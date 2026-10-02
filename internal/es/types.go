package es

import "time"

// ClusterInfo contains general information returned by the root endpoint.
type ClusterInfo struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number                           string    `json:"number"`
		BuildFlavor                      string    `json:"build_flavor,omitempty"`
		BuildType                        string    `json:"build_type,omitempty"`
		BuildHash                        string    `json:"build_hash,omitempty"`
		BuildDate                        time.Time `json:"build_date,omitempty"`
		BuildSnapshot                    bool      `json:"build_snapshot,omitempty"`
		LuceneVersion                    string    `json:"lucene_version,omitempty"`
		MinimumWireCompatibilityVersion  string    `json:"minimum_wire_compatibility_version,omitempty"`
		MinimumIndexCompatibilityVersion string    `json:"minimum_index_compatibility_version,omitempty"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

// ClusterHealth contains health status returned by /_cluster/health.
type ClusterHealth struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"` // green, yellow, red
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// IndexItem represents an Elasticsearch index returned by _cat/indices.
type IndexItem struct {
	Name         string   `json:"name"`
	Health       string   `json:"health"` // green, yellow, red
	Status       string   `json:"status"` // open, close
	DocsCount    int64    `json:"docs_count"`
	StoreSize    string   `json:"store_size"`
	PrimaryShard int      `json:"pri"`
	ReplicaShard int      `json:"rep"`
	Aliases      []string `json:"aliases,omitempty"`
}

// QueryRequest represents an incoming query execution request from the UI.
type QueryRequest struct {
	RawInput  string            `json:"raw_input"` // e.g. "GET /my-index/_search\n{...}" or pure JSON
	Method    string            `json:"method"`    // GET, POST, PUT, DELETE
	Path      string            `json:"path"`      // e.g. "/my-index/_search"
	Body      string            `json:"body"`      // JSON body
	Index     string            `json:"index"`     // Target index if not in path
	Variables map[string]string `json:"variables,omitempty"`
}

// ShardsInfo reflects shard execution details.
type ShardsInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

// QueryResult represents the processed outcome of executing a query.
type QueryResult struct {
	StatusCode        int                      `json:"status_code"`
	StatusText        string                   `json:"status_text"`
	LatencyMs         int64                    `json:"latency_ms"`
	TookMs            int64                    `json:"took_ms"`
	TimedOut          bool                     `json:"timed_out"`
	TotalHits         int64                    `json:"total_hits"`
	TotalHitsRelation string                   `json:"total_hits_relation"` // "eq", "gte"
	Shards            ShardsInfo               `json:"shards"`
	RawJSON           string                   `json:"raw_json"`
	ExtractedHits     []map[string]interface{} `json:"extracted_hits"`
	HitColumns        []string                 `json:"hit_columns"`
	Aggregations      map[string]interface{}   `json:"aggregations,omitempty"`
	ErrorMessage      string                   `json:"error_message,omitempty"`
}

// AnalyzeToken represents an individual analyzed token from _analyze.
type AnalyzeToken struct {
	Token       string `json:"token"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Type        string `json:"type"`
	Position    int    `json:"position"`
}

// AnalyzeRequest holds parameters for the _analyze API.
type AnalyzeRequest struct {
	Index       string   `json:"index,omitempty"`
	Text        string   `json:"text"`
	Analyzer    string   `json:"analyzer,omitempty"`
	Tokenizer   string   `json:"tokenizer,omitempty"`
	CharFilters []string `json:"char_filter,omitempty"`
	TokenFilter []string `json:"filter,omitempty"`
}

// AnalyzeResponse holds the tokens returned by _analyze.
type AnalyzeResponse struct {
	Tokens []AnalyzeToken `json:"tokens"`
}

// DocUpdateRequest represents updating a specific document.
type DocUpdateRequest struct {
	Index string                 `json:"index"`
	ID    string                 `json:"id"`
	Doc   map[string]interface{} `json:"doc"`
}
