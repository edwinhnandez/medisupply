#!/bin/bash

# Script para cargar imágenes de Docker a Amazon ECR
# Uso: ./scripts/push-to-ecr.sh <aws-region> <ecr-repository-prefix>

set -e

# Verificar parámetros
if [ $# -ne 2 ]; then
    echo "Uso: $0 <aws-region> <ecr-repository-prefix>"
    echo "Ejemplo: $0 us-east-1 mediplus"
    exit 1
fi

AWS_REGION=$1
ECR_PREFIX=$2
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "=== Configuración ==="
echo "AWS Region: $AWS_REGION"
echo "ECR Prefix: $ECR_PREFIX"
echo "AWS Account ID: $AWS_ACCOUNT_ID"
echo ""

# URLs de ECR
SUPPLIER_ECR_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_PREFIX}-supplier-service"
ORDER_ECR_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_PREFIX}-purchase-order-service"

echo "=== URLs de ECR ==="
echo "Supplier Service: $SUPPLIER_ECR_URI"
echo "Purchase Order Service: $ORDER_ECR_URI"
echo ""

# Autenticar Docker con ECR
echo "=== Autenticando Docker con ECR ==="
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

# Crear repositorios ECR si no existen
echo "=== Creando repositorios ECR ==="
aws ecr create-repository --repository-name ${ECR_PREFIX}-supplier-service --region $AWS_REGION 2>/dev/null || echo "Repositorio supplier-service ya existe"
aws ecr create-repository --repository-name ${ECR_PREFIX}-purchase-order-service --region $AWS_REGION 2>/dev/null || echo "Repositorio purchase-order-service ya existe"

# Etiquetar imágenes para ECR
echo "=== Etiquetando imágenes para ECR ==="
docker tag medisupply-supplier-service:latest $SUPPLIER_ECR_URI:latest
docker tag medisupply-purchase-order-service:latest $ORDER_ECR_URI:latest

# También etiquetar con timestamp para versionado
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
docker tag medisupply-supplier-service:latest $SUPPLIER_ECR_URI:$TIMESTAMP
docker tag medisupply-purchase-order-service:latest $ORDER_ECR_URI:$TIMESTAMP

echo "=== Subiendo imágenes a ECR ==="
echo "Subiendo supplier-service..."
docker push $SUPPLIER_ECR_URI:latest
docker push $SUPPLIER_ECR_URI:$TIMESTAMP

echo "Subiendo purchase-order-service..."
docker push $ORDER_ECR_URI:latest
docker push $ORDER_ECR_URI:$TIMESTAMP

echo ""
echo "=== Completado ==="
echo "Imágenes subidas exitosamente a ECR:"
echo "- Supplier Service: $SUPPLIER_ECR_URI:latest"
echo "- Supplier Service: $SUPPLIER_ECR_URI:$TIMESTAMP"
echo "- Purchase Order Service: $ORDER_ECR_URI:latest"
echo "- Purchase Order Service: $ORDER_ECR_URI:$TIMESTAMP"
echo ""
echo "Para usar estas imágenes en Kubernetes, actualiza los deployments con:"
echo "image: $SUPPLIER_ECR_URI:latest"
echo "image: $ORDER_ECR_URI:latest"
