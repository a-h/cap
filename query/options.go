package query

import "github.com/a-h/cap/model"

// Options controls how a tree, forest, or graph is built. The zero value imposes no
// limit and excludes nothing.
type Options struct {
	// MaxDepth limits how many edges from a root the traversal expands. A value of 0
	// means unlimited. With MaxDepth 1 a root shows only its direct children.
	MaxDepth int
	// Exclude lists the kinds to omit. An excluded entity is neither a node nor a
	// child, and the traversal does not descend into it.
	Exclude map[model.Kind]bool
}

// excluded reports whether the entity of the given identifier is of an excluded kind.
// An unloaded identifier is never excluded, so a dangling reference still shows.
func (o Options) excluded(m *model.Model, id model.ID) bool {
	if len(o.Exclude) == 0 {
		return false
	}
	kind, ok := m.Lookup(id)
	if !ok {
		return false
	}
	return o.Exclude[kind]
}

// children returns the children of an entity with excluded kinds removed.
func (o Options) children(m *model.Model, id model.ID) []model.ID {
	all := Children(m, id)
	if len(o.Exclude) == 0 {
		return all
	}
	out := make([]model.ID, 0, len(all))
	for _, child := range all {
		if o.excluded(m, child) {
			continue
		}
		out = append(out, child)
	}
	return out
}

// reachedDepth reports whether a traversal at the given depth has reached the limit
// and should not expand further.
func (o Options) reachedDepth(depth int) bool {
	return o.MaxDepth > 0 && depth >= o.MaxDepth
}
