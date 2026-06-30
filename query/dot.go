package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/a-h/cap/model"
)

// Edge is a directed composition link from one entity to another, in the downward
// direction returned by Children.
type Edge struct {
	From model.ID `json:"from"`
	To   model.ID `json:"to"`
}

// Graph is the whole model, or a reachable part of it, as a set of nodes and the
// directed composition edges between them. Unlike a tree, each entity is a single
// node however many other entities link to it, and a cycle is a set of edges rather
// than a truncated path. Nodes and edges are ordered by identifier.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// BuildGraph builds the composition graph of the whole model: every loaded entity is
// a node, and every downward link an edge, with the given depth limit and kind
// exclusions applied. Depth is measured from the forest roots: the bounded contexts
// and the top-level entities the text forest starts from. Any entity not reached from
// those roots, such as an orphaned invariant, is then seeded as its own root so the
// graph contains every loaded entity. When a depth limit is set the orphan seeding is
// skipped, since a depth limit asks for a trimmed view measured from the roots. A
// referenced but unloaded identifier becomes an unresolved node so a dangling edge
// still has an endpoint.
func BuildGraph(m *model.Model, opts Options) Graph {
	var roots []model.ID
	roots = append(roots, sortedIDs(contextIDs(m))...)
	roots = append(roots, sortedIDs(topLevelIDs(m))...)
	if opts.MaxDepth == 0 {
		roots = append(roots, allIDs(m)...)
	}
	return buildGraph(m, roots, opts)
}

// BuildGraphFrom builds the subgraph reachable downward from the given entity,
// following composition edges, with the given depth limit and kind exclusions applied.
// The entity itself is included; identifiers it does not reach are omitted.
func BuildGraphFrom(m *model.Model, id model.ID, opts Options) Graph {
	return buildGraph(m, []model.ID{id}, opts)
}

// buildGraph assembles the graph reachable from the given root identifiers by a
// breadth-first walk over filtered children, so the depth limit counts edges from the
// nearest root. Unloaded identifiers are added as unresolved nodes so every edge has
// both endpoints.
func buildGraph(m *model.Model, roots []model.ID, opts Options) Graph {
	nodes := map[model.ID]Node{}
	addNode := func(id model.ID) {
		if _, ok := nodes[id]; ok {
			return
		}
		_, resolved := m.Lookup(id)
		nodes[id] = Node{ID: id, Title: Title(m, id), Resolved: resolved}
	}
	seenEdge := map[Edge]struct{}{}
	var edges []Edge

	type queued struct {
		id    model.ID
		depth int
	}
	var queue []queued
	visited := map[model.ID]struct{}{}
	enqueue := func(id model.ID, depth int) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}
		queue = append(queue, queued{id: id, depth: depth})
	}
	for _, id := range roots {
		if opts.excluded(m, id) {
			continue
		}
		enqueue(id, 0)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		addNode(cur.id)
		if opts.reachedDepth(cur.depth) {
			continue
		}
		for _, child := range opts.children(m, cur.id) {
			addNode(child)
			edge := Edge{From: cur.id, To: child}
			if _, ok := seenEdge[edge]; !ok {
				seenEdge[edge] = struct{}{}
				edges = append(edges, edge)
			}
			enqueue(child, cur.depth+1)
		}
	}

	out := Graph{Nodes: make([]Node, 0, len(nodes)), Edges: edges}
	for _, n := range nodes {
		out.Nodes = append(out.Nodes, n)
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].ID < out.Nodes[j].ID })
	sort.Slice(out.Edges, func(i, j int) bool {
		if out.Edges[i].From != out.Edges[j].From {
			return out.Edges[i].From < out.Edges[j].From
		}
		return out.Edges[i].To < out.Edges[j].To
	})
	return out
}

// RenderDOT produces a Graphviz DOT description of the graph: one node statement per
// entity, labelled with its identifier and title, and one edge statement per link.
// The output is a directed graph named cap, suitable for piping to dot.
func RenderDOT(g Graph) string {
	var b strings.Builder
	b.WriteString("digraph cap {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  node [shape=box];\n")
	for _, n := range g.Nodes {
		fmt.Fprintf(&b, "  %s [label=%s];\n", encodeDOT(string(n.ID)), encodeDOT(dotLabel(n)))
	}
	for _, e := range g.Edges {
		fmt.Fprintf(&b, "  %s -> %s;\n", encodeDOT(string(e.From)), encodeDOT(string(e.To)))
	}
	b.WriteString("}\n")
	return b.String()
}

// dotLabel returns the node label: the identifier and title on one line, or the
// identifier marked unresolved when the entity was referenced but not loaded.
func dotLabel(n Node) string {
	if !n.Resolved {
		return fmt.Sprintf("%s (unresolved)", n.ID)
	}
	if n.Title == "" {
		return string(n.ID)
	}
	return fmt.Sprintf("%s\n%s", n.ID, n.Title)
}

// encodeDOT encodes a value as a DOT quoted string: it escapes the characters that are
// special within such a string and wraps the result in double quotes.
func encodeDOT(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return `"` + s + `"`
}
