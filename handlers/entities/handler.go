package entities

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

// Handler serves entity detail panels and handles file saves.
type Handler struct {
	Root string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodPut:
		h.Put(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := model.ID(r.PathValue("id")).Canonical()
	res, err := store.Load(h.Root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view, err := newEntityView(res, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	Detail(view).Render(r.Context(), w)
}

func (h Handler) Put(w http.ResponseWriter, r *http.Request) {
	id := model.ID(r.PathValue("id")).Canonical()
	res, err := store.Load(h.Root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path := resolveFilePath(res, id)
	if path == "" {
		http.Error(w, "file not found for "+string(id), http.StatusNotFound)
		return
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = os.WriteFile(path, content, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EntityView holds all data the detail template needs to render.
type EntityView struct {
	ID          string
	Kind        string
	Title       string
	Status      string
	Color       string
	FilePath    string
	FileContent string
	LinksTo     []RelLink
	LinkedBy    []RelLink
}

// RelLink is one entry in the links-to or linked-by section.
type RelLink struct {
	ID    string
	Kind  string
	Title string
	Color string
}

func newEntityView(res store.LoadResult, id model.ID) (EntityView, error) {
	m := res.Model
	kind, title, status := lookupEntity(m, id)
	if kind == "" {
		return EntityView{}, &notFoundError{id: string(id)}
	}
	path := resolveFilePath(res, id)
	var content string
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			content = string(data)
		}
	}
	edges := buildEdges(m)
	var linksTo, linkedBy []RelLink
	for _, tid := range edges.to[id] {
		tk, tt, _ := lookupEntity(m, tid)
		linksTo = append(linksTo, RelLink{ID: string(tid), Kind: tk, Title: tt, Color: kindColor(tk)})
	}
	for _, sid := range edges.from[id] {
		sk, st, _ := lookupEntity(m, sid)
		linkedBy = append(linkedBy, RelLink{ID: string(sid), Kind: sk, Title: st, Color: kindColor(sk)})
	}
	return EntityView{
		ID:          string(id),
		Kind:        kind,
		Title:       title,
		Status:      status,
		Color:       kindColor(kind),
		FilePath:    path,
		FileContent: content,
		LinksTo:     linksTo,
		LinkedBy:    linkedBy,
	}, nil
}

type notFoundError struct{ id string }

func (e *notFoundError) Error() string { return "entity not found: " + e.id }

type edgeIndex struct {
	to   map[model.ID][]model.ID
	from map[model.ID][]model.ID
}

func (idx *edgeIndex) add(s, t model.ID) {
	idx.to[s] = append(idx.to[s], t)
	idx.from[t] = append(idx.from[t], s)
}

func buildEdges(m *model.Model) edgeIndex {
	idx := edgeIndex{
		to:   map[model.ID][]model.ID{},
		from: map[model.ID][]model.ID{},
	}
	for _, cap := range m.Capabilities {
		if cap.Context != "" {
			idx.add(cap.Context, cap.ID)
		}
		for _, id := range cap.Invariants {
			idx.add(cap.ID, id)
		}
		for _, id := range cap.Specifications {
			idx.add(cap.ID, id)
		}
		for _, id := range cap.ADRs {
			idx.add(cap.ID, id)
		}
		for _, id := range cap.Scenarios {
			idx.add(cap.ID, id)
		}
		for _, id := range cap.Verification {
			idx.add(cap.ID, id)
		}
		for _, id := range cap.Tasks {
			idx.add(cap.ID, id)
		}
	}
	for _, con := range m.Concepts {
		if con.Context != "" {
			idx.add(con.Context, con.ID)
		}
	}
	for _, inv := range m.Invariants {
		for _, id := range inv.Capabilities {
			idx.add(inv.ID, id)
		}
	}
	for _, scn := range m.Scenarios {
		for _, id := range scn.Capabilities {
			idx.add(scn.ID, id)
		}
	}
	return idx
}

func lookupEntity(m *model.Model, id model.ID) (kind, title, status string) {
	if e, ok := m.Contexts[id]; ok {
		return "context", e.Name, ""
	}
	if e, ok := m.Concepts[id]; ok {
		return "concept", e.Name, ""
	}
	if e, ok := m.Capabilities[id]; ok {
		return "capability", e.Name, string(e.Status)
	}
	if e, ok := m.Invariants[id]; ok {
		return "invariant", e.Title, ""
	}
	if e, ok := m.Specifications[id]; ok {
		return "specification", e.Title, ""
	}
	if e, ok := m.ADRs[id]; ok {
		return "adr", e.Title, ""
	}
	if e, ok := m.Scenarios[id]; ok {
		return "scenario", e.Name, ""
	}
	if e, ok := m.Verification[id]; ok {
		return "verification", e.Title, ""
	}
	if e, ok := m.Tasks[id]; ok {
		return "task", e.Title, string(e.Status)
	}
	return "", "", ""
}

// resolveFilePath returns the absolute path for an entity's file. For inline
// entities (e.g. cap-0001/spec-1) it falls back to the parent's file.
func resolveFilePath(res store.LoadResult, id model.ID) string {
	if path, ok := res.Files[id]; ok {
		return path
	}
	raw := string(id)
	if idx := strings.LastIndex(raw, "/"); idx >= 0 {
		parent := model.ID(raw[:idx])
		if path, ok := res.Files[parent]; ok {
			return path
		}
	}
	return ""
}

func kindColor(kind string) string {
	colors := map[string]string{
		"context":       "#4a7fd4",
		"capability":    "#3daa6e",
		"concept":       "#d4a535",
		"invariant":     "#e05c5c",
		"scenario":      "#9a6dd4",
		"specification": "#3db8c8",
		"verification":  "#7aad4a",
		"adr":           "#c07840",
		"task":          "#888888",
	}
	if c, ok := colors[kind]; ok {
		return c
	}
	return "#888888"
}

func StatusColor(status string) string {
	colors := map[string]string{
		"done":        "#3daa6e",
		"in-progress": "#5b8af5",
		"proposed":    "#d4a535",
		"draft":       "#7880a8",
	}
	if c, ok := colors[status]; ok {
		return c
	}
	return ""
}
