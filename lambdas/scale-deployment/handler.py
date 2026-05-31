"""
Lambda: scale-deployment
Triggered by Step Functions when p99 latency exceeds threshold.
Action: kubectl scale deployment/app-<app> --replicas=current+delta
"""
from shared import audit, cp_client, notify


def lambda_handler(event: dict, _context) -> dict:
    playbook = event["playbook"]
    action = event.get("action", {})
    app = action.get("app", "")
    params = action.get("params", {})
    metric = event.get("metric", "http_latency_p99")
    metric_value = float(event.get("value", 0))

    if not app:
        raise ValueError("action.app is required — check playbook YAML")

    replicas_delta = int(params.get("replicas_delta", 1))
    replicas_exact = params.get("replicas")  # optional hard override

    notify.anomaly_detected(playbook, app, metric, metric_value)

    try:
        if replicas_exact is not None:
            result = cp_client.scale(app, replicas=int(replicas_exact))
        else:
            result = cp_client.scale(app, replicas_delta=replicas_delta)

        new_count = result.get("replicas", "?")
        outcome = f"scaled to {new_count} replicas"
        notify.healed(playbook, app, "scale", outcome)
        audit.log(playbook, app, "scale", metric_value, "success", outcome)
        return {"status": "ok", "app": app, "replicas": new_count}
    except Exception as exc:
        error = str(exc)
        notify.failed(playbook, app, "scale", error)
        audit.log(playbook, app, "scale", metric_value, "failed", error)
        raise
