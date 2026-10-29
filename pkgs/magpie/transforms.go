package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

// A rule owns its edits and imports. Unsupported candidates cannot invalidate unrelated edits.
func candidate(files []*sourceFile, rule func() bool) bool {
	unique := map[*sourceFile]bool{}
	var selected []*sourceFile
	for _, f := range files {
		if !unique[f] {
			unique[f] = true
			selected = append(selected, f)
		}
	}
	files = selected
	sources := make(map[*sourceFile][]byte, len(files))
	for _, f := range files {
		sources[f] = append([]byte(nil), f.source...)
	}
	rollback := func() {
		for _, f := range files {
			_ = reparse(f, sources[f])
		}
	}
	if !rule() {
		rollback()
		return false
	}
	result := map[*sourceFile][]byte{}
	for _, f := range files {
		source, err := applyEdits(f.source, f.edits)
		if err != nil {
			rollback()
			return false
		}
		if !bytes.Equal(source, f.source) {
			source, err = format.Source(source)
			if err != nil {
				rollback()
				return false
			}
		}
		result[f] = source
	}
	for _, f := range files {
		if !bytes.Equal(result[f], f.source) {
			if reparse(f, result[f]) != nil {
				rollback()
				return false
			}
		} else {
			f.edits = nil
		}
	}
	return true
}

// The protocol adapter identifies a diagnostic wait formatter, not arbitrary mutable reads.
func snapshotReports(files []*sourceFile, o options) {
	x := indexSources(files, o)
	for _, fn := range x.functions {
		if strings.HasSuffix(fn.file.path, "_test.go") {
			continue
		}
		var report *ast.CallExpr
		var target *function
		var publish *ast.IfStmt
		for _, stmt := range fn.body.List {
			if branch, ok := stmt.(*ast.IfStmt); ok {
				ast.Inspect(branch.Cond, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "answer" {
							publish = branch
						}
					}
					return true
				})
			}
			if assignment, ok := stmt.(*ast.AssignStmt); ok {
				for _, value := range assignment.Rhs {
					if call, ok := value.(*ast.CallExpr); ok {
						callee := x.resolve(call.Fun, map[types.Object]bool{})
						if callee == nil {
							continue
						}
						if fn.file.text(call.Fun) != "" && callee.file.hasCall(callee.body, "runtime/pprof", "Lookup") {
							continue
						}
						if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "heldReport" {
							report, target = call, callee
						}
					}
				}
			}
		}
		if report == nil || target == nil || publish == nil || report.Pos() < publish.Pos() {
			continue
		}
		var waits *ast.CallExpr
		ast.Inspect(target.body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "waitsSaid" && len(call.Args) == 0 {
					waits = call
				}
			}
			return true
		})
		if waits == nil || len(report.Args) == 0 || len(target.typ.Params.List) == 0 {
			continue
		}
		// The formatter returns a string copy under its own lock; it never holds that lock on return.
		formatter := x.resolve(waits.Fun, map[types.Object]bool{})
		if formatter == nil || formatter.typ.Results == nil || len(formatter.typ.Results.List) != 1 || formatter.file.text(formatter.typ.Results.List[0].Type) != "string" {
			continue
		}
		if !formatter.file.hasCall(formatter.body, "time", "Now") {
			continue
		}
		locks, unlocks := 0, 0
		ast.Inspect(formatter.body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
					if selector.Sel.Name == "Lock" {
						locks++
					}
					if selector.Sel.Name == "Unlock" {
						unlocks++
					}
				}
			}
			return true
		})
		if locks != 1 || unlocks != 1 {
			continue
		}
		calls := 0
		for _, source := range files {
			ast.Inspect(source.file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok && x.resolve(call.Fun, map[types.Object]bool{}) == target {
					calls++
				}
				return true
			})
		}
		if calls != 1 {
			continue
		}
		if strings.Contains(fn.file.text(fn.body), "nixWaitSnapshot") {
			continue
		}
		publishText := fn.file.text(publish)
		candidate([]*sourceFile{fn.file, target.file}, func() bool {
			expression := fn.file.text(report.Args[0]) + "." + waits.Fun.(*ast.SelectorExpr).Sel.Name + "()"
			fn.file.insert(publish.Pos(), "nixWaitSnapshot := "+expression+"\n")
			fn.file.insert(report.Rparen, ", nixWaitSnapshot")
			target.file.insert(target.typ.Params.Closing, ", nixWaitSnapshot string")
			target.file.replace(waits, "nixWaitSnapshot")
			// No cancellation, response handling or report I/O is moved before publication.
			return publishText != ""
		})
		break // rebuild bindings before considering another cross-file candidate
	}
}

