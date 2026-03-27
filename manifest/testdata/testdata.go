package testdata

import "embed"

//go:embed *.star *.yaml
var FS embed.FS
