// Package context assembles the deterministic context bundle for a capability.
//
// The bundle is the unit an AI coding agent starts from: it gathers a capability
// together with the invariants, specifications, ADRs, verification, scenarios,
// and tasks linked to it, so the agent can answer why the capability exists, what
// behaviour is expected, what decisions constrain it, what verifies it, and which
// scenarios depend on it.
package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/a-h/cap/model"
)

// Bundle is the resolved context for a single capability. Referenced entities are
// resolved to their full form; references that do not resolve are recorded in
// Unresolved rather than causing an error, consistent with the warn-by-default
// principle.
type Bundle struct {
	Capability     model.Capability      `json:"capability"`
	Context        *model.Context        `json:"context,omitempty"`
	Invariants     []model.Invariant     `json:"invariants"`
	Specifications []model.Specification `json:"specifications"`
	ADRs           []model.ADR           `json:"adrs"`
	Verification   []model.Verification  `json:"verification"`
	Scenarios      []model.Scenario      `json:"scenarios"`
	Tasks          []model.Task          `json:"tasks"`
	Unresolved     []model.ID            `json:"unresolved,omitempty"`
}

// For assembles the context bundle for the capability with the given identifier.
// It reports ok=false when no capability with that identifier exists.
func For(m *model.Model, id model.ID) (b Bundle, ok bool) {
	cap, ok := m.Capabilities[id]
	if !ok {
		return Bundle{}, false
	}
	b.Capability = cap
	if cap.Context != "" {
		if ctx, found := m.Contexts[cap.Context]; found {
			b.Context = &ctx
		} else {
			b.Unresolved = append(b.Unresolved, cap.Context)
		}
	}
	for _, ref := range invariantsFor(m, cap) {
		if inv, found := m.Invariants[ref]; found {
			b.Invariants = append(b.Invariants, inv)
			continue
		}
		b.Unresolved = append(b.Unresolved, ref)
	}
	for _, ref := range cap.Specifications {
		if s, found := m.Specifications[ref]; found {
			b.Specifications = append(b.Specifications, s)
			continue
		}
		b.Unresolved = append(b.Unresolved, ref)
	}
	for _, ref := range cap.ADRs {
		if a, found := m.ADRs[ref]; found {
			b.ADRs = append(b.ADRs, a)
			continue
		}
		b.Unresolved = append(b.Unresolved, ref)
	}
	for _, ref := range cap.Verification {
		if v, found := m.Verification[ref]; found {
			b.Verification = append(b.Verification, v)
			continue
		}
		b.Unresolved = append(b.Unresolved, ref)
	}
	for _, ref := range cap.Tasks {
		if t, found := m.Tasks[ref]; found {
			b.Tasks = append(b.Tasks, t)
			continue
		}
		b.Unresolved = append(b.Unresolved, ref)
	}
	b.Scenarios = findScenarios(m, cap)
	return b, true
}

// invariantsFor returns the identifiers of every invariant linked to the
// capability, whether named by the capability or naming it. The union is
// deduplicated and ordered by identifier for deterministic output.
func invariantsFor(m *model.Model, cap model.Capability) []model.ID {
	seen := map[model.ID]struct{}{}
	for _, ref := range cap.Invariants {
		seen[ref] = struct{}{}
	}
	for id, inv := range m.Invariants {
		for _, c := range inv.Capabilities {
			if c == cap.ID {
				seen[id] = struct{}{}
			}
		}
	}
	out := make([]model.ID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// findScenarios returns every scenario linked to the capability, whether declared on
// the capability or discovered by a scenario that references it. The union is
// deduplicated and ordered by identifier for deterministic output.
func findScenarios(m *model.Model, cap model.Capability) []model.Scenario {
	seen := map[model.ID]model.Scenario{}
	for _, ref := range cap.Scenarios {
		if j, found := m.Scenarios[ref]; found {
			seen[j.ID] = j
		}
	}
	for _, j := range m.Scenarios {
		for _, c := range j.Capabilities {
			if c == cap.ID {
				seen[j.ID] = j
			}
		}
	}
	scenarios := make([]model.Scenario, 0, len(seen))
	for _, j := range seen {
		scenarios = append(scenarios, j)
	}
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	return scenarios
}

// String renders a human-readable form of the bundle.
func (b Bundle) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s: %s\n", b.Capability.ID, b.Capability.Name)
	if b.Context != nil {
		fmt.Fprintf(&sb, "Context: %s (%s)\n", b.Context.Name, b.Context.ID)
	}
	if b.Capability.Status != "" {
		fmt.Fprintf(&sb, "Status: %s\n", b.Capability.Status)
	}
	if len(b.Invariants) > 0 {
		sb.WriteString("\nInvariants:\n")
		for _, inv := range b.Invariants {
			fmt.Fprintf(&sb, "  - %s: %s\n", inv.ID, inv.Title)
		}
	}
	if len(b.Specifications) > 0 {
		sb.WriteString("\nSpecifications:\n")
		for _, s := range b.Specifications {
			fmt.Fprintf(&sb, "  - %s: %s\n", s.ID, s.Title)
		}
	}
	if len(b.ADRs) > 0 {
		sb.WriteString("\nADRs:\n")
		for _, a := range b.ADRs {
			fmt.Fprintf(&sb, "  - %s: %s\n", a.ID, a.Title)
		}
	}
	if len(b.Verification) > 0 {
		sb.WriteString("\nVerification:\n")
		for _, v := range b.Verification {
			fmt.Fprintf(&sb, "  - %s: %s\n", v.ID, getTitle(v))
		}
	}
	if len(b.Scenarios) > 0 {
		sb.WriteString("\nScenarios:\n")
		for _, s := range b.Scenarios {
			fmt.Fprintf(&sb, "  - %s: %s\n", s.ID, s.Name)
		}
	}
	if len(b.Tasks) > 0 {
		sb.WriteString("\nTasks:\n")
		for _, t := range b.Tasks {
			fmt.Fprintf(&sb, "  - %s: %s\n", t.ID, t.Title)
		}
	}
	if len(b.Unresolved) > 0 {
		sb.WriteString("\nUnresolved references:\n")
		for _, id := range b.Unresolved {
			fmt.Fprintf(&sb, "  - %s\n", id)
		}
	}
	return sb.String()
}

// getTitle returns the human-readable title of a verification entity: its own title
// when present, otherwise its first path, otherwise its identifier.
func getTitle(v model.Verification) string {
	if v.Title != "" {
		return v.Title
	}
	if len(v.Paths) > 0 {
		return v.Paths[0]
	}
	return string(v.ID)
}
