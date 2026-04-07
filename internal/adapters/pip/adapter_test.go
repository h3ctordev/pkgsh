package pip

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := `[{"name": "requests", "version": "2.31.0"}, {"name": "numpy", "version": "1.26.4"}]`
	pkgs, err := parseList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "requests" {
		t.Errorf("expected requests, got %q", pkgs[0].Name)
	}
	if pkgs[0].Version != "2.31.0" {
		t.Errorf("expected 2.31.0, got %q", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerPip {
		t.Errorf("expected pip manager")
	}
}

func TestParseList_Empty(t *testing.T) {
	pkgs, err := parseList("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}
