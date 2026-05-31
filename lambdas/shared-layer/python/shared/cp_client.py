"""HTTP client for CloudForge control-plane internal ops endpoints."""
import os

import requests

_BASE = os.environ["CONTROL_PLANE_URL"]  # e.g. https://cloudforge.is-a.dev
_KEY = os.environ.get("INTERNAL_API_KEY", "")


def _headers() -> dict:
    h = {"Content-Type": "application/json"}
    if _KEY:
        h["X-Internal-Key"] = _KEY
    return h


def restart(app: str) -> dict:
    """kubectl rollout restart deployment/app-<app>."""
    r = requests.post(f"{_BASE}/api/apps/{app}/restart", headers=_headers(), timeout=30)
    r.raise_for_status()
    return r.json()


def scale(app: str, *, replicas: int | None = None, replicas_delta: int | None = None) -> dict:
    """kubectl scale deployment/app-<app> --replicas=N."""
    body: dict = {}
    if replicas is not None:
        body["replicas"] = replicas
    elif replicas_delta is not None:
        body["replicas_delta"] = replicas_delta
    else:
        raise ValueError("replicas or replicas_delta is required")
    r = requests.post(f"{_BASE}/api/apps/{app}/scale", json=body, headers=_headers(), timeout=30)
    r.raise_for_status()
    return r.json()


def rollback(app: str) -> dict:
    """kubectl rollout undo deployment/app-<app>."""
    r = requests.post(f"{_BASE}/api/apps/{app}/rollback", headers=_headers(), timeout=30)
    r.raise_for_status()
    return r.json()
