package ui

import (
	"strings"
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestDetailModel_RenderPackage(t *testing.T) {
	dm := newDetailModel()
	pkg := &domain.Package{
		Name:        "firefox",
		Version:     "126.0",
		NewVersion:  "127.0",
		Manager:     domain.ManagerApt,
		Size:        234 * 1024 * 1024,
		Description: "Web browser",
		IsNative:    false,
		Origin:      "ubuntu",
	}
	view := dm.View(pkg, 40, 20, false)
	for _, want := range []string{"firefox", "126.0", "127.0", "apt", "Web browser", "ubuntu"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected %q en detail view, got:\n%s", want, view)
		}
	}
}

func TestDetailModel_RenderNil(t *testing.T) {
	dm := newDetailModel()
	view := dm.View(nil, 40, 20, false)
	if !strings.Contains(view, "Selecciona") {
		t.Errorf("expected placeholder para nil package, got: %s", view)
	}
}

func TestDetailModel_FormatSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{2048, "2.0 KB"},
		{5 * 1024 * 1024, "5.0 MB"},
		{2 * 1024 * 1024 * 1024, "2.0 GB"},
	}
	for _, c := range cases {
		got := formatSize(c.bytes)
		if got != c.want {
			t.Errorf("formatSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}
