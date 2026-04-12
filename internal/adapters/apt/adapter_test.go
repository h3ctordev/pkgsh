package apt

import (
	"strings"
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := "firefox\t126.0\t239616\tubuntu\ncurl\t7.88.1\t512\tubuntu\n"
	pkgs := parseList(raw)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	ff := pkgs[0]
	if ff.Name != "firefox" {
		t.Errorf("expected firefox, got %q", ff.Name)
	}
	if ff.Version != "126.0" {
		t.Errorf("expected version 126.0, got %q", ff.Version)
	}
	if ff.Size != 239616*1024 {
		t.Errorf("expected size %d, got %d", 239616*1024, ff.Size)
	}
	if ff.Manager != domain.ManagerApt {
		t.Errorf("expected manager apt, got %q", ff.Manager)
	}
	if ff.Origin != "ubuntu" {
		t.Errorf("expected origin ubuntu, got %q", ff.Origin)
	}
}

func TestParseUpgradable(t *testing.T) {
	raw := "firefox/jammy-updates 127.0 amd64 [upgradable from: 126.0]\ncurl/jammy-updates 7.90.0 amd64 [upgradable from: 7.88.1]\n"
	upgrades := parseUpgradable(raw)
	if upgrades["firefox"] != "127.0" {
		t.Errorf("expected firefox→127.0, got %q", upgrades["firefox"])
	}
	if upgrades["curl"] != "7.90.0" {
		t.Errorf("expected curl→7.90.0, got %q", upgrades["curl"])
	}
}

func TestParseInfo(t *testing.T) {
	raw := `Package: firefox
Version: 126.0
Description: Mozilla Firefox web browser
Homepage: https://www.mozilla.org
`
	info := parseInfo(raw, domain.Package{Name: "firefox", Manager: domain.ManagerApt})
	if !strings.Contains(info.Description, "Firefox") {
		t.Errorf("expected description to contain Firefox, got %q", info.Description)
	}
}

func TestParseAutoInstalled(t *testing.T) {
	raw := "libssl3\nlibgcc-s1\nlibc6\n"
	result := parseAutoInstalled(raw)
	if !result["libssl3"] {
		t.Error("expected libssl3 to be auto-installed")
	}
	if !result["libc6"] {
		t.Error("expected libc6 to be auto-installed")
	}
	if result["firefox"] {
		t.Error("firefox should not be in auto-installed set")
	}
}

func TestListSetsIsOrphan(t *testing.T) {
	raw := "firefox\t126.0\t239616\tubuntu\ncurl\t7.88.1\t512\tubuntu\n"
	pkgs := parseList(raw)
	autoInstalled := parseAutoInstalled("curl\n")
	for i := range pkgs {
		pkgs[i].IsOrphan = autoInstalled[pkgs[i].Name]
	}
	ff := pkgs[0]
	cu := pkgs[1]
	if ff.IsOrphan {
		t.Error("firefox should not be orphan")
	}
	if !cu.IsOrphan {
		t.Error("curl should be orphan")
	}
}
