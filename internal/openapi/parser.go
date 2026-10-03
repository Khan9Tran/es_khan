package openapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var httpMethods = map[string]bool{
	"get":     true,
	"post":    true,
	"put":     true,
	"delete":  true,
	"patch":   true,
	"head":    true,
	"options": true,
}

// FetchAndParse fetches an OpenAPI/Swagger spec from a URL and parses it.
func FetchAndParse(ctx context.Context, specURL string) (*ParsedSpec, error) {
	if specURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", specURL, err)
	}
	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spec from %s: %w", specURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch spec, HTTP status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read spec response: %w", err)
	}

	spec, err := ParseSpec(bodyBytes)
	if err != nil {
		return nil, err
	}

	// If BaseURL is relative or empty, default to the origin of specURL
	if spec.Info.BaseURL == "" || strings.HasPrefix(spec.Info.BaseURL, "/") {
		if parsed, err := url.Parse(specURL); err == nil {
			origin := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
			if strings.HasPrefix(spec.Info.BaseURL, "/") {
				spec.Info.BaseURL = origin + spec.Info.BaseURL
			} else {
				spec.Info.BaseURL = origin
			}
		}
	}

	return spec, nil
}

// ParseSpec parses raw JSON or YAML OpenAPI/Swagger document bytes.
func ParseSpec(content []byte) (*ParsedSpec, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("spec content cannot be empty")
	}

	var rootDoc map[string]any

	// 1. Try parsing as JSON first
	if err := json.Unmarshal(content, &rootDoc); err != nil {
		// 2. If JSON fails, try YAML
		var rawYAML any
		if yamlErr := yaml.Unmarshal(content, &rawYAML); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: %w", yamlErr)
		}
		cleaned := cleanYAMLNode(rawYAML)
		m, ok := cleaned.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("spec root must be a YAML/JSON object")
		}
		rootDoc = m
	}

	spec := &ParsedSpec{
		ByTag: make(map[string][]Endpoint),
	}

	// Extract Info & Base URL
	extractInfo(rootDoc, spec)

	// Extract Tags
	tagSet := make(map[string]bool)
	if tagsArr, ok := rootDoc["tags"].([]any); ok {
		for _, t := range tagsArr {
			if tm, ok := t.(map[string]any); ok {
				if name, ok := tm["name"].(string); ok && name != "" {
					tagSet[name] = true
				}
			}
		}
	}

	// Extract Paths and Endpoints
	pathsMap, _ := rootDoc["paths"].(map[string]any)
	for pathStr, pathItem := range pathsMap {
		pathObj, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}

		// Common path parameters
		var commonParams []Parameter
		if paramsArr, ok := pathObj["parameters"].([]any); ok {
			commonParams = parseParameters(paramsArr)
		}

		for methodKey, methodItem := range pathObj {
			methodLower := strings.ToLower(methodKey)
			if !httpMethods[methodLower] {
				continue
			}

			opObj, ok := methodItem.(map[string]any)
			if !ok {
				continue
			}

			endpoint := Endpoint{
				ID:          fmt.Sprintf("%s_%s", strings.ToUpper(methodLower), pathStr),
				Method:      strings.ToUpper(methodLower),
				Path:        pathStr,
				Summary:     getString(opObj, "summary"),
				Description: getString(opObj, "description"),
				Responses:   make(map[string]string),
			}

			// Tags
			if tags, ok := opObj["tags"].([]any); ok && len(tags) > 0 {
				for _, t := range tags {
					if ts, ok := t.(string); ok && ts != "" {
						endpoint.Tags = append(endpoint.Tags, ts)
						tagSet[ts] = true
					}
				}
			}
			if len(endpoint.Tags) == 0 {
				endpoint.Tags = []string{"General"}
				tagSet["General"] = true
			}

			// Operation parameters
			endpoint.Parameters = append(endpoint.Parameters, commonParams...)
			if paramsArr, ok := opObj["parameters"].([]any); ok {
				endpoint.Parameters = append(endpoint.Parameters, parseParameters(paramsArr)...)
			}

			// Request Body & Mock Payload
			extractRequestBody(opObj, rootDoc, &endpoint)

			// Responses
			if respMap, ok := opObj["responses"].(map[string]any); ok {
				for code, respItem := range respMap {
					if respObj, ok := respItem.(map[string]any); ok {
						endpoint.Responses[code] = getString(respObj, "description")
					}
				}
			}

			spec.Endpoints = append(spec.Endpoints, endpoint)
			for _, tag := range endpoint.Tags {
				spec.ByTag[tag] = append(spec.ByTag[tag], endpoint)
			}
		}
	}

	// Sort tags
	for t := range tagSet {
		spec.Tags = append(spec.Tags, t)
	}
	sort.Strings(spec.Tags)

	// Sort endpoints inside each tag
	for tag := range spec.ByTag {
		sort.Slice(spec.ByTag[tag], func(i, j int) bool {
			if spec.ByTag[tag][i].Path == spec.ByTag[tag][j].Path {
				return spec.ByTag[tag][i].Method < spec.ByTag[tag][j].Method
			}
			return spec.ByTag[tag][i].Path < spec.ByTag[tag][j].Path
		})
	}

	return spec, nil
}

