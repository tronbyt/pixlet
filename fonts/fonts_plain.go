//go:build !gzip_fonts

package fonts

import "embed"

//go:embed *.bdf
var FS embed.FS

const Ext = ".bdf"

func GetBytes(name string) ([]byte, error) {
	b, err := FS.ReadFile(name + Ext)
	if err != nil {
		return nil, err
	}

	return b, nil
}
