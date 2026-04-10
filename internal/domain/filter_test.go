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
	filtered := domain.Filter(testPkgs(), "", domain.ManagerNpm)
	if len(filtered) != 1 || filtered[0].Name != "node" {
		t.Errorf("expected 1 npm package, got %v", filtered)
	}
}

func TestFilterByQuery(t *testing.T) {
	filtered := domain.Filter(testPkgs(), "fire", "")
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