func extractInfo(rootDoc map[string]any, spec *ParsedSpec) {
	if infoObj, ok := rootDoc["info"].(map[string]any); ok {
		spec.Info.Title = getString(infoObj, "title")
		spec.Info.Version = getString(infoObj, "version")
		spec.Info.Description = getString(infoObj, "description")
	}

	// OpenAPI 3.x servers
	if serversArr, ok := rootDoc["servers"].([]any); ok && len(serversArr) > 0 {
		if server0, ok := serversArr[0].(map[string]any); ok {
			spec.Info.BaseURL = getString(server0, "url")
		}
	}

	// Swagger 2.0 host & basePath
	if spec.Info.BaseURL == "" {
		host := getString(rootDoc, "host")
		basePath := getString(rootDoc, "basePath")
		schemes, _ := rootDoc["schemes"].([]any)
		scheme := "http"
		if len(schemes) > 0 {
			if s, ok := schemes[0].(string); ok && s != "" {
				scheme = s
			}
		}
		if host != "" {
			spec.Info.BaseURL = fmt.Sprintf("%s://%s%s", scheme, host, basePath)
		} else if basePath != "" {
			spec.Info.BaseURL = basePath
		}
	}
}

func extractRequestBody(opObj map[string]any, rootDoc map[string]any, endpoint *Endpoint) {
	// 1. OpenAPI 3.x requestBody
	if rb, ok := opObj["requestBody"].(map[string]any); ok {
		if contentMap, ok := rb["content"].(map[string]any); ok {
			if jsonContent, ok := contentMap["application/json"].(map[string]any); ok {
				endpoint.RequestBodyType = "json"
				if schema, ok := jsonContent["schema"].(map[string]any); ok {
					endpoint.MockBody = GenerateMockFromSchema(schema, rootDoc)
				}
				return
			}
			// Check form data
			if formContent, ok := contentMap["application/x-www-form-urlencoded"].(map[string]any); ok {
				endpoint.RequestBodyType = "x_www_form_urlencoded"
				if schema, ok := formContent["schema"].(map[string]any); ok {
					endpoint.MockBody = GenerateMockFromSchema(schema, rootDoc)
				}
				return
			}
			if multipartContent, ok := contentMap["multipart/form-data"].(map[string]any); ok {
				endpoint.RequestBodyType = "form_data"
				if schema, ok := multipartContent["schema"].(map[string]any); ok {
					endpoint.MockBody = GenerateMockFromSchema(schema, rootDoc)
				}
				return
			}
		}
	}

	// 2. Swagger 2.0 parameters where in == "body"
	if paramsArr, ok := opObj["parameters"].([]any); ok {
		for _, p := range paramsArr {
			if pm, ok := p.(map[string]any); ok {
				if getString(pm, "in") == "body" {
					endpoint.RequestBodyType = "json"
					if schema, ok := pm["schema"].(map[string]any); ok {
						endpoint.MockBody = GenerateMockFromSchema(schema, rootDoc)
					}
					return
				}
			}
		}
	}
}

func parseParameters(paramsArr []any) []Parameter {
	var results []Parameter
	for _, p := range paramsArr {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		inVal := getString(pm, "in")
		if inVal == "body" {
			continue // handled as request body
		}

		param := Parameter{
			Name:        getString(pm, "name"),
			In:          inVal,
			Required:    getBool(pm, "required"),
			Description: getString(pm, "description"),
			Default:     pm["default"],
		}

		// Type in OpenAPI 3 vs Swagger 2
		if typeStr := getString(pm, "type"); typeStr != "" {
			param.Type = typeStr
		} else if schemaObj, ok := pm["schema"].(map[string]any); ok {
			param.Type = getString(schemaObj, "type")
			if param.Default == nil {
				param.Default = schemaObj["default"]
			}
		}

		results = append(results, param)
	}
	return results
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func cleanYAMLNode(v any) any {
	switch val := v.(type) {
	case map[any]any:
		res := make(map[string]any)
		for k, valItem := range val {
			res[fmt.Sprintf("%v", k)] = cleanYAMLNode(valItem)
		}
		return res
	case map[string]any:
		res := make(map[string]any)
		for k, valItem := range val {
			res[k] = cleanYAMLNode(valItem)
		}
		return res
	case []any:
		res := make([]any, len(val))
		for i, item := range val {
			res[i] = cleanYAMLNode(item)
		}
		return res
	default:
		return val
	}
}
