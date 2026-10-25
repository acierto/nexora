package embed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/config"
)

func TestProviders(t *testing.T) {
	var gotAuth, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		var body struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path == "/api/embed" {
			out := [][]float32{}
			for range body.Input {
				out = append(out, []float32{1, 2, 3})
			}
			json.NewEncoder(w).Encode(map[string]any{"embeddings": out})
			return
		}
		data := []map[string]any{}
		for i := len(body.Input) - 1; i >= 0; i-- { // out of order on purpose
			data = append(data, map[string]any{"index": i, "embedding": []float32{float32(i), 0}})
		}
		json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	defer server.Close()

	if e, err := New(config.Embeddings{Provider: config.ProviderNone}, nil, nil); err != nil || e != nil {
		t.Fatalf("none provider: %v %v", e, err)
	}
	ollama, err := New(config.Embeddings{Provider: config.ProviderOllama, Model: "nomic", URL: server.URL}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	texts := make([]string, 70) // more than one batch
	vectors, err := ollama.Embed(context.Background(), texts)
	if err != nil || len(vectors) != 70 || gotPath != "/api/embed" || ollama.Model() != "ollama:nomic" {
		t.Fatalf("ollama: %v %d %s", err, len(vectors), gotPath)
	}
	env := func(name string) string {
		if name == "KEY" {
			return "sk-test"
		}
		return ""
	}
	openai, err := New(config.Embeddings{Provider: config.ProviderOpenAI, Model: "small", URL: server.URL, APIKeyEnv: "KEY"}, env, nil)
	if err != nil {
		t.Fatal(err)
	}
	vectors, err = openai.Embed(context.Background(), []string{"a", "b", "c"})
	if err != nil || gotAuth != "Bearer sk-test" || gotPath != "/embeddings" || vectors[2][0] != 2 {
		t.Fatalf("openai: %v %q %s %v", err, gotAuth, gotPath, vectors)
	}
	if _, err := New(config.Embeddings{Provider: config.ProviderOpenAI, Model: "small", URL: server.URL, APIKeyEnv: "MISSING"}, env, nil); err == nil {
		t.Fatal("missing key accepted")
	}
	wrong, _ := New(config.Embeddings{Provider: config.ProviderOllama, Model: "nomic", URL: server.URL, Dimensions: 768}, nil, nil)
	if _, err := wrong.Embed(context.Background(), []string{"a"}); err == nil || !strings.Contains(err.Error(), "768") {
		t.Fatalf("dimension mismatch not reported: %v", err)
	}
}

func TestProviderHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer server.Close()
	e, _ := New(config.Embeddings{Provider: config.ProviderOllama, Model: "missing", URL: server.URL}, nil, nil)
	if _, err := e.Embed(context.Background(), []string{"x"}); err == nil || !strings.Contains(err.Error(), "404") || !strings.Contains(err.Error(), "model not found") {
		t.Fatalf("error: %v", err)
	}
}
