#!/bin/bash

# Script para exportar imágenes de Docker para distribución
# Uso: ./scripts/export-docker-images.sh

set -e

echo "=== Exportando Imágenes de Docker para MediPlus ==="
echo ""

# Verificar que Docker esté funcionando
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker no está funcionando. Por favor, inicia Docker y vuelve a intentar."
    exit 1
fi

# Crear directorio de exportación
mkdir -p docker-images

# Construir las imágenes si no existen
echo "=== Construyendo Imágenes ==="
echo "Construyendo supplier-service..."
docker-compose build supplier-service

echo "Construyendo purchase-order-service..."
docker-compose build purchase-order-service

# Exportar imágenes
echo ""
echo "=== Exportando Imágenes ==="
echo "Exportando supplier-service..."
docker save medisupply-supplier-service:latest -o docker-images/medisupply-supplier-service.tar

echo "Exportando purchase-order-service..."
docker save medisupply-purchase-order-service:latest -o docker-images/medisupply-purchase-order-service.tar

# Comprimir las imágenes para reducir tamaño
echo ""
echo "=== Comprimiendo Imágenes ==="
echo "Comprimiendo supplier-service..."
gzip -c docker-images/medisupply-supplier-service.tar > docker-images/medisupply-supplier-service.tar.gz

echo "Comprimiendo purchase-order-service..."
gzip -c docker-images/medisupply-purchase-order-service.tar > docker-images/medisupply-purchase-order-service.tar.gz

# Mostrar información de los archivos
echo ""
echo "=== Archivos Generados ==="
echo "Imágenes sin comprimir:"
ls -lh docker-images/*.tar

echo ""
echo "Imágenes comprimidas:"
ls -lh docker-images/*.tar.gz

echo ""
echo "=== Instrucciones de Uso ==="
echo "1. Para cargar las imágenes en otro sistema:"
echo "   gunzip -c docker-images/medisupply-supplier-service.tar.gz | docker load"
echo "   gunzip -c docker-images/medisupply-purchase-order-service.tar.gz | docker load"
echo ""
echo "2. Para subir a ECR:"
echo "   ./scripts/load-and-push-to-ecr.sh <region> <prefix>"
echo ""
echo "3. Los archivos .tar y .tar.gz están en el directorio docker-images/"
echo "   (Este directorio está excluido del control de versiones)"
