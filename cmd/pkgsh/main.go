package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		// Un solo gestor por ahora; multi-gestor es v2
		manager = domain.ManagerType(strings.ToLower(strings.SplitN(*managerFlag, ",", 2)[0]))
	}

	opts := ui.Options{
		Manager:     manager,
		Upgradeable: *upgradeableFlag,
		Native:      *nativeFlag,
		Search:      *searchFlag,
	}

	p := tea.NewProgram(ui.New(mockPackages(), opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func mockPackages() []domain.Package {
	now := time.Now()
	return []domain.Package{
		{
			Name: "firefox", Version: "126.0", NewVersion: "127.0",
			Manager: domain.ManagerApt, Size: 234 * 1024 * 1024,
			Description: "Navegador web de Mozilla Firefox", IsNative: false, Origin: "ubuntu",
			InstallDate: now.AddDate(0, -3, 0), Path: "/usr/bin/firefox",
		},
		{
			Name: "vim", Version: "9.0.1672",
			Manager: domain.ManagerApt, Size: 3 * 1024 * 1024,
			Description: "Editor de texto modal avanzado", IsNative: true, Origin: "ubuntu",
			Path: "/usr/bin/vim",
		},
		{
			Name: "curl", Version: "7.88.1",
			Manager: domain.ManagerApt, Size: 500 * 1024,
			Description: "Herramienta de transferencia de datos con URLs", IsNative: true, Origin: "ubuntu",
			Path: "/usr/bin/curl",
		},
		{
			Name: "git", Version: "2.43.0",
			Manager: domain.ManagerApt, Size: 10 * 1024 * 1024,
			Description: "Sistema de control de versiones distribuido", IsNative: true, Origin: "ubuntu",
			Path: "/usr/bin/git",
		},
		{
			Name: "snapd", Version: "2.63",
			Manager: domain.ManagerSnap, Size: 90 * 1024 * 1024,
			Description: "Demonio de gestión de paquetes snap", IsNative: false, Origin: "canonical",
			Path: "/snap/snapd/current",
		},
		{
			Name: "spotify", Version: "1.2.26", NewVersion: "1.2.28",
			Manager: domain.ManagerSnap, Size: 180 * 1024 * 1024,
			Description: "Cliente de música en streaming de Spotify", IsNative: false, Origin: "spotify",
			Path: "/snap/spotify/current",
		},
		{
			Name: "vlc", Version: "3.0.20",
			Manager: domain.ManagerFlatpak, Size: 210 * 1024 * 1024,
			Description: "Reproductor multimedia universal", IsNative: false, Origin: "flathub",
			Path: "/var/lib/flatpak/app/org.videolan.VLC",
		},
		{
			Name: "gimp", Version: "2.10.36",
			Manager: domain.ManagerFlatpak, Size: 320 * 1024 * 1024,
			Description: "Editor de imágenes avanzado y gratuito", IsNative: false, Origin: "flathub",
			Path: "/var/lib/flatpak/app/org.gimp.GIMP",
		},
		{
			Name: "python3-pip", Version: "23.0.1",
			Manager: domain.ManagerDpkg, Size: 5 * 1024 * 1024,
			Description: "Gestor de paquetes para Python 3", IsNative: true, Origin: "ubuntu",
			Path: "/usr/bin/pip3",
		},
		{
			Name: "requests", Version: "2.31.0",
			Manager: domain.ManagerPip, Size: 300 * 1024,
			Description: "HTTP para humanos", IsNative: false, Origin: "pypi",
			Path: "/usr/lib/python3/dist-packages/requests",
		},
		{
			Name: "numpy", Version: "1.26.4", NewVersion: "2.0.0",
			Manager: domain.ManagerPip, Size: 25 * 1024 * 1024,
			Description: "Computación científica y matrices en Python", IsNative: false, Origin: "pypi",
			Path: "/usr/lib/python3/dist-packages/numpy",
		},
		{
			Name: "typescript", Version: "5.4.5",
			Manager: domain.ManagerNpm, Size: 15 * 1024 * 1024,
			Description: "JavaScript tipado a escala", IsNative: false, Origin: "npmjs",
			Path: "/usr/local/lib/node_modules/typescript",
		},
	}
}
