# Guía de Despliegue en AWS ECR

Esta guía explica cómo desplegar las imágenes de Docker de MediPlus en Amazon Elastic Container Registry (ECR) y usarlas en Kubernetes.

## Prerrequisitos

- AWS CLI configurado con credenciales apropiadas
- Docker instalado y funcionando
- Acceso a una cuenta de AWS con permisos para ECR
- Kubernetes cluster configurado

## Opción 1: Usar Imágenes Exportadas (Recomendado)

### 1. Verificar Archivos de Imagen

Los archivos de imagen ya están exportados en el directorio raíz:

```bash
ls -lh *.tar
# medisupply-supplier-service.tar (122M)
# medisupply-purchase-order-service.tar (122M)
```

### 2. Cargar y Subir a ECR

Usar el script automatizado:

```bash
# Ejemplo para región us-east-1 con prefijo mediplus
./scripts/load-and-push-to-ecr.sh us-east-1 mediplus
```

El script:
- Carga las imágenes desde los archivos tar
- Autentica Docker con ECR
- Crea los repositorios ECR si no existen
- Etiqueta las imágenes con URLs de ECR
- Sube las imágenes a ECR

### 3. Verificar en AWS Console

1. Ir a AWS Console > ECR
2. Verificar que los repositorios fueron creados:
   - `mediplus-supplier-service`
   - `mediplus-purchase-order-service`
3. Confirmar que las imágenes están disponibles

## Opción 2: Construir y Subir Directamente

Si prefieres construir las imágenes directamente:

```bash
# Construir las imágenes
docker-compose build

# Usar el script de push directo
./scripts/push-to-ecr.sh us-east-1 mediplus
```

## Configuración de Kubernetes

### 1. Actualizar Deployments

Actualizar los archivos de deployment en `k8s/` con las URLs de ECR:

```yaml
# supplier-service-deployment.yaml
spec:
  containers:
  - name: supplier-service
    image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/mediplus-supplier-service:latest
    # ... resto de configuración

# purchase-order-service-deployment.yaml
spec:
  containers:
  - name: purchase-order-service
    image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/mediplus-purchase-order-service:latest
    # ... resto de configuración
```

### 2. Configurar Secrets para ECR

Crear secret para autenticación con ECR:

```bash
kubectl create secret docker-registry ecr-secret \
  --docker-server=123456789012.dkr.ecr.us-east-1.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region us-east-1) \
  --docker-email=your-email@example.com
```

### 3. Actualizar Deployments con ImagePullSecrets

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
      - name: ecr-secret
      containers:
      - name: supplier-service
        image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/mediplus-supplier-service:latest
```

## Variables de Entorno para Producción

### Supplier Service

```yaml
env:
- name: PORT
  value: "8080"
- name: AWS_REGION
  value: "us-east-1"
- name: DYNAMODB_ENDPOINT
  value: ""  # Usar DynamoDB real en producción
- name: RABBITMQ_URL
  value: "amqp://user:password@rabbitmq-cluster:5672/"
- name: ENVIRONMENT
  value: "production"
```

### Purchase Order Service

```yaml
env:
- name: PORT
  value: "8081"
- name: AWS_REGION
  value: "us-east-1"
- name: DYNAMODB_ENDPOINT
  value: ""  # Usar DynamoDB real en producción
- name: RABBITMQ_URL
  value: "amqp://user:password@rabbitmq-cluster:5672/"
- name: ENVIRONMENT
  value: "production"
- name: SUPPLIER_SERVICE_URL
  value: "http://supplier-service:8080"
```

## Configuración de Recursos

### Límites y Requests Recomendados

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Monitoreo y Logging

### Configurar Fluentd/Fluent Bit

Para capturar logs de los contenedores:

```yaml
# ConfigMap para Fluentd
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*mediplus*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
    </source>
```

### Métricas de Prometheus

Los servicios exponen métricas en `/metrics` para Prometheus.

## Escalado Automático con KEDA

### Configuración de ScaledObject

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: supplier-service-scaledobject
spec:
  scaleTargetRef:
    name: supplier-service
  minReplicaCount: 1
  maxReplicaCount: 10
  triggers:
  - type: rabbitmq
    metadata:
      queueName: proveedor.events
      host: amqp://user:password@rabbitmq-cluster:5672/
      queueLength: '5'
```

## Seguridad

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mediplus-network-policy
spec:
  podSelector:
    matchLabels:
      app: mediplus
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: mediplus
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 8081
```

### Pod Security Standards

```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
  - name: supplier-service
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

## Troubleshooting

### Problemas Comunes

1. **Error de autenticación ECR**
   ```bash
   aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-east-1.amazonaws.com
   ```

2. **Imagen no encontrada**
   - Verificar que la imagen existe en ECR
   - Confirmar que el secret de ECR está configurado
   - Verificar permisos de IAM

3. **Problemas de conectividad**
   - Verificar configuración de red en Kubernetes
   - Confirmar que RabbitMQ y DynamoDB son accesibles
   - Revisar logs de los pods

### Comandos Útiles

```bash
# Ver logs de un pod
kubectl logs -f deployment/supplier-service

# Describir un pod
kubectl describe pod <pod-name>

# Ver eventos
kubectl get events --sort-by=.metadata.creationTimestamp

# Verificar conectividad
kubectl exec -it <pod-name> -- curl http://localhost:8080/health
```

## Backup y Recuperación

### Backup de DynamoDB

```bash
# Crear backup
aws dynamodb create-backup \
  --table-name suppliers \
  --backup-name suppliers-backup-$(date +%Y%m%d)
```

### Backup de Configuraciones

```bash
# Exportar configuraciones de Kubernetes
kubectl get all -l app=mediplus -o yaml > mediplus-backup.yaml
```

## Actualizaciones

### Rolling Updates

```bash
# Actualizar imagen
kubectl set image deployment/supplier-service supplier-service=123456789012.dkr.ecr.us-east-1.amazonaws.com/mediplus-supplier-service:v2.0.0

# Verificar rollout
kubectl rollout status deployment/supplier-service

# Rollback si es necesario
kubectl rollout undo deployment/supplier-service
```

Esta guía proporciona una base sólida para desplegar MediPlus en un entorno de producción usando AWS ECR y Kubernetes.
