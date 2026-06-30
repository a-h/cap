package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI drives the CLI in-process with the model root set to dir, returning the
// captured output and the command error. Tests assert on these directly rather than
// shelling out to a built binary.
func runCLI(t *testing.T, dir string, args ...string) (output string, err error) {
	t.Helper()
	var buf bytes.Buffer
	err = run(append(args, "--root", filepath.Join(dir, "cap")), &buf)
	return buf.String(), err
}

// initProject creates a temporary project with templates installed and returns its
// root directory.
func initProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := runCLI(t, dir, "init"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	return dir
}

func TestInitInstallsLayoutAndTemplates(t *testing.T) {
	dir := initProject(t)
	for _, sub := range []string{"capabilities", "invariants", ".templates"} {
		if _, err := os.Stat(filepath.Join(dir, "cap", sub)); err != nil {
			t.Errorf("expected cap/%s to exist: %v", sub, err)
		}
	}
}

func TestNewScaffoldsAnEntity(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	path := filepath.Join(dir, "cap", "capabilities", "cap-0001-evaluate-policies.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected the scaffolded file: %v", err)
	}
	if !strings.HasPrefix(string(b), "# Evaluate policies\n") {
		t.Errorf("expected the title to be set, got:\n%s", string(b))
	}
}

func TestValidateSucceedsForAScaffoldedModel(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	if _, err := runCLI(t, dir, "validate"); err != nil {
		t.Errorf("expected validate to succeed for a scaffolded model, got: %v", err)
	}
}

func TestValidateWarnsButSucceedsForADanglingReference(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-thing.md",
		"# Thing\n\n## Description\n\nA thing.\n\n## Scope\n\nIn scope:\n\n- x\n\n## Invariants\n\n- inv-9999\n")
	out, err := runCLI(t, dir, "validate")
	if err != nil {
		t.Errorf("expected a dangling reference to warn but not fail, got: %v", err)
	}
	if !strings.Contains(out, "inv-9999") || !strings.Contains(out, "warning") {
		t.Errorf("expected a warning naming inv-9999, got:\n%s", out)
	}
}

func TestValidateStrictFailsOnAWarning(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-thing.md",
		"# Thing\n\n## Description\n\nA thing.\n\n## Scope\n\nIn scope:\n\n- x\n\n## Invariants\n\n- inv-9999\n")
	if _, err := runCLI(t, dir, "validate", "--strict"); err == nil {
		t.Errorf("expected --strict to fail on a warning")
	}
}

