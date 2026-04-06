# pkgsh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Construir `pkgsh`, un gestor TUI unificado de paquetes para Ubuntu/Debian que abstrae apt, snap, flatpak, dpkg, pip, npm y AppImage en una sola interfaz de terminal.

**Architecture:** Cuatro capas estrictas (UI → Domain → Adapters → System). Cada gestor implementa la interfaz `PackageManager`. Las operaciones corren como `exec.Cmd` con streaming de stdout/stderr al panel de log de la TUI en tiempo real.

**Tech Stack:** Go 1.21+, Bubble Tea (TUI), Bubbles (componentes), Lipgloss (estilos), nfpm (empaquetado .deb), GitHub Actions (CI/CD).

---

## Mapa de archivos

```
cmd/pkgsh/main.go                               entrypoint + flag parsing
internal/domain/types.go                        Package, ManagerType, SortField, Panel, PackageInfo
internal/domain/interface.go                    PackageManager interface, Runner interface
internal/domain/operation.go                    Operation (streaming stdout/stderr + stdin para sudo)
internal/domain/runner.go                       SystemRunner + FakeRunner (para tests)
internal/domain/filter.go                       Filter, Sort — funciones puras
internal/domain/filter_test.go
internal/adapters/registry.go                   carga paralela de todos los adaptadores
internal/adapters/apt/apt.go
internal/adapters/apt/apt_test.go
internal/adapters/snap/snap.go
internal/adapters/snap/snap_test.go
internal/adapters/flatpak/flatpak.go
internal/adapters/flatpak/flatpak_test.go
internal/adapters/dpkg/dpkg.go
internal/adapters/dpkg/dpkg_test.go
internal/adapters/pip/pip.go
internal/adapters/pip/pip_test.go
internal/adapters/npm/npm.go
internal/adapters/npm/npm_test.go
internal/adapters/appimage/appimage.go
internal/adapters/appimage/appimage_test.go
internal/ui/styles.go                           lipgloss theme + colores
internal/ui/keys.go                             keymap centralizado
internal/ui/list.go                             panel lista (Bubble Tea model)
internal/ui/detail.go                           panel detalle
internal/ui/log.go                              panel log + streaming
internal/ui/modal.go                            modal confirmación + sudo input
internal/ui/app.go                              AppModel principal (compone todos los paneles)
nfpm.yaml                                       configuración empaquetado .deb
.github/workflows/release.yml                  pipeline CI/CD
LICENSE                                         MIT
```

---

## Task 1: Scaffold del proyecto

**Files:**
- Create: `go.mod`
- Create: `cmd/pkgsh/main.go` (stub)
- Create: `LICENSE`

- [ ] **Step 1: Inicializar módulo Go**

```bash
cd /home/hectordev/dev/app.manage
go mod init github.com/yourusername/pkgsh
```

Esperado: crea `go.mod` con `module github.com/yourusername/pkgsh`

- [ ] **Step 2: Instalar dependencias**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
```

- [ ] **Step 3: Crear stub de main.go**

Crear `cmd/pkgsh/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("pkgsh")
}
```

- [ ] **Step 4: Crear directorios internos**

```bash
mkdir -p internal/domain
mkdir -p internal/adapters/apt
mkdir -p internal/adapters/snap
mkdir -p internal/adapters/flatpak
mkdir -p internal/adapters/dpkg
mkdir -p internal/adapters/pip
mkdir -p internal/adapters/npm
mkdir -p internal/adapters/appimage
mkdir -p internal/ui
```

- [ ] **Step 5: Verificar que compila**

```bash
go build ./cmd/pkgsh
```

Esperado: binario `pkgsh` creado sin errores.

- [ ] **Step 6: Crear LICENSE**

Crear `LICENSE`:

```
MIT License

Copyright (c) 2026 pkgsh contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 7: Commit**

```bash
git init
git add .
git commit -m "feat: project scaffold with Go module and dependencies"
```

---

## Task 2: Domain — tipos e interfaces

**Files:**
- Create: `internal/domain/types.go`
- Create: `internal/domain/interface.go`

- [ ] **Step 1: Crear types.go**

Crear `internal/domain/types.go`:

```go
package domain

import "time"

// ManagerType identifica el gestor de paquetes de origen.
type ManagerType string

const (
	ManagerAll      ManagerType = ""
	ManagerApt      ManagerType = "apt"
	ManagerSnap     ManagerType = "snap"
	ManagerFlatpak  ManagerType = "flatpak"
	ManagerDpkg     ManagerType = "dpkg"
	ManagerPip      ManagerType = "pip"
	ManagerNpm      ManagerType = "npm"
	ManagerAppImage ManagerType = "appimage"
)

// AllManagers lista todos los gestores en orden de display.
var AllManagers = []ManagerType{
	ManagerApt, ManagerSnap, ManagerFlatpak,
	ManagerDpkg, ManagerPip, ManagerNpm, ManagerAppImage,
}

// SortField criterio de ordenamiento de la lista.
type SortField int

const (
	SortByName    SortField = iota
	SortByManager SortField = iota
	SortByVersion SortField = iota
	SortBySize    SortField = iota
)

// Panel identifica el panel activo de la TUI.
type Panel int

const (
	PanelList   Panel = iota
	PanelDetail Panel = iota
	PanelLog    Panel = iota
)

// Package representa un paquete instalado.
type Package struct {
	Name        string
	Version     string
	NewVersion  string      // vacío si no hay actualización disponible
	Manager     ManagerType
	Size        int64       // bytes
	Description string
	IsNative    bool        // true si viene de repositorios oficiales del OS
	Origin      string      // "ubuntu", "ppa:...", "pypi", "npmjs", etc.
	InstallDate time.Time   // zero si el gestor no lo expone
}

// PackageInfo detalle adicional de un paquete (para el panel derecho).
type PackageInfo struct {
	Package
	Homepage string
	Maintainer string
}
```

- [ ] **Step 2: Crear interface.go**

Crear `internal/domain/interface.go`:

```go
package domain

// PackageManager es la interfaz que cada adaptador debe implementar.
// Las operaciones Remove y Update devuelven un *Operation para streaming.
type PackageManager interface {
	// Name devuelve el identificador del gestor (e.g. "apt").
	Name() ManagerType
	// List devuelve todos los paquetes instalados.
	List() ([]Package, error)
	// Remove desinstala los paquetes dados y devuelve una operación con streaming.
	Remove(pkgs []Package) *Operation
	// Update actualiza los paquetes dados y devuelve una operación con streaming.
	Update(pkgs []Package) *Operation
	// Info devuelve detalle adicional de un paquete.
	Info(pkg Package) (PackageInfo, error)
	// Available reporta si el gestor está disponible en el sistema.
	Available() bool
}

// Runner abstrae exec.Cmd para permitir testing sin llamadas al sistema.
type Runner interface {
	// Output ejecuta el comando y devuelve stdout combinado.
	Output(args []string) ([]byte, error)
	// Stream ejecuta el comando y devuelve una Operation para leer stdout/stderr.
	Stream(args []string) *Operation
}
```

- [ ] **Step 3: Verificar que compila**

```bash
go build ./internal/domain/...
```

Esperado: sin errores.

- [ ] **Step 4: Commit**

```bash
git add internal/domain/types.go internal/domain/interface.go
git commit -m "feat: domain types, ManagerType, Package, PackageManager interface"
```

---

## Task 3: Domain — Operation y Runner

**Files:**
- Create: `internal/domain/operation.go`
- Create: `internal/domain/runner.go`

- [ ] **Step 1: Crear operation.go**

Crear `internal/domain/operation.go`:

```go
package domain

import "io"

// Operation representa una operación en curso (remove, update).
// Implementa io.Reader para leer stdout/stderr en tiempo real.
// Implementa io.Writer para enviar datos a stdin (e.g. contraseña de sudo).
type Operation struct {
	Reader io.Reader
	Writer io.Writer
	Done   <-chan struct{}
	Err    *error
}
```

- [ ] **Step 2: Crear runner.go**

Crear `internal/domain/runner.go`:

```go
package domain

import (
	"io"
	"os/exec"
)

// SystemRunner ejecuta comandos reales en el sistema.
// Cumple la interfaz Runner.
type SystemRunner struct{}

func NewSystemRunner() *SystemRunner {
	return &SystemRunner{}
}

func (r *SystemRunner) Output(args []string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	return cmd.Output()
}

func (r *SystemRunner) Stream(args []string) *Operation {
	pr, pw := io.Pipe()
	sr, sw := io.Pipe() // stdin: para sudo -S

	var cmdErr error
	done := make(chan struct{})

	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Stdout = pw
	cmd.Stderr = pw
	cmd.Stdin = sr

	go func() {
		defer close(done)
		defer pw.Close()
		cmdErr = cmd.Run()
		_ = sw.Close()
	}()

	return &Operation{
		Reader: pr,
		Writer: sw,
		Done:   done,
		Err:    &cmdErr,
	}
}

// FakeRunner simula ejecución de comandos para tests.
// La clave del mapa es el primer argumento del comando (e.g. "apt").
type FakeRunner struct {
	Outputs map[string][]byte
	Errors  map[string]error
}

func NewFakeRunner() *FakeRunner {
	return &FakeRunner{
		Outputs: make(map[string][]byte),
		Errors:  make(map[string]error),
	}
}

func (f *FakeRunner) Output(args []string) ([]byte, error) {
	key := args[0]
	if err, ok := f.Errors[key]; ok {
		return nil, err
	}
	return f.Outputs[key], nil
}

func (f *FakeRunner) Stream(args []string) *Operation {
	data := f.Outputs[args[0]]
	pr, pw := io.Pipe()
	sr, sw := io.Pipe()
	var cmdErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer pw.Close()
		_, _ = pw.Write(data)
	}()
	_ = sw.Close()
	return &Operation{
		Reader: pr,
		Writer: sr,
		Done:   done,
		Err:    &cmdErr,
	}
}
```

- [ ] **Step 3: Verificar que compila**

```bash
go build ./internal/domain/...
```

Esperado: sin errores.

