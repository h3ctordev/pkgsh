package npm

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := `{
  "dependencies": {
    "typescript": { "version": "5.4.5" },
    "npm":        { "version": "10.5.0" }
  }
}`
	pkgs, err := parseList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	found := false
	for _, p := range pkgs {
		if p.Name == "typescript" && p.Version == "5.4.5" {
			found = true
		}
		if p.Manager != domain.ManagerNpm {
			t.Errorf("expected npm manager, got %q", p.Manager)
		}
	}
	if !found {
		t.Error("typescript 5.4.5 not found in parsed packages")
	}
}

func TestParseList_Empty(t *testing.T) {
	raw := `{"dependencies": {}}`
	pkgs, err := parseList(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}
