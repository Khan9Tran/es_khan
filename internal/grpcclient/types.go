package grpcclient

// TargetConfig represents configuration for connecting to a gRPC target.
type TargetConfig struct {
	Target             string            `json:"target"` // e.g. "localhost:50051"
	Plaintext          bool              `json:"plaintext"`
	InsecureSkipVerify bool              `json:"insecure_skip_verify"`
	Authority          string            `json:"authority,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Variables          map[string]string `json:"variables,omitempty"`
}

// ServiceInfo represents a gRPC service and its methods discovered via reflection or proto files.
type ServiceInfo struct {
	Name    string       `json:"name"`
	Methods []MethodInfo `json:"methods"`
}

// MethodInfo represents a single RPC method in a service.
type MethodInfo struct {
	Name              string `json:"name"`
	FullName          string `json:"full_name"`
	IsClientStreaming bool   `json:"is_client_streaming"`
	IsServerStreaming bool   `json:"is_server_streaming"`
	InputType         string `json:"input_type"`
	OutputType        string `json:"output_type"`
	MockRequest       string `json:"mock_request"` // Auto-generated indented JSON template
}

// InvokeRequest represents a request to invoke an RPC method dynamically.
type InvokeRequest struct {
	Target             string            `json:"target"`
	Plaintext          bool              `json:"plaintext"`
	InsecureSkipVerify bool              `json:"insecure_skip_verify"`
	Service            string            `json:"service"` // Full service name e.g. "order.v1.OrderService"
	Method             string            `json:"method"`  // Method name e.g. "GetOrder"
	Metadata           map[string]string `json:"metadata,omitempty"`
	Body               string            `json:"body"` // JSON request body from editor
	TimeoutMs          int               `json:"timeout_ms,omitempty"`
	Variables          map[string]string `json:"variables,omitempty"`
}

// InvokeResponse represents the result of a dynamic RPC call.
type InvokeResponse struct {
	StatusCode string              `json:"status_code"` // e.g. "OK", "NOT_FOUND", "INTERNAL"
	Code       uint32              `json:"code"`        // Numeric code (0 = OK)
	Message    string              `json:"message,omitempty"`
	Response   any                 `json:"response,omitempty"` // Parsed JSON object/array
	RawJSON    string              `json:"raw_json,omitempty"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Trailers   map[string][]string `json:"trailers,omitempty"`
	TookMs     int64               `json:"took_ms"`
}
