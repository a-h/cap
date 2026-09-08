package main

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/a-h/cap/handlers"
)

//go:embed frontend/index.html
var landingPage []byte

// switchableHandler is an http.Handler that delegates to an inner handler,
// and falls back to a landing page when no root folder has been selected.
type switchableHandler struct {
	next atomic.Value // stores http.Handler
}

func (s *switchableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, _ := s.next.Load().(http.Handler)
	if h == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(landingPage)
		return
	}
	h.ServeHTTP(w, r)
}

// SetRoot swaps the handler to serve a new cap model root. If the selected
// directory contains a "cap" subdirectory, that subdirectory is used as the
// model root, matching the CLI default of --root cap.
func (s *switchableHandler) SetRoot(dir string) {
	root := dir
	if sub := filepath.Join(dir, "cap"); dirExists(sub) {
		root = sub
	}
	s.next.Store(handlers.NewMux(root))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
