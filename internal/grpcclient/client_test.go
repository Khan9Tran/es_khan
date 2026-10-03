package grpcclient

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func startTestServer(t *testing.T) (string, func()) {
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

func TestDiscoverServices(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := TargetConfig{
		Target:    addr,
		Plaintext: true,
	}

	services, err := DiscoverServices(ctx, cfg)
	if err != nil {
		t.Fatalf("DiscoverServices failed: %v", err)
	}

	if len(services) == 0 {
		t.Fatalf("expected at least 1 service, got 0")
	}

	var foundHealth bool
	for _, svc := range services {
		if svc.Name == "grpc.health.v1.Health" {
			foundHealth = true
			if len(svc.Methods) == 0 {
				t.Errorf("expected methods in Health service, got 0")
			}
			var foundCheck bool
			for _, m := range svc.Methods {
				if m.Name == "Check" {
					foundCheck = true
					if m.MockRequest == "" {
						t.Errorf("expected MockRequest for Check method, got empty")
					}
				}
			}
			if !foundCheck {
				t.Errorf("method Check not found in Health service")
			}
		}
	}

	if !foundHealth {
		t.Errorf("service grpc.health.v1.Health not found in %v", services)
	}
}

func TestInvokeUnarySuccess(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := InvokeRequest{
		Target:    addr,
		Plaintext: true,
		Service:   "grpc.health.v1.Health",
		Method:    "Check",
		Body:      `{"service": ""}`,
		TimeoutMs: 3000,
		Metadata: map[string]string{
			"x-test-header": "eskhan-v2",
		},
	}

	resp, err := Invoke(ctx, req)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if resp.StatusCode != "OK" {
		t.Errorf("expected status OK, got %s (msg: %s)", resp.StatusCode, resp.Message)
	}
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.RawJSON == "" {
		t.Errorf("expected non-empty RawJSON")
	}
}

func TestInvokeValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Missing target
	_, err := Invoke(ctx, InvokeRequest{})
	if err == nil {
		t.Errorf("expected error for empty target")
	}

	// Missing service
	_, err = Invoke(ctx, InvokeRequest{Target: "localhost:50051"})
	if err == nil {
		t.Errorf("expected error for empty service")
	}

	// Missing method
	_, err = Invoke(ctx, InvokeRequest{Target: "localhost:50051", Service: "test.Service"})
	if err == nil {
		t.Errorf("expected error for empty method")
	}
}

func TestInvokeInvalidMethod(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := InvokeRequest{
		Target:    addr,
		Plaintext: true,
		Service:   "grpc.health.v1.Health",
		Method:    "NonExistentMethod",
		Body:      `{}`,
	}

	_, err := Invoke(ctx, req)
	if err == nil {
		t.Fatalf("expected error for non-existent method, got nil")
	}
}

func TestParseProtoContent(t *testing.T) {
	protoStr := `
syntax = "proto3";
package bookstore.v1;

message BookRequest {
  string isbn = 1;
  int32 max_results = 2;
}

message BookResponse {
  string title = 1;
  string author = 2;
}

service BookService {
  rpc SearchBook(BookRequest) returns (BookResponse);
}
`
	services, err := ParseProtoContent(protoStr)
	if err != nil {
		t.Fatalf("ParseProtoContent failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	svc := services[0]
	if svc.Name != "bookstore.v1.BookService" {
		t.Errorf("expected bookstore.v1.BookService, got %s", svc.Name)
	}

	if len(svc.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(svc.Methods))
	}

	method := svc.Methods[0]
	if method.Name != "SearchBook" {
		t.Errorf("expected SearchBook method, got %s", method.Name)
	}

	if method.MockRequest == "" {
		t.Errorf("expected non-empty mock request")
	}
}

func TestInvokeWithVariables(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := InvokeRequest{
		Target:    "{{grpc_host}}",
		Plaintext: true,
		Service:   "{{service_name}}",
		Method:    "Check",
		Metadata: map[string]string{
			"authorization": "Bearer {{token}}",
		},
		Body: `{"service": "{{check_service}}"}`,
		Variables: map[string]string{
			"grpc_host":     addr,
			"service_name":  "grpc.health.v1.Health",
			"token":         "my-grpc-jwt",
			"check_service": "",
		},
	}

	resp, err := Invoke(ctx, req)
	if err != nil {
		t.Fatalf("Invoke with variables failed: %v", err)
	}

	if resp.StatusCode != "OK" {
		t.Errorf("expected OK status, got %s", resp.StatusCode)
	}
}
