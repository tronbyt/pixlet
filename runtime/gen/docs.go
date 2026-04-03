package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func generateDocs(pkg Package, types []*GeneratedType) {
	tmpl := must2(loadTemplate(pkg.DocTemplate))
	renderTemplateToFile(tmpl, types, pkg.DocPath)
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
	docs := make(map[string]string, len(pkgDoc.Types))
	for _, type_ := range pkgDoc.Types {
		docs[type_.Name] = type_.Doc
	}
	fieldDocs, embedded := collectFieldDocs(pkgs[0].Syntax)

	for _, type_ := range types {
		type_.Documentation, type_.Examples = splitDocAndExamples(docs[type_.GoName])

		// Attribute docs from field comments only.
		for _, attr := range type_.Attributes {
			attr.Documentation = resolveFieldDoc(fieldDocs, embedded, type_.GoName, attr.GoName(), map[string]bool{})
		}
	}
}
