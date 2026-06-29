package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

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

func TestAssemble(t *testing.T) {
	t.Run("an unknown identifier is not found", func(t *testing.T) {
		if _, ok := Assemble(store.LoadResult{Model: model.NewModel(), Files: map[model.ID]string{}}, "cap-9999"); ok {
			t.Errorf("expected ok=false for an unknown identifier")
		}
	})

	t.Run("a capability packet includes its content, linked context, and a capability checklist", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\nIt evaluates policies.\n\n## Scope\n\nIn scope:\n\n- Evaluation.\n")
		writeFile(t, root, "invariants", "inv-0001-consistency.md", "# Policies must be evaluated consistently.\n")
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\nIt evaluates policies.\n\n## Scope\n\nIn scope:\n\n- Evaluation.\n\n## Invariants\n\n- inv-0001\n")

		res, err := store.Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		p, ok := Assemble(res, "cap-0003")
		if !ok {
			t.Fatalf("expected the capability to be found")
		}
		if p.Kind != model.KindCapability {
			t.Errorf("got kind %q, expected capability", p.Kind)
		}
		if !strings.Contains(p.Content, "It evaluates policies.") {
			t.Errorf("expected the packet to contain the file content, got:\n%s", p.Content)
		}
		if p.Context == nil {
			t.Fatalf("expected a capability packet to carry linked context")
		}
		if len(p.Context.Invariants) != 1 {
			t.Errorf("expected the linked invariant to be resolved, got %v", p.Context.Invariants)
		}
		if len(p.Checklist) == 0 || !strings.Contains(strings.Join(p.Checklist, "\n"), "verb plus noun") {
			t.Errorf("expected a capability checklist, got %v", p.Checklist)
		}
	})

	t.Run("a non-capability packet carries no linked context", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "invariants", "inv-0001-consistency.md", "# Policies must be evaluated consistently.\n\n## Description\n\nThe invariant.\n")
		res, err := store.Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		p, ok := Assemble(res, "inv-0001")
		if !ok {
			t.Fatalf("expected the invariant to be found")
		}
		if p.Context != nil {
			t.Errorf("expected no linked context for an invariant, got %v", p.Context)
		}
		if !strings.Contains(strings.Join(p.Checklist, "\n"), "invariant") {
			t.Errorf("expected an invariant-specific checklist, got %v", p.Checklist)
		}
	})
}

func TestAssembleAll(t *testing.T) {
	t.Run("a packet is produced for every loaded entity, ordered by identifier", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n\n## Description\n\nIt evaluates.\n\n## Scope\n\nIn scope:\n\n- x\n")
		writeFile(t, root, "contexts", "ctx-0001-policy.md", "# Policy enforcement\n\n## Description\n\nThe context.\n")
		res, err := store.Load(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		packets := AssembleAll(res)
		if len(packets) != 2 {
			t.Fatalf("expected 2 packets, got %d", len(packets))
		}
		if packets[0].ID != "cap-0003" || packets[1].ID != "ctx-0001" {
			t.Errorf("expected packets ordered by identifier, got %s then %s", packets[0].ID, packets[1].ID)
		}
	})
}

func TestRender(t *testing.T) {
	t.Run("the rendered packet names the entity, its content, and its checklist", func(t *testing.T) {
		p := Packet{
			ID:        "cap-0003",
			Kind:      model.KindCapability,
			File:      "cap/capabilities/cap-0003-evaluate-policies.md",
			Content:   "# Evaluate policies\n\n## Description\n\nIt evaluates policies.",
			Checklist: []string{"Is the name an imperative verb plus noun?"},
		}
		got := p.Render()
		for _, want := range []string{
			"# Review: cap-0003 (capability)",
			"Source: cap/capabilities/cap-0003-evaluate-policies.md",
			"It evaluates policies.",
			"## Checklist",
			"- Is the name an imperative verb plus noun?",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("expected the rendered packet to contain %q, got:\n%s", want, got)
			}
		}
	})
}
