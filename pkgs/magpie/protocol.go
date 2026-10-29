package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// mcpStreams prepares the call's SSE receiver before the child can ask the client anything.
// Restrict reordering to the known wait/send/Accept/open/defer lifecycle; unknown shapes fail.
func (f *sourceFile) mcpStream(fn *ast.FuncDecl) error {
	matched := 0
	if fn.Body == nil {
		return nil
	}
	{
		httpHandler := false
		for _, field := range fn.Type.Params.List {
			if selector, ok := field.Type.(*ast.SelectorExpr); ok && f.text(selector) == f.importName("net/http")+".ResponseWriter" {
				httpHandler = true
			}
		}
		if !httpHandler {
			return nil
		}
		body := fn.Body.List
		for i, stmt := range body {
			assignment, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
				continue
			}
			wait, ok := assignment.Rhs[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			child, ok := wait.Fun.(*ast.SelectorExpr)
			if !ok || child.Sel.Name != "wait" {
				continue
			}
			matched++
			sendAt, guardAt, openAt := -1, -1, -1
			var guard *ast.IfStmt
			var registration *ast.AssignStmt
			var open *ast.CallExpr
			for j := i + 1; j < len(body); j++ {
				switch node := body[j].(type) {
				case *ast.IfStmt:
					if init, ok := node.Init.(*ast.AssignStmt); ok && len(init.Rhs) == 1 {
						if call, ok := init.Rhs[0].(*ast.CallExpr); ok && f.methodCall(call, child.X, "send") {
							if sendAt != -1 {
								return fmt.Errorf("%s: ambiguous MCP send", f.path)
							}
							sendAt = j
						}
					}
					if condition, ok := node.Cond.(*ast.UnaryExpr); ok && condition.Op == token.NOT {
						if call, ok := condition.X.(*ast.CallExpr); ok && f.call(call, "strings", "Contains") && len(call.Args) == 2 && firstString(call.Args[1]) == "text/event-stream" {
							if guard != nil {
								return fmt.Errorf("%s: ambiguous MCP Accept guard", f.path)
							}
							guardAt, guard = j, node
						}
					}
				case *ast.AssignStmt:
					if len(node.Lhs) == 1 && len(node.Rhs) == 1 && node.Tok == token.DEFINE {
						if call, ok := node.Rhs[0].(*ast.CallExpr); ok {
							if method, ok := call.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "open" && len(call.Args) == 0 {
								if openAt != -1 {
									return fmt.Errorf("%s: ambiguous MCP stream registration", f.path)
								}
								openAt, registration, open = j, node, call
							}
						}
					}
				}
			}
			if sendAt == -1 || guard == nil || guard.Init != nil || guard.Else != nil {
				return fmt.Errorf("%s: cannot locate MCP send and SSE guard", f.path)
			}
			condition := guard.Cond.(*ast.UnaryExpr).X
			// Our conditional registration is already before send on a second pass.
			if openAt == -1 && sendAt > i+1 {
				if prepared, ok := body[sendAt-1].(*ast.IfStmt); ok && f.text(prepared.Cond) == f.text(condition) {
					if len(prepared.Body.List) == 2 {
						set, a := prepared.Body.List[0].(*ast.AssignStmt)
						cleanup, b := prepared.Body.List[1].(*ast.DeferStmt)
						if a && b && len(set.Rhs) == 1 && len(set.Lhs) == 1 {
							if call, ok := set.Rhs[0].(*ast.CallExpr); ok {
								if method, ok := call.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "open" && f.methodCall(cleanup.Call, method.X, "close") && len(cleanup.Call.Args) == 1 && f.text(cleanup.Call.Args[0]) == f.text(set.Lhs[0]) {
									continue
								}
							}
						}
					}
				}
			}
			if sendAt != i+1 || guardAt <= sendAt || openAt != guardAt+1 || openAt+1 >= len(body) {
				return fmt.Errorf("%s: unsupported MCP stream ordering", f.path)
			}
			receiver, known := open.Fun.(*ast.SelectorExpr).X.(*ast.Ident)
			if !known || receiver.Obj == nil || receiver.Obj.Pos() >= body[sendAt].Pos() {
				return fmt.Errorf("%s: unknown MCP stream receiver", f.path)
			}
			cleanup, ok := body[openAt+1].(*ast.DeferStmt)
			if !ok || !f.methodCall(cleanup.Call, receiver, "close") || len(cleanup.Call.Args) != 1 || f.text(cleanup.Call.Args[0]) != f.text(registration.Lhs[0]) {
				return fmt.Errorf("%s: cannot locate MCP stream cleanup", f.path)
			}
			name := f.text(registration.Lhs[0])
			f.insert(body[sendAt].Pos(), "var "+name+" chan []byte\nif "+f.text(condition)+" {\n"+name+" = "+f.text(open)+"\n"+f.text(cleanup)+"\n}\n")
			f.replace(registration, "")
			f.replace(cleanup, "")
		}
	}
	if matched != 1 {
		return fmt.Errorf("%s: expected one MCP stream lifecycle, found %d", f.path, matched)
	}
	return nil
}

// handoverListeners follows cancellation handles returned by listening, not test names.
// Wait for the listener to close, not for the server to finish draining its live stream.
func (f *sourceFile) handoverListeners(fn *ast.FuncDecl) {
	test := f.testReceiver(fn.Type.Params)
	if test == "" {
		return
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			assignment, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 2 || len(assignment.Rhs) != 1 {
				continue
			}
			call, ok := assignment.Rhs[0].(*ast.CallExpr)
			if !ok || !f.call(call, "", "listening") || len(call.Args) != 2 {
				continue
			}
			cancel, ok := assignment.Lhs[0].(*ast.Ident)
			if !ok || cancel.Name == "_" {
				continue
			}
			server := call.Args[1]
			handedOver := false
			for j := i + 1; j < len(block.List); j++ {
				expr, ok := block.List[j].(*ast.ExprStmt)
				if !ok {
					continue
				}
				operation, ok := expr.X.(*ast.CallExpr)
				if !ok {
					continue
				}
				if f.methodCall(operation, server, "HandOff") {
					handedOver = true
				}
				if !handedOver || !f.call(operation, "", cancel.Name) {
					continue
				}
				if j+1 < len(block.List) && strings.Contains(f.text(block.List[j+1]), "handed-over listener did not close") {
					continue
				}
				f.insert(expr.End(), `
for deadline := `+f.qualify("time", "Now")+`().Add(10 * `+f.qualify("time", "Second")+`); ; {
 file, err := `+f.text(server)+`.ListenerFile()
 if `+f.qualify("errors", "Is")+`(err, `+f.qualify("net", "ErrClosed")+`) { break }
 if err != nil { `+test+`.Fatal(err) }
 file.Close()
 if `+f.qualify("time", "Now")+`().After(deadline) { `+test+`.Fatal("handed-over listener did not close") }
 `+f.qualify("time", "Sleep")+`(`+f.qualify("time", "Millisecond")+`)
}`)
			}
		}
		return true
	})
}
