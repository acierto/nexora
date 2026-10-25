// Package embed calls optional embedding providers over HTTP. Nexora works
// without embeddings; these are used only when configured.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/acierto/nexora/internal/config"
)

// Embedder turns texts into vectors with one fixed model.
type Embedder interface {
	Model() string
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

const batchSize = 64

// New returns nil when the provider is none.
func New(cfg config.Embeddings, getenv func(string) string, client *http.Client) (Embedder, error) {
	if client == nil {
		client = http.DefaultClient
	}
	switch cfg.Provider {
	case config.ProviderNone, "":
		return nil, nil
	case config.ProviderOllama:
		return &provider{cfg: cfg, client: client, kind: config.ProviderOllama}, nil
	case config.ProviderOpenAI:
		key := getenv(cfg.APIKeyEnv)
		if strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("embeddings: environment variable %s is not set", cfg.APIKeyEnv)
		}
		return &provider{cfg: cfg, client: client, kind: config.ProviderOpenAI, key: key}, nil
	}
	return nil, fmt.Errorf("embeddings: unknown provider %q", cfg.Provider)
}

type provider struct {
	cfg    config.Embeddings
	client *http.Client
	kind   string
	key    string
}

func (p *provider) Model() string { return p.kind + ":" + p.cfg.Model }

func (p *provider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += batchSize {
		end := min(start+batchSize, len(texts))
		vectors, err := p.batch(ctx, texts[start:end])
		if err != nil {
			return nil, err
		}
		out = append(out, vectors...)
	}
	dims := 0
	for i, v := range out {
		if len(v) == 0 {
			return nil, fmt.Errorf("embeddings: empty vector for input %d", i)
		}
		if dims == 0 {
			dims = len(v)
		}
		if len(v) != dims {
			return nil, fmt.Errorf("embeddings: inconsistent vector sizes %d and %d", dims, len(v))
		}
	}
	if p.cfg.Dimensions > 0 && dims != 0 && dims != p.cfg.Dimensions {
		return nil, fmt.Errorf("embeddings: model returned %d dimensions but index.embeddings.dimensions is %d", dims, p.cfg.Dimensions)
	}
	return out, nil
}

func (p *provider) batch(ctx context.Context, texts []string) ([][]float32, error) {
	var endpoint string
	body := map[string]any{"model": p.cfg.Model, "input": texts}
	if p.kind == config.ProviderOllama {
		endpoint = p.cfg.URL + "/api/embed"
	} else {
		endpoint = p.cfg.URL + "/embeddings"
		if p.cfg.Dimensions > 0 {
			body["dimensions"] = p.cfg.Dimensions
		}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.key != "" {
		req.Header.Set("Authorization", "Bearer "+p.key)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embeddings (%s): %w", p.Model(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return nil, fmt.Errorf("embeddings (%s): HTTP %d: %s", p.Model(), resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	if p.kind == config.ProviderOllama {
		var r struct {
			Embeddings [][]float32 `json:"embeddings"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, fmt.Errorf("embeddings: decode response: %w", err)
		}
		if len(r.Embeddings) != len(texts) {
			return nil, fmt.Errorf("embeddings: expected %d vectors, got %d", len(texts), len(r.Embeddings))
		}
		return r.Embeddings, nil
	}
	var r struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("embeddings: decode response: %w", err)
	}
	if len(r.Data) != len(texts) {
		return nil, fmt.Errorf("embeddings: expected %d vectors, got %d", len(texts), len(r.Data))
	}
	out := make([][]float32, len(texts))
	for _, d := range r.Data {
		if d.Index < 0 || d.Index >= len(texts) || out[d.Index] != nil {
			return nil, fmt.Errorf("embeddings: invalid response index %d", d.Index)
		}
		out[d.Index] = d.Embedding
	}
	return out, nil
}
