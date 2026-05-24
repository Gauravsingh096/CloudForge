"""Z-score based anomaly detection over a rolling Prometheus metric window."""

import logging
from typing import Optional, Tuple

import numpy as np
import requests

log = logging.getLogger(__name__)

# How many historic data points to collect for z-score baseline
WINDOW_MINUTES = 10
STEP_SECONDS = 30


class AnomalyDetector:
    def __init__(self, prometheus_url: str):
        self.url = prometheus_url.rstrip("/")

    def check(
        self,
        metric: str,
        hard_threshold: Optional[float],
        zscore_threshold: float = 3.0,
    ) -> Tuple[Optional[float], bool]:
        """
        Returns (current_value, is_anomaly).
        Anomaly is True if:
          - current_value exceeds hard_threshold, OR
          - z-score of current_value vs rolling window > zscore_threshold
        """
        current = self._instant_query(metric)
        if current is None:
            return None, False

        if hard_threshold is not None and current > hard_threshold:
            log.debug("hard threshold breach metric=%s value=%.3f threshold=%.3f",
                      metric, current, hard_threshold)
            return current, True

        window = self._range_query(metric, minutes=WINDOW_MINUTES)
        if len(window) < 5:
            # not enough history for z-score
            return current, False

        mean = np.mean(window)
        std = np.std(window)
        if std == 0:
            return current, False

        z = abs((current - mean) / std)
        is_anomaly = z > zscore_threshold
        log.debug("metric=%s value=%.3f mean=%.3f std=%.3f z=%.2f anomaly=%s",
                  metric, current, mean, std, z, is_anomaly)
        return current, is_anomaly

    def _instant_query(self, query: str) -> Optional[float]:
        try:
            resp = requests.get(
                f"{self.url}/api/v1/query",
                params={"query": query},
                timeout=5,
            )
            resp.raise_for_status()
            result = resp.json()["data"]["result"]
            if not result:
                return None
            return float(result[0]["value"][1])
        except Exception as exc:
            log.warning("prometheus instant query failed: %s", exc)
            return None

    def _range_query(self, query: str, minutes: int) -> list[float]:
        try:
            import time
            end = int(time.time())
            start = end - minutes * 60
            resp = requests.get(
                f"{self.url}/api/v1/query_range",
                params={
                    "query": query,
                    "start": start,
                    "end": end,
                    "step": STEP_SECONDS,
                },
                timeout=5,
            )
            resp.raise_for_status()
            result = resp.json()["data"]["result"]
            if not result:
                return []
            return [float(v[1]) for v in result[0]["values"]]
        except Exception as exc:
            log.warning("prometheus range query failed: %s", exc)
            return []
