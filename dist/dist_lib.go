//go:build lib

package dist

import "embed"

// dummy values not used in library build
var (
	Static embed.FS
	Index  []byte
)
