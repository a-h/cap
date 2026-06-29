package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/cap/model"
)

func TestProblemString(t *testing.T) {
	tests := []struct {
		name    string
		problem Problem
		want    string
	}{
		{
			name:    "a problem with a line includes the line in the location",
			problem: Problem{File: "cap/capabilities/cap-0003.md", Line: 7, Severity: SeverityWarning, Message: "thing is wrong"},
			want:    "cap/capabilities/cap-0003.md:7: warning: thing is wrong",
		},
		{
			name:    "a problem without a line omits the line from the location",
			problem: Problem{File: "cap/capabilities/cap-0003.md", Severity: SeverityError, Message: "thing is broken"},
			want:    "cap/capabilities/cap-0003.md: error: thing is broken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.problem.String(); got != tt.want {
				t.Errorf("got %q, expected %q", got, tt.want)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		wantID model.ID
		wantOK bool
	}{
		{name: "a canonical identifier with a descriptive slug is parsed", path: "cap/capabilities/cap-0003-evaluate-policies.md", wantID: "cap-0003", wantOK: true},
		{name: "an uppercase prefix is normalised to lowercase", path: "cap/capabilities/CAP-0003-evaluate-policies.md", wantID: "cap-0003", wantOK: true},
		{name: "an unpadded number is zero-padded", path: "cap/invariants/inv-1.md", wantID: "inv-0001", wantOK: true},
		{name: "an uppercase unpadded identifier is fully normalised", path: "cap/invariants/INV-1-thing.md", wantID: "inv-0001", wantOK: true},
		{name: "a verification identifier is parsed", path: "cap/verification/ver-0008-smoke.md", wantID: "ver-0008", wantOK: true},
		{name: "a filename without an identifier reports not ok", path: "cap/contexts/notes.md", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := ParseID(tt.path)
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && id != tt.wantID {
				t.Errorf("got %q, expected %q", id, tt.wantID)
			}
		})
	}
}

