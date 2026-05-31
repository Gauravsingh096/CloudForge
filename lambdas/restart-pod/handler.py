"""
Lambda: restart-pod
Triggered by Step Functions when pod OOM or memory spike is detected.
Action: kubectl rollout restart deployment/app-<app>
"""
from shared import audit, cp_client, notify


def lambda_handler(event: dict, _context) -> dict:
    playbook = event["playbook"]
    action = event.get("action", {})
    app = action.get("app", "")
    metric = event.get("metric", "memory_usage")
    metric_value = float(event.get("value", 0))

    if not app:
        raise ValueError("action.app is required — check playbook YAML")

    notify.anomaly_detected(playbook, app, metric, metric_value)

    try:
        result = cp_client.restart(app)
        outcome = result.get("output", "restarted")
        notify.healed(playbook, app, "restart", outcome)
        audit.log(playbook, app, "restart", metric_value, "success", outcome)
        return {"status": "ok", "app": app, "output": outcome}
    except Exception as exc:
        error = str(exc)
        notify.failed(playbook, app, "restart", error)
        audit.log(playbook, app, "restart", metric_value, "failed", error)
        raise
