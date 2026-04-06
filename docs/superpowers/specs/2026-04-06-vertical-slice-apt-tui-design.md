# pkgsh — Slice Vertical: TUI completa + Adapter apt

**Fecha:** 2026-04-06
**Versión:** 1.0
**Estado:** Aprobado

---

## 1. Resumen

Implementar el primer corte vertical funcional de `pkgsh` en tres PRs secuenciales:

1. **PR 1** — TUI completa con datos mock (todos los paneles, keybindings, operaciones simuladas)
2. **PR 2** — Adapter `apt` real con tests
3. **PR 3** — Conexión: reemplaza mock con goroutines reales del adapter

Al terminar los tres PRs, `pkgsh` arranca, llama a `apt`, muestra la lista real de paquetes instalados, permite navegar, buscar, seleccionar múltiples paquetes, desinstalar y actualizar con output en tiempo real.

---

## 2. Decisiones de diseño

| Pregunta | Decisión |
|---|---|
| Layout TUI | Layout completo del spec (3 paneles: lista + detalle + log) |
| Operaciones en este slice | `List()` + `Remove()` + `Update()` con streaming y sudo modal |
| Selección de paquetes | Múltiple con `Space` + operaciones en lote agrupadas por gestor |
| Estrategia de implementación | TUI mock primero → adapter real → conexión |

---

## 3. Arquitectura

### 3.1 Estructura de archivos

```
internal/ui/
  app.go      — modelo raíz Bubble Tea, orquesta todo, maneja keybindings globales
  list.go     — panel lista: navegación, checkboxes, tabs, búsqueda inline, sort
  detail.go   — panel detalle: info completa del paquete bajo cursor
  log.go      — panel log: output de operaciones en tiempo real, scrollable
  modal.go    — modal de confirmación (d/u) y sudo prompt (sin eco)

cmd/pkgsh/main.go  — inicializa TUI con mock (PR 1) → adapter real (PR 3)
internal/adapters/apt/adapter.go       — implementación real
internal/adapters/apt/adapter_test.go  — tests de integración y unitarios
```

### 3.2 Flujo de datos (PR 1)

```
main.go
  └─ crea AppState con []domain.Package mock hardcodeados
  └─ lanza ui.New(state) → tea.NewProgram(model)
         │
         ├─ app.go recibe: tea.KeyMsg, tickMsg, operationLineMsg
         ├─ list.go renderiza lista filtrada + checkboxes
         ├─ detail.go renderiza info del paquete bajo cursor
         └─ log.go acumula lines leídas de Operation.Reader()
```

### 3.3 Flujo de datos (PR 3)

```
main.go
  └─ registra adapters: [apt.New()]
  └─ lanza goroutines paralelas → cada una llama adapter.List()
  └─ resultados se envían al modelo vía tea.Cmd (packagesLoadedMsg)
  └─ spinner visible mientras carga
```

---

## 4. Componentes TUI

### 4.1 `app.go`

- Modelo raíz que contiene `domain.AppState`
- Maneja todos los keybindings globales
- Delega rendering a `list`, `detail`, `log` según `ActivePanel`
- Gestiona el ciclo de vida de `*domain.Operation` activa
- Layout:

```
┌─ pkgsh ──────────────────────────────────────────────────────┐
│ [1] Todos  [2] apt  [3] snap  ...                            │
├──────────────────────────┬───────────────────────────────────┤
│  LISTA                   │  DETALLE                          │
│                          │                                   │
├──────────────────────────┴───────────────────────────────────┤
│  LOG                                                         │
├──────────────────────────────────────────────────────────────┤
│  keybindings bar                                             │
└──────────────────────────────────────────────────────────────┘
```

### 4.2 `list.go`

Keybindings activos:

| Tecla | Acción |
|---|---|
| `↑↓` | Mover cursor |
| `Space` | Marcar/desmarcar paquete |
| `a` | Seleccionar todos los visibles |
| `Esc` | Limpiar selección |
| `s` | Ciclar sort: nombre → gestor → versión → tamaño |
| `/` | Activar búsqueda inline |
| `[1-8]` | Filtrar por gestor (1=Todos, 2=apt, ...) |
| `d` | Abrir modal de confirmación para desinstalar seleccionados |
| `u` | Abrir modal de confirmación para actualizar seleccionados |
| `Tab` | Cambiar panel activo |
| `q` | Salir (con confirmación si hay operación activa) |

### 4.3 `detail.go`

Se re-renderiza en cada movimiento del cursor. Campos mostrados:
- Nombre, Versión, Nueva versión (si disponible, con indicador `↑`)
- Manager, Origen
- Tamaño (formateado: MB/KB)
- Nativo (sí/no)
- Descripción

### 4.4 `log.go`

