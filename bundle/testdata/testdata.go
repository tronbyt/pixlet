package testdata

import "embed"

//go:embed *.tar.gz testapp/*
var FS embed.FS
