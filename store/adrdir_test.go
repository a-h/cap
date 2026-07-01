package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/a-h/cap/model"
)

func TestParseADRToolsID(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		wantID model.ID
		wantOK bool
	}{
		{name: "an adr-tools filename maps to a canonical adr identifier", path: "doc/adr/0001-record-architecture-decisions.md", wantID: "adr-0001", wantOK: true},
		{name: "an unpadded number is zero-padded", path: "0014-engine.md", wantID: "adr-0014", wantOK: true},
		{name: "a number with no slug is accepted", path: "0007.md", wantID: "adr-0007", wantOK: true},
		{name: "a cap-style adr prefix is accepted", path: "adr-0003-policy.md", wantID: "adr-0003", wantOK: true},
		{name: "a filename without a leading number is rejected", path: "README.md", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := parseADRToolsID(tt.path)
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && id != tt.wantID {
				t.Errorf("got id %q, expected %q", id, tt.wantID)
			}
		})
	}
}

func TestResolveADRDir(t *testing.T) {
	t.Run("a config in the model root names a directory relative to that root", func(t *testing.T) {
		root := t.TempDir()
		mustWrite(t, filepath.Join(root, ".adr-dir"), "doc/adr")
		mustMkdir(t, filepath.Join(root, "doc", "adr"))

		dir, ok := resolveADRDir(root)
		if !ok {
			t.Fatalf("expected the directory to resolve")
		}
		if want := filepath.Join(root, "doc", "adr"); dir != want {
			t.Errorf("got %q, expected %q", dir, want)
		}
	})

	t.Run("a config in a parent of the model root is found by walking up", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "doc/adr")
		mustMkdir(t, filepath.Join(repo, "doc", "adr"))
		modelRoot := filepath.Join(repo, "cap")
		mustMkdir(t, modelRoot)

		dir, ok := resolveADRDir(modelRoot)
		if !ok {
			t.Fatalf("expected the directory to resolve")
		}
		if want := filepath.Join(repo, "doc", "adr"); dir != want {
			t.Errorf("got %q, expected %q", dir, want)
		}
	})

	t.Run("an absolute path in the config is used as written", func(t *testing.T) {
		root := t.TempDir()
		external := t.TempDir()
		mustWrite(t, filepath.Join(root, ".adr-dir"), external)

		dir, ok := resolveADRDir(root)
		if !ok {
			t.Fatalf("expected the directory to resolve")
		}
		if dir != external {
			t.Errorf("got %q, expected %q", dir, external)
		}
	})

	t.Run("a conventional doc/adr directory is used when no config is present", func(t *testing.T) {
		repo := t.TempDir()
		mustMkdir(t, filepath.Join(repo, "doc", "adr"))
		modelRoot := filepath.Join(repo, "cap")
		mustMkdir(t, modelRoot)

		dir, ok := resolveADRDir(modelRoot)
		if !ok {
			t.Fatalf("expected the conventional directory to resolve")
		}
		if want := filepath.Join(repo, "doc", "adr"); dir != want {
			t.Errorf("got %q, expected %q", dir, want)
		}
	})

	t.Run("a config takes precedence over the conventional doc/adr directory", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "decisions")
		mustMkdir(t, filepath.Join(repo, "decisions"))
		mustMkdir(t, filepath.Join(repo, "doc", "adr"))

		dir, ok := resolveADRDir(repo)
		if !ok {
			t.Fatalf("expected the configured directory to resolve")
		}
		if want := filepath.Join(repo, "decisions"); dir != want {
			t.Errorf("got %q, expected %q", dir, want)
		}
	})

	t.Run("a config naming a missing directory does not resolve", func(t *testing.T) {
		root := t.TempDir()
		mustWrite(t, filepath.Join(root, ".adr-dir"), "doc/adr")

		if _, ok := resolveADRDir(root); ok {
			t.Errorf("expected no directory to resolve when the named directory is absent")
		}
	})

	t.Run("no config resolves to no directory", func(t *testing.T) {
		if _, ok := resolveADRDir(t.TempDir()); ok {
			t.Errorf("expected no directory to resolve without a config")
		}
	})
}

func TestLoadExternalADRs(t *testing.T) {
	t.Run("ADRs are loaded from the adr-tools directory named by the config", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "doc/adr")
		adrDir := filepath.Join(repo, "doc", "adr")
		mustMkdir(t, adrDir)
		mustWrite(t, filepath.Join(adrDir, "0001-record-architecture-decisions.md"), "# Record architecture decisions\n")
		mustWrite(t, filepath.Join(adrDir, "0002-policy-evaluation-engine.md"), "# Policy evaluation engine\n")

		modelRoot := filepath.Join(repo, "cap")
		mustMkdir(t, modelRoot)

		res, err := Load(modelRoot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, p := range res.Problems {
			if p.Severity == SeverityError {
				t.Fatalf("expected no errors, got %v", res.Problems)
			}
		}
		if a, ok := res.Model.ADRs["adr-0001"]; !ok || a.Title != "Record architecture decisions" {
			t.Errorf("adr-0001 not loaded correctly: %#v", a)
		}
		if a, ok := res.Model.ADRs["adr-0002"]; !ok || a.Title != "Policy evaluation engine" {
			t.Errorf("adr-0002 not loaded correctly: %#v", a)
		}
		if got := res.Files["adr-0001"]; got != filepath.Join(adrDir, "0001-record-architecture-decisions.md") {
			t.Errorf("adr-0001 file path not recorded correctly: %q", got)
		}
	})

	t.Run("when an external ADR directory is configured, the internal adrs directory is not read", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "doc/adr")
		mustMkdir(t, filepath.Join(repo, "doc", "adr"))

		modelRoot := filepath.Join(repo, "cap")
		writeFile(t, modelRoot, "adrs", "adr-0099-internal.md", "# Internal decision\n")

		res, err := Load(modelRoot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := res.Model.ADRs["adr-0099"]; ok {
			t.Errorf("expected the internal adrs directory to be ignored when an external one is configured")
		}
	})

	t.Run("files without a leading number are ignored", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "doc/adr")
		adrDir := filepath.Join(repo, "doc", "adr")
		mustMkdir(t, adrDir)
		mustWrite(t, filepath.Join(adrDir, "README.md"), "# How to write ADRs\n")
		mustWrite(t, filepath.Join(adrDir, "0001-first.md"), "# First decision\n")

		res, err := Load(filepath.Join(repo, "cap"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Model.ADRs) != 1 {
			t.Errorf("expected only the numbered ADR to be loaded, got %d", len(res.Model.ADRs))
		}
	})
}

func TestScaffoldADRWithExternalDir(t *testing.T) {
	t.Run("scaffolding an ADR defers to adr-tools when an external directory is configured", func(t *testing.T) {
		repo := t.TempDir()
		mustWrite(t, filepath.Join(repo, ".adr-dir"), "doc/adr")
		mustMkdir(t, filepath.Join(repo, "doc", "adr"))
		modelRoot := filepath.Join(repo, "cap")
		mustMkdir(t, modelRoot)

		if _, err := Scaffold(modelRoot, model.KindADR, "A decision"); err == nil {
			t.Fatalf("expected an error directing the user to adr-tools")
		}
	})
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}
