//go:build generate

package main

import (
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func main() {
	root, err := os.OpenRoot(".")
	if err != nil {
		panic(err)
	}
	defer root.Close()

	d, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		panic(err)
	}

	for _, f := range d {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".gz" {
			if err := root.Remove(f.Name()); err != nil {
				panic(err)
			}
		}
	}

	d, err = fs.ReadDir(root.FS(), ".")
	if err != nil {
		panic(err)
	}

	for _, f := range d {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".bdf" {
			if err := compress(root, f.Name()); err != nil {
				panic(err)
			}
		}
	}
}

func compress(root *os.Root, path string) error {
	in, err := root.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := root.Create(path + ".gz")
	if err != nil {
		return err
	}
	defer out.Close()

	gzw := gzip.NewWriter(out)
	gzw.Header.ModTime = time.Time{}
	gzw.Header.Name = path

	if _, err := io.Copy(gzw, in); err != nil {
		return err
	}

	if err := gzw.Close(); err != nil {
		return err
	}

	return nil
}
