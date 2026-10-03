package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"eskhan/internal/openapi"
)

func (s *Server) handleOpenAPIParse(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL     string `json:"url"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	payload.URL = strings.TrimSpace(payload.URL)
	payload.Content = strings.TrimSpace(payload.Content)

	if payload.URL == "" && payload.Content == "" {
		writeError(w, http.StatusBadRequest, "Either 'url' or 'content' must be provided")
		return
	}

	var spec *openapi.ParsedSpec
	var err error

	if payload.URL != "" {
		spec, err = openapi.FetchAndParse(r.Context(), payload.URL)
	} else {
		spec, err = openapi.ParseSpec([]byte(payload.Content))
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse OpenAPI/Swagger spec: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, spec)
}