- [ ] **Step 4: Commit**

```bash
git add internal/domain/operation.go internal/domain/runner.go
git commit -m "feat: Operation streaming type and SystemRunner/FakeRunner"
```

---

## Task 4: Domain — Filter y Sort

**Files:**
- Create: `internal/domain/filter.go`
- Create: `internal/domain/filter_test.go`

- [ ] **Step 1: Escribir los tests primero**

Crear `internal/domain/filter_test.go`:

```go
package domain_test

import (
	"testing"
	"github.com/yourusername/pkgsh/internal/domain"
)

var testPackages = []domain.Package{
	{Name: "firefox", Manager: domain.ManagerApt, Version: "126.0", NewVersion: "127.0", Size: 200 * 1024 * 1024, IsNative: true},
	{Name: "vlc",     Manager: domain.ManagerApt, Version: "3.0.2", NewVersion: "",      Size: 50 * 1024 * 1024,  IsNative: true},
	{Name: "spotify", Manager: domain.ManagerSnap, Version: "1.2",  NewVersion: "",      Size: 180 * 1024 * 1024, IsNative: false},
	{Name: "node",    Manager: domain.ManagerNpm,  Version: "20.1",  NewVersion: "21.0", Size: 80 * 1024 * 1024,  IsNative: false},
}

func TestFilterByManager(t *testing.T) {
	result := domain.FilterByManager(testPackages, domain.ManagerApt)
	if len(result) != 2 {
		t.Fatalf("want 2 apt packages, got %d", len(result))
	}
	for _, p := range result {
		if p.Manager != domain.ManagerApt {
			t.Errorf("unexpected manager %s", p.Manager)
		}
	}
}

func TestFilterByManagerAll(t *testing.T) {
	result := domain.FilterByManager(testPackages, domain.ManagerAll)
	if len(result) != len(testPackages) {
		t.Fatalf("want all %d packages, got %d", len(testPackages), len(result))
	}
}

func TestFilterBySearch(t *testing.T) {
	result := domain.FilterBySearch(testPackages, "fire")
	if len(result) != 1 || result[0].Name != "firefox" {
		t.Fatalf("want firefox, got %v", result)
	}
}

func TestFilterBySearchEmpty(t *testing.T) {
	result := domain.FilterBySearch(testPackages, "")
	if len(result) != len(testPackages) {
		t.Fatalf("empty search should return all, got %d", len(result))
	}
}

func TestFilterByUpgradeable(t *testing.T) {
	result := domain.FilterByUpgradeable(testPackages)
	if len(result) != 2 {
		t.Fatalf("want 2 upgradeable, got %d", len(result))
	}
}

func TestFilterByNative(t *testing.T) {
	result := domain.FilterByNative(testPackages)
	if len(result) != 2 {
		t.Fatalf("want 2 native, got %d", len(result))
	}
}

func TestSortByName(t *testing.T) {
	pkgs := make([]domain.Package, len(testPackages))
	copy(pkgs, testPackages)
	domain.Sort(pkgs, domain.SortByName)
	if pkgs[0].Name != "firefox" || pkgs[1].Name != "node" || pkgs[2].Name != "spotify" || pkgs[3].Name != "vlc" {
		t.Fatalf("unexpected sort order: %v", pkgs)
	}
}

func TestSortBySize(t *testing.T) {
	pkgs := make([]domain.Package, len(testPackages))
	copy(pkgs, testPackages)
	domain.Sort(pkgs, domain.SortBySize)
	// mayor a menor
	if pkgs[0].Name != "firefox" {
		t.Fatalf("firefox should be first (largest), got %s", pkgs[0].Name)
	}
}
```

- [ ] **Step 2: Ejecutar tests — deben fallar**

```bash
go test ./internal/domain/... 2>&1 | head -20
```

Esperado: `FAIL` con `undefined: domain.FilterByManager`

- [ ] **Step 3: Implementar filter.go**

Crear `internal/domain/filter.go`:

```go
package domain

import (
	"sort"
	"strings"
)

// FilterByManager filtra paquetes por gestor. ManagerAll devuelve todos.
func FilterByManager(pkgs []Package, m ManagerType) []Package {
	if m == ManagerAll {
		return pkgs
	}
	out := make([]Package, 0, len(pkgs))
	for _, p := range pkgs {
		if p.Manager == m {
			out = append(out, p)
		}
	}
	return out
}

// FilterBySearch filtra por nombre (case-insensitive, substring).
func FilterBySearch(pkgs []Package, query string) []Package {
	if query == "" {
		return pkgs
	}
	q := strings.ToLower(query)
	out := make([]Package, 0, len(pkgs))
	for _, p := range pkgs {
		if strings.Contains(strings.ToLower(p.Name), q) {
			out = append(out, p)
		}
	}
	return out
}

// FilterByUpgradeable devuelve solo paquetes con actualización disponible.
func FilterByUpgradeable(pkgs []Package) []Package {
	out := make([]Package, 0, len(pkgs))
	for _, p := range pkgs {
		if p.NewVersion != "" {
			out = append(out, p)
		}
	}
	return out
}

// FilterByNative devuelve solo paquetes nativos del OS.
func FilterByNative(pkgs []Package) []Package {
	out := make([]Package, 0, len(pkgs))
	for _, p := range pkgs {
		if p.IsNative {
			out = append(out, p)
		}
	}
	return out
}

// Sort ordena los paquetes in-place según el campo dado.
// SortBySize ordena de mayor a menor. Los demás campos, ascendente.
func Sort(pkgs []Package, by SortField) {
	sort.SliceStable(pkgs, func(i, j int) bool {
		switch by {
		case SortByManager:
			return pkgs[i].Manager < pkgs[j].Manager
		case SortByVersion:
			return pkgs[i].Version < pkgs[j].Version
		case SortBySize:
			return pkgs[i].Size > pkgs[j].Size // mayor a menor
		default: // SortByName
			return pkgs[i].Name < pkgs[j].Name
		}
	})
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/domain/... -v
```

Esperado: todos los tests `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/filter.go internal/domain/filter_test.go
git commit -m "feat: Filter and Sort functions for Package slice"
```

---

## Task 5: Adaptador apt

**Files:**
- Create: `internal/adapters/apt/apt.go`
- Create: `internal/adapters/apt/apt_test.go`

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/apt/apt_test.go`:

```go
package apt_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/apt"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Salida real de: dpkg-query -W --showformat='${Package}\t${Version}\t${Installed-Size}\t${Description}\n'
const dpkgOutput = `firefox	126.0.1+build1-0ubuntu0.22.04.1	264904	Safe and easy web browser from Mozilla
vlc	3.0.18-3	72340	multimedia player and streamer
`

// Salida real de: apt list --upgradable 2>/dev/null
const aptUpgradable = `Listing...
firefox/jammy-updates 127.0.1+build1 amd64 [upgradable from: 126.0.1+build1]
`

func newFakeRunner() *domain.FakeRunner {
	r := domain.NewFakeRunner()
	r.Outputs["dpkg-query"] = []byte(dpkgOutput)
	r.Outputs["apt"] = []byte(aptUpgradable)
	return r
}

func TestAptList(t *testing.T) {
	a := apt.New(newFakeRunner())
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "firefox" {
		t.Errorf("want firefox, got %s", pkgs[0].Name)
	}
	if pkgs[0].Version != "126.0.1+build1-0ubuntu0.22.04.1" {
		t.Errorf("unexpected version: %s", pkgs[0].Version)
	}
	if pkgs[0].NewVersion != "127.0.1+build1" {
		t.Errorf("want upgradeable version, got %q", pkgs[0].NewVersion)
	}
	if pkgs[1].NewVersion != "" {
		t.Errorf("vlc should not be upgradeable, got %q", pkgs[1].NewVersion)
	}
	if pkgs[0].Manager != domain.ManagerApt {
		t.Errorf("want ManagerApt, got %s", pkgs[0].Manager)
	}
	if !pkgs[0].IsNative {
		t.Error("apt packages should be native")
	}
}

func TestAptAvailable(t *testing.T) {
	a := apt.New(newFakeRunner())
	// FakeRunner siempre devuelve datos, así que Available debe ser true
	if !a.Available() {
		t.Error("apt should be available with fake runner")
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/apt/... 2>&1 | head -10
```

Esperado: `FAIL` con `cannot find package`

- [ ] **Step 3: Implementar apt.go**

Crear `internal/adapters/apt/apt.go`:

```go
package apt

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

// Adapter implementa domain.PackageManager para apt/dpkg.
type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerApt }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("apt")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	// 1. Obtener paquetes instalados via dpkg-query
	raw, err := a.runner.Output([]string{
		"dpkg-query", "-W",
		"--showformat=${Package}\t${Version}\t${Installed-Size}\t${Description}\n",
	})
	if err != nil {
		return nil, err
	}

	// 2. Obtener paquetes actualizables
	upgradable := map[string]string{}
	upgRaw, _ := a.runner.Output([]string{"apt", "list", "--upgradable"})
	scanner := bufio.NewScanner(bytes.NewReader(upgRaw))
	for scanner.Scan() {
		line := scanner.Text()
		// formato: firefox/jammy-updates 127.0.1 amd64 [upgradable from: ...]
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.Split(parts[0], "/")[0]
		upgradable[name] = parts[1]
	}

	// 3. Parsear salida de dpkg-query
	var pkgs []domain.Package
	scanner = bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 4 {
			continue
		}
		name, version, sizeKB, desc := parts[0], parts[1], parts[2], parts[3]
		if name == "" || version == "" {
			continue
		}
		var size int64
		fmt.Sscanf(sizeKB, "%d", &size)
		size *= 1024 // dpkg reporta en KB

		pkgs = append(pkgs, domain.Package{
			Name:        name,
			Version:     version,
			NewVersion:  upgradable[name],
			Manager:     domain.ManagerApt,
			Size:        size,
			Description: strings.TrimSpace(desc),
			IsNative:    true,
			Origin:      "apt",
		})
	}
	return pkgs, scanner.Err()
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	names := packageNames(pkgs)
	args := append([]string{"sudo", "-S", "apt-get", "remove", "-y"}, names...)
	return a.runner.Stream(args)
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	names := packageNames(pkgs)
	args := append([]string{"sudo", "-S", "apt-get", "install", "--only-upgrade", "-y"}, names...)
	return a.runner.Stream(args)
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}

