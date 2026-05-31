"""DynamoDB audit log — append-only record of every self-healing event."""
import os
import time
import uuid

import boto3

_TABLE = os.environ.get("AUDIT_TABLE", "cloudforge-audit-log")
_REGION = os.environ.get("AWS_REGION", "ap-south-1")
_TTL_SECONDS = 90 * 24 * 60 * 60  # 90-day auto-expiry (free, no ops)

_table = None


def _get_table():
    global _table
    if _table is None:
        _table = boto3.resource("dynamodb", region_name=_REGION).Table(_TABLE)
    return _table


def log(
    playbook: str,
    app: str,
    action: str,
    metric_value: float,
    outcome: str,
    message: str,
) -> None:
    now = int(time.time())
    try:
        _get_table().put_item(Item={
            "event_id":     str(uuid.uuid4()),
            "timestamp":    now,
            "ttl":          now + _TTL_SECONDS,
            "playbook":     playbook,
            "app":          app,
            "action":       action,
            "metric_value": str(round(metric_value, 6)),
            "outcome":      outcome,
            "message":      message[:1000],
        })
    except Exception as exc:
        print(f"audit.log failed (non-fatal): {exc}")
