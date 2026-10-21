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
	componentImportPath = "github.com/DannyFestor/TuiSnip/internal/adapters/tui/"
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
	syntax      *ast.File
	imports     map[string]string
	isComponent bool
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

func newSourceFile(syntax *ast.File, isComponent bool) sourceFile {
	imports := make(map[string]string, len(syntax.Imports))
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(spec.Path.Value, "\"`")
		imports[importName(spec, importPath)] = importPath
	}

	return sourceFile{syntax: syntax, imports: imports, isComponent: isComponent}
}

func importName(spec *ast.ImportSpec, importPath string) string {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(importPath)
}

func (f sourceFile) componentType(typeExpr ast.Expr) (string, bool) {
	switch typeExpr := typeExpr.(type) {
	case *ast.Ident:
		return f.syntax.Name.Name + "." + typeExpr.Name, f.isComponent
	case *ast.SelectorExpr:
		return f.importedComponentType(typeExpr)
	default:
		return "", false
	}
}

func (f sourceFile) importedComponentType(selector *ast.SelectorExpr) (string, bool) {
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	importPath := f.imports[qualifier.Name]

	return path.Base(importPath) + "." + selector.Sel.Name, strings.HasPrefix(importPath, componentImportPath)
}

func (f sourceFile) returnsMsg(signature *ast.FuncType) bool {
	if signature.Results == nil || len(signature.Results.List) != 1 {
		return false
	}

	selector, ok := signature.Results.List[0].Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	qualifier, ok := selector.X.(*ast.Ident)

	return ok && f.imports[qualifier.Name] == teaImportPath && selector.Sel.Name == "Msg"
}

func componentMessagesReturnedAsMsg(sources []sourceFile) []string {
	var returned []string

	for _, source := range sources {
		ast.Inspect(source.syntax, func(node ast.Node) bool {
			if body, ok := source.msgFunctionBody(node); ok {
				returned = append(returned, source.componentLiteralsReturnedBy(body)...)
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

func (f sourceFile) componentLiteralsReturnedBy(body *ast.BlockStmt) []string {
	var returned []string

	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			returned = append(returned, f.componentLiteralTypes(node.Results)...)
		}

		return true
	})

	return returned
}

func (f sourceFile) componentLiteralTypes(results []ast.Expr) []string {
	var types []string

	for _, result := range results {
		literal, ok := result.(*ast.CompositeLit)
		if !ok {
			continue
		}

		if name, isComponent := f.componentType(literal.Type); isComponent {
			types = append(types, name)
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
	if f.isComponent {
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
			routed = append(routed, f.componentTypes(clause.List)...)
		}

		return true
	})
	require.NotEmpty(t, routed, "found no case in Model.Update that calls "+routingMethod)

	return routed
}

func (f sourceFile) componentTypes(typeExprs []ast.Expr) []string {
	var types []string

	for _, typeExpr := range typeExprs {
		if name, isComponent := f.componentType(typeExpr); isComponent {
			types = append(types, name)
		}
	}

	return types
}

func callsMethod(statements []ast.Stmt, method string) bool {
	calls := false

	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			calls = calls || (ok && selector.Sel.Name == method)

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
