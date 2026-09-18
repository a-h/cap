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
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Unresolved bool   `json:"unresolved,omitempty"`
}

type graphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type graphData struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

// graph accumulates nodes and edges, guaranteeing that every edge endpoint has a
// corresponding node. References to unknown identifiers produce unresolved nodes so
// D3 always has both endpoints of each edge.
type graph struct {
	nodeSet map[string]struct{}
	nodes   []graphNode
	edgeSet map[string]struct{}
	edges   []graphEdge
}

func newGraph() *graph {
	return &graph{
		nodeSet: map[string]struct{}{},
		edgeSet: map[string]struct{}{},
	}
}

func (g *graph) addNode(id, kind, title string) {
	if _, ok := g.nodeSet[id]; ok {
		return
	}
	g.nodeSet[id] = struct{}{}
	g.nodes = append(g.nodes, graphNode{ID: id, Kind: kind, Title: title})
}

func (g *graph) ensureNode(id string) {
	if _, ok := g.nodeSet[id]; ok {
		return
	}
	g.nodeSet[id] = struct{}{}
	g.nodes = append(g.nodes, graphNode{ID: id, Kind: "unresolved", Title: id, Unresolved: true})
}

func (g *graph) addEdge(source, target string) {
	key := source + "→" + target
	if _, ok := g.edgeSet[key]; ok {
		return
	}
	g.ensureNode(source)
	g.ensureNode(target)
	g.edgeSet[key] = struct{}{}
	g.edges = append(g.edges, graphEdge{Source: source, Target: target})
}

func newGraphData(m *model.Model) graphData {
	g := newGraph()

	for _, id := range sortedIDs(m.Contexts) {
		n := m.Contexts[id]
		g.addNode(string(id), "context", n.Name)
	}
	for _, id := range sortedIDs(m.Concepts) {
		n := m.Concepts[id]
		g.addNode(string(id), "concept", n.Name)
	}
	for _, id := range sortedIDs(m.Capabilities) {
		n := m.Capabilities[id]
		g.addNode(string(id), "capability", n.Name)
	}
	for _, id := range sortedIDs(m.Invariants) {
		n := m.Invariants[id]
		g.addNode(string(id), "invariant", n.Title)
	}
	for _, id := range sortedIDs(m.Specifications) {
		n := m.Specifications[id]
		g.addNode(string(id), "specification", n.Title)
	}
	for _, id := range sortedIDs(m.Scenarios) {
		n := m.Scenarios[id]
		g.addNode(string(id), "scenario", n.Name)
	}
	for _, id := range sortedIDs(m.Verification) {
		n := m.Verification[id]
		g.addNode(string(id), "verification", n.Title)
	}
	for _, id := range sortedIDs(m.ADRs) {
		n := m.ADRs[id]
		g.addNode(string(id), "adr", n.Title)
	}
	for _, id := range sortedIDs(m.Tasks) {
		n := m.Tasks[id]
		g.addNode(string(id), "task", n.Title)
	}
	for _, id := range sortedIDs(m.Services) {
		n := m.Services[id]
		g.addNode(string(id), "service", n.Name)
	}
	for _, id := range sortedIDs(m.ExternalSystems) {
		n := m.ExternalSystems[id]
		g.addNode(string(id), "external-system", n.Name)
	}
	for _, id := range sortedIDs(m.Requirements) {
		n := m.Requirements[id]
		g.addNode(string(id), "requirement", n.Title)
	}

	for _, id := range sortedIDs(m.Capabilities) {
		cap := m.Capabilities[id]
		if cap.Context != "" {
			g.addEdge(string(cap.Context), string(id))
		}
		for _, x := range cap.Invariants {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.Specifications {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.ADRs {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.Scenarios {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.Verification {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.Tasks {
			g.addEdge(string(id), string(x))
		}
		for _, x := range cap.Requirements {
			g.addEdge(string(x), string(id))
		}
	}
	for _, id := range sortedIDs(m.Concepts) {
		con := m.Concepts[id]
		if con.Context != "" {
			g.addEdge(string(con.Context), string(id))
		}
	}
	for _, id := range sortedIDs(m.Invariants) {
		inv := m.Invariants[id]
		for _, x := range inv.Capabilities {
			g.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Scenarios) {
		scn := m.Scenarios[id]
		for _, x := range scn.Capabilities {
			g.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Services) {
		svc := m.Services[id]
		for _, x := range svc.Capabilities {
			g.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Requirements) {
		req := m.Requirements[id]
		for _, x := range req.Capabilities {
			g.addEdge(string(id), string(x))
		}
	}

	return graphData{Nodes: g.nodes, Edges: g.edges}
}

func sortedIDs[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
