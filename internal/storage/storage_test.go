package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageHistoryAndSnippets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eskhan-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dataPath := filepath.Join(tempDir, "data.json")
	store, err := NewStore(dataPath, 5) // max 5 items
	if err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	// Verify default snippets
	snippets := store.GetSnippets()
	if len(snippets) == 0 {
		t.Errorf("expected default snippets")
	}

	// Add new custom snippet
	customSnippet := SnippetItem{
		Title:   "Custom Aggregation",
		Method:  "POST",
		Path:    "/orders/_search",
		Content: "{\"size\": 0}",
	}
	if err := store.SaveSnippet(customSnippet); err != nil {
		t.Fatalf("failed to save snippet: %v", err)
	}

	allSnippets := store.GetSnippets()
	if len(allSnippets) != len(snippets)+1 {
		t.Errorf("expected %d snippets, got %d", len(snippets)+1, len(allSnippets))
	}

	// Test history ring buffer trimming
	for i := 1; i <= 8; i++ {
		err := store.AddHistory(HistoryItem{
			Index:     "logs",
			Method:    "GET",
			Path:      "/logs/_search",
			RawInput:  "{}",
			Status:    200,
			TookMs:    int64(i),
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Fatalf("failed to add history: %v", err)
		}
	}

	history := store.GetHistory(100)
	if len(history) != 5 {
		t.Errorf("expected history capped at maxHistory 5, got %d", len(history))
	}
	// Latest item should be took = 8
	if history[0].TookMs != 8 {
		t.Errorf("expected latest history item took=8, got %d", history[0].TookMs)
	}

	// Clear history
	if err := store.ClearHistory(); err != nil {
		t.Fatalf("failed to clear history: %v", err)
	}
	if len(store.GetHistory(10)) != 0 {
		t.Errorf("expected 0 history items after clear")
	}
}
