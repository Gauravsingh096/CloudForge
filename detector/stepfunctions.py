"""Trigger AWS Step Functions state machine when an anomaly is detected."""

import json
import logging
import os
import time

import boto3

log = logging.getLogger(__name__)

AWS_REGION = os.environ.get("AWS_REGION", "ap-south-1")
STATE_MACHINE_ARN_PREFIX = os.environ.get("STATE_MACHINE_ARN_PREFIX", "")


def trigger_remediation(playbook: dict, metric_value: float):
    sfn = boto3.client("stepfunctions", region_name=AWS_REGION)

    state_machine_arn = playbook.get("state_machine_arn") or (
        STATE_MACHINE_ARN_PREFIX + playbook["name"]
    )

    input_payload = {
        "playbook": playbook["name"],
        "metric": playbook["trigger"]["metric"],
        "value": metric_value,
        "action": playbook["action"],
        "timestamp": int(time.time()),
    }

    try:
        resp = sfn.start_execution(
            stateMachineArn=state_machine_arn,
            input=json.dumps(input_payload),
        )
        log.info(
            "step functions execution started: playbook=%s arn=%s",
            playbook["name"],
            resp["executionArn"],
        )
    except Exception as exc:
        log.error("failed to start step functions execution: %s", exc)
