"""
Lambda: rollback
Triggered by Step Functions when HTTP 5xx error rate exceeds threshold.
Action: kubectl rollout undo deployment/app-<app>
"""
from shared import audit, cp_client, notify


def lambda_handler(event: dict, _context) -> dict:
    playbook = event["playbook"]
    action = event.get("action", {})
    app = action.get("app", "")
    metric = event.get("metric", "http_error_rate")
    metric_value = float(event.get("value", 0))

    if not app:
        raise ValueError("action.app is required — check playbook YAML")

    notify.anomaly_detected(playbook, app, metric, metric_value)

    try:
        result = cp_client.rollback(app)
        outcome = result.get("output", "rolled back")
        notify.healed(playbook, app, "rollback", outcome)
        audit.log(playbook, app, "rollback", metric_value, "success", outcome)
        return {"status": "ok", "app": app, "output": outcome}
    except Exception as exc:
        error = str(exc)
        notify.failed(playbook, app, "rollback", error)
        audit.log(playbook, app, "rollback", metric_value, "failed", error)
        raise
