package markdown

import (
	"os"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/knowledge"
)

func TestAuthoringGuideSamples(t *testing.T) {
	data, e := os.ReadFile("../../docs/knowledge-format.md")
	if e != nil {
		t.Fatal(e)
	}
	parts := strings.Split(string(data), "```markdown\n")
	if len(parts) != 7 {
		t.Fatalf("expected five valid and one invalid sample; got %d", len(parts)-1)
	}
	for idx, part := range parts[1:] {
		document := strings.SplitN(part, "```", 2)[0]
		_, err := Parse([]byte(document), knowledge.Source{ID: "docs", Path: "sample.md"})
		if idx < 5 && err != nil {
			t.Fatalf("sample %d: %v", idx, err)
		}
		if idx == 5 && err == nil {
			t.Fatal("invalid sample accepted")
		}
	}
}
