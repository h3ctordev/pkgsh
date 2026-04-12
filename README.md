# pkgsh

Gestor unificado de paquetes para Ubuntu/Debian con interfaz TUI. Abstrae apt, snap, flatpak, dpkg, pip, npm y AppImage en una sola experiencia de terminal.

```
 pkgsh  312 paquetes  ·  4 actualizables  ·  12 huérfanos              [SEC]
─────────────────────────────────────────────────────────────────────────────
 > Buscar: ______________
─────────────────────────────────────────────────────────────────────────────
  NOMBRE              VERSIÓN     GESTOR     TAMAÑO    TAGS      INSTALADO
  ──────────────────────────────────────────────────────────────────────────
  ☐ firefox           126.0.1     apt        234 MB    UPD       2024-01-15
  ☐ snapd             2.63        snap        89 MB              2023-11-02
  ☒ vlc               3.0.21      apt         67 MB    ORP       2023-08-20
  ☒ node              20.1.0      npm          —                 2024-02-01
  ☐ spotify           1.2.31      snap       180 MB              2023-09-10
─────────────────────────────────────────────────────────────────────────────
 [/] Buscar  [Space] Sel  [d] Desinstalar  [u] Actualizar  [r] Recargar  [?] Ayuda
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
pkgsh                            # vista completa, todos los gestores
pkgsh --manager apt              # arrancar filtrado a apt
pkgsh --manager pip,npm          # arrancar con pip y npm activos
pkgsh --upgradeable              # solo paquetes con actualizaciones disponibles
pkgsh --native                   # solo paquetes nativos del OS
pkgsh --search firefox           # arrancar con búsqueda activa
pkgsh --no-security-mode         # desactivar bloqueo de paquetes del sistema
pkgsh --version                  # mostrar versión instalada
```

## Keybindings

| Tecla      | Acción                                                       |
|------------|--------------------------------------------------------------|
| `↑↓` / `jk` | Navegar la lista                                            |
| `g` / `G`  | Ir al primer / último paquete                                |
| `Space`    | Marcar/desmarcar paquete para operación en lote              |
| `a`        | Seleccionar todos los paquetes visibles                      |
| `Esc`      | Limpiar selección                                            |
| `d`        | Desinstalar seleccionado(s) — pide confirmación              |
| `u`        | Actualizar seleccionado(s) — pide confirmación               |
| `s`        | Ciclar ordenamiento: nombre → gestor → versión → tamaño      |
| `S`        | Activar/desactivar Security Mode                             |
| `/`        | Activar búsqueda                                             |
| `r`        | Recargar lista de paquetes                                   |
| `?`        | Abrir ayuda con todos los atajos                             |
| `Tab`      | Cambiar panel activo                                         |
| `[1-8]`    | Filtrar por gestor (1=Todos, 2=apt … 8=AppImage)             |
| `q`        | Salir                                                        |
| `Ctrl+C`   | Salir inmediatamente                                         |

## Security Mode

Activado por defecto. Oculta los paquetes del sistema (kernel, libc, systemd…) y bloquea operaciones sobre ellos para evitar daños accidentales. El badge `[SEC]` en el header indica que está activo.

```bash
pkgsh                   # security mode ON (por defecto)
pkgsh --no-security-mode  # sin restricciones
```

Dentro de la TUI, `S` alterna el modo en caliente.

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
