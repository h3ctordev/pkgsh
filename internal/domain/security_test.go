package domain_test

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestIsSystemPackage(t *testing.T) {
	cases := []struct {
		name    string
		manager domain.ManagerType
		want    bool
	}{
		// Nombres exactos protegidos
		{"libc6", domain.ManagerApt, true},
		{"bash", domain.ManagerApt, true},
		{"systemd", domain.ManagerApt, true},
		{"sudo", domain.ManagerApt, true},
		{"dpkg", domain.ManagerApt, true},
		{"apt", domain.ManagerApt, true},
		{"coreutils", domain.ManagerDpkg, true},
		// Patrones de prefijo
		{"linux-image-6.5.0-generic", domain.ManagerApt, true},
		{"linux-headers-6.5.0", domain.ManagerDpkg, true},
		{"linux-modules-6.5.0-generic", domain.ManagerApt, true},
		{"linux-modules-extra-6.5.0-generic", domain.ManagerApt, true},
		{"libc-bin", domain.ManagerApt, true},
		{"libgcc-12-dev", domain.ManagerApt, true},
		{"libstdc++6", domain.ManagerApt, true},
		// Paquetes normales — no protegidos
		{"vim", domain.ManagerApt, false},
		{"curl", domain.ManagerApt, false},
		{"vlc", domain.ManagerApt, false},
		{"htop", domain.ManagerApt, false},
		// Otros managers — nunca protegidos
		{"bash", domain.ManagerSnap, false},
		{"libc6", domain.ManagerPip, false},
		{"systemd", domain.ManagerFlatpak, false},
		{"coreutils", domain.ManagerNpm, false},
		{"linux-image-6.5.0", domain.ManagerAppImage, false},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/"+string(tc.manager), func(t *testing.T) {
			pkg := domain.Package{Name: tc.name, Manager: tc.manager}
			got := domain.IsSystemPackage(pkg)
			if got != tc.want {
				t.Errorf("IsSystemPackage(%q, %q) = %v, want %v", tc.name, tc.manager, got, tc.want)
			}
		})
	}
}
