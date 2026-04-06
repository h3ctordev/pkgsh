# pkgsh — Software Design Document

**Fecha:** 2026-04-05
**Versión:** 1.0
**Estado:** Aprobado

---

## 1. Resumen Ejecutivo

`pkgsh` es un gestor unificado de paquetes para Ubuntu/Debian con interfaz TUI (Terminal User Interface). Abstrae apt, snap, flatpak, dpkg, pip, npm y AppImage en una sola experiencia coherente. El usuario puede ver, buscar, clasificar, actualizar y desinstalar cualquier paquete instalado sin importar su origen, sin salir de la terminal y sin ejecutar comandos manuales.

### Stack tecnológico

| Componente | Tecnología | Justificación |
|---|---|---|
| Lenguaje | Go | Binario único, sin runtime, compilación cruzada trivial |
| TUI framework | Bubble Tea (charmbracelet) | El estándar de facto en TUI modernas con Go, usado por lazygit, glow |
| Empaquetado | nfpm | Genera .deb sin scripts complejos, compatible con CI/CD |
| CI/CD | GitHub Actions | Release automático con matrix de arquitecturas |

---

## 2. Alcance

### En scope — v1

- Listar todos los paquetes instalados de: apt, snap, flatpak, dpkg, pip, npm, AppImage
- Buscar paquetes por nombre en tiempo real
- Filtrar por gestor de origen
- Ordenar por nombre, gestor, versión, tamaño
- Ver detalle de cada paquete (versión, origen, tamaño, nativo/externo, descripción)
- Desinstalar uno o múltiples paquetes (operación en lote)
- Actualizar uno o múltiples paquetes (operación en lote)
- Output en tiempo real de operaciones dentro de la TUI
- Flags de arranque para pre-filtrar la vista
- Distribución como binario estático y paquete `.deb`

### Fuera de scope — v1

- Instalación de paquetes nuevos (solo gestión de los ya instalados)
- Multi-distro (Fedora, Arch, etc.) — arquitectura lo permite, no se implementa aún
- Repositorios privados o autenticados
- Caché persistente en disco entre sesiones

---

## 3. Arquitectura

### 3.1 Capas

```
┌─────────────────────────────────────┐
│           UI Layer (Bubble Tea)      │
│  Panels · Search · Modals · Output  │
└──────────────────┬──────────────────┘
                   │
┌──────────────────▼──────────────────┐
│          Domain Layer               │
│  PackageManager interface           │
│  BulkOperation · Filter · Sort      │
└──────────────────┬──────────────────┘
                   │
┌──────────────────▼──────────────────┐
│         Adapter Layer               │
│  apt │ snap │ flatpak │ dpkg        │
│  pip │ npm  │ AppImage              │
└──────────────────┬──────────────────┘
                   │
┌──────────────────▼──────────────────┐
│        System Layer                 │
│  exec.Cmd · streaming stdout/stderr │
│  sudo elevation · signal handling  │
└─────────────────────────────────────┘
```

### 3.2 Interfaz común de adaptadores

Cada gestor implementa la siguiente interfaz:

```go
type PackageManager interface {
    Name() string
    List() ([]Package, error)
    Remove(pkgs []Package) *Operation
    Update(pkgs []Package) *Operation
    Info(pkg Package) (PackageInfo, error)
}
```

`Operation` expone un `io.Reader` para streaming del output en tiempo real al panel de log de la TUI.

### 3.3 Carga de datos al arrancar

Al iniciar, `pkgsh` lanza todos los adaptadores en goroutines paralelas. La lista se va poblando progresivamente con un spinner visible. El caché vive en memoria por sesión — no persiste en disco.

---

## 4. Modelo de Datos

### 4.1 Paquete

```go
type Package struct {
    Name        string
    Version     string
    NewVersion  string        // vacío si no hay actualización disponible
    Manager     ManagerType   // apt, snap, flatpak, pip, npm, appimage
    Size        int64         // bytes
    Description string
    IsNative    bool          // true si viene de repositorios oficiales del OS
    Origin      string        // "ubuntu", "ppa:...", "pypi", "npmjs", etc.
    InstallDate time.Time     // si el gestor lo expone
}
```

### 4.2 Estado global de la aplicación

```go
type AppState struct {
    Packages     []Package
    Filtered     []Package     // resultado de búsqueda/filtro activo
    Selected     map[int]bool  // índices seleccionados para operación en lote
    ActiveTab    ManagerType   // filtro de gestor activo
    SearchQuery  string
    SortBy       SortField
    ActivePanel  Panel         // list | detail | log
    Operation    *Operation    // operación en curso (nil si ninguna)
    LogLines     []string      // output acumulado del log
}
```

---

## 5. Interfaz TUI

### 5.1 Layout

```
┌─ pkgsh ─────────────────────────────────────────────────────────────────────────┐
│ [1] Todos  [2] apt  [3] snap  [4] flatpak  [5] dpkg  [6] pip  [7] npm  [8] AppImage│
├──────────────────────────────────┬───────────────────────────────────────┤
│ > Buscar: ___________________    │  DETALLE                              │
├──────────────────────────────────┤  Nombre:   firefox                   │
│ ☐ firefox          apt  126.0  ✓ │  Versión:  126.0.1                   │
│ ☐ snapd            snap  2.63  ✓ │  Origen:   apt (ubuntu)              │
│ ☒ vlc              apt  3.0.2  ✓ │  Tamaño:   234 MB                    │
│ ☒ node             npm  20.1   ✓ │  Nativo:   No                        │
│ ☐ spotify          snap  1.2   ✓ │  Desc:     Web browser               │
│ ☐ ...                            │                                       │
├──────────────────────────────────┴───────────────────────────────────────┤
│ LOG OUTPUT                                                                │
│ Removing vlc... done.                                                     │
│ Removing node... done.                                                    │
└──────────────────────────────────────────────────────────────────────────┘
│ [/] Buscar [Tab] Panel [Space] Sel [d] Desinstalar [u] Actualizar [q] Salir │
```

