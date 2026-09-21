package fields

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sooapps/sooauth/server/internal/store"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDynamicOptionsService_RootArray(t *testing.T) {
	mockData, _ := json.Marshal([]map[string]any{
		{"id": "iron", "name": "Iron IV"},
		{"id": "bronze", "name": "Bronze I"},
		{"id": "gold", "name": "Gold II"},
	})

	svc := NewDynamicOptionsService()
	svc.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(mockData)),
			Header:     make(http.Header),
		}, nil
	})

	opts, err := svc.FetchOptions(context.Background(), store.OptionsSource{
		Type:     "dynamic_api",
		URL:      "https://api.example.com/ranks",
		LabelKey: "name",
		ValueKey: "id",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts) != 3 {
		t.Fatalf("expected 3 options, got %d", len(opts))
	}
	if opts[0].Label != "Iron IV" || opts[0].Value != "iron" {
		t.Errorf("unexpected first option: %+v", opts[0])
	}
}

func TestDynamicOptionsService_NestedPathWithHeaders(t *testing.T) {
	mockData, _ := json.Marshal(map[string]any{
		"status": "ok",
		"data": map[string]any{
			"ranks": []map[string]any{
				{"tier": "PLATINUM", "division": "I", "formatted": "Platinum I"},
				{"tier": "DIAMOND", "division": "IV", "formatted": "Diamond IV"},
			},
		},
	})

	calls := 0
	svc := NewDynamicOptionsService()
	svc.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.Header.Get("X-Riot-Token") != "secret-token-123" {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte("unauthorized"))),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(mockData)),
			Header:     make(http.Header),
		}, nil
	})

	opts, err := svc.FetchOptions(context.Background(), store.OptionsSource{
		Type: "dynamic_api",
		URL:  "https://api.example.com/nested-ranks",
		Headers: map[string]string{
			"X-Riot-Token": "secret-token-123",
		},
		ItemsPath:       "data.ranks",
		LabelKey:        "formatted",
		ValueKey:        "tier",
		CacheTTLSeconds: 60,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
	if opts[0].Label != "Platinum I" || opts[0].Value != "PLATINUM" {
		t.Errorf("unexpected first option: %+v", opts[0])
	}
	if calls != 1 {
		t.Fatalf("expected 1 http call, got %d", calls)
	}

	// Verify caching works: second call should NOT trigger roundtripper
	cachedOpts, err := svc.FetchOptions(context.Background(), store.OptionsSource{
		Type: "dynamic_api",
		URL:  "https://api.example.com/nested-ranks",
		Headers: map[string]string{
			"X-Riot-Token": "secret-token-123",
		},
		ItemsPath:       "data.ranks",
		LabelKey:        "formatted",
		ValueKey:        "tier",
		CacheTTLSeconds: 60,
	})
	if err != nil {
		t.Fatalf("expected cache hit, got error: %v", err)
	}
	if len(cachedOpts) != 2 {
		t.Fatalf("expected 2 cached options, got %d", len(cachedOpts))
	}
	if calls != 1 {
		t.Fatalf("expected still 1 http call due to cache, got %d", calls)
	}
}
