package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"eskhan/internal/schema"
)

// GenerateRequest defines input parameters for generating Query DSL.
type GenerateRequest struct {
	Prompt        string                   `json:"prompt"`
	IndexName     string                   `json:"index_name"`
	Index         string                   `json:"index,omitempty"`
	Fields        []schema.FieldSuggestion `json:"fields,omitempty"`
	ESVersion     string                   `json:"es_version,omitempty"`
	ExistingQuery string                   `json:"existing_query,omitempty"`
}

// ExtractIndexFromQuery extracts the index name from the first line of a Kibana DevTools style query (e.g., POST /my_index/_search).
func ExtractIndexFromQuery(raw string) string {
	trimmed := strings.TrimSpace(raw)
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return ""
	}
	firstLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(firstLine)
	if len(parts) >= 2 {
		method := strings.ToUpper(parts[0])
		if method == "GET" || method == "POST" || method == "PUT" || method == "DELETE" {
			path := strings.TrimPrefix(parts[1], "/")
			pathParts := strings.Split(path, "/")
			if len(pathParts) > 0 && pathParts[0] != "" && !strings.HasPrefix(pathParts[0], "_") {
				return pathParts[0]
			}
		}
	}
	return ""
}

// GenerateResponse holds the generated query, explanation, or interactive clarification questions.
type GenerateResponse struct {
	QueryDSL           string                `json:"query_dsl,omitempty"`
	Explanation        string                `json:"explanation,omitempty"`
	Method             string                `json:"method,omitempty"`
	Path               string                `json:"path,omitempty"`
	NeedsClarification bool                  `json:"needs_clarification,omitempty"`
	Question           string                `json:"question,omitempty"`
	SuggestedOptions   []ClarificationOption `json:"suggested_options,omitempty"`
	InvalidFields      []string              `json:"invalid_fields,omitempty"`
	VerifiedFields     []string              `json:"verified_fields,omitempty"`
}

// StatusInfo represents the readiness of Antigravity CLI local bridge.
type StatusInfo struct {
	Available bool   `json:"available"`
	Path      string `json:"path"`
	Version   string `json:"version,omitempty"`
	Account   string `json:"account,omitempty"`
	Message   string `json:"message"`
}

// Bridge handles communication with the local Antigravity CLI (`agy`).
type Bridge struct {
	agyPath string
	timeout time.Duration
}

// NewBridge initializes an Antigravity CLI bridge, discovering the local binary.
func NewBridge(customPath string) *Bridge {
	targetPath := customPath
	if targetPath == "" {
		targetPath = FindAgyBinary()
	}

	return &Bridge{
		agyPath: targetPath,
		timeout: 60 * time.Second,
	}
}

// FindAgyBinary searches for the local `agy` executable in standard paths.
func FindAgyBinary() string {
	// 1. Check user home .local/bin
	home, err := os.UserHomeDir()
	if err == nil {
		localPath := filepath.Join(home, ".local", "bin", "agy")
		if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
			return localPath
		}
	}

	// 2. Check current PATH
	if p, err := exec.LookPath("agy"); err == nil {
		return p
	}

	// 3. Common macOS / Linux paths
	fallbacks := []string{
		"/usr/local/bin/agy",
		"/opt/homebrew/bin/agy",
	}
	for _, fb := range fallbacks {
		if info, err := os.Stat(fb); err == nil && !info.IsDir() {
			return fb
		}
	}

	return ""
}

// Status checks whether Antigravity CLI is installed and ready to use.
func (b *Bridge) Status() StatusInfo {
	path := b.agyPath
	if path == "" {
		path = FindAgyBinary()
	}

	if path == "" {
		return StatusInfo{
			Available: false,
			Path:      "",
			Message:   "Antigravity CLI (agy) chưa được cài đặt trên máy. Vui lòng cài đặt agy để sử dụng.",
		}
	}

	// Verify file is executable
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return StatusInfo{
			Available: false,
			Path:      path,
			Message:   fmt.Sprintf("Không tìm thấy tệp nhị phân agy tại %s", path),
		}
	}

	// Quick execution check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return StatusInfo{
			Available: false,
			Path:      path,
			Message:   fmt.Sprintf("Lỗi khởi chạy agy: %v", err),
		}
	}

	return StatusInfo{
		Available: true,
		Path:      path,
		Account:   "Đã kết nối tài khoản Google Antigravity",
		Message:   "Antigravity CLI Local Bridge đã sẵn sàng hỗ trợ truy vấn!",
	}
}

