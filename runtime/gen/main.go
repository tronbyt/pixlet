package main

// Generates starlark bindings for the pixlet/render package.
//
// Also produces widget documentation and extracts example snippets
// that can be run with docs/gen.go to produce images for the widget
// docs.

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/render/animation"
	"golang.org/x/tools/go/packages"
)

//go:embed *.tmpl */*.tmpl
var tmplFS embed.FS

// Given a `reflect.Type` representing a pointer or slice, get the pointed-to or element type.
func decay(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		return t.Elem()
	}

	return t
}

// Given an `interface{}` return a `reflect.Type`, with pointer or slice unwrapped.
func toDecayedType(v any) reflect.Type {
	return decay(reflect.TypeOf(v))
}

type Package struct {
	Name           string
	Directory      string
	ImportPath     string
	HeaderTemplate string
	TypeTemplate   string
	CodePath       string
	DocTemplate    string
	DocPath        string
	GoRootName     string
	GoWidgetName   string
	Types          []reflect.Value
}

// Type defines how to generate code and documentation for type.
type Type struct {
	GoType        string
	DocType       string
	TemplatePath  string
	GenerateField bool
	DefaultValue  string
}

// GeneratedAttr defines a generated "Go to Starlark" attribute.
// This definition is passed to the templating engine.
type GeneratedAttr struct {
	GoName        string
	GoPath        string
	GoType        string
	GoWidgetName  string
	StarlarkName  string
	GenerateField bool
	IsRequired    bool
	IsReadOnly    bool
	DefaultValue  string

	// Template and generated code for handling this attribute.
	Template *template.Template
	Code     string

	// Documentation for this attribute.
	Documentation string
	DocType       string
}

// GeneratedType defines a generated "Go to Starlark" binding type.
// This definition is passed to the templating engine.
type GeneratedType struct {
	GoName            string
	GoNameWithPackage string
	GoRootName        string
	GoWidgetName      string
	Attributes        []*GeneratedAttr
	HasSize           bool
	HasInit           bool
	HasTransform      bool
	Documentation     string
	Examples          []string
}

// Given a `reflect.Value`, return all its fields, including fields of anonymous composed types.
func allFields(val reflect.Value) []reflect.StructField {
	fields := make([]reflect.StructField, 0)
	typ := val.Type()

	for i := range typ.NumField() {
		t := typ.Field(i)
		v := val.Field(i)

		if t.Anonymous && t.Type.Kind() == reflect.Struct {
			fields = append(fields, allFields(v)...)
		} else {
			fields = append(fields, t)
		}
	}

	return fields
}

// Given a `reflect.StructField`, return a `GeneratedAttr` parse its `starlark:` field tag.
func toGeneratedAttribute(typ reflect.Type, field reflect.StructField) (*GeneratedAttr, error) {
	result := &GeneratedAttr{
		GoName:       field.Name,
		GoPath:       field.Name,
		StarlarkName: strings.ToLower(field.Name),
	}

	if field.Name == typ.Name() {
		result.GoPath = typ.Name() + "." + field.Name
	}

	// Fields can be tagged `starlark:"<name>[<param>...]"` to control the attribute name in Starlark.
	//
	// Additional supported flags:
	//   * "required" - field is required on instantiation
	//   * "readonly" - field is read-only, and not passed to constructor
	//
	if tag, ok := field.Tag.Lookup("starlark"); ok {
		attrs := strings.Split(tag, ",")
		if len(attrs) == 0 {
			return nil, fmt.Errorf("%s.%s has invalid tag: '%s'", typ.Name(), field.Name, tag)
		}

		result.StarlarkName = strings.TrimSpace(attrs[0])

		for _, attr := range attrs[1:] {
			attr = strings.TrimSpace(attr)
			switch attr {
			case "required":
				result.IsRequired = true
			case "readonly":
				result.IsReadOnly = true
			default:
				return nil, fmt.Errorf("%s.%s has unsupported tag attribute: '%s'", typ.Name(), field.Name, attr)
			}
		}
	}

	if result.StarlarkName == "" {
		result.StarlarkName = strings.ToLower(field.Name)
	}

	return result, nil
}

