package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/a-h/cap/model"
	"github.com/a-h/cap/template"
)

// prefixForKind maps an entity kind to its canonical identifier prefix.
var prefixForKind = map[model.Kind]string{
	model.KindContext:        "ctx",
	model.KindConcept:        "con",
	model.KindCapability:     "cap",
	model.KindInvariant:      "inv",
	model.KindSpecification:  "spec",
	model.KindADR:            "adr",
	model.KindScenario:       "scn",
	model.KindVerification:   "ver",
	model.KindTask:           "task",
	model.KindService:        "svc",
	model.KindExternalSystem: "ext",
	model.KindRequirement:    "req",
}

// optionalHeading matches a heading whose title ends with the optional marker, so
// the marker can be stripped when scaffolding.
var optionalHeading = regexp.MustCompile(`(?i)^(#+\s+.*?)\s*\(optional\)\s*$`)

// Scaffold creates a new entity file of the given kind from its template. The name
// is used as the document title. When the slugified name forms a valid identifier
// (for example "SOW-0023" slugifies to "sow-0023", giving "req-sow-0023") and no
// file with that name already exists, it is used as the identifier directly. This
// lets external references such as SOW numbers or JIRA tickets become the entity's
// identifier without a separate auto-number. When the slug does not form a valid
// identifier, or the file already exists, the next free number for the prefix is
// allocated and the slug is appended as a descriptive suffix. It returns the path
// written.
func Scaffold(root string, kind model.Kind, name string) (path string, err error) {
	if kind == model.KindADR {
		if external, ok := resolveADRDir(root); ok {
			return "", fmt.Errorf("store: ADRs are managed by adr-tools in %s; create one with 'adr new %q'", external, name)
		}
	}
	dir, ok := DirForKind[kind]
	if !ok {
		return "", fmt.Errorf("store: no directory for kind %q", kind)
	}
	prefix, ok := prefixForKind[kind]
	if !ok {
		return "", fmt.Errorf("store: no identifier prefix for kind %q", kind)
	}

	content, ok := template.Load(root, kind)
	if !ok {
		return "", fmt.Errorf("store: no template for kind %q", kind)
	}

	slug := slugify(name)
	var filename string
	if slug != "" {
		candidate := string(model.ID(prefix + "-" + slug).Canonical())
		if _, ok := ParseID(candidate + ".md"); ok {
			if _, err := os.Stat(filepath.Join(root, dir, candidate+".md")); os.IsNotExist(err) {
				filename = candidate
			}
		}
	}
	if filename == "" {
		id := model.ID(fmt.Sprintf("%s-%d", prefix, findNextNumber(root, dir, prefix))).Canonical()
		filename = string(id)
		if slug != "" {
			filename += "-" + slug
		}
	}
	path = filepath.Join(root, dir, filename+".md")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("store: %s already exists", path)
	}

	if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
		return "", fmt.Errorf("store: creating %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(renderTemplate(content, name)), 0o644); err != nil {
		return "", fmt.Errorf("store: writing %s: %w", path, err)
	}
	return path, nil
}

// findNextNumber returns the next free identifier number for a prefix within a
// directory, by scanning existing filenames. Numbering starts at 1.
func findNextNumber(root, dir, prefix string) int {
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		return 1
	}
	highest := 0
	for _, entry := range entries {
		id, ok := ParseID(entry.Name())
		if !ok {
			continue
		}
		p, ok := id.Prefix()
		if !ok || p != prefix {
			continue
		}
		if n := parseNumber(id); n > highest {
			highest = n
		}
	}
	return highest + 1
}

// parseNumber returns the numeric part of a simple prefix-N identifier (for example
// 3 from "req-0003"), or zero when the identifier is compound (for example
// "req-sow-0023") or has no numeric part. Compound identifiers are excluded so they
// do not advance the auto-number counter when a directory contains a mix of
// auto-numbered and slug-based identifiers.
func parseNumber(id model.ID) int {
	s := string(id)
	if strings.Count(s, "-") != 1 {
		return 0
	}
	i := strings.Index(s, "-")
	n, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return 0
	}
	return n
}

// renderTemplate produces the scaffolded document: the title line is replaced with
// the given name, and the optional marker is stripped from headings so the authored
// file has clean headings.
func renderTemplate(content, name string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = "# " + name
			continue
		}
		if m := optionalHeading.FindStringSubmatch(line); m != nil {
			lines[i] = m[1]
		}
	}
	return strings.Join(lines, "\n")
}

// slugify converts a name to a lowercase, hyphen-separated slug.
func slugify(name string) string {
	var b strings.Builder
	var lastHyphen bool
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// Init creates the system directory layout beneath root and installs the default
// templates. It returns the template filenames that were written.
func Init(root string) (templatesWritten []string, err error) {
	for _, dir := range DirForKind {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return nil, fmt.Errorf("store: creating %s: %w", dir, err)
		}
	}
	return template.Install(root)
}
