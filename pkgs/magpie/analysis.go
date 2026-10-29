package main

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
)

const modulePath = "github.com/yetone/magpie/"
const taskImport = modulePath + "internal/nixtasks"

type function struct {
	file   *sourceFile
	typ    *ast.FuncType
	body   *ast.BlockStmt
	object types.Object
}
type sourceIndex struct {
	info      *types.Info
	functions []*function
	objects   map[types.Object]*function
	nodes     map[*ast.FuncLit]*function
	packages  map[string]*types.Package
	byPath    map[string][]*sourceFile
	loading   map[string]bool
	std       types.Importer
	mutated   map[types.Object]bool
}

// Missing dependency types limit individual transformations, not source compilation.
func indexSources(files []*sourceFile, o options) *sourceIndex {
	x := &sourceIndex{info: &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}, objects: map[types.Object]*function{}, nodes: map[*ast.FuncLit]*function{}, packages: map[string]*types.Package{}, byPath: map[string][]*sourceFile{}, loading: map[string]bool{}, std: importer.Default()}
	context := build.Default
	context.GOOS, context.CgoEnabled = o.platform, o.gui
	if o.arch != "" {
		context.GOARCH = o.arch
	}
	if o.gui {
		context.BuildTags = []string{"production"}
		if o.platform == "linux" {
			context.BuildTags = append(context.BuildTags, "gtk3")
		}
	} else {
		context.BuildTags = []string{"nogui"}
	}
	for _, f := range files {
		active, err := context.MatchFile(filepath.Dir(f.path), filepath.Base(f.path))
		if err != nil || !active {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(f.path))
		path := modulePath + dir
		if dir == "." {
			path = strings.TrimSuffix(modulePath, "/")
		}
		if strings.HasSuffix(f.file.Name.Name, "_test") {
			path += "_test"
		}
		x.byPath[path] = append(x.byPath[path], f)
	}
	paths := make([]string, 0, len(x.byPath))
	for p := range x.byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		x.Import(p)
	}
	x.mutated = map[types.Object]bool{}
	for _, p := range paths {
		for _, f := range x.byPath[p] {
			ast.Inspect(f.file, func(n ast.Node) bool {
				if assignment, ok := n.(*ast.AssignStmt); ok && assignment.Tok != token.DEFINE {
					for _, lhs := range assignment.Lhs {
						if id, ok := lhs.(*ast.Ident); ok {
							x.mutated[x.info.ObjectOf(id)] = true
						}
					}
				}
				return true
			})
		}
	}
	for _, p := range paths {
		for _, f := range x.byPath[p] {
			ast.Inspect(f.file, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.FuncDecl:
					if n.Body != nil {
						fn := &function{file: f, typ: n.Type, body: n.Body, object: x.info.Defs[n.Name]}
						x.functions = append(x.functions, fn)
						if fn.object != nil {
							x.objects[fn.object] = fn
						}
					}
				case *ast.FuncLit:
					fn := &function{file: f, typ: n.Type, body: n.Body}
					x.nodes[n] = fn
					x.functions = append(x.functions, fn)
				}
				return true
			})
		}
	}
	return x
}

func (x *sourceIndex) Import(path string) (*types.Package, error) {
	if p := x.packages[path]; p != nil {
		return p, nil
	}
	files := x.byPath[path]
	if len(files) == 0 {
		return x.std.Import(path)
	}
	if x.loading[path] {
		return types.NewPackage(path, files[0].file.Name.Name), nil
	}
	x.loading[path] = true
	// All sources use one FileSet; token positions remain comparable across files.
	config := types.Config{Importer: x, Error: func(error) {}, DisableUnusedImportCheck: true}
	syntax := make([]*ast.File, 0, len(files))
	for _, f := range files {
		syntax = append(syntax, f.file)
	}
	p, _ := config.Check(path, files[0].fset, syntax, x.info)
	if p == nil {
		p = types.NewPackage(path, files[0].file.Name.Name)
	}
	x.packages[path] = p
	delete(x.loading, path)
	return p, nil
}

func (x *sourceIndex) resolve(expr ast.Expr, seen map[types.Object]bool) *function {
	switch n := expr.(type) {
	case *ast.FuncLit:
		return x.nodes[n]
	case *ast.ParenExpr:
		return x.resolve(n.X, seen)
	case *ast.Ident:
		object := x.info.ObjectOf(n)
		if fn := x.objects[object]; fn != nil {
			return fn
		}
		if object == nil || seen[object] {
			return nil
		}
		seen[object] = true
		if x.mutated[object] {
			return nil
		}
		if n.Obj != nil {
			switch decl := n.Obj.Decl.(type) {
			case *ast.AssignStmt:
				if decl.Tok != token.DEFINE || len(decl.Lhs) != len(decl.Rhs) {
					return nil
				}
				for i, lhs := range decl.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && id.Obj == n.Obj {
						return x.resolve(decl.Rhs[i], seen)
					}
				}
			case *ast.ValueSpec:
				if len(decl.Names) == len(decl.Values) {
					for i, id := range decl.Names {
						if id.Obj == n.Obj {
							return x.resolve(decl.Values[i], seen)
						}
					}
				}
			}
		}
	case *ast.SelectorExpr:
		if selection := x.info.Selections[n]; selection != nil {
			return x.objects[selection.Obj()]
		}
		return x.objects[x.info.ObjectOf(n.Sel)]
	}
	return nil
}

