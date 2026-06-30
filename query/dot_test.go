package query

import (
	"strings"
	"testing"

	"github.com/a-h/cap/model"
)

func hasEdge(g Graph, from, to model.ID) bool {
	for _, e := range g.Edges {
		if e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func hasNode(g Graph, id model.ID) bool {
	for _, n := range g.Nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

func TestBuildGraph(t *testing.T) {
	t.Run("every loaded entity is a node and every downward link an edge", func(t *testing.T) {
		g := BuildGraph(sample(), Options{})
		for _, id := range []model.ID{"ctx-0001", "con-0001", "cap-0003", "inv-0001", "spec-0012", "ver-0008", "scn-0001"} {
			if !hasNode(g, id) {
				t.Errorf("expected a node for %s", id)
			}
		}
		if !hasEdge(g, "ctx-0001", "cap-0003") {
			t.Error("expected an edge from the context to its capability")
		}
		if !hasEdge(g, "scn-0001", "cap-0003") {
			t.Error("expected an edge from the scenario to its capability")
		}
		if !hasEdge(g, "cap-0003", "spec-0012") {
			t.Error("expected an edge from the capability to its specification")
		}
	})

	t.Run("a shared specification is one node with an edge from each capability", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0001"] = model.Capability{ID: "cap-0001", Name: "Publish events"}
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Subscribe to event streams"}
		m.Specifications["spec-0001"] = model.Specification{ID: "spec-0001", Title: "WebSocket protocol", Specifies: []model.ID{"cap-0001", "cap-0003"}}
		g := BuildGraph(m, Options{})
		var count int
		for _, n := range g.Nodes {
			if n.ID == "spec-0001" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected spec-0001 as a single node, got %d", count)
		}
		if !hasEdge(g, "cap-0001", "spec-0001") || !hasEdge(g, "cap-0003", "spec-0001") {
			t.Errorf("expected an edge from each capability to the shared spec, got %v", g.Edges)
		}
	})

	t.Run("a referenced but unloaded identifier is an unresolved node", func(t *testing.T) {
		m := model.NewModel()
		m.Scenarios["scn-0001"] = model.Scenario{ID: "scn-0001", Name: "J", Capabilities: []model.ID{"cap-9999"}}
		g := BuildGraph(m, Options{})
		var found bool
		for _, n := range g.Nodes {
			if n.ID == "cap-9999" {
				found = true
				if n.Resolved {
					t.Error("expected cap-9999 to be unresolved")
				}
			}
		}
		if !found {
			t.Error("expected an unresolved node for the dangling reference")
		}
	})
}

func TestBuildGraphFrom(t *testing.T) {
	t.Run("the subgraph contains only entities reachable downward from the root", func(t *testing.T) {
		g := BuildGraphFrom(sample(), "scn-0001", Options{})
		for _, id := range []model.ID{"scn-0001", "cap-0003", "inv-0001", "spec-0012", "ver-0008"} {
			if !hasNode(g, id) {
				t.Errorf("expected %s in the subgraph", id)
			}
		}
		if hasNode(g, "ctx-0001") {
			t.Error("did not expect the context, which is not reachable downward from the scenario")
		}
		if hasNode(g, "con-0001") {
			t.Error("did not expect a concept, which is not reachable from the scenario")
		}
	})
}

func TestRenderDOT(t *testing.T) {
	t.Run("the output is a directed graph with quoted nodes and edges", func(t *testing.T) {
		out := RenderDOT(BuildGraphFrom(sample(), "scn-0001", Options{}))
		if !strings.HasPrefix(out, "digraph cap {") {
			t.Errorf("expected a digraph header, got:\n%s", out)
		}
		if !strings.Contains(out, `"scn-0001" -> "cap-0003";`) {
			t.Errorf("expected a quoted edge, got:\n%s", out)
		}
		if !strings.Contains(out, `"cap-0003" [label="cap-0003\nEvaluate policies"];`) {
			t.Errorf("expected a labelled node, got:\n%s", out)
		}
	})

	t.Run("a title containing a quote is escaped", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0001"] = model.Capability{ID: "cap-0001", Name: `The "main" thing`}
		out := RenderDOT(BuildGraph(m, Options{}))
		if !strings.Contains(out, `\"main\"`) {
			t.Errorf("expected the quote to be escaped, got:\n%s", out)
		}
	})
}
