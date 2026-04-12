package apt

import (
	"strconv"
	"strings"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "apt" }

func (a *Adapter) List() ([]domain.Package, error) {
	out, err := adapters.RunCmd([]string{
		"dpkg-query", "-W",
		"-f=${Package}\t${Version}\t${Installed-Size}\t${Source}\n",
	})
	if err != nil {
		return nil, err
	}
	pkgs := parseList(out)

	upOut, _ := adapters.RunCmd([]string{"apt", "list", "--upgradable"})
	upgrades := parseUpgradable(upOut)
	for i := range pkgs {
		if nv, ok := upgrades[pkgs[i].Name]; ok {
			pkgs[i].NewVersion = nv
		}
	}

	autoOut, _ := adapters.RunCmd([]string{"apt-mark", "showauto"})
	autoInstalled := parseAutoInstalled(autoOut)
	for i := range pkgs {
		pkgs[i].IsOrphan = autoInstalled[pkgs[i].Name]
	}
	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"sudo", "-S", "-p", "PKGSH_SUDO:\n", "apt", "remove", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmdStdin(args, op.StdinReader(), op.Writer())
	return op
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	args := []string{"sudo", "-S", "-p", "PKGSH_SUDO:\n", "apt", "install", "--only-upgrade", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	go adapters.StreamCmdStdin(args, op.StdinReader(), op.Writer())
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := adapters.RunCmd([]string{"apt", "show", pkg.Name})
	if err != nil {
		return domain.PackageInfo{Package: pkg}, err
	}
	return parseInfo(out, pkg), nil
}

func parseList(out string) []domain.Package {
	var pkgs []domain.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		sizeKB, _ := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		origin := ""
		if len(parts) >= 4 {
			origin = strings.TrimSpace(parts[3])
		}
		if name == "" || version == "" {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:     name,
			Version:  version,
			Manager:  domain.ManagerApt,
			Size:     sizeKB * 1024,
			Origin:   origin,
			IsNative: origin == "ubuntu" || origin == "",
		})
	}
	return pkgs
}

func parseUpgradable(out string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "[upgradable") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.Split(parts[0], "/")[0]
		newVersion := parts[1]
		result[name] = newVersion
	}
	return result
}

func parseInfo(out string, base domain.Package) domain.PackageInfo {
	info := domain.PackageInfo{Package: base}
	var descLines []string
	inDesc := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Description:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			inDesc = true
			continue
		}
		if inDesc {
			if line == "" || (len(line) > 0 && line[0] != ' ') {
				inDesc = false
			} else {
				descLines = append(descLines, strings.TrimSpace(line))
			}
		}
	}
	if len(descLines) > 0 {
		info.Description += "\n" + strings.Join(descLines, "\n")
	}
	return info
}

func parseAutoInstalled(out string) map[string]bool {
	result := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result[line] = true
		}
	}
	return result
}