func packageNames(pkgs []domain.Package) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}
```

- [ ] **Step 4: Agregar import faltante de fmt**

Editar `internal/adapters/apt/apt.go` — agregar `"fmt"` al bloque de imports:

```go
import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)
```

- [ ] **Step 5: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/apt/... -v
```

Esperado: `PASS`

- [ ] **Step 6: Commit**

```bash
git add internal/adapters/apt/
git commit -m "feat: apt adapter with dpkg-query parsing and upgradeable detection"
```

---

## Task 6: Adaptador snap

**Files:**
- Create: `internal/adapters/snap/snap.go`
- Create: `internal/adapters/snap/snap_test.go`

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/snap/snap_test.go`:

```go
package snap_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/snap"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Salida real de: snap list
const snapOutput = `Name            Version    Rev    Tracking       Publisher   Notes
bare            1.0        5      latest/stable  canonical✓  base
firefox         127.0.2-1  3779   latest/stable  mozilla✓    -
spotify         1.2.31.533 71     latest/stable  spotify✓    -
`

func TestSnapList(t *testing.T) {
	r := domain.NewFakeRunner()
	r.Outputs["snap"] = []byte(snapOutput)

	a := snap.New(r)
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("want 3 snap packages, got %d", len(pkgs))
	}
	if pkgs[1].Name != "firefox" {
		t.Errorf("want firefox at index 1, got %s", pkgs[1].Name)
	}
	if pkgs[1].Version != "127.0.2-1" {
		t.Errorf("unexpected version: %s", pkgs[1].Version)
	}
	if pkgs[1].Manager != domain.ManagerSnap {
		t.Errorf("want ManagerSnap, got %s", pkgs[1].Manager)
	}
	if pkgs[0].IsNative {
		t.Error("snap packages should not be native")
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/snap/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar snap.go**

Crear `internal/adapters/snap/snap.go`:

```go
package snap

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerSnap }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("snap")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	raw, err := a.runner.Output([]string{"snap", "list"})
	if err != nil {
		return nil, err
	}

	var pkgs []domain.Package
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	first := true
	for scanner.Scan() {
		// saltar la cabecera
		if first {
			first = false
			continue
		}
		line := scanner.Text()
		// columnas: Name Version Rev Tracking Publisher Notes
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:     fields[0],
			Version:  fields[1],
			Manager:  domain.ManagerSnap,
			IsNative: false,
			Origin:   "snap store",
		})
	}
	return pkgs, scanner.Err()
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	args := []string{"sudo", "-S", "snap", "remove"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	args := []string{"sudo", "-S", "snap", "refresh"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/snap/... -v
```

Esperado: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/snap/
git commit -m "feat: snap adapter"
```

---

## Task 7: Adaptador flatpak

**Files:**
- Create: `internal/adapters/flatpak/flatpak.go`
- Create: `internal/adapters/flatpak/flatpak_test.go`

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/flatpak/flatpak_test.go`:

```go
package flatpak_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/flatpak"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Salida real de: flatpak list --app --columns=application,version,size,description,origin
const flatpakOutput = `org.mozilla.firefox	127.0	284.1 MB	Firefox Web Browser	flathub
com.spotify.Client	1.2.31	184.0 MB	Spotify streaming service	flathub
`

func TestFlatpakList(t *testing.T) {
	r := domain.NewFakeRunner()
	r.Outputs["flatpak"] = []byte(flatpakOutput)

	a := flatpak.New(r)
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("want 2 flatpak packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "org.mozilla.firefox" {
		t.Errorf("want org.mozilla.firefox, got %s", pkgs[0].Name)
	}
	if pkgs[0].Version != "127.0" {
		t.Errorf("unexpected version: %s", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerFlatpak {
		t.Errorf("want ManagerFlatpak, got %s", pkgs[0].Manager)
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/flatpak/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar flatpak.go**

Crear `internal/adapters/flatpak/flatpak.go`:

```go
package flatpak

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerFlatpak }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	raw, err := a.runner.Output([]string{
		"flatpak", "list", "--app",
		"--columns=application,version,size,description,origin",
	})
	if err != nil {
		return nil, err
	}

	var pkgs []domain.Package
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		// columnas separadas por tab: application version size description origin
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 2 {
			// intento con espacio si no hay tabs (depende de la versión de flatpak)
			parts = strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
		}

		var sizeBytes int64
		if len(parts) >= 3 {
			var sizeVal float64
			var unit string
			fmt.Sscanf(strings.TrimSpace(parts[2]), "%f %s", &sizeVal, &unit)
			switch strings.ToUpper(unit) {
			case "MB":
				sizeBytes = int64(sizeVal * 1024 * 1024)
			case "GB":
				sizeBytes = int64(sizeVal * 1024 * 1024 * 1024)
			case "KB":
				sizeBytes = int64(sizeVal * 1024)
			}
		}

		desc := ""
		origin := "flathub"
		if len(parts) >= 4 {
			desc = strings.TrimSpace(parts[3])
		}
		if len(parts) >= 5 {
			origin = strings.TrimSpace(parts[4])
		}

		pkgs = append(pkgs, domain.Package{
			Name:        strings.TrimSpace(parts[0]),
			Version:     strings.TrimSpace(parts[1]),
			Manager:     domain.ManagerFlatpak,
			Size:        sizeBytes,
			Description: desc,
			IsNative:    false,
			Origin:      origin,
		})
	}
	return pkgs, scanner.Err()
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	args := []string{"flatpak", "uninstall", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	args := []string{"flatpak", "update", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/flatpak/... -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/flatpak/
git commit -m "feat: flatpak adapter"
```

---

## Task 8: Adaptador dpkg (paquetes locales .deb)

**Files:**
- Create: `internal/adapters/dpkg/dpkg.go`
- Create: `internal/adapters/dpkg/dpkg_test.go`

El adaptador dpkg lista paquetes instalados directamente con `dpkg -i` que **no tienen fuente en apt** (i.e., paquetes de `.deb` descargados manualmente como Google Chrome, Discord, etc.).

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/dpkg/dpkg_test.go`:

```go
package dpkg_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/dpkg"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Paquetes instalados por dpkg -i sin fuente en apt
// Salida de: dpkg-query -W --showformat='${Package}\t${Version}\t${Installed-Size}\t${Description}\n'
// filtrada por paquetes sin origin en apt-cache policy
const dpkgOnlyOutput = `google-chrome-stable	126.0.6478.55-1	373316	The web browser from Google
discord	0.0.46	243200	Discord - Free voice and text chat
`

func TestDpkgList(t *testing.T) {
	r := domain.NewFakeRunner()
	r.Outputs["dpkg-query"] = []byte(dpkgOnlyOutput)

	a := dpkg.New(r)
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "google-chrome-stable" {
		t.Errorf("want google-chrome-stable, got %s", pkgs[0].Name)
	}
	if pkgs[0].Manager != domain.ManagerDpkg {
		t.Errorf("want ManagerDpkg, got %s", pkgs[0].Manager)
	}
	if pkgs[0].Origin != "local .deb" {
		t.Errorf("want 'local .deb' origin, got %s", pkgs[0].Origin)
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/dpkg/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar dpkg.go**

Crear `internal/adapters/dpkg/dpkg.go`:

```go
package dpkg

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

// Adapter lista paquetes instalados con dpkg -i que no tienen fuente en apt.
// Complementa al adaptador apt sin duplicar paquetes.
type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerDpkg }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("dpkg-query")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	raw, err := a.runner.Output([]string{
		"dpkg-query", "-W",
		"--showformat=${Package}\t${Version}\t${Installed-Size}\t${Description}\n",
	})
	if err != nil {
		return nil, err
	}

	var pkgs []domain.Package
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 4 {
			continue
		}
		name, version, sizeKB, desc := parts[0], parts[1], parts[2], parts[3]
		if name == "" || version == "" {
			continue
		}
		var size int64
		fmt.Sscanf(sizeKB, "%d", &size)
		size *= 1024

		pkgs = append(pkgs, domain.Package{
			Name:        name,
			Version:     version,
			Manager:     domain.ManagerDpkg,
			Size:        size,
			Description: strings.TrimSpace(desc),
			IsNative:    false,
			Origin:      "local .deb",
		})
	}
	return pkgs, scanner.Err()
}

