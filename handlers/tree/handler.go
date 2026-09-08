package tree

import (
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

// Handler serves the tree panel as an HTMX fragment.
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
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	activeKinds := parseKinds(r.URL.Query().Get("kinds"))
	rows := newRows(res.Model, q, activeKinds)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	Tree(rows).Render(r.Context(), w)
}

var defaultKinds = map[string]bool{
	"context": true, "capability": true, "concept": true, "invariant": true,
	"scenario": true, "specification": true, "verification": true, "adr": true, "task": true,
}

func parseKinds(raw string) map[string]bool {
	if raw == "" {
		return defaultKinds
	}
	out := map[string]bool{}
	for _, k := range strings.Split(raw, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			out[k] = true
		}
	}
	return out
}

// Row is a single flat row in the tree view.
type Row struct {
	ID      string
	Kind    string
	Title   string
	Status  string
	Color   string
	Depth   int
	Parent  string
	HasKids bool
	Shared  bool
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

func statusColor(status string) string {
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

func newRows(m *model.Model, q string, activeKinds map[string]bool) []Row {
	var rows []Row

	ctxIDs := sortedIDs(m.Contexts)
	for _, ctxID := range ctxIDs {
		ctx := m.Contexts[ctxID]
		ctxRow := Row{
			ID:    string(ctxID),
			Kind:  "context",
			Title: ctx.Name,
			Color: kindColor("context"),
			Depth: 0,
		}
		var children []Row

		// Concepts in this context.
		conIDs := sortedIDs(m.Concepts)
		for _, conID := range conIDs {
			con := m.Concepts[conID]
			if con.Context != ctxID {
				continue
			}
			children = append(children, Row{
				ID:     string(conID),
				Kind:   "concept",
				Title:  con.Name,
				Color:  kindColor("concept"),
				Depth:  1,
				Parent: string(ctxID),
			})
		}

		// Capabilities in this context.
		capIDs := sortedIDs(m.Capabilities)
		for _, capID := range capIDs {
			cap := m.Capabilities[capID]
			if cap.Context != ctxID {
				continue
			}
			capRow := Row{
				ID:     string(capID),
				Kind:   "capability",
				Title:  cap.Name,
				Status: string(cap.Status),
				Color:  kindColor("capability"),
				Depth:  1,
				Parent: string(ctxID),
			}
			var capKids []Row
			for _, invID := range cap.Invariants {
				inv := m.Invariants[invID]
				capKids = append(capKids, Row{
					ID:     string(invID),
					Kind:   "invariant",
					Title:  inv.Title,
					Color:  kindColor("invariant"),
					Depth:  2,
					Parent: string(capID),
				})
			}
			for _, specID := range cap.Specifications {
				spec := m.Specifications[specID]
				capKids = append(capKids, Row{
					ID:     string(specID),
					Kind:   "specification",
					Title:  spec.Title,
					Color:  kindColor("specification"),
					Depth:  2,
					Parent: string(capID),
				})
			}
			for _, adrID := range cap.ADRs {
				adr := m.ADRs[adrID]
				capKids = append(capKids, Row{
					ID:     string(adrID),
					Kind:   "adr",
					Title:  adr.Title,
					Color:  kindColor("adr"),
					Depth:  2,
					Parent: string(capID),
				})
			}
			for _, verID := range cap.Verification {
				ver := m.Verification[verID]
				capKids = append(capKids, Row{
					ID:     string(verID),
					Kind:   "verification",
					Title:  ver.Title,
					Color:  kindColor("verification"),
					Depth:  2,
					Parent: string(capID),
				})
			}
			for _, taskID := range cap.Tasks {
				task := m.Tasks[taskID]
				capKids = append(capKids, Row{
					ID:     string(taskID),
					Kind:   "task",
					Title:  task.Title,
					Status: string(task.Status),
					Color:  kindColor("task"),
					Depth:  2,
					Parent: string(capID),
				})
			}
			capRow.HasKids = len(capKids) > 0
			children = append(children, capRow)
			children = append(children, capKids...)
		}

		ctxRow.HasKids = len(children) > 0
		if !activeKinds["context"] {
			continue
		}
		if q != "" && !rowMatchesQuery(ctxRow, q) && !anyChildMatches(children, q) {
			continue
		}
		rows = append(rows, ctxRow)
		rows = append(rows, children...)
	}

	// Scenarios as top-level items.
	if activeKinds["scenario"] {
		scnIDs := sortedIDs(m.Scenarios)
		for _, scnID := range scnIDs {
			scn := m.Scenarios[scnID]
			scnRow := Row{
				ID:    string(scnID),
				Kind:  "scenario",
				Title: scn.Name,
				Color: kindColor("scenario"),
				Depth: 0,
			}
			var scnKids []Row
			for _, capID := range scn.Capabilities {
				cap := m.Capabilities[capID]
				scnKids = append(scnKids, Row{
					ID:     string(capID),
					Kind:   "capability",
					Title:  cap.Name,
					Status: string(cap.Status),
					Color:  kindColor("capability"),
					Depth:  1,
					Parent: string(scnID),
					Shared: true,
				})
			}
			scnRow.HasKids = len(scnKids) > 0
			if q != "" && !rowMatchesQuery(scnRow, q) && !anyChildMatches(scnKids, q) {
				continue
			}
			rows = append(rows, scnRow)
			rows = append(rows, scnKids...)
		}
	}

	// When a parent kind is not active, promote its children to top-level so
	// they remain visible (e.g. soloing "concept" shows all concepts).
	if !activeKinds["context"] {
		if activeKinds["concept"] {
			for _, id := range sortedIDs(m.Concepts) {
				con := m.Concepts[id]
				rows = append(rows, Row{ID: string(id), Kind: "concept", Title: con.Name, Color: kindColor("concept")})
			}
		}
		if activeKinds["capability"] {
			for _, id := range sortedIDs(m.Capabilities) {
				cap := m.Capabilities[id]
				rows = append(rows, Row{ID: string(id), Kind: "capability", Title: cap.Name, Status: string(cap.Status), Color: kindColor("capability")})
			}
		}
	}
	if !activeKinds["capability"] {
		if activeKinds["invariant"] {
			for _, id := range sortedIDs(m.Invariants) {
				inv := m.Invariants[id]
				rows = append(rows, Row{ID: string(id), Kind: "invariant", Title: inv.Title, Color: kindColor("invariant")})
			}
		}
		if activeKinds["specification"] {
			for _, id := range sortedIDs(m.Specifications) {
				spec := m.Specifications[id]
				rows = append(rows, Row{ID: string(id), Kind: "specification", Title: spec.Title, Color: kindColor("specification")})
			}
		}
		if activeKinds["adr"] {
			for _, id := range sortedIDs(m.ADRs) {
				adr := m.ADRs[id]
				rows = append(rows, Row{ID: string(id), Kind: "adr", Title: adr.Title, Color: kindColor("adr")})
			}
		}
		if activeKinds["verification"] {
			for _, id := range sortedIDs(m.Verification) {
				ver := m.Verification[id]
				rows = append(rows, Row{ID: string(id), Kind: "verification", Title: ver.Title, Color: kindColor("verification")})
			}
		}
		if activeKinds["task"] {
			for _, id := range sortedIDs(m.Tasks) {
				task := m.Tasks[id]
				rows = append(rows, Row{ID: string(id), Kind: "task", Title: task.Title, Status: string(task.Status), Color: kindColor("task")})
			}
		}
	}

	return filterByQuery(rows, q)
}

func rowMatchesQuery(r Row, q string) bool {
	return strings.Contains(strings.ToLower(r.ID), q) ||
		strings.Contains(strings.ToLower(r.Title), q)
}

func anyChildMatches(children []Row, q string) bool {
	if q == "" {
		return false
	}
	for _, c := range children {
		if rowMatchesQuery(c, q) {
			return true
		}
	}
	return false
}

func filterByQuery(rows []Row, q string) []Row {
	if q == "" {
		return rows
	}
	out := make([]Row, 0, len(rows))
	seen := map[string]bool{}
	for _, r := range rows {
		if !rowMatchesQuery(r, q) || seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		r.Parent = ""
		r.Depth = 0
		r.HasKids = false
		out = append(out, r)
	}
	return out
}

func sortedIDs[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// StatusColor returns a CSS color for a status value.
func StatusColor(status string) string {
	return statusColor(status)
}
