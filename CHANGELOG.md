# Changelog

All notable changes to pkgsh will be documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.2.1] - 2026-04-12

### Fixed
- **Paquetes duplicados apt/dpkg** — `DeduplicatePackages` nunca se invocaba tras el refactor de carga progresiva que eliminó `loadPackages()` de `main.go`. Ahora se llama en los dos puntos donde se consolidan los paquetes: al terminar la carga inicial y tras cada recarga post-operación.

---

## [0.2.0] - 2026-04-11

### Added
- **Security Mode** — activado por defecto; oculta y bloquea paquetes del sistema en operaciones destructivas. Toggle con `S`, badge `[SEC]` en footer. Flag `--no-security-mode` para desactivarlo
- **Smart Panels UI redesign** — paleta GitHub Dark completa; tags semánticos `UPD` / `SYS` / `ORP`; navegación `j`/`k`/`g`/`G`; tecla `r` para recargar; `?` para ayuda
- **ModalHelp** — pantalla de ayuda completa con todos los atajos de teclado
- **ModalSudo** — intercepción del prompt de contraseña sudo en el stream de salida; input seguro con enmascarado `*`
- **Header stats** — contador de paquetes totales, actualizables y huérfanos en el header
- **Footer contextual** — muestra las acciones disponibles según el panel activo
- **Log auto-collapse** — el panel de log se colapsa automáticamente al terminar una operación
- **Columna `Instalado`** — fecha de instalación por paquete, poblada en todos los adapters (apt, snap, flatpak, dpkg, pip, npm, appimage)
- **Columna `Tamaño`** en lista de paquetes
- **`IsOrphan`** en `domain.Package` — adapter `apt` vía `apt-mark showauto`; adapter `flatpak` marca runtimes como huérfanos
- **`IsSystemPackage()`** — lista estática + patrones de prefijo para detectar paquetes del sistema
- **`DeduplicatePackages()`** — si un paquete aparece en apt y dpkg con el mismo nombre, gana apt
- **`SECURITY.md`** — política de seguridad y proceso de reporte de vulnerabilidades

### Changed
- Panel de lista rediseñado con colores, columnas dinámicas Name/Version/Manager/Size/Tags/Instalado
- Panel de detalle muestra flecha de versión actual → nueva, manager coloreado, descripción truncada
- Panel de log con título spinner, colores semánticos por tipo de línea y auto-scroll
- `StreamCmd` usa `StdinPipe` para evitar bloqueo en `cmd.Wait()`
- Flatpak usa Application ID en operaciones y nombre de display en la UI
- Footer ajusta ancho del badge para prevenir overflow en terminales estrechas
- Padding vertical en header y barra de búsqueda

### Fixed
- Alineación de columnas con códigos ANSI y checkboxes Unicode
- Alias de colores de manager a constantes de paleta semántica
- Prefijos `linux-modules` reordenados; `python3-apt` movido a mapa de nombre exacto
- Strips de `\r` en líneas de log; truncado de líneas largas al ancho del panel
- Race condition en `TestStreamCmdStdin` usando `io.Pipe` directo
- Guard nil en `Operation` al confirmar `ModalSudo`

---

## [0.1.0] - 2026-04-06

### Added
- Scaffolding inicial del proyecto (`go.mod`, estructura de directorios, dependencias)
- Definición de la interfaz `PackageManager` y tipos de dominio (`Package`, `Operation`, `AppState`, `Filter`, `Sort`)
- Helper compartido `exec` para adapters (`RunCmd`, `StreamCmd`)
- Adapter `apt` — List, Remove, Update, Info + tests
- Adapter `snap` — List, Remove, Update, Info + tests
- Adapter `flatpak` — List, Remove, Update, Info + tests
- Adapter `dpkg` — List, Info (solo lectura) + tests
- Adapter `pip` — List, Remove, Update, Info + tests
- Adapter `npm` — List, Remove, Update, Info + tests
- Adapter `appimage` — List, Remove + tests
- UI: ListModel, DetailModel, LogModel, ModalConfirm, ModalSudo base
- CLI entrypoint con flags (`--manager`, `--upgradeable`, `--native`, `--search`, `--version`)
- Pipeline CI/CD con GitHub Actions (amd64 + arm64, `.deb` via nfpm, SHA256)
- `nfpm.yaml` para empaquetado `.deb`
- `docs/pkgsh.1` — man page
- `scripts/postinstall.sh`