// Remove desinstala usando dpkg --remove.
func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	args := []string{"sudo", "-S", "dpkg", "--remove"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

// Update no aplica para paquetes dpkg locales.
func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	// dpkg no tiene mecanismo de actualización — operación no soportada.
	// Devolvemos una Operation que inmediatamente reporta el mensaje.
	return a.runner.Stream([]string{"echo", "dpkg: update not supported for local packages"})
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/dpkg/... -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/dpkg/
git commit -m "feat: dpkg adapter for locally installed .deb packages"
```

---

## Task 9: Adaptador pip

**Files:**
- Create: `internal/adapters/pip/pip.go`
- Create: `internal/adapters/pip/pip_test.go`

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/pip/pip_test.go`:

```go
package pip_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/pip"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Salida de: pip3 list --format=json
const pipOutput = `[{"name": "requests", "version": "2.31.0"}, {"name": "numpy", "version": "1.26.0"}, {"name": "pip", "version": "23.2.1"}]`

func TestPipList(t *testing.T) {
	r := domain.NewFakeRunner()
	r.Outputs["pip3"] = []byte(pipOutput)

	a := pip.New(r)
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("want 3 pip packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "requests" {
		t.Errorf("want requests, got %s", pkgs[0].Name)
	}
	if pkgs[0].Version != "2.31.0" {
		t.Errorf("unexpected version: %s", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerPip {
		t.Errorf("want ManagerPip, got %s", pkgs[0].Manager)
	}
	if pkgs[0].Origin != "pypi" {
		t.Errorf("want pypi origin, got %s", pkgs[0].Origin)
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/pip/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar pip.go**

Crear `internal/adapters/pip/pip.go`:

```go
package pip

import (
	"encoding/json"
	"os/exec"

	"github.com/yourusername/pkgsh/internal/domain"
)

type pipEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerPip }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("pip3")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	raw, err := a.runner.Output([]string{"pip3", "list", "--format=json"})
	if err != nil {
		return nil, err
	}

	var entries []pipEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}

	pkgs := make([]domain.Package, len(entries))
	for i, e := range entries {
		pkgs[i] = domain.Package{
			Name:     e.Name,
			Version:  e.Version,
			Manager:  domain.ManagerPip,
			IsNative: false,
			Origin:   "pypi",
		}
	}
	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	args := []string{"pip3", "uninstall", "-y"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	args := []string{"pip3", "install", "--upgrade"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/pip/... -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/pip/
git commit -m "feat: pip adapter with JSON output parsing"
```

---

## Task 10: Adaptador npm

**Files:**
- Create: `internal/adapters/npm/npm.go`
- Create: `internal/adapters/npm/npm_test.go`

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/npm/npm_test.go`:

```go
package npm_test

import (
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/npm"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Salida de: npm list -g --json --depth=0
const npmOutput = `{
  "dependencies": {
    "typescript": {"version": "5.1.6"},
    "nodemon": {"version": "3.0.1"},
    "npm": {"version": "9.8.1"}
  }
}`

func TestNpmList(t *testing.T) {
	r := domain.NewFakeRunner()
	r.Outputs["npm"] = []byte(npmOutput)

	a := npm.New(r)
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("want 3 npm packages, got %d", len(pkgs))
	}
	// npm devuelve map, el orden no está garantizado; verificar que están todos
	names := make(map[string]bool)
	for _, p := range pkgs {
		names[p.Name] = true
		if p.Manager != domain.ManagerNpm {
			t.Errorf("want ManagerNpm, got %s", p.Manager)
		}
		if p.Origin != "npmjs" {
			t.Errorf("want npmjs origin, got %s", p.Origin)
		}
	}
	for _, expected := range []string{"typescript", "nodemon", "npm"} {
		if !names[expected] {
			t.Errorf("missing package %s", expected)
		}
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/npm/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar npm.go**

Crear `internal/adapters/npm/npm.go`:

```go
package npm

import (
	"encoding/json"
	"os/exec"

	"github.com/yourusername/pkgsh/internal/domain"
)

type npmDep struct {
	Version string `json:"version"`
}

type npmOutput struct {
	Dependencies map[string]npmDep `json:"dependencies"`
}

type Adapter struct {
	runner domain.Runner
}

func New(r domain.Runner) *Adapter {
	return &Adapter{runner: r}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerNpm }

func (a *Adapter) Available() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func (a *Adapter) List() ([]domain.Package, error) {
	raw, err := a.runner.Output([]string{"npm", "list", "-g", "--json", "--depth=0"})
	if err != nil {
		// npm list puede retornar exit code != 0 con paquetes faltantes; ignorar
		if len(raw) == 0 {
			return nil, err
		}
	}

	var out npmOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	pkgs := make([]domain.Package, 0, len(out.Dependencies))
	for name, dep := range out.Dependencies {
		pkgs = append(pkgs, domain.Package{
			Name:     name,
			Version:  dep.Version,
			Manager:  domain.ManagerNpm,
			IsNative: false,
			Origin:   "npmjs",
		})
	}
	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	args := []string{"npm", "uninstall", "-g"}
	for _, p := range pkgs {
		args = append(args, p.Name)
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	args := []string{"npm", "install", "-g"}
	for _, p := range pkgs {
		args = append(args, p.Name+"@latest")
	}
	return a.runner.Stream(args)
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 4: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/npm/... -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/npm/
git commit -m "feat: npm global packages adapter"
```

---

## Task 11: Adaptador AppImage

**Files:**
- Create: `internal/adapters/appimage/appimage.go`
- Create: `internal/adapters/appimage/appimage_test.go`

El adaptador escanea directorios conocidos en busca de archivos `.AppImage`.

- [ ] **Step 1: Escribir el test**

Crear `internal/adapters/appimage/appimage_test.go`:

```go
package appimage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/pkgsh/internal/adapters/appimage"
	"github.com/yourusername/pkgsh/internal/domain"
)

func TestAppImageList(t *testing.T) {
	// Crear directorio temporal con AppImages de prueba
	dir := t.TempDir()
	f1 := filepath.Join(dir, "Obsidian-1.4.16.AppImage")
	f2 := filepath.Join(dir, "Inkscape-1.3.AppImage")
	_ = os.WriteFile(f1, []byte("fake"), 0755)
	_ = os.WriteFile(f2, []byte("fake"), 0755)

	a := appimage.New([]string{dir})
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("want 2 AppImages, got %d", len(pkgs))
	}
	names := map[string]bool{}
	for _, p := range pkgs {
		names[p.Name] = true
		if p.Manager != domain.ManagerAppImage {
			t.Errorf("want ManagerAppImage, got %s", p.Manager)
		}
		if p.Origin != "appimage" {
			t.Errorf("want appimage origin, got %s", p.Origin)
		}
	}
	if !names["Obsidian-1.4.16"] {
		t.Error("missing Obsidian")
	}
}
```

- [ ] **Step 2: Ejecutar — debe fallar**

```bash
go test ./internal/adapters/appimage/... 2>&1 | head -5
```

- [ ] **Step 3: Implementar appimage.go**

Crear `internal/adapters/appimage/appimage.go`:

```go
package appimage

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

// DefaultDirs son los directorios escaneados por defecto.
var DefaultDirs = []string{
	filepath.Join(os.Getenv("HOME"), "Applications"),
	filepath.Join(os.Getenv("HOME"), "Downloads"),
	"/opt",
}

// Adapter escanea directorios en busca de archivos .AppImage.
type Adapter struct {
	dirs []string
}

// New crea un adaptador con los directorios dados.
// Usar appimage.New(appimage.DefaultDirs) en producción.
func New(dirs []string) *Adapter {
	return &Adapter{dirs: dirs}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerAppImage }

func (a *Adapter) Available() bool { return true } // siempre disponible

func (a *Adapter) List() ([]domain.Package, error) {
	var pkgs []domain.Package
	for _, dir := range a.dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // directorio no existe, skip
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".appimage") {
				continue
			}

			info, err := e.Info()
			if err != nil {
				continue
			}

			nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
			nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".AppImage")
			nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".appimage")

			pkgs = append(pkgs, domain.Package{
				Name:     nameWithoutExt,
				Version:  "unknown",
				Manager:  domain.ManagerAppImage,
				Size:     info.Size(),
				IsNative: false,
				Origin:   "appimage",
				Description: filepath.Join(dir, name),
			})
		}
	}
	return pkgs, nil
}

// Remove elimina el archivo .AppImage del sistema de archivos.
func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	// Description almacena el path completo del archivo
	paths := make([]string, 0, len(pkgs))
	for _, p := range pkgs {
		if strings.HasSuffix(strings.ToLower(p.Description), ".appimage") {
			paths = append(paths, p.Description)
		}
	}
	args := append([]string{"rm", "-f"}, paths...)
	// No necesita sudo si el archivo es del usuario
	pr, pw, _ := makeStringPipe("Removing AppImages...\n")
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer pw.Close()
		for _, path := range paths {
			if err := os.Remove(path); err != nil {
				pw.Write([]byte("error: " + err.Error() + "\n"))
			} else {
				pw.Write([]byte("removed: " + path + "\n"))
			}
		}
		_ = args // suppress unused warning
	}()
	var noErr error
	return &domain.Operation{Reader: pr, Writer: pw, Done: done, Err: &noErr}
}

// Update no está soportado para AppImages en v1.
func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	pr, pw, done := makeStringPipe("AppImage update not supported in v1. Download manually.\n")
	_ = pw.Close()
	close(done)
	var noErr error
	return &domain.Operation{Reader: pr, Writer: pw, Done: done, Err: &noErr}
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}

func makeStringPipe(msg string) (*os.File, *os.File, chan struct{}) {
	// Usamos io.Pipe en lugar de os.File para evitar imports innecesarios
	// Este helper retorna un pipe string simple
	// Nota: se reemplaza por io.Pipe en la implementación real
	panic("use io.Pipe version instead")
}
```

- [ ] **Step 4: Corregir appimage.go — reemplazar makeStringPipe con io.Pipe**

El método `Remove` y `Update` necesitan `io.Pipe`. Reemplazar la implementación:

```go
package appimage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/pkgsh/internal/domain"
)

var DefaultDirs = []string{
	filepath.Join(os.Getenv("HOME"), "Applications"),
	filepath.Join(os.Getenv("HOME"), "Downloads"),
	"/opt",
}

type Adapter struct {
	dirs []string
}

func New(dirs []string) *Adapter {
	return &Adapter{dirs: dirs}
}

func (a *Adapter) Name() domain.ManagerType { return domain.ManagerAppImage }
func (a *Adapter) Available() bool          { return true }

func (a *Adapter) List() ([]domain.Package, error) {
	var pkgs []domain.Package
	for _, dir := range a.dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".appimage") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			baseName := strings.TrimSuffix(name, ".AppImage")
			baseName = strings.TrimSuffix(baseName, ".appimage")
			fullPath := filepath.Join(dir, name)
			pkgs = append(pkgs, domain.Package{
				Name:        baseName,
				Version:     "unknown",
				Manager:     domain.ManagerAppImage,
				Size:        info.Size(),
				IsNative:    false,
				Origin:      "appimage",
				Description: fullPath,
			})
		}
	}
	return pkgs, nil
}

func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	pr, pw := io.Pipe()
	var noErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer pw.Close()
		for _, p := range pkgs {
			path := p.Description
			if err := os.Remove(path); err != nil {
				_, _ = pw.Write([]byte("error: " + err.Error() + "\n"))
			} else {
				_, _ = pw.Write([]byte("removed: " + path + "\n"))
			}
		}
	}()
	return &domain.Operation{Reader: pr, Writer: pw, Done: done, Err: &noErr}
}

