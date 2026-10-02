package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"eskhan/internal/ai"
)

// User's exact DSL definitions (for reference, compilation, and evaluation):
type EsFiltersRange struct {
	Key string
	Min interface{}
	Max interface{}
}

type EsFilters struct {
	Type  string // support for term, wildcard and match
	Key   string
	Value any
}

const (
	ASC  = "asc"
	DESC = "desc"
)

type EsSort struct {
	Key   string
	Order string
}

type ESQuery map[string]interface{}
type ESQueryArr []map[string]interface{}
type EsFiltersArr []EsFilters
type EsRangesArr []EsFiltersRange
type EsSortArr []EsSort

func (f EsFiltersArr) Add(input EsFilters) EsFiltersArr {
	f = append(f, input)
	return f
}

func (f EsFiltersArr) Exists(field string) EsFiltersArr {
	f = append(f, EsFilters{Key: "field", Type: "exists", Value: field})
	return f
}

func (f EsFiltersArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		if f[i].Type == "wildcard" {
			value := map[string]interface{}{}
			value["value"] = f[i].Value
			value["case_insensitive"] = true
			rs = append(rs, map[string]interface{}{
				f[i].Type: map[string]interface{}{
					f[i].Key: value,
				},
			})
		} else {
			rs = append(rs, map[string]interface{}{
				f[i].Type: map[string]interface{}{
					f[i].Key: f[i].Value,
				},
			})
		}
	}
	return rs
}

func (f EsRangesArr) Add(input EsFiltersRange) EsRangesArr {
	f = append(f, input)
	return f
}

func (f EsRangesArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		rs = append(rs, map[string]interface{}{
			"range": map[string]interface{}{
				f[i].Key: map[string]interface{}{
					"gte": f[i].Min,
					"lte": f[i].Max,
				},
			},
		})
	}
	return rs
}

func (f EsSortArr) Add(input EsSort) EsSortArr {
	f = append(f, input)
	return f
}

func (f EsSortArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		rs = append(rs, map[string]interface{}{
			f[i].Key: map[string]interface{}{
				"order": f[i].Order,
			},
		})
	}
	return rs
}

func (f ESQueryArr) Must() ESQuery {
	return ESQuery(map[string]interface{}{
		"must": f,
	})
}

func (f ESQueryArr) MustNot() ESQuery {
	return ESQuery(map[string]interface{}{
		"must_not": f,
	})
}

func (f ESQueryArr) Should() ESQuery {
	return ESQuery(map[string]interface{}{
		"should": f,
	})
}

func (f ESQuery) Bool() ESQuery {
	return ESQuery(map[string]interface{}{
		"bool": f,
	})
}

func (f ESQuery) Query() ESQuery {
	return ESQuery(map[string]interface{}{
		"query": f,
	})
}

func (f ESQuery) Nested(nestedName string) ESQuery {
	return ESQuery(map[string]interface{}{
		"nested": map[string]interface{}{
			"path":  nestedName,
			"query": f,
		},
	})
}

func (f ESQuery) Sort(sort EsSortArr) ESQuery {
	f["sort"] = sort
	return f
}

func (f ESQuery) Paging(page, size int) ESQuery {
	f["size"] = size
	f["from"] = page * size
	return f
}

func QueryMatchAll() ESQuery {
	return ESQuery(map[string]interface{}{
		"match_all": map[string]interface{}{},
	}).Query()
}

// Converter handles bidirectional translation between Query DSL and Go code.
type Converter struct {
	aiBridge *ai.Bridge
}

// NewConverter creates a new DSL converter.
func NewConverter(bridge *ai.Bridge) *Converter {
	return &Converter{aiBridge: bridge}
}

// CleanJSONInput strips Kibana REST line (e.g. POST /index/_search) to keep only the JSON body.
func CleanJSONInput(raw string) string {
	trimmed := strings.TrimSpace(raw)
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return ""
	}
	firstLine := strings.TrimSpace(lines[0])
	if regexp.MustCompile(`^(?i)(GET|POST|PUT|DELETE)\s+`).MatchString(firstLine) || strings.HasPrefix(firstLine, "/") {
		return strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}
	return trimmed
}

