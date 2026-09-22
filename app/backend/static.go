package ojozone

import (
	"embed"
	"io/fs"
)

//go:embed static/*
var embeddedStatic embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(embeddedStatic, "static")
}