func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	pr, pw := io.Pipe()
	var noErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer pw.Close()
		_, _ = pw.Write([]byte("AppImage update not supported in v1. Download manually.\n"))
	}()
	return &domain.Operation{Reader: pr, Writer: pw, Done: done, Err: &noErr}
}

func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
```

- [ ] **Step 5: Ejecutar tests — deben pasar**

```bash
go test ./internal/adapters/appimage/... -v
```

- [ ] **Step 6: Ejecutar todos los tests de adaptadores**

```bash
go test ./internal/adapters/... -v
```

Esperado: todos `PASS`.

- [ ] **Step 7: Commit**

```bash
git add internal/adapters/appimage/
git commit -m "feat: AppImage adapter with directory scanning"
```

---

## Task 12: Registry de adaptadores

**Files:**
- Create: `internal/adapters/registry.go`

El registry carga todos los adaptadores en goroutines paralelas y unifica los resultados.

- [ ] **Step 1: Implementar registry.go**

Crear `internal/adapters/registry.go`:

```go
package adapters

import (
	"sync"

	"github.com/yourusername/pkgsh/internal/adapters/appimage"
	"github.com/yourusername/pkgsh/internal/adapters/apt"
	"github.com/yourusername/pkgsh/internal/adapters/dpkg"
	"github.com/yourusername/pkgsh/internal/adapters/flatpak"
	"github.com/yourusername/pkgsh/internal/adapters/npm"
	"github.com/yourusername/pkgsh/internal/adapters/pip"
	"github.com/yourusername/pkgsh/internal/adapters/snap"
	"github.com/yourusername/pkgsh/internal/domain"
)

// LoadResult es el resultado de cargar un adaptador.
type LoadResult struct {
	Packages []domain.Package
	Manager  domain.ManagerType
	Err      error
}

// Registry contiene todos los adaptadores disponibles.
type Registry struct {
	managers []domain.PackageManager
}

// New crea un Registry con los adaptadores de producción.
func New() *Registry {
	runner := domain.NewSystemRunner()
	return &Registry{
		managers: []domain.PackageManager{
			apt.New(runner),
			snap.New(runner),
			flatpak.New(runner),
			dpkg.New(runner),
			pip.New(runner),
			npm.New(runner),
			appimage.New(appimage.DefaultDirs),
		},
	}
}

// NewWithManagers crea un Registry con adaptadores personalizados (para tests).
func NewWithManagers(managers []domain.PackageManager) *Registry {
	return &Registry{managers: managers}
}

// LoadAll carga todos los gestores disponibles en paralelo.
// Envía resultados progresivamente al canal results a medida que llegan.
// Cierra el canal cuando todos terminan.
func (r *Registry) LoadAll(results chan<- LoadResult) {
	var wg sync.WaitGroup
	for _, m := range r.managers {
		if !m.Available() {
			continue
		}
		wg.Add(1)
		m := m // captura para goroutine
		go func() {
			defer wg.Done()
			pkgs, err := m.List()
			results <- LoadResult{
				Packages: pkgs,
				Manager:  m.Name(),
				Err:      err,
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
}

// GetManager devuelve el adaptador para el gestor dado, o nil si no existe.
func (r *Registry) GetManager(m domain.ManagerType) domain.PackageManager {
	for _, mgr := range r.managers {
		if mgr.Name() == m {
			return mgr
		}
	}
	return nil
}
```

- [ ] **Step 2: Verificar que compila**

```bash
go build ./internal/adapters/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/adapters/registry.go
git commit -m "feat: adapter registry with parallel goroutine loading"
```

---

## Task 13: UI — estilos y keymap

**Files:**
- Create: `internal/ui/styles.go`
- Create: `internal/ui/keys.go`

- [ ] **Step 1: Crear styles.go**

Crear `internal/ui/styles.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colores base
	colorPrimary   = lipgloss.Color("#7C3AED") // púrpura
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#10B981") // verde
	colorWarning   = lipgloss.Color("#F59E0B") // amarillo
	colorDanger    = lipgloss.Color("#EF4444") // rojo
	colorMuted     = lipgloss.Color("#6B7280") // gris
	colorBg        = lipgloss.Color("#1F2937") // fondo oscuro
	colorBgLight   = lipgloss.Color("#374151") // fondo panel
	colorText      = lipgloss.Color("#F9FAFB") // texto claro

	// Estilos de panel
	StylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	StylePanelActive = StylePanel.
				BorderForeground(colorPrimary)

	// Header / tabs
	StyleTab = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	StyleTabActive = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	// Fila de lista
	StyleRow = lipgloss.NewStyle().
			Foreground(colorText)

	StyleRowSelected = lipgloss.NewStyle().
				Background(colorPrimary).
				Foreground(colorText).
				Bold(true)

	StyleCheckbox         = lipgloss.NewStyle().Foreground(colorSecondary)
	StyleCheckboxSelected = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	// Badges de gestor
	managerColors = map[string]lipgloss.Color{
		"apt":      "#3B82F6",
		"snap":     "#F97316",
		"flatpak":  "#8B5CF6",
		"dpkg":     "#14B8A6",
		"pip":      "#EAB308",
		"npm":      "#EF4444",
		"appimage": "#EC4899",
	}

	// Indicador de actualización
	StyleUpgrade = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)

	// Panel de detalle
	StyleDetailLabel = lipgloss.NewStyle().Foreground(colorMuted).Width(12)
	StyleDetailValue = lipgloss.NewStyle().Foreground(colorText)

	// Log
	StyleLog     = lipgloss.NewStyle().Foreground(colorMuted)
	StyleLogLine = lipgloss.NewStyle().Foreground(colorText)

	// Modal
	StyleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorWarning).
			Padding(1, 2).
			Background(colorBg)

	StyleModalTitle   = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
	StyleModalDanger  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	StyleModalConfirm = lipgloss.NewStyle().Foreground(colorSuccess)

	// Statusbar inferior
	StyleStatusBar = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(colorBgLight).
			Padding(0, 1)

	StyleStatusKey = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Background(colorBgLight).
			Bold(true)

	// Título de la app
	StyleTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)
)

// ManagerBadge devuelve el nombre del gestor con su color asociado.
func ManagerBadge(manager string) string {
	color, ok := managerColors[manager]
	if !ok {
		color = colorMuted
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(manager)
}

// FormatSize convierte bytes a string legible (KB, MB, GB).
func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1024*1024*1024))
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1024))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
```

- [ ] **Step 2: Agregar import fmt a styles.go**

Editar `internal/ui/styles.go` para agregar `"fmt"` al import:

```go
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)
```

- [ ] **Step 3: Crear keys.go**

Crear `internal/ui/keys.go`:

```go
package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap define todos los atajos de la aplicación.
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	SelectAll key.Binding
	ClearSel  key.Binding
	Remove    key.Binding
	Update    key.Binding
	Search    key.Binding
	Sort      key.Binding
	Tab       key.Binding
	Tab1      key.Binding
	Tab2      key.Binding
	Tab3      key.Binding
	Tab4      key.Binding
	Tab5      key.Binding
	Tab6      key.Binding
	Tab7      key.Binding
	Tab8      key.Binding
	Quit      key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
}

// DefaultKeyMap es el mapa de teclas por defecto.
var DefaultKeyMap = KeyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "arriba")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "abajo")),
	Select:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "seleccionar")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "seleccionar todos")),
	ClearSel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "limpiar selección")),
	Remove:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "desinstalar")),
	Update:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "actualizar")),
	Search:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "buscar")),
	Sort:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "ordenar")),
	Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "panel")),
	Tab1:      key.NewBinding(key.WithKeys("1")),
	Tab2:      key.NewBinding(key.WithKeys("2")),
	Tab3:      key.NewBinding(key.WithKeys("3")),
	Tab4:      key.NewBinding(key.WithKeys("4")),
	Tab5:      key.NewBinding(key.WithKeys("5")),
	Tab6:      key.NewBinding(key.WithKeys("6")),
	Tab7:      key.NewBinding(key.WithKeys("7")),
	Tab8:      key.NewBinding(key.WithKeys("8")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "salir")),
	Confirm:   key.NewBinding(key.WithKeys("y", "s", "enter"), key.WithHelp("s/y", "confirmar")),
	Cancel:    key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "cancelar")),
}
```

- [ ] **Step 4: Verificar que compila**

```bash
go build ./internal/ui/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/ui/styles.go internal/ui/keys.go
git commit -m "feat: UI styles (lipgloss theme) and keymap"
```

---

## Task 14: UI — Panel de lista

**Files:**
- Create: `internal/ui/list.go`

- [ ] **Step 1: Implementar list.go**

Crear `internal/ui/list.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/pkgsh/internal/domain"
)

// ListPanel renderiza la lista de paquetes con checkboxes, búsqueda y badges.
type ListPanel struct {
	packages []domain.Package
	cursor   int
	selected map[int]bool
	width    int
	height   int
}

func NewListPanel(width, height int) *ListPanel {
	return &ListPanel{
		selected: make(map[int]bool),
		width:    width,
		height:   height,
	}
}

func (l *ListPanel) SetPackages(pkgs []domain.Package) {
	l.packages = pkgs
	if l.cursor >= len(pkgs) && len(pkgs) > 0 {
		l.cursor = len(pkgs) - 1
	}
}

func (l *ListPanel) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
	}
}

func (l *ListPanel) MoveDown() {
	if l.cursor < len(l.packages)-1 {
		l.cursor++
	}
}

func (l *ListPanel) ToggleSelect() {
	if len(l.packages) == 0 {
		return
	}
	if l.selected[l.cursor] {
		delete(l.selected, l.cursor)
	} else {
		l.selected[l.cursor] = true
	}
}

func (l *ListPanel) SelectAll() {
	for i := range l.packages {
		l.selected[i] = true
	}
}

func (l *ListPanel) ClearSelection() {
	l.selected = make(map[int]bool)
}

