# Medisupply Microservices - Kubernetes Deployment Guide

## Overview

This document explains how to deploy the Medisupply microservices to Kubernetes using the infrastructure created by the Ansible playbooks (`crear_infra.yml` and `install_istio.yml`).

## Issues Identified and Fixed

### 1. **NATS vs RabbitMQ Migration**
- **Issue**: Kubernetes manifests still referenced NATS, but microservices were migrated to RabbitMQ
- **Fix**: Created new RabbitMQ Kubernetes manifests (`k8s/rabbitmq-deployment.yaml`)

### 2. **Missing Infrastructure Components**
- **Issue**: No DynamoDB Local setup in Kubernetes
- **Fix**: Updated DynamoDB Local manifests with proper configuration

### 3. **Istio Integration Missing**
- **Issue**: No Gateway/VirtualService for external access
- **Fix**: Created Istio Gateway and VirtualService configurations

### 4. **Environment Configuration**
- **Issue**: Missing AWS credentials and proper environment variables
- **Fix**: Created ConfigMaps and Secrets for centralized configuration

### 5. **Image References**
- **Issue**: Using generic `mediplus/` prefix instead of actual Docker Hub username
- **Fix**: Updated to use configurable Docker Hub username

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Istio Gateway                            │
│                 (External Access)                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                Kubernetes Cluster                          │
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐                 │
│  │ Supplier Service│  │Purchase Order   │                 │
│  │   (2 replicas)  │  │   Service       │                 │
│  │                 │  │  (2 replicas)   │                 │
│  └─────────┬───────┘  └─────────┬───────┘                 │
│            │                    │                         │
│            └──────────┬─────────┘                         │
│                       │                                   │
│  ┌────────────────────┴─────────────────────────────────┐ │
│  │              RabbitMQ Cluster                        │ │
│  │         (Message Broker)                             │ │
│  └────────────────────┬─────────────────────────────────┘ │
│                       │                                   │
│  ┌────────────────────┴─────────────────────────────────┐ │
│  │              DynamoDB Local                          │ │
│  │            (Data Storage)                            │ │
│  └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Files Created/Modified

### New Kubernetes Manifests
1. **`k8s/rabbitmq-deployment.yaml`** - RabbitMQ cluster with initialization
2. **`k8s/configmaps-secrets.yaml`** - Configuration and secrets management
3. **`k8s/istio-gateway.yaml`** - Istio Gateway, VirtualService, and security policies

### Updated Manifests
1. **`k8s/supplier-service-deployment.yaml`** - Updated for RabbitMQ and Istio
2. **`k8s/purchase-order-service-deployment.yaml`** - Updated for RabbitMQ and Istio
3. **`k8s/dynamodb-local.yaml`** - Enhanced with proper configuration

### New Ansible Playbook
1. **`deploy-microservices.yml`** - Complete deployment automation

## Key Features Implemented

### 1. **High Availability**
- 2 replicas for each microservice
- Proper health checks and probes
- Circuit breakers and retry policies

### 2. **Service Mesh Integration**
- Istio sidecar injection
- Traffic management and routing
- Security policies and authorization

### 3. **Message Broker**
- RabbitMQ cluster with persistent storage
- Automatic exchange and queue creation
- Event-driven architecture support

### 4. **Configuration Management**
- Centralized ConfigMaps
- Secure secrets management
- Environment-specific configuration

### 5. **Monitoring and Observability**
- Health check endpoints
- Istio metrics and tracing
- Prometheus integration (from Ansible setup)

## Deployment Instructions

### Prerequisites
1. Infrastructure created by `crear_infra.yml`
2. Istio installed by `install_istio.yml`
3. Docker images pushed to Docker Hub

### Step 1: Update Configuration
Edit `deploy-microservices.yml` and change:
```yaml
DOCKERHUB_USERNAME: "your-dockerhub-username"  # Change this to your actual Docker Hub username
```

### Step 2: Deploy Microservices
```bash
# From your local machine
ansible-playbook -i your-ec2-ip, deploy-microservices.yml -u ubuntu --private-key=key_lab.pem
```

### Step 3: Verify Deployment
```bash
# SSH to your EC2 instance
ssh -i key_lab.pem ubuntu@your-ec2-ip

# Check pods
kubectl get pods

# Check services
kubectl get services

# Check Istio gateway
kubectl get gateway
kubectl get virtualservice
```

