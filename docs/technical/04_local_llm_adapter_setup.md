# Local LLM Adapter Setup

The Local LLM Adapter is an optional local-only HTTP service that enriches quick friction notes by calling an Ollama-compatible runtime. It is disabled by default and quick event saving continues to use deterministic rules when the adapter is unavailable.

## Ubuntu with Ollama and Qwen

Install and start Ollama:

```bash
curl -fsSL https://ollama.com/install.sh | sh
sudo systemctl enable ollama
sudo systemctl start ollama
ollama --version
ollama pull qwen3.6
ollama run qwen3.6
```

If `qwen3.6` is too large for the developer machine, use the smaller fallback:

```bash
ollama pull qwen3:8b
ollama run qwen3:8b
```

Developers with enough RAM/VRAM may pin a concrete Qwen3.6 size:

```bash
ollama pull qwen3.6:27b
# or, for stronger hardware:
ollama pull qwen3.6:35b
```

Optional GPU and service checks:

```bash
nvidia-smi
journalctl -e -u ollama
```

## Windows 11 with Ollama and Qwen

1. Download and run `OllamaSetup.exe` from the official Ollama download page.
2. Open a new PowerShell window after installation so the updated PATH is available.
3. Verify Ollama is running:

```powershell
ollama --version
Invoke-WebRequest http://localhost:11434/api/tags
```

4. Pull and run the first supported Qwen model:

```powershell
ollama pull qwen3.6
ollama run qwen3.6
```

If `qwen3.6` is too large, use `qwen3:8b`. To pin a larger Qwen3.6 size, use `qwen3.6:27b` or `qwen3.6:35b`.


## Optional Logarift Ollama Modelfiles

The repository includes project-specific Ollama Modelfiles under `llm-adapter/modelfiles/`. They are tailored to the adapter response contract (`fields` plus `warnings`) instead of the generic suggestion fields. This keeps the local model aligned with Logarift's ontology, confidence gates, and MVP non-goals.

Create the lower-resource alias:

```bash
ollama pull qwen3:8b
ollama create logarift-enricher-qwen3-8b -f llm-adapter/modelfiles/logarift-enricher-qwen3-8b.Modelfile
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-8b
```

Create the stronger Qwen3.6 alias:

```bash
ollama pull qwen3.6
ollama create logarift-enricher-qwen36 -f llm-adapter/modelfiles/logarift-enricher-qwen36.Modelfile
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen36
```

These Modelfiles intentionally do not emit `suggested_next_action`, recommendations, coaching, or productivity judgments. The adapter is only allowed to propose structured enrichment candidates; the backend still validates every field and falls back to deterministic rules when needed.

## Project setup with adapter enabled

Smoke-test the local runtime before starting Logarift:

```bash
curl http://localhost:11434/api/chat \
  -d '{"model":"qwen3.6","messages":[{"role":"user","content":"Return only JSON: {\"ok\":true}"}],"stream":false,"format":"json"}'
```

If using the fallback model, replace `qwen3.6` with `qwen3:8b`. If using a Logarift Modelfile alias, replace it with `logarift-enricher-qwen3-8b` or `logarift-enricher-qwen36`.

Enable the adapter for Docker Compose:

```bash
export LOGARIFT_LLM_ADAPTER_ENABLED=true
export LOGARIFT_LLM_MODEL=qwen3.6
# Optional on Linux if host-gateway is unavailable; the compose default is usually correct:
# export LOGARIFT_LLM_RUNTIME_URL=http://host.docker.internal:11434
docker compose up --build
```

Windows PowerShell equivalent:

```powershell
$env:LOGARIFT_LLM_ADAPTER_ENABLED = "true"
$env:LOGARIFT_LLM_MODEL = "qwen3.6"
# Optional; the compose default is usually correct:
# $env:LOGARIFT_LLM_RUNTIME_URL = "http://host.docker.internal:11434"
docker compose up --build
```

For Docker Compose, the backend reaches the adapter at `http://llm-adapter:8091`, while the adapter reaches host Ollama through `host.docker.internal:11434` by default. Do not set `LOGARIFT_LLM_RUNTIME_URL=http://localhost:11434` for Docker Compose, because `localhost` inside the adapter container is the container itself, not the host machine. Use `http://localhost:11434` only when running the adapter directly on the host outside Docker. Prompts and responses are not logged unless `LOGARIFT_LLM_LOG_PROMPTS=true` or `LOGARIFT_LLM_LOG_RESPONSES=true` is explicitly configured.

## Troubleshooting adapter runtime errors

If adapter logs show `ollama_chat_http_error`, `ollama_tags_http_error`, or `runtime_error`, first check the adapter readiness endpoint:

```bash
curl http://localhost:8091/health/ready
curl http://localhost:8091/v1/models/current
```

Useful log fields are `error_code`, `error_message`, `runtime_endpoint`, `http_status`, and `hint`. The logs intentionally do not include raw note text.

For Docker Compose on Linux, host Ollama must be reachable from containers through `host.docker.internal:11434`. If the hint says the runtime cannot be reached, configure Ollama to listen on a non-loopback interface and restart it:

```bash
sudo systemctl edit ollama
```

Add:

```ini
[Service]
Environment="OLLAMA_HOST=0.0.0.0:11434"
```

Then restart and verify:

```bash
sudo systemctl daemon-reload
sudo systemctl restart ollama
curl http://localhost:11434/api/tags
```

If the model exists but requests time out, increase both deadlines so the backend waits long enough for the adapter and the adapter waits long enough for Ollama:

```bash
export LOGARIFT_LLM_ADAPTER_TIMEOUT_MS=30000
export LOGARIFT_LLM_REQUEST_TIMEOUT_MS=30000
docker compose up --build
```

## Runtime caches and files

Ollama stores model files and runtime caches according to its platform defaults. Logarift does not copy model weights into the repository and does not send prompts to hosted LLM APIs in local adapter mode.

## Official references

- Ollama Linux documentation: <https://docs.ollama.com/linux>
- Ollama Windows documentation: <https://docs.ollama.com/windows>
- Ollama API documentation: <https://docs.ollama.com/api>
- Ollama Qwen3.6 library page: <https://ollama.com/library/qwen3.6>
- Qwen3.6 upstream repository: <https://github.com/QwenLM/Qwen3.6>