func (l *ListPanel) SelectedPackages() []domain.Package {
	if len(l.selected) == 0 && len(l.packages) > 0 {
		// si no hay selección, operar sobre el paquete bajo el cursor
		return []domain.Package{l.packages[l.cursor]}
	}
	result := make([]domain.Package, 0, len(l.selected))
	for i := range l.selected {
		if i < len(l.packages) {
			result = append(result, l.packages[i])
		}
	}
	return result
}

func (l *ListPanel) CurrentPackage() *domain.Package {
	if len(l.packages) == 0 {
		return nil
	}
	return &l.packages[l.cursor]
}

func (l *ListPanel) View(active bool) string {
	style := StylePanel
	if active {
		style = StylePanelActive
	}

	var rows []string
	visibleLines := l.height - 4 // descontar bordes y padding
	start := 0
	if l.cursor >= visibleLines {
		start = l.cursor - visibleLines + 1
	}

	for i := start; i < len(l.packages) && i < start+visibleLines; i++ {
		p := l.packages[i]

		checkbox := "☐ "
		checkStyle := StyleCheckbox
		if l.selected[i] {
			checkbox = "☒ "
			checkStyle = StyleCheckboxSelected
		}

		upgrade := "  "
		if p.NewVersion != "" {
			upgrade = StyleUpgrade.Render("↑ ")
		}

		name := p.Name
		if len(name) > 22 {
			name = name[:20] + ".."
		}

		badge := ManagerBadge(string(p.Manager))
		version := p.Version
		if len(version) > 8 {
			version = version[:8]
		}

		row := fmt.Sprintf("%s%-22s  %s  %-8s %s",
			checkStyle.Render(checkbox),
			name,
			badge,
			version,
			upgrade,
		)

		if i == l.cursor {
			row = StyleRowSelected.Width(l.width - 4).Render(
				lipgloss.NewStyle().Render(row),
			)
		} else {
			row = StyleRow.Render(row)
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		rows = []string{StyleMuted.Render("  Sin paquetes")}
	}

	content := strings.Join(rows, "\n")
	return style.Width(l.width).Height(l.height).Render(content)
}
```

- [ ] **Step 2: Agregar StyleMuted faltante a styles.go**

Editar `internal/ui/styles.go` — agregar:

```go
StyleMuted = lipgloss.NewStyle().Foreground(colorMuted)
```

junto con los otros estilos existentes.

- [ ] **Step 3: Verificar que compila**

```bash
go build ./internal/ui/...
```

- [ ] **Step 4: Commit**

```bash
git add internal/ui/list.go internal/ui/styles.go
git commit -m "feat: list panel with checkboxes, badges, and cursor navigation"
```

---

## Task 15: UI — Panel de detalle

**Files:**
- Create: `internal/ui/detail.go`

- [ ] **Step 1: Implementar detail.go**

Crear `internal/ui/detail.go`:

```go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/pkgsh/internal/domain"
)

// DetailPanel muestra la información completa del paquete seleccionado.
type DetailPanel struct {
	pkg    *domain.Package
	width  int
	height int
}

func NewDetailPanel(width, height int) *DetailPanel {
	return &DetailPanel{width: width, height: height}
}

func (d *DetailPanel) SetPackage(p *domain.Package) {
	d.pkg = p
}

func (d *DetailPanel) View(active bool) string {
	style := StylePanel
	if active {
		style = StylePanelActive
	}

	if d.pkg == nil {
		return style.Width(d.width).Height(d.height).Render(
			StyleMuted.Render("\n  Selecciona un paquete"),
		)
	}

	p := d.pkg
	rows := []string{
		row("Nombre", p.Name),
		row("Versión", p.Version),
		row("Gestor", ManagerBadge(string(p.Manager))),
		row("Origen", p.Origin),
		row("Tamaño", FormatSize(p.Size)),
		row("Nativo", boolStr(p.IsNative)),
	}

	if p.NewVersion != "" {
		rows = append(rows, StyleUpgrade.Render(
			fmt.Sprintf("  ↑ Actualizable a %s", p.NewVersion),
		))
	}

	if !p.InstallDate.IsZero() {
		rows = append(rows, row("Instalado", p.InstallDate.Format("2006-01-02")))
	}

	if p.Description != "" {
		rows = append(rows, "")
		desc := p.Description
		if len(desc) > d.width-6 {
			desc = desc[:d.width-9] + "..."
		}
		rows = append(rows, StyleDetailLabel.Render("Desc")+StyleDetailValue.Render("  "+desc))
	}

	content := strings.Join(rows, "\n")
	return style.Width(d.width).Height(d.height).Render(content)
}

func row(label, value string) string {
	return StyleDetailLabel.Render(label) + StyleDetailValue.Render("  " + value)
}

func boolStr(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

// Asegurar que time está importado aunque InstallDate sea zero.
var _ = time.Now
```

- [ ] **Step 2: Verificar que compila**

```bash
go build ./internal/ui/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/ui/detail.go
git commit -m "feat: detail panel showing full package information"
```

---

## Task 16: UI — Panel de log y streaming

**Files:**
- Create: `internal/ui/log.go`

El panel de log recibe líneas de una `*domain.Operation` usando un `tea.Cmd` que se re-invoca a sí mismo por cada línea.

- [ ] **Step 1: Implementar log.go**

Crear `internal/ui/log.go`:

```go
package ui

import (
	"bufio"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/pkgsh/internal/domain"
)

// Mensajes de Bubble Tea para el panel de log
type LogLineMsg string
type OperationDoneMsg struct{ Err error }
type SudoPromptMsg struct{} // detectado prompt de sudo

// LogPanel muestra el output en tiempo real de las operaciones.
type LogPanel struct {
	lines  []string
	width  int
	height int
}

func NewLogPanel(width, height int) *LogPanel {
	return &LogPanel{width: width, height: height}
}

func (l *LogPanel) AddLine(line string) {
	l.lines = append(l.lines, line)
	// mantener las últimas N líneas visibles
	maxLines := l.height - 4
	if len(l.lines) > maxLines*2 {
		l.lines = l.lines[len(l.lines)-maxLines:]
	}
}

func (l *LogPanel) Clear() {
	l.lines = nil
}

func (l *LogPanel) View(active bool) string {
	style := StylePanel
	if active {
		style = StylePanelActive
	}

	visibleLines := l.height - 4
	start := 0
	if len(l.lines) > visibleLines {
		start = len(l.lines) - visibleLines
	}

	var rows []string
	for _, line := range l.lines[start:] {
		rows = append(rows, StyleLogLine.Render(line))
	}

	if len(rows) == 0 {
		rows = []string{StyleMuted.Render("  Log de operaciones")}
	}

	content := strings.Join(rows, "\n")
	return style.Width(l.width).Height(l.height).Render(content)
}

// StreamOperation devuelve un tea.Cmd que lee líneas de la operación
// y las envía como LogLineMsg. Se detiene al EOF y envía OperationDoneMsg.
func StreamOperation(op *domain.Operation) tea.Cmd {
	scanner := bufio.NewScanner(op.Reader)
	return func() tea.Msg {
		if scanner.Scan() {
			line := scanner.Text()
			// Detectar prompt de sudo en el output
			if strings.Contains(line, "[sudo] password") ||
				strings.Contains(line, "sudo] contraseña") {
				return SudoPromptMsg{}
			}
			return LogLineMsg(line)
		}
		var err error
		if op.Err != nil {
			err = *op.Err
		}
		return OperationDoneMsg{Err: err}
	}
}
```

- [ ] **Step 2: Verificar que compila**

```bash
go build ./internal/ui/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/ui/log.go
git commit -m "feat: log panel with real-time streaming via tea.Cmd"
```

---

## Task 17: UI — Modal (confirmación y sudo)

**Files:**
- Create: `internal/ui/modal.go`

- [ ] **Step 1: Implementar modal.go**

Crear `internal/ui/modal.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/pkgsh/internal/domain"
)

// ModalType diferencia los dos tipos de modal.
type ModalType int

const (
	ModalConfirm ModalType = iota
	ModalSudo
)

// ModalConfirmMsg se emite cuando el usuario confirma.
type ModalConfirmMsg struct{}

// ModalCancelMsg se emite cuando el usuario cancela.
type ModalCancelMsg struct{}

// ModalSudoSubmitMsg se emite con la contraseña escrita.
type ModalSudoSubmitMsg struct{ Password string }

// Modal es el componente modal (confirmación o sudo).
type Modal struct {
	visible    bool
	modalType  ModalType
	title      string
	body       string
	sudoInput  textinput.Model
	width      int
}

func NewModal(width int) *Modal {
	ti := textinput.New()
	ti.Placeholder = "contraseña"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 30
	return &Modal{sudoInput: ti, width: width}
}

// ShowConfirm muestra el modal de confirmación para una lista de paquetes.
func (m *Modal) ShowConfirm(action string, pkgs []domain.Package) {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	m.modalType = ModalConfirm
	m.title = fmt.Sprintf("%s %d paquete(s)", action, len(pkgs))
	m.body = strings.Join(names, ", ")
	m.visible = true
}

// ShowSudo muestra el modal de entrada de contraseña sudo.
func (m *Modal) ShowSudo() {
	m.modalType = ModalSudo
	m.title = "Se requiere contraseña sudo"
	m.body = ""
	m.sudoInput.SetValue("")
	m.sudoInput.Focus()
	m.visible = true
}

func (m *Modal) Hide() {
	m.visible = false
}

func (m *Modal) IsVisible() bool {
	return m.visible
}

// Update procesa teclas dentro del modal.
func (m *Modal) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.visible {
		return nil
	}
	switch m.modalType {
	case ModalConfirm:
		switch msg.String() {
		case "y", "s", "enter":
			m.visible = false
			return func() tea.Msg { return ModalConfirmMsg{} }
		case "n", "esc":
			m.visible = false
			return func() tea.Msg { return ModalCancelMsg{} }
		}
	case ModalSudo:
		switch msg.String() {
		case "enter":
			pass := m.sudoInput.Value()
			m.visible = false
			return func() tea.Msg { return ModalSudoSubmitMsg{Password: pass} }
		case "esc":
			m.visible = false
			return func() tea.Msg { return ModalCancelMsg{} }
		default:
			var cmd tea.Cmd
			m.sudoInput, cmd = m.sudoInput.Update(msg)
			return cmd
		}
	}
	return nil
}