// GenerateQuery invokes `agy` to generate a valid Elasticsearch Query DSL JSON.
func (b *Bridge) GenerateQuery(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if b.agyPath == "" {
		b.agyPath = FindAgyBinary()
	}
	if b.agyPath == "" {
		return nil, fmt.Errorf("antigravity CLI (agy) is not installed or not found in PATH")
	}

	if req.IndexName == "" && req.Index != "" {
		req.IndexName = req.Index
	}
	if req.IndexName == "" && req.ExistingQuery != "" {
		req.IndexName = ExtractIndexFromQuery(req.ExistingQuery)
	}

	var sb strings.Builder
	sb.WriteString("Bạn là kiến trúc sư chuyên gia cao cấp về Elasticsearch Query DSL.\n")
	sb.WriteString("Nhiệm vụ: Tạo một Elasticsearch Query DSL định dạng JSON hoàn chỉnh, tối ưu và TUÂN THỦ NGHIÊM NGẶT SCHEMA cho yêu cầu sau.\n\n")

	if req.IndexName != "" {
		sb.WriteString(fmt.Sprintf("Chỉ mục mục tiêu (Target Index): %s\n", req.IndexName))
	}
	if req.ESVersion != "" {
		sb.WriteString(fmt.Sprintf("Phiên bản Elasticsearch: %s\n", req.ESVersion))
	}

	if len(req.Fields) > 0 {
		sb.WriteString("=== DANH SÁCH TOÀN BỘ TRƯỜNG TRONG SCHEMA MAPPING CỦA INDEX NÀY ===\n")
		count := 0
		for _, f := range req.Fields {
			if strings.HasPrefix(f.Name, "_") {
				continue
			}
			if count >= 250 {
				sb.WriteString(fmt.Sprintf("- ... (và %d trường khác trong schema)\n", len(req.Fields)-count))
				break
			}
			detail := ""
			if f.Detail != "" {
				detail = fmt.Sprintf(" [%s]", f.Detail)
			}
			sb.WriteString(fmt.Sprintf("- %s (kiểu: %s%s)\n", f.Name, f.Type, detail))
			count++
		}
		sb.WriteString("===================================================================\n\n")
	} else if req.IndexName != "" {
		sb.WriteString(fmt.Sprintf("LƯU Ý: Không lấy được schema mapping cho index '%s'.\n\n", req.IndexName))
	}

	if req.ExistingQuery != "" {
		sb.WriteString(fmt.Sprintf("Truy vấn hiện tại trên editor (để tham khảo hoặc chỉnh sửa):\n%s\n\n", req.ExistingQuery))
	}

	sb.WriteString(fmt.Sprintf("YÊU CẦU CỦA NGƯỜI DÙNG: \"%s\"\n\n", req.Prompt))
	sb.WriteString("QUY TẮC BẮT BUỘC ĐỂ QUERY CHÍNH XÁC 100% THEO SCHEMA:\n")
	sb.WriteString("1. TUÂN THỦ TUYỆT ĐỐI SCHEMA MAPPING:\n")
	sb.WriteString("   - CHỈ ĐƯỢC PHÉP dùng các tên trường (field names) có trong danh sách mapping ở trên.\n")
	sb.WriteString("   - TUYỆT ĐỐI KHÔNG TỰ BỊA RA TÊN TRƯỜNG KHÔNG CÓ TRONG SCHEMA!\n")
	sb.WriteString("2. NGUYÊN TẮC HỎI LÀM RÕ (CLARIFICATION) NẾU KHÔNG CÓ TRƯỜNG HOẶC MƠ HỒ:\n")
	sb.WriteString("   - Nếu người dùng yêu cầu lọc/tìm kiếm theo một khái niệm mà TRONG SCHEMA KHÔNG CÓ TRƯỜNG NÀO TƯƠNG ỨNG, hoặc CÓ NHIỀU TRƯỜNG MƠ HỒ (ví dụ: người dùng nói 'lọc theo giá' nhưng schema có 'unit_price' và 'total_amount', hoặc người dùng nói 'trạng thái' nhưng không có 'status' mà chỉ có 'order_status'):\n")
	sb.WriteString("     -> BẠN KHÔNG ĐƯỢC TỰ BỊA TRƯỜNG VÀ KHÔNG ĐƯỢC ĐOÁN BỪA!\n")
	sb.WriteString("     -> BẠN PHẢI TRẢ VỀ ĐỊNH DẠNG HỎI LÀM RÕ (CLARIFICATION JSON) sau:\n")
	sb.WriteString("     ```json\n")
	sb.WriteString("     {\n")
	sb.WriteString("       \"needs_clarification\": true,\n")
	sb.WriteString("       \"question\": \"Trong schema của index không có trường [tên_trường_sai]. Bạn có muốn dùng một trong các trường gợi ý sau không?\",\n")
	sb.WriteString("       \"suggested_options\": [\n")
	sb.WriteString("         {\"label\": \"Dùng 'tên_trường_1' (kiểu)\", \"field\": \"tên_trường_1\"},\n")
	sb.WriteString("         {\"label\": \"Dùng 'tên_trường_2' (kiểu)\", \"field\": \"tên_trường_2\"}\n")
	sb.WriteString("       ]\n")
	sb.WriteString("     }\n")
	sb.WriteString("     ```\n")
	sb.WriteString("3. NẾU YÊU CẦU RÕ RÀNG VÀ TẤT CẢ CÁC TRƯỜNG ĐỀU TỒN TẠI TRONG SCHEMA:\n")
	sb.WriteString("   - Trả về JSON Query DSL hợp lệ nằm giữa cặp thẻ ```json và ```.\n")
	sb.WriteString("   - Với trường 'text' có subfield '.keyword': bắt buộc dùng '.keyword' cho 'term', 'terms', 'sort', 'aggs'.\n")
	sb.WriteString("   - Với trường 'text' thuần: dùng 'match', 'multi_match'. Không dùng 'term' trên text.\n")
	sb.WriteString("   - Với trường số hoặc 'date': dùng 'range' hoặc 'term'.\n")
	sb.WriteString("   - Đưa tất cả điều kiện lọc chính xác vào 'filter' của bool query.\n")
	sb.WriteString("   - KHÔNG viết bất kỳ lời giải thích nào bên ngoài khối json.\n\n")

	prompt := sb.String()

	callCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cmd := exec.CommandContext(callCtx, b.agyPath, "-p", prompt, "--disable-slash-commands", "--dangerously-skip-permissions", "--effort", "low")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if errStr == "" {
			errStr = stdout.String()
		}
		return nil, fmt.Errorf("antigravity CLI execution error: %v (details: %s)", err, strings.TrimSpace(errStr))
	}

	rawOutput := stdout.String()
	jsonDSL := extractJSON(rawOutput)
	if jsonDSL == "" {
		return nil, fmt.Errorf("antigravity did not return a valid JSON block: %s", rawOutput)
	}

	// 1. Check if Antigravity itself returned a Clarification JSON
	var clarResp struct {
		NeedsClarification bool                  `json:"needs_clarification"`
		Question           string                `json:"question"`
		SuggestedOptions   []ClarificationOption `json:"suggested_options"`
		SuggestedFields    []string              `json:"suggested_fields"`
	}
	if err := json.Unmarshal([]byte(jsonDSL), &clarResp); err == nil && (clarResp.NeedsClarification || clarResp.Question != "") {
		options := clarResp.SuggestedOptions
		if len(options) == 0 && len(clarResp.SuggestedFields) > 0 {
			for _, sf := range clarResp.SuggestedFields {
				options = append(options, ClarificationOption{
					Label: fmt.Sprintf("Dùng '%s'", sf),
					Field: sf,
				})
			}
		}
		return &GenerateResponse{
			NeedsClarification: true,
			Question:           clarResp.Question,
			SuggestedOptions:   options,
		}, nil
	}

	// 2. Validate JSON syntax
	var parsed interface{}
	if err := json.Unmarshal([]byte(jsonDSL), &parsed); err != nil {
		return nil, fmt.Errorf("generated output is not valid JSON: %w", err)
	}

	// 3. Schema Verification Layer: Inspect every field in generated Query DSL against index schema
	if len(req.Fields) > 0 {
		valRes := ValidateQueryAgainstSchema(jsonDSL, req.IndexName, req.Fields)
		if !valRes.IsValid {
			// Query contained fields not in schema! Intercept and return interactive clarification instead of bad query!
			return &GenerateResponse{
				NeedsClarification: true,
				Question:           valRes.Question,
				SuggestedOptions:   valRes.SuggestedOptions,
				InvalidFields:      valRes.InvalidFields,
				QueryDSL:           jsonDSL,
			}, nil
		}
	}

	// Format pretty
	prettyBytes, err := json.MarshalIndent(parsed, "", "  ")
	if err == nil {
		jsonDSL = string(prettyBytes)
	}

	path := "/_search"
	if req.IndexName != "" {
		path = fmt.Sprintf("/%s/_search", req.IndexName)
	}

	// Extract verified fields for metadata
	var verified []string
	if len(req.Fields) > 0 {
		vRes := ValidateQueryAgainstSchema(jsonDSL, req.IndexName, req.Fields)
		verified = vRes.VerifiedFields
	}

	return &GenerateResponse{
		QueryDSL:       jsonDSL,
		Method:         "POST",
		Path:           path,
		VerifiedFields: verified,
	}, nil
}

