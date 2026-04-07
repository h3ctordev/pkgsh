package appimage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestScanDir(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"MyApp-1.0.AppImage", "OtherTool-2.3-x86_64.AppImage", "notanapp.txt"} {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	pkgs := scanDir(dir)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 AppImages, got %d", len(pkgs))
	}
	for _, p := range pkgs {
		if p.Manager != domain.ManagerAppImage {
			t.Errorf("expected appimage manager, got %q", p.Manager)
		}
		if p.Path == "" {
			t.Error("expected non-empty path")
		}
	}
}

func TestParseAppImageName(t *testing.T) {
	cases := []struct {
		filename string
		name     string
		version  string
	}{
		{"MyApp-1.0.AppImage", "MyApp", "1.0"},
		{"OtherTool-2.3-x86_64.AppImage", "OtherTool", "2.3"},
		{"SimpleApp.AppImage", "SimpleApp", ""},
	}
	for _, c := range cases {
		name, version := parseAppImageName(c.filename)
		if name != c.name {
			t.Errorf("parseAppImageName(%q) name = %q, want %q", c.filename, name, c.name)
		}
		if version != c.version {
			t.Errorf("parseAppImageName(%q) version = %q, want %q", c.filename, version, c.version)
		}
	}
}
