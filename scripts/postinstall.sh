#!/bin/sh
set -e

# Verifica presencia de gestores del sistema al instalar el .deb.
# pkgsh funciona igual si alguno no está instalado — simplemente no muestra esos paquetes.

for cmd in sudo apt dpkg; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "pkgsh: advertencia — '$cmd' no encontrado. Algunas funciones no estarán disponibles."
    fi
done

echo "pkgsh instalado correctamente. Ejecuta: pkgsh"