func receiverTempDir(f *sourceFile, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "TempDir" || len(call.Args) != 0 {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && f.testingParams()[receiver.Obj] != ""
}

// Shell fixtures explicitly source their own rc; suppress host startup files only there.
func shellFixtures(files []*sourceFile, o options) {
	x := indexSources(files, o)
	for _, fn := range x.functions {
		if !x.testFunction(fn) && !strings.HasSuffix(fn.file.path, "_test.go") {
			continue
		}
		ast.Inspect(fn.body, func(n ast.Node) bool {
			if _, closure := n.(*ast.FuncLit); closure {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok || !fn.file.call(call, "os/exec", "Command") || len(call.Args) < 3 {
				return true
			}
			if firstString(call.Args[1]) != "-c" {
				return true
			}
			shell := fn.file.commandName(call)
			if shell == "" {
				if id, ok := call.Args[0].(*ast.Ident); ok && id.Obj != nil {
					if assign, ok := id.Obj.Decl.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 {
						if lookup, ok := assign.Rhs[0].(*ast.CallExpr); ok && fn.file.call(lookup, "os/exec", "LookPath") {
							shell = "fixture"
						}
					}
				}
			}
			script := fn.file.text(call.Args[2])
			if shell != "zsh" && shell != "fixture" {
				return true
			}
			if !strings.Contains(script, "$HOME/") || (!strings.Contains(script, ". ") && !strings.Contains(script, "source ")) {
				return true
			}
			// -f is shared by bash and zsh, and preserves the explicit rc load below.
			if !strings.Contains(fn.file.text(fn.body), "PATH=") {
				return true
			}
			if shell == "zsh" {
				fn.file.insert(call.Args[1].Pos(), strconv.Quote("-f")+", ")
			} else if len(call.Args) == 3 {
				path := fn.file.text(call.Args[0])
				args := `func() []string { if ` + fn.file.qualify("path/filepath", "Base") + `(` + path + `) == "zsh" { return []string{"-f", "-c", ` + script + `} }; return []string{"-c", ` + script + `} }()...`
				fn.file.replace(call, fn.file.text(call.Fun)+"("+path+", "+args+")")
			}
			return true
		})
	}
}

func typeString(f *sourceFile, t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p.Name() == f.file.Name.Name && p.Path() == strings.TrimSuffix(modulePath, "/")+"/"+strings.TrimPrefix(f.path[:strings.LastIndex(f.path, "/")], "./") {
			return ""
		}
		return strings.TrimSuffix(f.qualify(p.Path(), ""), ".")
	})
}

// Reproduce Go's function/argument evaluation at the caller with explicit parameter types.
func evaluatedCall(f *sourceFile, call *ast.CallExpr, signature *types.Signature) (string, string, bool) {
	if signature.TypeParams() != nil && signature.TypeParams().Len() > 0 {
		return "", "", false
	}
	var setup, args []string
	setup = append(setup, "nixTaskCall := "+f.text(call.Fun))
	if len(call.Args) == 1 && signature.Params().Len() > 1 {
		// Go permits f(g()) when g returns exactly f's parameter tuple.
		for i := 0; i < signature.Params().Len(); i++ {
			name := fmt.Sprintf("nixTaskArg%d", i)
			setup = append(setup, "var "+name+" "+typeString(f, signature.Params().At(i).Type()))
			args = append(args, name)
		}
		setup = append(setup, strings.Join(args, ",")+" = "+f.text(call.Args[0]))
		return strings.Join(setup, "\n"), "nixTaskCall(" + strings.Join(args, ",") + ")", true
	}
	if len(call.Args) != signature.Params().Len() && !signature.Variadic() {
		return "", "", false
	}
	for i, arg := range call.Args {
		index := i
		if index >= signature.Params().Len() {
			index = signature.Params().Len() - 1
		}
		if index < 0 {
			return "", "", false
		}
		typ := signature.Params().At(index).Type()
		if signature.Variadic() && index == signature.Params().Len()-1 && !call.Ellipsis.IsValid() {
			slice, ok := typ.(*types.Slice)
			if !ok {
				return "", "", false
			}
			typ = slice.Elem()
		}
		if _, tuple := typ.(*types.Tuple); tuple {
			return "", "", false
		}
		name := fmt.Sprintf("nixTaskArg%d", i)
		setup = append(setup, "var "+name+" "+typeString(f, typ)+" = "+f.text(arg))
		args = append(args, name)
	}
	if call.Ellipsis.IsValid() && len(args) > 0 {
		args[len(args)-1] += "..."
	}
	return strings.Join(setup, "\n"), "nixTaskCall(" + strings.Join(args, ",") + ")", true
}

