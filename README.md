# MediSupply - Sistema de Gestión de Aprovisionamiento

Sistema de microservicios para la gestión de aprovisionamiento médico basado en Domain-Driven Design (DDD) con event mesh y escalado automático mediante KEDA.

## Arquitectura

El sistema está compuesto por dos microservicios principales:

### 1. Supplier Management Service (Puerto 8082)
- **Responsabilidad**: Gestión de proveedores calificados, certificaciones y evaluaciones
- **Agregado Principal**: ProveedorCalificado
- **Eventos**: ProveedorCalificado, ProveedorSuspendido, CertificacionPorVencer, EvaluacionActualizada
- **Event Listeners**: Escucha eventos de órdenes de compra y genera solicitudes de proveedor automáticamente

### 2. Purchase Order Service (Puerto 8081)
- **Responsabilidad**: Gestión de órdenes de compra automáticas y procesamiento de eventos externos
- **Agregado Principal**: OrdenCompraAutomatica
- **Eventos**: OrdenCompraGenerada, OrdenCompraEnviada, OrdenCompraConfirmada, OrdenCompraRecibida

- **Event Listeners**: Escucha eventos de stock bajo y genera órdenes automáticamente

## Tecnologías Utilizadas

- **Lenguaje**: Go 1.23
- **Base de Datos**: Amazon DynamoDB
- **Event Mesh**: RabbitMQ
- **Orquestación**: Kubernetes
- **Escalado Automático**: KEDA
- **Contenedores**: Docker
- **API**: Gin Framework

## Estructura del Proyecto

```
medisupply/
├── supplier-service/           # Microservicio de proveedores
│   ├── main.go
│   ├── Dockerfile
│   └── internal/
│       ├── config/
│       ├── database/
│       ├── events/
│       │   ├── events.go
│       │   └── rabbitmq.go
│       ├── handlers/
│       ├── models/
│       ├── repository/
│       └── service/
├── purchase-order-service/     # Microservicio de órdenes
│   ├── main.go
│   ├── Dockerfile
│   └── internal/
│       ├── config/
│       ├── database/
│       ├── events/
│       │   ├── events.go
│       │   └── rabbitmq.go
│       ├── handlers/
│       │   ├── order_handler.go
│       │   ├── external_event_handler.go
│       │   └── external_simulator_handler.go
│       ├── models/
│       ├── repository/
│       └── service/
├── k8s/                       # Configuración de Kubernetes
│   ├── supplier-service-deployment.yaml
│   ├── purchase-order-service-deployment.yaml
│   ├── supplier-service-scaledobject.yaml
│   ├── purchase-order-service-scaledobject.yaml
│   ├── rabbitmq-cluster.yaml
│   ├── dynamodb-local.yaml
│   └── dynamodb-tables.yaml
├── scripts/                   # Scripts de inicialización
│   ├── init-rabbitmq.sh
│   └── create-tables.sh
├── docker-compose.yml          # Desarrollo local
└── go.mod
```

## Event Mesh con RabbitMQ

El sistema utiliza RabbitMQ como event mesh para la comunicación entre microservicios:

### Exchanges y Topics
- `supplier.events`: Eventos relacionados con proveedores
- `order.events`: Eventos relacionados con órdenes
- `stock.events`: Eventos relacionados con stock
- `notifications.events`: Eventos de notificaciones
- `external.events`: Eventos desde sistemas externos

### Tipos de Eventos

#### Supplier Service
- `proveedor.calificado`: Proveedor calificado
- `proveedor.suspendido`: Proveedor suspendido
- `proveedor.activado`: Proveedor activado
- `certificacion.por_vencer`: Certificación por vencer
- `evaluacion.actualizada`: Evaluación actualizada
- `solicitud.proveedor`: Solicitud de proveedor generada automáticamente

#### Purchase Order Service
- `orden.generada`: Orden de compra generada
- `orden.confirmada`: Orden de compra confirmada
- `orden.recibida`: Orden de compra recibida
- `stock.bajo`: Stock bajo punto de reorden
- `stock.lote_danado`: Lote dañado por temperatura
- `stock.demanda_alta`: Pronóstico de alta demanda

#### Eventos Externos (Nuevos)
- `external.stock.bajo`: Stock bajo detectado por sistema externo
- `external.demanda.alta`: Demanda alta pronosticada por sistema externo
- `external.lote.danado`: Lote dañado detectado por sistema externo
- `external.alerta.inventario`: Alerta de inventario desde sistema externo

