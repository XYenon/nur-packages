package main

import (
	"go/ast"
	"path/filepath"
	"strings"
)

// Directory clock adaptations follow packages using a timestamp-keyed directory listing.
// Content-only file clocks and tests deliberately setting Chtimes stay unchanged.
func discoveryPackages(files []*sourceFile) map[string]bool {
	dependencies := map[string][]string{}
	clock := map[string]bool{}
	for _, f := range files {
		dir := filepath.ToSlash(filepath.Dir(f.path))
		if strings.HasSuffix(f.path, "_test.go") {
			continue
		}
		listing, stamp := false, false
		ast.Inspect(f.file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if f.call(call, "os", "ReadDir") {
					listing = true
				}
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "ModTime" {
					stamp = true
				}
			}
			return true
		})
		clock[dir] = clock[dir] || listing && stamp
		for _, spec := range f.file.Imports {
			p := firstString(spec.Path)
			if strings.HasPrefix(p, modulePath) {
				dependencies[dir] = append(dependencies[dir], strings.TrimPrefix(p, modulePath))
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for dir, deps := range dependencies {
			if clock[dir] {
				continue
			}
			for _, dep := range deps {
				if clock[dep] {
					clock[dir] = true
					changed = true
					break
				}
			}
		}
	}
	return clock
}
