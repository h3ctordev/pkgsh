package domain_test

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func testPkgs() []domain.Package {
	return []domain.Package{
		{Name: "zsh", Manager: domain.ManagerApt, Version: "5.9", Size: 2_900_000},
		{Name: "apt", Manager: domain.ManagerApt, Version: "2.6.1", Size: 4_096_000},
		{Name: "node", Manager: domain.ManagerNpm, Version: "20.1.0", Size: 98_000_000},
		{Name: "firefox", Manager: domain.ManagerApt, Version: "126.0", Size: 245_000_000},
	}
}

func TestSortByName(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortByName)
	if sorted[0].Name != "apt" || sorted[3].Name != "zsh" {
		t.Errorf("expected alphabetical order, got %v", namesOf(sorted))
	}
}

func TestSortBySize(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortBySize)
	if sorted[0].Name != "firefox" {
		t.Errorf("expected largest first, got %s", sorted[0].Name)
	}
}

func TestSortByManager(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortByManager)
	// apt < npm alphabetically, so npm should be last
	if sorted[len(sorted)-1].Manager != domain.ManagerNpm {
		t.Errorf("expected npm last, got %s", sorted[len(sorted)-1].Manager)
	}
}

func TestFilterByManager(t *testing.T) {
	filtered := domain.Filter(testPkgs(), "", domain.ManagerNpm, false)
	if len(filtered) != 1 || filtered[0].Name != "node" {
		t.Errorf("expected 1 npm package, got %v", filtered)
	}
}

func TestFilterByQuery(t *testing.T) {
	filtered := domain.Filter(testPkgs(), "fire", "", false)
	if len(filtered) != 1 || filtered[0].Name != "firefox" {
		t.Errorf("expected firefox, got %v", filtered)
	}
}

func namesOf(pkgs []domain.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.Name
	}
	return out
}

func TestDeduplicatePackages_RemovesDpkgWhenAptExists(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "bash", Manager: domain.ManagerApt, Version: "5.2"},
		{Name: "bash", Manager: domain.ManagerDpkg, Version: "5.2"},
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "only-dpkg", Manager: domain.ManagerDpkg, Version: "1.0"},
		{Name: "node", Manager: domain.ManagerNpm, Version: "20.0"},
	}
	got := domain.DeduplicatePackages(pkgs)
	byNameManager := map[string]bool{}
	for _, p := range got {
		byNameManager[p.Name+":"+string(p.Manager)] = true
	}

	if byNameManager["bash:dpkg"] {
		t.Error("bash:dpkg should be removed when bash:apt exists")
	}
	if !byNameManager["bash:apt"] {
		t.Error("bash:apt should be kept")
	}
	if !byNameManager["vim:apt"] {
		t.Error("vim:apt should be kept")
	}
	if !byNameManager["only-dpkg:dpkg"] {
		t.Error("only-dpkg:dpkg should be kept when no apt equivalent exists")
	}
	if !byNameManager["node:npm"] {
		t.Error("node:npm should be kept unchanged")
	}
	if len(got) != 4 {
		t.Errorf("expected 4 packages, got %d: %v", len(got), namesOf(got))
	}
}

func TestFilter_SecurityModeHidesSystemPackages(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "bash", Manager: domain.ManagerApt, Version: "5.2"},
		{Name: "libc6", Manager: domain.ManagerApt, Version: "2.35"},
		{Name: "node", Manager: domain.ManagerNpm, Version: "20.0"},
	}

	// Con securityMode=true: bash y libc6 deben desaparecer
	filtered := domain.Filter(pkgs, "", "", true)
	for _, p := range filtered {
		if domain.IsSystemPackage(p) {
			t.Errorf("expected system package %q to be hidden in security mode", p.Name)
		}
	}
	if len(filtered) != 2 { // vim + node
		t.Errorf("expected 2 packages with securityMode=true, got %d: %v", len(filtered), namesOf(filtered))
	}

	// Con securityMode=false: todos visibles
	all := domain.Filter(pkgs, "", "", false)
	if len(all) != 4 {
		t.Errorf("expected 4 packages with securityMode=false, got %d", len(all))
	}
}

func TestFilter_SecurityMode_MixedSelection(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "curl", Manager: domain.ManagerApt, Version: "7.88"},
		{Name: "linux-image-6.5.0-generic", Manager: domain.ManagerApt, Version: "6.5.0"},
		{Name: "htop", Manager: domain.ManagerApt, Version: "3.2"},
	}
	filtered := domain.Filter(pkgs, "", "", true)
	if len(filtered) != 2 {
		t.Errorf("expected 2 non-system packages, got %d", len(filtered))
	}
	for _, p := range filtered {
		if p.Name == "linux-image-6.5.0-generic" {
			t.Error("linux-image should be hidden in security mode")
		}
	}
}
