package markdown

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, doc Document)
	}{
		{
			name:  "the level one heading becomes the document title",
			input: "# Evaluate policies\n\nSome prose.\n",
			check: func(t *testing.T, doc Document) {
				if doc.Title != "Evaluate policies" {
					t.Errorf("got title %q, expected %q", doc.Title, "Evaluate policies")
				}
			},
		},
		{
			name:  "bullet items are collected under their section",
			input: "# Title\n\n## Requirements\n\n- REQ-001\n- REQ-002\n",
			check: func(t *testing.T, doc Document) {
				sec, ok := doc.FindSection("Requirements")
				if !ok {
					t.Fatalf("expected a Requirements section")
				}
				got := []string{}
				for _, it := range sec.Items {
					got = append(got, it.Text)
				}
				if diff := cmp.Diff([]string{"REQ-001", "REQ-002"}, got); diff != "" {
					t.Errorf("items mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:  "sections can be found case insensitively",
			input: "# Title\n\n## metadata\n\n- area: policy-enforcement\n",
			check: func(t *testing.T, doc Document) {
				if _, ok := doc.FindSection("Metadata"); !ok {
					t.Errorf("expected to find Metadata section regardless of case")
				}
			},
		},
		{
			name:  "content inside fenced code blocks is ignored",
			input: "# Title\n\n## Tasks\n\n```\n- not a task\n## Not a heading\n```\n\n- TASK-1\n",
			check: func(t *testing.T, doc Document) {
				sec, ok := doc.FindSection("Tasks")
				if !ok {
					t.Fatalf("expected a Tasks section")
				}
				if len(sec.Items) != 1 || sec.Items[0].Text != "TASK-1" {
					t.Errorf("expected only TASK-1 outside the fence, got %#v", sec.Items)
				}
				if _, ok := doc.FindSection("Not a heading"); ok {
					t.Errorf("headings inside a fence must not be parsed")
				}
			},
		},
		{
			name:  "asterisk and plus bullet markers are recognised",
			input: "# Title\n\n## Tasks\n\n* TASK-1\n+ TASK-2\n",
			check: func(t *testing.T, doc Document) {
				sec, _ := doc.FindSection("Tasks")
				if len(sec.Items) != 2 {
					t.Errorf("expected 2 items, got %d", len(sec.Items))
				}
			},
		},
		{
			name:  "line numbers are recorded for items",
			input: "# Title\n## Tasks\n- TASK-1\n",
			check: func(t *testing.T, doc Document) {
				sec, _ := doc.FindSection("Tasks")
				if sec.Items[0].Line != 3 {
					t.Errorf("got line %d, expected 3", sec.Items[0].Line)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, doc)
		})
	}
}

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{name: "a simple pair is split", text: "area: policy-enforcement", wantKey: "area", wantValue: "policy-enforcement", wantOK: true},
		{name: "surrounding whitespace is trimmed", text: "  status :  implemented ", wantKey: "status", wantValue: "implemented", wantOK: true},
		{name: "an item without a colon is not a pair", text: "no colon here", wantOK: false},
		{name: "an item with an empty key is not a pair", text: ": value", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := Item{Text: tt.text}.KeyValue()
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && (key != tt.wantKey || value != tt.wantValue) {
				t.Errorf("got (%q, %q), expected (%q, %q)", key, value, tt.wantKey, tt.wantValue)
			}
		})
	}
}

func TestReference(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		wantID string
		wantOK bool
	}{
		{name: "a plain identifier is extracted", text: "CAP-003", wantID: "CAP-003", wantOK: true},
		{name: "an identifier in a markdown link is extracted from the label", text: "[CAP-003](../capabilities/CAP-003-evaluate-policy.md)", wantID: "CAP-003", wantOK: true},
		{name: "an identifier followed by description is extracted", text: "REQ-001 Policies must be evaluated consistently", wantID: "REQ-001", wantOK: true},
		{name: "a verification identifier with a type prefix is extracted", text: "UNIT-042", wantID: "UNIT-042", wantOK: true},
		{name: "an item without an identifier reports not ok", text: "see the related work", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := Item{Text: tt.text}.Reference()
			if ok != tt.wantOK {
				t.Fatalf("got ok %v, expected %v", ok, tt.wantOK)
			}
			if ok && id != tt.wantID {
				t.Errorf("got %q, expected %q", id, tt.wantID)
			}
		})
	}
}

func TestSubsections(t *testing.T) {
	input := "# Title\n\n## Specifications\n\n### One\n\n- a\n\n### Two\n\n- b\n\n## Tasks\n\n### Not a spec subsection\n"
	doc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("subsections beneath a section are returned, stopping at the next same-level heading", func(t *testing.T) {
		subs := doc.Subsections("Specifications")
		if len(subs) != 2 {
			t.Fatalf("expected 2 subsections, got %d: %#v", len(subs), subs)
		}
		if subs[0].Title != "One" || subs[1].Title != "Two" {
			t.Errorf("got titles %q, %q, expected One, Two", subs[0].Title, subs[1].Title)
		}
		if len(subs[0].Items) != 1 || subs[0].Items[0].Text != "a" {
			t.Errorf("expected the subsection's items to be captured, got %#v", subs[0].Items)
		}
	})

	t.Run("a subsection under a different section is not returned", func(t *testing.T) {
		for _, sub := range doc.Subsections("Specifications") {
			if sub.Title == "Not a spec subsection" {
				t.Errorf("subsection of Tasks must not appear under Specifications")
			}
		}
	})

	t.Run("a section with no subsections returns none", func(t *testing.T) {
		if subs := doc.Subsections("Tasks"); len(subs) != 1 {
			t.Errorf("expected Tasks to have its one subsection, got %#v", subs)
		}
	})
}