## Funcionalidad de Eventos Externos

### Descripción
El sistema ahora puede recibir eventos desde sistemas externos (como sistemas de inventario hospitalarios, sensores de temperatura, sistemas de pronóstico de demanda) y generar órdenes de compra automáticamente.

### Tipos de Eventos Externos Soportados

#### 1. Stock Bajo Externo
- **Trigger**: Sistema de inventario detecta stock por debajo del punto de reorden
- **Acción**: Genera orden automática con prioridad ALTA
- **Datos**: Producto, stock actual, punto reorden, cantidad requerida

#### 2. Demanda Alta Externa
- **Trigger**: Sistema de pronóstico predice alta demanda
- **Acción**: Genera orden automática solo si confianza >= 80%
- **Datos**: Producto, demanda pronosticada, confianza, período

#### 3. Lote Dañado Externo
- **Trigger**: Sistema de monitoreo detecta lote dañado
- **Acción**: Genera orden de reposición con prioridad ALTA
- **Datos**: Producto, lote, cantidad dañada, temperatura, motivo

#### 4. Alerta de Inventario Externa
- **Trigger**: Sistema de gestión de inventario emite alerta
- **Acción**: Genera orden basada en tipo de alerta
- **Datos**: Producto, tipo alerta, descripción, stock actual

### Flujo de Procesamiento de Eventos Externos

1. **Sistema Externo** publica evento en exchange `external.events`
2. **RabbitMQ** enruta evento a cola `purchase-order-external-events`
3. **Purchase Order Service** recibe y procesa evento
4. **ExternalEventHandler** analiza evento y determina acción
5. **OrderService** crea orden automáticamente
6. **EventBus** publica evento `OrdenCompraGenerada`

### Lógica de Decisión Inteligente

- **Filtrado por confianza**: Solo procesa eventos de demanda alta con confianza >= 80%
- **Mapeo de prioridades**: Convierte prioridades externas a enums internos
- **Generación de números únicos**: Crea números de orden con prefijos específicos
- **Trazabilidad completa**: Registra origen del evento en motivo de generación

## Escalado Automático con KEDA

El sistema utiliza KEDA para el escalado automático basado en:

### Métricas de Escalado
- **RabbitMQ Queue Length**: Número de mensajes pendientes en colas
- **CPU Utilization**: Uso de CPU del pod
- **Memory Utilization**: Uso de memoria del pod
- **Prometheus Metrics**: Métricas personalizadas

### Configuración de Escalado
- **Mínimo de réplicas**: 1
- **Máximo de réplicas**: 10-15 (dependiendo del servicio)
- **Intervalo de polling**: 30 segundos
- **Período de cooldown**: 300 segundos

## Base de Datos DynamoDB

### Tablas Principales

#### suppliers
- **Clave primaria**: proveedor_id (String)
- **GSI**: estado-index (estado_proveedor)
- **Atributos**: nombre_legal, razon_social, estado_proveedor, certificaciones, etc.

#### audit_traces
- **Clave primaria**: traza_id (String)
- **GSI**: proveedor-fecha-index (proveedor_id, fecha_cambio)
- **Atributos**: proveedor_id, tipo_cambio, descripcion, etc.

#### orders
- **Clave primaria**: orden_id (String)
- **GSI**: estado-index (estado_orden)
- **GSI**: proveedor-fecha-index (proveedor_id, fecha_generacion)
- **Atributos**: numero_orden, proveedor_id, estado_orden, items, etc.

#### products
- **Clave primaria**: producto_id (String)
- **GSI**: stock-index (stock_actual)
- **Atributos**: nombre, stock_actual, punto_reorden, condiciones, etc.

## Desarrollo Local

### Prerrequisitos
- Docker y Docker Compose
- Go 1.23+
- AWS CLI (para DynamoDB local)

### Nota sobre Imágenes Docker
Las imágenes Docker exportadas (.tar) exceden el límite de 100MB de GitHub y no están incluidas en el repositorio. Para generar las imágenes cuando sea necesario, usa:

```bash
# Generar imágenes de Docker
./scripts/export-docker-images.sh
```

Ver `DOCKER_IMAGES.md` para más detalles sobre el manejo de imágenes.

### Ejecutar el Sistema

1. **Clonar el repositorio**
```bash
git clone <repository-url>
cd medisupply
```

2. **Ejecutar con Docker Compose**
```bash
docker-compose up -d
```