### 5.2 Paneles

| Panel | Descripción |
|---|---|
| Lista | Navegable con `↑↓`, filtrable por gestor con tabs `[1-7]` |
| Detalle | Se actualiza al mover el cursor, muestra info completa del paquete |
| Log | Output en tiempo real de operaciones, scrollable |

### 5.3 Keybindings

| Tecla | Acción |
|---|---|
| `↑↓` | Navegar lista |
| `Space` | Marcar/desmarcar paquete para operación en lote |
| `a` | Seleccionar todos los paquetes visibles |
| `Esc` | Limpiar selección |
| `d` | Desinstalar seleccionado(s) — pide confirmación |
| `u` | Actualizar seleccionado(s) — pide confirmación |
| `s` | Ciclar criterio de ordenamiento (nombre → gestor → versión → tamaño) |
| `/` | Activar búsqueda |
| `Tab` | Cambiar panel activo |
| `[1-8]` | Filtrar por gestor (1=Todos, 2=apt, 3=snap, 4=flatpak, 5=dpkg, 6=pip, 7=npm, 8=AppImage) |
| `q` | Salir |

### 5.4 Indicadores visuales

- `↑` junto a la versión indica actualización disponible
- `☒` checkbox marcado indica paquete seleccionado para operación en lote
- Spinner durante carga inicial de paquetes

---

## 6. Operaciones

### 6.1 Flujo de operación en lote

```
Usuario presiona [d]
       │
       ▼
Modal confirmación: "Desinstalar 3 paquetes: vlc, node, gimp. ¿Confirmar? [s/n]"
       │
       ▼
pkgsh construye comandos agrupados por gestor:
  apt remove vlc gimp
  npm uninstall -g node
       │
       ▼
Cada comando corre como exec.Cmd con stdout/stderr
streameado al panel LOG en tiempo real
       │
       ▼
Al terminar: refresca lista (re-ejecuta adaptadores afectados)
```

### 6.2 Elevación de privilegios

- Operaciones que requieren `sudo` (apt, dpkg, snap, flatpak) se ejecutan vía `sudo`
- Si se requiere contraseña, `pkgsh` detecta el prompt y abre un modal de input seguro (sin eco) dentro de la TUI
- pip y npm en modo usuario (`--user`) no requieren sudo

### 6.3 Seguridad

- Los comandos se construyen como slices `[]string` — nunca interpolación de strings en shell. Elimina riesgo de command injection.
- No se almacenan credenciales en disco.
- Operaciones destructivas siempre requieren confirmación explícita en modal.

---

## 7. Flags de Arranque

```bash
pkgsh                          # vista completa, todos los gestores
pkgsh --manager apt            # arranca filtrado a apt
pkgsh --manager pip,npm        # arranca con pip y npm activos
pkgsh --upgradeable            # solo paquetes con actualizaciones disponibles
pkgsh --native                 # solo paquetes nativos del OS
pkgsh --search firefox         # arranca con búsqueda activa
```

---

## 8. Distribución y Empaquetado

### 8.1 Artefactos de release

```
pkgsh_1.0.0_linux_amd64          # binario estático, sin dependencias
pkgsh_1.0.0_linux_arm64          # para Raspberry Pi / ARM
pkgsh_1.0.0_amd64.deb            # instalable con dpkg -i o apt install ./
```

### 8.2 Binario estático

- Compilado con `CGO_ENABLED=0 GOOS=linux go build`
- Sin dependencias de runtime
- Publicado en GitHub Releases con checksums SHA256

### 8.3 Paquete .deb

- Generado con `nfpm`
- Instala el binario en `/usr/local/bin/pkgsh`
- Incluye manpage en `/usr/share/man/man1/pkgsh.1`
- Script `postinst` que verifica dependencias del sistema (sudo, apt, snap, etc.)

### 8.4 Pipeline CI/CD (GitHub Actions)

```
tag v*  →  go build (matrix: amd64/arm64)
        →  nfpm package → .deb
        →  sha256sum
        →  gh release create + upload artifacts
```

### 8.5 Licencia

MIT — distribución libre sin restricciones.

---

## 9. Extensibilidad Futura

La interfaz `PackageManager` permite agregar soporte multi-distro sin modificar las capas superiores:

- **Fedora/RHEL:** adaptador `dnf`
- **Arch Linux:** adaptador `pacman` / `yay`
- **macOS:** adaptador `brew`

Cada nuevo gestor solo necesita implementar la interfaz y registrarse en el registry de adaptadores.

---

## 10. Estructura de Directorios del Proyecto

```
pkgsh/
├── cmd/
│   └── pkgsh/
│       └── main.go           # entrypoint, parsing de flags
├── internal/
│   ├── ui/                   # capa TUI (Bubble Tea models)
│   │   ├── app.go
│   │   ├── list.go
│   │   ├── detail.go
│   │   └── log.go
│   ├── domain/               # tipos y lógica de negocio
│   │   ├── package.go
│   │   ├── operation.go
│   │   └── filter.go
│   └── adapters/             # un subdirectorio por gestor
│       ├── apt/
│       ├── snap/
│       ├── flatpak/
│       ├── dpkg/
│       ├── pip/
│       ├── npm/
│       └── appimage/
├── pkg/                      # utilidades exportables (si aplica)
├── .github/
│   └── workflows/
│       └── release.yml
├── nfpm.yaml                 # configuración de empaquetado .deb
├── go.mod
└── go.sum
```
