package store

import (
	"fmt"
	"strings"

	"github.com/a-h/cap/markdown"
	"github.com/a-h/cap/model"
)

// build constructs the entity of the given kind from a parsed document and inserts
// it into the model, recording problems on res. It reports ok=false when the kind
// is not recognised, in which case nothing is inserted.
func (res *LoadResult) build(kind model.Kind, id model.ID, file string, doc markdown.Document) (ok bool) {
	m := res.Model
	meta := res.parseMetadata(file, doc)
	switch kind {
	case model.KindContext:
		m.Contexts[id] = model.Context{ID: id, Name: doc.Title}
	case model.KindConcept:
		m.Concepts[id] = model.Concept{ID: id, Name: doc.Title, Context: parseMetaID(meta, "context")}
	case model.KindCapability:
		m.Capabilities[id] = res.buildCapability(id, file, doc, meta)
	case model.KindInvariant:
		m.Invariants[id] = model.Invariant{ID: id, Title: doc.Title, Capabilities: res.parseReferences(file, doc, SectionCapabilities)}
	case model.KindSpecification:
		m.Specifications[id] = model.Specification{ID: id, Title: doc.Title, Specifies: res.parseReferences(file, doc, SectionSpecifies)}
	case model.KindADR:
		m.ADRs[id] = model.ADR{ID: id, Title: doc.Title}
	case model.KindScenario:
		m.Scenarios[id] = model.Scenario{ID: id, Name: doc.Title, Capabilities: res.parseReferences(file, doc, SectionCapabilities)}
	case model.KindVerification:
		m.Verification[id] = model.Verification{ID: id, Title: doc.Title, Paths: res.parsePaths(doc)}
	case model.KindTask:
		m.Tasks[id] = model.Task{ID: id, Title: doc.Title, Status: res.parseStatus(file, meta)}
	default:
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("unknown entity kind %q", kind)})
		return false
	}
	return true
}

func (res *LoadResult) buildCapability(id model.ID, file string, doc markdown.Document, meta map[string]string) model.Capability {
	return model.Capability{
		ID:             id,
		Name:           doc.Title,
		Context:        parseMetaID(meta, "context"),
		Status:         res.parseStatus(file, meta),
		Invariants:     res.parseInvariants(id, file, doc),
		Specifications: res.parseSpecifications(id, file, doc),
		ADRs:           res.parseReferences(file, doc, SectionADRs),
		Scenarios:      res.parseReferences(file, doc, SectionScenarios),
		Verification:   res.parseVerification(id, file, doc),
		Tasks:          res.parseReferences(file, doc, SectionTasks),
	}
}

// parseVerification reads a capability's Verification section. A bullet that
// references an identifier links to a file-backed verification entity; a bullet
// that is a path is inline verification, registered with a synthesised identifier
// and that path as its evidence.
func (res *LoadResult) parseVerification(owner model.ID, file string, doc markdown.Document) []model.ID {
	sec, ok := doc.FindSection(SectionVerification)
	if !ok {
		return nil
	}
	var ids []model.ID
	ordinal := 0
	for _, item := range sec.Items {
		if ref, ok := item.Reference(); ok {
			ids = append(ids, model.ID(ref).Canonical())
			continue
		}
		ordinal++
		inlineID := model.SynthesiseID(owner, "ver", ordinal)
		res.Model.Verification[inlineID] = model.Verification{ID: inlineID, Paths: []string{item.Text}, Owner: owner}
		ids = append(ids, inlineID)
	}
	return ids
}

// parseInvariants reads a capability's Invariants section, returning the identifiers
// to record on the capability. A bullet that references an identifier is a link to a
// file-backed invariant; a prose bullet is an inline invariant registered in the
// model with a synthesised, owner-scoped identifier.
func (res *LoadResult) parseInvariants(owner model.ID, file string, doc markdown.Document) []model.ID {
	sec, ok := doc.FindSection(SectionInvariants)
	if !ok {
		return nil
	}
	var ids []model.ID
	ordinal := 0
	for _, item := range sec.Items {
		if ref, ok := item.Reference(); ok {
			ids = append(ids, model.ID(ref).Canonical())
			continue
		}
		ordinal++
		inlineID := model.SynthesiseID(owner, "inv", ordinal)
		res.Model.Invariants[inlineID] = model.Invariant{ID: inlineID, Title: item.Text, Owner: owner}
		ids = append(ids, inlineID)
	}
	return ids
}

