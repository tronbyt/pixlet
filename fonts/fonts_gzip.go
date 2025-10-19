//go:build gzip_fonts

package fonts

import (
	"compress/gzip"
	"embed"
	"io"
)

//go:generate sh -c "gzip -nkf *.bdf"

//go:embed *.bdf.gz
var FS embed.FS

const Ext = ".bdf.gz"

func GetBytes(name string) ([]byte, error) {
	f, err := FS.Open(name + Ext)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(gzr)
	if err != nil {
		return nil, err
	}

	if err := gzr.Close(); err != nil {
		return nil, err
	}

	return b, nil
}
