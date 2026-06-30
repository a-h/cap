package model

import "testing"

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{name: "draft is valid", status: StatusDraft, want: true},
		{name: "proposed is valid", status: StatusProposed, want: true},
		{name: "in-progress is valid", status: StatusInProgress, want: true},
		{name: "done is valid", status: StatusDone, want: true},
		{name: "an unknown status is not valid", status: "wibble", want: false},
		{name: "the empty status is not valid", status: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Errorf("got %v, expected %v", got, tt.want)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	t.Run("an entity of each kind is found by its identifier", func(t *testing.T) {
		m := NewModel()
		m.Contexts["ctx-0001"] = Context{ID: "ctx-0001"}
		m.Concepts["con-0001"] = Concept{ID: "con-0001"}
		m.Capabilities["cap-0003"] = Capability{ID: "cap-0003"}
		m.Invariants["inv-0001"] = Invariant{ID: "inv-0001"}
		m.Specifications["spec-0012"] = Specification{ID: "spec-0012"}
		m.ADRs["adr-0014"] = ADR{ID: "adr-0014"}
		m.Scenarios["scn-0001"] = Scenario{ID: "scn-0001"}
		m.Verification["ver-0008"] = Verification{ID: "ver-0008"}
		m.Tasks["task-0341"] = Task{ID: "task-0341"}

		want := map[ID]Kind{
			"ctx-0001":  KindContext,
			"con-0001":  KindConcept,
			"cap-0003":  KindCapability,
			"inv-0001":  KindInvariant,
			"spec-0012": KindSpecification,
			"adr-0014":  KindADR,
			"scn-0001":  KindScenario,
			"ver-0008":  KindVerification,
			"task-0341": KindTask,
		}
		for id, wantKind := range want {
			kind, ok := m.Lookup(id)
			if !ok {
				t.Errorf("expected %s to be found", id)
				continue
			}
			if kind != wantKind {
				t.Errorf("%s: got kind %q, expected %q", id, kind, wantKind)
			}
		}
	})

	t.Run("an identifier that was never loaded is not found", func(t *testing.T) {
		if _, ok := NewModel().Lookup("cap-9999"); ok {
			t.Errorf("expected a missing identifier to report not found")
		}
	})
}

func TestCanonical(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want ID
	}{
		{name: "a canonical identifier is unchanged", id: "cap-0003", want: "cap-0003"},
		{name: "an uppercase prefix is lowercased", id: "CAP-0003", want: "cap-0003"},
		{name: "an unpadded number is zero-padded to four digits", id: "cap-3", want: "cap-0003"},
		{name: "a mixed-case unpadded identifier is fully normalised", id: "Cap-3", want: "cap-0003"},
		{name: "a wider number is preserved", id: "cap-12345", want: "cap-12345"},
		{name: "an alphanumeric prefix is normalised", id: "E2E-8", want: "e2e-0008"},
		{name: "an identifier without a number lowercases the body", id: "cap-ABC", want: "cap-abc"},
		{name: "an identifier without a hyphen is lowercased", id: "DRAFT", want: "draft"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Canonical(); got != tt.want {
				t.Errorf("got %q, expected %q", got, tt.want)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		name       string
		id         ID
		wantPrefix string
		wantOK     bool
	}{
		{name: "a prefix is returned in lowercase", id: "CAP-0003", wantPrefix: "cap", wantOK: true},
		{name: "a canonical prefix is returned as is", id: "req-0001", wantPrefix: "req", wantOK: true},
		{name: "an identifier without a hyphen has no prefix", id: "draft", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, ok := tt.id.Prefix()
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && prefix != tt.wantPrefix {
				t.Errorf("got %q, expected %q", prefix, tt.wantPrefix)
			}
		})
	}
}

func TestMapToEntityKind(t *testing.T) {
	tests := []struct {
		name     string
		id       ID
		wantKind Kind
		wantOK   bool
	}{
		{name: "a context identifier maps to the context kind", id: "ctx-0001", wantKind: KindContext, wantOK: true},
		{name: "a concept identifier maps to the concept kind", id: "con-0001", wantKind: KindConcept, wantOK: true},
		{name: "a capability identifier maps to the capability kind", id: "cap-0003", wantKind: KindCapability, wantOK: true},
		{name: "an invariant identifier maps to the invariant kind", id: "inv-0001", wantKind: KindInvariant, wantOK: true},
		{name: "a specification identifier maps to the specification kind", id: "spec-0012", wantKind: KindSpecification, wantOK: true},
		{name: "an adr identifier maps to the adr kind", id: "adr-0014", wantKind: KindADR, wantOK: true},
		{name: "a scenario identifier maps to the scenario kind", id: "scn-0001", wantKind: KindScenario, wantOK: true},
		{name: "a verification identifier maps to the verification kind", id: "ver-0008", wantKind: KindVerification, wantOK: true},
		{name: "a task identifier maps to the task kind", id: "task-0341", wantKind: KindTask, wantOK: true},
		{name: "an uppercase identifier maps to its kind", id: "CAP-0003", wantKind: KindCapability, wantOK: true},
		{name: "an unknown prefix maps to nothing", id: "e2e-0008", wantOK: false},
		{name: "an identifier without a prefix maps to nothing", id: "draft", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, ok := tt.id.MapToEntityKind()
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && kind != tt.wantKind {
				t.Errorf("got %q, expected %q", kind, tt.wantKind)
			}
		})
	}
}
