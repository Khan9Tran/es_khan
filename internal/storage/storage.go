package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// HistoryItem represents an executed query or API request log entry.
type HistoryItem struct {
	ID           string    `json:"id"`
	Protocol     string    `json:"protocol,omitempty"` // "es", "http", "grpc"
	Timestamp    time.Time `json:"timestamp"`
	ConnectionID string    `json:"connection_id,omitempty"`
	Index        string    `json:"index,omitempty"`
	Method       string    `json:"method"` // GET, POST, or RPC method name
	Path         string    `json:"path"`   // ES path, HTTP URL, or gRPC service/method
	RawInput     string    `json:"raw_input,omitempty"`
	Status       int       `json:"status"` // HTTP status or gRPC code (0 = OK)
	StatusText   string    `json:"status_text,omitempty"`
	TookMs       int64     `json:"took_ms"`
	TotalHits    int64     `json:"total_hits,omitempty"`
}

// SnippetItem represents a saved, reusable query snippet or saved API request in Collections.
type SnippetItem struct {
	ID          string            `json:"id"`
	Protocol    string            `json:"protocol,omitempty"`   // "es", "http", "grpc"
	Collection  string            `json:"collection,omitempty"` // Group/collection name
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Method      string            `json:"method"`
	Path        string            `json:"path"` // ES path, HTTP URL, or gRPC target/service/method
	Content     string            `json:"content"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	BodyType    string            `json:"body_type,omitempty"`
	FormData    map[string]string `json:"form_data,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// StorageData holds the persisted state.
type StorageData struct {
	History  []HistoryItem `json:"history"`
	Snippets []SnippetItem `json:"snippets"`
}

// Store provides thread-safe access to persistent history and snippets.
type Store struct {
	mu         sync.RWMutex
	filePath   string
	maxHistory int
	history    []HistoryItem
	snippets   []SnippetItem
}

// NewStore initializes a history and snippet store.
func NewStore(customPath string, maxHistory int) (*Store, error) {
	if maxHistory <= 0 {
		maxHistory = 500
	}

	path := customPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			path = filepath.Join(".", ".eskhan", "data.json")
		} else {
			path = filepath.Join(home, ".eskhan", "data.json")
		}
	}

	s := &Store{
		filePath:   path,
		maxHistory: maxHistory,
		history:    []HistoryItem{},
		snippets:   DefaultSnippets(),
	}

	if err := s.load(); err != nil {
		if os.IsNotExist(err) {
			if saveErr := s.save(); saveErr != nil {
				return nil, fmt.Errorf("failed to save initial store: %w", saveErr)
			}
		} else {
			// If corrupted, backup damaged file and recover with defaults
			backupPath := fmt.Sprintf("%s.corrupted.%d", path, time.Now().Unix())
			_ = os.Rename(path, backupPath)
			log.Printf("⚠️ Corrupted store file %s backed up to %s. Initializing fresh store.", path, backupPath)
			if saveErr := s.save(); saveErr != nil {
				return nil, fmt.Errorf("failed to recover store: %w", saveErr)
			}
		}
	}

	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var d StorageData
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("failed to decode store json: %w", err)
	}

	s.history = d.History
	if len(d.Snippets) > 0 {
		existing := make(map[string]bool)
		for _, snip := range d.Snippets {
			existing[snip.ID] = true
		}
		merged := d.Snippets
		for _, def := range DefaultSnippets() {
			if !existing[def.ID] {
				merged = append(merged, def)
			}
		}
		s.snippets = merged
	}

	return nil
}

func (s *Store) save() error {
	s.mu.RLock()
	d := StorageData{
		History:  s.history,
		Snippets: s.snippets,
	}
	s.mu.RUnlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store: %w", err)
	}

	return atomicWriteFile(s.filePath, data, 0600)
}

func atomicWriteFile(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Chmod(perm); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpName, filePath); err != nil {
		return fmt.Errorf("failed to atomically replace %s: %w", filePath, err)
	}
	return nil
}

// AddHistory appends a new query execution record and trims old records.
func (s *Store) AddHistory(item HistoryItem) error {
	s.mu.Lock()
	if item.ID == "" {
		item.ID = fmt.Sprintf("hist-%d", time.Now().UnixNano())
	}
	if item.Timestamp.IsZero() {
		item.Timestamp = time.Now()
	}

	// Prepend to show latest first
	s.history = append([]HistoryItem{item}, s.history...)
	if len(s.history) > s.maxHistory {
		s.history = s.history[:s.maxHistory]
	}
	s.mu.Unlock()

	return s.save()
}

// GetHistory returns query history, latest first.
func (s *Store) GetHistory(limit int) []HistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.history) {
		limit = len(s.history)
	}

	result := make([]HistoryItem, limit)
	copy(result, s.history[:limit])
	return result
}

// GetHistoryByProtocol returns history filtered by protocol ("es", "http", "grpc").
func (s *Store) GetHistoryByProtocol(protocol string, limit int) []HistoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []HistoryItem
	for _, item := range s.history {
		p := item.Protocol
		if p == "" {
			p = "es" // legacy fallback
		}
		if strings.EqualFold(p, protocol) {
			filtered = append(filtered, item)
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}
	return filtered
}

// ClearHistory deletes all query history records.
func (s *Store) ClearHistory() error {
	s.mu.Lock()
	s.history = []HistoryItem{}
	s.mu.Unlock()
	return s.save()
}

