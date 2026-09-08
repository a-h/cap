package graph

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/a-h/cap/model"
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
	json.NewEncoder(w).Encode(newGraphData(res.Model))
}

type graphNode struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

type graphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type graphData struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

// edgeSet accumulates directed edges, dropping duplicates.
type edgeSet struct {
	seen  map[string]bool
	edges []graphEdge
}

func newEdgeSet() edgeSet {
	return edgeSet{seen: map[string]bool{}}
}

func (es *edgeSet) add(source, target string) {
	key := source + "→" + target
	if es.seen[key] {
		return
	}
	es.seen[key] = true
	es.edges = append(es.edges, graphEdge{Source: source, Target: target})
}

func newGraphData(m *model.Model) graphData {
	var nodes []graphNode
	for _, id := range sortedIDs(m.Contexts) {
		n := m.Contexts[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "context", Title: n.Name})
	}
	for _, id := range sortedIDs(m.Capabilities) {
		n := m.Capabilities[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "capability", Title: n.Name})
	}
	for _, id := range sortedIDs(m.Concepts) {
		n := m.Concepts[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "concept", Title: n.Name})
	}
	for _, id := range sortedIDs(m.Invariants) {
		n := m.Invariants[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "invariant", Title: n.Title})
	}
	for _, id := range sortedIDs(m.Specifications) {
		n := m.Specifications[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "specification", Title: n.Title})
	}
	for _, id := range sortedIDs(m.Scenarios) {
		n := m.Scenarios[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "scenario", Title: n.Name})
	}
	for _, id := range sortedIDs(m.Verification) {
		n := m.Verification[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "verification", Title: n.Title})
	}
	for _, id := range sortedIDs(m.ADRs) {
		n := m.ADRs[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "adr", Title: n.Title})
	}
	for _, id := range sortedIDs(m.Tasks) {
		n := m.Tasks[id]
		nodes = append(nodes, graphNode{ID: string(id), Kind: "task", Title: n.Title})
	}

	es := newEdgeSet()
	for _, id := range sortedIDs(m.Capabilities) {
		cap := m.Capabilities[id]
		if cap.Context != "" {
			es.add(string(cap.Context), string(id))
		}
		for _, x := range cap.Invariants {
			es.add(string(id), string(x))
		}
		for _, x := range cap.Specifications {
			es.add(string(id), string(x))
		}
		for _, x := range cap.ADRs {
			es.add(string(id), string(x))
		}
		for _, x := range cap.Scenarios {
			es.add(string(id), string(x))
		}
		for _, x := range cap.Verification {
			es.add(string(id), string(x))
		}
		for _, x := range cap.Tasks {
			es.add(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Concepts) {
		con := m.Concepts[id]
		if con.Context != "" {
			es.add(string(con.Context), string(id))
		}
	}
	for _, id := range sortedIDs(m.Invariants) {
		inv := m.Invariants[id]
		for _, x := range inv.Capabilities {
			es.add(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Scenarios) {
		scn := m.Scenarios[id]
		for _, x := range scn.Capabilities {
			es.add(string(id), string(x))
		}
	}
	return graphData{Nodes: nodes, Edges: es.edges}
}

func sortedIDs[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
