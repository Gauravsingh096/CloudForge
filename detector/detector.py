"""
CloudForge Detector — polls Prometheus + k8s every 30s, runs z-score anomaly
detection on key metrics, and triggers self-healing redeploys via control plane.
"""

import json
import logging
import os
import subprocess
import time
from collections import defaultdict

import numpy as np
import yaml

from anomaly import AnomalyDetector
from remediate import trigger_remediation

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
log = logging.getLogger(__name__)

PROMETHEUS_URL = os.environ.get("PROMETHEUS_URL", "http://localhost:9090")
CONTROL_PLANE_URL = os.environ.get("CONTROL_PLANE_URL", "http://localhost:8080")
POLL_INTERVAL = int(os.environ.get("POLL_INTERVAL_SECONDS", "30"))
RESTART_THRESHOLD = int(os.environ.get("RESTART_THRESHOLD", "5"))
KUBECONFIG = os.environ.get("KUBECONFIG", "/etc/rancher/k3s/k3s.yaml")
PLAYBOOKS_DIR = os.path.join(os.path.dirname(__file__), "playbooks")

# Rolling restart history per app for z-score baseline
_restart_history: dict[str, list[int]] = defaultdict(list)
HISTORY_WINDOW = 20


def load_playbooks() -> list[dict]:
    playbooks = []
    for fname in sorted(os.listdir(PLAYBOOKS_DIR)):
        if fname.endswith(".yaml"):
            with open(os.path.join(PLAYBOOKS_DIR, fname)) as f:
                playbooks.append(yaml.safe_load(f))
    log.info("loaded %d playbooks", len(playbooks))
    return playbooks


def get_pod_restarts(namespace: str = "user-apps") -> dict[str, int]:
    """Returns {app_label: total_restart_count} by querying kubectl."""
    try:
        out = subprocess.check_output(
            ["kubectl", "get", "pods", "-n", namespace, "-o", "json"],
            timeout=10,
            env={**os.environ, "KUBECONFIG": KUBECONFIG},
        )
        pods = json.loads(out)
        app_restarts: dict[str, int] = defaultdict(int)
        for pod in pods.get("items", []):
            app = pod["metadata"].get("labels", {}).get("app", pod["metadata"]["name"])
            for cs in pod.get("status", {}).get("containerStatuses", []):
                app_restarts[app] += cs.get("restartCount", 0)
        return dict(app_restarts)
    except Exception as exc:
        log.warning("kubectl get pods failed: %s", exc)
        return {}


def check_restarts(playbooks: list[dict]):
    """Detect pod crash-loops via kubectl restart counts and redeploy if anomalous."""
    restarts = get_pod_restarts()
    if not restarts:
        return

    redeploy_by_app = {
        pb["action"].get("app"): pb
        for pb in playbooks
        if pb.get("action", {}).get("type") == "redeploy"
    }

    for app, count in restarts.items():
        history = _restart_history[app]
        history.append(count)
        if len(history) > HISTORY_WINDOW:
            history.pop(0)

        is_anomaly = count >= RESTART_THRESHOLD

        if not is_anomaly and len(history) >= 5:
            mean = np.mean(history[:-1])
            std = np.std(history[:-1])
            if std > 0:
                z = abs((count - mean) / std)
                if z > 3.0:
                    is_anomaly = True
                    log.info("z-score anomaly app=%s restarts=%d z=%.2f", app, count, z)

        if is_anomaly:
            log.warning("ANOMALY app=%s restarts=%d — triggering redeploy", app, count)
            playbook = redeploy_by_app.get(app)
            if playbook:
                trigger_remediation(playbook=playbook, metric_value=float(count))
            else:
                log.warning("no redeploy playbook configured for app=%s", app)
        else:
            log.debug("app=%s restarts=%d OK", app, count)


def check_prometheus(detector: AnomalyDetector, playbooks: list[dict]):
    """Check Prometheus-based metrics (latency, error rate) against playbooks."""
    for playbook in playbooks:
        if playbook.get("action", {}).get("type") == "redeploy":
            continue  # handled by check_restarts

        name = playbook["name"]
        trigger = playbook.get("trigger", {})
        metric = trigger.get("metric")
        if not metric:
            continue

        value, is_anomaly = detector.check(
            metric=metric,
            hard_threshold=trigger.get("threshold"),
            zscore_threshold=trigger.get("zscore", 3.0),
        )

        if value is None:
            log.debug("playbook=%s metric=%s no data", name, metric)
            continue

        if is_anomaly:
            log.warning("ANOMALY playbook=%s metric=%s value=%.3f", name, metric, value)
            trigger_remediation(playbook=playbook, metric_value=value)
        else:
            log.debug("playbook=%s metric=%s value=%.3f OK", name, metric, value)


def run():
    playbooks = load_playbooks()
    detector = AnomalyDetector(prometheus_url=PROMETHEUS_URL)

    log.info("detector starting — polling every %ds", POLL_INTERVAL)
    while True:
        try:
            check_prometheus(detector, playbooks)
            check_restarts(playbooks)
        except Exception as exc:
            log.error("check cycle failed: %s", exc)
        time.sleep(POLL_INTERVAL)


if __name__ == "__main__":
    run()
