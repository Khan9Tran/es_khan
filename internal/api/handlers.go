package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"eskhan/internal/ai"
	"eskhan/internal/config"
	"eskhan/internal/es"
	"eskhan/internal/linter"
	"eskhan/internal/schema"
	"eskhan/internal/storage"
)

// Connections handlers

func (s *Server) handleGetConnections(w http.ResponseWriter, r *http.Request) {
	cfg := s.configMgr.GetConfig()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active_id":   cfg.ActiveConnectionID,
		"connections": cfg.Connections,
	})
}

func (s *Server) handleSaveConnection(w http.ResponseWriter, r *http.Request) {
	var conn config.ConnectionProfile
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		writeError(w, http.StatusBadRequest, "invalid connection payload: "+err.Error())
		return
	}

	if conn.URL == "" {
		writeError(w, http.StatusBadRequest, "connection URL is required")
		return
	}
	if conn.Name == "" {
		conn.Name = conn.URL
	}

	if err := s.configMgr.SaveConnection(conn); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save connection: "+err.Error())
		return
	}

	// If this connection is now active, refresh client
	active, _ := s.configMgr.GetActiveConnection()
	if active != nil && active.ID == conn.ID {
		s.mu.Lock()
		s.esClient = es.NewClient(*active)
		s.schemaMgr.UpdateClient(s.esClient)
		s.mu.Unlock()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "saved",
		"connection": conn,
	})
}

func (s *Server) handleDeleteConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	if err := s.configMgr.DeleteConnection(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete connection: "+err.Error())
		return
	}

	// Update client with new active connection if changed
	active, _ := s.configMgr.GetActiveConnection()
	s.mu.Lock()
	if active != nil {
		s.esClient = es.NewClient(*active)
	} else {
		s.esClient = nil
	}
	s.schemaMgr.UpdateClient(s.esClient)
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleActivateConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	if err := s.configMgr.SetActiveConnection(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	active, err := s.configMgr.GetActiveConnection()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.mu.Lock()
	s.esClient = es.NewClient(*active)
	s.schemaMgr.UpdateClient(s.esClient)
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "activated",
		"active_id": active.ID,
	})
}

func (s *Server) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	var conn config.ConnectionProfile
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		writeError(w, http.StatusBadRequest, "invalid connection payload: "+err.Error())
		return
	}

	if conn.URL == "" {
		writeError(w, http.StatusBadRequest, "connection URL is required")
		return
	}

	testClient := es.NewClient(conn)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := testClient.Ping(ctx)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	health, _ := testClient.GetHealth(ctx)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"info":    info,
		"health":  health,
	})
}

// Cluster & Indices Handlers

func (s *Server) handleGetClusterInfo(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := client.Ping(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, "cluster ping failed: "+err.Error())
		return
	}

	health, _ := client.GetHealth(ctx)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"info":   info,
		"health": health,
	})
}

func (s *Server) handleGetIndices(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	indices, err := client.GetIndices(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to get indices: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"indices": indices,
	})
}

func (s *Server) handleGetIndexMapping(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "index name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	fields, err := s.schemaMgr.GetIndexFields(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get mapping: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"index":  name,
		"fields": fields,
	})
}

// Query Execution Handler

func (s *Server) handleExecuteQuery(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req es.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid query payload: "+err.Error())
		return
	}

	activeConn, _ := s.configMgr.GetActiveConnection()
	connID := ""
	if activeConn != nil {
		connID = activeConn.ID
	}

	result, err := client.ExecuteQuery(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query execution error: "+err.Error())
		return
	}

	// Record in history asynchronously or sync
	method, path, _ := es.ParseRawInput(req.RawInput, req.Index)
	_ = s.storage.AddHistory(storage.HistoryItem{
		ConnectionID: connID,
		Index:        req.Index,
		Method:       method,
		Path:         path,
		RawInput:     req.RawInput,
		Status:       result.StatusCode,
		TookMs:       result.TookMs,
		TotalHits:    result.TotalHits,
		Timestamp:    time.Now(),
	})

	writeJSON(w, http.StatusOK, result)
}

