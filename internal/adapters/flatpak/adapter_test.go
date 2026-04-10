package flatpak

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := "Firefox Web Browser\torg.mozilla.firefox\t126.0\tflathub\tsystem\nVLC media player\torg.videolan.VLC\t3.0.20\tflathub\tsystem\n"
	pkgs := parseList(raw)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "Firefox Web Browser" {
		t.Errorf("expected 'Firefox Web Browser', got %q", pkgs[0].Name)
	}
	if pkgs[0].Version != "126.0" {
		t.Errorf("expected 126.0, got %q", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerFlatpak {
		t.Errorf("expected flatpak manager")
	}
	if pkgs[0].Origin != "flathub" {
		t.Errorf("expected flathub, got %q", pkgs[0].Origin)
	}
}
