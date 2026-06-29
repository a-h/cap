// Package model defines the entities of the cap system model and the
// relationships between them
//
// Capabilities are the stable hub of the model. Every other traceable entity
// links to one or more capabilities. Bounded contexts group capabilities for
// navigation and are not part of the traceability graph.
package model

import (
	"fmt"
	"strconv"
	"strings"
)

// Kind identifies the type of an entity.
type Kind string

const (
	KindContext       Kind = "context"
	KindCapability    Kind = "capability"
	KindInvariant     Kind = "invariant"
	KindSpecification Kind = "specification"
	KindADR           Kind = "adr"
	KindScenario      Kind = "scenario"
	KindVerification  Kind = "verification"
	KindTask          Kind = "task"
)

// numberWidth is the zero-padded width of the numeric part of a canonical
// identifier, for example the four digits in "cap-0003".
const numberWidth = 4

// ID is the identifier of an entity. The canonical form is a lowercase prefix, a
// hyphen, and a zero-padded number, for example "cap-0003". The prefix indicates
// the kind of entity. Identifiers are compared and stored in canonical form, so
// "CAP-3", "cap-003", and "cap-0003" all refer to the same entity.
type ID string

// Prefix returns the lowercase identifier prefix, for example "cap" for
// "CAP-003". It reports ok=false when the identifier has no prefix.
func (id ID) Prefix() (prefix string, ok bool) {
	before, _, found := strings.Cut(string(id), "-")
	if !found || before == "" {
		return "", false
	}
	return strings.ToLower(before), true
}

// Canonical returns the identifier in canonical form: a lowercase prefix and a
// zero-padded number, for example "cap-0003". When the identifier does not have a
// numeric part it is returned with only its prefix lowercased, so non-numeric
// identifiers remain usable.
func (id ID) Canonical() ID {
	if strings.Contains(string(id), "/") {
		return id
	}
	before, rest, found := strings.Cut(string(id), "-")
	if !found || before == "" {
		return ID(strings.ToLower(string(id)))
	}
	prefix := strings.ToLower(before)
	n, err := strconv.Atoi(rest)
	if err != nil {
		return ID(prefix + "-" + strings.ToLower(rest))
	}
	return ID(fmt.Sprintf("%s-%0*d", prefix, numberWidth, n))
}

// SynthesiseID builds the identifier of an inline entity from its owning
// capability, the kind's prefix, and a one-based ordinal, for example
// "cap-0003/inv-1". Inline identifiers are local to their owner and cannot be
// referenced from other files.
func SynthesiseID(owner ID, prefix string, ordinal int) ID {
	return ID(fmt.Sprintf("%s/%s-%d", owner, prefix, ordinal))
}

// Capability is the primary, long-lived entity. It describes a single thing the
// system can do, independent of UI and implementation.
type Capability struct {
	ID             ID     `json:"id"`
	Name           string `json:"name"`
	Context        ID     `json:"context,omitempty"`
	Status         Status `json:"status,omitempty"`
	Invariants     []ID   `json:"invariants,omitempty"`
	Specifications []ID   `json:"specifications,omitempty"`
	ADRs           []ID   `json:"adrs,omitempty"`
	Scenarios      []ID   `json:"scenarios,omitempty"`
	Verification   []ID   `json:"verification,omitempty"`
	Tasks          []ID   `json:"tasks,omitempty"`
}

// Status records the lifecycle state of an entity that represents work in progress,
// such as a capability or a task.
type Status string

