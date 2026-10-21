package tui_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	teaImportPath       = "charm.land/bubbletea/v2"
	tuiSubpackagePrefix = "github.com/DannyFestor/TuiSnip/internal/adapters/tui/"
	routingMethod       = "overlaysUpdatedSharingArrivals"
	modelPackageDir     = "."
	testFileSuffix      = "_test.go"
)

func TestModel_routesComponentMessages(t *testing.T) {
	t.Parallel()

	sources := parsedSources(t)
	returned := componentMessagesReturnedAsMsg(sources)
	routed := componentMessagesRouted(t, sources)

	require.NotEmpty(t, returned)
	assert.Empty(t, unrouted(returned, routed),
		"Model.Update drops these component messages: list each in the case that calls "+routingMethod)
}

type sourceFile struct {
	syntax       *ast.File
	imports      map[string]string
	inSubpackage bool
}

func parsedSources(t *testing.T) []sourceFile {
	t.Helper()

	var sources []sourceFile

	fileSet := token.NewFileSet()
	err := filepath.WalkDir(modelPackageDir, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil || !isProductionGoFile(entry) {
			return err
		}

		syntax, err := parser.ParseFile(fileSet, filePath, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		sources = append(sources, newSourceFile(syntax, filepath.Dir(filePath) != modelPackageDir))

		return nil
	})
	require.NoError(t, err)

	return sources
}

func isProductionGoFile(entry fs.DirEntry) bool {
	name := entry.Name()

	return !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, testFileSuffix)
}

func newSourceFile(syntax *ast.File, inSubpackage bool) sourceFile {
	imports := make(map[string]string, len(syntax.Imports))
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(spec.Path.Value, "\"`")
		imports[importName(spec, importPath)] = importPath
	}

	return sourceFile{syntax: syntax, imports: imports, inSubpackage: inSubpackage}
}

func importName(spec *ast.ImportSpec, importPath string) string {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(importPath)
}

func (f sourceFile) subpackageTypes(typeExprs []ast.Expr) []string {
	var types []string

	for _, typeExpr := range typeExprs {
		if name, ok := f.subpackageType(typeExpr); ok {
			types = append(types, name)
		}
	}

	return types
}

func (f sourceFile) subpackageType(typeExpr ast.Expr) (string, bool) {
	switch typeExpr := typeExpr.(type) {
	case *ast.Ident:
		return f.syntax.Name.Name + "." + typeExpr.Name, f.inSubpackage
	case *ast.SelectorExpr:
		importPath, ok := f.importPathOf(typeExpr)

		return path.Base(importPath) + "." + typeExpr.Sel.Name, ok && strings.HasPrefix(importPath, tuiSubpackagePrefix)
	default:
		return "", false
	}
}

func (f sourceFile) importPathOf(selector *ast.SelectorExpr) (string, bool) {
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	importPath, ok := f.imports[qualifier.Name]

	return importPath, ok
}

func (f sourceFile) returnsMsg(signature *ast.FuncType) bool {
	if signature.Results == nil || len(signature.Results.List) != 1 {
		return false
	}

	selector, ok := signature.Results.List[0].Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	importPath, ok := f.importPathOf(selector)

	return ok && importPath == teaImportPath && selector.Sel.Name == "Msg"
}

func componentMessagesReturnedAsMsg(sources []sourceFile) []string {
	var returned []string

	for _, source := range sources {
		ast.Inspect(source.syntax, func(node ast.Node) bool {
			if body, ok := source.msgFunctionBody(node); ok {
				returned = append(returned, source.subpackageTypes(literalTypesReturnedBy(body))...)
			}

			return true
		})
	}

	return returned
}

func (f sourceFile) msgFunctionBody(node ast.Node) (*ast.BlockStmt, bool) {
	switch function := node.(type) {
	case *ast.FuncLit:
		return function.Body, f.returnsMsg(function.Type)
	case *ast.FuncDecl:
		return function.Body, function.Body != nil && f.returnsMsg(function.Type)
	default:
		return nil, false
	}
}

func literalTypesReturnedBy(body *ast.BlockStmt) []ast.Expr {
	var types []ast.Expr

	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			types = append(types, literalTypes(node.Results)...)
		}

		return true
	})

	return types
}

func literalTypes(results []ast.Expr) []ast.Expr {
	var types []ast.Expr

	for _, result := range results {
		if literal, ok := result.(*ast.CompositeLit); ok {
			types = append(types, literal.Type)
		}
	}

	return types
}

func componentMessagesRouted(t *testing.T, sources []sourceFile) []string {
	t.Helper()

	for _, source := range sources {
		if update, ok := source.modelUpdate(); ok {
			return source.routedCaseTypes(t, update)
		}
	}

	require.FailNow(t, "found no Model.Update")

	return nil
}

func (f sourceFile) modelUpdate() (*ast.FuncDecl, bool) {
	if f.inSubpackage {
		return nil, false
	}

	for _, decl := range f.syntax.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if ok && function.Name.Name == "Update" && isModelReceiver(function.Recv) {
			return function, true
		}
	}

	return nil, false
}

func isModelReceiver(receiver *ast.FieldList) bool {
	if receiver == nil {
		return false
	}

	receiverType, ok := receiver.List[0].Type.(*ast.Ident)

	return ok && receiverType.Name == "Model"
}

func (f sourceFile) routedCaseTypes(t *testing.T, update *ast.FuncDecl) []string {
	t.Helper()

	var routed []string

	ast.Inspect(update.Body, func(node ast.Node) bool {
		clause, ok := node.(*ast.CaseClause)
		if ok && callsMethod(clause.Body, routingMethod) {
			routed = append(routed, f.subpackageTypes(clause.List)...)
		}

		return true
	})
	require.NotEmpty(t, routed, "found no case in Model.Update that calls "+routingMethod)

	return routed
}

func callsMethod(statements []ast.Stmt, method string) bool {
	calls := false

	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				calls = calls || (isSelector && selector.Sel.Name == method)
			}

			return !calls
		})
	}

	return calls
}

func unrouted(returned, routed []string) []string {
	missing := slices.DeleteFunc(slices.Clone(returned), func(name string) bool {
		return slices.Contains(routed, name)
	})
	slices.Sort(missing)

	return slices.Compact(missing)
}
