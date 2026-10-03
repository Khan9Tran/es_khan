package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"eskhan/internal/config"
)

func (s *Server) handleGetEnvironments(w http.ResponseWriter, r *http.Request) {
	envs := s.configMgr.GetEnvironments()
	activeEnv, _ := s.configMgr.GetActiveEnvironment()
	activeID := ""
	if activeEnv != nil {
		activeID = activeEnv.ID
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"environments":          envs,
		"active_environment_id": activeID,
	})
}

func (s *Server) handleSaveEnvironment(w http.ResponseWriter, r *http.Request) {
	var env config.Environment
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if env.Name == "" {
		writeError(w, http.StatusBadRequest, "Environment name is required")
		return
	}

	if err := s.configMgr.SaveEnvironment(env); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save environment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "saved",
		"environment": env,
	})
}

func (s *Server) handleDeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Environment ID is required")
		return
	}

	if err := s.configMgr.DeleteEnvironment(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete environment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

func (s *Server) handleActivateEnvironment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// If id is "none" or "_none_", deactivate active environment
	if id == "none" || id == "_none_" {
		id = ""
	}

	if err := s.configMgr.SetActiveEnvironment(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":                "activated",
		"active_environment_id": id,
	})
}
