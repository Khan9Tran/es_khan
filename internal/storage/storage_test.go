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

func TestMultiProtocolStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eskhan-store-proto-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dataPath := filepath.Join(tempDir, "data.json")
	store, err := NewStore(dataPath, 20)
	if err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}

	// Add HTTP snippet
	httpSnippet := SnippetItem{
		Protocol:   "http",
		Collection: "Auth Service",
		Title:      "Login API",
		Method:     "POST",
		Path:       "https://api.example.com/v1/auth/login",
		Content:    `{"username":"admin"}`,
	}
	if err := store.SaveSnippet(httpSnippet); err != nil {
		t.Fatalf("failed to save http snippet: %v", err)
	}

	// Add gRPC snippet
	grpcSnippet := SnippetItem{
		Protocol:   "grpc",
		Collection: "User Service",
		Title:      "GetUser RPC",
		Method:     "GetUser",
		Path:       "localhost:50051/user.v1.UserService/GetUser",
		Content:    `{"id":123}`,
	}
	if err := store.SaveSnippet(grpcSnippet); err != nil {
		t.Fatalf("failed to save grpc snippet: %v", err)
	}

	httpSnips := store.GetSnippetsByProtocol("http")
	if len(httpSnips) != 1 || httpSnips[0].Title != "Login API" {
		t.Errorf("expected 1 http snippet, got %v", httpSnips)
	}

	grpcSnips := store.GetSnippetsByProtocol("grpc")
	if len(grpcSnips) != 1 || grpcSnips[0].Title != "GetUser RPC" {
		t.Errorf("expected 1 grpc snippet, got %v", grpcSnips)
	}

	// Add multi-protocol history
	_ = store.AddHistory(HistoryItem{
		Protocol: "http",
		Method:   "GET",
		Path:     "https://api.example.com/users",
		Status:   200,
		TookMs:   45,
	})
	_ = store.AddHistory(HistoryItem{
		Protocol: "grpc",
		Method:   "Check",
		Path:     "localhost:50051/grpc.health.v1.Health/Check",
		Status:   0,
		TookMs:   12,
	})

	httpHistory := store.GetHistoryByProtocol("http", 10)
	if len(httpHistory) != 1 || httpHistory[0].Path != "https://api.example.com/users" {
		t.Errorf("expected 1 http history, got %v", httpHistory)
	}

	grpcHistory := store.GetHistoryByProtocol("grpc", 10)
	if len(grpcHistory) != 1 || grpcHistory[0].Path != "localhost:50051/grpc.health.v1.Health/Check" {
		t.Errorf("expected 1 grpc history, got %v", grpcHistory)
	}
}
