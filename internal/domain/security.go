package domain

import "strings"

// systemPackageNames es el conjunto de nombres exactos de paquetes críticos del sistema.
// Solo aplica a managers apt y dpkg.
var systemPackageNames = map[string]bool{
	"libc6": true, "libc6-dev": true, "libc-bin": true, "libc-dev-bin": true,
	"bash": true, "dash": true, "sh": true,
	"coreutils": true, "util-linux": true,
	"systemd": true, "systemd-sysv": true, "systemd-timesyncd": true,
	"init": true, "sysvinit-utils": true,
	"dpkg": true, "apt": true, "apt-utils": true,
	"login": true, "passwd": true, "shadow-utils": true,
	"sudo": true,
	"mount": true,
	"e2fsprogs": true, "btrfs-progs": true,
	"udev": true,
	"dbus": true, "libdbus-1-3": true,
	"libpam-runtime": true, "libpam-modules": true, "libpam0g": true,
	"linux-base": true,
	"gcc-12-base": true, "libgcc-s1": true,
	"libstdc++6": true,
	"tar": true, "gzip": true, "bzip2": true, "xz-utils": true,
	"sed": true, "grep": true, "gawk": true,
	"findutils": true,
	"procps": true,
	"python3-minimal": true,
}

// systemPackagePrefixes son prefijos que marcan automáticamente un paquete como del sistema.
var systemPackagePrefixes = []string{
	"linux-image-",
	"linux-headers-",
	"linux-modules-",
	"linux-modules-extra-",
	"libc-",
	"libgcc-",
	"libstdc++",
	"python3-apt",
}

// IsSystemPackage devuelve true si el paquete es considerado crítico del sistema.
// Solo aplica a managers apt y dpkg; siempre devuelve false para snap, flatpak, pip, npm, appimage.
func IsSystemPackage(pkg Package) bool {
	if pkg.Manager != ManagerApt && pkg.Manager != ManagerDpkg {
		return false
	}
	name := strings.ToLower(pkg.Name)
	if systemPackageNames[name] {
		return true
	}
	for _, prefix := range systemPackagePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
