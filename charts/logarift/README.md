# Logarift Helm Chart

This chart deploys the Logarift frontend, Go backend, C++ math engine, optional LLM adapter, MongoDB, and Valkey Streams.

## Install

Install from a local checkout:

```bash
helm upgrade --install logarift charts/logarift
```

Install a published chart from GHCR:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 0.1.0
```

The chart defaults to application images published under `ghcr.io/mistervvp`: `logarift-api`, `logarift-frontend`, `logarift-math-engine`, and `logarift-llm-adapter`. If the GHCR packages are private, authenticate Helm and configure Kubernetes image pull secrets before installing.

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

## Development package testing

Pushes to `dev-*` branches publish development images with the branch name and short SHA tags, plus a development Helm chart version such as `0.1.0-dev.42`. Use those versions to validate chart and image changes before creating a GitHub Release. The development chart app version defaults to the matching branch image tag:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 0.1.0-dev.42
```

For immutable deployments, set image digests with `backend.image.digest`, `frontend.image.digest`, `mathEngine.image.digest`, and `llmAdapter.image.digest`.
