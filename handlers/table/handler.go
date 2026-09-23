package table

import (
	"net/http"
	"strings"

	"github.com/a-h/cap/graph"
	"github.com/a-h/cap/store"
)

// Handler serves an HTML relationship table for a given from/to entity kind pair.
type Handler struct {
	Root string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fromKind := r.URL.Query().Get("from")
	toKind := r.URL.Query().Get("to")

	res, err := store.Load(h.Root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groups := graph.Build(res.Model).Groups(fromKind, toKind)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = Table(capitalize(fromKind), capitalize(toKind), groups).Render(r.Context(), w)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
