# pkgsh

Gestor unificado de paquetes para Ubuntu/Debian con interfaz TUI. Abstrae apt, snap, flatpak, dpkg, pip, npm y AppImage en una sola experiencia de terminal.

```
┌─ pkgsh ──────────────────────────────────────────────────────────────────────────┐
│ [1] Todos  [2] apt  [3] snap  [4] flatpak  [5] dpkg  [6] pip  [7] npm  [8] AppImage │
├──────────────────────────────────┬───────────────────────────────────────────────┤
│ > Buscar: ___________________    │  DETALLE                                      │
├──────────────────────────────────┤  Nombre:   firefox                            │
│ ☐ firefox          apt  126.0 ↑  │  Versión:  126.0.1                            │
│ ☐ snapd            snap  2.63    │  Origen:   apt (ubuntu)                       │
│ ☒ vlc              apt  3.0.2    │  Tamaño:   234 MB                             │
│ ☒ node             npm  20.1     │  Nativo:   Sí                                 │
│ ☐ spotify          snap  1.2     │  Desc:     Web browser                        │
├──────────────────────────────────┴───────────────────────────────────────────────┤
│ LOG OUTPUT                                                                        │
│ Removing vlc... done.                                                             │
└──────────────────────────────────────────────────────────────────────────────────┘
│ [/] Buscar  [Tab] Panel  [Space] Sel  [d] Desinstalar  [u] Actualizar  [q] Salir  │
```

## Instalación

### Binario estático (recomendado)

```bash
curl -L https://github.com/hbustos/pkgsh/releases/latest/download/pkgsh_linux_amd64 -o pkgsh
chmod +x pkgsh
sudo mv pkgsh /usr/local/bin/
```

### Paquete .deb

```bash
curl -L https://github.com/hbustos/pkgsh/releases/latest/download/pkgsh_amd64.deb -o pkgsh.deb
sudo dpkg -i pkgsh.deb
```

### Desde fuente

```bash
git clone https://github.com/hbustos/pkgsh.git
cd pkgsh
CGO_ENABLED=0 go build -o pkgsh ./cmd/pkgsh
sudo mv pkgsh /usr/local/bin/
```

## Uso

```bash
pkgsh                          # vista completa, todos los gestores
pkgsh --manager apt            # arrancar filtrado a apt
pkgsh --manager pip,npm        # arrancar con pip y npm activos
pkgsh --upgradeable            # solo paquetes con actualizaciones disponibles
pkgsh --native                 # solo paquetes nativos del OS
pkgsh --search firefox         # arrancar con búsqueda activa
```

## Keybindings

| Tecla     | Acción                                                        |
|-----------|---------------------------------------------------------------|
| `↑↓`      | Navegar la lista                                              |
| `Space`   | Marcar/desmarcar paquete para operación en lote               |
| `a`       | Seleccionar todos los paquetes visibles                       |
| `Esc`     | Limpiar selección                                             |
| `d`       | Desinstalar seleccionado(s) — pide confirmación               |
| `u`       | Actualizar seleccionado(s) — pide confirmación                |
| `s`       | Ciclar ordenamiento: nombre → gestor → versión → tamaño       |
| `/`       | Activar búsqueda                                              |
| `Tab`     | Cambiar panel activo                                          |
| `[1-8]`   | Filtrar por gestor (1=Todos, 2=apt … 8=AppImage)              |
| `q`       | Salir                                                         |
| `Ctrl+C`  | Salir inmediatamente                                          |

## Arquitectura

Cuatro capas con dependencias estrictamente hacia abajo:

```
UI (Bubble Tea) → Domain (interfaces, tipos) → Adapters (uno por gestor) → System (exec.Cmd)
```

Regla de seguridad principal: **todos los comandos son `[]string` pasados a `exec.Cmd` — nunca interpolación de strings en shell.**

## Gestores soportados

| Gestor   | Listar | Desinstalar | Actualizar | Sudo requerido |
|----------|--------|-------------|------------|----------------|
| apt      | ✓      | ✓           | ✓          | Sí             |
| snap     | ✓      | ✓           | ✓          | Sí             |
| flatpak  | ✓      | ✓           | ✓          | Sí             |
| dpkg     | ✓      | —           | —          | —              |
| pip      | ✓      | ✓           | ✓          | No (`--user`)  |
| npm      | ✓      | ✓           | ✓          | No (`--user`)  |
| AppImage | ✓      | ✓           | —          | No             |

## Desarrollo

```bash
go build ./cmd/pkgsh          # compilar
go run ./cmd/pkgsh            # ejecutar sin compilar
go test ./...                 # todos los tests
go test ./internal/adapters/apt/...  # adapter específico

# Release local
nfpm package --packager deb --target dist/
```

### Agregar un nuevo gestor

1. Crear `internal/adapters/<manager>/adapter.go`
2. Implementar la interfaz `domain.PackageManager`
3. Registrar en `cmd/pkgsh/main.go`

No se modifica ninguna otra capa.

## Roadmap

Funcionalidades planeadas para versiones futuras:

| Feature | Descripción |
|---------|-------------|
| **Búsqueda e instalación** | Buscar paquetes en repositorios remotos (apt, snap, flatpak…) e instalar directamente desde la TUI |
| **Historial de operaciones** | Log persistente en disco de cada install/remove/update con timestamp y resultado |
| **Menú de configuración** | Panel de ajustes accesible desde la TUI: gestores activos, comportamiento de SecurityMode, preferencias de UI |
| **Temas y paletas de color** | Soporte para múltiples temas (GitHub Dark, Dracula, Solarized…) y cambio de paleta en caliente |
| **Multi-idioma (i18n)** | Interfaz disponible en inglés, español y otros idiomas configurables |

## Licencia

MIT