func TestContextEmitsAJSONBundle(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	out, err := runCLI(t, dir, "context", "cap-0001", "--format", "json")
	if err != nil {
		t.Fatalf("context failed: %v", err)
	}
	var bundle struct {
		Capability struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"capability"`
	}
	if err := json.Unmarshal([]byte(out), &bundle); err != nil {
		t.Fatalf("expected valid JSON, got error %v from:\n%s", err, out)
	}
	if bundle.Capability.ID != "cap-0001" || bundle.Capability.Name != "Evaluate policies" {
		t.Errorf("unexpected bundle: %+v", bundle.Capability)
	}
}

func TestContextAcceptsNonCanonicalIdentifiers(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	if _, err := runCLI(t, dir, "context", "CAP-1"); err != nil {
		t.Errorf("expected CAP-1 to resolve to cap-0001, got: %v", err)
	}
}

func TestContextFailsForAnUnknownCapability(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "context", "cap-9999"); err == nil {
		t.Errorf("expected a missing capability to fail")
	}
}

func TestKindListListsEntities(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	if _, err := runCLI(t, dir, "new", "capability", "Process payments"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	out, err := runCLI(t, dir, "list", "capabilities")
	if err != nil {
		t.Fatalf("list capabilities failed: %v", err)
	}
	if !strings.Contains(out, "cap-0001  Evaluate policies") || !strings.Contains(out, "cap-0002  Process payments") {
		t.Errorf("expected both capabilities listed, got:\n%s", out)
	}
}

func TestShowReportsLinksInBothDirections(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n\n## Invariants\n\n- inv-0001\n")
	writeEntity(t, dir, "invariants", "inv-0001-consistency.md",
		"# Policies must be evaluated consistently.\n\n## Description\n\nx\n")
	out, err := runCLI(t, dir, "show", "inv-0001")
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(out, "Linked by:") || !strings.Contains(out, "cap-0001") {
		t.Errorf("expected inv-0001 to be linked by cap-0001, got:\n%s", out)
	}
}

func TestGraphPrintsATopDownTree(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n")
	writeEntity(t, dir, "scenarios", "scn-0001-claim.md",
		"# Claim approval\n\n## Description\n\nx\n\n## Steps\n\n- x\n\n## Capabilities\n\n- cap-0001\n")
	out, err := runCLI(t, dir, "graph", "scn-0001")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if !strings.Contains(out, "scn-0001  Claim approval") || !strings.Contains(out, "cap-0001  Evaluate policies") {
		t.Errorf("expected a tree from scenario to capability, got:\n%s", out)
	}
}

func TestGraphWithoutAnIDGraphsTheWholeModelFromContexts(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "contexts", "ctx-0001-claims.md",
		"# Claims\n\n## Description\n\nx\n")
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Metadata\n\n- context: ctx-0001\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n")
	writeEntity(t, dir, "scenarios", "scn-0001-claim.md",
		"# Claim approval\n\n## Description\n\nx\n\n## Steps\n\n- x\n\n## Capabilities\n\n- cap-0001\n")
	out, err := runCLI(t, dir, "graph")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if !strings.Contains(out, "ctx-0001  Claims") {
		t.Errorf("expected the context to be a root, got:\n%s", out)
	}
	if !strings.Contains(out, "cap-0001  Evaluate policies") {
		t.Errorf("expected the capability beneath its context, got:\n%s", out)
	}
	if !strings.Contains(out, "scn-0001  Claim approval") {
		t.Errorf("expected the scenario as a top-level root, got:\n%s", out)
	}
}

func TestGraphFormatDOTEmitsADirectedGraph(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n")
	writeEntity(t, dir, "scenarios", "scn-0001-claim.md",
		"# Claim approval\n\n## Description\n\nx\n\n## Steps\n\n- x\n\n## Capabilities\n\n- cap-0001\n")
	out, err := runCLI(t, dir, "graph", "--format", "dot")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if !strings.HasPrefix(out, "digraph cap {") {
		t.Errorf("expected a DOT digraph, got:\n%s", out)
	}
	if !strings.Contains(out, `"scn-0001" -> "cap-0001";`) {
		t.Errorf("expected an edge from the scenario to its capability, got:\n%s", out)
	}
}

func TestGraphExcludeOmitsAKind(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n\n## Verification\n\n- tests/policy_test.go\n")
	out, err := runCLI(t, dir, "graph", "cap-0001", "--exclude", "verification")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if strings.Contains(out, "policy_test.go") {
		t.Errorf("expected verification to be excluded, got:\n%s", out)
	}
}

func TestGraphDepthLimitsExpansion(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-evaluate.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n\n## Verification\n\n- tests/policy_test.go\n")
	writeEntity(t, dir, "scenarios", "scn-0001-claim.md",
		"# Claim approval\n\n## Description\n\nx\n\n## Steps\n\n- x\n\n## Capabilities\n\n- cap-0001\n")
	out, err := runCLI(t, dir, "graph", "scn-0001", "--depth", "1")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if !strings.Contains(out, "cap-0001  Evaluate policies") {
		t.Errorf("expected the capability at depth 1, got:\n%s", out)
	}
	if strings.Contains(out, "policy_test.go") {
		t.Errorf("expected the verification, two links deep, to be trimmed, got:\n%s", out)
	}
}

func TestGraphExcludeRejectsAnUnknownKind(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "graph", "--exclude", "wibble"); err == nil {
		t.Error("expected an error for an unknown kind")
	}
}

func TestValidateWarnsAboutCapabilitiesWithoutVerification(t *testing.T) {
	dir := initProject(t)
	writeEntity(t, dir, "capabilities", "cap-0001-untested.md",
		"# Authenticate users\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n")
	writeEntity(t, dir, "capabilities", "cap-0002-tested.md",
		"# Evaluate policies\n\n## Description\n\nx\n\n## Scope\n\nIn scope:\n\n- x\n\n## Verification\n\n- internal/policy/eval_test.go\n")
	out, err := runCLI(t, dir, "validate")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !strings.Contains(out, "cap-0001 has no verification") {
		t.Errorf("expected a no-verification warning for cap-0001, got:\n%s", out)
	}
	if strings.Contains(out, "cap-0002 has no verification") {
		t.Errorf("did not expect a warning for the tested capability, got:\n%s", out)
	}
}

func TestReviewEmitsAPacketWithAChecklist(t *testing.T) {
	dir := initProject(t)
	if _, err := runCLI(t, dir, "new", "capability", "Evaluate policies"); err != nil {
		t.Fatalf("new failed: %v", err)
	}
	out, err := runCLI(t, dir, "review", "cap-0001")
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if !strings.Contains(out, "# Review: cap-0001 (capability)") {
		t.Errorf("expected a review header, got:\n%s", out)
	}
	if !strings.Contains(out, "## Checklist") || !strings.Contains(out, "verb plus noun") {
		t.Errorf("expected a capability checklist, got:\n%s", out)
	}
}

func TestVersionWritesTheBuildVersion(t *testing.T) {
	var buf bytes.Buffer
	if err := run([]string{"version"}, &buf); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if buf.String() != Version {
		t.Errorf("expected %q, got %q", Version, buf.String())
	}
}

func TestRunWithoutArgumentsPrintsHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := run(nil, &buf); err != nil {
		t.Fatalf("run without arguments failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Usage: cap <command>") {
		t.Errorf("expected usage text, got %q", out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Errorf("expected the command list, got %q", out)
	}
}

// writeEntity writes a Markdown entity file directly into the model layout, for
// tests that need a specific document rather than a scaffolded one.
func writeEntity(t *testing.T, dir, kindDir, name, content string) {
	t.Helper()
	target := filepath.Join(dir, "cap", kindDir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