3. **Verificar servicios**
```bash
# Health checks
curl http://localhost:8082/health  # Supplier Service
curl http://localhost:8081/health  # Purchase Order Service

# RabbitMQ Management UI
# http://localhost:15672 (usuario: mediplus, contraseña: mediplus123)
```

### APIs Disponibles

#### Supplier Service (Puerto 8082)
- `GET /api/v1/suppliers` - Listar proveedores
- `POST /api/v1/suppliers` - Crear proveedor
- `GET /api/v1/suppliers/:id` - Obtener proveedor
- `PUT /api/v1/suppliers/:id` - Actualizar proveedor
- `DELETE /api/v1/suppliers/:id` - Eliminar proveedor
- `POST /api/v1/suppliers/:id/evaluate` - Evaluar proveedor
- `POST /api/v1/suppliers/:id/suspend` - Suspender proveedor
- `POST /api/v1/suppliers/:id/activate` - Activar proveedor

**Event Listeners:**
- Escucha `orden.generada` → Genera `solicitud.proveedor`
- Escucha `orden.confirmada` → Registra confirmación en auditoría
- Escucha `orden.recibida` → Registra recepción en auditoría

#### Purchase Order Service (Puerto 8081)
- `GET /api/v1/orders` - Listar órdenes
- `POST /api/v1/orders` - Crear orden
- `GET /api/v1/orders/:id` - Obtener orden
- `PUT /api/v1/orders/:id` - Actualizar orden
- `DELETE /api/v1/orders/:id` - Eliminar orden
- `POST /api/v1/orders/:id/confirm` - Confirmar orden
- `POST /api/v1/orders/:id/receive` - Marcar como recibida
- `POST /api/v1/orders/auto-generate` - Generar orden automáticamente

#### APIs de Eventos Externos (Puerto 8081)
- `GET /api/v1/external/event-types` - Listar tipos de eventos externos disponibles
- `POST /api/v1/external/simulate/stock-bajo` - Simular evento de stock bajo
- `POST /api/v1/external/simulate/demanda-alta` - Simular evento de demanda alta
- `POST /api/v1/external/simulate/lote-danado` - Simular evento de lote dañado
- `POST /api/v1/external/simulate/alerta-inventario` - Simular alerta de inventario

### Ejemplos de Uso de Eventos Externos

#### Simular Stock Bajo
```bash
curl -X POST http://localhost:8081/api/v1/external/simulate/stock-bajo \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "PROD-001",
    "nombre_producto": "Vacuna COVID-19",
    "stock_actual": 5,
    "punto_reorden": 20,
    "stock_maximo": 100,
    "cantidad_requerida": 50,
    "prioridad": "ALTA",
    "urgencia": "ALTA",
    "source": "Sistema de Inventario Hospital Central"
  }'
```

#### Simular Demanda Alta
```bash
curl -X POST http://localhost:8081/api/v1/external/simulate/demanda-alta \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "PROD-002",
    "nombre_producto": "Insulina",
    "demanda_pronosticada": 200,
    "stock_actual": 30,
    "cantidad_requerida": 100,
    "confianza_pronostico": 0.85,
    "periodo_pronostico": "30 días",
    "prioridad": "MEDIA",
    "source": "Sistema de Pronóstico de Demanda"
  }'
```

#### Simular Lote Dañado
```bash
curl -X POST http://localhost:8081/api/v1/external/simulate/lote-danado \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "PROD-003",
    "nombre_producto": "Vacuna Influenza",
    "lote_id": "LOTE-2025-001",
    "cantidad_danada": 25,
    "temperatura_registrada": 8.5,
    "temperatura_requerida": 2.0,
    "motivo_danio": "Temperatura fuera de rango",
    "urgencia": "ALTA",
    "source": "Sistema de Monitoreo de Temperatura"
  }'
```

## Despliegue en Kubernetes

### Prerrequisitos
- Kubernetes cluster
- KEDA instalado
- RabbitMQ cluster configurado

### Desplegar el Sistema

1. **Aplicar configuraciones**
```bash
kubectl apply -f k8s/rabbitmq-cluster.yaml
kubectl apply -f k8s/dynamodb-local.yaml
kubectl apply -f k8s/dynamodb-tables.yaml
kubectl apply -f k8s/supplier-service-deployment.yaml
kubectl apply -f k8s/purchase-order-service-deployment.yaml
kubectl apply -f k8s/supplier-service-scaledobject.yaml
kubectl apply -f k8s/purchase-order-service-scaledobject.yaml
```

2. **Verificar despliegue**
```bash
kubectl get pods
kubectl get scaledobjects
```