func taskResources(files []*sourceFile, o options) {
	for taskResourcePass(files, o) {
	}
}

func taskResourcePass(files []*sourceFile, o options) bool {
	x := indexSources(files, o)
	reachable := x.reachableTests()
	recursive := x.recursiveFunctions()
	summaries := map[*function]taskSummary{}
	visiting := map[*function]bool{}
	var summarize func(*function) taskSummary
	summarize = func(fn *function) taskSummary {
		if recursive[fn] {
			return taskSummary{unknown: true}
		}
		if result, ok := summaries[fn]; ok {
			return result
		}
		if visiting[fn] {
			return taskSummary{unknown: true}
		}
		visiting[fn] = true
		result := taskSummary{}
		ast.Inspect(fn.body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncLit:
				return false
			case *ast.BranchStmt:
				if n.Tok == token.GOTO {
					result.unknown = true
				}
			case *ast.GoStmt, *ast.SelectStmt, *ast.SendStmt:
				result.unknown = true
				return false
			case *ast.UnaryExpr:
				if n.Op == token.ARROW {
					result.unknown = true
				}
			case *ast.ForStmt:
				if !boundedFor(x, n) {
					result.unknown = true
				}
			case *ast.RangeStmt:
				typ := x.info.TypeOf(n.X)
				if typ == nil {
					result.unknown = true
				} else {
					if _, channel := typ.Underlying().(*types.Chan); channel {
						result.unknown = true
					}
				}
			case *ast.CallExpr:
				p, name := x.callName(n)
				if p == "time" && name == "Sleep" && len(n.Args) == 1 {
					if _, constant := x.info.Types[n.Args[0]]; constant && x.info.Types[n.Args[0]].Value != nil {
						return true
					}
				}
				if pathArg, ok := fileOperation(p, name); ok {
					if pathArg < len(n.Args) {
						result.paths = append(result.paths, n.Args[pathArg])
						if name == "Rename" {
							result.paths = append(result.paths, n.Args[0])
						}
					}
					return true
				}
				if safeOperation(p, name) {
					if p == "fmt" {
						for _, arg := range n.Args {
							typ := x.info.TypeOf(arg)
							if typ == nil {
								result.unknown = true
								continue
							}
							if _, basic := typ.Underlying().(*types.Basic); !basic {
								result.unknown = true
							}
						}
					}
					return true
				}
				if x.info.Types[n.Fun].IsType() {
					return true
				}
				callee := x.resolve(n.Fun, map[types.Object]bool{})
				if callee == nil {
					result.unknown = true
					return true
				}
				child := summarize(callee)
				result.unknown = result.unknown || child.unknown
				for _, path := range child.paths {
					mapped := mapParameterPath(x, callee, n, path)
					if mapped == nil {
						result.unknown = true
					} else {
						result.paths = append(result.paths, mapped)
					}
				}
			}
			return true
		})
		delete(visiting, fn)
		summaries[fn] = result
		return result
	}
	changed := false
	for _, fn := range x.functions {
		if changed {
			break
		}
		if !reachable[fn] || fn.file.hasCall(fn.body, "testing/synctest", "Test") {
			continue
		}
		ast.Inspect(fn.body, func(n ast.Node) bool {
			if changed {
				return false
			}
			if _, closure := n.(*ast.FuncLit); closure {
				return false
			}
			stmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}
			callee := x.resolve(stmt.Call.Fun, map[types.Object]bool{})
			var summary taskSummary
			if callee == nil {
				p, name := x.callName(stmt.Call)
				index, ok := fileOperation(p, name)
				if !ok || len(stmt.Call.Args) <= index {
					return false
				}
				summary.paths = []ast.Expr{stmt.Call.Args[index]}
				if name == "Rename" {
					summary.paths = append(summary.paths, stmt.Call.Args[0])
				}
			} else {
				summary = summarize(callee)
			}
			if summary.unknown || len(summary.paths) == 0 {
				return false
			}
			// Anonymous closures retain outer bindings; named callees map path parameters at this call.
			var paths []ast.Expr
			for _, path := range summary.paths {
				mapped := path
				if callee != nil {
					mapped = mapParameterPath(x, callee, stmt.Call, path)
				}
				if mapped == nil || !x.stablePath(fn.file, mapped) {
					return false
				}
				paths = append(paths, mapped)
			}
			signature, ok := x.info.TypeOf(stmt.Call.Fun).(*types.Signature)
			if !ok {
				return false
			}
			changed = candidate([]*sourceFile{fn.file}, func() bool {
				setup, invoke, ok := evaluatedCall(fn.file, stmt.Call, signature)
				if !ok {
					return false
				}
				var resources []string
				for _, path := range paths {
					argument := -1
					for i, arg := range stmt.Call.Args {
						if path == arg {
							argument = i
							break
						}
					}
					if argument >= 0 {
						resources = append(resources, fmt.Sprintf("nixTaskArg%d", argument))
						continue
					}
					text, ok := resourcePath(x, fn.file, path)
					if !ok {
						return false
					}
					resources = append(resources, text)
				}
				begin := fn.file.qualify(taskImport, "Begin")
				fn.file.replace(stmt, "{\n"+setup+"\nnixTaskDone := "+begin+"("+strings.Join(resources, ",")+")\ngo func() { defer nixTaskDone(); "+invoke+" }()\n}")
				return true
			})
			return false
		})
	}
	return changed
}