func writeFile(t *testing.T, root, dir, name, content string) {
	t.Helper()
	full := filepath.Join(root, dir)
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(full, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestLoad(t *testing.T) {
	t.Run("a capability is loaded with its metadata and references normalised to canonical form", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", `# Evaluate policies

## Metadata

- context: CTX-1
- status: done

## Invariants

- [INV-001](../invariants/INV-001.md)

## Verification

- unit-42
`)
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cap, ok := res.Model.Capabilities["cap-0003"]
		if !ok {
			t.Fatalf("expected cap-0003 to be loaded")
		}
		if cap.Name != "Evaluate policies" {
			t.Errorf("got name %q, expected %q", cap.Name, "Evaluate policies")
		}
		if cap.Context != "ctx-0001" {
			t.Errorf("got context %q, expected %q", cap.Context, "ctx-0001")
		}
		if cap.Status != model.StatusDone {
			t.Errorf("got status %q, expected %q", cap.Status, model.StatusDone)
		}
		if len(cap.Invariants) != 1 || cap.Invariants[0] != "inv-0001" {
			t.Errorf("got invariants %v, expected [inv-0001]", cap.Invariants)
		}
		if len(cap.Verification) != 1 || cap.Verification[0] != "unit-0042" {
			t.Errorf("got verification %v, expected [unit-0042]", cap.Verification)
		}
	})

	t.Run("every entity kind is loaded from its directory", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "contexts", "ctx-0001-policy-enforcement.md", "# Policy enforcement\n")
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n")
		writeFile(t, root, "invariants", "inv-0001-consistency.md", "# Policies must be evaluated consistently.\n")
		writeFile(t, root, "specifications", "spec-0012-semantics.md", "# Policy evaluation semantics\n\n## Metadata\n\n- of: cap-0003\n")
		writeFile(t, root, "adrs", "adr-0014-engine.md", "# Policy evaluation engine\n")
		writeFile(t, root, "scenarios", "scn-0001-claim-approval.md", "# Claim approval\n\n## Capabilities\n\n- cap-0003\n")
		writeFile(t, root, "verification", "ver-0008-smoke.md", "# Pre-release smoke test\n\n## Description\n\nx\n\n## Paths\n\n- tests/e2e/policy.spec.ts\n")
		writeFile(t, root, "tasks", "task-0341-cache.md", "# Implement a policy evaluation cache\n")

		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, p := range res.Problems {
			if p.Severity == SeverityError {
				t.Fatalf("expected no errors, got %v", res.Problems)
			}
		}
		m := res.Model
		if a, ok := m.Contexts["ctx-0001"]; !ok || a.Name != "Policy enforcement" {
			t.Errorf("context not loaded correctly: %#v", a)
		}
		if c, ok := m.Capabilities["cap-0003"]; !ok || c.Name != "Evaluate policies" {
			t.Errorf("capability not loaded correctly: %#v", c)
		}
		if r, ok := m.Invariants["inv-0001"]; !ok || r.Title != "Policies must be evaluated consistently." {
			t.Errorf("invariant not loaded correctly: %#v", r)
		}
		if s, ok := m.Specifications["spec-0012"]; !ok || s.Of != "cap-0003" {
			t.Errorf("specification not loaded correctly: %#v", s)
		}
		if a, ok := m.ADRs["adr-0014"]; !ok || a.Title != "Policy evaluation engine" {
			t.Errorf("adr not loaded correctly: %#v", a)
		}
		j, ok := m.Scenarios["scn-0001"]
		if !ok || len(j.Capabilities) != 1 || j.Capabilities[0] != "cap-0003" {
			t.Errorf("scenario not loaded correctly: %#v", j)
		}
		v, ok := m.Verification["ver-0008"]
		if !ok || v.Title != "Pre-release smoke test" || len(v.Paths) != 1 || v.Paths[0] != "tests/e2e/policy.spec.ts" {
			t.Errorf("verification not loaded correctly: %#v", v)
		}
		if task, ok := m.Tasks["task-0341"]; !ok || task.Title != "Implement a policy evaluation cache" {
			t.Errorf("task not loaded correctly: %#v", task)
		}
		if res.Files["cap-0003"] == "" {
			t.Errorf("expected the source file to be recorded for cap-0003")
		}
	})

	t.Run("a duplicate identifier is reported as an error", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-one.md", "# One\n")
		writeFile(t, root, "capabilities", "cap-0003-two.md", "# Two\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityError, "duplicate") {
			t.Errorf("expected a duplicate identifier error, got %v", res.Problems)
		}
	})

	t.Run("an identifier whose prefix does not match its directory is warned", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "inv-0001-misfiled.md", "# Misfiled invariant\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityWarning, "stored under") {
			t.Errorf("expected a prefix/directory mismatch warning, got %v", res.Problems)
		}
	})

	t.Run("a malformed metadata item is warned but does not stop loading", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-thing.md", "# Thing\n\n## Metadata\n\n- not a pair\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := res.Model.Capabilities["cap-0003"]; !ok {
			t.Errorf("expected the capability to load despite the malformed metadata")
		}
		if !hasProblem(res.Problems, SeverityWarning, "key: value") {
			t.Errorf("expected a malformed metadata warning, got %v", res.Problems)
		}
	})

	t.Run("a reference item without an identifier is warned in a reference-only section", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-thing.md", "# Thing\n\n## Scenarios\n\n- see the related work\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityWarning, "does not reference an identifier") {
			t.Errorf("expected a missing-identifier warning, got %v", res.Problems)
		}
	})

	t.Run("a prose item under Invariants is loaded as an inline invariant", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-thing.md", "# Thing\n\n## Invariants\n\n- Policies are evaluated consistently.\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		inv, ok := res.Model.Invariants["cap-0003/inv-1"]
		if !ok {
			t.Fatalf("expected an inline invariant cap-0003/inv-1, got %v", res.Model.Invariants)
		}
		if inv.Title != "Policies are evaluated consistently." || inv.Owner != "cap-0003" {
			t.Errorf("unexpected inline invariant: %#v", inv)
		}
		cap := res.Model.Capabilities["cap-0003"]
		if len(cap.Invariants) != 1 || cap.Invariants[0] != "cap-0003/inv-1" {
			t.Errorf("expected the capability to link the inline invariant, got %v", cap.Invariants)
		}
	})

	t.Run("a path under Verification is loaded as inline verification", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-thing.md", "# Thing\n\n## Verification\n\n- internal/policy/eval_test.go\n- ver-0008\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := res.Model.Verification["cap-0003/ver-1"]
		if !ok {
			t.Fatalf("expected an inline verification cap-0003/ver-1, got %v", res.Model.Verification)
		}
		if v.Owner != "cap-0003" || len(v.Paths) != 1 || v.Paths[0] != "internal/policy/eval_test.go" {
			t.Errorf("unexpected inline verification: %#v", v)
		}
		cap := res.Model.Capabilities["cap-0003"]
		if len(cap.Verification) != 2 || cap.Verification[0] != "cap-0003/ver-1" || cap.Verification[1] != "ver-0008" {
			t.Errorf("expected the capability to link inline verification and the reference, got %v", cap.Verification)
		}
	})

	t.Run("a subsection under Specifications is loaded as an inline specification", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-thing.md", "# Thing\n\n## Specifications\n\n### Policy evaluation semantics\n\n- An empty policy set denies by default.\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		spec, ok := res.Model.Specifications["cap-0003/spec-1"]
		if !ok {
			t.Fatalf("expected an inline specification cap-0003/spec-1, got %v", res.Model.Specifications)
		}
		if spec.Title != "Policy evaluation semantics" || spec.Owner != "cap-0003" {
			t.Errorf("unexpected inline specification: %#v", spec)
		}
		if len(spec.Detail) != 1 || spec.Detail[0] != "An empty policy set denies by default." {
			t.Errorf("expected the specification detail to be captured, got %v", spec.Detail)
		}
	})

	t.Run("a missing required section is reported as a warning", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\nIt evaluates policies.\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityWarning, `required section "Scope" is missing`) {
			t.Errorf("expected a missing Scope warning, got %v", res.Problems)
		}
	})

	t.Run("an empty required section is reported as a warning", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\n## Scope\n\nIn scope:\n\n- Evaluation.\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityWarning, `required section "Description" is empty`) {
			t.Errorf("expected an empty Description warning, got %v", res.Problems)
		}
	})

	t.Run("an optional section may be omitted without warning", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\nIt evaluates policies.\n\n## Scope\n\nIn scope:\n\n- Evaluation.\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hasProblem(res.Problems, SeverityWarning, "Actors") {
			t.Errorf("did not expect a warning about the optional Actors section, got %v", res.Problems)
		}
		if hasProblem(res.Problems, SeverityWarning, "required section") {
			t.Errorf("expected no missing-section warnings, got %v", res.Problems)
		}
	})

	t.Run("an unknown capability status is reported as an error", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "CAP-001-thing.md", "# Thing\n\n## Metadata\n\n- status: wibble\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityError, "wibble") {
			t.Errorf("expected an error about the unknown status, got %v", res.Problems)
		}
		if !hasProblem(res.Problems, SeverityError, "valid statuses are") {
			t.Errorf("expected the error to list the valid statuses, got %v", res.Problems)
		}
		for _, s := range model.Statuses {
			if !hasProblem(res.Problems, SeverityError, string(s)) {
				t.Errorf("expected the error to name the valid status %q, got %v", s, res.Problems)
			}
		}
	})

	t.Run("a filename without an identifier is reported as an error", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "notes.md", "# Notes\n")
		res, err := Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !hasProblem(res.Problems, SeverityError, "identifier") {
			t.Errorf("expected an error about the identifier, got %v", res.Problems)
		}
	})

	t.Run("a missing system directory loads an empty model without error", func(t *testing.T) {
		res, err := Load(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Model.Capabilities) != 0 {
			t.Errorf("expected an empty model")
		}
	})
}

func hasProblem(problems []Problem, severity Severity, substr string) bool {
	for _, p := range problems {
		if p.Severity == severity && strings.Contains(p.Message, substr) {
			return true
		}
	}
	return false
}