// SaveSnippet adds or updates a saved query snippet.
func (s *Store) SaveSnippet(snippet SnippetItem) error {
	s.mu.Lock()
	now := time.Now()
	if snippet.ID == "" {
		snippet.ID = fmt.Sprintf("snip-%d", now.UnixNano())
		snippet.CreatedAt = now
	}
	snippet.UpdatedAt = now

	found := false
	for i, existing := range s.snippets {
		if existing.ID == snippet.ID {
			s.snippets[i] = snippet
			found = true
			break
		}
	}
	if !found {
		s.snippets = append(s.snippets, snippet)
	}
	s.mu.Unlock()

	return s.save()
}

// DeleteSnippet removes a snippet by ID.
func (s *Store) DeleteSnippet(id string) error {
	s.mu.Lock()
	var updated []SnippetItem
	for _, snip := range s.snippets {
		if snip.ID != id {
			updated = append(updated, snip)
		}
	}
	s.snippets = updated
	s.mu.Unlock()

	return s.save()
}

// GetSnippets returns all saved snippets.
func (s *Store) GetSnippets() []SnippetItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SnippetItem, len(s.snippets))
	copy(result, s.snippets)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})

	return result
}

// GetSnippetsByProtocol returns snippets filtered by protocol ("es", "http", "grpc").
func (s *Store) GetSnippetsByProtocol(protocol string) []SnippetItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []SnippetItem
	for _, snip := range s.snippets {
		p := snip.Protocol
		if p == "" {
			p = "es"
		}
		if strings.EqualFold(p, protocol) {
			filtered = append(filtered, snip)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Title < filtered[j].Title
	})

	return filtered
}

// DefaultSnippets returns useful out-of-the-box templates.
func DefaultSnippets() []SnippetItem {
	now := time.Now()
	return []SnippetItem{
		{
			ID:          "default-match-all",
			Title:       "Match All Documents",
			Description: "Simple query to retrieve first 20 documents",
			Tags:        []string{"basic", "search"},
			Method:      "POST",
			Path:        "/_search",
			Content:     "{\n  \"size\": 20,\n  \"query\": {\n    \"match_all\": {}\n  }\n}",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "default-bool-filter",
			Title:       "Bool Query with Filter",
			Description: "High-performance filtered search (exact term + text match)",
			Tags:        []string{"filter", "bool"},
			Method:      "POST",
			Path:        "/_search",
			Content:     "{\n  \"query\": {\n    \"bool\": {\n      \"must\": [\n        { \"match\": { \"name\": \"sample\" } }\n      ],\n      \"filter\": [\n        { \"term\": { \"status\": \"active\" } }\n      ]\n    }\n  }\n}",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "default-cat-indices",
			Title:       "List Indices & Health",
			Description: "Cat API to view indices, doc count, and store size",
			Tags:        []string{"cat", "admin"},
			Method:      "GET",
			Path:        "/_cat/indices?v&s=index",
			Content:     "",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "vietnamese-multi-match",
			Title:       "Tìm kiếm Tiếng Việt (Có dấu & Không dấu)",
			Description: "Multi-match tiếng Việt có trọng số trường và highlight từ khóa",
			Tags:        []string{"vietnamese", "search", "highlight"},
			Method:      "POST",
			Path:        "/_search",
			Content:     "{\n  \"query\": {\n    \"multi_match\": {\n      \"query\": \"điện thoại thông minh\",\n      \"fields\": [\"name^3\", \"title^2\", \"description\"],\n      \"type\": \"best_fields\",\n      \"operator\": \"and\"\n    }\n  },\n  \"highlight\": {\n    \"pre_tags\": [\"<mark>\"],\n    \"post_tags\": [\"</mark>\"],\n    \"fields\": {\n      \"name\": {},\n      \"description\": {}\n    }\n  }\n}",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "vietnamese-autocomplete",
			Title:       "Gợi ý Tiếng Việt (Phrase Prefix)",
			Description: "Autocomplete / tìm kiếm tiền tố từ ghép tiếng Việt tức thì",
			Tags:        []string{"vietnamese", "autocomplete", "prefix"},
			Method:      "POST",
			Path:        "/_search",
			Content:     "{\n  \"query\": {\n    \"match_phrase_prefix\": {\n      \"name\": {\n        \"query\": \"máy giặt\",\n        \"max_expansions\": 10\n      }\n    }\n  }\n}",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "vietnamese-bool-filter",
			Title:       "Tìm kiếm Tiếng Việt + Bộ lọc",
			Description: "Kết hợp tìm kiếm văn bản tiếng Việt và lọc điều kiện chính xác",
			Tags:        []string{"vietnamese", "filter", "bool"},
			Method:      "POST",
			Path:        "/_search",
			Content:     "{\n  \"query\": {\n    \"bool\": {\n      \"must\": [\n        {\n          \"multi_match\": {\n            \"query\": \"áo sơ mi\",\n            \"fields\": [\"title^2\", \"content\"],\n            \"operator\": \"and\"\n          }\n        }\n      ],\n      \"filter\": [\n        { \"term\": { \"status\": \"active\" } },\n        { \"range\": { \"price\": { \"gte\": 100000, \"lte\": 500000 } } }\n      ]\n    }\n  },\n  \"sort\": [\n    { \"_score\": { \"order\": \"desc\" } }\n  ]\n}",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}
