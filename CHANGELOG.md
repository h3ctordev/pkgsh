# Changelog

All notable changes to pkgsh will be documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Campo `Path` en `domain.Package` para almacenar la ruta de instalación
- Panel de lista con 4 columnas: checkbox+nombre, versión, gestor, indicador de actualización
- Header y separador en el panel de lista para mayor legibilidad
- Campo **Ruta** en el panel de detalle del paquete
- Soporte para salir con `Ctrl+C`

### Changed
- Panel de lista ocupa 2/3 del ancho de la terminal (antes 40%)
- Panel de detalle con padding superior para mejor respiración visual

### Added
- Estructura base del proyecto
- Diseño de arquitectura en cuatro capas
- Spec de interfaz TUI

---

## [0.1.0] - 2026-04-06

### Added
- Scaffolding inicial del proyecto
- Definición de la interfaz `PackageManager`
- Tipos de dominio: `Package`, `Operation`, `AppState`
- Estructura de directorios definitiva
- Pipeline de CI/CD con GitHub Actions
- Configuración de empaquetado `.deb` con nfpm
