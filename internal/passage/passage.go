// Package passage turns documentation into retrievable passages with provenance.
// Passages are context, never knowledge items: they carry no applicability
// metadata and can never become mandatory.
package passage

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Document is one source document converted to Markdown-like text.
type Document struct {
	// ID is unique within its source: a relative path or "page-<id>".
	ID       string
	Title    string
	Location string
	Text     string
	Digest   string
}

// Passage is one retrievable chunk of a document.
type Passage struct {
	ID         string   `json:"id"`
	SourceID   string   `json:"source_id"`
	SourceType string   `json:"source_type"`
	DocumentID string   `json:"document_id"`
	Title      string   `json:"title"`
	Heading    string   `json:"heading"`
	Text       string   `json:"text"`
	Location   string   `json:"location"`
	Digest     string   `json:"digest"`
	Projects   []string `json:"projects"`
	// Embedding is stored by the index, never printed.
	Embedding []float32 `json:"-"`
}

// Clone returns a copy with independent slices and non-nil projects.
func (p Passage) Clone() Passage {
	p.Projects = append([]string{}, p.Projects...)
	if p.Embedding != nil {
		p.Embedding = append([]float32{}, p.Embedding...)
	}
	return p
}

// AllowedFor reports whether a passage may be returned for a project.
// This is an applicability filter, not authorization.
func (p Passage) AllowedFor(project string) bool {
	return len(p.Projects) == 0 || slices.Contains(p.Projects, project)
}

// SearchText is the text indexed for keyword search.
func (p Passage) SearchText() string { return p.Title + "\n" + p.Heading + "\n" + p.Text }

// Options controls chunk sizes in characters (runes).
type Options struct {
	MaxChars     int
	OverlapChars int
}

var (
	headingLine = regexp.MustCompile(`^(#{1,6})[ \t]+(.*?)[ \t]*#*[ \t]*$`)
	fenceLine   = regexp.MustCompile("^[ \t]*(```|~~~)")
	blankRuns   = regexp.MustCompile(`\n{3,}`)
)

type section struct {
	trail      []string
	paragraphs []string
}

// Chunk splits a document at headings and paragraphs into passages of at most
// MaxChars runes, repeating up to OverlapChars runes of the previous passage
// within the same section.
func Chunk(doc Document, sourceID, sourceType string, projects []string, opt Options) []Passage {
	if opt.MaxChars <= 0 {
		opt.MaxChars = 1200
	}
	if opt.OverlapChars < 0 || opt.OverlapChars*2 > opt.MaxChars {
		opt.OverlapChars = 0
	}
	title := strings.TrimSpace(doc.Title)
	sections := splitSections(doc.Text)
	if title == "" {
		for _, s := range sections {
			if len(s.trail) > 0 {
				title = s.trail[0]
				break
			}
		}
	}
	if title == "" {
		title = doc.ID
	}
	out := []Passage{}
	for _, s := range sections {
		heading := strings.Join(s.trail, " › ")
		for _, text := range pack(s.paragraphs, opt) {
			out = append(out, Passage{
				ID:         fmt.Sprintf("%s:%s#%d", sourceID, doc.ID, len(out)+1),
				SourceID:   sourceID,
				SourceType: sourceType,
				DocumentID: doc.ID,
				Title:      title,
				Heading:    heading,
				Text:       text,
				Location:   doc.Location,
				Digest:     doc.Digest,
				Projects:   append([]string{}, projects...),
			})
		}
	}
	return out
}

