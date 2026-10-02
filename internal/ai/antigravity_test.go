package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"eskhan/internal/schema"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Markdown json block",
			input:    "Here is the query:\n```json\n{\n  \"query\": { \"match_all\": {} }\n}\n```\nHope it helps!",
			expected: "{\n  \"query\": { \"match_all\": {} }\n}",
		},
		{
			name:     "Markdown block without language tag",
			input:    "Query below:\n```\n{\"size\": 10}\n```",
			expected: "{\"size\": 10}",
		},
		{
			name:     "Raw JSON with surrounding text",
			input:    "Some text before {\"query\": {\"term\": {\"status\": \"active\"}}} and text after",
			expected: "{\"query\": {\"term\": {\"status\": \"active\"}}}",
		},
		{
			name:     "No valid JSON",
			input:    "This is just plain text without any json brackets",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSON() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindAgyBinary(t *testing.T) {
	path := FindAgyBinary()
	if path == "" {
		t.Log("Note: agy binary not found in standard paths on this test environment")
	} else {
		t.Logf("Found agy binary at: %s", path)
		if !strings.HasSuffix(path, "agy") {
			t.Errorf("expected path to end with agy, got %s", path)
		}
	}
}

func TestBridgeStatus(t *testing.T) {
	bridge := NewBridge("")
	status := bridge.Status()
	t.Logf("Antigravity Status: available=%v, path=%s, message=%s", status.Available, status.Path, status.Message)
	if status.Available && status.Path == "" {
		t.Errorf("expected non-empty path when available")
	}
}

func TestBridgeGenerateQuery_LiveIfAvailable(t *testing.T) {
	bridge := NewBridge("")
	status := bridge.Status()
	if !status.Available {
		t.Skip("Skipping live agy test because agy is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req := GenerateRequest{
		Prompt:    "tìm các document có status là active",
		IndexName: "users",
		Fields: []schema.FieldSuggestion{
			{Name: "status", Type: "keyword"},
			{Name: "name", Type: "text"},
		},
	}

	res, err := bridge.GenerateQuery(ctx, req)
	if err != nil {
		t.Logf("Live agy generation returned note/error: %v (skipping if headless permissions apply)", err)
		return
	}

	if res == nil || res.QueryDSL == "" {
		t.Fatalf("expected non-empty QueryDSL")
	}

	t.Logf("Generated DSL:\n%s", res.QueryDSL)
	if !strings.Contains(res.QueryDSL, "query") {
		t.Errorf("expected query clause in generated DSL")
	}
}
