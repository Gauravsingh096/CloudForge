"""
CloudForge Detector — polls Prometheus every 30s, runs z-score anomaly
detection on key metrics, and triggers Step Functions remediation workflows.
"""

import logging
import os
import time
from typing import Optional

import yaml

from anomaly import AnomalyDetector
from stepfunctions import trigger_remediation

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
log = logging.getLogger(__name__)

PROMETHEUS_URL = os.environ.get("PROMETHEUS_URL", "http://localhost:9090")
POLL_INTERVAL = int(os.environ.get("POLL_INTERVAL_SECONDS", "30"))
PLAYBOOKS_DIR = os.path.join(os.path.dirname(__file__), "playbooks")


def load_playbooks() -> list[dict]:
    playbooks = []
    for fname in os.listdir(PLAYBOOKS_DIR):
        if fname.endswith(".yaml"):
            with open(os.path.join(PLAYBOOKS_DIR, fname)) as f:
                playbooks.append(yaml.safe_load(f))
    log.info("loaded %d playbooks", len(playbooks))
    return playbooks


def run():
    playbooks = load_playbooks()
    detector = AnomalyDetector(prometheus_url=PROMETHEUS_URL)

    log.info("detector starting — polling every %ds", POLL_INTERVAL)
    while True:
        try:
            check_all(detector, playbooks)
        except Exception as exc:
            log.error("check cycle failed: %s", exc)
        time.sleep(POLL_INTERVAL)


def check_all(detector: AnomalyDetector, playbooks: list[dict]):
    for playbook in playbooks:
        name = playbook["name"]
        metric = playbook["trigger"]["metric"]
        threshold = playbook["trigger"].get("threshold")
        zscore_threshold = playbook["trigger"].get("zscore", 3.0)

        value, is_anomaly = detector.check(
            metric=metric,
            hard_threshold=threshold,
            zscore_threshold=zscore_threshold,
        )

        if value is None:
            log.debug("playbook=%s metric=%s no data", name, metric)
            continue

        if is_anomaly:
            log.warning(
                "ANOMALY detected playbook=%s metric=%s value=%.3f",
                name, metric, value,
            )
            trigger_remediation(playbook=playbook, metric_value=value)
        else:
            log.debug("playbook=%s metric=%s value=%.3f OK", name, metric, value)


if __name__ == "__main__":
    run()
