package main

// Generates starlark bindings for the pixlet/render package.
//
// Also produces widget documentation and extracts example snippets
// that can be run with docs/gen.go to produce images for the widget
// docs.

import (
	"bytes"
	"cmp"
	"embed"
	"fmt"
	"go/ast"
	"go/format"
	"iter"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/render/animation"
)

//go:embed *.tmpl */*.tmpl
var tmplFS embed.FS

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
	Package
	Type

	typ   reflect.Type
	field reflect.StructField

	StarlarkName  string
	IsRequired    bool
	IsReadOnly    bool
	Documentation string
}

func (g GeneratedAttr) GoName() string {
	return g.field.Name
}

func (g GeneratedAttr) StarlarkGoName() string {
	return "starlark" + g.GoName()
}

func (g GeneratedAttr) GoPath() string {
	if g.field.Name == g.typ.Name() {
		return g.typ.Name() + "." + g.field.Name
	}
	return g.field.Name
}

func (g GeneratedAttr) Code() (string, error) {
	tmpl, err := loadTemplate(g.TemplatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template for attribute %s: %w", g.StarlarkName, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, g); err != nil {
		return "", fmt.Errorf("failed to render template for attribute %s: %w", g.StarlarkName, err)
	}
	return buf.String(), nil
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
func allFields(val reflect.Value) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		typ := val.Type()
		for i := range typ.NumField() {
			t := typ.Field(i)
			v := val.Field(i)

			if t.Anonymous && t.Type.Kind() == reflect.Struct {
				allFields(v)(func(field reflect.StructField) bool {
					return yield(field)
				})
			} else {
				if !yield(t) {
					return
				}
			}
		}
	}
}

// Given a `reflect.StructField`, return a `GeneratedAttr` parse its `starlark:` field tag.
func newGeneratedAttribute(pkg Package, t Type, typ reflect.Type, field reflect.StructField) (*GeneratedAttr, error) {
	result := &GeneratedAttr{
		Package: pkg,
		Type:    t,
		typ:     typ,
		field:   field,
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
	typ := val.Type()
	result := &GeneratedType{
		HasSize:      typ.Implements(reflect.TypeFor[render.WidgetStaticSize]()),
		HasInit:      typ.Implements(reflect.TypeFor[render.WidgetWithInit]()),
		HasTransform: typ.Implements(reflect.TypeFor[animation.Transform]()),
	}

	if typ == reflect.TypeFor[*render.Root]() {
		result.GoRootName = pkg.GoRootName
	}

	if typ.Implements(reflect.TypeFor[render.Widget]()) {
		result.GoWidgetName = pkg.GoWidgetName
	}

	// Unwrap any pointer types.
	val = reflect.Indirect(val)
	typ = val.Type()

	if val.Kind() != reflect.Struct {
		panic("type is neither struct nor pointer to struct, wtf?")
	}

	result.GoName = typ.Name()
	result.GoNameWithPackage = typ.String()
	result.Attributes = make([]*GeneratedAttr, 0, val.NumField())

	for field := range allFields(val) {
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

		t, ok := TypeMap[field.Type]
		if !ok {
			return nil, fmt.Errorf("%s.%s has unsupported type", typ.Name(), field.Name)
		}

		attribute, err := newGeneratedAttribute(pkg, t, typ, field)
		if err != nil {
			return nil, err
		}

		result.Attributes = append(result.Attributes, attribute)
	}

	// Reorder attributes so that required fields appear first.
	slices.SortStableFunc(result.Attributes, func(a, b *GeneratedAttr) int {
		switch {
		case a.IsRequired == b.IsRequired:
			return 0
		case a.IsRequired:
			return -1
		default:
			return 1
		}
	})

	return result, nil
}

var funcMap = template.FuncMap{
	"ToLower": strings.ToLower,
}

func loadTemplate(p string) (*template.Template, error) {
	return template.New(path.Base(p)).
		Funcs(funcMap).
		ParseFS(tmplFS, p)
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

func generateCode(pkg Package, types []*GeneratedType) {
	// Then render templates for the header and for each type.
	headerTmpl := must2(loadTemplate(pkg.HeaderTemplate))
	typeTmpl := must2(loadTemplate(pkg.TypeTemplate))

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

func main() {
	// Generate code and documentation for each package.
	for _, pkg := range Packages {
		types := make([]*GeneratedType, 0, len(pkg.Types))

		for _, typ := range pkg.Types {
			if result, err := toGeneratedType(pkg, typ); err == nil {
				types = append(types, result)
			} else {
				panic(err)
			}
		}

		slices.SortFunc(types, func(a, b *GeneratedType) int {
			return cmp.Compare(a.GoName, b.GoName)
		})

		attachDocs(pkg, types)
		generateCode(pkg, types)
		generateDocs(pkg, types)
	}

	genEmoji()
}
