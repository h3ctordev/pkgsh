package snap

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := `Name    Version  Rev  Tracking       Publisher  Notes
firefox 126.0    123  latest/stable  mozilla    -
snapd   2.63     456  latest/stable  canonical  snapd
`
	pkgs := parseList(raw)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" {
		t.Errorf("expected firefox, got %q", pkgs[0].Name)
	}
	if pkgs[0].Version != "126.0" {
		t.Errorf("expected 126.0, got %q", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerSnap {
		t.Errorf("expected snap manager")
	}
	if pkgs[1].Origin != "canonical" {
		t.Errorf("expected canonical, got %q", pkgs[1].Origin)
	}
}
