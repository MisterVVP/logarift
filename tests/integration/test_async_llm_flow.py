import json
import os
import shutil
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[2]
PROJECT = f"logarift-e2e-{int(time.time())}"
BACKEND = "http://127.0.0.1:8080"
FRONTEND = "http://127.0.0.1:5173"


def compose(*args: str, env: dict[str, str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["docker", "compose", "-p", PROJECT, *args],
        cwd=ROOT,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        check=True,
    )


def get_json(url: str) -> dict:
    with urllib.request.urlopen(url, timeout=5) as response:
        return json.loads(response.read().decode("utf-8"))


def post_json(url: str, payload: dict) -> dict:
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(req, timeout=10) as response:
        return json.loads(response.read().decode("utf-8"))


def wait_for_json(url: str, predicate, timeout: float = 90.0) -> dict:
    deadline = time.time() + timeout
    last_error: Exception | None = None
    while time.time() < deadline:
        try:
            payload = get_json(url)
            if predicate(payload):
                return payload
        except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
            last_error = exc
        time.sleep(1)
    raise AssertionError(f"Timed out waiting for {url}; last_error={last_error!r}")


def wait_for_text(url: str, expected: str, timeout: float = 90.0) -> str:
    deadline = time.time() + timeout
    last_error: Exception | None = None
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=5) as response:
                text = response.read().decode("utf-8")
            if expected in text:
                return text
        except urllib.error.URLError as exc:
            last_error = exc
        time.sleep(1)
    raise AssertionError(f"Timed out waiting for {url}; last_error={last_error!r}")


def test_async_ui_backend_llm_adapter_flow_through_compose() -> None:
    if shutil.which("docker") is None:
        pytest.skip("Docker CLI is required for compose E2E tests.")
    env = os.environ.copy()
    env.update(
        {
            "LOGARIFT_LLM_ADAPTER_ENABLED": "true",
            "LOGARIFT_LLM_MOCK_RESPONSE_ENABLED": "true",
            "LOGARIFT_LLM_MODEL": "logarift-e2e-mock",
            "LOGARIFT_LLM_ADAPTER_TIMEOUT_MS": "5000",
            "LOGARIFT_LLM_REQUEST_TIMEOUT_MS": "5000",
        }
    )
    try:
        compose("up", "--build", "-d", env=env)
        wait_for_json(f"{BACKEND}/api/v1/status", lambda data: data.get("database", {}).get("ready") is True)
        wait_for_text(FRONTEND, "Logarift")

        created = post_json(
            f"{BACKEND}/api/v1/friction-events/quick",
            {
                "occurred_at": "2026-06-06T19:00:00Z",
                "friction_level": "orange",
                "notes_markdown": "CI failed again after 20 min with an unclear timeout.",
            },
        )

        event = created["event"]
        enrichment = created["enrichment"]
        assert enrichment["llm_status"] == "queued"
        assert enrichment["job_id"]
        assert enrichment["trace_id"]
        assert event["enrichment"]["llm_status"] == "queued"

        event_id = event["id"]
        final = wait_for_json(
            f"{BACKEND}/api/v1/friction-events/{event_id}",
            lambda data: data["event"].get("enrichment", {}).get("llm_status")
            in {"succeeded", "partially_succeeded"},
        )["event"]

        assert final["workflow_stage"] == "test"
        assert final["friction_layer"] == "technical"
        assert final["friction_type"] == "failed_feedback"
        assert "llm-mock" in final.get("tags", [])
        assert final["inference"]["local_llm"]["accepted_fields"]
        assert final["enrichment"]["merge_summary"]["accepted_field_count"] >= 3

        job = get_json(f"{BACKEND}/api/v1/enrichment-jobs/{enrichment['job_id']}")["job"]
        assert job["status"] in {"succeeded", "partially_succeeded"}
        assert job["trace_id"] == enrichment["trace_id"]
        assert job["merge_summary"]["field_decisions"]["workflow_stage"]["decision"] == "accepted"
    finally:
        compose("down", "-v", "--remove-orphans", env=env)
