package handlers

import (
	"net/http"

	"github.com/a-h/cap/handlers/entities"
	"github.com/a-h/cap/handlers/graph"
	"github.com/a-h/cap/handlers/root"
	"github.com/a-h/cap/handlers/tree"
)

// NewMux constructs the HTTP mux for the cap web UI.
func NewMux(modelRoot string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", root.Handler{})
	mux.Handle("/tree", tree.Handler{Root: modelRoot})
	mux.Handle("/entities/{id}", entities.Handler{Root: modelRoot})
	mux.Handle("/graph", graph.Handler{Root: modelRoot})
	mux.Handle("/static/", StaticHandler)
	return mux
}