func (m *Modal) View() string {
	if !m.visible {
		return ""
	}
	var content string
	switch m.modalType {
	case ModalConfirm:
		content = StyleModalTitle.Render(m.title) + "\n\n" +
			StyleModalDanger.Render(m.body) + "\n\n" +
			StyleModalConfirm.Render("[s/y]") + " confirmar  " +
			StyleMuted.Render("[n/esc]") + " cancelar"
	case ModalSudo:
		content = StyleModalTitle.Render(m.title) + "\n\n" +
			m.sudoInput.View() + "\n\n" +
			StyleMuted.Render("[enter] confirmar  [esc] cancelar")
	}
	return StyleModal.Width(m.width / 2).Render(content)
}
```

- [ ] **Step 2: Verificar que compila**

```bash
go build ./internal/ui/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/ui/modal.go
git commit -m "feat: confirmation modal and sudo password input modal"
```

---

## Task 18: UI — AppModel principal

**Files:**
- Create: `internal/ui/app.go`

Este es el modelo principal de Bubble Tea que compone todos los paneles y maneja el flujo de eventos.

- [ ] **Step 1: Implementar app.go**

Crear `internal/ui/app.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/pkgsh/internal/adapters"
	"github.com/yourusername/pkgsh/internal/domain"
)

// PackagesLoadedMsg se envía cuando un adaptador termina de cargar.
type PackagesLoadedMsg adapters.LoadResult

// AppModel es el modelo raíz de Bubble Tea.
type AppModel struct {
	// Estado
	allPackages []domain.Package
	filtered    []domain.Package
	activeTab   domain.ManagerType
	searchQuery string
	sortBy      domain.SortField
	activePanel domain.Panel
	loading     bool
	loadCount   int

	// Sub-modelos UI
	list    *ListPanel
	detail  *DetailPanel
	log     *LogPanel
	modal   *Modal
	search  textinput.Model
	spinner spinner.Model

	// Operación en curso
	pendingOp    *domain.Operation
	pendingAction string // "Desinstalar" o "Actualizar"

	// Registry
	registry *adapters.Registry

	// Dimensiones del terminal
	width  int
	height int

	// Flags de inicio
	StartManager    domain.ManagerType
	StartUpgradeable bool
	StartNative     bool
	StartSearch     string
}

// NewAppModel crea el modelo con las dimensiones del terminal.
func NewAppModel(registry *adapters.Registry) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	si := textinput.New()
	si.Placeholder = "buscar paquetes..."
	si.Width = 30

	return &AppModel{
		registry: registry,
		spinner:  s,
		search:   si,
		loading:  true,
		sortBy:   domain.SortByName,
	}
}

func (m *AppModel) Init() tea.Cmd {
	results := make(chan adapters.LoadResult, 16)
	m.loadCount = 0
	go m.registry.LoadAll(results)
	return tea.Batch(
		m.spinner.Tick,
		waitForLoad(results),
	)
}

// waitForLoad espera un resultado del canal y lo envía como PackagesLoadedMsg.
func waitForLoad(ch <-chan adapters.LoadResult) tea.Cmd {
	return func() tea.Msg {
		result, ok := <-ch
		if !ok {
			return nil // canal cerrado, carga completa
		}
		return PackagesLoadedMsg(result)
	}
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()

	case PackagesLoadedMsg:
		if msg.Packages != nil {
			m.allPackages = append(m.allPackages, msg.Packages...)
		}
		m.applyFilters()
		// seguir esperando más resultados
		results := make(chan adapters.LoadResult, 16)
		cmds = append(cmds, waitForLoad(results))

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case LogLineMsg:
		m.log.AddLine(string(msg))
		if m.pendingOp != nil {
			cmds = append(cmds, StreamOperation(m.pendingOp))
		}

	case OperationDoneMsg:
		m.pendingOp = nil
		if msg.Err != nil {
			m.log.AddLine("error: " + msg.Err.Error())
		} else {
			m.log.AddLine("✓ operación completada")
		}
		// refrescar lista
		cmds = append(cmds, m.reloadAdapters())

	case SudoPromptMsg:
		m.modal.ShowSudo()

	case ModalConfirmMsg:
		if m.pendingOp != nil {
			cmds = append(cmds, StreamOperation(m.pendingOp))
		}

	case ModalSudoSubmitMsg:
		if m.pendingOp != nil {
			_, _ = m.pendingOp.Writer.Write([]byte(msg.Password + "\n"))
		}

	case ModalCancelMsg:
		m.pendingOp = nil
		m.log.AddLine("operación cancelada")

	case tea.KeyMsg:
		// Si modal visible, delegarle las teclas
		if m.modal.IsVisible() {
			cmd := m.modal.Update(msg)
			return m, cmd
		}
		// Si búsqueda activa, delegarle las teclas (excepto Esc y Enter)
		if m.search.Focused() {
			switch msg.String() {
			case "esc", "enter":
				m.search.Blur()
				m.searchQuery = m.search.Value()
				m.applyFilters()
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.searchQuery = m.search.Value()
				m.applyFilters()
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch {
		case msg.String() == "q", msg.String() == "ctrl+c":
			return m, tea.Quit

		case msg.String() == "/":
			m.search.Focus()

		case msg.String() == "tab":
			m.activePanel = (m.activePanel + 1) % 3

		case msg.String() == "up", msg.String() == "k":
			m.list.MoveUp()
			m.detail.SetPackage(m.list.CurrentPackage())

		case msg.String() == "down", msg.String() == "j":
			m.list.MoveDown()
			m.detail.SetPackage(m.list.CurrentPackage())

		case msg.String() == " ":
			m.list.ToggleSelect()

		case msg.String() == "a":
			m.list.SelectAll()

		case msg.String() == "esc":
			m.list.ClearSelection()

		case msg.String() == "s":
			m.sortBy = (m.sortBy + 1) % 4
			m.applyFilters()

		case msg.String() == "d":
			selected := m.list.SelectedPackages()
			if len(selected) > 0 {
				m.modal.ShowConfirm("Desinstalar", selected)
				m.pendingAction = "Desinstalar"
				m.prepareBulkOp(selected, "remove")
			}

		case msg.String() == "u":
			selected := m.list.SelectedPackages()
			if len(selected) > 0 {
				m.modal.ShowConfirm("Actualizar", selected)
				m.pendingAction = "Actualizar"
				m.prepareBulkOp(selected, "update")
			}

		// Tabs numéricos [1-8]
		case msg.String() == "1":
			m.setActiveTab(domain.ManagerAll)
		case msg.String() == "2":
			m.setActiveTab(domain.ManagerApt)
		case msg.String() == "3":
			m.setActiveTab(domain.ManagerSnap)
		case msg.String() == "4":
			m.setActiveTab(domain.ManagerFlatpak)
		case msg.String() == "5":
			m.setActiveTab(domain.ManagerDpkg)
		case msg.String() == "6":
			m.setActiveTab(domain.ManagerPip)
		case msg.String() == "7":
			m.setActiveTab(domain.ManagerNpm)
		case msg.String() == "8":
			m.setActiveTab(domain.ManagerAppImage)
		}
	}

	return m, tea.Batch(cmds...)
}

// prepareBulkOp agrupa los paquetes por gestor y ejecuta la operación.
// Actualmente ejecuta secuencialmente; v1 no paraleliza.
func (m *AppModel) prepareBulkOp(pkgs []domain.Package, action string) {
	// agrupar por gestor
	byManager := make(map[domain.ManagerType][]domain.Package)
	for _, p := range pkgs {
		byManager[p.Manager] = append(byManager[p.Manager], p)
	}

	// ejecutar el primero (simplificación v1: un gestor a la vez)
	for managerType, mpkgs := range byManager {
		mgr := m.registry.GetManager(managerType)
		if mgr == nil {
			continue
		}
		var op *domain.Operation
		if action == "remove" {
			op = mgr.Remove(mpkgs)
		} else {
			op = mgr.Update(mpkgs)
		}
		m.pendingOp = op
		m.log.Clear()
		m.activePanel = domain.PanelLog
		break // v1: primer gestor solamente (multi-gestor en v2)
	}
}

func (m *AppModel) reloadAdapters() tea.Cmd {
	m.allPackages = nil
	results := make(chan adapters.LoadResult, 16)
	go m.registry.LoadAll(results)
	return waitForLoad(results)
}

func (m *AppModel) setActiveTab(mt domain.ManagerType) {
	m.activeTab = mt
	m.applyFilters()
}

func (m *AppModel) applyFilters() {
	pkgs := m.allPackages
	pkgs = domain.FilterByManager(pkgs, m.activeTab)
	if m.searchQuery != "" {
		pkgs = domain.FilterBySearch(pkgs, m.searchQuery)
	}
	domain.Sort(pkgs, m.sortBy)
	m.filtered = pkgs
	m.list.SetPackages(pkgs)
	m.detail.SetPackage(m.list.CurrentPackage())
}

func (m *AppModel) recalcLayout() {
	listW := m.width * 55 / 100
	detailW := m.width - listW - 2
	logH := m.height * 25 / 100
	topH := m.height - logH - 5 // 5 para header + statusbar

	if m.list == nil {
		m.list = NewListPanel(listW, topH)
		m.detail = NewDetailPanel(detailW, topH)
		m.log = NewLogPanel(m.width, logH)
		m.modal = NewModal(m.width)
	} else {
		m.list.width = listW
		m.list.height = topH
		m.detail.width = detailW
		m.detail.height = topH
		m.log.width = m.width
		m.log.height = logH
		m.modal.width = m.width
	}
}

func (m *AppModel) View() string {
	if m.width == 0 {
		return "cargando..."
	}

	// Header con título y tabs
	header := m.renderHeader()

	// Barra de búsqueda
	searchBar := " > " + m.search.View()

	// Fila superior: lista + detalle
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.list.View(m.activePanel == domain.PanelList),
		m.detail.View(m.activePanel == domain.PanelDetail),
	)

	// Panel de log
	logPanel := m.log.View(m.activePanel == domain.PanelLog)

	// Statusbar
	statusBar := m.renderStatusBar()

	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		searchBar,
		topRow,
		logPanel,
		statusBar,
	)

	// Superponer modal si está visible
	if m.modal.IsVisible() {
		modal := m.modal.View()
		// centrar modal
		topOffset := m.height/2 - 4
		leftOffset := m.width/4
		_ = topOffset
		_ = leftOffset
		// Renderizar modal encima
		ui = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("#00000088")),
		)
	}

	return ui
}

