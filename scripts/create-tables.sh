#!/bin/bash

# Script para crear tablas de DynamoDB Local
echo "Creating DynamoDB tables..."

# Esperar a que DynamoDB esté listo
sleep 5

# Crear tabla suppliers
aws dynamodb create-table \
  --table-name suppliers \
  --attribute-definitions \
    AttributeName=proveedor_id,AttributeType=S \
    AttributeName=estado_proveedor,AttributeType=S \
  --key-schema \
    AttributeName=proveedor_id,KeyType=HASH \
  --global-secondary-indexes \
    IndexName=estado-index,KeySchema='[{AttributeName=estado_proveedor,KeyType=HASH}]',Projection='{ProjectionType=ALL}',ProvisionedThroughput='{ReadCapacityUnits=5,WriteCapacityUnits=5}' \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table suppliers already exists"

# Crear tabla audit_traces
aws dynamodb create-table \
  --table-name audit_traces \
  --attribute-definitions \
    AttributeName=traza_id,AttributeType=S \
    AttributeName=proveedor_id,AttributeType=S \
    AttributeName=fecha_cambio,AttributeType=S \
  --key-schema \
    AttributeName=traza_id,KeyType=HASH \
  --global-secondary-indexes \
    IndexName=proveedor-fecha-index,KeySchema='[{AttributeName=proveedor_id,KeyType=HASH},{AttributeName=fecha_cambio,KeyType=RANGE}]',Projection='{ProjectionType=ALL}',ProvisionedThroughput='{ReadCapacityUnits=5,WriteCapacityUnits=5}' \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table audit_traces already exists"

# Crear tabla orders
aws dynamodb create-table \
  --table-name orders \
  --attribute-definitions \
    AttributeName=orden_id,AttributeType=S \
    AttributeName=estado_orden,AttributeType=S \
    AttributeName=proveedor_id,AttributeType=S \
    AttributeName=fecha_generacion,AttributeType=S \
  --key-schema \
    AttributeName=orden_id,KeyType=HASH \
  --global-secondary-indexes \
    IndexName=estado-index,KeySchema='[{AttributeName=estado_orden,KeyType=HASH}]',Projection='{ProjectionType=ALL}',ProvisionedThroughput='{ReadCapacityUnits=5,WriteCapacityUnits=5}' \
    IndexName=proveedor-fecha-index,KeySchema='[{AttributeName=proveedor_id,KeyType=HASH},{AttributeName=fecha_generacion,KeyType=RANGE}]',Projection='{ProjectionType=ALL}',ProvisionedThroughput='{ReadCapacityUnits=5,WriteCapacityUnits=5}' \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table orders already exists"

# Crear tabla products
aws dynamodb create-table \
  --table-name products \
  --attribute-definitions \
    AttributeName=producto_id,AttributeType=S \
    AttributeName=stock_actual,AttributeType=N \
  --key-schema \
    AttributeName=producto_id,KeyType=HASH \
  --global-secondary-indexes \
    IndexName=stock-index,KeySchema='[{AttributeName=stock_actual,KeyType=HASH}]',Projection='{ProjectionType=ALL}',ProvisionedThroughput='{ReadCapacityUnits=5,WriteCapacityUnits=5}' \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --endpoint-url http://localhost:8000 \
  --region us-east-1 || echo "Table products already exists"

echo "All tables created successfully!"
