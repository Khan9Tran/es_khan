package es

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"eskhan/internal/config"
	"eskhan/internal/variable"
)

// Client handles all HTTP interactions with an Elasticsearch cluster.
type Client struct {
	profile    config.ConnectionProfile
	httpClient *http.Client
	baseURL    string
}

// NewClient initializes a new Elasticsearch client from a connection profile.
func NewClient(profile config.ConnectionProfile) *Client {
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: profile.InsecureSkipVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	baseURL := strings.TrimRight(profile.URL, "/")

	return &Client{
		profile:    profile,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// applyAuthAndHeaders attaches authentication headers and custom headers to the request.
func (c *Client) applyAuthAndHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	switch c.profile.AuthType {
	case config.AuthBasic:
		if c.profile.Username != "" || c.profile.Password != "" {
			req.SetBasicAuth(c.profile.Username, c.profile.Password)
		}
	case config.AuthAPIKey:
		if c.profile.APIKey != "" {
			req.Header.Set("Authorization", "ApiKey "+c.profile.APIKey)
		}
	case config.AuthBearer:
		if c.profile.BearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.profile.BearerToken)
		}
	}

	for k, v := range c.profile.CustomHeaders {
		req.Header.Set(k, v)
	}
}

// Ping checks cluster connectivity and retrieves version details.
func (c *Client) Ping(ctx context.Context) (*ClusterInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ping request: %w", err)
	}
	c.applyAuthAndHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ping failed to connect to %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ping returned HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var info ClusterInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse cluster info: %w", err)
	}

	return &info, nil
}

// GetHealth fetches cluster health status from /_cluster/health.
func (c *Client) GetHealth(ctx context.Context) (*ClusterHealth, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_cluster/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health request: %w", err)
	}
	c.applyAuthAndHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	var health ClusterHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to parse cluster health: %w", err)
	}

	return &health, nil
}

// GetIndices fetches index list and basic stats via /_cat/indices?format=json.
func (c *Client) GetIndices(ctx context.Context) ([]IndexItem, error) {
	targetURL := c.baseURL + "/_cat/indices?format=json&h=index,health,status,docs.count,store.size,pri,rep"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create indices request: %w", err)
	}
	c.applyAuthAndHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch indices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indices endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rawList []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawList); err != nil {
		return nil, fmt.Errorf("failed to parse indices list: %w", err)
	}

	var items []IndexItem
	for _, raw := range rawList {
		name, _ := raw["index"].(string)
		if name == "" || strings.HasPrefix(name, ".") {
			// Skip internal system indices if needed, or include with tag
		}
		health, _ := raw["health"].(string)
		status, _ := raw["status"].(string)
		storeSize, _ := raw["store.size"].(string)

		var docsCount int64
		if dcStr, ok := raw["docs.count"].(string); ok {
			docsCount, _ = strconv.ParseInt(dcStr, 10, 64)
		} else if dcNum, ok := raw["docs.count"].(float64); ok {
			docsCount = int64(dcNum)
		}

		pri, _ := strconv.Atoi(fmt.Sprintf("%v", raw["pri"]))
		rep, _ := strconv.Atoi(fmt.Sprintf("%v", raw["rep"]))

		items = append(items, IndexItem{
			Name:         name,
			Health:       health,
			Status:       status,
			DocsCount:    docsCount,
			StoreSize:    storeSize,
			PrimaryShard: pri,
			ReplicaShard: rep,
		})
	}

	// Sort indices alphabetically
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

// GetMapping fetches raw mapping definition for a specific index or all indices.
func (c *Client) GetMapping(ctx context.Context, indexName string) (map[string]interface{}, error) {
	path := "/_mapping"
	if indexName != "" && indexName != "*" {
		path = "/" + url.PathEscape(indexName) + "/_mapping"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mapping request: %w", err)
	}
	c.applyAuthAndHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mapping: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mapping json: %w", err)
	}

	return result, nil
}

