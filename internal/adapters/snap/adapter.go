package snap

import (
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "snap" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{"snap", "list", "--unicode=never"})
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"sudo", "snap", "remove"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"sudo", "snap", "refresh"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmd(args, op.Writer())
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"snap", "info", pkg.Name})
	if err != nil {
		return domain.PackageInfo{Package: pkg}, err
	}
	info := domain.PackageInfo{Package: pkg}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "summary:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "summary:"))
		}
	}
	return info, nil
}

func parseList(out string) []domain.Package {
	var pkgs []domain.Package
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return pkgs
	}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:    fields[0],
			Version: fields[1],
			Manager: domain.ManagerSnap,
			Origin:  fields[4],
		})
	}
	return pkgs
}
