package fields

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sooapps/sooauth/server/internal/store"
)

type cacheEntry struct {
	options   []store.SelectOption
	expiresAt time.Time
}

type DynamicOptionsService struct {
	client *http.Client
	mu     sync.RWMutex
	cache  map[string]cacheEntry
}

func NewDynamicOptionsService() *DynamicOptionsService {
	return &DynamicOptionsService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]cacheEntry),
	}
}

func (s *DynamicOptionsService) FetchOptions(ctx context.Context, source store.OptionsSource) ([]store.SelectOption, error) {
	if source.Type == "static" {
		return source.Options, nil
	}

	url := strings.TrimSpace(source.URL)
	if url == "" {
		return nil, errors.New("url is required for dynamic options")
	}

	ttl := source.CacheTTLSeconds
	if ttl <= 0 {
		ttl = 300 // default 5 minutes
	}

	cacheKey := s.buildCacheKey(source)

	// Check cache
	s.mu.RLock()
	entry, ok := s.cache[cacheKey]
	s.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.options, nil
	}

	method := strings.ToUpper(strings.TrimSpace(source.Method))
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	for k, v := range source.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("external api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // limit to 5MB
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	options, err := parseOptionsFromJSON(bodyBytes, source.ItemsPath, source.LabelKey, source.ValueKey)
	if err != nil {
		return nil, err
	}

	// Cache options
	s.mu.Lock()
	s.cache[cacheKey] = cacheEntry{
		options:   options,
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	s.mu.Unlock()

	return options, nil
}

func (s *DynamicOptionsService) buildCacheKey(source store.OptionsSource) string {
	var headersStr strings.Builder
	for k, v := range source.Headers {
		headersStr.WriteString(k)
		headersStr.WriteString(":")
		headersStr.WriteString(v)
		headersStr.WriteString(";")
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", source.Method, source.URL, headersStr.String(), source.ItemsPath, source.LabelKey, source.ValueKey)
}

func parseOptionsFromJSON(data []byte, itemsPath, labelKey, valueKey string) ([]store.SelectOption, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	items, err := extractItems(raw, itemsPath)
	if err != nil {
		return nil, err
	}

	out := make([]store.SelectOption, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case map[string]any:
			lbl := extractString(v, labelKey, []string{"name", "label", "title", "text"})
			val := extractString(v, valueKey, []string{"id", "value", "key", "slug", "code"})
			if lbl == "" {
				lbl = val
			}
			if val == "" {
				val = lbl
			}
			out = append(out, store.SelectOption{
				Label: lbl,
				Value: val,
			})
		case string:
			out = append(out, store.SelectOption{
				Label: v,
				Value: v,
			})
		case float64:
			str := fmt.Sprintf("%v", v)
			out = append(out, store.SelectOption{
				Label: str,
				Value: str,
			})
		default:
			str := fmt.Sprintf("%v", v)
			out = append(out, store.SelectOption{
				Label: str,
				Value: str,
			})
		}
	}

	return out, nil
}

func extractItems(raw any, itemsPath string) ([]any, error) {
	itemsPath = strings.TrimSpace(itemsPath)
	if itemsPath == "" {
		if arr, ok := raw.([]any); ok {
			return arr, nil
		}
		return nil, errors.New("expected json array at root or specify items_path")
	}

	parts := strings.Split(itemsPath, ".")
	curr := raw
	for _, part := range parts {
		m, ok := curr.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("path segment '%s' is not an object", part)
		}
		val, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("path segment '%s' not found in json", part)
		}
		curr = val
	}

	arr, ok := curr.([]any)
	if !ok {
		return nil, fmt.Errorf("value at '%s' is not an array", itemsPath)
	}
	return arr, nil
}

func extractString(m map[string]any, preferredKey string, fallbackKeys []string) string {
	if preferredKey != "" {
		if val, ok := m[preferredKey]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
	}
	for _, k := range fallbackKeys {
		if val, ok := m[k]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}