// ParseRawInput parses Kibana DevTools style or pure JSON requests into method, path, and body.
// Example:
//
//	POST my-index/_search
//	{ "query": { "match_all": {} } }
//
// or pure JSON:
//
//	{ "query": { ... } }
func ParseRawInput(raw string, defaultIndex string) (method, path string, body []byte) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return http.MethodGet, "/_search", nil
	}

	lines := strings.Split(trimmed, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Check if first line starts with HTTP method
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
	hasMethod := false
	for _, m := range methods {
		if strings.HasPrefix(strings.ToUpper(firstLine), m+" ") {
			hasMethod = true
			method = m
			remainder := strings.TrimSpace(firstLine[len(m):])
			path = remainder
			break
		}
	}

	if hasMethod {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if len(lines) > 1 {
			bodyText := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			if bodyText != "" {
				body = []byte(bodyText)
			}
		}
		return method, path, body
	}

	// If no method in first line, check if it's pure JSON
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		idx := defaultIndex
		if idx == "" {
			path = "/_search"
		} else {
			path = "/" + url.PathEscape(idx) + "/_search"
		}
		return http.MethodPost, path, []byte(trimmed)
	}

	// Just a path line like "/_cat/indices"
	if strings.HasPrefix(firstLine, "/") {
		path = firstLine
		method = http.MethodGet
		if len(lines) > 1 {
			bodyText := strings.TrimSpace(strings.Join(lines[1:], "\n"))
			if bodyText != "" {
				body = []byte(bodyText)
				method = http.MethodPost
			}
		}
		return method, path, body
	}

	// Fallback
	return http.MethodGet, "/" + firstLine, nil
}

// ExecuteQuery executes a query against Elasticsearch and processes the response.
func (c *Client) ExecuteQuery(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	method := req.Method
	path := req.Path
	var bodyBytes []byte

	rawInput := req.RawInput
	if rawInput != "" {
		rawInput = SubstituteVariables(rawInput, req.Variables)
		method, path, bodyBytes = ParseRawInput(rawInput, req.Index)
	} else {
		if method == "" {
			method = http.MethodPost
		}
		if path == "" {
			if req.Index != "" {
				path = "/" + url.PathEscape(req.Index) + "/_search"
			} else {
				path = "/_search"
			}
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if req.Body != "" {
			bodyText := SubstituteVariables(req.Body, req.Variables)
			bodyBytes = []byte(bodyText)
		}
	}

	targetURL := c.baseURL + path
	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}
	c.applyAuthAndHeaders(httpReq)

	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		return &QueryResult{
			StatusCode:   0,
			StatusText:   "Connection Error",
			LatencyMs:    latency,
			ErrorMessage: err.Error(),
		}, nil
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &QueryResult{
			StatusCode:   httpResp.StatusCode,
			StatusText:   httpResp.Status,
			LatencyMs:    latency,
			ErrorMessage: fmt.Sprintf("failed to read response body: %v", err),
		}, nil
	}

	result := &QueryResult{
		StatusCode: httpResp.StatusCode,
		StatusText: httpResp.Status,
		LatencyMs:  latency,
		RawJSON:    string(respBytes),
	}

	// Parse JSON response details
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(respBytes, &parsed); jsonErr == nil {
		// Extract took
		if took, ok := parsed["took"].(float64); ok {
			result.TookMs = int64(took)
		}
		if timedOut, ok := parsed["timed_out"].(bool); ok {
			result.TimedOut = timedOut
		}

		// Extract shards
		if shardsMap, ok := parsed["_shards"].(map[string]interface{}); ok {
			if total, ok := shardsMap["total"].(float64); ok {
				result.Shards.Total = int(total)
			}
			if succ, ok := shardsMap["successful"].(float64); ok {
				result.Shards.Successful = int(succ)
			}
			if skip, ok := shardsMap["skipped"].(float64); ok {
				result.Shards.Skipped = int(skip)
			}
			if fail, ok := shardsMap["failed"].(float64); ok {
				result.Shards.Failed = int(fail)
			}
		}

		// Extract aggregations
		if aggs, ok := parsed["aggregations"].(map[string]interface{}); ok {
			result.Aggregations = aggs
		}

		// Extract hits and total (Compatibility ES 7.x vs 8.x)
		if hitsMap, ok := parsed["hits"].(map[string]interface{}); ok {
			// Handle hits.total: could be number or {value, relation}
			if totalNum, ok := hitsMap["total"].(float64); ok {
				result.TotalHits = int64(totalNum)
				result.TotalHitsRelation = "eq"
			} else if totalObj, ok := hitsMap["total"].(map[string]interface{}); ok {
				if val, ok := totalObj["value"].(float64); ok {
					result.TotalHits = int64(val)
				}
				if rel, ok := totalObj["relation"].(string); ok {
					result.TotalHitsRelation = rel
				}
			}

			// Extract hits list for table view
			if hitList, ok := hitsMap["hits"].([]interface{}); ok {
				columnsSet := make(map[string]bool)
				var rows []map[string]interface{}

				for _, h := range hitList {
					hitObj, ok := h.(map[string]interface{})
					if !ok {
						continue
					}

					row := make(map[string]interface{})
					if id, ok := hitObj["_id"].(string); ok {
						row["_id"] = id
						columnsSet["_id"] = true
					}
					if score, ok := hitObj["_score"].(float64); ok {
						row["_score"] = score
						columnsSet["_score"] = true
					}
					if source, ok := hitObj["_source"].(map[string]interface{}); ok {
						for k, v := range source {
							row[k] = v
							columnsSet[k] = true
						}
					}
					rows = append(rows, row)
				}

				result.ExtractedHits = rows

				// Order columns: _id, _score, then alphabetized fields
				var fieldCols []string
				hasID := false
				hasScore := false
				for col := range columnsSet {
					if col == "_id" {
						hasID = true
					} else if col == "_score" {
						hasScore = true
					} else {
						fieldCols = append(fieldCols, col)
					}
				}
				sort.Strings(fieldCols)

				var ordered []string
				if hasID {
					ordered = append(ordered, "_id")
				}
				if hasScore {
					ordered = append(ordered, "_score")
				}
				ordered = append(ordered, fieldCols...)
				result.HitColumns = ordered
			}
		}

		// Check for top-level error in ES response
		if errObj, ok := parsed["error"].(map[string]interface{}); ok {
			if reason, ok := errObj["reason"].(string); ok {
				result.ErrorMessage = reason
			} else {
				errBytes, _ := json.Marshal(errObj)
				result.ErrorMessage = string(errBytes)
			}
		}
	} else {
		// Non-JSON response (e.g. text/plain from _cat)
		if httpResp.StatusCode >= 400 {
			result.ErrorMessage = string(respBytes)
		}
	}

	return result, nil
}

