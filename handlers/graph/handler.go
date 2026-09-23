package graph

import (
	"encoding/json"
	"net/http"

	graphpkg "github.com/a-h/cap/graph"
	"github.com/a-h/cap/store"
)

// Handler serves graph data as JSON for D3.
type Handler struct {
	Root string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Get(w, r)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	res, err := store.Load(h.Root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphpkg.Build(res.Model))
}
