package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuthType represents the authentication method used for Elasticsearch.
type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBasic  AuthType = "basic"
	AuthAPIKey AuthType = "apikey"
	AuthBearer AuthType = "bearer"
)

// ConnectionProfile holds connection details for an Elasticsearch cluster.
type ConnectionProfile struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	AuthType           AuthType          `json:"auth_type"`
	Username           string            `json:"username,omitempty"`
	Password           string            `json:"password,omitempty"`
	APIKey             string            `json:"api_key,omitempty"`
	BearerToken        string            `json:"bearer_token,omitempty"`
	InsecureSkipVerify bool              `json:"insecure_skip_verify"`
	CustomHeaders      map[string]string `json:"custom_headers,omitempty"`
	DefaultIndex       string            `json:"default_index,omitempty"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	CreatedAt          time.Time         `json:"created_at"`
}

// Environment represents a named collection of variables for API requests.
type Environment struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// AppConfig represents global application settings and profiles.
type AppConfig struct {
	ActiveConnectionID  string              `json:"active_connection_id"`
	Connections         []ConnectionProfile `json:"connections"`
	ActiveEnvironmentID string              `json:"active_environment_id,omitempty"`
	Environments        []Environment       `json:"environments,omitempty"`
	Theme               string              `json:"theme"`
	MaxHistoryItems     int                 `json:"max_history_items"`
	ServerPort          int                 `json:"server_port"`
}

// Manager handles thread-safe loading and saving of configuration.
type Manager struct {
	mu       sync.RWMutex
	filePath string
	config   AppConfig
}

// NewManager creates a new config manager.
func NewManager(customPath string) (*Manager, error) {
	path := customPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			path = filepath.Join(".", ".eskhan", "config.json")
		} else {
			path = filepath.Join(home, ".eskhan", "config.json")
		}
	}

	m := &Manager{
		filePath: path,
		config: AppConfig{
			Connections:     []ConnectionProfile{},
			Theme:           "vs-dark",
			MaxHistoryItems: 500,
			ServerPort:      8989,
		},
	}

	if err := m.Load(); err != nil {
		// If file doesn't exist, save default with a local cluster profile
		if os.IsNotExist(err) {
			defaultProfile := ConnectionProfile{
				ID:                 "default-local",
				Name:               "Local Elasticsearch",
				URL:                "http://localhost:9200",
				AuthType:           AuthNone,
				InsecureSkipVerify: true,
				TimeoutSeconds:     30,
				CreatedAt:          time.Now(),
			}
			m.config.Connections = append(m.config.Connections, defaultProfile)
			m.config.ActiveConnectionID = defaultProfile.ID

			defaultEnv := Environment{
				ID:   "default-dev",
				Name: "Local Dev",
				Variables: map[string]string{
					"base_url":    "http://localhost:8080",
					"grpc_target": "localhost:50051",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			m.config.Environments = append(m.config.Environments, defaultEnv)
			m.config.ActiveEnvironmentID = defaultEnv.ID

			if saveErr := m.Save(); saveErr != nil {
				return nil, fmt.Errorf("failed to save initial config: %w", saveErr)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	return m, nil
}

// Load reads the config file from disk.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	m.config = cfg
	return nil
}

// Save writes the current config to disk.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", m.filePath, err)
	}

	return nil
}

// GetConfig returns a copy of current config.
func (m *Manager) GetConfig() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetActiveConnection returns the currently active connection profile.
func (m *Manager) GetActiveConnection() (*ConnectionProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.config.Connections) == 0 {
		return nil, fmt.Errorf("no connection profiles configured")
	}

	for _, conn := range m.config.Connections {
		if conn.ID == m.config.ActiveConnectionID {
			connCopy := conn
			return &connCopy, nil
		}
	}

	// Fallback to first profile
	first := m.config.Connections[0]
	return &first, nil
}

// SaveConnection adds or updates a connection profile.
func (m *Manager) SaveConnection(conn ConnectionProfile) error {
	m.mu.Lock()
	if conn.ID == "" {
		conn.ID = fmt.Sprintf("conn-%d", time.Now().UnixNano())
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}
	if conn.TimeoutSeconds <= 0 {
		conn.TimeoutSeconds = 30
	}

	found := false
	for i, c := range m.config.Connections {
		if c.ID == conn.ID {
			m.config.Connections[i] = conn
			found = true
			break
		}
	}
	if !found {
		m.config.Connections = append(m.config.Connections, conn)
	}

	if m.config.ActiveConnectionID == "" || len(m.config.Connections) == 1 {
		m.config.ActiveConnectionID = conn.ID
	}
	m.mu.Unlock()

	return m.Save()
}

// DeleteConnection removes a connection profile.
func (m *Manager) DeleteConnection(id string) error {
	m.mu.Lock()
	var updated []ConnectionProfile
	for _, c := range m.config.Connections {
		if c.ID != id {
			updated = append(updated, c)
		}
	}
	m.config.Connections = updated

	if m.config.ActiveConnectionID == id {
		if len(updated) > 0 {
			m.config.ActiveConnectionID = updated[0].ID
		} else {
			m.config.ActiveConnectionID = ""
		}
	}
	m.mu.Unlock()

	return m.Save()
}

// SetActiveConnection sets the active connection ID.
func (m *Manager) SetActiveConnection(id string) error {
	m.mu.Lock()
	found := false
	for _, c := range m.config.Connections {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		m.mu.Unlock()
		return fmt.Errorf("connection profile %s not found", id)
	}

	m.config.ActiveConnectionID = id
	m.mu.Unlock()

	return m.Save()
}

// GetEnvironments returns a copy of all environments.
func (m *Manager) GetEnvironments() []Environment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]Environment, len(m.config.Environments))
	copy(res, m.config.Environments)
	return res
}

// GetActiveEnvironment returns the currently active environment or nil if none.
func (m *Manager) GetActiveEnvironment() (*Environment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.ActiveEnvironmentID == "" {
		return nil, nil
	}

	for _, env := range m.config.Environments {
		if env.ID == m.config.ActiveEnvironmentID {
			envCopy := env
			return &envCopy, nil
		}
	}

	return nil, nil
}

// SaveEnvironment adds or updates an environment.
func (m *Manager) SaveEnvironment(env Environment) error {
	m.mu.Lock()
	now := time.Now()
	if env.ID == "" {
		env.ID = fmt.Sprintf("env-%d", now.UnixNano())
		env.CreatedAt = now
	}
	env.UpdatedAt = now
	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}

	found := false
	for i, e := range m.config.Environments {
		if e.ID == env.ID {
			m.config.Environments[i] = env
			found = true
			break
		}
	}
	if !found {
		m.config.Environments = append(m.config.Environments, env)
	}

	if m.config.ActiveEnvironmentID == "" {
		m.config.ActiveEnvironmentID = env.ID
	}
	m.mu.Unlock()

	return m.Save()
}

// DeleteEnvironment removes an environment.
func (m *Manager) DeleteEnvironment(id string) error {
	m.mu.Lock()
	var updated []Environment
	for _, e := range m.config.Environments {
		if e.ID != id {
			updated = append(updated, e)
		}
	}
	m.config.Environments = updated

	if m.config.ActiveEnvironmentID == id {
		if len(updated) > 0 {
			m.config.ActiveEnvironmentID = updated[0].ID
		} else {
			m.config.ActiveEnvironmentID = ""
		}
	}
	m.mu.Unlock()

	return m.Save()
}

// SetActiveEnvironment switches the active environment.
func (m *Manager) SetActiveEnvironment(id string) error {
	m.mu.Lock()
	if id == "" {
		m.config.ActiveEnvironmentID = ""
		m.mu.Unlock()
		return m.Save()
	}

	found := false
	for _, e := range m.config.Environments {
		if e.ID == id {
			found = true
			break
		}
	}
	if !found {
		m.mu.Unlock()
		return fmt.Errorf("environment %s not found", id)
	}

	m.config.ActiveEnvironmentID = id
	m.mu.Unlock()

	return m.Save()
}
