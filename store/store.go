// Package store loads cap entities from a Markdown-backed repository on disk.
//
// The on-disk layout groups entities by kind under the model root directory,
// "cap" by default:
//
//	cap/
//	├── contexts/
//	├── capabilities/
//	├── invariants/
//	├── specifications/
//	├── scenarios/
//	├── verification/
//	├── adrs/
//	└── tasks/
//
// Each file is a structured Markdown document whose identifier is derived from
// its filename.
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/cap/markdown"
	"github.com/a-h/cap/model"
	"github.com/a-h/cap/template"
)

// Well-known section headings used within entity documents. These names form the
// contract between authors and the parser, enforced by cap validate.
const (
	SectionMetadata       = "Metadata"
	SectionInvariants     = "Invariants"
	SectionSpecifications = "Specifications"
	SectionADRs           = "ADRs"
	SectionScenarios      = "Scenarios"
	SectionVerification   = "Verification"
	SectionTasks          = "Tasks"
	SectionCapabilities   = "Capabilities"
	SectionPaths          = "Paths"
)

// DirForKind maps an entity kind to its subdirectory beneath the system root.
var DirForKind = map[model.Kind]string{
	model.KindContext:       "contexts",
	model.KindCapability:    "capabilities",
	model.KindInvariant:     "invariants",
	model.KindSpecification: "specifications",
	model.KindADR:           "adrs",
	model.KindScenario:      "scenarios",
	model.KindVerification:  "verification",
	model.KindTask:          "tasks",
}

// idFromFilename matches an identifier at the start of a filename, for example
// "cap-0003" in "cap-0003-evaluate-policies.md". The prefix may be any case and
// the numeric part may be unpadded; the captured identifier is canonicalised by
// the caller. An optional descriptive slug follows the number.
var idFromFilename = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9]*-[0-9]+)(?:-.*)?$`)

// Problem is a single finding raised while loading or validating the model.
type Problem struct {
	File     string
	Line     int
	Severity Severity
	Message  string
}

// Severity classifies a Problem. Errors prevent an entity from being understood.
// Warnings indicate a stale or mistyped reference in an otherwise usable entity.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

func (p Problem) String() string {
	loc := p.File
	if p.Line > 0 {
		loc = fmt.Sprintf("%s:%d", p.File, p.Line)
	}
	return fmt.Sprintf("%s: %s: %s", loc, p.Severity, p.Message)
}

// LoadResult holds the loaded model together with any problems found while
// reading and parsing entity files. Loading is tolerant: a model is returned
// even when problems are present.
type LoadResult struct {
	Model    *model.Model
	Files    map[model.ID]string
	Problems []Problem
}

// ParseID derives an entity identifier from a file path. The identifier is the
// leading prefixed, numbered token of the base filename, before the descriptive
// slug, returned in canonical form. It reports ok=false when the filename does not
// begin with an identifier.
func ParseID(path string) (id model.ID, ok bool) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := idFromFilename.FindStringSubmatch(base)
	if m == nil {
		return "", false
	}
	return model.ID(m[1]).Canonical(), true
}

// Load reads every entity beneath the system root directory and returns the
// assembled model and any problems found. It does not resolve cross-entity
// references; see the validate package for reference checking.
func Load(root string) (LoadResult, error) {
	res := LoadResult{Model: model.NewModel(), Files: map[model.ID]string{}}
	for kind, dir := range DirForKind {
		path := filepath.Join(root, dir)
		entries, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return res, fmt.Errorf("store: reading %s: %w", path, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				continue
			}
			file := filepath.Join(path, entry.Name())
			res.loadFile(root, kind, file)
		}
	}
	return res, nil
}

func (res *LoadResult) loadFile(root string, kind model.Kind, file string) {
	id, ok := ParseID(file)
	if !ok {
		res.add(Problem{File: file, Severity: SeverityError, Message: "filename does not begin with a valid identifier"})
		return
	}
	f, err := os.Open(file)
	if err != nil {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("cannot open file: %v", err)})
		return
	}
	defer f.Close()
	doc, err := markdown.Parse(f)
	if err != nil {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("cannot parse document: %v", err)})
		return
	}
	if k, known := id.MapToEntityKind(); known && k != kind {
		res.add(Problem{File: file, Severity: SeverityWarning, Message: fmt.Sprintf("identifier %s has prefix for %s but is stored under %s", id, k, kind)})
	}
	if existingKind, exists := res.Model.Lookup(id); exists {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("duplicate identifier %s (already loaded as %s)", id, existingKind)})
		return
	}

	if !res.build(kind, id, file, doc) {
		return
	}
	res.checkRequiredSections(root, kind, file, doc)
	res.Files[id] = file
}

// checkRequiredSections warns when a required section from the kind's template
// schema is missing or empty in the document. These are warnings; cap validate
// promotes them to errors under --strict.
func (res *LoadResult) checkRequiredSections(root string, kind model.Kind, file string, doc markdown.Document) {
	sections, ok := template.ParseSchema(root, kind)
	if !ok {
		return
	}
	for _, want := range sections {
		if !want.Required {
			continue
		}
		sec, present := doc.FindSection(want.Title)
		if !present {
			res.add(Problem{File: file, Severity: SeverityWarning, Message: fmt.Sprintf("required section %q is missing", want.Title)})
			continue
		}
		if !sec.HasContent {
			res.add(Problem{File: file, Line: sec.Line, Severity: SeverityWarning, Message: fmt.Sprintf("required section %q is empty", want.Title)})
		}
	}
}

func (res *LoadResult) add(p Problem) {
	res.Problems = append(res.Problems, p)
}