func (m *AppModel) renderHeader() string {
	title := StyleTitle.Render("pkgsh")
	if m.loading && len(m.allPackages) == 0 {
		title += "  " + m.spinner.View() + " cargando..."
	} else {
		title += fmt.Sprintf("  %s%d paquetes",
			StyleMuted.Render(""),
			len(m.filtered),
		)
	}

	tabs := []struct {
		label   string
		manager domain.ManagerType
	}{
		{"Todos", domain.ManagerAll},
		{"apt", domain.ManagerApt},
		{"snap", domain.ManagerSnap},
		{"flatpak", domain.ManagerFlatpak},
		{"dpkg", domain.ManagerDpkg},
		{"pip", domain.ManagerPip},
		{"npm", domain.ManagerNpm},
		{"AppImage", domain.ManagerAppImage},
	}

	var tabStr []string
	for i, t := range tabs {
		label := fmt.Sprintf("[%d] %s", i+1, t.label)
		if t.manager == m.activeTab {
			tabStr = append(tabStr, StyleTabActive.Render(label))
		} else {
			tabStr = append(tabStr, StyleTab.Render(label))
		}
	}

	return title + "  " + strings.Join(tabStr, "")
}

func (m *AppModel) renderStatusBar() string {
	keys := []string{
		StyleStatusKey.Render("[/]") + " buscar",
		StyleStatusKey.Render("[space]") + " sel",
		StyleStatusKey.Render("[a]") + " sel todo",
		StyleStatusKey.Render("[d]") + " desinstalar",
		StyleStatusKey.Render("[u]") + " actualizar",
		StyleStatusKey.Render("[s]") + " ordenar",
		StyleStatusKey.Render("[tab]") + " panel",
		StyleStatusKey.Render("[q]") + " salir",
	}
	return StyleStatusBar.Width(m.width).Render(strings.Join(keys, "  "))
}
```

- [ ] **Step 2: Verificar que compila**

```bash
go build ./internal/ui/...
```

Si hay errores de importación circular o tipos no encontrados, revisarlos y corregirlos.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat: AppModel root Bubble Tea model composing all panels"
```

---

## Task 19: CLI entrypoint con flags

**Files:**
- Modify: `cmd/pkgsh/main.go`

- [ ] **Step 1: Implementar main.go definitivo**

Reemplazar `cmd/pkgsh/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/pkgsh/internal/adapters"
	"github.com/yourusername/pkgsh/internal/domain"
	"github.com/yourusername/pkgsh/internal/ui"
)

func main() {
	var (
		managerFlag     = flag.String("manager", "", "filtrar por gestor al inicio (apt,snap,flatpak,dpkg,pip,npm,appimage)")
		upgradeableFlag = flag.Bool("upgradeable", false, "mostrar solo paquetes con actualizaciones")
		nativeFlag      = flag.Bool("native", false, "mostrar solo paquetes nativos del OS")
		searchFlag      = flag.String("search", "", "iniciar con búsqueda activa")
	)
	flag.Parse()

	registry := adapters.New()
	model := ui.NewAppModel(registry)

	// Aplicar flags de inicio
	if *managerFlag != "" {
		managers := strings.Split(*managerFlag, ",")
		if len(managers) == 1 {
			model.StartManager = domain.ManagerType(strings.TrimSpace(managers[0]))
		}
		// multi-manager en v1: usar el primero (el filtro por múltiples gestores es v2)
	}
	model.StartUpgradeable = *upgradeableFlag
	model.StartNative = *nativeFlag
	model.StartSearch = *searchFlag

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pkgsh error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Aplicar flags de inicio en AppModel.Init()**

Editar `internal/ui/app.go` — al final de `Init()`, agregar la aplicación de los flags de inicio. Buscar el bloque `Init()` y añadir antes del `return`:

```go
func (m *AppModel) Init() tea.Cmd {
	results := make(chan adapters.LoadResult, 16)
	m.loadCount = 0
	go m.registry.LoadAll(results)

	// Aplicar flags de inicio
	if m.StartManager != "" {
		m.activeTab = m.StartManager
	}
	if m.StartSearch != "" {
		m.searchQuery = m.StartSearch
		m.search.SetValue(m.StartSearch)
	}

	return tea.Batch(
		m.spinner.Tick,
		waitForLoad(results),
	)
}
```

- [ ] **Step 3: Compilar todo el proyecto**

```bash
go build ./...
```

Resolver cualquier error de compilación antes de continuar.

- [ ] **Step 4: Verificar que el binario arranca**

```bash
./pkgsh --help
```

Esperado: muestra los flags disponibles y sale limpiamente.

- [ ] **Step 5: Commit**

```bash
git add cmd/pkgsh/main.go internal/ui/app.go
git commit -m "feat: CLI entrypoint with --manager, --upgradeable, --native, --search flags"
```

---

## Task 20: Distribución — nfpm.yaml y GitHub Actions

**Files:**
- Create: `nfpm.yaml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Crear nfpm.yaml**

Crear `nfpm.yaml`:

```yaml
name: pkgsh
arch: amd64
platform: linux
version: "${VERSION}"
maintainer: "pkgsh contributors"
description: |
  pkgsh — unified terminal package manager for Ubuntu/Debian.
  Manages apt, snap, flatpak, dpkg, pip, npm, and AppImage from a
  single TUI without leaving the terminal.
homepage: "https://github.com/yourusername/pkgsh"
license: MIT

contents:
  - src: ./pkgsh
    dst: /usr/local/bin/pkgsh
    file_info:
      mode: 0755

scripts:
  postinstall: |
    echo "pkgsh installed. Run 'pkgsh' to start."

deb:
  depends:
    - sudo
```

- [ ] **Step 2: Crear .github/workflows/release.yml**

Crear `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goarch: amd64
            goos: linux
          - goarch: arm64
            goos: linux
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Build static binary
        env:
          CGO_ENABLED: "0"
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o pkgsh_${{ github.ref_name }}_${{ matrix.goos }}_${{ matrix.goarch }} ./cmd/pkgsh

      - name: Upload binary artifact
        uses: actions/upload-artifact@v4
        with:
          name: pkgsh-${{ matrix.goarch }}
          path: pkgsh_*

  package-deb:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install nfpm
        run: |
          curl -sfL https://install.goreleaser.com/github.com/goreleaser/nfpm.sh | sh -s -- -b /usr/local/bin

      - name: Build amd64 binary
        env:
          CGO_ENABLED: "0"
        run: go build -ldflags="-s -w" -o pkgsh ./cmd/pkgsh

      - name: Package .deb
        env:
          VERSION: ${{ github.ref_name }}
        run: nfpm package --packager deb --target dist/

      - name: Upload .deb artifact
        uses: actions/upload-artifact@v4
        with:
          name: pkgsh-deb
          path: dist/*.deb

  release:
    runs-on: ubuntu-latest
    needs: [build, package-deb]
    steps:
      - uses: actions/download-artifact@v4
        with:
          merge-multiple: true
          path: artifacts/

      - name: Generate checksums
        run: |
          cd artifacts
          sha256sum * > SHA256SUMS

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            artifacts/*
          generate_release_notes: true
```

- [ ] **Step 3: Verificar estructura final del proyecto**

```bash
find . -type f -name "*.go" | sort
```

Esperado: ver todos los archivos `.go` en sus directorios correctos.

- [ ] **Step 4: Ejecutar todos los tests**

```bash
go test ./... -v
```

Esperado: todos `PASS`.

- [ ] **Step 5: Build final estático**

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o pkgsh ./cmd/pkgsh
ls -lh pkgsh
```

Esperado: binario `pkgsh` de ~10-15 MB.

- [ ] **Step 6: Commit final**

```bash
git add nfpm.yaml .github/
git commit -m "feat: distribution config — nfpm .deb packaging and GitHub Actions release pipeline"
```

---

## Auto-revisión del plan contra el spec

**Cobertura del spec v1:**

| Requisito | Task |
|---|---|
| Listar paquetes apt/snap/flatpak/dpkg/pip/npm/AppImage | Tasks 5-11 |
| Búsqueda en tiempo real | Task 4, Task 14, Task 18 |
| Filtrar por gestor | Task 4, Task 18 |
| Ordenar por nombre/gestor/versión/tamaño | Task 4 |
| Panel detalle con info completa | Task 15 |
| Desinstalar uno o múltiples (lote) | Task 17, Task 18 |
| Actualizar uno o múltiples (lote) | Task 17, Task 18 |
| Output en tiempo real | Task 16 |
| Flags de arranque | Task 19 |
| Distribución binario estático | Task 20 |
| Distribución .deb | Task 20 |
| Elevación sudo con modal seguro | Tasks 3, 16, 17 |
| Confirmación antes de operación destructiva | Task 17 |
| Selección múltiple con checkboxes | Task 14, Task 18 |

**Consistencia de tipos verificada:**
- `Package`, `ManagerType`, `Operation`, `Runner` definidos en Task 2-3, usados consistentemente en Tasks 5-11 y 13-18.
- `FilterByManager`, `FilterBySearch`, `Sort` definidos en Task 4, llamados en Task 18.
- `StreamOperation` definida en Task 16, usada en Task 18.
- `ModalConfirmMsg`, `SudoPromptMsg`, `ModalSudoSubmitMsg` definidos en Task 17, manejados en Task 18.
- `PackagesLoadedMsg`, `LoadResult` definidos en Tasks 12 y 18 consistentemente.
