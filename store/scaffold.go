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
	model.KindContext:       "ctx",
	model.KindConcept:       "con",
	model.KindCapability:    "cap",
	model.KindInvariant:     "inv",
	model.KindSpecification: "spec",
	model.KindADR:           "adr",
	model.KindScenario:      "scn",
	model.KindVerification:  "ver",
	model.KindTask:          "task",
}

// optionalHeading matches a heading whose title ends with the optional marker, so
// the marker can be stripped when scaffolding.
var optionalHeading = regexp.MustCompile(`(?i)^(#+\s+.*?)\s*\(optional\)\s*$`)

// Scaffold creates a new entity file of the given kind from its template. The name
// is used both as the document title and, slugified, as the descriptive part of
// the filename. The next free number for the kind's identifier prefix is allocated.
// It returns the path written.
func Scaffold(root string, kind model.Kind, name string) (path string, err error) {
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

	id := model.ID(fmt.Sprintf("%s-%d", prefix, findNextNumber(root, dir, prefix))).Canonical()
	filename := string(id)
	if slug := slugify(name); slug != "" {
		filename += "-" + slug
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

// parseNumber returns the numeric part of a canonical identifier, or zero when it has
// none.
func parseNumber(id model.ID) int {
	i := strings.LastIndex(string(id), "-")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(string(id)[i+1:])
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
