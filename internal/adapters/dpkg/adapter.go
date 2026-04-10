package dpkg

import (
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "dpkg" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{"dpkg-query", "-l"})
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

// Remove es no-op: dpkg es read-only en pkgsh (usar apt para desinstalar).
func (a *Adapter) Remove(_ []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	op.Done(nil)
	return op
}

// Update es no-op.
func (a *Adapter) Update(_ []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	op.Done(nil)
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"dpkg", "-s", pkg.Name})
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

// parseList parsea la salida de dpkg-query -l.
// Solo incluye paquetes con estado "ii" (instalados correctamente).
func parseList(out string) []domain.Package {
	var pkgs []domain.Package
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "ii ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := fields[1]
		if idx := strings.Index(name, ":"); idx != -1 {
			name = name[:idx]
		}
		pkgs = append(pkgs, domain.Package{
			Name:     name,
			Version:  fields[2],
			Manager:  domain.ManagerDpkg,
			IsNative: true,
		})
	}
	return pkgs
}
