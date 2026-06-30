package context

import (
	"strings"
	"testing"

	"github.com/a-h/cap/model"
)

func newModel(t *testing.T, entities ...any) *model.Model {
	t.Helper()
	m := model.NewModel()
	for _, e := range entities {
		switch v := e.(type) {
		case model.Context:
			m.Contexts[v.ID] = v
		case model.Capability:
			m.Capabilities[v.ID] = v
		case model.Invariant:
			m.Invariants[v.ID] = v
		case model.Specification:
			m.Specifications[v.ID] = v
		case model.ADR:
			m.ADRs[v.ID] = v
		case model.Scenario:
			m.Scenarios[v.ID] = v
		case model.Verification:
			m.Verification[v.ID] = v
		case model.Task:
			m.Tasks[v.ID] = v
		default:
			t.Fatalf("newModel: unsupported entity type %T", e)
		}
	}
	return m
}

func TestFor(t *testing.T) {
	t.Run("a capability is not found when it does not exist", func(t *testing.T) {
		m := model.NewModel()
		if _, ok := For(m, "CAP-999"); ok {
			t.Errorf("expected ok=false for a missing capability")
		}
	})

	t.Run("declared references are resolved to full entities", func(t *testing.T) {
		m := newModel(t,
			model.Context{ID: "CTX-001", Name: "Policy enforcement"},
			model.Invariant{ID: "INV-001", Title: "Policies must be evaluated consistently."},
			model.Specification{ID: "SPEC-012", Title: "Policy evaluation semantics", Specifies: []model.ID{"CAP-003"}},
			model.Capability{
				ID: "CAP-003", Name: "Evaluate policies", Context: "CTX-001",
				Invariants:     []model.ID{"INV-001"},
				Specifications: []model.ID{"SPEC-012"},
			},
		)
		b, ok := For(m, "CAP-003")
		if !ok {
			t.Fatalf("expected the capability to be found")
		}
		if b.Context == nil || b.Context.ID != "CTX-001" {
			t.Errorf("expected the context to be resolved, got %v", b.Context)
		}
		if len(b.Invariants) != 1 || b.Invariants[0].Title == "" {
			t.Errorf("expected the invariant to be resolved to its full form, got %v", b.Invariants)
		}
		if len(b.Specifications) != 1 || b.Specifications[0].ID != "SPEC-012" {
			t.Errorf("expected the specification to be resolved, got %v", b.Specifications)
		}
	})

	t.Run("a dangling reference is recorded as unresolved rather than failing", func(t *testing.T) {
		m := newModel(t, model.Capability{ID: "CAP-003", Name: "Evaluate policies", Invariants: []model.ID{"INV-999"}})
		b, ok := For(m, "CAP-003")
		if !ok {
			t.Fatalf("expected the capability to be found")
		}
		if len(b.Invariants) != 0 {
			t.Errorf("expected no resolved invariants, got %v", b.Invariants)
		}
		if len(b.Unresolved) != 1 || b.Unresolved[0] != "INV-999" {
			t.Errorf("expected INV-999 to be unresolved, got %v", b.Unresolved)
		}
	})

	t.Run("scenarios that reference the capability are discovered in reverse", func(t *testing.T) {
		m := newModel(t,
			model.Capability{ID: "CAP-003", Name: "Evaluate policies"},
			model.Scenario{ID: "SCN-001", Name: "Claim approval", Capabilities: []model.ID{"CAP-003"}},
		)
		b, _ := For(m, "CAP-003")
		if len(b.Scenarios) != 1 || b.Scenarios[0].ID != "SCN-001" {
			t.Errorf("expected the scenario to be discovered by reverse lookup, got %v", b.Scenarios)
		}
	})

	t.Run("a scenario linked both ways appears only once", func(t *testing.T) {
		m := newModel(t,
			model.Capability{ID: "CAP-003", Name: "Evaluate policies", Scenarios: []model.ID{"SCN-001"}},
			model.Scenario{ID: "SCN-001", Name: "Claim approval", Capabilities: []model.ID{"CAP-003"}},
		)
		b, _ := For(m, "CAP-003")
		if len(b.Scenarios) != 1 {
			t.Errorf("expected the scenario to be deduplicated, got %d scenarios", len(b.Scenarios))
		}
	})

	t.Run("adrs, verification, and tasks are resolved to full entities", func(t *testing.T) {
		m := newModel(t,
			model.ADR{ID: "adr-0014", Title: "Policy evaluation engine"},
			model.Verification{ID: "ver-0008", Title: "Policy evaluation", Paths: []string{"tests/e2e/policy.spec.ts"}},
			model.Task{ID: "task-0341", Title: "Implement a policy evaluation cache"},
			model.Capability{
				ID: "cap-0003", Name: "Evaluate policies",
				ADRs:         []model.ID{"adr-0014"},
				Verification: []model.ID{"ver-0008"},
				Tasks:        []model.ID{"task-0341"},
			},
		)
		b, ok := For(m, "cap-0003")
		if !ok {
			t.Fatalf("expected the capability to be found")
		}
		if len(b.ADRs) != 1 || b.ADRs[0].Title != "Policy evaluation engine" {
			t.Errorf("expected the ADR to be resolved, got %v", b.ADRs)
		}
		if len(b.Verification) != 1 || b.Verification[0].Title != "Policy evaluation" {
			t.Errorf("expected the verification to be resolved, got %v", b.Verification)
		}
		if len(b.Tasks) != 1 || b.Tasks[0].Title != "Implement a policy evaluation cache" {
			t.Errorf("expected the task to be resolved, got %v", b.Tasks)
		}
	})
}

func TestBundleString(t *testing.T) {
	t.Run("the rendered bundle names the capability, its context, status, and linked entities", func(t *testing.T) {
		m := newModel(t,
			model.Context{ID: "ctx-0001", Name: "Policy enforcement"},
			model.Invariant{ID: "inv-0001", Title: "Policies must be evaluated consistently."},
			model.Scenario{ID: "scn-0001", Name: "Claim approval", Capabilities: []model.ID{"cap-0003"}},
			model.Capability{
				ID: "cap-0003", Name: "Evaluate policies", Context: "ctx-0001",
				Status:     model.StatusDone,
				Invariants: []model.ID{"inv-0001"},
			},
		)
		b, _ := For(m, "cap-0003")
		got := b.String()
		for _, want := range []string{
			"cap-0003: Evaluate policies",
			"Context: Policy enforcement (ctx-0001)",
			"Status: done",
			"inv-0001: Policies must be evaluated consistently.",
			"scn-0001: Claim approval",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("expected the rendered bundle to contain %q, got:\n%s", want, got)
			}
		}
	})

	t.Run("unresolved references are listed in the rendered bundle", func(t *testing.T) {
		m := newModel(t, model.Capability{ID: "cap-0003", Name: "Evaluate policies", Invariants: []model.ID{"inv-0999"}})
		b, _ := For(m, "cap-0003")
		got := b.String()
		if !strings.Contains(got, "Unresolved references:") || !strings.Contains(got, "inv-0999") {
			t.Errorf("expected unresolved references to be listed, got:\n%s", got)
		}
	})
}