## Monitoreo y Observabilidad

### Métricas de KEDA
- Queue length de RabbitMQ
- CPU y memoria de pods
- Métricas personalizadas de Prometheus

### Health Checks
- Endpoints `/health` en ambos servicios
- Probes de liveness y readiness configurados

### Logging
- Logs estructurados en formato JSON
- Niveles de log configurables
- Integración con sistemas de logging centralizados
- Trazabilidad completa de eventos externos

## Eventos de Negocio

### Flujo de Eventos Típico

1. **Stock Bajo** → `stock.bajo` event → **Purchase Order Service escucha** → Auto-generación de orden
2. **Lote Dañado** → `lote.danado` event → **Purchase Order Service escucha** → Auto-generación de orden
3. **Pronóstico Alta Demanda** → `pronostico.demanda_alta` event → **Purchase Order Service escucha** → Auto-generación de orden
4. **Orden Generada** → `orden_compra.generada` event → Notificaciones
5. **Orden Enviada** → `orden_compra.enviada` event → Seguimiento
6. **Orden Confirmada** → `orden_compra.confirmada` event → Actualización de estado
7. **Orden Recibida** → `orden_compra.recibida` event → Finalización del proceso

### Event-Driven Order Generation

El **Purchase Order Service** ahora escucha automáticamente los siguientes eventos y genera órdenes de compra:

- **`stock.bajo`**: Cuando el stock de un producto está por debajo del punto de reorden
- **`stock.lote_danado`**: Cuando se detecta un lote dañado por temperatura
- **`stock.demanda_alta`**: Cuando se pronostica alta demanda para un producto

#### Lógica de Generación Automática:
- **Prioridad Inteligente**: 
  - Stock = 0 → Prioridad CRÍTICA
  - Stock ≤ PuntoReorden/2 → Prioridad ALTA
  - Stock ≤ PuntoReorden → Prioridad MEDIA
- **Cantidad Calculada**: Hasta el stock máximo del producto
- **Prevención de Duplicados**: Verifica que no exista una orden pendiente para el mismo producto

### Event-Driven Supplier Requests

El **Supplier Service** ahora escucha automáticamente los siguientes eventos de órdenes y genera solicitudes de proveedor:

- **`orden.generada`**: Cuando se crea una nueva orden de compra
- **`orden.confirmada`**: Cuando un proveedor confirma una orden
- **`orden.recibida`**: Cuando se recibe la entrega de una orden

#### Lógica de Solicitud de Proveedor:
- **Requisitos Especiales**: Basados en la prioridad de la orden
  - `CRITICA`: Entrega urgente, respuesta 24/7, certificaciones médicas vigentes
  - `ALTA`: Entrega rápida, certificaciones médicas vigentes
  - `MEDIA`: Certificaciones médicas vigentes
  - `BAJA`: Certificaciones básicas
- **Productos Requeridos**: Información detallada de productos y cantidades
- **Auditoría**: Trazabilidad completa de eventos de órdenes
- **Evento Generado**: `solicitud.proveedor` con todos los requisitos

## Consideraciones de Escalabilidad

### Escalado Horizontal
- KEDA maneja el escalado automático basado en métricas
- Cada microservicio puede escalar independientemente
- RabbitMQ distribuye la carga entre instancias

### Escalado Vertical
- Límites de CPU y memoria configurados
- Recursos ajustables según necesidades

### Persistencia
- DynamoDB como base de datos principal
- Eventos persistentes en RabbitMQ
- Backup y recuperación automatizados

## Seguridad

### Autenticación y Autorización
- Endpoints protegidos (implementar según necesidades)
- Tokens JWT para autenticación
- Roles y permisos por servicio

### Comunicación Segura
- TLS para comunicación entre servicios
- Encriptación de datos en tránsito
- Secrets management en Kubernetes

## Integración con Sistemas Externos

### Sistemas Soportados
- Sistemas de inventario hospitalarios
- Sistemas de pronóstico de demanda
- Sensores de temperatura y monitoreo
- Sistemas de gestión de cadena de frío
- Sistemas de alertas de inventario

### Protocolos de Comunicación
- RabbitMQ AMQP para eventos
- REST APIs para simulación y testing
- Webhooks para integración directa

### Configuración de Integración
- Credenciales de RabbitMQ configurables
- Routing keys personalizables
- Filtros de eventos configurables

## Contribución

1. Fork el repositorio
2. Crear una rama para la feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit los cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear un Pull Request

## Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.