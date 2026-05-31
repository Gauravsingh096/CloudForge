"""Discord webhook notifications for self-healing events."""
import os

import requests

_WEBHOOK = os.environ.get("DISCORD_WEBHOOK_URL", "")

_RED = 0xFF4444
_GREEN = 0x44BB44
_ORANGE = 0xFF8800


def _send(embed: dict) -> None:
    if not _WEBHOOK:
        return
    try:
        requests.post(_WEBHOOK, json={"embeds": [embed]}, timeout=5)
    except Exception:
        pass  # alerts must never crash the Lambda


def anomaly_detected(playbook: str, app: str, metric: str, value: float) -> None:
    _send({
        "title": f"⚠️  Anomaly Detected — {playbook}",
        "color": _RED,
        "fields": [
            {"name": "App",    "value": f"`{app}`",       "inline": True},
            {"name": "Metric", "value": f"`{metric}`",    "inline": True},
            {"name": "Value",  "value": f"`{value:.4f}`", "inline": True},
        ],
    })


def healed(playbook: str, app: str, action: str, outcome: str) -> None:
    _send({
        "title": f"✅  Healed — {playbook}",
        "color": _GREEN,
        "fields": [
            {"name": "App",     "value": f"`{app}`",    "inline": True},
            {"name": "Action",  "value": f"`{action}`", "inline": True},
            {"name": "Outcome", "value": outcome[:200], "inline": False},
        ],
    })


def failed(playbook: str, app: str, action: str, error: str) -> None:
    _send({
        "title": f"❌  Remediation Failed — {playbook}",
        "color": _ORANGE,
        "fields": [
            {"name": "App",    "value": f"`{app}`",   "inline": True},
            {"name": "Action", "value": f"`{action}`","inline": True},
            {"name": "Error",  "value": error[:200],  "inline": False},
        ],
    })
