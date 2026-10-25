package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPagesPaginationAndAuth(t *testing.T) {
	var server *httptest.Server
	auth := []string{}
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = append(auth, r.Header.Get("Authorization"))
		if r.URL.Path != "/wiki/rest/api/content" || r.URL.Query().Get("spaceKey") != "PAY" || r.URL.Query().Get("expand") != "body.storage,version" {
			http.Error(w, "bad request "+r.URL.String(), http.StatusBadRequest)
			return
		}
		page := func(id, title string) map[string]any {
			return map[string]any{"id": id, "title": title, "version": map[string]any{"number": 3},
				"body":   map[string]any{"storage": map[string]any{"value": "<h1>" + title + "</h1><p>Body " + id + "</p>"}},
				"_links": map[string]any{"webui": "/spaces/PAY/pages/" + id}}
		}
		resp := map[string]any{"_links": map[string]any{"base": server.URL + "/wiki"}}
		if r.URL.Query().Get("start") == "" {
			resp["results"] = []any{page("1", "Refunds")}
			resp["_links"].(map[string]any)["next"] = "/rest/api/content?spaceKey=PAY&type=page&status=current&expand=body.storage,version&limit=1&start=1"
		} else {
			resp["results"] = []any{page("2", "Chargebacks")}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	pages, err := Client{BaseURL: server.URL + "/wiki", User: "me@example.com", Token: "secret", PageLimit: 1}.Pages(context.Background(), "PAY")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0].Title != "Refunds" || pages[1].URL != server.URL+"/wiki/spaces/PAY/pages/2" || pages[1].Version != 3 {
		t.Fatalf("pages: %+v", pages)
	}
	if !strings.HasPrefix(auth[0], "Basic ") {
		t.Fatalf("expected basic auth, got %q", auth[0])
	}
	doc := pages[0].Document()
	if doc.ID != "page-1" || doc.Location != pages[0].URL || !strings.Contains(doc.Text, "# Refunds") || len(doc.Digest) != 64 {
		t.Fatalf("document: %+v", doc)
	}

	auth = nil
	if _, err := (Client{BaseURL: server.URL + "/wiki", Token: "pat"}).Pages(context.Background(), "PAY"); err != nil {
		t.Fatal(err)
	}
	if auth[0] != "Bearer pat" {
		t.Fatalf("expected bearer auth, got %q", auth[0])
	}
}

func TestPagesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer server.Close()
	_, err := Client{BaseURL: server.URL, Token: "x"}.Pages(context.Background(), "ARCH")
	if err == nil || !strings.Contains(err.Error(), "space ARCH") || !strings.Contains(err.Error(), "401") {
		t.Fatalf("error: %v", err)
	}

	loop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"results": []any{}, "_links": map[string]any{"next": "/rest/api/content?same=1"}})
	}))
	defer loop.Close()
	if _, err := (Client{BaseURL: loop.URL, Token: "x"}).Pages(context.Background(), "A"); err == nil || !strings.Contains(err.Error(), "pagination loop") {
		t.Fatalf("loop not detected: %v", err)
	}
}
