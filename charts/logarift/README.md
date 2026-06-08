# Logarift Helm Chart

This chart deploys the Logarift frontend, Go backend, C++ math engine, optional LLM adapter, MongoDB, and Valkey Streams.

## Install

```bash
helm upgrade --install logarift charts/logarift
```

Port-forward the frontend when ingress is disabled:

```bash
kubectl port-forward svc/logarift-frontend 5173:5173
```

## External MongoDB and Valkey

MongoDB and Valkey are enabled by default for local or small-cluster installs. Disable them and provide externally managed connection strings when your cluster already has these services:

```bash
helm upgrade --install logarift charts/logarift \
  --set mongodb.enabled=false \
  --set mongodb.external.uri='mongodb://mongo.example:27017' \
  --set valkey.enabled=false \
  --set valkey.external.url='redis://valkey.example:6379'
```

Existing Secrets are supported to avoid placing connection strings in release values:

```yaml
mongodb:
  enabled: false
  external:
    existingSecret: logarift-mongodb
    existingSecretKey: uri
valkey:
  enabled: false
  external:
    existingSecret: logarift-valkey
    existingSecretKey: url
```

## Scheduling and operations

Each component exposes optional Kubernetes placement settings with safe defaults:

- `nodeSelector`
- `affinity` including pod anti-affinity
- `tolerations`
- `topologySpreadConstraints`
- `resources`
- `podAnnotations` and `podLabels`

The backend also supports configurable upload/export persistence, probes, a PodDisruptionBudget, and additional environment variables.
