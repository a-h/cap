package query

import (
	"fmt"
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

// BuildTree builds the downward composition tree rooted at id. An identifier
// already present on the path from the root is marked Repeated and not expanded, so
// the tree stays finite when the graph contains cycles or shared nodes.
func BuildTree(m *model.Model, id model.ID) Node {
	return buildNode(m, id, map[model.ID]struct{}{})
}

func buildNode(m *model.Model, id model.ID, path map[model.ID]struct{}) Node {
	_, resolved := m.Lookup(id)
	node := Node{ID: id, Title: Title(m, id), Resolved: resolved}
	if !resolved {
		return node
	}
	if _, onPath := path[id]; onPath {
		node.Repeated = true
		return node
	}
	path[id] = struct{}{}
	for _, child := range Children(m, id) {
		node.Children = append(node.Children, buildNode(m, child, path))
	}
	delete(path, id)
	return node
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