- Lee líneas de `Operation.Reader()` mediante un `tea.Cmd` que hace `bufio.Scanner`
- Cada línea leída genera un `operationLineMsg` que el modelo acumula en `AppState.LogLines`
- Scrollable con `PgUp/PgDn` cuando `ActivePanel == PanelLog`
- Líneas de error se muestran en rojo

### 4.5 `modal.go`

Dos modos:

**Confirmación:**
```
┌─ Confirmar ──────────────────────────┐
│ Desinstalar 2 paquetes:              │
│   vlc, node                          │
│                                      │
│  [s] Confirmar    [n/Esc] Cancelar   │
└──────────────────────────────────────┘
```

**Sudo prompt:**
```
┌─ Contraseña requerida ───────────────┐
│ sudo necesita tu contraseña:         │
│ > ████████                           │
│                                      │
│  [Enter] Confirmar   [Esc] Cancelar  │
└──────────────────────────────────────┘
```

La contraseña se escribe al stdin del proceso `exec.Cmd` mediante `sudo -S`.

> **PR 1 (mock):** Al confirmar una operación, se crea un `domain.NewOperation()` y se lanza una goroutine que escribe líneas de texto simulado al pipe (`"Removing vlc... done."`) con un pequeño delay. Esto valida todo el flujo TUI sin ejecutar comandos reales. El modal de sudo **no aparece** en mock — las operaciones simuladas no requieren contraseña.

---

## 5. Adapter apt (PR 2)

### 5.1 `List()`

Dos comandos ejecutados al inicio:

```go
// Paquetes instalados
exec.Command("dpkg-query", "-W", "-f=${Package}\t${Version}\t${Installed-Size}\t${db:Status-Abbrev}\t${binary:Summary}\n")

// Paquetes actualizables
exec.Command("apt", "list", "--upgradeable")
```

Ambos se ejecutan como `[]string` sin interpolación de shell. Los resultados se combinan: si un paquete aparece en upgradeable, se popula `Package.NewVersion`.

Solo se incluyen paquetes con status `ii` (instalado correctamente).

### 5.2 `Remove(pkgs []Package)`

```go
args := append([]string{"apt", "remove", "-y"}, names...)
cmd := exec.Command("sudo", "-S", args...)
// stdin conectado al modal de sudo
// stdout/stderr copiados a op.Writer() en goroutine
// op.Done(err) al terminar
```

### 5.3 `Update(pkgs []Package)`

```go
args := append([]string{"apt-get", "install", "--only-upgrade", "-y"}, names...)
cmd := exec.Command("sudo", "-S", args...)
```

### 5.4 `Info(pkg Package)`

```go
exec.Command("apt-cache", "show", pkg.Name)
```

Parsea campos: `Depends:`, `Installed-By:`.

### 5.5 Regla de seguridad

**Todos los comandos son `[]string` pasados a `exec.Command` directamente. Nunca interpolación de strings en shell.** Este es el invariante de seguridad central del proyecto.

---

## 6. Tests

### PR 1 (TUI mock)

- Tests unitarios de `domain.Filter()` — ya existe, ampliar con casos de sort
- Tests de transiciones de estado del modelo Bubble Tea usando input sintético (sin levantar terminal)

### PR 2 (adapter apt)

| Test | Tipo | Descripción |
|---|---|---|
| `TestListParsing` | Unitario | Output fixture de `dpkg-query` hardcodeado → verifica parsing |
| `TestListIntegration` | Integración | Llama `dpkg-query` real, verifica que retorna `[]Package` no vacío |
| `TestUpgradeableParsing` | Unitario | Output fixture de `apt list --upgradeable` → verifica `NewVersion` |

Tests de integración marcados con `//go:build linux` y `t.Skip` si `dpkg-query` no está disponible.

---

## 7. Manejo de errores

| Escenario | Comportamiento |
|---|---|
| `dpkg-query` no disponible | `List()` retorna `(nil, err)`, se muestra en log, no es fatal |
| Operación falla (apt error) | `op.Done(err)` cierra el pipe, última línea en rojo en log |
| `q` durante operación activa | Modal: `"Operación en curso, ¿salir? [s/n]"` |
| Sudo incorrecto | El proceso termina con error, se muestra en log |

---

## 8. Dependencias a agregar (PR 1)

```bash
go get github.com/charmbracelet/bubbletea@v0.26.6
go get github.com/charmbracelet/bubbles@v0.18.0
go get github.com/charmbracelet/lipgloss@v0.11.0
```

---

## 9. Secuencia de PRs

```
master
  └─ feat/tui-mock          → PR 1: TUI completa con datos mock
  └─ feat/adapter-apt       → PR 2: adapter apt real + tests
  └─ feat/connect-apt-tui   → PR 3: conectar adapter con TUI, reemplazar mock
```

Cada PR es mergeable y deja el proyecto en un estado consistente (compilable y ejecutable).
