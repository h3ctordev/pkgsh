package npm

import (
	"encoding/json"
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "npm" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{"npm", "list", "-g", "--json", "--depth=0"})
	if err != nil && out == "" {
		return nil, err
	}
	return parseList(out)
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"npm", "uninstall", "-g"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"npm", "update", "-g"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"npm", "info", pkg.Name, "description"})
	if err != nil {
		return domain.PackageInfo{Package: pkg}, err
	}
	info := domain.PackageInfo{Package: pkg}
	info.Description = strings.TrimSpace(out)
	return info, nil
}

type npmDep struct {
	Version string `json:"version"`
}

type npmOutput struct {
	Dependencies map[string]npmDep `json:"dependencies"`
}

func parseList(out string) ([]domain.Package, error) {
	var result npmOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		return nil, err
	}
	pkgs := make([]domain.Package, 0, len(result.Dependencies))
	for name, dep := range result.Dependencies {
		pkgs = append(pkgs, domain.Package{
			Name:    name,
			Version: dep.Version,
			Manager: domain.ManagerNpm,
			Origin:  "npmjs",
		})
	}
	return pkgs, nil
}
