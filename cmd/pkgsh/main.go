package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
		managerFlag       = flag.String("manager", "", "filtrar por gestor (apt,snap,flatpak,dpkg,pip,npm,appimage)")
		upgradeableFlag   = flag.Bool("upgradeable", false, "solo paquetes con actualizaciones disponibles")
		nativeFlag        = flag.Bool("native", false, "solo paquetes nativos del OS")
		searchFlag        = flag.String("search", "", "búsqueda inicial al arrancar")
		noSecurityMode    = flag.Bool("no-security-mode", false, "desactivar protección de paquetes del sistema")
		verFlag           = flag.Bool("version", false, "mostrar versión")
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
		Manager:       manager,
		Upgradeable:   *upgradeableFlag,
		Native:        *nativeFlag,
		Search:        *searchFlag,
		SecurityMode:  !*noSecurityMode,
	}

	adapterMap := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt:      apt.New(),
		domain.ManagerSnap:     snap.New(),
		domain.ManagerFlatpak:  flatpak.New(),
		domain.ManagerDpkg:     dpkg.New(),
		domain.ManagerPip:      pip.New(),
		domain.ManagerNpm:      npm.New(),
		domain.ManagerAppImage: appimage.New(),
	}

	p := tea.NewProgram(ui.New(nil, adapterMap, opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
