package appimage

import "github.com/hbustos/pkgsh/internal/domain"

// Adapter implementa domain.PackageManager para appimage.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "appimage" }

func (a *Adapter) List() ([]domain.Package, error) {
	// TODO: implementar
	return nil, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	// TODO: implementar
	return domain.NewOperation()
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	// TODO: implementar
	return domain.NewOperation()
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	// TODO: implementar
	return domain.PackageInfo{}, nil
}
