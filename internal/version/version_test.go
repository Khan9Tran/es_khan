package version

import (
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Errorf("expected non-empty Version")
	}
	if info.BuiltBy != "Khan9Tran" {
		t.Errorf("expected BuiltBy 'Khan9Tran', got '%s'", info.BuiltBy)
	}
	if info.GoVersion == "" {
		t.Errorf("expected non-empty GoVersion")
	}
	if info.OS == "" || info.Arch == "" {
		t.Errorf("expected non-empty OS/Arch")
	}
}
