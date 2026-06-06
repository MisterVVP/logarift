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

## Project setup with adapter enabled

Smoke-test the local runtime before starting Logarift:

```bash
curl http://localhost:11434/api/chat \
  -d '{"model":"qwen3.6","messages":[{"role":"user","content":"Return only JSON: {\"ok\":true}"}],"stream":false,"format":"json"}'
```

If using the fallback model, replace `qwen3.6` with `qwen3:8b`.

Enable the adapter for Docker Compose:

```bash
export LOGARIFT_LLM_ADAPTER_ENABLED=true
export LOGARIFT_LLM_ADAPTER_URL=http://localhost:8091
export LOGARIFT_LLM_RUNTIME_URL=http://localhost:11434
export LOGARIFT_LLM_MODEL=qwen3.6
docker compose up --build
```

Windows PowerShell equivalent:

```powershell
$env:LOGARIFT_LLM_ADAPTER_ENABLED = "true"
$env:LOGARIFT_LLM_ADAPTER_URL = "http://localhost:8091"
$env:LOGARIFT_LLM_RUNTIME_URL = "http://localhost:11434"
$env:LOGARIFT_LLM_MODEL = "qwen3.6"
docker compose up --build
```

For Docker Compose, the backend reaches the adapter at `http://llm-adapter:8091`, while the adapter reaches host Ollama through `host.docker.internal:11434` by default. Prompts and responses are not logged unless `LOGARIFT_LLM_LOG_PROMPTS=true` or `LOGARIFT_LLM_LOG_RESPONSES=true` is explicitly configured.

## Runtime caches and files

Ollama stores model files and runtime caches according to its platform defaults. Logarift does not copy model weights into the repository and does not send prompts to hosted LLM APIs in local adapter mode.

## Official references

- Ollama Linux documentation: <https://docs.ollama.com/linux>
- Ollama Windows documentation: <https://docs.ollama.com/windows>
- Ollama API documentation: <https://docs.ollama.com/api>
- Ollama Qwen3.6 library page: <https://ollama.com/library/qwen3.6>
- Qwen3.6 upstream repository: <https://github.com/QwenLM/Qwen3.6>