func splitSections(text string) []section {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	sections := []section{{}}
	current := func() *section { return &sections[len(sections)-1] }
	var paragraph []string
	flush := func() {
		joined := strings.TrimSpace(strings.Join(paragraph, "\n"))
		if joined != "" {
			current().paragraphs = append(current().paragraphs, joined)
		}
		paragraph = nil
	}
	fence := ""
	for _, line := range strings.Split(text, "\n") {
		if fence != "" {
			paragraph = append(paragraph, line)
			if m := fenceLine.FindStringSubmatch(line); m != nil && m[1] == fence {
				fence = ""
				flush()
			}
			continue
		}
		if m := fenceLine.FindStringSubmatch(line); m != nil {
			flush()
			fence = m[1]
			paragraph = append(paragraph, line)
			continue
		}
		if m := headingLine.FindStringSubmatch(line); m != nil {
			flush()
			level := len(m[1])
			trail := append([]string{}, current().trail...)
			if len(trail) >= level {
				trail = trail[:level-1]
			}
			for len(trail) < level-1 {
				trail = append(trail, "")
			}
			trail = append(trail, strings.TrimSpace(m[2]))
			// Trail keeps level positions (empty for skipped levels) so a deeper
			// heading replaces the right entries; empties are dropped on output.
			sections = append(sections, section{trail: trail})
			continue
		}
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		paragraph = append(paragraph, line)
	}
	flush()
	out := []section{}
	for _, s := range sections {
		if len(s.paragraphs) == 0 {
			continue
		}
		trail := []string{}
		for _, t := range s.trail {
			if t != "" {
				trail = append(trail, t)
			}
		}
		out = append(out, section{trail: trail, paragraphs: s.paragraphs})
	}
	return out
}

// pack groups paragraphs into chunks, splitting oversized paragraphs.
func pack(paragraphs []string, opt Options) []string {
	pieces := []string{}
	for _, p := range paragraphs {
		pieces = append(pieces, splitLong(p, opt.MaxChars)...)
	}
	chunks := []string{}
	var current strings.Builder
	runes := 0
	emit := func() {
		if runes == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(current.String()))
		current.Reset()
		runes = 0
	}
	for _, piece := range pieces {
		n := utf8.RuneCountInString(piece)
		if runes > 0 && runes+2+n > opt.MaxChars {
			previous := current.String()
			emit()
			if tail := overlap(previous, opt.OverlapChars); tail != "" && utf8.RuneCountInString(tail)+3+n <= opt.MaxChars {
				current.WriteString("…" + tail)
				runes = utf8.RuneCountInString(tail) + 1
			}
		}
		if runes > 0 {
			current.WriteString("\n\n")
			runes += 2
		}
		current.WriteString(piece)
		runes += n
	}
	emit()
	for i, c := range chunks {
		chunks[i] = blankRuns.ReplaceAllString(c, "\n\n")
	}
	return chunks
}

// splitLong breaks a paragraph longer than max at sentence, then word, then
// rune boundaries. Code fences are split on line boundaries.
func splitLong(p string, max int) []string {
	if utf8.RuneCountInString(p) <= max {
		return []string{p}
	}
	separators := []string{"\n", ". ", "; ", ", ", " "}
	for _, sep := range separators {
		parts := strings.SplitAfter(p, sep)
		if len(parts) < 2 {
			continue
		}
		out := []string{}
		var b strings.Builder
		for _, part := range parts {
			if b.Len() > 0 && utf8.RuneCountInString(b.String())+utf8.RuneCountInString(part) > max {
				out = append(out, strings.TrimSpace(b.String()))
				b.Reset()
			}
			b.WriteString(part)
		}
		if strings.TrimSpace(b.String()) != "" {
			out = append(out, strings.TrimSpace(b.String()))
		}
		fine := true
		for _, o := range out {
			if utf8.RuneCountInString(o) > max {
				fine = false
				break
			}
		}
		if fine {
			return out
		}
	}
	r := []rune(p)
	out := []string{}
	for len(r) > 0 {
		n := min(max, len(r))
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}

// overlap returns up to n runes from the end of text, starting at a word.
func overlap(text string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(strings.TrimSpace(text))
	if len(r) <= n {
		return ""
	}
	tail := string(r[len(r)-n:])
	if idx := strings.IndexAny(tail, " \n"); idx >= 0 && idx < len(tail)-1 {
		tail = tail[idx+1:]
	}
	return strings.TrimSpace(tail)
}
