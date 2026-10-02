package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"eskhan/internal/ai"
	"eskhan/internal/config"
	"eskhan/internal/converter"
	"eskhan/internal/es"
	"eskhan/internal/schema"
	"eskhan/internal/storage"
)

// Server coordinates the API router, ES client, schema manager, storage, AI bridge, and converter.
type Server struct {
	configMgr *config.Manager
	storage   *storage.Store
	schemaMgr *schema.Manager
	esClient  *es.Client
	aiBridge  *ai.Bridge
	converter *converter.Converter
	mu        sync.RWMutex
	router    chi.Router
	staticFS  fs.FS
}

// NewServer creates a new API server instance.
func NewServer(cfgMgr *config.Manager, store *storage.Store, staticFS fs.FS) (*Server, error) {
	aiBr := ai.NewBridge("")
	s := &Server{
		configMgr: cfgMgr,
		storage:   store,
		staticFS:  staticFS,
		aiBridge:  aiBr,
		converter: converter.NewConverter(aiBr),
	}

	// Initialize active ES client if profile exists
	activeConn, err := cfgMgr.GetActiveConnection()
	if err == nil && activeConn != nil {
		s.esClient = es.NewClient(*activeConn)
	}
	s.schemaMgr = schema.NewManager(s.esClient)

	s.setupRoutes()
	return s, nil
}

// Router returns the Chi router instance.
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS handling for local dev if needed
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Connections
		r.Get("/connections", s.handleGetConnections)
		r.Post("/connections", s.handleSaveConnection)
		r.Delete("/connections/{id}", s.handleDeleteConnection)
		r.Post("/connections/{id}/activate", s.handleActivateConnection)
		r.Post("/connections/test", s.handleTestConnection)

		// Cluster & Indices
		r.Get("/cluster/info", s.handleGetClusterInfo)
		r.Get("/indices", s.handleGetIndices)
		r.Get("/indices/{name}/mapping", s.handleGetIndexMapping)

		// Query execution
		r.Post("/query/execute", s.handleExecuteQuery)
		r.Post("/linter/check", s.handleLinterCheck)

		// Antigravity (Local CLI Bridge)
		r.Get("/antigravity/status", s.handleAntigravityStatus)
		r.Post("/antigravity/generate", s.handleAntigravityGenerate)
		r.Post("/antigravity/explain", s.handleAntigravityExplain)
		r.Post("/antigravity/fix", s.handleAntigravityFix)

		// Go DSL Converter (Bidirectional: Query DSL ⇋ Go Code)
		r.Post("/converter/to-golang", s.handleConvertToGolang)
		r.Post("/converter/to-dsl", s.handleConvertToDSL)

		// Schema, Terms & Snippets for autocomplete
		r.Get("/snippets/dsl", s.handleGetDSLSnippets)
		r.Get("/indices/{name}/terms", s.handleGetTopTerms)
		r.Post("/indices/{name}/action", s.handleIndexAction)

		// Text analysis & Playground
		r.Post("/analyze", s.handleAnalyzeText)

		// Document CRUD
		r.Post("/docs/update", s.handleUpdateDocument)
		r.Delete("/docs/{index}/{id}", s.handleDeleteDocument)

		// History
		r.Get("/history", s.handleGetHistory)
		r.Delete("/history", s.handleClearHistory)

		// Saved Snippets
		r.Get("/snippets", s.handleGetSnippets)
		r.Post("/snippets", s.handleSaveSnippet)
		r.Delete("/snippets/{id}", s.handleDeleteSnippet)

		// gRPC Studio (v2)
		r.Post("/grpc/reflect", s.handleGRPCReflect)
		r.Post("/grpc/invoke", s.handleGRPCInvoke)
		r.Post("/grpc/proto/parse", s.handleGRPCParseProto)
	})

	// Static Web UI assets
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		r.Handle("/*", fileServer)
	}

	s.router = r
}

// Helpers
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// getActiveClient safely retrieves the current ES client.
func (s *Server) getActiveClient() (*es.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.esClient == nil {
		return nil, fmt.Errorf("no active cluster connection")
	}
	return s.esClient, nil
}
