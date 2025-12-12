// Package parser provides Go source file parsing functionality.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"gogen/internal/model"
)

// Parser parses Go source files and extracts type definitions.
type Parser struct {
	fset *token.FileSet
}

// New creates a new Parser.
func New() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

// ParseFile parses a single Go source file and returns its type definitions.
func (p *Parser) ParseFile(path string) (*model.File, error) {
	file, err := parser.ParseFile(p.fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	result := &model.File{
		Package: file.Name.Name,
		Path:    path,
	}

	// Extract imports
	result.Imports = p.extractImports(file)

	// Extract types using ast.Inspect
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			t := p.extractType(typeSpec, genDecl.Doc)
			result.Types = append(result.Types, t)
		}
		return true
	})

	return result, nil
}

// extractImports extracts import statements from a Go file.
func (p *Parser) extractImports(file *ast.File) []model.Import {
	var imports []model.Import
	for _, imp := range file.Imports {
		i := model.Import{
			Path: strings.Trim(imp.Path.Value, `"`),
		}
		if imp.Name != nil {
			i.Alias = imp.Name.Name
		}
		imports = append(imports, i)
	}
	return imports
}

// extractType extracts type information from an ast.TypeSpec.
func (p *Parser) extractType(spec *ast.TypeSpec, doc *ast.CommentGroup) model.Type {
	t := model.Type{
		Name:       spec.Name.Name,
		IsExported: ast.IsExported(spec.Name.Name),
		Doc:        commentText(doc),
	}

	// Determine type kind and extract details
	switch typeExpr := spec.Type.(type) {
	case *ast.StructType:
		t.Kind = model.KindStruct
		t.Fields = p.extractFields(typeExpr.Fields)

	case *ast.InterfaceType:
		t.Kind = model.KindInterface
		// We don't extract interface methods for now

	case *ast.Ident:
		// Type alias or named type
		if spec.Assign.IsValid() {
			t.Kind = model.KindAlias
		} else {
			t.Kind = model.KindNamed
		}
		t.Underlying = p.typeRefFromExpr(typeExpr)

	default:
		// Other types (slices, maps, etc.) as named types
		if spec.Assign.IsValid() {
			t.Kind = model.KindAlias
		} else {
			t.Kind = model.KindNamed
		}
		t.Underlying = p.typeRefFromExpr(typeExpr)
	}

	return t
}

// extractFields extracts fields from a struct.
func (p *Parser) extractFields(fieldList *ast.FieldList) []model.Field {
	if fieldList == nil {
		return nil
	}

	var fields []model.Field
	for _, f := range fieldList.List {
		typeRef := p.typeRefFromExpr(f.Type)
		tag := p.parseTag(f.Tag)
		doc := commentText(f.Doc)

		if len(f.Names) == 0 {
			// Embedded field
			fields = append(fields, model.Field{
				Name:       typeRef.Name, // Use the type name
				Type:       *typeRef,
				Tag:        tag,
				Doc:        doc,
				IsEmbedded: true,
				IsExported: ast.IsExported(typeRef.Name),
			})
		} else {
			for _, name := range f.Names {
				fields = append(fields, model.Field{
					Name:       name.Name,
					Type:       *typeRef,
					Tag:        tag,
					Doc:        doc,
					IsExported: ast.IsExported(name.Name),
				})
			}
		}
	}
	return fields
}

// typeRefFromExpr converts an ast.Expr to a TypeRef.
func (p *Parser) typeRefFromExpr(expr ast.Expr) *model.TypeRef {
	switch t := expr.(type) {
	case *ast.Ident:
		return &model.TypeRef{
			Kind: model.KindBasic,
			Name: t.Name,
			Raw:  t.Name,
		}

	case *ast.SelectorExpr:
		// Package-qualified type (e.g., time.Time)
		pkg := ""
		if ident, ok := t.X.(*ast.Ident); ok {
			pkg = ident.Name
		}
		return &model.TypeRef{
			Kind:    model.KindNamed,
			Name:    t.Sel.Name,
			Package: pkg,
			Raw:     fmt.Sprintf("%s.%s", pkg, t.Sel.Name),
		}

	case *ast.StarExpr:
		elem := p.typeRefFromExpr(t.X)
		return &model.TypeRef{
			Kind: model.KindPointer,
			Elem: elem,
			Raw:  "*" + elem.Raw,
		}

	case *ast.ArrayType:
		elem := p.typeRefFromExpr(t.Elt)
		if t.Len == nil {
			// Slice
			return &model.TypeRef{
				Kind: model.KindSlice,
				Elem: elem,
				Raw:  "[]" + elem.Raw,
			}
		}
		// Array
		return &model.TypeRef{
			Kind: model.KindArray,
			Elem: elem,
			Raw:  fmt.Sprintf("[...]%s", elem.Raw),
		}

	case *ast.MapType:
		key := p.typeRefFromExpr(t.Key)
		value := p.typeRefFromExpr(t.Value)
		return &model.TypeRef{
			Kind:  model.KindMap,
			Key:   key,
			Value: value,
			Raw:   fmt.Sprintf("map[%s]%s", key.Raw, value.Raw),
		}

	case *ast.InterfaceType:
		return &model.TypeRef{
			Kind: model.KindInterface,
			Name: "interface{}",
			Raw:  "interface{}",
		}

	case *ast.ChanType:
		elem := p.typeRefFromExpr(t.Value)
		return &model.TypeRef{
			Kind: model.KindBasic,
			Name: "chan",
			Elem: elem,
			Raw:  "chan " + elem.Raw,
		}

	case *ast.FuncType:
		return &model.TypeRef{
			Kind: model.KindBasic,
			Name: "func",
			Raw:  "func",
		}

	case *ast.Ellipsis:
		// Variadic parameter, treat as slice
		elem := p.typeRefFromExpr(t.Elt)
		return &model.TypeRef{
			Kind: model.KindSlice,
			Elem: elem,
			Raw:  "..." + elem.Raw,
		}

	default:
		return &model.TypeRef{
			Kind: model.KindBasic,
			Name: "unknown",
			Raw:  "unknown",
		}
	}
}

// parseTag parses a struct tag.
func (p *Parser) parseTag(lit *ast.BasicLit) model.StructTag {
	if lit == nil {
		return model.StructTag{Values: make(map[string]string)}
	}

	raw := strings.Trim(lit.Value, "`")
	tag := reflect.StructTag(raw)

	// Parse common tag keys
	values := make(map[string]string)
	for _, key := range []string{"json", "yaml", "xml", "db", "form", "validate", "binding", "bson"} {
		if v, ok := tag.Lookup(key); ok {
			values[key] = v
		}
	}

	return model.StructTag{
		Raw:    raw,
		Values: values,
	}
}

// commentText extracts text from a comment group.
func commentText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return strings.TrimSpace(cg.Text())
}