// SubstituteVariables replaces template variables like {{var}}, {{$timestamp}}, {{$uuid}}, {{$date}}, {{$randomInt}}.
func SubstituteVariables(input string, userVars map[string]string) string {
	return variable.Eval(input, userVars)
}

// AnalyzeText runs text through the _analyze API to test tokenizers and filters.
func (c *Client) AnalyzeText(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	path := "/_analyze"
	if req.Index != "" {
		path = "/" + url.PathEscape(req.Index) + "/_analyze"
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analyze request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create analyze request: %w", err)
	}
	c.applyAuthAndHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("analyze request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("analyze returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var res AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to parse analyze response: %w", err)
	}

	return &res, nil
}

// GetTopTerms runs an aggregation to retrieve the most frequent terms of a keyword field.
func (c *Client) GetTopTerms(ctx context.Context, indexName, field string, size int) ([]string, error) {
	if size <= 0 {
		size = 10
	}

	path := "/" + url.PathEscape(indexName) + "/_search"
	queryMap := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"top_terms": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": field,
					"size":  size,
				},
			},
		},
	}

	payload, _ := json.Marshal(queryMap)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	c.applyAuthAndHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	var terms []string
	if aggs, ok := parsed["aggregations"].(map[string]interface{}); ok {
		if topAgg, ok := aggs["top_terms"].(map[string]interface{}); ok {
			if buckets, ok := topAgg["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					if bMap, ok := b.(map[string]interface{}); ok {
						if keyStr, ok := bMap["key_as_string"].(string); ok {
							terms = append(terms, keyStr)
						} else if key, exists := bMap["key"]; exists {
							terms = append(terms, fmt.Sprintf("%v", key))
						}
					}
				}
			}
		}
	}

	return terms, nil
}

// UpdateDocument updates an individual document via POST /{index}/_update/{id}.
func (c *Client) UpdateDocument(ctx context.Context, index, id string, doc map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/%s/_update/%s", url.PathEscape(index), url.PathEscape(id))
	payload, _ := json.Marshal(map[string]interface{}{
		"doc": doc,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	c.applyAuthAndHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("update returned HTTP %d", resp.StatusCode)
	}

	return result, nil
}

// DeleteDocument deletes an individual document via DELETE /{index}/_doc/{id}.
func (c *Client) DeleteDocument(ctx context.Context, index, id string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(index), url.PathEscape(id))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	c.applyAuthAndHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("delete returned HTTP %d", resp.StatusCode)
	}

	return result, nil
}

// ExecuteIndexAction executes an administrative maintenance action (refresh, flush, cache/clear, forcemerge).
func (c *Client) ExecuteIndexAction(ctx context.Context, index, action string) (map[string]interface{}, error) {
	validActions := map[string]string{
		"refresh":     "/_refresh",
		"flush":       "/_flush",
		"cache_clear": "/_cache/clear",
		"forcemerge":  "/_forcemerge",
	}

	actPath, valid := validActions[action]
	if !valid {
		return nil, fmt.Errorf("invalid index action: %s", action)
	}

	path := "/" + url.PathEscape(index) + actPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	c.applyAuthAndHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("action %s returned HTTP %d", action, resp.StatusCode)
	}

	return result, nil
}
