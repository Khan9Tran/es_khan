package api

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func startAPITestGRPCServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthSrv)
	reflection.Register(s)

	go func() {
		_ = s.Serve(lis)
	}()

	cleanup := func() {
		s.Stop()
		_ = lis.Close()
	}

	return lis.Addr().String(), cleanup
}

func TestAPIGRPCReflect(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	grpcAddr, cleanupGRPC := startAPITestGRPCServer(t)
	defer cleanupGRPC()

	reqBody := map[string]any{
		"target":    grpcAddr,
		"plaintext": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/grpc/reflect", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Services []struct {
			Name    string `json:"name"`
			Methods []any  `json:"methods"`
		} `json:"services"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Services) == 0 {
		t.Errorf("expected at least 1 service, got 0")
	}
}

func TestAPIGRPCInvoke(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	grpcAddr, cleanupGRPC := startAPITestGRPCServer(t)
	defer cleanupGRPC()

	reqBody := map[string]any{
		"target":     grpcAddr,
		"plaintext":  true,
		"service":    "grpc.health.v1.Health",
		"method":     "Check",
		"body":       `{"service": ""}`,
		"timeout_ms": 3000,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/grpc/invoke", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		StatusCode string `json:"status_code"`
		Code       uint32 `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.StatusCode != "OK" {
		t.Errorf("expected status OK, got %s", resp.StatusCode)
	}
}

func TestAPIGRPCParseProto(t *testing.T) {
	srv, cleanupAPI := setupTestServer(t)
	defer cleanupAPI()

	protoContent := `
syntax = "proto3";
package order.v1;

message OrderRequest {
  string order_id = 1;
}

message OrderResponse {
  string status = 1;
}

service OrderService {
  rpc GetOrder(OrderRequest) returns (OrderResponse);
}
`
	reqBody := map[string]any{
		"content": protoContent,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/grpc/proto/parse", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Services []struct {
			Name string `json:"name"`
		} `json:"services"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(resp.Services))
	}
	if resp.Services[0].Name != "order.v1.OrderService" {
		t.Errorf("expected order.v1.OrderService, got %s", resp.Services[0].Name)
	}
}
