package passage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/acierto/nexora/internal/config"
)

var (
	firstHeading = regexp.MustCompile(`(?m)^#[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	nameSplitter = regexp.MustCompile(`[-_]+`)
)

// Extensions lists documentation file types read from docs sources.
var Extensions = map[string]string{".md": "markdown", ".markdown": "markdown", ".txt": "text", ".html": "html", ".htm": "html"}

// LoadDocuments reads every supported file below a docs source root in stable
// order. Like knowledge loading it rejects symlink entries and requires UTF-8.
func LoadDocuments(ctx context.Context, source config.Source) ([]Document, error) {
	docs := []Document{}
	err := filepath.WalkDir(source.Root, func(file string, entry fs.DirEntry, walkErr error) error {
		rel, err := filepath.Rel(source.Root, file)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		wrap := func(err error) error { return fmt.Errorf("source %s, %s: %w", source.ID, rel, err) }
		if err = ctx.Err(); err != nil {
			return wrap(err)
		}
		if walkErr != nil {
			return wrap(walkErr)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return wrap(fmt.Errorf("symlink entries are not allowed"))
		}
		if entry.IsDir() {
			return nil
		}
		format, ok := Extensions[strings.ToLower(filepath.Ext(file))]
		if !entry.Type().IsRegular() || !ok {
			return nil
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return wrap(err)
		}
		doc, err := Convert(rel, rel, format, data)
		if err != nil {
			return wrap(err)
		}
		docs = append(docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// Convert turns raw file bytes into a Document. Leading YAML front matter in
// Markdown is skipped without being parsed.
func Convert(id, location, format string, data []byte) (Document, error) {
	if !utf8.Valid(data) {
		return Document{}, fmt.Errorf("invalid UTF-8")
	}
	doc := Document{ID: id, Location: location, Digest: fmt.Sprintf("%x", sha256.Sum256(data))}
	text := string(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")))
	switch format {
	case "markdown":
		text = stripFrontMatter(text)
		if m := firstHeading.FindStringSubmatch(text); m != nil {
			doc.Title = strings.TrimSpace(m[1])
		}
	case "html":
		doc.Title = HTMLTitle(text)
		text = HTMLToText(text)
	case "text":
	default:
		return Document{}, fmt.Errorf("unsupported format %q", format)
	}
	if doc.Title == "" {
		base := strings.TrimSuffix(filepath.Base(id), filepath.Ext(id))
		doc.Title = strings.TrimSpace(nameSplitter.ReplaceAllString(base, " "))
		if doc.Title != "" {
			doc.Title = strings.ToUpper(doc.Title[:1]) + doc.Title[1:]
		}
	}
	doc.Text = text
	return doc, nil
}

func stripFrontMatter(text string) string {
	if !strings.HasPrefix(text, "---\n") {
		return text
	}
	lines := strings.SplitAfter(text, "\n")
	for i := 1; i < len(lines) && i < 200; i++ {
		if strings.TrimRight(lines[i], "\r\n") == "---" {
			return strings.Join(lines[i+1:], "")
		}
	}
	return text
}
