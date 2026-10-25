package passage

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/acierto/nexora/internal/config"
)

func TestChunkHeadingsAndProvenance(t *testing.T) {
	doc := Document{ID: "payments/refunds.md", Location: "payments/refunds.md", Digest: "abc", Text: "---\nignored: front matter\n---\n# Payments\n\nIntro text.\n\n## Refunds\n\nRefund within 30 days.\n\n```go\n# not a heading\nfunc main() {}\n```\n\n### Partial\n\nNever exceed the capture.\n\n## Disputes\n\nOpen a case."}
	converted, err := Convert(doc.ID, doc.Location, "markdown", []byte(doc.Text))
	if err != nil {
		t.Fatal(err)
	}
	got := Chunk(converted, "handbook", "docs", []string{"commerce-api"}, Options{MaxChars: 1200, OverlapChars: 100})
	headings := []string{}
	for _, p := range got {
		headings = append(headings, p.Heading)
		if p.Title != "Payments" || p.SourceID != "handbook" || p.SourceType != "docs" || p.Location != "payments/refunds.md" || !reflect.DeepEqual(p.Projects, []string{"commerce-api"}) {
			t.Fatalf("provenance: %+v", p)
		}
		if strings.Contains(p.Text, "ignored: front matter") {
			t.Fatal("front matter leaked into passages")
		}
	}
	want := []string{"Payments", "Payments › Refunds", "Payments › Refunds › Partial", "Payments › Disputes"}
	if !reflect.DeepEqual(headings, want) {
		t.Fatalf("headings: %q", headings)
	}
	if !strings.Contains(got[1].Text, "# not a heading") {
		t.Fatalf("code fence lost: %q", got[1].Text)
	}
	if got[0].ID != "handbook:payments/refunds.md#1" || got[3].ID != "handbook:payments/refunds.md#4" {
		t.Fatalf("ids: %s %s", got[0].ID, got[3].ID)
	}
	again := Chunk(converted, "handbook", "docs", []string{"commerce-api"}, Options{MaxChars: 1200, OverlapChars: 100})
	if !reflect.DeepEqual(got, again) {
		t.Fatal("chunking is not deterministic")
	}
}

func TestChunkSizeAndOverlap(t *testing.T) {
	var b strings.Builder
	b.WriteString("# Long\n\n")
	for i := 0; i < 40; i++ {
		b.WriteString("Sentence number ")
		b.WriteString(strings.Repeat("x", i%7+1))
		b.WriteString(" explains one more detail about refunds. ")
		if i%5 == 4 {
			b.WriteString("\n\n")
		}
	}
	b.WriteString(strings.Repeat("y", 900))
	got := Chunk(Document{ID: "long.md", Text: b.String()}, "s", "docs", nil, Options{MaxChars: 300, OverlapChars: 60})
	if len(got) < 4 {
		t.Fatalf("expected several passages, got %d", len(got))
	}
	for _, p := range got {
		if n := utf8.RuneCountInString(p.Text); n > 300 {
			t.Fatalf("passage too long: %d", n)
		}
		if p.Projects == nil {
			t.Fatal("projects must be non-nil")
		}
	}
	if !strings.HasPrefix(got[1].Text, "…") {
		t.Fatalf("expected overlap marker: %q", got[1].Text[:20])
	}
}

func TestHTMLAndConfluenceStorage(t *testing.T) {
	html := `<html><head><title>Deploy guide</title><script>alert(1)</script></head><body><h1>Deploy</h1><p>Use the <b>pipeline</b>&nbsp;only.</p><ul><li>Check migrations</li><li>Announce</li></ul><table><tr><th>Env</th><th>Approval</th></tr><tr><td>prod</td><td>one</td></tr></table><pre>make deploy
make verify</pre></body></html>`
	text := HTMLToText(html)
	for _, want := range []string{"# Deploy", "Use the pipeline only.", "- Check migrations", "Env | Approval", "prod | one", "```\nmake deploy\nmake verify\n```"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "alert") || strings.Contains(text, "Deploy guide") {
		t.Fatalf("script or head leaked:\n%s", text)
	}
	if HTMLTitle(html) != "Deploy guide" {
		t.Fatalf("title: %q", HTMLTitle(html))
	}
	storage := `<h2>Rotate keys</h2><p>Run the job.</p><ac:structured-macro ac:name="code"><ac:parameter ac:name="language">bash</ac:parameter><ac:plain-text-body><![CDATA[./rotate --all]]></ac:plain-text-body></ac:structured-macro><ac:structured-macro ac:name="info"><ac:rich-text-body><p>Ask security first.</p></ac:rich-text-body></ac:structured-macro><p>See <ac:link><ri:page ri:content-title="Keys"/></ac:link> too.</p>`
	text = HTMLToText(storage)
	for _, want := range []string{"## Rotate keys", "```\n./rotate --all\n```", "Ask security first."} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "bash") {
		t.Fatalf("macro parameter leaked:\n%s", text)
	}
}

func TestLoadDocuments(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{"guide.md": "# Guide\n\nHello.", "nested/page.html": "<h1>Page</h1><p>Body</p>", "notes.txt": "plain notes", "image.png": "binary", "nested/deep/run-book_v2.markdown": "no heading here"}
	for name, content := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	docs, err := LoadDocuments(context.Background(), config.Source{ID: "h", Type: config.SourceDocs, Root: root})
	if err != nil {
		t.Fatal(err)
	}
	ids, titles := []string{}, []string{}
	for _, d := range docs {
		ids = append(ids, d.ID)
		titles = append(titles, d.Title)
	}
	if !reflect.DeepEqual(ids, []string{"guide.md", "nested/deep/run-book_v2.markdown", "nested/page.html", "notes.txt"}) {
		t.Fatalf("ids: %q", ids)
	}
	if !reflect.DeepEqual(titles, []string{"Guide", "Run book v2", "Page", "Notes"}) {
		t.Fatalf("titles: %q", titles)
	}
	if err := os.WriteFile(filepath.Join(root, "bad.md"), []byte{0xff, 0xfe}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDocuments(context.Background(), config.Source{ID: "h", Root: root}); err == nil || !strings.Contains(err.Error(), "bad.md") {
		t.Fatalf("invalid UTF-8 accepted: %v", err)
	}
	os.Remove(filepath.Join(root, "bad.md"))
	if err := os.Symlink(filepath.Join(root, "guide.md"), filepath.Join(root, "link.md")); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, err := LoadDocuments(context.Background(), config.Source{ID: "h", Root: root}); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink accepted: %v", err)
	}
}
