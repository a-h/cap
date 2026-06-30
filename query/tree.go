package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/a-h/cap/model"
)

// Node is an entity and its children in a downward composition tree. Resolved is
// false when the identifier was referenced but not loaded, and Repeated is true
// when the identifier already appeared higher on the current path and so was not
// expanded again.
type Node struct {
	ID       model.ID `json:"id"`
	Title    string   `json:"title"`
	Resolved bool     `json:"resolved"`
	Repeated bool     `json:"repeated,omitempty"`
	Children []Node   `json:"children,omitempty"`
}

// BuildTree builds the downward composition tree rooted at id, applying the given
// depth limit and kind exclusions. An identifier already present on the path from the
// root is marked Repeated and not expanded, so the tree stays finite when the graph
// contains cycles or shared nodes.
func BuildTree(m *model.Model, id model.ID, opts Options) Node {
	return buildNode(m, id, 0, opts, map[model.ID]struct{}{})
}

func buildNode(m *model.Model, id model.ID, depth int, opts Options, path map[model.ID]struct{}) Node {
	_, resolved := m.Lookup(id)
	node := Node{ID: id, Title: Title(m, id), Resolved: resolved}
	if !resolved {
		return node
	}
	if _, onPath := path[id]; onPath {
		node.Repeated = true
		return node
	}
	if opts.reachedDepth(depth) {
		return node
	}
	path[id] = struct{}{}
	for _, child := range opts.children(m, id) {
		node.Children = append(node.Children, buildNode(m, child, depth+1, opts, path))
	}
	delete(path, id)
	return node
}

// BuildForest builds the downward composition trees for the whole model, applying the
// given depth limit and kind exclusions. The bounded contexts come first, each
// expanded into the concepts and capabilities it groups, followed by the remaining
// top-level entities: those that nothing links to and that no context already showed,
// such as scenarios and orphaned capabilities. A root of an excluded kind is omitted
// entirely. Roots within each group are ordered by identifier.
func BuildForest(m *model.Model, opts Options) []Node {
	covered := map[model.ID]struct{}{}

	var roots []Node
	for _, id := range sortedIDs(contextIDs(m)) {
		if opts.excluded(m, id) {
			continue
		}
		node := buildNode(m, id, 0, opts, map[model.ID]struct{}{})
		markCovered(node, covered)
		roots = append(roots, node)
	}

	for _, id := range sortedIDs(topLevelIDs(m)) {
		if _, ok := covered[id]; ok {
			continue
		}
		if opts.excluded(m, id) {
			continue
		}
		node := buildNode(m, id, 0, opts, map[model.ID]struct{}{})
		markCovered(node, covered)
		roots = append(roots, node)
	}
	return roots
}

// markCovered records the identifier of a node and all of its descendants, so an
// entity shown beneath a context or earlier root is not repeated as a top-level root.
func markCovered(n Node, covered map[model.ID]struct{}) {
	covered[n.ID] = struct{}{}
	for _, child := range n.Children {
		markCovered(child, covered)
	}
}

// contextIDs returns the identifiers of every loaded bounded context.
func contextIDs(m *model.Model) []model.ID {
	out := make([]model.ID, 0, len(m.Contexts))
	for id := range m.Contexts {
		out = append(out, id)
	}
	return out
}

// allIDs returns the identifiers of every loaded entity, ordered by identifier.
func allIDs(m *model.Model) []model.ID {
	var out []model.ID
	for id := range m.Contexts {
		out = append(out, id)
	}
	for id := range m.Concepts {
		out = append(out, id)
	}
	for id := range m.Capabilities {
		out = append(out, id)
	}
	for id := range m.Invariants {
		out = append(out, id)
	}
	for id := range m.Specifications {
		out = append(out, id)
	}
	for id := range m.ADRs {
		out = append(out, id)
	}
	for id := range m.Scenarios {
		out = append(out, id)
	}
	for id := range m.Verification {
		out = append(out, id)
	}
	for id := range m.Tasks {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// topLevelIDs returns the identifiers of entities that no other entity links to and
// so sit at the top of the composition graph, such as scenarios and capabilities
// that declare no context.
func topLevelIDs(m *model.Model) []model.ID {
	var out []model.ID
	for id := range m.Scenarios {
		if len(Parents(m, id)) == 0 {
			out = append(out, id)
		}
	}
	for id := range m.Capabilities {
		if len(Parents(m, id)) == 0 {
			out = append(out, id)
		}
	}
	return out
}

func sortedIDs(ids []model.ID) []model.ID {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// RenderForest produces an indented text tree for each root, separated by a blank
// line.
func RenderForest(roots []Node) string {
	var b strings.Builder
	for i, root := range roots {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(root.Render())
	}
	return b.String()
}

// Render produces an indented text tree for a node.
func (n Node) Render() string {
	var b strings.Builder
	b.WriteString(n.label())
	b.WriteByte('\n')
	renderChildren(&b, n.Children, "")
	return b.String()
}

func renderChildren(b *strings.Builder, children []Node, prefix string) {
	for i, child := range children {
		last := i == len(children)-1
		branch := "├─ "
		nextPrefix := prefix + "│  "
		if last {
			branch = "└─ "
			nextPrefix = prefix + "   "
		}
		fmt.Fprintf(b, "%s%s%s\n", prefix, branch, child.label())
		renderChildren(b, child.Children, nextPrefix)
	}
}

// label returns the single-line description of a node.
func (n Node) label() string {
	switch {
	case !n.Resolved:
		return fmt.Sprintf("%s  (unresolved)", n.ID)
	case n.Repeated:
		return fmt.Sprintf("%s  %s (shown above)", n.ID, n.Title)
	default:
		return fmt.Sprintf("%s  %s", n.ID, n.Title)
	}
}
