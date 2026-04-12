package flatpak

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "flatpak" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{
		"flatpak", "list", "--app",
		"--columns=name,application,version,origin,installation",
	})
	if err != nil {
		return nil, err
	}
	pkgs := parseList(out)
	rtOut, _ := adapters.RunCmd([]string{
		"flatpak", "list", "--runtime",
		"--columns=name,application,version,origin",
	})
	pkgs = append(pkgs, parseRuntimes(rtOut)...)

	home := os.Getenv("HOME")
	for i := range pkgs {
		dirs := []string{
			"/var/lib/flatpak/app/" + pkgs[i].Path,
			filepath.Join(home, ".local/share/flatpak/app/", pkgs[i].Path),
		}
		for _, d := range dirs {
			if info, err := os.Stat(d); err == nil {
				pkgs[i].InstallDate = info.ModTime()
				break
			}
		}
	}

	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"flatpak", "uninstall", "-y"}
	for _, p := range pkgs {
		args = append(args, flatpakID(p))
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"flatpak", "update", "-y"}
	for _, p := range pkgs {
		args = append(args, flatpakID(p))
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

// flatpakID returns the application ID stored in Path, falling back to Name.
func flatpakID(p domain.Package) string {
	if p.Path != "" {
		return p.Path
	}
	return p.Name
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"flatpak", "info", flatpakID(pkg)})
	if err != nil {
		return domain.PackageInfo{Package: pkg}, err
	}
	info := domain.PackageInfo{Package: pkg}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Description:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
		}
	}
	return info, nil
}

func parseRuntimes(out string) []domain.Package {
	var pkgs []domain.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		origin := ""
		if len(parts) >= 4 {
			origin = strings.TrimSpace(parts[3])
		}
		pkgs = append(pkgs, domain.Package{
			Name:     strings.TrimSpace(parts[0]),
			Path:     strings.TrimSpace(parts[1]),
			Version:  strings.TrimSpace(parts[2]),
			Manager:  domain.ManagerFlatpak,
			Origin:   origin,
			IsOrphan: true,
		})
	}
	return pkgs
}

func parseList(out string) []domain.Package {
	var pkgs []domain.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:    strings.TrimSpace(parts[0]), // display name
			Path:    strings.TrimSpace(parts[1]), // application ID (used for operations)
			Version: strings.TrimSpace(parts[2]),
			Manager: domain.ManagerFlatpak,
			Origin:  strings.TrimSpace(parts[3]),
		})
	}
	return pkgs
}
