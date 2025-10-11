#!/bin/bash

# Script para limpiar archivos de imagen Docker del historial de git
# Uso: ./scripts/cleanup-git.sh

set -e

echo "=== Limpiando Archivos de Imagen Docker del Git ==="
echo ""

# Verificar que estamos en un repositorio git
if [ ! -d ".git" ]; then
    echo "Error: No estás en un repositorio git"
    exit 1
fi

# Verificar que no hay cambios sin commitear
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Hay cambios sin commitear. Por favor, haz commit o stash de los cambios primero."
    git status --short
    exit 1
fi

echo "=== Archivos a limpiar ==="
echo "Buscando archivos .tar en el historial de git..."

# Buscar archivos .tar en el historial
TAR_FILES=$(git log --name-only --pretty=format: | grep -E '\.tar$' | sort | uniq)

if [ -z "$TAR_FILES" ]; then
    echo "No se encontraron archivos .tar en el historial de git."
    exit 0
fi

echo "Archivos encontrados:"
echo "$TAR_FILES"
echo ""

# Confirmar acción
read -p "¿Estás seguro de que quieres eliminar estos archivos del historial de git? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operación cancelada."
    exit 0
fi

echo ""
echo "=== Eliminando archivos del historial ==="

# Eliminar archivos del historial usando git filter-branch
for file in $TAR_FILES; do
    echo "Eliminando $file del historial..."
    git filter-branch --force --index-filter \
        "git rm --cached --ignore-unmatch '$file'" \
        --prune-empty --tag-name-filter cat -- --all
done

echo ""
echo "=== Limpiando referencias ==="
# Limpiar referencias
git for-each-ref --format="delete %(refname)" refs/original | git update-ref --stdin
git reflog expire --expire=now --all
git gc --prune=now --aggressive

echo ""
echo "=== Verificando limpieza ==="
# Verificar que los archivos fueron eliminados
REMAINING_FILES=$(git log --name-only --pretty=format: | grep -E '\.tar$' | sort | uniq)

if [ -z "$REMAINING_FILES" ]; then
    echo "✅ Limpieza completada exitosamente. No quedan archivos .tar en el historial."
else
    echo "⚠️  Aún quedan algunos archivos .tar en el historial:"
    echo "$REMAINING_FILES"
fi

echo ""
echo "=== Instrucciones adicionales ==="
echo "1. Si trabajas con otros desarrolladores, necesitarán hacer:"
echo "   git fetch origin"
echo "   git reset --hard origin/main"
echo ""
echo "2. Para subir los cambios:"
echo "   git push origin --force --all"
echo "   git push origin --force --tags"
echo ""
echo "3. Los archivos .tar ahora están excluidos por .gitignore"
echo "   y se pueden generar con: ./scripts/export-docker-images.sh"
echo ""
echo "⚠️  ADVERTENCIA: Este proceso reescribe el historial de git."
echo "   Asegúrate de que todos los colaboradores estén informados."
