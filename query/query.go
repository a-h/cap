// Package query provides read and navigation operations over a loaded model.
//
// It concentrates the per-kind title and link logic in one place so the list,
// show, and graph commands do not reintroduce per-entity accessor methods.
package query

import (
	"slices"
	"sort"

	"github.com/a-h/cap/model"
)

// Entry is an entity reduced to its identifier and display title, for listings.
// Inline is true for an entity written within a capability rather than its own file.
type Entry struct {
	ID     model.ID `json:"id"`
	Title  string   `json:"title"`
	Inline bool     `json:"inline,omitempty"`
}

// Title returns the display title of an entity regardless of kind. Contexts,
// capabilities, and scenarios use a name; the other kinds use a title. It returns an
// empty string when the identifier is not loaded.
func Title(m *model.Model, id model.ID) string {
	if c, ok := m.Contexts[id]; ok {
		return c.Name
	}
	if c, ok := m.Capabilities[id]; ok {
		return c.Name
	}
	if inv, ok := m.Invariants[id]; ok {
		return inv.Title
	}
	if s, ok := m.Specifications[id]; ok {
		return s.Title
	}
	if a, ok := m.ADRs[id]; ok {
		return a.Title
	}
	if s, ok := m.Scenarios[id]; ok {
		return s.Name
	}
	if v, ok := m.Verification[id]; ok {
		return verificationTitle(v)
	}
	if t, ok := m.Tasks[id]; ok {
		return t.Title
	}
	return ""
}

// List returns the entities of a kind as id and title entries, ordered by
// identifier.
func List(m *model.Model, kind model.Kind) []Entry {
	var entries []Entry
	switch kind {
	case model.KindContext:
		for id, c := range m.Contexts {
			entries = append(entries, Entry{ID: id, Title: c.Name})
		}
	case model.KindCapability:
		for id, c := range m.Capabilities {
			entries = append(entries, Entry{ID: id, Title: c.Name})
		}
	case model.KindInvariant:
		for id, inv := range m.Invariants {
			entries = append(entries, Entry{ID: id, Title: inv.Title, Inline: inv.Inline()})
		}
	case model.KindSpecification:
		for id, s := range m.Specifications {
			entries = append(entries, Entry{ID: id, Title: s.Title, Inline: s.Inline()})
		}
	case model.KindADR:
		for id, a := range m.ADRs {
			entries = append(entries, Entry{ID: id, Title: a.Title})
		}
	case model.KindScenario:
		for id, j := range m.Scenarios {
			entries = append(entries, Entry{ID: id, Title: j.Name})
		}
	case model.KindVerification:
		for id, v := range m.Verification {
			entries = append(entries, Entry{ID: id, Title: verificationTitle(v), Inline: v.Inline()})
		}
	case model.KindTask:
		for id, t := range m.Tasks {
			entries = append(entries, Entry{ID: id, Title: t.Title})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries
}

// Children returns the identifiers an entity links to, in the downward
// composition direction: a scenario is composed of capabilities; a capability is
// composed of its invariants, specifications, ADRs, verification, and tasks; a
// context lists the capabilities it groups. The capability-to-invariant link may
// be declared from either end, so a capability's children include invariants that
// name it as well as those it names.
func Children(m *model.Model, id model.ID) []model.ID {
	if c, ok := m.Capabilities[id]; ok {
		out := dedupe(append(c.Invariants, getInvariantsNaming(m, id)...))
		out = append(out, c.Specifications...)
		out = append(out, c.ADRs...)
		out = append(out, c.Verification...)
		out = append(out, c.Tasks...)
		return out
	}
	if s, ok := m.Scenarios[id]; ok {
		return append([]model.ID(nil), s.Capabilities...)
	}
	if _, ok := m.Contexts[id]; ok {
		return getCapabilitiesInContext(m, id)
	}
	return nil
}

// Parents returns the identifiers that link to an entity: the context a capability
// belongs to, the scenarios that use a capability, the capabilities an invariant
// constrains, and the capability a specification specifies. The result is ordered
// by identifier.
func Parents(m *model.Model, id model.ID) []model.ID {
	seen := map[model.ID]struct{}{}
	add := func(p model.ID) { seen[p] = struct{}{} }
	if c, ok := m.Capabilities[id]; ok {
		if c.Context != "" {
			add(c.Context)
		}
	}
	for jid, j := range m.Scenarios {
		if slices.Contains(j.Capabilities, id) {
			add(jid)
		}
	}
	for cid, c := range m.Capabilities {
		if capabilityLinksTo(c, id) {
			add(cid)
		}
	}
	if inv, ok := m.Invariants[id]; ok {
		for _, cid := range inv.Capabilities {
			add(cid)
		}
	}
	for sid, s := range m.Specifications {
		if s.Of == id {
			add(sid)
		}
	}
	out := make([]model.ID, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// getInvariantsNaming returns the invariants that name the given capability, so the
// many-to-many link is visible whether declared on the capability or the invariant.
func getInvariantsNaming(m *model.Model, capability model.ID) []model.ID {
	var out []model.ID
	for id, inv := range m.Invariants {
		if slices.Contains(inv.Capabilities, capability) {
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// getCapabilitiesInContext returns the identifiers of capabilities that declare the
// given context, ordered by identifier.
func getCapabilitiesInContext(m *model.Model, ctx model.ID) []model.ID {
	var out []model.ID
	for id, c := range m.Capabilities {
		if c.Context == ctx {
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// capabilityLinksTo reports whether the capability references the given identifier in
// any of its link sections.
func capabilityLinksTo(c model.Capability, id model.ID) bool {
	for _, set := range [][]model.ID{c.Invariants, c.Specifications, c.ADRs, c.Scenarios, c.Verification, c.Tasks} {
		if slices.Contains(set, id) {
			return true
		}
	}
	return false
}

func dedupe(ids []model.ID) []model.ID {
	seen := map[model.ID]struct{}{}
	out := make([]model.ID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// verificationTitle returns a display title for a verification entity, using its
// title when present, otherwise its first path.
func verificationTitle(v model.Verification) string {
	if v.Title != "" {
		return v.Title
	}
	if len(v.Paths) > 0 {
		return v.Paths[0]
	}
	return string(v.ID)
}