func toGeneratedType(pkg Package, val reflect.Value) (*GeneratedType, error) {
	result := &GeneratedType{}

	typ := val.Type()

	if decay(typ) == toDecayedType(new(render.Root)) {
		result.GoRootName = pkg.GoRootName
	}

	if typ.ConvertibleTo(toDecayedType(new(render.Widget))) {
		result.GoWidgetName = pkg.GoWidgetName
	}

	if typ.ConvertibleTo(toDecayedType(new(render.WidgetStaticSize))) {
		result.HasSize = true
	}

	if typ.ConvertibleTo(toDecayedType(new(render.WidgetWithInit))) {
		result.HasInit = true
	}

	if typ.Implements(reflect.TypeFor[animation.Transform]()) {
		result.HasTransform = true
	}

	// Unwrap any pointer types.
	val = reflect.Indirect(val)
	typ = val.Type()

	if val.Kind() != reflect.Struct {
		panic("type is neither struct nor pointer to struct, wtf?")
	}

	result.GoName = typ.Name()
	result.GoNameWithPackage = typ.String()

	for _, field := range allFields(val) {
		if field.PkgPath != "" {
			// Field is not exported.
			continue
		}

		if field.Anonymous {
			if _, hasTag := field.Tag.Lookup("starlark"); !hasTag {
				// Anonymous fields without Starlark metadata are structural.
				continue
			}
		}

		if attribute, err := toGeneratedAttribute(typ, field); err == nil {
			result.Attributes = append(result.Attributes, attribute)

			if t, ok := TypeMap[field.Type]; ok {
				attribute.GoType = t.GoType
				attribute.GoWidgetName = pkg.GoWidgetName
				attribute.DocType = t.DocType
				attribute.Template = loadTemplate(t.TemplatePath)
				attribute.GenerateField = t.GenerateField
				attribute.DefaultValue = t.DefaultValue
			} else {
				return nil, fmt.Errorf("%s.%s has unsupported type", typ.Name(), field.Name)
			}
		} else {
			return nil, err
		}
	}

	// Reorder attributes so that required fields appear first.
	sort.SliceStable(result.Attributes, func(i, j int) bool {
		return result.Attributes[i].IsRequired && !result.Attributes[j].IsRequired
	})

	return result, nil
}

func loadTemplate(path string) *template.Template {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	content := must2(tmplFS.ReadFile(path))

	tmpl := must2(template.New(path).Funcs(funcMap).Parse(string(content)))
	return tmpl
}

func renderTemplateToFile(tmpl *template.Template, data any, path string) {
	outf := must2(os.Create(path))
	defer func() {
		must(outf.Close())
	}()
	must(tmpl.Execute(outf, data))
}

func renderTemplateToBuffer(tmpl *template.Template, data any, buf *bytes.Buffer) {
	must(tmpl.Execute(buf, data))
}

func renderTemplateToString(tmpl *template.Template, data any) string {
	var buf bytes.Buffer
	renderTemplateToBuffer(tmpl, data, &buf)
	return buf.String()
}

func commentText(field *ast.Field) string {
	if field.Doc != nil {
		if text := strings.TrimSpace(field.Doc.Text()); text != "" {
			return text
		}
	}
	if field.Comment != nil {
		return strings.TrimSpace(field.Comment.Text())
	}
	return ""
}

func fieldNameFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.StarExpr:
		return fieldNameFromExpr(e.X)
	default:
		return ""
	}
}

func collectFieldDocs(files []*ast.File) (map[string]map[string]string, map[string][]string) {
	fieldDocs := map[string]map[string]string{}
	embedded := map[string][]string{}

	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if _, ok := fieldDocs[typeSpec.Name.Name]; !ok {
					fieldDocs[typeSpec.Name.Name] = map[string]string{}
				}
				if _, ok := embedded[typeSpec.Name.Name]; !ok {
					embedded[typeSpec.Name.Name] = []string{}
				}

				for _, field := range structType.Fields.List {
					text := commentText(field)

					// Normal named fields; one comment applies to all names in this line.
					if len(field.Names) > 0 {
						if text == "" {
							continue
						}
						for _, name := range field.Names {
							fieldDocs[typeSpec.Name.Name][name.Name] = text
						}
						continue
					}

					// Embedded field.
					if name := fieldNameFromExpr(field.Type); name != "" {
						embedded[typeSpec.Name.Name] = append(embedded[typeSpec.Name.Name], name)
						if text != "" {
							fieldDocs[typeSpec.Name.Name][name] = text
						}
					}
				}
			}
		}
	}

	return fieldDocs, embedded
}