func (x *sourceIndex) callName(call *ast.CallExpr) (string, string) {
	var object types.Object
	switch n := call.Fun.(type) {
	case *ast.Ident:
		object = x.info.ObjectOf(n)
	case *ast.SelectorExpr:
		if s := x.info.Selections[n]; s != nil {
			object = s.Obj()
		} else {
			object = x.info.ObjectOf(n.Sel)
		}
	}
	if object == nil {
		return "", ""
	}
	if object.Pkg() == nil {
		return "", object.Name()
	}
	return object.Pkg().Path(), object.Name()
}

func (x *sourceIndex) testFunction(fn *function) bool {
	for _, field := range fn.typ.Params.List {
		typ := x.info.TypeOf(field.Type)
		if pointer, ok := typ.(*types.Pointer); ok {
			typ = pointer.Elem()
		}
		if named, ok := typ.(*types.Named); ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "testing" {
			return true
		}
	}
	return false
}

func (x *sourceIndex) reachableTests() map[*function]bool {
	reachable := map[*function]bool{}
	var visit func(*function)
	visit = func(fn *function) {
		if reachable[fn] {
			return
		}
		reachable[fn] = true
		ast.Inspect(fn.body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if next := x.resolve(call.Fun, map[types.Object]bool{}); next != nil {
					visit(next)
				}
			}
			if closure, ok := n.(*ast.FuncLit); ok {
				if next := x.nodes[closure]; next != nil {
					visit(next)
				}
				return false
			}
			return true
		})
	}
	for _, fn := range x.functions {
		if strings.HasSuffix(fn.file.path, "_test.go") && x.testFunction(fn) {
			visit(fn)
		}
	}
	return reachable
}

// stablePath follows immutable local aliases, parameters and filepath operations.
// Expressions with reads or side effects cannot be evaluated again for task registration.
func (x *sourceIndex) stablePath(f *sourceFile, e ast.Expr) bool {
	switch n := e.(type) {
	case *ast.Ident:
		object, ok := x.info.ObjectOf(n).(*types.Var)
		if !ok || object.Parent() == nil || object.Pkg() != nil && object.Parent() == object.Pkg().Scope() {
			return false
		}
		mutable := false
		ast.Inspect(f.file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				if node.Tok != token.DEFINE {
					for _, lhs := range node.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && x.info.ObjectOf(id) == object {
							mutable = true
						}
					}
				}
			case *ast.UnaryExpr:
				if node.Op == token.AND {
					if id, ok := node.X.(*ast.Ident); ok && x.info.ObjectOf(id) == object {
						mutable = true
					}
				}
			}
			return true
		})
		return !mutable
	case *ast.BasicLit:
		return n.Kind == token.STRING
	case *ast.BinaryExpr:
		return n.Op == token.ADD && x.stablePath(f, n.X) && x.stablePath(f, n.Y)
	case *ast.CallExpr:
		p, name := x.callName(n)
		if p != "path/filepath" || (name != "Join" && name != "Dir" && name != "Clean") {
			return false
		}
		for _, arg := range n.Args {
			if !x.stablePath(f, arg) {
				return false
			}
		}
		return true
	}
	return false
}

func (x *sourceIndex) sourcePath(f *sourceFile, e ast.Expr, seen map[types.Object]bool) bool {
	if id, ok := e.(*ast.Ident); ok {
		object := x.info.ObjectOf(id)
		if object == nil || seen[object] {
			return false
		}
		seen[object] = true
		defer delete(seen, object)
		if id.Obj != nil {
			switch decl := id.Obj.Decl.(type) {
			case *ast.Field:
				return x.info.TypeOf(decl.Type) != nil && x.info.TypeOf(decl.Type).String() == "string"
			case *ast.AssignStmt:
				if len(decl.Lhs) == len(decl.Rhs) {
					for i, lhs := range decl.Lhs {
						if name, ok := lhs.(*ast.Ident); ok && name.Obj == id.Obj {
							return x.sourcePath(f, decl.Rhs[i], seen)
						}
					}
				}
			}
		}
	}
	if call, ok := e.(*ast.CallExpr); ok {
		p, name := x.callName(call)
		if p == "testing" && name == "TempDir" {
			return true
		}
		if p == "path/filepath" && name == "Join" && len(call.Args) > 0 {
			return x.sourcePath(f, call.Args[0], seen)
		}
	}
	return false
}

func reparse(f *sourceFile, source []byte) error {
	parsed, err := parser.ParseFile(f.fset, f.path, source, parser.ParseComments)
	if err != nil {
		return err
	}
	f.source = source
	f.file = parsed
	f.edits = nil
	return nil
}
