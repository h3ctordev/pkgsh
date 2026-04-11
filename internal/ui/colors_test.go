package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

func TestManagerColor(t *testing.T) {
	tests := []struct {
		mgr      domain.ManagerType
		expected lipgloss.Color
	}{
		{domain.ManagerApt, colorApt},
		{domain.ManagerSnap, colorSnap},
		{domain.ManagerFlatpak, colorFlatpak},
		{domain.ManagerNpm, colorNpm},
		{domain.ManagerPip, colorPip},
		{domain.ManagerDpkg, colorDpkg},
		{domain.ManagerAppImage, colorAppImage},
	}
	for _, tt := range tests {
		got := managerColor(tt.mgr)
		if got != tt.expected {
			t.Errorf("managerColor(%s) = %v, want %v", tt.mgr, got, tt.expected)
		}
	}
}

func TestManagerColor_Unknown(t *testing.T) {
	got := managerColor("unknown")
	if got != colorMuted {
		t.Errorf("expected muted color for unknown manager, got %v", got)
	}
}
