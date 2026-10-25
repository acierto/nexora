// Package confluence reads pages from Confluence spaces through the REST API.
// It is read-only and used only by explicit indexing.
package confluence

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/acierto/nexora/internal/passage"
)

// Client reads pages with basic auth (User set, e.g. Cloud email + API token)
// or a bearer token (Data Center personal access token).
type Client struct {
	BaseURL string
	User    string
	Token   string
	HTTP    *http.Client
	// PageLimit is the page size requested per call (default 50).
	PageLimit int
}

// Page is one current Confluence page in storage format.
type Page struct {
	ID      string
	Title   string
	Space   string
	URL     string
	Version int
	Storage string
}

type listResponse struct {
	Results []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
		Body struct {
			Storage struct {
				Value string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
		Links struct {
			WebUI string `json:"webui"`
		} `json:"_links"`
	} `json:"results"`
	Links struct {
		Base string `json:"base"`
		Next string `json:"next"`
	} `json:"_links"`
}

const maxPagesPerSpace = 20000

// Pages lists every current page of a space, following pagination links.
func (c Client) Pages(ctx context.Context, space string) ([]Page, error) {
	limit := c.PageLimit
	if limit <= 0 {
		limit = 50
	}
	base := strings.TrimRight(c.BaseURL, "/")
	query := url.Values{"spaceKey": {space}, "type": {"page"}, "status": {"current"}, "expand": {"body.storage,version"}, "limit": {strconv.Itoa(limit)}}
	next := base + "/rest/api/content?" + query.Encode()
	pages := []Page{}
	seen := map[string]bool{}
	for next != "" {
		if seen[next] {
			return nil, fmt.Errorf("space %s: pagination loop at %s", space, next)
		}
		seen[next] = true
		var resp listResponse
		if err := c.get(ctx, next, &resp); err != nil {
			return nil, fmt.Errorf("space %s: %w", space, err)
		}
		linkBase := strings.TrimRight(resp.Links.Base, "/")
		if linkBase == "" {
			linkBase = base
		}
		for _, r := range resp.Results {
			page := Page{ID: r.ID, Title: r.Title, Space: space, Version: r.Version.Number, Storage: r.Body.Storage.Value}
			if r.Links.WebUI != "" {
				page.URL = join(linkBase, r.Links.WebUI)
			} else {
				page.URL = base + "/pages/viewpage.action?pageId=" + url.QueryEscape(r.ID)
			}
			pages = append(pages, page)
			if len(pages) > maxPagesPerSpace {
				return nil, fmt.Errorf("space %s: more than %d pages; narrow the configured spaces", space, maxPagesPerSpace)
			}
		}
		next = ""
		if resp.Links.Next != "" {
			next = join(linkBase, resp.Links.Next)
		}
	}
	return pages, nil
}

func join(base, link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	if !strings.HasPrefix(link, "/") {
		link = "/" + link
	}
	return base + link
}

func (c Client) get(ctx context.Context, target string, value any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if c.User != "" {
		req.SetBasicAuth(c.User, c.Token)
	} else if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	client := c.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		hint := ""
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			hint = " (check user_env/token_env credentials and space permissions)"
		}
		return fmt.Errorf("HTTP %d%s: %s", resp.StatusCode, hint, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(value); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// Document converts a page into a passage document. The digest covers the
// storage body, so unchanged pages keep identical passages.
func (p Page) Document() passage.Document {
	text := passage.HTMLToText(p.Storage)
	return passage.Document{
		ID:       "page-" + p.ID,
		Title:    p.Title,
		Location: p.URL,
		Text:     text,
		Digest:   fmt.Sprintf("%x", sha256.Sum256([]byte(p.Storage))),
	}
}
