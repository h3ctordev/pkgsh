package domain

import "time"

type ManagerType string

const (
	ManagerApt      ManagerType = "apt"
	ManagerSnap     ManagerType = "snap"
	ManagerFlatpak  ManagerType = "flatpak"
	ManagerDpkg     ManagerType = "dpkg"
	ManagerPip      ManagerType = "pip"
	ManagerNpm      ManagerType = "npm"
	ManagerAppImage ManagerType = "appimage"
)

type Package struct {
	Name        string
	Version     string
	NewVersion  string
	Manager     ManagerType
	Size        int64
	Description string
	IsNative    bool
	Origin      string
	InstallDate time.Time
}

type PackageManager interface {
	Name() string
	List() ([]Package, error)
	Remove(pkgs []Package) *Operation
	Update(pkgs []Package) *Operation
	Info(pkg Package) (PackageInfo, error)
}

type PackageInfo struct {
	Package
	Depends     []string
	InstalledBy string
}
