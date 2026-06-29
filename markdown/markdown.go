// Package markdown parses the structured Markdown documents that store cap
// entities.
//
// A document carries structure through well-known section headings. The parser
// extracts those sections and tolerates everything else: free prose anywhere is
// ignored, and malformed sections are reported as problems rather than causing
// the parse to fail.
package markdown

import (
	"bufio"
	"io"
	"strings"
)

// initialLineBufferBytes is the starting capacity of the scanner's line buffer. It
// is grown as needed up to maxLineBytes.
const initialLineBufferBytes = 64 * 1024

// maxLineBytes is the largest single line the parser accepts. It is raised above
// bufio.Scanner's 64 KiB default so a long line, such as a wide table row, does not
// cause Scan to fail with bufio.ErrTooLong.
const maxLineBytes = 1024 * 1024

// Document is the parsed structure of a single entity Markdown file. Free prose
// outside the recognised sections is not retained.
type Document struct {
	Title    string
	Sections []Section
}

// Section is a heading and the structured content found beneath it, up to the
// next heading of the same or higher level.
type Section struct {
	Level int
	Title string
	Line  int
	Items []Item
	// HasContent reports whether any non-blank line, prose or bullet, appears
	// beneath the heading. It distinguishes an empty section from one with content
	// even when that content is prose rather than parsed items.
	HasContent bool
}

// Item is a single bullet-list entry within a section, with its source line for
// reporting.
type Item struct {
	Text string
	Line int
}

// FindSection returns the first section whose title matches name,
// case-insensitively, reporting ok=false when no such section exists.
func (d Document) FindSection(name string) (s Section, ok bool) {
	for _, sec := range d.Sections {
		if strings.EqualFold(sec.Title, name) {
			return sec, true
		}
	}
	return Section{}, false
}

// Subsections returns the sections nested directly beneath the named section: those
// of a deeper level that follow it in document order, up to the next section of the
// same or a shallower level. It is used to read inline definitions written as deeper
// headings under a link section.
func (d Document) Subsections(name string) []Section {
	var subs []Section
	var parentLevel int
	collecting := false
	for _, sec := range d.Sections {
		if collecting {
			if sec.Level <= parentLevel {
				break
			}
			subs = append(subs, sec)
			continue
		}
		if strings.EqualFold(sec.Title, name) {
			collecting = true
			parentLevel = sec.Level
		}
	}
	return subs
}

// Parse reads a structured Markdown document. It never returns an error for
// malformed content; structural problems are detected by callers inspecting the
// returned Document.
func Parse(r io.Reader) (doc Document, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, initialLineBufferBytes), maxLineBytes)

	var current *Section
	var inFence bool
	var fence string
	line := 0
	for scanner.Scan() {
		line++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		if f, ok := parseFenceMarker(trimmed); ok {
			if current != nil {
				current.HasContent = true
			}
			if !inFence {
				inFence, fence = true, f
				continue
			}
			if f == fence {
				inFence, fence = false, ""
			}
			continue
		}
		if inFence {
			continue
		}

		if level, title, ok := parseHeading(raw); ok {
			if level == 1 && doc.Title == "" {
				doc.Title = title
			}
			doc.Sections = append(doc.Sections, Section{Level: level, Title: title, Line: line})
			current = &doc.Sections[len(doc.Sections)-1]
			continue
		}

		if current == nil {
			continue
		}
		if trimmed != "" {
			current.HasContent = true
		}
		if text, ok := parseBullet(trimmed); ok {
			current.Items = append(current.Items, Item{Text: text, Line: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return doc, err
	}
	return doc, nil
}

// parseFenceMarker reports whether a trimmed line opens or closes a fenced code block
// and returns the fence token (``` or ~~~) used.
func parseFenceMarker(trimmed string) (fence string, ok bool) {
	switch {
	case strings.HasPrefix(trimmed, "```"):
		return "```", true
	case strings.HasPrefix(trimmed, "~~~"):
		return "~~~", true
	}
	return "", false
}

// parseHeading parses an ATX heading line, returning its level and trimmed title.
func parseHeading(raw string) (level int, title string, ok bool) {
	i := 0
	for i < len(raw) && raw[i] == '#' {
		i++
	}
	if i == 0 || i > 6 {
		return 0, "", false
	}
	if i < len(raw) && raw[i] != ' ' {
		return 0, "", false
	}
	return i, strings.TrimSpace(raw[i:]), true
}

// parseBullet parses an unordered list item, returning its trimmed text.
func parseBullet(trimmed string) (text string, ok bool) {
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(trimmed, marker) {
			return strings.TrimSpace(trimmed[len(marker):]), true
		}
	}
	return "", false
}
