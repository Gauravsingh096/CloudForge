# detector

Python service — polls Prometheus every 30s, detects anomalies via z-score, triggers AWS Step Functions remediation workflows.

## Playbooks

| Playbook | Trigger | Action |
|---|---|---|
| `pod-oom.yaml` | Memory > 90% or OOMKilled | Restart pod |
| `crashloop.yaml` | CrashLoopBackOff detected | Restart + alert |
| `high-latency.yaml` | p99 latency > 2s for 2 min | Scale up replicas |
| `error-spike.yaml` | Error rate > 10% for 2 min | Rollback deployment |

## Structure

```
detector/
├── detector.py          Main polling loop
├── anomaly.py           Z-score detection logic
├── stepfunctions.py     AWS Step Functions trigger
├── playbooks/
│   ├── pod-oom.yaml
│   ├── crashloop.yaml
│   ├── high-latency.yaml
│   └── error-spike.yaml
├── Dockerfile
└── requirements.txt
```