// parseSpecifications reads a capability's Specifications section. A bullet that
// references an identifier links to a file-backed specification; each deeper
// subsection is an inline specification whose heading is its title and whose bullet
// items are its detail, registered with a synthesised identifier.
func (res *LoadResult) parseSpecifications(owner model.ID, file string, doc markdown.Document) []model.ID {
	sec, ok := doc.FindSection(SectionSpecifications)
	if !ok {
		return nil
	}
	var ids []model.ID
	for _, item := range sec.Items {
		if ref, ok := item.Reference(); ok {
			ids = append(ids, model.ID(ref).Canonical())
			continue
		}
		res.add(Problem{File: file, Line: item.Line, Severity: SeverityWarning, Message: fmt.Sprintf("%s item %q is neither a reference nor an inline specification; write an inline specification as a subsection", SectionSpecifications, item.Text)})
	}
	ordinal := 0
	for _, sub := range doc.Subsections(SectionSpecifications) {
		ordinal++
		inlineID := model.SynthesiseID(owner, "spec", ordinal)
		res.Model.Specifications[inlineID] = model.Specification{ID: inlineID, Title: sub.Title, Specifies: []model.ID{owner}, Detail: getItemTexts(sub.Items), Owner: owner}
		ids = append(ids, inlineID)
	}
	return ids
}

// getItemTexts returns the text of each item in a section.
func getItemTexts(items []markdown.Item) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Text)
	}
	return out
}

// parseMetadata reads the Metadata section as a set of key: value pairs. Bullet items
// that are not key: value pairs are reported as warnings.
func (res *LoadResult) parseMetadata(file string, doc markdown.Document) (out map[string]string) {
	out = map[string]string{}
	sec, ok := doc.FindSection(SectionMetadata)
	if !ok {
		return
	}
	for _, item := range sec.Items {
		key, value, ok := item.KeyValue()
		if !ok {
			res.add(Problem{File: file, Line: item.Line, Severity: SeverityWarning, Message: fmt.Sprintf("metadata item %q is not a key: value pair", item.Text)})
			continue
		}
		out[key] = value
	}
	return out
}

// parseStatus reads the status metadata value, reporting an error for an unknown
// value. An absent status is left empty for the missing-status check in validate.
func (res *LoadResult) parseStatus(file string, meta map[string]string) model.Status {
	status := model.Status(meta["status"])
	if status != "" && !status.Valid() {
		res.add(Problem{File: file, Severity: SeverityError, Message: fmt.Sprintf("unknown status %q; valid statuses are %s", status, joinStatuses())})
	}
	return status
}

// joinStatuses renders the valid statuses as a quoted, comma-separated list for an
// error message, for example `"draft", "proposed", "in-progress", "done"`.
func joinStatuses() string {
	quoted := make([]string, len(model.Statuses))
	for i, s := range model.Statuses {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}

// parseMetaID reads a metadata value as an entity identifier in canonical form,
// returning the empty identifier when the key is absent or empty.
func parseMetaID(meta map[string]string, key string) model.ID {
	value := meta[key]
	if value == "" {
		return ""
	}
	return model.ID(value).Canonical()
}

// parseReferences reads a link section as a list of entity identifiers. Items that do
// not contain an identifier are reported as warnings.
func (res *LoadResult) parseReferences(file string, doc markdown.Document, section string) []model.ID {
	sec, ok := doc.FindSection(section)
	if !ok {
		return nil
	}
	var ids []model.ID
	for _, item := range sec.Items {
		ref, ok := item.Reference()
		if !ok {
			res.add(Problem{File: file, Line: item.Line, Severity: SeverityWarning, Message: fmt.Sprintf("%s item %q does not reference an identifier", section, item.Text)})
			continue
		}
		ids = append(ids, model.ID(ref).Canonical())
	}
	return ids
}

// parsePaths reads the Paths section as a list of file paths.
func (res *LoadResult) parsePaths(doc markdown.Document) []string {
	sec, ok := doc.FindSection(SectionPaths)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range sec.Items {
		out = append(out, item.Text)
	}
	return out
}
