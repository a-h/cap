package markdown

import (
	"regexp"
	"strings"
)

// idPattern matches an entity identifier such as cap-0003 or unit-0042. The
// prefix is one or more letters of any case, followed by a hyphen and a number.
// The identifier is canonicalised by the caller.
var idPattern = regexp.MustCompile(`\b([A-Za-z][A-Za-z0-9]*-[0-9]+)\b`)

// KeyValue splits a metadata item of the form "key: value" into its parts. It
// reports ok=false when the item does not contain a colon-separated key.
func (it Item) KeyValue() (key, value string, ok bool) {
	i := strings.Index(it.Text, ":")
	if i <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(it.Text[:i])
	value = strings.TrimSpace(it.Text[i+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

// Reference extracts an entity identifier from a link item. It first looks at the
// text of a Markdown link ("[CAP-003](...)") and otherwise scans the item text.
// It reports ok=false when no identifier is present.
func (it Item) Reference() (id string, ok bool) {
	if label, found := parseLinkLabel(it.Text); found {
		if m := idPattern.FindString(label); m != "" {
			return m, true
		}
	}
	if m := idPattern.FindString(it.Text); m != "" {
		return m, true
	}
	return "", false
}

// parseLinkLabel returns the label text of the first Markdown link in s.
func parseLinkLabel(s string) (label string, ok bool) {
	open := strings.Index(s, "[")
	if open < 0 {
		return "", false
	}
	close := strings.Index(s[open:], "]")
	if close < 0 {
		return "", false
	}
	return s[open+1 : open+close], true
}
