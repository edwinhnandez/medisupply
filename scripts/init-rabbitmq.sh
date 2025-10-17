#!/bin/bash

# Script para inicializar RabbitMQ con las configuraciones necesarias
# Este script se ejecuta cuando RabbitMQ está listo

echo "Inicializando RabbitMQ para MediPlus..."

# Esperar a que RabbitMQ esté listo
until rabbitmq-diagnostics ping > /dev/null 2>&1; do
    echo "Esperando a que RabbitMQ esté listo..."
    sleep 2
done

echo "RabbitMQ está listo. Configurando exchanges y colas..."

# Crear exchanges
rabbitmqadmin declare exchange name=supplier.events type=topic durable=true
rabbitmqadmin declare exchange name=notifications.events type=topic durable=true
rabbitmqadmin declare exchange name=stock.events type=topic durable=true
rabbitmqadmin declare exchange name=order.events type=topic durable=true
rabbitmqadmin declare exchange name=external.events type=topic durable=true

# Crear colas para supplier-service
rabbitmqadmin declare queue name=proveedor.events durable=true
rabbitmqadmin declare queue name=proveedor.audit durable=true
rabbitmqadmin declare queue name=proveedor.evaluation durable=true

# Crear colas para purchase-order-service
rabbitmqadmin declare queue name=order.events durable=true
rabbitmqadmin declare queue name=stock.events durable=true
rabbitmqadmin declare queue name=purchase-order-external-events durable=true

# Crear colas para notificaciones
rabbitmqadmin declare queue name=notifications durable=true

# Bind colas a exchanges
rabbitmqadmin declare binding source=supplier.events destination=proveedor.events routing_key="#"
rabbitmqadmin declare binding source=supplier.events destination=proveedor.audit routing_key="proveedor.*"
rabbitmqadmin declare binding source=supplier.events destination=proveedor.evaluation routing_key="proveedor.evaluacion.*"

rabbitmqadmin declare binding source=order.events destination=order.events routing_key="#"
rabbitmqadmin declare binding source=stock.events destination=stock.events routing_key="#"
rabbitmqadmin declare binding source=external.events destination=purchase-order-external-events routing_key="#"
rabbitmqadmin declare binding source=notifications.events destination=notifications routing_key="#"

echo "Configuración de RabbitMQ completada exitosamente!"

# Mostrar estado de las colas
echo "Estado de las colas:"
rabbitmqadmin list queues name messages consumers

echo "Estado de los exchanges:"
rabbitmqadmin list exchanges name type durable

