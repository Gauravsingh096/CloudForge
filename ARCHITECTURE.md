# CloudForge — Architecture

## System Overview

CloudForge is a single-tenant PaaS built on k3s (lightweight Kubernetes) running on a single Oracle Cloud ARM VM. All infrastructure is free-tier.

---

## Component Map

```
┌─────────────────────────────────────────────────────────────────────┐
│  Developer Laptop                                                    │
│  git push → github.com/Gauravsingh096/CloudForge                    │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ webhook
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  GitHub Actions                                                      │
│  1. docker build (Kaniko or buildx)                                  │
│  2. docker push → ghcr.io                                            │
│  3. SSH → Oracle VM → kubectl apply                                  │
└──────────────────────────┬──────────────────────────────────────────┘
                           │ SSH deploy
                           ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Oracle Cloud ARM VM  (ap-mumbai-1, 4 OCPU, 24 GB, free forever)    │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  k3s Cluster                                                 │   │
│  │                                                              │   │
│  │  ┌─────────────────┐   ┌──────────────────────────────────┐ │   │
│  │  │ CloudForge CP   │   │ NGINX Ingress Controller         │ │   │
│  │  │ (Go API)        │   │ *.cloudforge.is-a.dev → pods     │ │   │
│  │  │ :8080           │   └──────────────────────────────────┘ │   │
│  │  └────────┬────────┘                    ▲                    │   │
│  │           │                             │ Cloudflare Tunnel  │   │
│  │     CRUD  │                             │ (no open ports)    │   │
│  │           ▼                             │                    │   │
│  │  ┌──────────────┐            ┌──────────┴───────────────┐   │   │
│  │  │ Neon Postgres │            │ User App Pods            │   │   │
│  │  │ (external)    │            │ (one pod per project)    │   │   │
│  │  └──────────────┘            └──────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌──────────────────────────────────────────────────────┐   │   │
│  │  │ Observability Stack                                   │   │   │
│  │  │  Prometheus  ←  kube-state-metrics + node-exporter   │   │   │
│  │  │  Grafana     ←  Prometheus datasource                 │   │   │
│  │  │  Loki        ←  Promtail (all pod logs)               │   │   │
│  │  └──────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌──────────────────────────────────────────────────────┐   │   │
│  │  │ Detector Service (Python)                             │   │   │
│  │  │  polls Prometheus every 30s                           │   │   │
│  │  │  z-score anomaly detection on CPU/mem/RPS/latency    │   │   │
│  │  └──────────────────┬───────────────────────────────────┘   │   │
│  └─────────────────────┼──────────────────────────────────────┘   │
└────────────────────────┼────────────────────────────────────────────┘
                         │ anomaly event
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│  AWS (free tier)                                                     │
│                                                                      │
│  Step Functions ──► Lambda (restart-pod / scale / rollback)         │
│       │                 │                                            │
│       │                 ├──► DynamoDB (audit log)                    │
│       │                 └──► Discord webhook (alert)                 │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow — Deploy Lifecycle

```
1. Developer runs:  git push origin main
2. GitHub webhook fires → GitHub Actions workflow starts
3. Actions: docker build → docker push ghcr.io/<image>:<sha>
4. Actions: SSH into Oracle VM
5. Control plane API: POST /api/deploys  {project, image, sha}
6. Control plane generates K8s Deployment + Service + Ingress YAML
7. kubectl apply → k3s schedules pod
8. NGINX Ingress picks up new Ingress rule
9. Cloudflare DNS *.cloudforge.is-a.dev → Cloudflare Tunnel → NGINX → pod
10. Control plane writes deploy record to Neon Postgres
11. Deploy status: RUNNING
```

---

## Self-Healing Flow

```
1. Detector polls Prometheus every 30s
2. Metric breaches threshold (z-score > 3σ or hard threshold)
3. Detector → AWS Step Functions: StartExecution with {app, metric, value, playbook}
4. Step Functions state machine:
   a. Wait 60s (confirm it's not transient)
   b. Re-check metric
   c. If still breached → invoke Lambda
5. Lambda action (based on playbook):
   - pod-oom:         kubectl rollout restart deployment/<app>
   - crashloop:       kubectl rollout restart + check image
   - high-latency:    kubectl scale deployment/<app> --replicas=+1
   - error-spike:     kubectl rollout undo deployment/<app>
6. Lambda → DynamoDB: write audit record
7. Lambda → Discord webhook: send alert
```

---

## Keploy Regression Gate Flow

```
1. New deploy triggers canary pod (new image, 0 traffic)
2. Keploy sidecar replays recorded API traffic → canary pod
3. Compare responses: status codes, body diff, latency
4. Regression score = % of requests with diff
5. Score < threshold (5%) → promote canary to production
6. Score ≥ threshold → block promotion, alert developer, keep old version
```

---

## Key Tech Decisions

| Decision | Choice | Why not the alternative |
|---|---|---|
| K8s distro | k3s | EKS = $73/mo control plane fee |
| VM host | Oracle Cloud ARM | AWS EC2 free tier = only 12 months + tiny |
| DB | Neon Postgres | Doesn't burn Oracle VM RAM; serverless |
| Image registry | ghcr.io | ECR = 500 MB limit; ghcr free for public |
| Tunnel | Cloudflare Tunnel | No open ports on Oracle VM required |
| Workflow | Step Functions | 4K transitions/mo free; native AWS story |
| Anomaly | z-score | ML is over-engineering for this scale |
