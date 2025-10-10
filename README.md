# MediPlus - Sistema de Gestión de Aprovisionamiento

Sistema de microservicios para la gestión de aprovisionamiento médico basado en Domain-Driven Design (DDD) con event mesh y escalado automático mediante KEDA.

## Arquitectura

El sistema está compuesto por dos microservicios principales:

### 1. Supplier Management Service (Puerto 8080)
- **Responsabilidad**: Gestión de proveedores calificados, certificaciones y evaluaciones
- **Agregado Principal**: ProveedorCalificado
- **Eventos**: ProveedorCalificado, ProveedorSuspendido, CertificacionPorVencer, EvaluacionActualizada

### 2. Purchase Order Service (Puerto 8081)
- **Responsabilidad**: Gestión de órdenes de compra automáticas
- **Agregado Principal**: OrdenCompraAutomatica
- **Eventos**: OrdenCompraGenerada, OrdenCompraEnviada, OrdenCompraConfirmada, OrdenCompraRecibida

## Tecnologías Utilizadas

- **Lenguaje**: Go 1.21
- **Base de Datos**: Amazon DynamoDB
- **Event Mesh**: NATS
- **Orquestación**: Kubernetes
- **Escalado Automático**: KEDA
- **Contenedores**: Docker
- **API**: Gin Framework

## Estructura del Proyecto

```
mediplus/
├── supplier-service/           # Microservicio de proveedores
│   ├── main.go
│   ├── Dockerfile
│   └── internal/
│       ├── config/
│       ├── database/
│       ├── events/
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
│       ├── handlers/
│       ├── models/
│       ├── repository/
│       └── service/
├── k8s/                       # Configuración de Kubernetes
│   ├── supplier-service-deployment.yaml
│   ├── purchase-order-service-deployment.yaml
│   ├── supplier-service-scaledobject.yaml
│   ├── purchase-order-service-scaledobject.yaml
│   ├── nats-cluster.yaml
│   ├── dynamodb-local.yaml
│   └── dynamodb-tables.yaml
├── docker-compose.yml          # Desarrollo local
└── go.mod
```

## Event Mesh

El sistema utiliza NATS como event mesh para la comunicación entre microservicios:

### Topics de Eventos
- `supplier.events`: Eventos relacionados con proveedores
- `order.events`: Eventos relacionados con órdenes
- `stock.events`: Eventos relacionados con stock
- `notifications.events`: Eventos de notificaciones

### Tipos de Eventos

#### Supplier Service
- `proveedor.calificado`: Proveedor calificado
- `proveedor.suspendido`: Proveedor suspendido
- `proveedor.activado`: Proveedor activado
- `certificacion.por_vencer`: Certificación por vencer
- `evaluacion.actualizada`: Evaluación actualizada

#### Purchase Order Service
- `orden_compra.generada`: Orden de compra generada
- `orden_compra.enviada`: Orden de compra enviada
- `orden_compra.confirmada`: Orden de compra confirmada
- `orden_compra.recibida`: Orden de compra recibida
- `stock.bajo`: Stock bajo punto de reorden
- `lote.danado`: Lote dañado por temperatura
- `pronostico.demanda_alta`: Pronóstico de alta demanda

## Escalado Automático con KEDA

El sistema utiliza KEDA para el escalado automático basado en:

### Métricas de Escalado
- **NATS Queue Length**: Número de mensajes pendientes en colas
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
- **Atributos**: nombre_legal, razon_social, estado_proveedor, etc.

#### audit_traces
- **Clave primaria**: traza_id (String)
- **GSI**: proveedor-fecha-index (proveedor_id, fecha_cambio)
- **Atributos**: proveedor_id, tipo_cambio, descripcion, etc.

#### orders
- **Clave primaria**: orden_id (String)
- **GSI**: estado-index (estado_orden)
- **GSI**: proveedor-fecha-index (proveedor_id, fecha_generacion)
- **Atributos**: numero_orden, proveedor_id, estado_orden, etc.

#### products
- **Clave primaria**: producto_id (String)
- **GSI**: stock-index (stock_actual)
- **Atributos**: nombre, stock_actual, punto_reorden, etc.

## Desarrollo Local

### Prerrequisitos
- Docker y Docker Compose
- Go 1.21+
- AWS CLI (para DynamoDB local)

### Ejecutar el Sistema

1. **Clonar el repositorio**
```bash
git clone <repository-url>
cd mediplus
```

2. **Ejecutar con Docker Compose**
```bash
docker-compose up -d
```

3. **Verificar servicios**
```bash
# Health checks
curl http://localhost:8080/health  # Supplier Service
curl http://localhost:8081/health  # Purchase Order Service
```

### APIs Disponibles

#### Supplier Service (Puerto 8080)
- `GET /api/v1/suppliers` - Listar proveedores
- `POST /api/v1/suppliers` - Crear proveedor
- `GET /api/v1/suppliers/:id` - Obtener proveedor
- `PUT /api/v1/suppliers/:id` - Actualizar proveedor
- `DELETE /api/v1/suppliers/:id` - Eliminar proveedor
- `POST /api/v1/suppliers/:id/evaluate` - Evaluar proveedor
- `POST /api/v1/suppliers/:id/suspend` - Suspender proveedor
- `POST /api/v1/suppliers/:id/activate` - Activar proveedor

#### Purchase Order Service (Puerto 8081)
- `GET /api/v1/orders` - Listar órdenes
- `POST /api/v1/orders` - Crear orden
- `GET /api/v1/orders/:id` - Obtener orden
- `PUT /api/v1/orders/:id` - Actualizar orden
- `DELETE /api/v1/orders/:id` - Eliminar orden
- `POST /api/v1/orders/:id/confirm` - Confirmar orden
- `POST /api/v1/orders/:id/receive` - Marcar como recibida
- `POST /api/v1/orders/auto-generate` - Generar orden automáticamente

## Despliegue en Kubernetes

### Prerrequisitos
- Kubernetes cluster
- KEDA instalado
- NATS cluster configurado

### Desplegar el Sistema

1. **Aplicar configuraciones**
```bash
kubectl apply -f k8s/nats-cluster.yaml
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
- Queue length de NATS
- CPU y memoria de pods
- Métricas personalizadas de Prometheus

### Health Checks
- Endpoints `/health` en ambos servicios
- Probes de liveness y readiness configurados

### Logging
- Logs estructurados en formato JSON
- Niveles de log configurables
- Integración con sistemas de logging centralizados

## Eventos de Negocio

### Flujo de Eventos Típico

1. **Stock Bajo** → `stock.bajo` event → Auto-generación de orden
2. **Lote Dañado** → `lote.danado` event → Auto-generación de orden
3. **Pronóstico Alta Demanda** → `pronostico.demanda_alta` event → Auto-generación de orden
4. **Orden Generada** → `orden_compra.generada` event → Notificaciones
5. **Orden Enviada** → `orden_compra.enviada` event → Seguimiento
6. **Orden Confirmada** → `orden_compra.confirmada` event → Actualización de estado
7. **Orden Recibida** → `orden_compra.recibida` event → Finalización del proceso

## Consideraciones de Escalabilidad

### Escalado Horizontal
- KEDA maneja el escalado automático basado en métricas
- Cada microservicio puede escalar independientemente
- NATS distribuye la carga entre instancias

### Escalado Vertical
- Límites de CPU y memoria configurados
- Recursos ajustables según necesidades

### Persistencia
- DynamoDB como base de datos principal
- Eventos persistentes en NATS JetStream
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

## Contribución

1. Fork el repositorio
2. Crear una rama para la feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit los cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear un Pull Request

## Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.
