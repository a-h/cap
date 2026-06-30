package validate

import (
	"strings"
	"testing"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

func TestCheck(t *testing.T) {
	// coveredCapability returns a capability with a status and a registered inline
	// verification and specification, so that the status, verification, and
	// specification gap checks do not interfere with reference-resolution assertions.
	// It registers both entities in m.
	coveredCapability := func(m *model.Model, id model.ID, name string) model.Capability {
		verID := model.SynthesiseID(id, "ver", 1)
		m.Verification[verID] = model.Verification{ID: verID, Paths: []string{"x_test.go"}, Owner: id}
		specID := model.SynthesiseID(id, "spec", 1)
		m.Specifications[specID] = model.Specification{ID: specID, Title: "Design", Of: id, Owner: id}
		return model.Capability{ID: id, Name: name, Status: model.StatusDone, Verification: []model.ID{verID}, Specifications: []model.ID{specID}}
	}

	t.Run("a resolvable reference of the expected kind produces no problems", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["inv-0001"] = model.Invariant{ID: "inv-0001", Title: "Policies must be evaluated consistently.", Capabilities: []model.ID{"cap-0003"}}
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"inv-0001"}
		m.Capabilities["cap-0003"] = c
		if problems := Check(m, nil); len(problems) != 0 {
			t.Errorf("expected no problems, got %v", problems)
		}
	})

	t.Run("a shared invariant and capability that name each other from only one side warn", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["inv-0001"] = model.Invariant{ID: "inv-0001", Title: "Policies must be evaluated consistently."}
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"inv-0001"}
		m.Capabilities["cap-0003"] = c
		problems := Check(m, nil)
		if len(problems) != 1 {
			t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
		}
		if problems[0].Severity != store.SeverityWarning {
			t.Errorf("got severity %q, expected warning", problems[0].Severity)
		}
		if !strings.Contains(problems[0].Message, "cap-0003") || !strings.Contains(problems[0].Message, "inv-0001") {
			t.Errorf("expected the message to name both entities, got %q", problems[0].Message)
		}
	})

	t.Run("a link declared on both entities produces no asymmetry warning", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["inv-0001"] = model.Invariant{ID: "inv-0001", Title: "Policies must be evaluated consistently.", Capabilities: []model.ID{"cap-0003"}}
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"inv-0001"}
		m.Capabilities["cap-0003"] = c
		if problems := Check(m, nil); len(problems) != 0 {
			t.Errorf("expected no problems for a symmetric link, got %v", problems)
		}
	})

	t.Run("a dangling reference is reported as a warning", func(t *testing.T) {
		m := model.NewModel()
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"inv-0999"}
		m.Capabilities["cap-0003"] = c
		problems := Check(m, nil)
		if len(problems) != 1 {
			t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
		}
		if problems[0].Severity != store.SeverityWarning {
			t.Errorf("got severity %q, expected warning", problems[0].Severity)
		}
		if !strings.Contains(problems[0].Message, "inv-0999") {
			t.Errorf("expected the message to name inv-0999, got %q", problems[0].Message)
		}
	})

	t.Run("a reference to the wrong kind is reported as a warning", func(t *testing.T) {
		m := model.NewModel()
		m.Tasks["task-0001"] = model.Task{ID: "task-0001", Title: "Implement a policy cache", Status: model.StatusDraft}
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"task-0001"}
		m.Capabilities["cap-0003"] = c
		problems := Check(m, nil)
		if len(problems) != 1 {
			t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
		}
		if !strings.Contains(problems[0].Message, "not a invariant") {
			t.Errorf("expected a kind-mismatch message, got %q", problems[0].Message)
		}
	})

	t.Run("a scenario referencing an existing capability produces no problems", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0003"] = coveredCapability(m, "cap-0003", "Evaluate policies")
		m.Scenarios["scn-0001"] = model.Scenario{ID: "scn-0001", Name: "Claim approval", Capabilities: []model.ID{"cap-0003"}}
		if problems := Check(m, nil); len(problems) != 0 {
			t.Errorf("expected no problems, got %v", problems)
		}
	})

	t.Run("a concept belonging to an existing context produces no problems", func(t *testing.T) {
		m := model.NewModel()
		m.Contexts["ctx-0001"] = model.Context{ID: "ctx-0001", Name: "Policy enforcement"}
		m.Concepts["con-0001"] = model.Concept{ID: "con-0001", Name: "Policy", Context: "ctx-0001"}
		if problems := Check(m, nil); len(problems) != 0 {
			t.Errorf("expected no problems, got %v", problems)
		}
	})

	t.Run("a concept naming a context that does not exist is warned", func(t *testing.T) {
		m := model.NewModel()
		m.Concepts["con-0001"] = model.Concept{ID: "con-0001", Name: "Policy", Context: "ctx-9999"}
		problems := Check(m, nil)
		if len(problems) != 1 {
			t.Fatalf("expected 1 problem, got %d: %v", len(problems), problems)
		}
		if !strings.Contains(problems[0].Message, "con-0001") || !strings.Contains(problems[0].Message, "ctx-9999") {
			t.Errorf("expected the message to name the concept and the missing context, got %q", problems[0].Message)
		}
	})

	t.Run("a capability's own inline invariant produces no problems", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["cap-0003/inv-1"] = model.Invariant{ID: "cap-0003/inv-1", Title: "Policies are evaluated consistently.", Owner: "cap-0003"}
		c := coveredCapability(m, "cap-0003", "Evaluate policies")
		c.Invariants = []model.ID{"cap-0003/inv-1"}
		m.Capabilities["cap-0003"] = c
		if problems := Check(m, nil); len(problems) != 0 {
			t.Errorf("expected no problems for a capability's own inline invariant, got %v", problems)
		}
	})

	t.Run("a capability with no verification is warned", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies"}
		m.Capabilities["cap-0007"] = coveredCapability(m, "cap-0007", "Process payments")
		problems := Check(m, nil)
		if !containsMessage(problems, "cap-0003 has no verification") {
			t.Errorf("expected a no-verification warning for cap-0003, got %v", problems)
		}
		if containsMessage(problems, "cap-0007 has no verification") {
			t.Errorf("did not expect a warning for the tested capability, got %v", problems)
		}
	})

	t.Run("a capability with no specification is warned", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies"}
		m.Capabilities["cap-0007"] = coveredCapability(m, "cap-0007", "Process payments")
		problems := Check(m, nil)
		if !containsMessage(problems, "cap-0003 has no specification") {
			t.Errorf("expected a no-specification warning for cap-0003, got %v", problems)
		}
		if containsMessage(problems, "cap-0007 has no specification") {
			t.Errorf("did not expect a warning for the documented capability, got %v", problems)
		}
	})

	t.Run("a capability or task with no status is warned", func(t *testing.T) {
		m := model.NewModel()
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies"}
		m.Capabilities["cap-0007"] = coveredCapability(m, "cap-0007", "Process payments")
		m.Tasks["task-0001"] = model.Task{ID: "task-0001", Title: "Implement a cache"}
		m.Tasks["task-0002"] = model.Task{ID: "task-0002", Title: "Add metrics", Status: model.StatusInProgress}
		problems := Check(m, nil)
		if !containsMessage(problems, "cap-0003 has no status") {
			t.Errorf("expected a no-status warning for cap-0003, got %v", problems)
		}
		if !containsMessage(problems, "task-0001 has no status") {
			t.Errorf("expected a no-status warning for task-0001, got %v", problems)
		}
		if containsMessage(problems, "cap-0007 has no status") || containsMessage(problems, "task-0002 has no status") {
			t.Errorf("did not expect a status warning for entities that set one, got %v", problems)
		}
	})

	t.Run("a capability documented by a context-level specification is not warned", func(t *testing.T) {
		m := model.NewModel()
		m.Contexts["ctx-0001"] = model.Context{ID: "ctx-0001", Name: "Policy enforcement"}
		m.Specifications["spec-0001"] = model.Specification{ID: "spec-0001", Title: "Auditing design", Of: "ctx-0001"}
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies", Context: "ctx-0001"}
		problems := Check(m, nil)
		if containsMessage(problems, "cap-0003 has no specification") {
			t.Errorf("a capability covered by a context-level specification must not be warned, got %v", problems)
		}
	})

	t.Run("referencing another capability's inline entity is warned", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["cap-0003/inv-1"] = model.Invariant{ID: "cap-0003/inv-1", Title: "Decisions are auditable.", Owner: "cap-0003"}
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies", Invariants: []model.ID{"cap-0003/inv-1"}}
		m.Capabilities["cap-0007"] = model.Capability{ID: "cap-0007", Name: "Process payments", Invariants: []model.ID{"cap-0003/inv-1"}}
		problems := Check(m, nil)
		if !containsMessage(problems, "cannot be referenced from elsewhere") {
			t.Errorf("expected a warning that an inline entity cannot be referenced from elsewhere, got %v", problems)
		}
	})

	t.Run("identical inline invariant text in two capabilities is hinted for extraction", func(t *testing.T) {
		m := model.NewModel()
		m.Invariants["cap-0003/inv-1"] = model.Invariant{ID: "cap-0003/inv-1", Title: "Decisions are auditable.", Owner: "cap-0003"}
		m.Invariants["cap-0007/inv-1"] = model.Invariant{ID: "cap-0007/inv-1", Title: "Decisions are auditable.", Owner: "cap-0007"}
		m.Capabilities["cap-0003"] = model.Capability{ID: "cap-0003", Name: "Evaluate policies", Invariants: []model.ID{"cap-0003/inv-1"}}
		m.Capabilities["cap-0007"] = model.Capability{ID: "cap-0007", Name: "Process payments", Invariants: []model.ID{"cap-0007/inv-1"}}
		problems := Check(m, nil)
		if !containsMessage(problems, "consider extracting it to a shared file") {
			t.Errorf("expected an extraction hint, got %v", problems)
		}
	})
}

func containsMessage(problems []store.Problem, substr string) bool {
	for _, p := range problems {
		if strings.Contains(p.Message, substr) {
			return true
		}
	}
	return false
}
