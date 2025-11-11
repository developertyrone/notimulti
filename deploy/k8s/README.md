# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying the notimulti notification server.

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured to access your cluster
- Container image built and pushed to a registry (or available locally)

## Quick Start

### 1. Build and Load Docker Image

```bash
# Build the Docker image
docker build -t notimulti:latest -f ../Dockerfile ..

# For local clusters (minikube, kind, k3s), load the image:
# Minikube:
minikube image load notimulti:latest

# Kind:
kind load docker-image notimulti:latest

# For production, push to your registry:
# docker tag notimulti:latest your-registry.com/notimulti:latest
# docker push your-registry.com/notimulti:latest
```

### 2. Configure Provider Settings

Edit `configmap.yaml` with your actual provider credentials:

```bash
# Edit the ConfigMap
kubectl edit configmap notimulti-config

# Or apply from edited file
kubectl apply -f configmap.yaml
```

**Important**: Update these values in `configmap.yaml`:
- Telegram: `bot_token`, `default_chat_id`
- Email: `smtp_host`, `smtp_port`, `smtp_username`, `smtp_password`, `from_address`

### 3. Deploy to Kubernetes

```bash
# Create namespace (optional)
kubectl create namespace notimulti

# Apply all manifests
kubectl apply -f configmap.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Optional: Apply ingress for external access
kubectl apply -f ingress.yaml
```

### 4. Verify Deployment

```bash
# Check pod status
kubectl get pods -l app=notimulti

# Check logs
kubectl logs -l app=notimulti -f

# Check service
kubectl get svc notimulti

# Port forward to access locally
kubectl port-forward svc/notimulti 8080:80
```

Access the application at http://localhost:8080

## Manifest Files

### deployment.yaml
- **Type**: StatefulSet
- **Replicas**: 1 (single instance for SQLite)
- **Security**: Runs as non-root user (UID 1000)
- **Probes**: 
  - Liveness: `/api/v1/health` (checks if container is alive)
  - Readiness: `/api/v1/ready` (checks database + providers)
- **Resources**:
  - Requests: 128Mi memory, 100m CPU
  - Limits: 512Mi memory, 500m CPU

### service.yaml
- **Type**: ClusterIP
- **Port**: 80 → 8080 (container)
- **Selector**: app=notimulti

### configmap.yaml
- **Purpose**: Provider configuration files
- **Files**: 
  - `telegram-alerts.json` - Telegram provider config
  - `email-prod.json` - Email provider config
- **⚠️ Update with real credentials before deployment**

### pvc.yaml
- **Type**: PersistentVolumeClaim
- **Access**: ReadWriteOnce
- **Size**: 10Gi
- **Purpose**: Database persistence

### ingress.yaml (optional)
- **Type**: Ingress
- **Purpose**: External HTTPS access
- **Requires**: Ingress controller (nginx, traefik, etc.)
- **⚠️ Update `host` with your domain**

## Configuration

### Environment Variables

Set via deployment.yaml or override with ConfigMap:

```yaml
env:
- name: PORT
  value: "8080"
- name: LOG_LEVEL
  value: "info"  # debug, info, warn, error
- name: CONFIG_DIR
  value: "/app/configs"
- name: DB_PATH
  value: "/app/data/notifications.db"
- name: LOG_RETENTION_DAYS
  value: "90"
```

### Volume Mounts

- **Config**: `/app/configs` (ConfigMap, read-only)
- **Data**: `/app/data` (PVC, read-write for database)

## Scaling Considerations

**⚠️ Important**: This application uses SQLite and should run with `replicas: 1`.

For high availability or horizontal scaling:
1. Replace SQLite with PostgreSQL/MySQL
2. Update deployment to use Deployment instead of StatefulSet
3. Increase replicas as needed
4. Configure external database connection

## Monitoring

### Health Checks

```bash
# Liveness probe endpoint
curl http://pod-ip:8080/api/v1/health

# Readiness probe endpoint (checks database + providers)
curl http://pod-ip:8080/api/v1/ready
```

### Metrics

```bash
# View pod metrics
kubectl top pod -l app=notimulti

# View logs
kubectl logs -l app=notimulti --tail=100 -f
```

## Troubleshooting

### Pod not starting

```bash
# Check pod events
kubectl describe pod -l app=notimulti

# Check logs
kubectl logs -l app=notimulti

# Common issues:
# - Image pull errors: Verify image name/registry
# - ConfigMap missing: Apply configmap.yaml first
# - PVC pending: Check storage class availability
```

### Database persistence issues

```bash
# Check PVC status
kubectl get pvc

# Check volume mount
kubectl exec -it notimulti-0 -- ls -la /app/data

# Verify permissions (should be owned by UID 1000)
kubectl exec -it notimulti-0 -- ls -ln /app/data
```

### Configuration not loading

```bash
# Verify ConfigMap exists
kubectl get configmap notimulti-config

# Check mounted config files
kubectl exec -it notimulti-0 -- ls -la /app/configs

# View actual config content
kubectl exec -it notimulti-0 -- cat /app/configs/telegram-alerts.json
```

## Updating

### Update Configuration

```bash
# Edit ConfigMap
kubectl edit configmap notimulti-config

# Restart pod to reload config (if no file watcher)
kubectl rollout restart statefulset/notimulti
```

### Update Application

```bash
# Build and push new image with version tag
docker build -t notimulti:v1.1.0 -f ../Dockerfile ..
docker push your-registry.com/notimulti:v1.1.0

# Update deployment image
kubectl set image statefulset/notimulti notimulti=notimulti:v1.1.0

# Or edit deployment.yaml and apply
kubectl apply -f deployment.yaml
```

## Cleanup

```bash
# Delete all resources
kubectl delete -f ingress.yaml
kubectl delete -f service.yaml
kubectl delete -f deployment.yaml
kubectl delete -f pvc.yaml
kubectl delete -f configmap.yaml

# Or delete by label
kubectl delete all,pvc,configmap -l app=notimulti
```

## Production Best Practices

1. **Use Secrets for sensitive data** (not ConfigMaps)
   ```bash
   kubectl create secret generic notimulti-secrets \
     --from-literal=telegram-token=YOUR_TOKEN \
     --from-literal=smtp-password=YOUR_PASSWORD
   ```

2. **Enable TLS/HTTPS** via Ingress with cert-manager

3. **Set resource limits** appropriate for your workload

4. **Use namespace isolation**
   ```bash
   kubectl create namespace notimulti
   kubectl apply -f . -n notimulti
   ```

5. **Implement monitoring** with Prometheus/Grafana

6. **Configure backup** for PVC data

7. **Use image tags** (not `latest`) for version control

## Support

For issues or questions:
- Check application logs: `kubectl logs -l app=notimulti`
- Review pod events: `kubectl describe pod -l app=notimulti`
- Verify provider configurations in ConfigMap
- Test health endpoints directly
