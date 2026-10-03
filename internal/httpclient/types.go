package httpclient

// Request represents an HTTP request to be executed by the client engine.
type Request struct {
	Method             string            `json:"method"` // GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
	URL                string            `json:"url"`
	QueryParams        map[string]string `json:"query_params,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	BodyType           string            `json:"body_type,omitempty"` // "none", "json", "raw", "x_www_form_urlencoded", "form_data"
	Body               string            `json:"body,omitempty"`
	FormData           map[string]string `json:"form_data,omitempty"`
	TimeoutMs          int               `json:"timeout_ms,omitempty"`
	InsecureSkipVerify bool              `json:"insecure_skip_verify,omitempty"`
	FollowRedirects    *bool             `json:"follow_redirects,omitempty"` // Defaults to true
	Variables          map[string]string `json:"variables,omitempty"`
}

// NetworkTimings records detailed network stages using net/http/httptrace.
type NetworkTimings struct {
	DNSLookupMs    int64 `json:"dns_lookup_ms"`
	TCPConnectMs   int64 `json:"tcp_connect_ms"`
	TLSHandshakeMs int64 `json:"tls_handshake_ms"`
	TTFBMs         int64 `json:"ttfb_ms"` // Time To First Byte
	TotalMs        int64 `json:"total_ms"`
}

// Response represents the result of an executed HTTP request.
type Response struct {
	StatusCode int                 `json:"status_code"`
	StatusText string              `json:"status_text"` // e.g. "200 OK"
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Size       int64               `json:"size"` // Body size in bytes
	Timings    NetworkTimings      `json:"timings"`
	Error      string              `json:"error,omitempty"`
}
