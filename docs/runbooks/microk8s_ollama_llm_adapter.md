# MicroK8s Ollama and LLM adapter runbook

This runbook verifies the chart-managed Ollama runtime used by the Logarift `llm-adapter` on a local MicroK8s cluster. The `llm-adapter` does not need GPU access because it calls Ollama over HTTP; Ollama is the pod that must see the GPU devices or Kubernetes GPU resources.

## GPU acceleration for chart-managed Ollama

The Helm chart defaults to CPU-only Ollama with `ollama.acceleration.type=none`. CPU mode does not request GPU resources, mount host GPU devices, set GPU-specific environment variables, configure a GPU RuntimeClass, or make the Ollama container privileged.

For AMD local MicroK8s testing, use the ROCm-capable `ollama/ollama:rocm` image and mount `/dev/kfd` plus `/dev/dri` from the host into the Ollama container. Some cards, including common RX 6900 XT setups, may need `HSA_OVERRIDE_GFX_VERSION="10.3.0"`.

For NVIDIA testing, install host NVIDIA drivers and the NVIDIA Kubernetes device plugin first. The chart requests the configured NVIDIA resource name, defaulting to `nvidia.com/gpu`, and adds `NVIDIA_VISIBLE_DEVICES` plus `NVIDIA_DRIVER_CAPABILITIES` only to the Ollama container.

AMD device-plugin clusters can use `ollama.acceleration.amd.resourceName`, such as `amd.com/gpu`, with `ollama.acceleration.amd.hostDevices.enabled=false`. For simple single-node MicroK8s testing, AMD host-device mode is usually enough and avoids requiring a separate AMD device plugin.

## Verify AMD devices on the host

```bash
ls -l /dev/kfd /dev/dri
```

## Deploy AMD ROCm config

```bash
microk8s helm3 upgrade --install logarift charts/logarift \
  --create-namespace --namespace logarift \
  --values charts/logarift/values.local.yaml \
  --values charts/logarift/examples/values.local-amd-rocm.yaml
```

## Verify AMD devices inside the Ollama pod

```bash
microk8s kubectl -n logarift exec statefulset/logarift-ollama -- \
  ls -l /dev/kfd /dev/dri
```

## Check Ollama GPU logs

```bash
microk8s kubectl -n logarift logs statefulset/logarift-ollama --tail=200 | \
  grep -Ei 'rocm|gpu|amd|gfx|hip|cuda|nvidia'
```

## Verify models

```bash
microk8s kubectl -n logarift run ollama-check --rm -i --restart=Never --image=curlimages/curl -- \
  curl -s http://logarift-ollama:11434/api/tags
```

## Test generation

```bash
time microk8s kubectl -n logarift run ollama-generate-check --rm -i --restart=Never --image=curlimages/curl -- \
  curl -s http://logarift-ollama:11434/api/generate \
    -H "Content-Type: application/json" \
    -d '{"model":"logarift-enricher-qwen3-8b","prompt":"Return only JSON: {\"ok\":true}","stream":false}'
```

## Watch GPU usage on the host

For AMD:

```bash
amdgpu_top
```

For NVIDIA:

```bash
nvidia-smi
```
