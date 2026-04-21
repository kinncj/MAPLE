package main

import "embed"

//go:embed all:template
var embeddedTemplate embed.FS