const (
	StatusDraft      Status = "draft"
	StatusProposed   Status = "proposed"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

// Statuses lists every valid status in lifecycle order. It is the single source of
// truth for what Valid accepts and for the values reported in messages.
var Statuses = []Status{StatusDraft, StatusProposed, StatusInProgress, StatusDone}

// Valid reports whether the status is one of the known values.
func (s Status) Valid() bool {
	for _, known := range Statuses {
		if s == known {
			return true
		}
	}
	return false
}

// Context is a bounded context: a boundary within which one model and language
// apply. It groups capabilities for organisation and is not itself a node in the
// traceability graph.
type Context struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

// Invariant is a domain invariant: a rule that must always hold. It states a
// required property as a positive assertion of behaviour and may constrain many
// capabilities, so the capability-to-invariant relationship is many-to-many.
//
// An invariant is either file-backed (its own file, shareable across capabilities)
// or inline (written within a single capability, with Owner set to that capability
// and a synthesised identifier).
type Invariant struct {
	ID           ID     `json:"id"`
	Title        string `json:"title"`
	Capabilities []ID   `json:"capabilities,omitempty"`
	Owner        ID     `json:"owner,omitempty"`
}

// Inline reports whether the invariant is written within a capability rather than
// in its own file.
func (i Invariant) Inline() bool { return i.Owner != "" }

// Specification is the overall design that realises a set of invariants: how the
// pieces fit together, including implementation detail. It does not restate
// behaviour, which the invariants already capture. A specification is of a
// capability or of a bounded context (a design spanning the context's
// capabilities), named by Of.
//
// Like an invariant, a specification is either file-backed or inline within a
// single capability.
type Specification struct {
	ID     ID       `json:"id"`
	Title  string   `json:"title"`
	Of     ID       `json:"of,omitempty"`
	Detail []string `json:"detail,omitempty"`
	Owner  ID       `json:"owner,omitempty"`
}

// Inline reports whether the specification is written within a capability rather
// than in its own file.
func (s Specification) Inline() bool { return s.Owner != "" }

// ADR is an architectural decision record that constrains how capabilities are
// implemented.
type ADR struct {
	ID    ID     `json:"id"`
	Title string `json:"title"`
}

// Scenario is an end-to-end path through the system enabled by one or more
// capabilities. Scenarios are consumers of capabilities.
type Scenario struct {
	ID           ID     `json:"id"`
	Name         string `json:"name"`
	Capabilities []ID   `json:"capabilities,omitempty"`
}

// Verification links the tests that cover a capability to that capability. Each
// path names evidence (a test file, or a procedure). Like an invariant or
// specification, it is either file-backed or inline within a single capability.
type Verification struct {
	ID    ID       `json:"id"`
	Title string   `json:"title,omitempty"`
	Paths []string `json:"paths,omitempty"`
	Owner ID       `json:"owner,omitempty"`
}

// Inline reports whether the verification is written within a capability rather
// than in its own file.
func (v Verification) Inline() bool { return v.Owner != "" }

// Task is transient implementation work that advances one or more capabilities.
type Task struct {
	ID     ID     `json:"id"`
	Title  string `json:"title"`
	Status Status `json:"status,omitempty"`
}

// MapToEntityKind derives the entity kind from the identifier's prefix, matched
// case-insensitively. It reports ok=false when the identifier has no prefix or the
// prefix is not a known kind.
func (id ID) MapToEntityKind() (kind Kind, ok bool) {
	prefix, ok := id.Prefix()
	if !ok {
		return "", false
	}
	switch prefix {
	case "ctx":
		return KindContext, true
	case "cap":
		return KindCapability, true
	case "inv":
		return KindInvariant, true
	case "spec":
		return KindSpecification, true
	case "adr":
		return KindADR, true
	case "scn":
		return KindScenario, true
	case "ver":
		return KindVerification, true
	case "task":
		return KindTask, true
	}
	return "", false
}

// Model is the complete, loaded system model, indexed by identifier.
type Model struct {
	Contexts       map[ID]Context
	Capabilities   map[ID]Capability
	Invariants     map[ID]Invariant
	Specifications map[ID]Specification
	ADRs           map[ID]ADR
	Scenarios      map[ID]Scenario
	Verification   map[ID]Verification
	Tasks          map[ID]Task
}

// NewModel returns a Model with all indexes initialised.
func NewModel() *Model {
	return &Model{
		Contexts:       map[ID]Context{},
		Capabilities:   map[ID]Capability{},
		Invariants:     map[ID]Invariant{},
		Specifications: map[ID]Specification{},
		ADRs:           map[ID]ADR{},
		Scenarios:      map[ID]Scenario{},
		Verification:   map[ID]Verification{},
		Tasks:          map[ID]Task{},
	}
}

// Lookup returns the kind of the entity with the given identifier, reporting
// ok=false when no entity with that identifier is loaded.
func (m *Model) Lookup(id ID) (kind Kind, ok bool) {
	if _, ok := m.Contexts[id]; ok {
		return KindContext, true
	}
	if _, ok := m.Capabilities[id]; ok {
		return KindCapability, true
	}
	if _, ok := m.Invariants[id]; ok {
		return KindInvariant, true
	}
	if _, ok := m.Specifications[id]; ok {
		return KindSpecification, true
	}
	if _, ok := m.ADRs[id]; ok {
		return KindADR, true
	}
	if _, ok := m.Scenarios[id]; ok {
		return KindScenario, true
	}
	if _, ok := m.Verification[id]; ok {
		return KindVerification, true
	}
	if _, ok := m.Tasks[id]; ok {
		return KindTask, true
	}
	return "", false
}
