package passage

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

var (
	tagPattern = regexp.MustCompile(`(?s)<[^>]*>`)
	// Confluence prefixes become plain names so HTML auto-close rules (which
	// include "link") cannot misfire on <ac:link>.
	confluenceTag = regexp.MustCompile(`<(/?)(ac|ri):`)
	spaceRuns     = regexp.MustCompile(`[ \t\f\v]+`)
	trailingSpace = regexp.MustCompile(`(?m)[ \t]+$`)
)

// HTMLToText converts HTML or Confluence storage XHTML into Markdown-like text.
// It keeps headings, paragraphs, list items, table cells, and code bodies. Content
// is tokenized only; nothing is rendered, fetched, or executed.
func HTMLToText(input string) string {
	decoder := xml.NewDecoder(strings.NewReader("<nexora-root>" + confluenceTag.ReplaceAllString(input, "<$1$2-") + "</nexora-root>"))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity
	var out strings.Builder
	skip := 0       // inside script/style/head/macro parameters
	pre := 0        // inside <pre> or a code macro body
	inCode := false // inside an ac:structured-macro name="code"
	cell := false
	block := func() { out.WriteString("\n\n") }
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Malformed markup: fall back to stripping tags.
			return tidy(tagPattern.ReplaceAllString(input, " "))
		}
		switch t := token.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			space := strings.ToLower(t.Name.Space)
			if skip > 0 {
				if name == "script" || name == "style" || name == "ac-parameter" || name == "head" || name == "title" {
					skip++
				}
				continue
			}
			switch {
			case name == "script" || name == "style" || name == "head" || name == "title":
				skip++
			case name == "ac-parameter":
				skip++
			case name == "ac-structured-macro":
				for _, a := range t.Attr {
					if strings.ToLower(a.Name.Local) == "name" && (a.Value == "code" || a.Value == "noformat") {
						inCode = true
					}
				}
			case name == "ac-plain-text-body" && inCode:
				block()
				out.WriteString("```\n")
				pre++
			case len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6' && space == "":
				block()
				out.WriteString(strings.Repeat("#", int(name[1]-'0')) + " ")
			case name == "p" || name == "div" || name == "section" || name == "blockquote" || name == "table" || name == "ul" || name == "ol":
				block()
			case name == "br":
				out.WriteString("\n")
			case name == "li":
				out.WriteString("\n- ")
			case name == "tr":
				out.WriteString("\n")
				cell = false
			case name == "td" || name == "th":
				if cell {
					out.WriteString(" | ")
				}
				cell = true
			case name == "pre":
				block()
				out.WriteString("```\n")
				pre++
			}
		case xml.EndElement:
			name := strings.ToLower(t.Name.Local)
			space := strings.ToLower(t.Name.Space)
			if skip > 0 {
				if name == "script" || name == "style" || name == "ac-parameter" || name == "head" || name == "title" {
					skip--
				}
				continue
			}
			switch {
			case name == "ac-plain-text-body" && pre > 0:
				out.WriteString("\n```")
				block()
				pre--
			case name == "ac-structured-macro":
				inCode = false
			case name == "pre" && pre > 0:
				out.WriteString("\n```")
				block()
				pre--
			case len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6' && space == "",
				name == "p", name == "div", name == "section", name == "blockquote", name == "table", name == "ul", name == "ol":
				block()
			}
		case xml.CharData:
			if skip > 0 {
				continue
			}
			text := strings.ReplaceAll(string(t), "\u00a0", " ")
			if pre > 0 {
				out.WriteString(text)
				continue
			}
			text = spaceRuns.ReplaceAllString(strings.ReplaceAll(text, "\n", " "), " ")
			out.WriteString(text)
		}
	}
	return tidy(out.String())
}

// HTMLTitle returns the text of the first <title> or <h1>, if any.
func HTMLTitle(input string) string {
	for _, tag := range []string{"title", "h1"} {
		re := regexp.MustCompile(`(?is)<` + tag + `\b[^>]*>(.*?)</` + tag + `>`)
		if m := re.FindStringSubmatch(input); m != nil {
			if t := strings.TrimSpace(tidy(HTMLToText(m[1]))); t != "" {
				return t
			}
		}
	}
	return ""
}

func tidy(text string) string {
	lines := strings.Split(text, "\n")
	inFence := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			lines[i] = strings.TrimSpace(line)
			continue
		}
		if !inFence {
			lines[i] = strings.TrimSpace(spaceRuns.ReplaceAllString(line, " "))
		}
	}
	text = trailingSpace.ReplaceAllString(strings.Join(lines, "\n"), "")
	return strings.TrimSpace(blankRuns.ReplaceAllString(text, "\n\n"))
}