func resolveFieldDoc(fieldDocs map[string]map[string]string, embedded map[string][]string, typeName, fieldName string, visited map[string]bool) string {
	if docs, ok := fieldDocs[typeName]; ok {
		if text := docs[fieldName]; text != "" {
			return text
		}
	}

	if visited[typeName] {
		return ""
	}
	visited[typeName] = true

	for _, embedType := range embedded[typeName] {
		if text := resolveFieldDoc(fieldDocs, embedded, embedType, fieldName, visited); text != "" {
			return text
		}
	}

	return ""
}

func splitDocAndExamples(docText string) (string, []string) {
	docText = strings.ReplaceAll(docText, "\nExample:", "\n")

	var cleanDoc strings.Builder
	cleanDoc.Grow(len(docText))

	examples := make([]string, 0, 2)

	for docText != "" {
		before, after, ok := strings.Cut(docText, "\n\t")

		cleanDoc.WriteString(strings.TrimSpace(before))

		if ok {
			before, after, _ = strings.Cut(after, "\n\n")

			example := strings.ReplaceAll(strings.TrimSpace(before), "\n\t", "\n")
			if example != "" {
				examples = append(examples, example)
			}
		}

		docText = after
	}

	return strings.TrimSpace(cleanDoc.String()), examples
}

func attachDocs(pkg Package, types []*GeneratedType) {
	// Parse all .go files in pixlet/render packages and extract all type doc comments
	fset := token.NewFileSet()
	docs := map[string]string{}

	abs, err := filepath.Abs(pkg.Directory)
	if err != nil {
		panic(err)
	}

	cfg := &packages.Config{
		Mode: packages.LoadFiles | packages.NeedSyntax,
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, abs)
	if err != nil {
		panic(err)
	}
	if len(pkgs) != 1 {
		panic(fmt.Errorf("expected 1 package, got %d", len(pkgs)))
	}
	if len(pkgs[0].Errors) > 0 {
		panic(pkgs[0].Errors[0])
	}

	pkgDoc := must2(doc.NewFromFiles(fset, pkgs[0].Syntax, pkg.ImportPath))
	for _, type_ := range pkgDoc.Types {
		docs[type_.Name] = type_.Doc
	}
	fieldDocs, embedded := collectFieldDocs(pkgs[0].Syntax)

	for _, type_ := range types {
		type_.Documentation, type_.Examples = splitDocAndExamples(docs[type_.GoName])

		// Attribute docs from field comments only.
		for _, attr := range type_.Attributes {
			attr.Documentation = resolveFieldDoc(fieldDocs, embedded, type_.GoName, attr.GoName, map[string]bool{})
		}
	}
}

func generateCode(pkg Package, types []*GeneratedType) {
	// First render templates for each attribute.
	for _, type_ := range types {
		for _, attr := range type_.Attributes {
			attr.Code = renderTemplateToString(attr.Template, attr)
		}
	}

	// Then render templates for the header and for each type.
	headerTmpl := loadTemplate(pkg.HeaderTemplate)
	typeTmpl := loadTemplate(pkg.TypeTemplate)

	outf := must2(os.Create(pkg.CodePath))
	defer func() {
		must(outf.Close())
	}()

	var buf bytes.Buffer
	renderTemplateToBuffer(headerTmpl, types, &buf)

	for _, typ := range types {
		renderTemplateToBuffer(typeTmpl, typ, &buf)
	}

	// Format and write the source to disk.
	source := must2(format.Source(buf.Bytes()))
	must2(outf.Write(source))
}

func generateDocs(pkg Package, types []*GeneratedType) {
	tmpl := loadTemplate(pkg.DocTemplate)
	renderTemplateToFile(tmpl, types, pkg.DocPath)
}

func main() {
	// Generate code and documentation for each package.
	for _, pkg := range Packages {
		types := []*GeneratedType{}

		for _, typ := range pkg.Types {
			if result, err := toGeneratedType(pkg, typ); err == nil {
				types = append(types, result)
			} else {
				panic(err)
			}
		}

		sort.SliceStable(types, func(i, j int) bool {
			return types[i].GoName < types[j].GoName
		})

		attachDocs(pkg, types)
		generateCode(pkg, types)
		generateDocs(pkg, types)
	}

	genEmoji()
}
