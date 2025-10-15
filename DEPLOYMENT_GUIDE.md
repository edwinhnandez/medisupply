# Medisupply Microservices - Complete Deployment Guide

## Completed Steps

### 1. Docker Hub Username Updated
- Updated `deploy-microservices.yml` to use `edwinhc93` as Docker Hub username
- Updated all Kubernetes deployment files with correct image references

### 2. Docker Images Pushed to Docker Hub
- `edwinhc93/medisupply-supplier-service:latest` - Successfully pushed
- `edwinhc93/medisupply-purchase-order-service:latest` - Successfully pushed

## Next Steps to Deploy

### Step 3: Run the Deployment

You need to run the Ansible playbook to deploy your microservices to your EC2 instance. Here's how:

#### Prerequisites Check:
1. **EC2 Instance**: Make sure your EC2 instance is running (created by `crear_infra.yml`)
2. **Istio Installed**: Ensure Istio is installed (via `install_istio.yml`)
3. **SSH Key**: Have your `key_lab.pem` file ready

#### Deploy Command:
```bash
# Replace YOUR_EC2_IP with your actual EC2 public IP address
ansible-playbook -i YOUR_EC2_IP, deploy-microservices.yml -u ubuntu --private-key=key_lab.pem
```

#### Example:
```bash
# If your EC2 IP is 54.123.45.67
ansible-playbook -i 54.123.45.67, deploy-microservices.yml -u ubuntu --private-key=key_lab.pem
```

### Step 4: Verify Deployment

After the deployment completes, SSH to your EC2 instance to verify:

```bash
# SSH to your EC2 instance
ssh -i key_lab.pem ubuntu@YOUR_EC2_IP

# Check if all pods are running
kubectl get pods

# Check services
kubectl get services

# Check Istio gateway
kubectl get gateway
kubectl get virtualservice
```

### Step 5: Test the APIs

#### Get the Gateway IP/Port:
```bash
# Get Istio Ingress Gateway IP
kubectl get service istio-ingressgateway -n istio-system

# If using NodePort, get the port
kubectl get service istio-ingressgateway -n istio-system -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}'
```

#### Test Commands:

**1. Health Check:**
```bash
curl http://GATEWAY_IP:PORT/health
```

**2. Create a Supplier:**
```bash
curl -X POST http://GATEWAY_IP:PORT/api/v1/suppliers \
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

**3. List Suppliers:**
```bash
curl http://GATEWAY_IP:PORT/api/v1/suppliers
```

**4. Create an Order:**
```bash
curl -X POST http://GATEWAY_IP:PORT/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "proveedor_id": "supplier-id-from-step-2",
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

**5. List Orders:**
```bash
curl http://GATEWAY_IP:PORT/api/v1/orders
```

**6. Simulate External Event (Stock Low):**
```bash
curl -X POST http://GATEWAY_IP:PORT/api/v1/external/simulate/stock-bajo \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "prod1",
    "stock_actual": 5,
    "stock_minimo": 10,
    "ubicacion": "warehouse-a"
  }'
```

**7. Simulate External Event (High Demand):**
```bash
curl -X POST http://GATEWAY_IP:PORT/api/v1/external/simulate/demanda-alta \
  -H "Content-Type: application/json" \
  -d '{
    "producto_id": "prod1",
    "demanda_actual": 150,
    "demanda_historica": 100,
    "factor_incremento": 1.5
  }'
```

## Troubleshooting

### If Deployment Fails:

**1. Check EC2 Instance Status:**
```bash
# From your local machine
aws ec2 describe-instances --instance-ids YOUR_INSTANCE_ID --query 'Reservations[0].Instances[0].State.Name'
```

**2. Check SSH Connection:**
```bash
ssh -i key_lab.pem ubuntu@YOUR_EC2_IP "echo 'SSH connection successful'"
```

**3. Check if Istio is Running:**
```bash
ssh -i key_lab.pem ubuntu@YOUR_EC2_IP "kubectl get pods -n istio-system"
```

**4. Check Minikube Status:**
```bash
ssh -i key_lab.pem ubuntu@YOUR_EC2_IP "minikube status"
```

### If Pods Don't Start:

**1. Check Pod Logs:**
```bash
kubectl logs -l app=supplier-service
kubectl logs -l app=purchase-order-service
kubectl logs -l app=rabbitmq
kubectl logs -l app=dynamodb-local
```

**2. Check Pod Status:**
```bash
kubectl describe pod POD_NAME
```

**3. Check Resource Usage:**
```bash
kubectl top pods
kubectl top nodes
```

### If Services Are Not Accessible:

**1. Check Service Endpoints:**
```bash
kubectl get endpoints
```

**2. Check Istio Gateway:**
```bash
kubectl get gateway
kubectl describe gateway medisupply-gateway
```

**3. Check VirtualService:**
```bash
kubectl get virtualservice
kubectl describe virtualservice medisupply-vs
```

## Monitoring

### View Logs:
```bash
# Real-time logs
kubectl logs -f -l app=supplier-service
kubectl logs -f -l app=purchase-order-service

# All logs
kubectl logs -l app=supplier-service --all-containers=true
```

### Check Resource Usage:
```bash
kubectl top pods
kubectl top nodes
```

### Check Istio Metrics:
```bash
# Access Istio Grafana (if installed)
kubectl port-forward -n istio-system svc/grafana 3000:3000
# Then open http://localhost:3000 in your browser
```

## Expected Results

After successful deployment, you should see:

1. **All Pods Running:**
   - 2x supplier-service pods
   - 2x purchase-order-service pods
   - 1x rabbitmq pod
   - 1x dynamodb-local pod

2. **Services Available:**
   - All services should have ClusterIP
   - Istio Gateway should be accessible

3. **APIs Working:**
   - Health checks return 200 OK
   - CRUD operations work for suppliers and orders
   - External event simulation triggers automatic order creation

## Configuration Files Updated

- `deploy-microservices.yml` - Docker Hub username updated
- `k8s/supplier-service-deployment.yaml` - Image reference updated
- `k8s/purchase-order-service-deployment.yaml` - Image reference updated

## Notes

- The deployment will take approximately 5-10 minutes to complete
- All services are configured with health checks and proper resource limits
- Istio service mesh provides traffic management and security
- RabbitMQ handles event-driven communication between services
- DynamoDB Local provides persistent data storage

Your microservices are now ready for deployment! 🚀
