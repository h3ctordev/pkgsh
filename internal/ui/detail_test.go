package ui

import (
	"strings"
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestDetailView_NoPackage(t *testing.T) {
	dm := newDetailModel()
	out := dm.View(nil, 30, 10, false)
	if !strings.Contains(out, "Selecciona") {
		t.Errorf("expected placeholder text, got: %q", out)
	}
}

func TestDetailView_WithPackage(t *testing.T) {
	pkg := &domain.Package{
		Name:        "firefox",
		Version:     "124.0",
		NewVersion:  "125.0",
		Manager:     domain.ManagerApt,
		Size:        248 * 1024 * 1024,
		Description: "Mozilla Firefox web browser",
	}
	dm := newDetailModel()
	out := dm.View(pkg, 40, 15, false)
	if !strings.Contains(out, "firefox") {
		t.Errorf("expected name in output, got: %q", out)
	}
	if !strings.Contains(out, "124.0") {
		t.Errorf("expected version in output, got: %q", out)
	}
	if !strings.Contains(out, "125.0") {
		t.Errorf("expected new version in output, got: %q", out)
	}
	if !strings.Contains(out, "→") {
		t.Errorf("expected arrow in version display, got: %q", out)
	}
}
