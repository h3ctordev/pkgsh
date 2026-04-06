# Guía de Contribución — pkgsh

Bienvenido. Esta guía cubre todo lo que necesitas para contribuir al proyecto de forma ordenada y consistente.

---

## Índice

1. [Requisitos previos](#requisitos-previos)
2. [Configuración del entorno](#configuración-del-entorno)
3. [Flujo de trabajo Git](#flujo-de-trabajo-git)
4. [Conventional Commits](#conventional-commits)
5. [Proceso de Pull Request](#proceso-de-pull-request)
6. [Estándares de código](#estándares-de-código)
7. [Tests](#tests)
8. [Reportar bugs](#reportar-bugs)

---

## Requisitos previos

| Herramienta | Versión mínima | Instalación |
|---|---|---|
| Go | 1.22 | `sudo apt install golang-go` |
| Git | 2.x | `sudo apt install git` |
| gh CLI | 2.x | [docs.github.com/cli](https://docs.github.com/en/github-cli) |
| nfpm | 2.x | `go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest` |

Gestores opcionales (para probar los adaptadores):

```bash
sudo apt install snapd flatpak
pip3 install --user pip
npm install -g npm
```

---

## Configuración del entorno

```bash
# 1. Forkear el repositorio desde GitHub, luego clonar tu fork
git clone git@github.com:<tu-usuario>/pkgsh.git
cd pkgsh

# 2. Agregar el repositorio original como upstream
git remote add upstream git@github.com:h3ctordev/pkgsh.git

# 3. Instalar dependencias
go mod download

# 4. Verificar que todo compila
go build ./...
go test ./...
```

---

## Flujo de trabajo Git

### Regla fundamental

**Nunca hagas commits ni push directos a `master`.** Siempre trabaja en una rama propia y abre un PR.

### Nomenclatura de ramas

```
<tipo>/<descripción-corta-en-kebab-case>
```

| Tipo | Cuándo usarlo |
|---|---|
| `feat/` | Nueva funcionalidad |
| `fix/` | Corrección de bug |
| `docs/` | Solo documentación |
| `refactor/` | Sin cambio de comportamiento |
| `test/` | Agregar o corregir tests |
| `chore/` | Mantenimiento, deps, config |
| `ci/` | Cambios en pipelines |

**Ejemplos correctos:**

```bash
feat/apt-adapter-list
fix/ui-crash-empty-packages
docs/contributing-guide
refactor/operation-streaming
test/apt-adapter-unit
```

### Ciclo de trabajo

```bash
# 1. Sincronizar master con upstream antes de crear una rama
git checkout master
git pull upstream master

# 2. Crear rama desde master actualizado
git checkout -b feat/mi-feature

# 3. Desarrollar, commitear con Conventional Commits (ver sección siguiente)
git add <archivos>
git commit -m "feat(adapters/apt): implement List() using dpkg-query"

# 4. Pushear la rama
git push -u origin feat/mi-feature

# 5. Abrir PR desde GitHub o con gh CLI
gh pr create --base master --title "feat(adapters/apt): implement List()"
```

---

## Conventional Commits

Todo commit **debe** seguir este formato:

```
<type>(<scope>): <description>

[body opcional — explica el por qué, no el qué]

[footer opcional — e.g. Closes #123]
```

### Tipos

| Tipo | Cuándo |
|---|---|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `chore` | Deps, config, build, sin cambio de lógica |
| `docs` | Solo documentación |
| `refactor` | Reestructuración sin cambio de comportamiento |
| `test` | Agregar o corregir tests |
| `ci` | Pipelines CI/CD |
| `style` | Formato, espacios — sin cambio de lógica |
| `perf` | Mejoras de rendimiento |
| `revert` | Revertir commit anterior |

### Reglas del mensaje

- Descripción en **imperativo**, minúsculas, sin punto al final
- Máximo 72 caracteres en la primera línea
- El scope es la carpeta o módulo afectado

### Ejemplos correctos

```
feat(adapters/apt): implement List() using dpkg-query
fix(ui): prevent crash when package list is empty
docs(readme): add ARM64 installation instructions
test(adapters/snap): add unit tests for Remove()
chore(deps): upgrade bubbletea to v0.27.0
ci(release): add arm64 build to release matrix
refactor(domain): extract filter logic into separate file
```

### Ejemplos incorrectos

```
# Sin tipo
added apt support

# Mayúsculas
Feat(apt): Add List function

# Con punto al final
fix(ui): prevent crash.

# Vago
fix: stuff
```

---

## Proceso de Pull Request

### Antes de abrir el PR

```bash
# Asegurarse de que los tests pasan
go test ./...

# Asegurarse de que compila
go build ./...

# Formatear el código
go fmt ./...

# Lint (si tienes golangci-lint)
golangci-lint run
```

### Título del PR

El título sigue el mismo formato que Conventional Commits:

```
feat(adapters/apt): implement List() using dpkg-query
```

### Descripción del PR

Usa esta estructura:

```markdown
## ¿Qué hace este PR?
Breve descripción del cambio.

## ¿Por qué?
Motivación del cambio. Referencia a issue si aplica: Closes #12

## Cómo probar
1. Pasos para verificar el cambio manualmente
2. ...

## Checklist
- [ ] Tests pasan (`go test ./...`)
- [ ] Compila sin errores (`go build ./...`)
- [ ] Código formateado (`go fmt ./...`)
- [ ] CHANGELOG.md actualizado si es un cambio visible para el usuario
```

### Revisión

- Se requiere **al menos 1 aprobación** antes de mergear
- Los checks de CI deben pasar
- No hacer squash de commits significativos — se mantiene el historial limpio
- El merge lo hace el maintainer, no el autor del PR

---

## Estándares de código

### Seguridad (regla no negociable)

**Todos los comandos del sistema son `[]string` pasados a `exec.Cmd`.** Nunca construir comandos concatenando strings.

```go
// CORRECTO
cmd := exec.Command("apt", "remove", pkg.Name)

// INCORRECTO — command injection
cmd := exec.Command("sh", "-c", "apt remove "+pkg.Name)
```

### Arquitectura

Respetar las cuatro capas. Las dependencias solo fluyen hacia abajo:

```
UI → Domain → Adapters → System
```

- La capa UI no importa paquetes de adapters directamente
- Los adapters no importan nada de UI
- El domain no importa nada de adapters ni UI

### Agregar un nuevo adapter

1. Crear `internal/adapters/<manager>/adapter.go`
2. Implementar `domain.PackageManager` completo
3. Registrar en `cmd/pkgsh/main.go`
4. Agregar tests en `internal/adapters/<manager>/adapter_test.go`

No se modifica ninguna otra capa.

---

## Tests

```bash
# Todos los tests
go test ./...

# Tests de un adapter específico
go test ./internal/adapters/apt/...

# Test específico por nombre
go test -run TestAptList ./internal/adapters/apt/...

# Con coverage
go test -cover ./...
```

Los tests de adapters que dependen de comandos del sistema deben usar interfaces para mockear `exec.Cmd` — no llamar binarios reales en los tests unitarios.

---

## Reportar bugs

Abre un issue en GitHub usando la plantilla de bug report. Incluye:

- Versión de pkgsh (`pkgsh --version`)
- Distribución y versión de Ubuntu/Debian
- Pasos para reproducir
- Comportamiento esperado vs. actual
- Output completo del error (si aplica)

---

## Preguntas

Si tienes dudas sobre el diseño o arquitectura, abre un issue con el label `question` antes de implementar. Así evitamos esfuerzo duplicado.
