// Package validate checks a loaded cap model for reference integrity and coverage
// gaps.
//
// Validation is tolerant by default: dangling or mistyped references are
// warnings, and structural problems surfaced during loading are errors. Callers
// decide, via a strict mode, whether warnings should affect the exit status.
package validate

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

// reference is one expected link from a source entity to a target identifier. want
// lists the kinds the target may be; an empty want accepts any kind.
type reference struct {
	source  model.ID
	section string
	target  model.ID
	want    []model.Kind
}

// allows reports whether kind is one of the wanted kinds, treating an empty want as
// accepting any kind.
func (r reference) allows(kind model.Kind) bool {
	if len(r.want) == 0 {
		return true
	}
	for _, w := range r.want {
		if kind == w {
			return true
		}
	}
	return false
}

// wantList renders the wanted kinds for a message, for example "a capability or a
// bounded context".
func wantList(want []model.Kind) string {
	switch len(want) {
	case 0:
		return ""
	case 1:
		return string(want[0])
	}
	names := make([]string, len(want))
	for i, w := range want {
		names[i] = string(w)
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// Check resolves every cross-entity reference in the model and returns warnings
// for references that do not resolve, or that resolve to the wrong kind. It does
// not repeat the structural problems already gathered during loading. The files
// map associates an identifier with the file it was loaded from, so that warnings
// name the source file; it may be nil.
func Check(m *model.Model, files map[model.ID]string) []store.Problem {
	var problems []store.Problem
	for _, ref := range collectReferences(m) {
		kind, ok := m.Lookup(ref.target)
		if !ok {
			problems = append(problems, store.Problem{
				File:     files[ref.source],
				Severity: store.SeverityWarning,
				Message:  fmt.Sprintf("%s references %s in %s, but %s does not exist", ref.source, ref.target, ref.section, ref.target),
			})
			continue
		}
		if !ref.allows(kind) {
			problems = append(problems, store.Problem{
				File:     files[ref.source],
				Severity: store.SeverityWarning,
				Message:  fmt.Sprintf("%s references %s in %s, but %s is a %s, not a %s", ref.source, ref.target, ref.section, ref.target, kind, wantList(ref.want)),
			})
			continue
		}
		if owner, inline := inlineOwner(m, ref.target); inline && owner != ref.source {
			problems = append(problems, store.Problem{
				File:     files[ref.source],
				Severity: store.SeverityWarning,
				Message:  fmt.Sprintf("%s references %s in %s, but %s is inline to %s and cannot be referenced from elsewhere; extract it to its own file to share it", ref.source, ref.target, ref.section, ref.target, owner),
			})
		}
	}
	problems = append(problems, hintInlineDuplicates(m, files)...)
	problems = append(problems, hintAsymmetricLinks(m, files)...)
	problems = append(problems, hintGaps(m, files)...)
	return problems
}

// hintAsymmetricLinks warns when a file-backed invariant and a capability name each
// other from only one side. The invariant-to-capability link may be declared from
// either end, so the model treats a one-sided link as present; the warning catches
// the likelier cause, a half-deleted or forgotten link, and prompts declaring it on
// both entities so either file shows the relationship on its own. Inline invariants
// are owned by a single capability and excluded, since they cannot be named from
// another file.
func hintAsymmetricLinks(m *model.Model, files map[model.ID]string) []store.Problem {
	var problems []store.Problem
	for _, inv := range m.Invariants {
		if inv.Inline() {
			continue
		}
		for _, cid := range inv.Capabilities {
			c, ok := m.Capabilities[cid]
			if !ok {
				continue
			}
			if !slices.Contains(c.Invariants, inv.ID) {
				problems = append(problems, store.Problem{
					File:     files[inv.ID],
					Severity: store.SeverityWarning,
					Message:  fmt.Sprintf("%s names %s, but %s does not name %s; declare the link on both entities or remove it", inv.ID, cid, cid, inv.ID),
				})
			}
		}
	}
	for _, c := range m.Capabilities {
		for _, iid := range c.Invariants {
			inv, ok := m.Invariants[iid]
			if !ok || inv.Inline() {
				continue
			}
			if !slices.Contains(inv.Capabilities, c.ID) {
				problems = append(problems, store.Problem{
					File:     files[c.ID],
					Severity: store.SeverityWarning,
					Message:  fmt.Sprintf("%s names %s, but %s does not name %s; declare the link on both entities or remove it", c.ID, iid, iid, c.ID),
				})
			}
		}
	}
	sort.Slice(problems, func(i, j int) bool { return problems[i].Message < problems[j].Message })
	return problems
}

// hintGaps warns about capabilities with a coverage gap: no verification (untested),
// or no specification (undocumented). A capability is documented by a specification
// of it, or by a specification of the bounded context it belongs to.
func hintGaps(m *model.Model, files map[model.ID]string) []store.Problem {
	documented := map[model.ID]bool{}
	for _, s := range m.Specifications {
		if _, ok := m.Capabilities[s.Of]; ok {
			documented[s.Of] = true
			continue
		}
		if _, ok := m.Contexts[s.Of]; ok {
			for cid, c := range m.Capabilities {
				if c.Context == s.Of {
					documented[cid] = true
				}
			}
		}
	}
	var ids []model.ID
	for id := range m.Capabilities {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var problems []store.Problem
	for _, id := range ids {
		if len(m.Capabilities[id].Verification) == 0 {
			problems = append(problems, store.Problem{File: files[id], Severity: store.SeverityWarning, Message: fmt.Sprintf("%s has no verification", id)})
		}
		if !documented[id] {
			problems = append(problems, store.Problem{File: files[id], Severity: store.SeverityWarning, Message: fmt.Sprintf("%s has no specification", id)})
		}
		if m.Capabilities[id].Status == "" {
			problems = append(problems, store.Problem{File: files[id], Severity: store.SeverityWarning, Message: fmt.Sprintf("%s has no status", id)})
		}
	}
	var taskIDs []model.ID
	for id := range m.Tasks {
		taskIDs = append(taskIDs, id)
	}
	sort.Slice(taskIDs, func(i, j int) bool { return taskIDs[i] < taskIDs[j] })
	for _, id := range taskIDs {
		if m.Tasks[id].Status == "" {
			problems = append(problems, store.Problem{File: files[id], Severity: store.SeverityWarning, Message: fmt.Sprintf("%s has no status", id)})
		}
	}
	return problems
}

// inlineOwner reports whether the identifier names an inline entity and, if so, the
// capability that owns it.
func inlineOwner(m *model.Model, id model.ID) (owner model.ID, inline bool) {
	if inv, ok := m.Invariants[id]; ok && inv.Inline() {
		return inv.Owner, true
	}
	if s, ok := m.Specifications[id]; ok && s.Inline() {
		return s.Owner, true
	}
	if v, ok := m.Verification[id]; ok && v.Inline() {
		return v.Owner, true
	}
	return "", false
}

// hintInlineDuplicates warns, as an advisory, when identical inline invariant text
// appears in more than one capability, suggesting it be extracted to a shared file.
func hintInlineDuplicates(m *model.Model, files map[model.ID]string) []store.Problem {
	owners := map[string]map[model.ID]struct{}{}
	for _, inv := range m.Invariants {
		if !inv.Inline() {
			continue
		}
		if owners[inv.Title] == nil {
			owners[inv.Title] = map[model.ID]struct{}{}
		}
		owners[inv.Title][inv.Owner] = struct{}{}
	}
	var titles []string
	for title, set := range owners {
		if len(set) >= 2 {
			titles = append(titles, title)
		}
	}
	sort.Strings(titles)
	var problems []store.Problem
	for _, title := range titles {
		problems = append(problems, store.Problem{
			Severity: store.SeverityWarning,
			Message:  fmt.Sprintf("inline invariant %q appears in %d capabilities; consider extracting it to a shared file", title, len(owners[title])),
		})
	}
	return problems
}

// collectReferences enumerates every expected link in the model.
func collectReferences(m *model.Model) []reference {
	var refs []reference
	add := func(source model.ID, section string, want model.Kind, targets []model.ID) {
		for _, t := range targets {
			refs = append(refs, reference{source: source, section: section, target: t, want: []model.Kind{want}})
		}
	}
	for _, c := range m.Capabilities {
		if c.Context != "" {
			refs = append(refs, reference{source: c.ID, section: store.SectionMetadata, target: c.Context, want: []model.Kind{model.KindContext}})
		}
		add(c.ID, store.SectionInvariants, model.KindInvariant, c.Invariants)
		add(c.ID, store.SectionSpecifications, model.KindSpecification, c.Specifications)
		add(c.ID, store.SectionADRs, model.KindADR, c.ADRs)
		add(c.ID, store.SectionScenarios, model.KindScenario, c.Scenarios)
		add(c.ID, store.SectionVerification, model.KindVerification, c.Verification)
		add(c.ID, store.SectionTasks, model.KindTask, c.Tasks)
	}
	for _, inv := range m.Invariants {
		add(inv.ID, store.SectionCapabilities, model.KindCapability, inv.Capabilities)
	}
	for _, s := range m.Specifications {
		if s.Of != "" {
			refs = append(refs, reference{source: s.ID, section: store.SectionMetadata, target: s.Of, want: []model.Kind{model.KindCapability, model.KindContext}})
		}
	}
	for _, j := range m.Scenarios {
		add(j.ID, store.SectionCapabilities, model.KindCapability, j.Capabilities)
	}
	return refs
}
