// Package review assembles the review packet for an entity.
//
// A review packet is the material an external assessor, typically an LLM agent,
// needs to judge an entity's quality: the entity's content, the context it relates
// to, and a checklist of judgement questions appropriate to its kind. cap makes no
// judgement and reaches no network; it only assembles the packet.
package review

import (
	"fmt"
	"os"
	"sort"
	"strings"

	capcontext "github.com/a-h/cap/context"
	"github.com/a-h/cap/markdown"
	"github.com/a-h/cap/model"
	"github.com/a-h/cap/store"
)

// Packet is the review material for a single entity.
type Packet struct {
	ID        model.ID           `json:"id"`
	Kind      model.Kind         `json:"kind"`
	Title     string             `json:"title"`
	File      string             `json:"file"`
	Content   string             `json:"content"`
	Context   *capcontext.Bundle `json:"context,omitempty"`
	Checklist []string           `json:"checklist"`
}

// Assemble builds the review packet for the entity with the given identifier. It
// reports ok=false when no entity with that identifier is loaded.
func Assemble(res store.LoadResult, id model.ID) (p Packet, ok bool) {
	kind, ok := res.Model.Lookup(id)
	if !ok {
		return Packet{}, false
	}
	p = Packet{ID: id, Kind: kind, File: res.Files[id], Checklist: buildChecklist(kind)}
	if raw, err := os.ReadFile(p.File); err == nil {
		p.Content = strings.TrimRight(string(raw), "\n")
		if doc, err := markdown.Parse(strings.NewReader(string(raw))); err == nil {
			p.Title = doc.Title
		}
	}
	if kind == model.KindCapability {
		if bundle, found := capcontext.For(res.Model, id); found {
			p.Context = &bundle
		}
	}
	return p, true
}

// AssembleAll builds a review packet for every entity in the model, ordered by
// identifier for deterministic output.
func AssembleAll(res store.LoadResult) []Packet {
	var ids []model.ID
	for id := range res.Files {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	packets := make([]Packet, 0, len(ids))
	for _, id := range ids {
		if p, ok := Assemble(res, id); ok {
			packets = append(packets, p)
		}
	}
	return packets
}

// Render produces the Markdown form of a packet, suitable for piping to an
// assessor.
func (p Packet) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Review: %s (%s)\n\n", p.ID, p.Kind)
	if p.File != "" {
		fmt.Fprintf(&b, "Source: %s\n\n", p.File)
	}
	b.WriteString("## Content under review\n\n")
	if p.Content != "" {
		b.WriteString(p.Content)
		b.WriteString("\n\n")
	}
	if p.Context != nil {
		b.WriteString("## Linked context\n\n")
		b.WriteString(p.Context.String())
		b.WriteString("\n")
	}
	b.WriteString("## Checklist\n\n")
	for _, q := range p.Checklist {
		fmt.Fprintf(&b, "- %s\n", q)
	}
	return b.String()
}
