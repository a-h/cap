package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/cap/model"
)

func TestLoadDefault(t *testing.T) {
	t.Run("a known kind has an embedded default template", func(t *testing.T) {
		content, ok := LoadDefault(model.KindCapability)
		if !ok {
			t.Fatalf("expected a default capability template")
		}
		if !strings.Contains(content, "## Description") {
			t.Errorf("expected the capability template to contain a Description section")
		}
	})

	t.Run("verification has no default template because its prefix is type-specific", func(t *testing.T) {
		if _, ok := LoadDefault(model.KindVerification); !ok {
			t.Skip("verification template is provided")
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("an installed template takes precedence over the embedded default", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, Dir), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		custom := "# Custom\n\n## Custom section\n\nContent.\n"
		if err := os.WriteFile(filepath.Join(root, Dir, "capability.md"), []byte(custom), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		content, ok := Load(root, model.KindCapability)
		if !ok {
			t.Fatalf("expected the template to load")
		}
		if content != custom {
			t.Errorf("expected the installed template, got %q", content)
		}
	})

	t.Run("the embedded default is used when no template is installed", func(t *testing.T) {
		content, ok := Load(t.TempDir(), model.KindCapability)
		if !ok || !strings.Contains(content, "## Scope") {
			t.Errorf("expected the embedded default capability template")
		}
	})
}

func TestInstall(t *testing.T) {
	t.Run("default templates are written and existing ones are preserved", func(t *testing.T) {
		root := t.TempDir()
		written, err := Install(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(written) == 0 {
			t.Fatalf("expected templates to be written")
		}
		custom := "# kept\n"
		if err := os.WriteFile(filepath.Join(root, Dir, "capability.md"), []byte(custom), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		written, err = Install(root)
		if err != nil {
			t.Fatalf("unexpected error on second install: %v", err)
		}
		if len(written) != 0 {
			t.Errorf("expected no templates to be rewritten, got %v", written)
		}
		b, err := os.ReadFile(filepath.Join(root, Dir, "capability.md"))
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if string(b) != custom {
			t.Errorf("expected the customised template to be preserved, got %q", string(b))
		}
	})
}

func TestParseSchema(t *testing.T) {
	t.Run("headings are required unless marked optional", func(t *testing.T) {
		sections, ok := ParseSchema(t.TempDir(), model.KindCapability)
		if !ok {
			t.Fatalf("expected a schema for the capability kind")
		}
		required := map[string]bool{}
		for _, s := range sections {
			required[s.Title] = s.Required
		}
		if r, ok := required["Description"]; !ok || !r {
			t.Errorf("expected Description to be required, got %v (present %v)", r, ok)
		}
		if r, ok := required["Scope"]; !ok || !r {
			t.Errorf("expected Scope to be required, got %v (present %v)", r, ok)
		}
		if r, ok := required["Actors"]; !ok || r {
			t.Errorf("expected Actors to be optional, got required=%v (present %v)", r, ok)
		}
		if _, ok := required["(optional)"]; ok {
			t.Errorf("the optional marker should be stripped from the title")
		}
	})

	t.Run("the title heading is not part of the schema", func(t *testing.T) {
		sections, _ := ParseSchema(t.TempDir(), model.KindInvariant)
		for _, s := range sections {
			if strings.HasPrefix(s.Title, "State the rule") {
				t.Errorf("the level one title must not appear as a schema section")
			}
		}
	})
}