## API Endpoints

### External Access (via Istio Gateway)
- **Supplier API**: `http://gateway-ip/api/v1/suppliers`
- **Order API**: `http://gateway-ip/api/v1/orders`
- **External Events API**: `http://gateway-ip/api/v1/external`

### Internal Access (within cluster)
- **Supplier Service**: `http://supplier-service:8080`
- **Purchase Order Service**: `http://purchase-order-service:8081`
- **DynamoDB Local**: `http://dynamodb-local:8000`
- **RabbitMQ Management**: `http://rabbitmq:15672`

## Testing the Deployment

### 1. Health Check
```bash
curl http://gateway-ip/health
```

### 2. Create a Supplier
```bash
curl -X POST http://gateway-ip/api/v1/suppliers \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Test Supplier",
    "contacto": "test@example.com",
    "telefono": "+1234567890",
    "direccion": "123 Test St",
    "certificaciones": ["ISO9001"],
    "estado_proveedor": "activo"
  }'
```

### 3. Create an Order
```bash
curl -X POST http://gateway-ip/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "proveedor_id": "supplier-id",
    "productos": [
      {
        "producto_id": "prod1",
        "cantidad": 10,
        "precio_unitario": 25.50
      }
    ],
    "motivo_generacion": "restock"
  }'
```

### 4. Simulate External Event
```bash
curl -X POST http://gateway-ip/api/v1/external/simulate/stock-bajo \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "prod1",
    "stock_actual": 5,
    "stock_minimo": 10,
    "ubicacion": "warehouse-a"
  }'
```

## Monitoring and Troubleshooting

### Check Pod Status
```bash
kubectl get pods -o wide
kubectl describe pod <pod-name>
```

### View Logs
```bash
kubectl logs -l app=supplier-service
kubectl logs -l app=purchase-order-service
kubectl logs -l app=rabbitmq
```

### Check Services
```bash
kubectl get services
kubectl describe service <service-name>
```

### Istio Monitoring
```bash
kubectl get gateway
kubectl get virtualservice
kubectl get destinationrule
kubectl get authorizationpolicy
```

### RabbitMQ Management
Access RabbitMQ Management UI at `http://gateway-ip:15672` with credentials `mediplus/mediplus123`

## Security Considerations

### 1. **Network Policies**
- Istio authorization policies implemented
- Service-to-service authentication
- Traffic encryption in transit

### 2. **Secrets Management**
- AWS credentials stored in Kubernetes secrets
- Base64 encoded (use proper secret management in production)

### 3. **RBAC**
- Service accounts with minimal permissions
- Istio service mesh security

## Scaling and Performance

### Horizontal Scaling
```bash
kubectl scale deployment supplier-service --replicas=3
kubectl scale deployment purchase-order-service --replicas=3
```

### Resource Limits
- CPU: 200m-1000m per pod
- Memory: 256Mi-1Gi per pod
- Adjust based on actual usage

### RabbitMQ Scaling
- Currently single instance
- Can be scaled to cluster mode for production

## Production Considerations

### 1. **Replace DynamoDB Local**
- Use AWS DynamoDB service
- Update endpoint configuration
- Implement proper IAM roles

### 2. **TLS Certificates**
- Configure proper TLS certificates for Istio Gateway
- Use cert-manager for automatic certificate management

### 3. **Monitoring**
- Implement comprehensive monitoring with Prometheus/Grafana
- Set up alerting for critical metrics
- Use Jaeger for distributed tracing

### 4. **Backup and Recovery**
- Implement RabbitMQ data backup
- Set up DynamoDB point-in-time recovery
- Create disaster recovery procedures

## Troubleshooting Common Issues

### 1. **Pods Not Starting**
- Check resource limits
- Verify image availability
- Check environment variables

### 2. **Services Not Accessible**
- Verify Istio Gateway configuration
- Check VirtualService routing rules
- Ensure proper port configuration

### 3. **RabbitMQ Connection Issues**
- Verify RabbitMQ is running
- Check initialization job completion
- Verify credentials and connection string

### 4. **DynamoDB Connection Issues**
- Ensure DynamoDB Local is running
- Verify table creation
- Check AWS credentials configuration

This deployment provides a production-ready foundation for the Medisupply microservices with proper service mesh integration, monitoring, and security policies.