// ExplainQuery asks Antigravity to explain the structure and logic of a query in Vietnamese.
func (b *Bridge) ExplainQuery(ctx context.Context, query string, indexName string) (string, error) {
	if b.agyPath == "" {
		b.agyPath = FindAgyBinary()
	}
	if b.agyPath == "" {
		return "", fmt.Errorf("antigravity CLI (agy) is not installed")
	}

	prompt := fmt.Sprintf(
		"Bạn là chuyên gia Elasticsearch. Hãy giải thích chi tiết, ngắn gọn, súc tích bằng Tiếng Việt xem truy vấn Elasticsearch sau thực hiện điều gì:\n\n"+
			"Chỉ mục: %s\n"+
			"Nội dung truy vấn Query DSL:\n```json\n%s\n```\n\n"+
			"Hãy giải thích rõ: mục đích tìm kiếm, các điều kiện lọc (filter), các mệnh đề tính điểm (must/should), sắp xếp (sort) hoặc gom nhóm (aggregations) nếu có.",
		indexName, query,
	)

	callCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cmd := exec.CommandContext(callCtx, b.agyPath, "-p", prompt, "--disable-slash-commands", "--dangerously-skip-permissions", "--effort", "low")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("antigravity explain failed: %v (%s)", err, strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

// FixQuery asks Antigravity to fix a broken query based on the error message.
func (b *Bridge) FixQuery(ctx context.Context, query string, esError string, indexName string, fields []schema.FieldSuggestion) (string, error) {
	if b.agyPath == "" {
		b.agyPath = FindAgyBinary()
	}
	if b.agyPath == "" {
		return "", fmt.Errorf("antigravity CLI (agy) is not installed")
	}

	var sb strings.Builder
	sb.WriteString("Bạn là chuyên gia sửa lỗi Elasticsearch. Truy vấn sau đây bị lỗi khi gửi tới Elasticsearch:\n\n")
	if indexName != "" {
		sb.WriteString(fmt.Sprintf("Chỉ mục: %s\n", indexName))
	}
	sb.WriteString(fmt.Sprintf("Truy vấn bị lỗi:\n```json\n%s\n```\n\n", query))
	sb.WriteString(fmt.Sprintf("Thông báo lỗi từ Elasticsearch:\n%s\n\n", esError))

	if len(fields) > 0 {
		sb.WriteString("Các trường có sẵn trong mapping:\n")
		count := 0
		for _, f := range fields {
			if count >= 40 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", f.Name, f.Type))
			count++
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Hãy sửa lại truy vấn trên để hết lỗi và thực thi thành công. Trả về DUY NHẤT một khối JSON hợp lệ nằm giữa cặp thẻ ```json và ```.")

	callCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cmd := exec.CommandContext(callCtx, b.agyPath, "-p", sb.String(), "--disable-slash-commands", "--dangerously-skip-permissions", "--effort", "low")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("antigravity fix failed: %v (%s)", err, strings.TrimSpace(string(out)))
	}

	fixedJSON := extractJSON(string(out))
	if fixedJSON == "" {
		return "", fmt.Errorf("antigravity did not return fixed JSON: %s", string(out))
	}

	// Validate JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(fixedJSON), &parsed); err != nil {
		return fixedJSON, nil
	}

	prettyBytes, err := json.MarshalIndent(parsed, "", "  ")
	if err == nil {
		return string(prettyBytes), nil
	}

	return fixedJSON, nil
}

// extractJSON extracts the JSON content between ```json and ``` or between outermost braces.
func extractJSON(text string) string {
	// 1. Look for ```json ... ```
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\}|\\[.*?\\])\\s*```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 2. Look for outermost { ... }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		candidate := strings.TrimSpace(text[start : end+1])
		var js json.RawMessage
		if json.Unmarshal([]byte(candidate), &js) == nil {
			return candidate
		}
	}

	return ""
}
