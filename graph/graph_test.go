package graph

import (
	"testing"

	"github.com/a-h/cap/model"
)

func TestDataGroups(t *testing.T) {
	m := &model.Model{
		Requirements: map[model.ID]model.Requirement{
			"req-0001": {ID: "req-0001", Title: "Auth requirement", Capabilities: []model.ID{"cap-0001", "cap-0002"}},
			"req-0002": {ID: "req-0002", Title: "Payment requirement", Capabilities: []model.ID{"cap-0002"}},
		},
		Capabilities: map[model.ID]model.Capability{
			"cap-0001": {ID: "cap-0001", Name: "Login"},
			"cap-0002": {ID: "cap-0002", Name: "Pay"},
		},
	}
	data := Build(m)

	tests := []struct {
		name     string
		fromKind string
		toKind   string
		expected []Group
	}{
		{
			name:     "requirement to capability groups are sorted with to-rows sorted within each group",
			fromKind: "requirement",
			toKind:   "capability",
			expected: []Group{
				{
					FromID: "req-0001", FromTitle: "Auth requirement",
					ToRows: []ToRow{{ToID: "cap-0001", ToTitle: "Login"}, {ToID: "cap-0002", ToTitle: "Pay"}},
				},
				{
					FromID: "req-0002", FromTitle: "Payment requirement",
					ToRows: []ToRow{{ToID: "cap-0002", ToTitle: "Pay"}},
				},
			},
		},
		{
			name:     "reversed kind pair returns no groups when no such edges exist",
			fromKind: "capability",
			toKind:   "requirement",
			expected: nil,
		},
		{
			name:     "an unknown kind pair returns no groups",
			fromKind: "scenario",
			toKind:   "adr",
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.Groups(tt.fromKind, tt.toKind)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d groups, expected %d: %v", len(got), len(tt.expected), got)
			}
			for i, g := range got {
				e := tt.expected[i]
				if g.FromID != e.FromID || g.FromTitle != e.FromTitle {
					t.Errorf("group %d from: got {%s %q}, expected {%s %q}", i, g.FromID, g.FromTitle, e.FromID, e.FromTitle)
				}
				if len(g.ToRows) != len(e.ToRows) {
					t.Fatalf("group %d: got %d to-rows, expected %d", i, len(g.ToRows), len(e.ToRows))
				}
				for j, tr := range g.ToRows {
					et := e.ToRows[j]
					if tr.ToID != et.ToID || tr.ToTitle != et.ToTitle {
						t.Errorf("group %d to-row %d: got {%s %q}, expected {%s %q}", i, j, tr.ToID, tr.ToTitle, et.ToID, et.ToTitle)
					}
				}
			}
		})
	}
}
