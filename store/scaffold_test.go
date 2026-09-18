package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/template"
)

func TestScaffold(t *testing.T) {
	t.Run("a capability is scaffolded from the template with its title set", func(t *testing.T) {
		root := t.TempDir()
		path, err := Scaffold(root, model.KindCapability, "Process payments")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "cap-0001-process-payments.md" {
			t.Errorf("got filename %q, expected cap-0001-process-payments.md", filepath.Base(path))
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		content := string(b)
		if !strings.HasPrefix(content, "# Process payments\n") {
			t.Errorf("expected the title to be set, got:\n%s", content)
		}
		if strings.Contains(content, "(optional)") {
			t.Errorf("expected optional markers to be stripped, got:\n%s", content)
		}
		if !strings.Contains(content, "## Actors") {
			t.Errorf("expected the optional Actors heading to remain, got:\n%s", content)
		}
	})

	t.Run("the next free number is allocated for the prefix", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "capabilities", "cap-0003-evaluate-policies.md", "# Evaluate policies\n")
		path, err := Scaffold(root, model.KindCapability, "Process payments")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "cap-0004-process-payments.md" {
			t.Errorf("got %q, expected cap-0004-process-payments.md", filepath.Base(path))
		}
	})

	t.Run("a concept uses the con prefix", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindConcept, "Policy")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "con-0001-policy.md" {
			t.Errorf("got %q, expected con-0001-policy.md", filepath.Base(path))
		}
	})

	t.Run("verification uses the ver prefix", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindVerification, "Pre-release smoke test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "ver-0001-pre-release-smoke-test.md" {
			t.Errorf("got %q, expected ver-0001-pre-release-smoke-test.md", filepath.Base(path))
		}
	})

	t.Run("a name with punctuation is slugified", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindTask, "Implement the policy cache (v2)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "task-0001-implement-the-policy-cache-v2.md" {
			t.Errorf("got %q, expected task-0001-implement-the-policy-cache-v2.md", filepath.Base(path))
		}
	})

	t.Run("a name that slugifies to a valid identifier uses the slug as the identifier", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindRequirement, "SOW-0023")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "req-sow-0023.md" {
			t.Errorf("got %q, expected req-sow-0023.md", filepath.Base(path))
		}
	})

	t.Run("a plain English name falls back to auto-numbering even for requirements", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindRequirement, "Provide T2O services")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "req-0001-provide-t2o-services.md" {
			t.Errorf("got %q, expected req-0001-provide-t2o-services.md", filepath.Base(path))
		}
	})

	t.Run("a slug-based identifier that already exists falls back to auto-numbering", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "requirements", "req-sow-0023.md", "# SOW-0023\n")
		path, err := Scaffold(root, model.KindRequirement, "SOW-0023")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "req-0001-sow-0023.md" {
			t.Errorf("got %q, expected req-0001-sow-0023.md", filepath.Base(path))
		}
	})

	t.Run("a JIRA-style name uses the slug as the identifier for any kind", func(t *testing.T) {
		path, err := Scaffold(t.TempDir(), model.KindTask, "PROJ-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Base(path) != "task-proj-123.md" {
			t.Errorf("got %q, expected task-proj-123.md", filepath.Base(path))
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("the directory layout and templates are created", func(t *testing.T) {
		root := t.TempDir()
		written, err := Init(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(written) == 0 {
			t.Errorf("expected templates to be installed")
		}
		for kind, dir := range DirForKind {
			if _, err := os.Stat(filepath.Join(root, dir)); err != nil {
				t.Errorf("expected directory for %s at %s: %v", kind, dir, err)
			}
		}
		if _, err := os.Stat(filepath.Join(root, template.Dir, "capability.md")); err != nil {
			t.Errorf("expected the capability template to be installed: %v", err)
		}
	})
}