// QueryToGolang converts Elasticsearch Query DSL JSON into Go code utilizing user's DSL structs.
func (c *Converter) QueryToGolang(jsonDSL string) (string, error) {
	cleanJSON := CleanJSONInput(jsonDSL)
	if cleanJSON == "" {
		return "// Cú pháp truy vấn rỗng\nquery := QueryMatchAll()", nil
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(cleanJSON), &root); err != nil {
		return "", fmt.Errorf("invalid JSON query: %w", err)
	}

	// 1. Check for match_all
	queryObj, hasQuery := root["query"].(map[string]interface{})
	if hasQuery {
		if _, isMatchAll := queryObj["match_all"]; isMatchAll && len(queryObj) == 1 {
			var sb strings.Builder
			sb.WriteString("query := QueryMatchAll()")
			appendPagingAndSort(&sb, root)
			return sb.String(), nil
		}
	} else if _, isMatchAll := root["match_all"]; isMatchAll {
		var sb strings.Builder
		sb.WriteString("query := QueryMatchAll()")
		appendPagingAndSort(&sb, root)
		return sb.String(), nil
	}

	// 2. Parse bool query clauses
	var boolObj map[string]interface{}
	if hasQuery {
		if b, ok := queryObj["bool"].(map[string]interface{}); ok {
			boolObj = b
		}
	} else if b, ok := root["bool"].(map[string]interface{}); ok {
		boolObj = b
	}

	var sb strings.Builder
	hasClauses := false

	// Handle Must / Filter / Should / MustNot
	clauseTypes := []struct {
		Key      string
		Method   string
		VarName  string
		ArrQuery string
	}{
		{"must", "Must()", "mustFilters", "mustRanges"},
		{"filter", "Must()", "filterList", "filterRanges"},
		{"should", "Should()", "shouldFilters", "shouldRanges"},
		{"must_not", "MustNot()", "mustNotFilters", "mustNotRanges"},
	}

	var queryArrays []string
	mainMethod := "Must()"

	for _, ct := range clauseTypes {
		if boolObj == nil {
			continue
		}
		rawClause, exists := boolObj[ct.Key]
		if !exists {
			continue
		}

		var items []interface{}
		switch v := rawClause.(type) {
		case []interface{}:
			items = v
		case map[string]interface{}:
			items = []interface{}{v}
		}

		var filterLines []string
		var rangeLines []string
		var nestedLines []string

		for _, itm := range items {
			itmMap, ok := itm.(map[string]interface{})
			if !ok {
				continue
			}

			// term
			if termObj, ok := itmMap["term"].(map[string]interface{}); ok {
				for k, v := range termObj {
					filterLines = append(filterLines, fmt.Sprintf("\tAdd(EsFilters{Type: \"term\", Key: %q, Value: %s})", k, formatGoValue(v)))
				}
				continue
			}

			// match
			if matchObj, ok := itmMap["match"].(map[string]interface{}); ok {
				for k, v := range matchObj {
					filterLines = append(filterLines, fmt.Sprintf("\tAdd(EsFilters{Type: \"match\", Key: %q, Value: %s})", k, formatGoValue(v)))
				}
				continue
			}

			// wildcard
			if wcObj, ok := itmMap["wildcard"].(map[string]interface{}); ok {
				for k, v := range wcObj {
					val := v
					if vMap, isMap := v.(map[string]interface{}); isMap {
						if innerVal, hasVal := vMap["value"]; hasVal {
							val = innerVal
						}
					}
					filterLines = append(filterLines, fmt.Sprintf("\tAdd(EsFilters{Type: \"wildcard\", Key: %q, Value: %s})", k, formatGoValue(val)))
				}
				continue
			}

			// exists
			if exObj, ok := itmMap["exists"].(map[string]interface{}); ok {
				if fld, ok := exObj["field"].(string); ok {
					filterLines = append(filterLines, fmt.Sprintf("\tExists(%q)", fld))
				}
				continue
			}

			// range
			if rangeObj, ok := itmMap["range"].(map[string]interface{}); ok {
				for k, v := range rangeObj {
					if conds, ok := v.(map[string]interface{}); ok {
						minVal := conds["gte"]
						if minVal == nil {
							minVal = conds["gt"]
						}
						maxVal := conds["lte"]
						if maxVal == nil {
							maxVal = conds["lt"]
						}
						rangeLines = append(rangeLines, fmt.Sprintf("\tAdd(EsFiltersRange{Key: %q, Min: %s, Max: %s})", k, formatGoValue(minVal), formatGoValue(maxVal)))
					}
				}
				continue
			}

			// nested
			if nestedObj, ok := itmMap["nested"].(map[string]interface{}); ok {
				path, _ := nestedObj["path"].(string)
				subQuery, _ := nestedObj["query"].(map[string]interface{})
				nestedLines = append(nestedLines, fmt.Sprintf("// Nested path: %s (subquery: %v)", path, subQuery))
			}
		}

		if len(filterLines) > 0 {
			hasClauses = true
			fVar := ct.VarName
			sb.WriteString(fmt.Sprintf("%s := EsFiltersArr{}.\n%s\n\n", fVar, strings.Join(filterLines, ".\n")))
			queryArrays = append(queryArrays, fmt.Sprintf("%s.Query()", fVar))
			mainMethod = ct.Method
		}

		if len(rangeLines) > 0 {
			hasClauses = true
			rVar := ct.ArrQuery
			sb.WriteString(fmt.Sprintf("%s := EsRangesArr{}.\n%s\n\n", rVar, strings.Join(rangeLines, ".\n")))
			queryArrays = append(queryArrays, fmt.Sprintf("%s.Query()", rVar))
			mainMethod = ct.Method
		}
	}

	// 3. Assemble query
	if hasClauses {
		if len(queryArrays) == 1 {
			sb.WriteString(fmt.Sprintf("query := %s.%s.Bool().Query()", queryArrays[0], mainMethod))
		} else {
			sb.WriteString(fmt.Sprintf("queryArr := append(%s)\n", strings.Join(queryArrays, ", ...)\nqueryArr = append(queryArr, ")))
			sb.WriteString(fmt.Sprintf("query := queryArr.%s.Bool().Query()", mainMethod))
		}
		appendPagingAndSort(&sb, root)
		return sb.String(), nil
	}

	// 4. Fallback: If not parsed into bool clauses, use Antigravity AI Bridge if available
	if c.aiBridge != nil && c.aiBridge.Status().Available {
		aiCode, err := c.convertQueryViaAI(cleanJSON)
		if err == nil && aiCode != "" {
			return aiCode, nil
		}
	}

	// Fallback pure builder
	var fallbackSB strings.Builder
	fallbackSB.WriteString("// Truy vấn cấu trúc tùy biến:\n")
	fallbackSB.WriteString("query := QueryMatchAll()")
	appendPagingAndSort(&fallbackSB, root)
	return fallbackSB.String(), nil
}

