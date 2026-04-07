package appimage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hbustos/pkgsh/internal/domain"
)

var searchDirs = []string{
	filepath.Join(os.Getenv("HOME"), "Applications"),
	filepath.Join(os.Getenv("HOME"), ".local", "bin"),
	"/opt",
}

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "appimage" }

func (a *Adapter) List() ([]domain.Package, error) {
	var pkgs []domain.Package
	for _, dir := range searchDirs {
		pkgs = append(pkgs, scanDir(dir)...)
	}
	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	go func() {
		w := op.Writer()
		for _, p := range pkgs {
			if err := os.Remove(p.Path); err != nil {
				w.Write([]byte("error removing " + p.Path + ": " + err.Error() + "\n"))
			} else {
				w.Write([]byte("removed " + p.Path + "\n"))
			}
		}
		w.Close()
	}()
	return op
}

// Update es no-op — AppImages se actualizan manualmente.
func (a *Adapter) Update(_ []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	op.Done(nil)
	return op
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}

func scanDir(dir string) []domain.Package {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var pkgs []domain.Package
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".appimage") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		info, err := e.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}
		name, version := parseAppImageName(e.Name())
		pkgs = append(pkgs, domain.Package{
			Name:    name,
			Version: version,
			Manager: domain.ManagerAppImage,
			Size:    size,
			Path:    path,
		})
	}
	return pkgs
}

var versionRe = regexp.MustCompile(`^(.+?)-(\d+\.\d+[\w.]*)(?:-[^-]+)?\.AppImage$`)

func parseAppImageName(filename string) (name, version string) {
	matches := versionRe.FindStringSubmatch(filename)
	if len(matches) == 3 {
		return matches[1], matches[2]
	}
	name = strings.TrimSuffix(filename, ".AppImage")
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return name, ""
}
