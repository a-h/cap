package root

import (
	"net/http"
)

// Handler serves the main UI shell page.
type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.Get(w, r)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	Layout().Render(r.Context(), w)
}
