package openapi

// SpecInfo contains metadata about the parsed OpenAPI/Swagger document.
type SpecInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
	BaseURL     string `json:"base_url"`
}

// Parameter represents a query, header, or path parameter.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // "query", "header", "path"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     any    `json:"default,omitempty"`
}

// Endpoint represents a single API operation.
type Endpoint struct {
	ID              string            `json:"id"`
	Method          string            `json:"method"` // GET, POST, PUT, DELETE, PATCH, etc.
	Path            string            `json:"path"`
	Summary         string            `json:"summary"`
	Description     string            `json:"description"`
	Tags            []string          `json:"tags"`
	Parameters      []Parameter       `json:"parameters"`
	RequestBodyType string            `json:"request_body_type,omitempty"` // "json", "form_data", etc.
	MockBody        string            `json:"mock_body,omitempty"`         // Auto-generated sample JSON
	Responses       map[string]string `json:"responses,omitempty"`         // code -> description
}

// ParsedSpec contains the complete parsed structure ready for IDE visualization and 1-click testing.
type ParsedSpec struct {
	Info      SpecInfo              `json:"info"`
	Tags      []string              `json:"tags"`
	Endpoints []Endpoint            `json:"endpoints"`
	ByTag     map[string][]Endpoint `json:"by_tag"`
}
