package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/tronbyt/pixlet/manifest"
)

const (
	appsDir      = "apps"
	manifestName = "manifest.yaml"
)

//go:embed templates/source.star.tmpl
var starSource string

// Generator provides a structure for generating apps.
type Generator struct {
	starTmpl   *template.Template
	root       string
	inAppsRepo bool
}

// NewGenerator creates an instantiated generator with the templates parsed.
func NewGenerator(root string, inAppsRepo bool) (*Generator, error) {
	starTmpl, err := template.New("star").Parse(starSource)
	if err != nil {
		return nil, err
	}

	return &Generator{
		starTmpl:   starTmpl,
		root:       root,
		inAppsRepo: inAppsRepo,
	}, nil
}

// GenerateApp creates the base app starlark, go package, and updates the app
// list.
func (g *Generator) GenerateApp(app *manifest.Manifest) (string, error) {
	if g.inAppsRepo {
		if err := g.createDir(app); err != nil {
			return "", err
		}
	}

	err := g.writeManifest(app)
	if err != nil {
		return "", err
	}

	return g.generateStarlark(app)
}

// RemoveApp removes an app from the apps directory.
func (g *Generator) RemoveApp(app *manifest.Manifest) error {
	return g.removeDir(app)
}

func (g *Generator) createDir(app *manifest.Manifest) error {
	p := path.Join(g.root, appsDir, manifest.GenerateDirName(app.Name))
	return os.MkdirAll(p, os.ModePerm)
}

func (g *Generator) removeDir(app *manifest.Manifest) error {
	p := path.Join(g.root, appsDir, manifest.GenerateDirName(app.Name))
	return os.RemoveAll(p)
}

func (g *Generator) writeManifest(app *manifest.Manifest) error {
	p := path.Join(g.root, manifestName)
	if g.inAppsRepo {
		p = path.Join(g.root, appsDir, manifest.GenerateDirName(app.Name), manifestName)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("couldn't create manifest file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return app.WriteManifest(f)
}

func (g *Generator) generateStarlark(app *manifest.Manifest) (string, error) {
	dir := manifest.GenerateDirName(app.Name)
	fn := manifest.GenerateFileName(app.Name)

	p := path.Join(g.root, fn)
	if g.inAppsRepo {
		p = path.Join(g.root, appsDir, dir, fn)
	}

	file, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	err = g.starTmpl.Execute(file, app)
	if err != nil {
		return "", err
	}

	return p, nil
}
