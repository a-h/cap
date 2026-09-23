package graph

import (
	"slices"

	"github.com/a-h/cap/model"
)

// Node is a graph vertex representing one entity.
type Node struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Unresolved bool   `json:"unresolved,omitempty"`
}

// Edge is a directed relationship between two entities.
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Data holds the full graph representation of a model.
type Data struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Group holds a from-entity and all the to-entities it relates to.
type Group struct {
	FromID    string
	FromTitle string
	ToRows    []ToRow
}

// ToRow is one to-entity within a Group.
type ToRow struct {
	ToID    string
	ToTitle string
}

// Groups returns all from→to relationships for the given entity kind pair, sorted
// by from-entity ID with to-rows sorted within each group.
func (d Data) Groups(fromKind, toKind string) []Group {
	titleByID := make(map[string]string, len(d.Nodes))
	kindByID := make(map[string]string, len(d.Nodes))
	for _, n := range d.Nodes {
		titleByID[n.ID] = n.Title
		kindByID[n.ID] = n.Kind
	}

	toByFrom := map[string][]ToRow{}
	var fromOrder []string
	seen := map[string]bool{}
	for _, e := range d.Edges {
		if kindByID[e.Source] != fromKind || kindByID[e.Target] != toKind {
			continue
		}
		if !seen[e.Source] {
			seen[e.Source] = true
			fromOrder = append(fromOrder, e.Source)
		}
		toByFrom[e.Source] = append(toByFrom[e.Source], ToRow{ToID: e.Target, ToTitle: titleByID[e.Target]})
	}

	slices.Sort(fromOrder)
	groups := make([]Group, 0, len(fromOrder))
	for _, fromID := range fromOrder {
		toRows := toByFrom[fromID]
		slices.SortFunc(toRows, func(a, b ToRow) int {
			if a.ToID < b.ToID {
				return -1
			}
			if a.ToID > b.ToID {
				return 1
			}
			return 0
		})
		groups = append(groups, Group{
			FromID:    fromID,
			FromTitle: titleByID[fromID],
			ToRows:    toRows,
		})
	}
	return groups
}

// accumulator builds nodes and edges, guaranteeing that every edge endpoint has a
// corresponding node. References to unknown identifiers produce unresolved nodes so
// D3 always has both endpoints of each edge.
type accumulator struct {
	nodeSet map[string]struct{}
	nodes   []Node
	edgeSet map[string]struct{}
	edges   []Edge
}

func newAccumulator() *accumulator {
	return &accumulator{
		nodeSet: map[string]struct{}{},
		edgeSet: map[string]struct{}{},
	}
}

func (a *accumulator) addNode(id, kind, title string) {
	if _, ok := a.nodeSet[id]; ok {
		return
	}
	a.nodeSet[id] = struct{}{}
	a.nodes = append(a.nodes, Node{ID: id, Kind: kind, Title: title})
}

func (a *accumulator) ensureNode(id string) {
	if _, ok := a.nodeSet[id]; ok {
		return
	}
	a.nodeSet[id] = struct{}{}
	a.nodes = append(a.nodes, Node{ID: id, Kind: "unresolved", Title: id, Unresolved: true})
}

func (a *accumulator) addEdge(source, target string) {
	key := source + "→" + target
	if _, ok := a.edgeSet[key]; ok {
		return
	}
	a.ensureNode(source)
	a.ensureNode(target)
	a.edgeSet[key] = struct{}{}
	a.edges = append(a.edges, Edge{Source: source, Target: target})
}

// Build constructs the full graph data from the given model.
func Build(m *model.Model) Data {
	a := newAccumulator()

	for _, id := range sortedIDs(m.Contexts) {
		n := m.Contexts[id]
		a.addNode(string(id), "context", n.Name)
	}
	for _, id := range sortedIDs(m.Concepts) {
		n := m.Concepts[id]
		a.addNode(string(id), "concept", n.Name)
	}
	for _, id := range sortedIDs(m.Capabilities) {
		n := m.Capabilities[id]
		a.addNode(string(id), "capability", n.Name)
	}
	for _, id := range sortedIDs(m.Invariants) {
		n := m.Invariants[id]
		a.addNode(string(id), "invariant", n.Title)
	}
	for _, id := range sortedIDs(m.Specifications) {
		n := m.Specifications[id]
		a.addNode(string(id), "specification", n.Title)
	}
	for _, id := range sortedIDs(m.Scenarios) {
		n := m.Scenarios[id]
		a.addNode(string(id), "scenario", n.Name)
	}
	for _, id := range sortedIDs(m.Verification) {
		n := m.Verification[id]
		a.addNode(string(id), "verification", n.Title)
	}
	for _, id := range sortedIDs(m.ADRs) {
		n := m.ADRs[id]
		a.addNode(string(id), "adr", n.Title)
	}
	for _, id := range sortedIDs(m.Tasks) {
		n := m.Tasks[id]
		a.addNode(string(id), "task", n.Title)
	}
	for _, id := range sortedIDs(m.Services) {
		n := m.Services[id]
		a.addNode(string(id), "service", n.Name)
	}
	for _, id := range sortedIDs(m.ExternalSystems) {
		n := m.ExternalSystems[id]
		a.addNode(string(id), "external-system", n.Name)
	}
	for _, id := range sortedIDs(m.Requirements) {
		n := m.Requirements[id]
		a.addNode(string(id), "requirement", n.Title)
	}
	for _, id := range sortedIDs(m.Teams) {
		n := m.Teams[id]
		a.addNode(string(id), "team", n.Name)
	}

	for _, id := range sortedIDs(m.Capabilities) {
		cap := m.Capabilities[id]
		if cap.Context != "" {
			a.addEdge(string(cap.Context), string(id))
		}
		for _, x := range cap.Invariants {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.Specifications {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.ADRs {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.Scenarios {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.Verification {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.Tasks {
			a.addEdge(string(id), string(x))
		}
		for _, x := range cap.Requirements {
			a.addEdge(string(x), string(id))
		}
	}
	for _, id := range sortedIDs(m.Concepts) {
		con := m.Concepts[id]
		if con.Context != "" {
			a.addEdge(string(con.Context), string(id))
		}
	}
	for _, id := range sortedIDs(m.Invariants) {
		inv := m.Invariants[id]
		for _, x := range inv.Capabilities {
			a.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Scenarios) {
		scn := m.Scenarios[id]
		for _, x := range scn.Capabilities {
			a.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Services) {
		svc := m.Services[id]
		if svc.Team != "" {
			a.addEdge(string(svc.Team), string(id))
		}
		for _, x := range svc.Capabilities {
			a.addEdge(string(id), string(x))
		}
	}
	for _, id := range sortedIDs(m.Teams) {
		team := m.Teams[id]
		if team.Parent != "" {
			a.addEdge(string(team.Parent), string(id))
		}
	}
	for _, id := range sortedIDs(m.Requirements) {
		req := m.Requirements[id]
		for _, x := range req.Capabilities {
			a.addEdge(string(id), string(x))
		}
	}

	return Data{Nodes: a.nodes, Edges: a.edges}
}

func sortedIDs[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