// Register each test directory independently; plugin cleanup wrappers keep their own ordering.
func registerTaskDirectories(files []*sourceFile) {
	for _, f := range files {
		ast.Inspect(f.file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if ok && receiverTempDir(f, call) {
				selector := call.Fun.(*ast.SelectorExpr)
				f.replace(call.Fun, f.qualify(taskImport, "TempDir"))
				f.insert(call.Lparen+1, f.text(selector.X))
			}
			return true
		})
	}
}

func boundedFor(x *sourceIndex, loop *ast.ForStmt) bool {
	initial, ok := loop.Init.(*ast.AssignStmt)
	if !ok || initial.Tok != token.DEFINE || len(initial.Lhs) != 1 || len(initial.Rhs) != 1 {
		return false
	}
	counter, ok := initial.Lhs[0].(*ast.Ident)
	if !ok {
		return false
	}
	condition, ok := loop.Cond.(*ast.BinaryExpr)
	if !ok || (condition.Op != token.LSS && condition.Op != token.LEQ) {
		return false
	}
	id, ok := condition.X.(*ast.Ident)
	if !ok || x.info.ObjectOf(id) != x.info.ObjectOf(counter) {
		return false
	}
	increment, ok := loop.Post.(*ast.IncDecStmt)
	if !ok || increment.Tok != token.INC {
		return false
	}
	id, ok = increment.X.(*ast.Ident)
	if !ok || x.info.ObjectOf(id) != x.info.ObjectOf(counter) {
		return false
	}
	if x.info.Types[condition.Y].Value == nil {
		return false
	}
	modified := false
	ast.Inspect(loop.Body, func(n ast.Node) bool {
		if increment, ok := n.(*ast.IncDecStmt); ok {
			if id, ok := increment.X.(*ast.Ident); ok && x.info.ObjectOf(id) == x.info.ObjectOf(counter) {
				modified = true
			}
		}
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && x.info.ObjectOf(id) == x.info.ObjectOf(counter) {
					modified = true
				}
			}
		}
		return true
	})
	return !modified
}

type taskSummary struct {
	paths   []ast.Expr
	unknown bool
}

