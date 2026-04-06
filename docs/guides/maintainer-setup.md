# Guía de Configuración del Maintainer — pkgsh

Esta guía cubre la configuración completa del repositorio en GitHub para garantizar un flujo de colaboración correcto: protección de ramas, revisiones obligatorias, CI, permisos y plantillas.

---

## Índice

1. [Protección de ramas](#1-protección-de-ramas)
2. [Configuración de GitHub Actions](#2-configuración-de-github-actions)
3. [Secrets y variables de entorno](#3-secrets-y-variables-de-entorno)
4. [CODEOWNERS](#4-codeowners)
5. [Plantillas de PR e Issues](#5-plantillas-de-pr-e-issues)
6. [Configuración de colaboradores](#6-configuración-de-colaboradores)
7. [Releases y versionado](#7-releases-y-versionado)
8. [Checklist inicial](#8-checklist-inicial)

---

## 1. Protección de ramas

Este repositorio usa el sistema de **GitHub Rulesets** (Settings → Rules → Rulesets), que es la API moderna de protección de ramas y reemplaza al sistema clásico de branch protection.

> **Importante:** No usar la API clásica `/branches/{branch}/protection` — las reglas creadas ahí aparecen bajo Settings → Branches (UI antigua) y no en la UI de Rulesets.

### Reglas configuradas en `master`

| Regla | Valor | Por qué |
|---|---|---|
| Require a pull request before merging | ✅ | Nadie, ni el owner, puede push directo |
| Required approvals | 0 | Solo maintainer activo, sin equipo externo aún |
| Dismiss stale reviews on push | ✅ | Un nuevo commit invalida aprobaciones anteriores |
| Require status checks to pass | ✅ (si hay CI activo) | CI debe pasar antes de mergear |
| Require branches to be up to date | ✅ | Sin merges con código desactualizado |
| Block force pushes | ✅ | Protege el historial |
| Restrict deletions | ✅ | Nadie puede borrar master |

### Crear el Ruleset vía gh CLI

```bash
gh api repos/h3ctordev/pkgsh/rulesets \
  --method POST \
  --input - <<'EOF'
{
  "name": "protect-master",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["refs/heads/master"],
      "exclude": []
    }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false
      }
    }
  ]
}
EOF
```

> **Nota:** `required_approving_review_count: 0` significa que se requiere abrir un PR pero no necesita aprobación externa — útil cuando aún no hay reviewers adicionales. Cuando el equipo crezca, subir este valor a `1`.

### Actualizar el Ruleset existente

Para modificar el Ruleset (p. ej. subir las aprobaciones requeridas a 1):

```bash
# Primero obtener el ID del ruleset
gh api repos/h3ctordev/pkgsh/rulesets

# Luego hacer PATCH con el ID
gh api repos/h3ctordev/pkgsh/rulesets/<ID> \
  --method PUT \
  --input - <<'EOF'
{
  "name": "protect-master",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["refs/heads/master"],
      "exclude": []
    }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 1,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false
      }
    }
  ]
}
EOF
```

---

## 2. Configuración de GitHub Actions

El pipeline `.github/workflows/release.yml` se dispara con `git tag v*`.

### Permisos requeridos del workflow

En **Settings → Actions → General**:

- **Actions permissions:** Allow all actions
- **Workflow permissions:** Read and write permissions (para poder crear releases)
- Marcar **Allow GitHub Actions to create and approve pull requests**

### Verificar que el pipeline funciona

```bash
# Simular un release con una tag de prueba
git tag v0.0.1-test
git push origin v0.0.1-test

# Ver el estado del workflow
gh run list --workflow=release.yml

# Eliminar la tag de prueba
git tag -d v0.0.1-test
git push origin --delete v0.0.1-test
```

---

## 3. Secrets y variables de entorno

El pipeline actual no requiere secrets adicionales — usa el `GITHUB_TOKEN` automático de Actions para crear releases.

Si en el futuro se agregan notificaciones o firmas GPG, configurar en **Settings → Secrets and variables → Actions**:

| Secret | Para qué | Cuándo agregar |
|---|---|---|
| `GPG_PRIVATE_KEY` | Firma de binarios | Si se activa firma de releases |
| `SLACK_WEBHOOK` | Notificaciones de release | Opcional |

---

## 4. CODEOWNERS

El archivo `CODEOWNERS` define quién debe revisar automáticamente los PRs según las rutas modificadas.

Crear el archivo en `.github/CODEOWNERS`:

```
# Owner global — revisor por defecto de todo el repo
*                          @h3ctordev

# Adapters — si en el futuro hay expertos por área
internal/adapters/apt/     @h3ctordev
internal/adapters/snap/    @h3ctordev
internal/adapters/flatpak/ @h3ctordev
internal/ui/               @h3ctordev

# CI/CD — cambios en pipelines requieren revisión del owner
.github/                   @h3ctordev
```

Con esto, GitHub asigna automáticamente el reviewer correcto al abrir un PR.

---

## 5. Plantillas de PR e Issues

### Plantilla de PR

Ya existe en `.github/pull_request_template.md`. GitHub la usa automáticamente al abrir cualquier PR.

### Plantillas de Issues

Están en `.github/ISSUE_TEMPLATE/`. GitHub muestra un selector al crear un issue.

**Bug report** (`.github/ISSUE_TEMPLATE/bug_report.md`):
- Versión de pkgsh
- Distribución y versión del OS
- Pasos para reproducir
- Comportamiento esperado vs. actual

**Feature request** (`.github/ISSUE_TEMPLATE/feature_request.md`):
- Problema que resuelve
- Solución propuesta
- Alternativas consideradas

> Estos archivos ya están creados en el repositorio. Si necesitas modificarlos, edítalos en una rama y abre PR.

---

## 6. Configuración de colaboradores

### Agregar un colaborador

En **Settings → Collaborators and teams → Add people**:

| Rol | Permisos | Para quién |
|---|---|---|
| **Write** | Push a ramas no protegidas, abrir PRs | Colaboradores activos |
| **Triage** | Gestionar issues y PRs, sin push | Moderadores de comunidad |
| **Read** | Solo lectura | Observers |

> Los colaboradores con rol **Write** pueden hacer push a ramas de feature pero **no** a `master` (protegida). Todo pasa por PR.

### Proceso de onboarding para colaboradores nuevos

1. Agregar con rol **Write**
2. Compartir el link a `CONTRIBUTING.md`
3. Pedirles que hagan fork + clonar + abrir un PR de prueba con algo menor (un typo en docs, etc.)
4. Revisar y mergear ese primer PR como verificación

---

## 7. Releases y versionado

El proyecto usa **Semantic Versioning** (`MAJOR.MINOR.PATCH`):

| Cambio | Versión |
|---|---|
| Breaking change en CLI o comportamiento | MAJOR |
| Nueva funcionalidad compatible | MINOR |
| Bugfix o mejora interna | PATCH |

### Proceso de release

```bash
# 1. Asegurarse de estar en master actualizado
git checkout master && git pull origin master

# 2. Actualizar CHANGELOG.md
#    Mover los items de [Unreleased] a la nueva versión con fecha
#    Ejemplo: ## [1.0.0] - 2026-05-01

# 3. Commitear el CHANGELOG
git checkout -b chore/release-v1.0.0
git add CHANGELOG.md
git commit -m "chore(release): prepare v1.0.0"
git push -u origin chore/release-v1.0.0

# 4. Abrir PR, revisar y mergear a master
gh pr create --base master --title "chore(release): prepare v1.0.0"

# 5. Una vez mergeado, crear y pushear la tag desde master
git checkout master && git pull origin master
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions construye y publica el release automáticamente
gh run watch  # ver el progreso
```

### Verificar el release

```bash
# Ver el release publicado
gh release view v1.0.0

# Descargar y probar el binario
gh release download v1.0.0 --pattern "pkgsh_linux_amd64"
chmod +x pkgsh_linux_amd64
./pkgsh_linux_amd64 --version
```

---

## 8. Checklist inicial

Completar estos pasos una sola vez al configurar el repositorio:

### GitHub Settings

- [ ] **Ruleset `protect-master`** creado y activo en `master` (ver sección 1)
- [ ] **Actions permissions** configuradas (ver sección 2)
- [ ] **Workflow permissions:** Read and write
- [ ] Verificar que el primer run de CI pasa

### Archivos del repositorio

- [ ] `.github/CODEOWNERS` creado con el owner
- [ ] `.github/pull_request_template.md` presente
- [ ] `.github/ISSUE_TEMPLATE/bug_report.md` presente
- [ ] `.github/ISSUE_TEMPLATE/feature_request.md` presente
- [ ] `CONTRIBUTING.md` en la raíz
- [ ] `CHANGELOG.md` con la versión inicial

### Primer release de prueba

- [ ] Hacer un tag `v0.1.0` y verificar que el pipeline genera los binarios
- [ ] Verificar que el `.deb` se genera correctamente
- [ ] Verificar que los checksums `SHA256SUMS.txt` están en el release

---

## Referencia rápida — comandos frecuentes

```bash
# Ver estado de todos los workflows
gh run list

# Ver PRs abiertos
gh pr list

# Ver issues abiertos
gh issue list

# Mergear un PR (desde master, solo maintainer)
gh pr merge <número> --squash --delete-branch

# Ver un release específico
gh release view v1.0.0

# Listar releases
gh release list
```
