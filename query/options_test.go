package query

import (
	"testing"

	"github.com/a-h/cap/model"
)

func TestOptionsOnTree(t *testing.T) {
	t.Run("a depth limit stops expansion at the given number of links", func(t *testing.T) {
		tree := BuildTree(sample(), "scn-0001", Options{MaxDepth: 1})
		if len(tree.Children) != 1 || tree.Children[0].ID != "cap-0003" {
			t.Fatalf("expected the capability at depth 1, got %#v", tree.Children)
		}
		if len(tree.Children[0].Children) != 0 {
			t.Errorf("expected no grandchildren at depth 1, got %#v", tree.Children[0].Children)
		}
	})

	t.Run("an excluded kind is omitted from the children", func(t *testing.T) {
		tree := BuildTree(sample(), "cap-0003", Options{Exclude: map[model.Kind]bool{model.KindVerification: true}})
		for _, child := range tree.Children {
			if child.ID == "ver-0008" {
				t.Errorf("expected verification to be excluded, got %#v", tree.Children)
			}
		}
	})
}

func TestOptionsOnForest(t *testing.T) {
	t.Run("a root of an excluded kind is omitted", func(t *testing.T) {
		roots := BuildForest(sample(), Options{Exclude: map[model.Kind]bool{model.KindScenario: true}})
		for _, r := range roots {
			if r.ID == "scn-0001" {
				t.Errorf("expected the scenario root to be excluded, got %v", roots)
			}
		}
	})
}

func TestOptionsOnGraph(t *testing.T) {
	t.Run("an excluded kind is neither a node nor an edge endpoint", func(t *testing.T) {
		g := BuildGraph(sample(), Options{Exclude: map[model.Kind]bool{model.KindVerification: true}})
		if hasNode(g, "ver-0008") {
			t.Errorf("expected ver-0008 to be excluded from nodes")
		}
		if hasEdge(g, "cap-0003", "ver-0008") {
			t.Errorf("expected no edge to the excluded verification")
		}
	})

	t.Run("a depth limit measured from the roots trims deep nodes", func(t *testing.T) {
		g := BuildGraph(sample(), Options{MaxDepth: 1})
		if !hasNode(g, "ctx-0001") || !hasNode(g, "cap-0003") {
			t.Error("expected the context and its capability within depth 1")
		}
		if hasNode(g, "inv-0001") {
			t.Error("expected the invariant, two links from the context, to be trimmed at depth 1")
		}
	})

	t.Run("the unfiltered graph still contains an orphaned entity", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["inv-9000"] = model.Invariant{ID: "inv-9000", Title: "An orphan rule."}
		g := BuildGraph(m, Options{})
		if !hasNode(g, "inv-9000") {
			t.Error("expected an orphaned invariant to appear as its own node")
		}
	})
}
