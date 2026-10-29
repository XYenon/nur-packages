package main

import (
	"go/ast"
	"go/types"
)

// Tarjan components distinguish recursion from repeated visits through an acyclic helper graph.
func (x *sourceIndex) recursiveFunctions() map[*function]bool {
	edges := map[*function][]*function{}
	for _, fn := range x.functions {
		ast.Inspect(fn.body, func(n ast.Node) bool {
			if _, closure := n.(*ast.FuncLit); closure {
				return false
			}
			if call, ok := n.(*ast.CallExpr); ok {
				if next := x.resolve(call.Fun, map[types.Object]bool{}); next != nil {
					edges[fn] = append(edges[fn], next)
				}
			}
			return true
		})
	}
	serial := 0
	number, low := map[*function]int{}, map[*function]int{}
	active := map[*function]bool{}
	recursive := map[*function]bool{}
	var stack []*function
	var visit func(*function)
	visit = func(fn *function) {
		serial++
		number[fn], low[fn] = serial, serial
		stack = append(stack, fn)
		active[fn] = true
		for _, next := range edges[fn] {
			if number[next] == 0 {
				visit(next)
				if low[next] < low[fn] {
					low[fn] = low[next]
				}
			} else if active[next] && number[next] < low[fn] {
				low[fn] = number[next]
			}
		}
		if low[fn] != number[fn] {
			return
		}
		var component []*function
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			active[last] = false
			component = append(component, last)
			if last == fn {
				break
			}
		}
		if len(component) > 1 {
			for _, member := range component {
				recursive[member] = true
			}
		} else {
			for _, next := range edges[fn] {
				if next == fn {
					recursive[fn] = true
				}
			}
		}
	}
	for _, fn := range x.functions {
		if number[fn] == 0 {
			visit(fn)
		}
	}
	return recursive
}
