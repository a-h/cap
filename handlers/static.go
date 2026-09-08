package handlers

import (
	"embed"
	"net/http"
)

//go:embed static
var staticFiles embed.FS

// StaticHandler serves vendored and application static assets.
var StaticHandler = http.FileServerFS(staticFiles)