func fileOperation(path, name string) (int, bool) {
	if path != "os" {
		return 0, false
	}
	switch name {
	case "WriteFile", "Remove", "RemoveAll", "Mkdir", "MkdirAll", "Create", "OpenFile":
		return 0, true
	case "Rename", "Symlink":
		return 1, true
	}
	return 0, false
}
func safeOperation(path, name string) bool {
	if path == "" {
		switch name {
		case "len", "cap", "append", "copy", "delete", "make", "new", "close":
			return true
		}
	}
	switch path {
	case "strings", "bytes":
		switch name {
		case "Contains", "HasPrefix", "HasSuffix", "TrimPrefix", "TrimSuffix", "TrimSpace", "Join", "Split", "Replace", "ReplaceAll", "Equal", "Compare":
			return true
		}
	case "path/filepath":
		switch name {
		case "Join", "Dir", "Base", "Clean", "Ext", "IsAbs":
			return true
		}
	case "errors":
		return name == "New"
	case "fmt":
		return name == "Sprintf" || name == "Sprint"
	case "os":
		switch name {
		case "Stat", "Lstat", "ReadFile", "IsNotExist":
			return true
		}
	case "time":
		return name == "Now" || name == "Since" || name == "Until"
	}
	return false
}
func resourcePath(x *sourceIndex, f *sourceFile, expr ast.Expr) (string, bool) {
	switch node := expr.(type) {
	case *ast.Ident:
		return node.Name, true
	case *ast.BasicLit:
		return node.Value, true
	case *ast.BinaryExpr:
		a, ok := resourcePath(x, f, node.X)
		if !ok {
			return "", false
		}
		b, ok := resourcePath(x, f, node.Y)
		return "(" + a + " + " + b + ")", ok
	case *ast.CallExpr:
		path, name := x.callName(node)
		if path != "path/filepath" {
			return "", false
		}
		var args []string
		for _, arg := range node.Args {
			text, ok := resourcePath(x, f, arg)
			if !ok {
				return "", false
			}
			args = append(args, text)
		}
		return f.qualify(path, name) + "(" + strings.Join(args, ",") + ")", true
	}
	return "", false
}

func mapParameterPath(x *sourceIndex, fn *function, call *ast.CallExpr, path ast.Expr) ast.Expr {
	return mappedPath(x, fn, call, path, map[types.Object]bool{})
}

func mappedPath(x *sourceIndex, fn *function, call *ast.CallExpr, path ast.Expr, seen map[types.Object]bool) ast.Expr {
	if binary, ok := path.(*ast.BinaryExpr); ok && binary.Op == token.ADD {
		a, b := mappedPath(x, fn, call, binary.X, seen), mappedPath(x, fn, call, binary.Y, seen)
		if a == nil || b == nil {
			return nil
		}
		return &ast.BinaryExpr{X: a, Op: token.ADD, Y: b}
	}
	if literal, ok := path.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		return literal
	}
	if nested, ok := path.(*ast.CallExpr); ok {
		p, name := x.callName(nested)
		if p != "path/filepath" || (name != "Join" && name != "Dir" && name != "Clean") {
			return nil
		}
		var args []ast.Expr
		for _, arg := range nested.Args {
			mapped := mappedPath(x, fn, call, arg, seen)
			if mapped == nil {
				return nil
			}
			args = append(args, mapped)
		}
		return &ast.CallExpr{Fun: nested.Fun, Args: args}
	}
	if id, ok := path.(*ast.Ident); ok {
		object := x.info.ObjectOf(id)
		if object == nil || seen[object] {
			return nil
		}
		seen[object] = true
		defer delete(seen, object)
		if fn.object == nil && object.Pos() < fn.body.Pos() {
			return id
		}
		if id.Obj != nil {
			if assignment, ok := id.Obj.Decl.(*ast.AssignStmt); ok && assignment.Tok == token.DEFINE && len(assignment.Lhs) == len(assignment.Rhs) {
				for i, lhs := range assignment.Lhs {
					if bound, ok := lhs.(*ast.Ident); ok && bound.Obj == id.Obj {
						return mappedPath(x, fn, call, assignment.Rhs[i], seen)
					}
				}
			}
		}
		index := 0
		for _, field := range fn.typ.Params.List {
			for _, param := range field.Names {
				if x.info.ObjectOf(param) == object {
					if index < len(call.Args) {
						return call.Args[index]
					}
					return nil
				}
				index++
			}
		}
	}
	return nil
}
