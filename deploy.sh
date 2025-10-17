#!/bin/bash

# Script de despliegue para MediPlus
# Este script despliega el sistema completo en Kubernetes

set -e

echo "Iniciando despliegue de MediPlus..."

# Verificar que kubectl está disponible
if ! command -v kubectl &> /dev/null; then
    echo "kubectl no está instalado. Por favor, instala kubectl primero."
    exit 1
fi

# Verificar que KEDA está instalado
if ! kubectl get crd scaledobjects.keda.sh &> /dev/null; then
    echo "KEDA no está instalado. Por favor, instala KEDA primero."
    echo "   Instrucciones: https://keda.sh/docs/latest/deploy/"
    exit 1
fi

echo "Desplegando RabbitMQ Cluster..."
kubectl apply -f k8s/rabbitmq-cluster.yaml

echo "Desplegando DynamoDB Local..."
kubectl apply -f k8s/dynamodb-local.yaml

echo "Esperando que DynamoDB esté listo..."
kubectl wait --for=condition=ready pod -l app=dynamodb-local --timeout=60s

echo "Creando tablas de DynamoDB..."
kubectl apply -f k8s/dynamodb-tables.yaml

echo "Esperando que las tablas se creen..."
sleep 30

echo "Desplegando Supplier Service..."
kubectl apply -f k8s/supplier-service-deployment.yaml

echo "Desplegando Purchase Order Service..."
kubectl apply -f k8s/purchase-order-service-deployment.yaml

echo "Configurando escalado automático con KEDA..."
kubectl apply -f k8s/supplier-service-scaledobject.yaml
kubectl apply -f k8s/purchase-order-service-scaledobject.yaml

echo "Esperando que los servicios estén listos..."
kubectl wait --for=condition=ready pod -l app=supplier-service --timeout=120s
kubectl wait --for=condition=ready pod -l app=purchase-order-service --timeout=120s

echo "Despliegue completado exitosamente!"

echo ""
echo "Información del despliegue:"
echo "   - Supplier Service: http://localhost:8082 (con event listeners para órdenes)"
echo "   - Purchase Order Service: http://localhost:8081 (con event listeners para stock)"
echo "   - RabbitMQ Management: http://localhost:15672"
echo "   - DynamoDB Local: http://localhost:8000"

echo ""
echo "Para verificar el estado:"
echo "   kubectl get pods"
echo "   kubectl get scaledobjects"
echo "   kubectl get services"

echo ""
echo "Para ver logs:"
echo "   kubectl logs -l app=supplier-service"
echo "   kubectl logs -l app=purchase-order-service"

echo ""
echo "Para probar los servicios:"
echo "   curl http://localhost:8082/health"
echo "   curl http://localhost:8081/health"

echo ""
echo "Funcionalidades Event-Driven:"
echo "   - Purchase Order Service escucha eventos de stock y genera órdenes automáticamente"
echo "   - Supplier Service escucha eventos de órdenes y genera solicitudes de proveedor"
echo "   - RabbitMQ maneja la comunicación entre microservicios"

echo ""
echo "¡MediPlus está listo para usar!"