// Autocomplete DSL Snippets Handler

func (s *Server) handleGetDSLSnippets(w http.ResponseWriter, r *http.Request) {
	snippets := schema.GetStandardSnippets()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"snippets": snippets,
	})
}

// History Handlers

func (s *Server) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items := s.storage.GetHistory(limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"history": items,
	})
}

func (s *Server) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if err := s.storage.ClearHistory(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

// Snippets Handlers

func (s *Server) handleGetSnippets(w http.ResponseWriter, r *http.Request) {
	snippets := s.storage.GetSnippets()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"snippets": snippets,
	})
}

func (s *Server) handleSaveSnippet(w http.ResponseWriter, r *http.Request) {
	var snip storage.SnippetItem
	if err := json.NewDecoder(r.Body).Decode(&snip); err != nil {
		writeError(w, http.StatusBadRequest, "invalid snippet: "+err.Error())
		return
	}

	if snip.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if err := s.storage.SaveSnippet(snip); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "saved",
		"snippet": snip,
	})
}

func (s *Server) handleDeleteSnippet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := s.storage.DeleteSnippet(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Analyze Handler

func (s *Server) handleAnalyzeText(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req es.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid analyze request: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.AnalyzeText(ctx, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Top Terms Handler

func (s *Server) handleGetTopTerms(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	name := chi.URLParam(r, "name")
	field := r.URL.Query().Get("field")
	sizeStr := r.URL.Query().Get("size")
	size := 10
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	if name == "" || field == "" {
		writeError(w, http.StatusBadRequest, "index name and field query parameter are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	terms, err := client.GetTopTerms(ctx, name, field, size)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"index": name,
		"field": field,
		"terms": terms,
	})
}

// Index Action Handler

func (s *Server) handleIndexAction(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	name := chi.URLParam(r, "name")
	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Action == "" {
		writeError(w, http.StatusBadRequest, "valid action is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result, err := client.ExecuteIndexAction(ctx, name, body.Action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"result": result,
	})
}

// Document CRUD Handlers

func (s *Server) handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req es.DocUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid update document payload: "+err.Error())
		return
	}

	if req.Index == "" || req.ID == "" {
		writeError(w, http.StatusBadRequest, "index and id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result, err := client.UpdateDocument(ctx, req.Index, req.ID, req.Doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	client, err := s.getActiveClient()
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	index := chi.URLParam(r, "index")
	id := chi.URLParam(r, "id")
	if index == "" || id == "" {
		writeError(w, http.StatusBadRequest, "index and id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result, err := client.DeleteDocument(ctx, index, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Query Linter Handler

func (s *Server) handleLinterCheck(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query string `json:"query"`
		Index string `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid linter payload: "+err.Error())
		return
	}

	fieldTypes := make(map[string]string)
	if body.Index != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if fields, err := s.schemaMgr.GetIndexFields(ctx, body.Index); err == nil {
			for _, f := range fields {
				fieldTypes[f.Name] = f.Type
			}
		}
	}

	warnings := linter.AnalyzeQuery(body.Query, fieldTypes)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"warnings": warnings,
	})
}

// Antigravity Handlers (Local CLI Bridge)

func (s *Server) handleAntigravityStatus(w http.ResponseWriter, r *http.Request) {
	status := s.aiBridge.Status()
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleAntigravityGenerate(w http.ResponseWriter, r *http.Request) {
	var req ai.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	// Resolve index name from IndexName, Index, or ExistingQuery
	if req.IndexName == "" && req.Index != "" {
		req.IndexName = req.Index
	}
	if req.IndexName == "" && req.ExistingQuery != "" {
		req.IndexName = ai.ExtractIndexFromQuery(req.ExistingQuery)
	}

	// If no index is resolved: ask user to choose an index, offering cluster indices as clickable options!
	if req.IndexName == "" {
		var options []ai.ClarificationOption
		if client, err := s.getActiveClient(); err == nil && client != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if indices, err := client.GetIndices(ctx); err == nil {
				for _, idx := range indices {
					if !strings.HasPrefix(idx.Name, ".") {
						options = append(options, ai.ClarificationOption{
							Label: fmt.Sprintf("Chỉ mục '%s' (%d docs)", idx.Name, idx.DocsCount),
							Field: idx.Name,
						})
					}
				}
			}
		}
		writeJSON(w, http.StatusOK, ai.GenerateResponse{
			NeedsClarification: true,
			Question:           "Bạn chưa chọn Chỉ mục (Index) mục tiêu. Vui lòng chọn một index dưới đây để Antigravity nạp đúng Schema Mapping trước khi sinh câu truy vấn:",
			SuggestedOptions:   options,
		})
		return
	}

	// Auto-enrich mapping fields if indexName is provided and fields are empty
	if req.IndexName != "" && len(req.Fields) == 0 {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if fields, err := s.schemaMgr.GetIndexFields(ctx, req.IndexName); err == nil && len(fields) > 0 {
			req.Fields = fields
		}
	}

	// Get ES version if connected
	if client, err := s.getActiveClient(); err == nil && client != nil {
		if info, err := client.Ping(r.Context()); err == nil && info.Version.Number != "" {
			req.ESVersion = info.Version.Number
		}
	}

	res, err := s.aiBridge.GenerateQuery(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleAntigravityExplain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query     string `json:"query"`
		IndexName string `json:"index_name"`
		Index     string `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if strings.TrimSpace(body.Query) == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	if body.IndexName == "" && body.Index != "" {
		body.IndexName = body.Index
	}
	if body.IndexName == "" && body.Query != "" {
		body.IndexName = ai.ExtractIndexFromQuery(body.Query)
	}

	explanation, err := s.aiBridge.ExplainQuery(r.Context(), body.Query, body.IndexName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"explanation": explanation,
	})
}

func (s *Server) handleAntigravityFix(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query     string `json:"query"`
		ESError   string `json:"es_error"`
		Error     string `json:"error"`
		IndexName string `json:"index_name"`
		Index     string `json:"index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if body.IndexName == "" && body.Index != "" {
		body.IndexName = body.Index
	}
	if body.IndexName == "" && body.Query != "" {
		body.IndexName = ai.ExtractIndexFromQuery(body.Query)
	}
	if body.ESError == "" && body.Error != "" {
		body.ESError = body.Error
	}

	var fields []schema.FieldSuggestion
	if body.IndexName != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if f, err := s.schemaMgr.GetIndexFields(ctx, body.IndexName); err == nil {
			fields = f
		}
	}

	fixedQuery, err := s.aiBridge.FixQuery(r.Context(), body.Query, body.ESError, body.IndexName, fields)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"fixed_query": fixedQuery,
	})
}

// --- Go DSL Converter Handlers (Bidirectional: Query DSL ⇋ Go Code) ---

func (s *Server) handleConvertToGolang(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QueryDSL string `json:"query_dsl"`
		Query    string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	raw := body.QueryDSL
	if raw == "" {
		raw = body.Query
	}

	goCode, err := s.converter.QueryToGolang(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"golang_code": goCode,
	})
}

func (s *Server) handleConvertToDSL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GolangCode string `json:"golang_code"`
		Code       string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	raw := body.GolangCode
	if raw == "" {
		raw = body.Code
	}

	if strings.TrimSpace(raw) == "" {
		writeError(w, http.StatusBadRequest, "mã nguồn Golang không được để trống")
		return
	}

	queryDSL, err := s.converter.GolangToQuery(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"query_dsl": queryDSL,
	})
}