// GolangToQuery parses user's Go code and generates the equivalent Elasticsearch Query DSL JSON.
func (c *Converter) GolangToQuery(goCode string) (string, error) {
	trimmed := strings.TrimSpace(goCode)
	if trimmed == "" {
		return "{\n  \"query\": {\n    \"match_all\": {}\n  }\n}", nil
	}

	// 1. Check for QueryMatchAll
	if strings.Contains(trimmed, "QueryMatchAll") {
		result := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		parsePagingAndSortFromCode(trimmed, result)
		pretty, _ := json.MarshalIndent(result, "", "  ")
		return string(pretty), nil
	}

	// 2. Parse EsFilters{Type: "...", Key: "...", Value: ...}
	filterRegex := regexp.MustCompile(`EsFilters\s*\{\s*Type:\s*"([^"]+)",\s*Key:\s*"([^"]+)",\s*Value:\s*([^}]+)\}`)
	matches := filterRegex.FindAllStringSubmatch(trimmed, -1)

	var filterItems []map[string]interface{}
	for _, m := range matches {
		fType := m[1]
		fKey := m[2]
		rawVal := strings.TrimSpace(m[3])
		val := parseGoLiteral(rawVal)

		if fType == "wildcard" {
			filterItems = append(filterItems, map[string]interface{}{
				"wildcard": map[string]interface{}{
					fKey: map[string]interface{}{
						"value":            val,
						"case_insensitive": true,
					},
				},
			})
		} else {
			filterItems = append(filterItems, map[string]interface{}{
				fType: map[string]interface{}{
					fKey: val,
				},
			})
		}
	}

	// Parse Exists("field") or .Exists("field")
	existsRegex := regexp.MustCompile(`(?:\.)?Exists\s*\(\s*"([^"]+)"\s*\)`)
	exMatches := existsRegex.FindAllStringSubmatch(trimmed, -1)
	for _, em := range exMatches {
		filterItems = append(filterItems, map[string]interface{}{
			"exists": map[string]interface{}{
				"field": em[1],
			},
		})
	}

	// Parse EsFiltersRange{Key: "...", Min: ..., Max: ...}
	rangeRegex := regexp.MustCompile(`EsFiltersRange\s*\{\s*Key:\s*"([^"]+)",\s*Min:\s*([^,]+),\s*Max:\s*([^}]+)\}`)
	rangeMatches := rangeRegex.FindAllStringSubmatch(trimmed, -1)
	for _, rm := range rangeMatches {
		rKey := rm[1]
		minRaw := strings.TrimSpace(rm[2])
		maxRaw := strings.TrimSpace(rm[3])

		rangeBody := map[string]interface{}{}
		if minRaw != "nil" && minRaw != `""` {
			rangeBody["gte"] = parseGoLiteral(minRaw)
		}
		if maxRaw != "nil" && maxRaw != `""` {
			rangeBody["lte"] = parseGoLiteral(maxRaw)
		}

		filterItems = append(filterItems, map[string]interface{}{
			"range": map[string]interface{}{
				rKey: rangeBody,
			},
		})
	}

	// Determine boolean clause (must vs should vs must_not)
	clause := "must"
	if strings.Contains(trimmed, ".Should()") {
		clause = "should"
	} else if strings.Contains(trimmed, ".MustNot()") {
		clause = "must_not"
	}

	if len(filterItems) > 0 {
		boolMap := map[string]interface{}{
			clause: filterItems,
		}
		queryResult := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": boolMap,
			},
		}

		// Check Nested
		nestedRegex := regexp.MustCompile(`\.Nested\s*\(\s*"([^"]+)"\s*\)`)
		if nMatch := nestedRegex.FindStringSubmatch(trimmed); len(nMatch) > 1 {
			queryResult["query"] = map[string]interface{}{
				"nested": map[string]interface{}{
					"path":  nMatch[1],
					"query": queryResult["query"],
				},
			}
		}

		parsePagingAndSortFromCode(trimmed, queryResult)
		pretty, _ := json.MarshalIndent(queryResult, "", "  ")
		return string(pretty), nil
	}

	// 3. Fallback to Antigravity AI Bridge if available
	if c.aiBridge != nil && c.aiBridge.Status().Available {
		dsl, err := c.convertCodeViaAI(trimmed)
		if err == nil && dsl != "" {
			return dsl, nil
		}
	}

	return "", fmt.Errorf("không thể suy luận Query DSL từ mã nguồn Go này. Vui lòng sử dụng cú pháp DSL chuẩn (EsFilters, EsFiltersRange, EsSort, QueryMatchAll)")
}

