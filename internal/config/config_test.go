package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eskhan-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("failed to init manager: %v", err)
	}

	cfg := mgr.GetConfig()
	if len(cfg.Connections) != 1 {
		t.Errorf("expected 1 default connection, got %d", len(cfg.Connections))
	}

	newConn := ConnectionProfile{
		ID:       "staging-es7",
		Name:     "Staging ES 7.17",
		URL:      "http://staging.internal:9200",
		AuthType: AuthBasic,
		Username: "elastic",
		Password: "password",
	}

	if err := mgr.SaveConnection(newConn); err != nil {
		t.Fatalf("failed to save connection: %v", err)
	}

	if err := mgr.SetActiveConnection("staging-es7"); err != nil {
		t.Fatalf("failed to set active connection: %v", err)
	}

	active, err := mgr.GetActiveConnection()
	if err != nil {
		t.Fatalf("failed to get active connection: %v", err)
	}
	if active.ID != "staging-es7" {
		t.Errorf("expected active connection ID staging-es7, got %s", active.ID)
	}

	// Reload manager to verify persistence
	mgr2, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("failed to reload manager: %v", err)
	}
	if len(mgr2.GetConfig().Connections) != 2 {
		t.Errorf("expected 2 connections after reload, got %d", len(mgr2.GetConfig().Connections))
	}

	// Delete connection
	if err := mgr2.DeleteConnection("staging-es7"); err != nil {
		t.Fatalf("failed to delete connection: %v", err)
	}
	if len(mgr2.GetConfig().Connections) != 1 {
		t.Errorf("expected 1 connection after delete, got %d", len(mgr2.GetConfig().Connections))
	}
}

func TestEnvironmentManagement(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "eskhan-env-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	mgr, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("failed to init manager: %v", err)
	}

	envs := mgr.GetEnvironments()
	if len(envs) != 1 {
		t.Fatalf("expected 1 default environment, got %d", len(envs))
	}

	// Add new environment
	stagingEnv := Environment{
		ID:   "env-staging",
		Name: "Staging Cluster",
		Variables: map[string]string{
			"base_url":    "https://staging.api.com",
			"grpc_target": "staging.grpc.internal:50051",
			"token":       "stage-secret",
		},
	}
	if err := mgr.SaveEnvironment(stagingEnv); err != nil {
		t.Fatalf("failed to save environment: %v", err)
	}

	if err := mgr.SetActiveEnvironment("env-staging"); err != nil {
		t.Fatalf("failed to set active environment: %v", err)
	}

	active, err := mgr.GetActiveEnvironment()
	if err != nil || active == nil {
		t.Fatalf("failed to get active environment: %v", err)
	}
	if active.ID != "env-staging" || active.Variables["token"] != "stage-secret" {
		t.Errorf("unexpected active environment: %+v", active)
	}

	// Persistence reload
	mgr2, err := NewManager(configPath)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if len(mgr2.GetEnvironments()) != 2 {
		t.Errorf("expected 2 environments after reload, got %d", len(mgr2.GetEnvironments()))
	}

	// Delete environment
	if err := mgr2.DeleteEnvironment("env-staging"); err != nil {
		t.Fatalf("failed to delete environment: %v", err)
	}
	if len(mgr2.GetEnvironments()) != 1 {
		t.Errorf("expected 1 environment after delete, got %d", len(mgr2.GetEnvironments()))
	}
}
