package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExecute_GET_WithQueryParamsAndHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("sort") != "desc" {
			t.Errorf("unexpected query params: %s", r.URL.RawQuery)
		}
		if r.Header.Get("X-Custom-Header") != "eskhan-test" {
			t.Errorf("missing or incorrect X-Custom-Header: %s", r.Header.Get("X-Custom-Header"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Server-Time", "2026-10-02")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","items":[1,2,3]}`))
	}))
	defer ts.Close()

	req := Request{
		Method: "GET",
		URL:    ts.URL,
		QueryParams: map[string]string{
			"page": "2",
			"sort": "desc",
		},
		Headers: map[string]string{
			"X-Custom-Header": "eskhan-test",
		},
		TimeoutMs: 5000,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Size == 0 {
		t.Errorf("expected non-zero response size")
	}
	if resp.Headers["X-Server-Time"][0] != "2026-10-02" {
		t.Errorf("expected header X-Server-Time to be 2026-10-02")
	}
	if resp.Timings.TotalMs < 0 {
		t.Errorf("expected total ms >= 0")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &parsed); err != nil {
		t.Fatalf("failed to parse JSON response body: %v", err)
	}
	if parsed["status"] != "success" {
		t.Errorf("expected status success, got %v", parsed["status"])
	}
}

func TestExecute_POST_JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(bodyBytes, &payload)
		if payload["name"] != "eskhan" {
			t.Errorf("expected name=eskhan, got %v", payload["name"])
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created":true,"id":"eskhan-123"}`))
	}))
	defer ts.Close()

	req := Request{
		Method:   "POST",
		URL:      ts.URL,
		BodyType: "json",
		Body:     `{"name":"eskhan","version":2}`,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestExecute_POST_FormUrlEncoded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("username") != "alice" || r.FormValue("role") != "admin" {
			t.Errorf("unexpected form values: %v", r.Form)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer ts.Close()

	req := Request{
		Method:   "POST",
		URL:      ts.URL,
		BodyType: "x_www_form_urlencoded",
		FormData: map[string]string{
			"username": "alice",
			"role":     "admin",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestExecute_POST_MultipartFormData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(10 << 20)
		if r.FormValue("file_id") != "abc-999" {
			t.Errorf("unexpected multipart form value: %s", r.FormValue("file_id"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Uploaded"))
	}))
	defer ts.Close()

	req := Request{
		Method:   "POST",
		URL:      ts.URL,
		BodyType: "form_data",
		FormData: map[string]string{
			"file_id": "abc-999",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestExecute_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req := Request{
		Method:    "GET",
		URL:       ts.URL,
		TimeoutMs: 30, // 30ms timeout will expire while server sleeps 200ms
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute should return error struct, not err: %v", err)
	}

	if resp.StatusCode != 0 || resp.Error == "" {
		t.Errorf("expected status 0 and error message on timeout, got code=%d error=%s", resp.StatusCode, resp.Error)
	}
}

func TestExecute_Validation(t *testing.T) {
	_, err := Execute(context.Background(), Request{URL: ""})
	if err == nil {
		t.Errorf("expected error for empty URL, got nil")
	}
}

func TestExecute_VariablesInterpolation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/42" {
			t.Errorf("expected path /v1/users/42, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("expected Authorization header Bearer secret-token, got %s", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"action":"created","user_id":"42"}` {
			t.Errorf("unexpected body: %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	req := Request{
		Method: "POST",
		URL:    "{{base_url}}/v1/users/{{user_id}}",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		BodyType: "json",
		Body:     `{"action":"created","user_id":"{{user_id}}"}`,
		Variables: map[string]string{
			"base_url": ts.URL,
			"user_id":  "42",
			"token":    "secret-token",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
