"""Direct control-plane API remediation — triggers redeploy via CloudForge API."""

import logging
import os

import requests

log = logging.getLogger(__name__)

CONTROL_PLANE_URL = os.environ.get("CONTROL_PLANE_URL", "http://localhost:8080")


def trigger_remediation(playbook: dict, metric_value: float):
    action = playbook.get("action", {})
    if action.get("type") == "redeploy":
        _redeploy(action, playbook["name"], metric_value)
    else:
        log.warning("unsupported action type: %s", action.get("type"))


def _redeploy(action: dict, playbook_name: str, metric_value: float):
    project_id = action.get("project_id")
    image = action.get("image")

    if not project_id or not image:
        log.error("redeploy action missing project_id or image in playbook %s", playbook_name)
        return

    log.info("remediating: playbook=%s project=%s metric_value=%.2f",
             playbook_name, project_id, metric_value)

    try:
        resp = requests.post(
            f"{CONTROL_PLANE_URL}/api/deploys",
            json={
                "project_id": project_id,
                "image": image,
                "commit_sha": f"auto-heal-{playbook_name}",
            },
            timeout=10,
        )
        resp.raise_for_status()
        deploy_id = resp.json()["id"]
        log.info("created remediation deploy %s", deploy_id)
    except Exception as exc:
        log.error("failed to create deploy: %s", exc)
        return

    try:
        resp = requests.post(
            f"{CONTROL_PLANE_URL}/api/deploys/{deploy_id}/apply",
            timeout=60,
        )
        resp.raise_for_status()
        log.info("remediation deploy %s applied: status=%s",
                 deploy_id, resp.json().get("status"))
    except Exception as exc:
        log.error("failed to apply deploy %s: %s", deploy_id, exc)
