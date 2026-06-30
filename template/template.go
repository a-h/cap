// Package template provides the per-kind content templates that define an entity's
// schema.
//
// A template is an ordinary entity document with guidance content under each
// heading. It is both the authoring skeleton scaffolded by cap new and the schema
// that cap validate enforces: every heading is a required section unless its title
// ends with the optional marker.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/cap/markdown"
	"github.com/a-h/cap/model"
)

//go:embed defaults/*.md
var defaults embed.FS

// Dir is the directory beneath the system root where templates are installed.
const Dir = ".templates"

// optionalMarker, when it ends a heading title, marks the section as optional.
const optionalMarker = "(optional)"

// fileForKind maps an entity kind to its template filename.
var fileForKind = map[model.Kind]string{
	model.KindContext:       "context.md",
	model.KindConcept:       "concept.md",
	model.KindCapability:    "capability.md",
	model.KindInvariant:     "invariant.md",
	model.KindSpecification: "specification.md",
	model.KindADR:           "adr.md",
	model.KindScenario:      "scenario.md",
	model.KindVerification:  "verification.md",
	model.KindTask:          "task.md",
}

// LoadDefault returns the embedded default template for a kind. It reports ok=false
// when the kind has no template.
func LoadDefault(kind model.Kind) (content string, ok bool) {
	name, ok := fileForKind[kind]
	if !ok {
		return "", false
	}
	b, err := defaults.ReadFile(filepath.Join("defaults", name))
	if err != nil {
		return "", false
	}
	return string(b), true
}

// Load returns the template for a kind, preferring the project's installed template
// under root/.templates and falling back to the embedded default. It reports
// ok=false when the kind has no template.
func Load(root string, kind model.Kind) (content string, ok bool) {
	name, ok := fileForKind[kind]
	if !ok {
		return "", false
	}
	if b, err := os.ReadFile(filepath.Join(root, Dir, name)); err == nil {
		return string(b), true
	}
	return LoadDefault(kind)
}

// Install writes the embedded default templates to root/.templates, skipping any
// that already exist so a project's customisations are preserved. It returns the
// names of the templates it wrote.
func Install(root string) (written []string, err error) {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("template: creating %s: %w", dir, err)
	}
	entries, err := fs.ReadDir(defaults, "defaults")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		dest := filepath.Join(dir, entry.Name())
		if _, err := os.Stat(dest); err == nil {
			continue
		}
		b, err := defaults.ReadFile(filepath.Join("defaults", entry.Name()))
		if err != nil {
			return written, err
		}
		if err := os.WriteFile(dest, b, 0o644); err != nil {
			return written, fmt.Errorf("template: writing %s: %w", dest, err)
		}
		written = append(written, entry.Name())
	}
	return written, nil
}

// Section describes a heading required or permitted by a template.
type Section struct {
	Title    string
	Required bool
}

// ParseSchema derives the section schema for a kind from its template. Every heading in
// the template is a required section unless its title ends with the optional
// marker, in which case it is optional and the marker is stripped from the title.
// It reports ok=false when the kind has no template.
func ParseSchema(root string, kind model.Kind) (sections []Section, ok bool) {
	content, ok := Load(root, kind)
	if !ok {
		return nil, false
	}
	doc, err := markdown.Parse(strings.NewReader(content))
	if err != nil {
		return nil, false
	}
	for _, sec := range doc.Sections {
		if sec.Level != 2 {
			continue
		}
		title := sec.Title
		required := true
		if strings.HasSuffix(strings.ToLower(title), optionalMarker) {
			title = strings.TrimSpace(title[:len(title)-len(optionalMarker)])
			required = false
		}
		sections = append(sections, Section{Title: title, Required: required})
	}
	return sections, true
}