func (c *Converter) convertQueryViaAI(jsonDSL string) (string, error) {
	prompt := fmt.Sprintf(
		"Bạn là chuyên gia Golang và Elasticsearch. Hãy chuyển đổi câu truy vấn Elasticsearch Query DSL JSON sau thành mã nguồn Golang dựa CHÍNH XÁC vào cấu trúc builder pattern sau đây:\n\n"+
			"```go\n"+
			"type EsFiltersRange struct { Key string; Min interface{}; Max interface{} }\n"+
			"type EsFilters struct { Type string; Key string; Value any }\n"+
			"type EsSort struct { Key string; Order string }\n"+
			"// methods: EsFiltersArr{}.Add(...).Exists(...).Query()\n"+
			"// EsRangesArr{}.Add(...).Query()\n"+
			"// EsSortArr{}.Add(...).Query()\n"+
			"// .Must().Bool().Query(), .Should().Bool().Query(), .Paging(page, size), .Sort(sorts)\n"+
			"// QueryMatchAll()\n"+
			"```\n\n"+
			"Query DSL cần chuyển đổi:\n```json\n%s\n```\n\n"+
			"QUY TẮC: Chỉ trả về đoạn code Go hợp lệ trong cặp thẻ ```go và ```, không giải thích dông dài.",
		jsonDSL,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	res, err := c.aiBridge.ExplainQuery(ctx, prompt, "")
	if err != nil {
		return "", err
	}

	code := extractCodeBlock(res, "go")
	if code == "" {
		code = res
	}
	return strings.TrimSpace(code), nil
}

func (c *Converter) convertCodeViaAI(goCode string) (string, error) {
	prompt := fmt.Sprintf(
		"Bạn là chuyên gia Golang và Elasticsearch. Hãy chuyển đổi đoạn mã Go sau thành câu truy vấn Elasticsearch Query DSL định dạng JSON hoàn chỉnh:\n\n"+
			"```go\n%s\n```\n\n"+
			"QUY TẮC: Chỉ trả về DUY NHẤT một khối JSON hợp lệ nằm giữa cặp thẻ ```json và ```.",
		goCode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	res, err := c.aiBridge.ExplainQuery(ctx, prompt, "")
	if err != nil {
		return "", err
	}

	jsonStr := extractCodeBlock(res, "json")
	if jsonStr == "" {
		jsonStr = res
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return "", fmt.Errorf("AI output is not valid JSON: %w", err)
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err == nil {
		return string(pretty), nil
	}
	return jsonStr, nil
}

func appendPagingAndSort(sb *strings.Builder, root map[string]interface{}) {
	// Paging
	sizeVal, hasSize := root["size"].(float64)
	fromVal, hasFrom := root["from"].(float64)
	if hasSize || hasFrom {
		size := int(sizeVal)
		from := int(fromVal)
		if size <= 0 {
			size = 10
		}
		page := 0
		if size > 0 {
			page = from / size
		}
		sb.WriteString(fmt.Sprintf(".Paging(%d, %d)", page, size))
	}

	// Sort
	if sortRaw, ok := root["sort"]; ok {
		var sortItems []string
		switch s := sortRaw.(type) {
		case []interface{}:
			for _, item := range s {
				if sm, ok := item.(map[string]interface{}); ok {
					for k, v := range sm {
						order := "DESC"
						if vm, ok := v.(map[string]interface{}); ok {
							if ord, ok := vm["order"].(string); ok && strings.ToLower(ord) == "asc" {
								order = "ASC"
							}
						} else if ordStr, ok := v.(string); ok && strings.ToLower(ordStr) == "asc" {
							order = "ASC"
						}
						sortItems = append(sortItems, fmt.Sprintf("\tAdd(EsSort{Key: %q, Order: %s})", k, order))
					}
				}
			}
		}

		if len(sortItems) > 0 {
			sb.WriteString(fmt.Sprintf("\n\nsorts := EsSortArr{}.\n%s\nquery = query.Sort(sorts)", strings.Join(sortItems, ".\n")))
		}
	}
}

func parsePagingAndSortFromCode(code string, result map[string]interface{}) {
	// Parse .Paging(page, size)
	pagingRegex := regexp.MustCompile(`\.Paging\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)`)
	if match := pagingRegex.FindStringSubmatch(code); len(match) == 3 {
		page, _ := strconv.Atoi(match[1])
		size, _ := strconv.Atoi(match[2])
		result["size"] = size
		result["from"] = page * size
	}

	// Parse EsSort{Key: "...", Order: ...}
	sortRegex := regexp.MustCompile(`EsSort\s*\{\s*Key:\s*"([^"]+)",\s*Order:\s*([A-Za-z0-9_"]+)\}`)
	sortMatches := sortRegex.FindAllStringSubmatch(code, -1)
	if len(sortMatches) > 0 {
		var sorts []map[string]interface{}
		for _, sm := range sortMatches {
			key := sm[1]
			ordRaw := strings.ToUpper(strings.Trim(sm[2], `"`))
			order := "desc"
			if ordRaw == "ASC" || strings.ToLower(ordRaw) == "asc" {
				order = "asc"
			}
			sorts = append(sorts, map[string]interface{}{
				key: map[string]interface{}{
					"order": order,
				},
			})
		}
		result["sort"] = sorts
	}
}

func formatGoValue(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case int, int64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	default:
		b, err := json.Marshal(val)
		if err == nil {
			return fmt.Sprintf("%q", string(b))
		}
		return fmt.Sprintf("%v", val)
	}
}

func parseGoLiteral(raw string) interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "nil" {
		return nil
	}
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		var s string
		if err := json.Unmarshal([]byte(trimmed), &s); err == nil {
			return s
		}
		return strings.Trim(trimmed, `"`)
	}
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(trimmed); err == nil {
		return b
	}
	return trimmed
}

func extractCodeBlock(content, lang string) string {
	pattern := fmt.Sprintf("(?s)```%s\\s*\\n(.*?)```", lang)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	reAny := regexp.MustCompile("(?s)```\\s*\\n(.*?)```")
	matchesAny := reAny.FindStringSubmatch(content)
	if len(matchesAny) > 1 {
		return strings.TrimSpace(matchesAny[1])
	}

	return ""
}
