package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/adapters/appimage"
	"github.com/hbustos/pkgsh/internal/adapters/apt"
	"github.com/hbustos/pkgsh/internal/adapters/dpkg"
	"github.com/hbustos/pkgsh/internal/adapters/flatpak"
	"github.com/hbustos/pkgsh/internal/adapters/npm"
	"github.com/hbustos/pkgsh/internal/adapters/pip"
	"github.com/hbustos/pkgsh/internal/adapters/snap"
	"github.com/hbustos/pkgsh/internal/domain"
	"github.com/hbustos/pkgsh/internal/ui"
)

var version = "dev"

func main() {
	var (
		managerFlag     = flag.String("manager", "", "filtrar por gestor (apt,snap,flatpak,dpkg,pip,npm,appimage)")
		upgradeableFlag = flag.Bool("upgradeable", false, "solo paquetes con actualizaciones disponibles")
		nativeFlag      = flag.Bool("native", false, "solo paquetes nativos del OS")
		searchFlag      = flag.String("search", "", "búsqueda inicial al arrancar")
		verFlag         = flag.Bool("version", false, "mostrar versión")
	)
	flag.Parse()

	if *verFlag {
		fmt.Printf("pkgsh %s\n", version)
		os.Exit(0)
	}

	var manager domain.ManagerType
	if *managerFlag != "" {
		manager = domain.ManagerType(strings.ToLower(strings.SplitN(*managerFlag, ",", 2)[0]))
	}

	opts := ui.Options{
		Manager:     manager,
		Upgradeable: *upgradeableFlag,
		Native:      *nativeFlag,
		Search:      *searchFlag,
	}

	pkgs := loadPackages()

	p := tea.NewProgram(ui.New(pkgs, opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// loadPackages ejecuta todos los adapters en paralelo y combina los resultados.
// Los errores por adapter no disponible se ignoran silenciosamente (el gestor puede no estar instalado).
func loadPackages() []domain.Package {
	managers := []domain.PackageManager{
		apt.New(),
		snap.New(),
		flatpak.New(),
		dpkg.New(),
		pip.New(),
		npm.New(),
		appimage.New(),
	}

	type result struct {
		pkgs []domain.Package
	}

	results := make([]result, len(managers))
	var wg sync.WaitGroup
	for i, mgr := range managers {
		wg.Add(1)
		go func(idx int, m domain.PackageManager) {
			defer wg.Done()
			pkgs, _ := m.List()
			results[idx] = result{pkgs: pkgs}
		}(i, mgr)
	}
	wg.Wait()

	var all []domain.Package
	for _, r := range results {
		all = append(all, r.pkgs...)
	}
	return all
}
