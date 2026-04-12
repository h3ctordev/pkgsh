package pip

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "pip" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{"pip", "list", "--format=json"})
	if err != nil {
		out, err = adapters.RunCmd([]string{"pip3", "list", "--format=json"})
		if err != nil {
			return nil, err
		}
	}
	pkgs, err := parseList(out)
	if err != nil {
		return nil, err
	}

	siteOut, _ := adapters.RunCmd([]string{"python3", "-c", "import site; print(site.getsitepackages()[0])"})
	siteDir := strings.TrimSpace(siteOut)
	if siteDir != "" {
		for i := range pkgs {
			pattern := filepath.Join(siteDir, pkgs[i].Name+"-*.dist-info")
			matches, _ := filepath.Glob(pattern)
			if len(matches) == 0 {
				pattern = filepath.Join(siteDir, strings.ReplaceAll(pkgs[i].Name, "-", "_")+"-*.dist-info")
				matches, _ = filepath.Glob(pattern)
			}
			if len(matches) > 0 {
				if info, err := os.Stat(matches[0]); err == nil {
					pkgs[i].InstallDate = info.ModTime()
				}
			}
		}
	}

	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"pip", "uninstall", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"pip", "install", "--upgrade"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"pip", "show", pkg.Name})
	if err != nil {
		return domain.PackageInfo{Package: pkg}, err
	}
	info := domain.PackageInfo{Package: pkg}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Summary:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "Summary:"))
		}
	}
	return info, nil
}

type pipEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func parseList(out string) ([]domain.Package, error) {
	var entries []pipEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &entries); err != nil {
		return nil, err
	}
	pkgs := make([]domain.Package, len(entries))
	for i, e := range entries {
		pkgs[i] = domain.Package{
			Name:    e.Name,
			Version: e.Version,
			Manager: domain.ManagerPip,
			Origin:  "pypi",
		}
	}
	return pkgs, nil
}
