# lambdas

AWS Lambda functions — remediation actions invoked by Step Functions state machines.

## Functions

| Function | Trigger | What it does |
|---|---|---|
| `restart-pod` | CrashLoop / OOM playbook | `kubectl rollout restart deployment/<app>` |
| `scale-deployment` | High-latency playbook | `kubectl scale deployment/<app> --replicas=N+1` |
| `rollback` | Error-spike playbook | `kubectl rollout undo deployment/<app>` |

Each function also:
- Writes an audit record to DynamoDB
- Sends a Discord webhook notification

## Structure

```
lambdas/
├── restart-pod/
│   ├── handler.py
│   └── requirements.txt
├── scale-deployment/
│   ├── handler.py
│   └── requirements.txt
└── rollback/
    ├── handler.py
    └── requirements.txt
```
