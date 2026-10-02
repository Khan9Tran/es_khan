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
