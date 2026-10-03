package api

import (
	"encoding/json"
	"net/http"

	"eskhan/internal/httpclient"
	"eskhan/internal/storage"
)

func (s *Server) handleHTTPSend(w http.ResponseWriter, r *http.Request) {
	var req httpclient.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Inject active environment variables if none provided
	if len(req.Variables) == 0 && s.configMgr != nil {
		activeEnv, _ := s.configMgr.GetActiveEnvironment()
		if activeEnv != nil && len(activeEnv.Variables) > 0 {
			req.Variables = activeEnv.Variables
		}
	}

	resp, err := httpclient.Execute(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "HTTP execution failed: "+err.Error())
		return
	}

	// Persist request in history
	if s.storage != nil && resp != nil {
		_ = s.storage.AddHistory(storage.HistoryItem{
			Protocol:   "http",
			Method:     req.Method,
			Path:       req.URL,
			RawInput:   req.Body,
			Status:     resp.StatusCode,
			StatusText: resp.StatusText,
			TookMs:     resp.Timings.TotalMs,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
