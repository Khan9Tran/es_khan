package api

import (
	"encoding/json"
	"net/http"

	"eskhan/internal/grpcclient"
)

func (s *Server) handleGRPCReflect(w http.ResponseWriter, r *http.Request) {
	var cfg grpcclient.TargetConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if cfg.Target == "" {
		writeError(w, http.StatusBadRequest, "Target address is required (e.g. localhost:50051)")
		return
	}

	services, err := grpcclient.DiscoverServices(r.Context(), cfg)
	if err != nil {
		writeError(w, http.StatusBadGateway, "gRPC reflection failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"services": services,
	})
}

func (s *Server) handleGRPCInvoke(w http.ResponseWriter, r *http.Request) {
	var req grpcclient.InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Target == "" || req.Service == "" || req.Method == "" {
		writeError(w, http.StatusBadRequest, "target, service, and method are required")
		return
	}

	resp, err := grpcclient.Invoke(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invocation error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGRPCParseProto(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	services, err := grpcclient.ParseProtoContent(payload.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse proto content: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"services": services,
	})
}
