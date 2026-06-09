# Logarift Helm Chart

This chart deploys the Logarift frontend, Go backend, C++ math engine, optional LLM adapter, optional Ollama runtime, MongoDB, and Valkey Streams.

## Deployment posture

This chart is the preferred shared deployment path for Logarift. It is intended for private Kubernetes clusters where a platform or DevEx team can make anonymous friction logging available to developers, technical leads, Developer Experience engineers, and engineering managers across a tech organization.

The chart does not introduce Logarift-managed user accounts or per-person authorization. If an organization needs an access gate, place it at the Gateway API, ingress, identity-aware proxy, or service-mesh boundary until a future SSO feature is designed.

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

The chart defaults to the `latest` application images published under `ghcr.io/mistervvp`: `logarift-api`, `logarift-frontend`, `logarift-math-engine`, and `logarift-llm-adapter`. Every push to `main` refreshes those `latest` image tags and publishes a `0.1.0-main.<run>` chart whose default app image tag is also `latest`. Packaged GitHub release and `dev-*` charts are rewritten by the release workflow to use the matching release or branch image tag. If the GHCR packages are private, authenticate Helm and configure Kubernetes image pull secrets before installing.

Port-forward the frontend when Gateway API exposure is disabled:

```bash
kubectl port-forward svc/logarift-frontend 5173:5173
```


## Local Kubernetes quick start

MicroK8s users can enable DNS, hostpath storage, Helm, and the MicroK8s routing addon. The addon is named `ingress`, but current MicroK8s versions include Gateway API support and expose a Traefik Gateway that this chart can attach to without creating Kubernetes Ingress resources.

Use `values.local.yaml` for a local MicroK8s install with chart-managed Ollama and the LLM adapter enabled. Ollama remains disabled in the base chart values. When enabled locally, the Ollama init container pulls `qwen3:8b` and creates the default `logarift-enricher-qwen3-8b` alias from the bundled Logarift Modelfile before the Ollama container starts.

```bash
microk8s status --wait-ready
microk8s enable dns hostpath-storage helm3 ingress
microk8s helm3 upgrade --install logarift charts/logarift \
  --create-namespace --namespace logarift \
  --values charts/logarift/values.local.yaml
microk8s kubectl -n logarift rollout status statefulset/logarift-ollama
microk8s kubectl -n logarift rollout status deploy/logarift-llm-adapter
```

For Minikube, kind, Docker Desktop Kubernetes, and other local clusters without a Gateway API controller, install with defaults and use port-forwarding:

```bash
kubectl create namespace logarift
helm upgrade --install logarift charts/logarift --namespace logarift
kubectl -n logarift rollout status deploy/logarift-frontend
kubectl -n logarift port-forward svc/logarift-frontend 5173:5173
```

Open `http://localhost:5173`. If your local cluster has a Gateway API implementation, set `gateway.enabled=true` and attach `httpRoute.parentRefs` to that implementation's Gateway.

## Gateway API exposure

The chart does not support legacy Kubernetes Ingress resources. Use Gateway API by enabling `gateway.enabled`. The chart can either create a Gateway or attach its HTTPRoute to an existing Gateway:

```bash
helm upgrade --install logarift charts/logarift \
  --set gateway.enabled=true \
  --set gateway.create=true \
  --set gateway.className=standard
```

To attach to an existing Gateway, leave `gateway.create=false` and set either `gateway.name` or explicit `httpRoute.parentRefs`. The default HTTPRoute sends `/api`, `/health`, and `/uploads` to the backend service and `/` to the frontend service.

## LLM adapter and Ollama service discovery

When `llmAdapter.enabled=true`, the backend integration is enabled. By default `llmAdapter.deploy=true` also creates a ClusterIP Service for the adapter. The backend defaults to the cluster-local DNS URL `http://<release>-logarift-llm-adapter.<namespace>.svc.<clusterDomain>:8091`, so pods can reach the adapter regardless of whether they run on the same or different nodes. Set `llmAdapter.deploy=false` and provide `llmAdapter.backendURL` when using an existing adapter Service or custom service name.

The chart can also deploy an in-cluster Ollama runtime with `ollama.enabled=true`; it is disabled by default. When both `llmAdapter.enabled=true` and `ollama.enabled=true`, an empty `llmAdapter.ollamaURL` makes the adapter use the chart-managed Ollama Service DNS name. The default chart model is `logarift-enricher-qwen3-8b`, and `ollama.modelInit.enabled=true` prepares that alias automatically from `qwen3:8b`. If you keep `ollama.enabled=false`, set `llmAdapter.ollamaURL` to an existing Ollama-compatible runtime Service URL.

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

The backend also supports configurable upload/export persistence, probes, a PodDisruptionBudget, and additional environment variables. The default app container `securityContext` drops all Linux capabilities and disables privilege escalation; this is safe for Logarift app containers because they listen on unprivileged ports and do not require kernel-level privileges. MongoDB and Valkey StatefulSets keep their own image defaults instead of inheriting this app-container security context.

## Development package testing

Pushes to `main` publish `latest` images, short-SHA image tags, and a rolling Helm chart version such as `0.1.0-main.42` whose app version defaults to `latest`. Pushes to `dev-*` branches publish development images with the branch name and short SHA tags, plus a development Helm chart version such as `0.1.0-dev.42`. Use those versions to validate chart and image changes before creating a GitHub Release. The development chart app version defaults to the matching branch image tag:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 0.1.0-dev.42
```

For immutable deployments, set image digests with `backend.image.digest`, `frontend.image.digest`, `mathEngine.image.digest`, `llmAdapter.image.digest`, and `ollama.image.digest`.
