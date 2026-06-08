# CloudForge

> A self-hosted, self-healing Platform-as-a-Service on Kubernetes — push code, get a live URL, watch the platform auto-remediate incidents.

[![CI](https://github.com/Gauravsingh096/CloudForge/actions/workflows/ci.yml/badge.svg)](https://github.com/Gauravsingh096/CloudForge/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## What is CloudForge ?

CloudForge is a mini Heroku / Vercel — but open, self-hosted, and wired for reliability:

| Capability | How |
|---|---|
| **Git-push deploys** | `git push` → GitHub Actions builds → pod live in under 3 min |
| **Custom HTTPS subdomains** | Every app gets `<name>.cloudforge.is-a.dev` automatically |
| **Observable by default** | Prometheus + Grafana + Loki per app, zero config |
| **Self-healing** | Python detector + AWS Step Functions auto-remediates OOM / CrashLoop / latency spikes |
| **Regression gate** | it records prod traffic, replays against canary — blocks bad deploys automatically |

---

## Architecture

```
Developer
  │ git push
  ▼
GitHub Actions ──► builds image ──► ghcr.io
  │ SSH deploy
  ▼
GitHub Codespaces  (k3s, 4 cores / 8 GB, free with GitHub Pro)
  ├── CloudForge Control Plane  (Go API)
  ├── NGINX Ingress  ◄── Cloudflare Tunnel  ◄── Browser
  ├── User app pods
  ├── Prometheus + Grafana + Loki
  └── Detector (Python)
        │ anomaly
        ▼
      AWS Step Functions → Lambda → kubectl + Discord + DynamoDB
```

Full diagram → [ARCHITECTURE.md](ARCHITECTURE.md)

---

## Stack (100% free tier, ₹0/month)

| Layer | Technology |
|---|---|
| Compute | GitHub Codespaces — 4 vCPU, 8 GB RAM (free with GitHub Pro) |
| Kubernetes | k3s (single-node, official Rancher) |
| Control plane API | Go (net/http + chi router) |
| Database | Neon Postgres (serverless, scales to zero) |
| Cache / Queue | Upstash Redis |
| Observability | Prometheus + Grafana + Loki (self-hosted) |
| Self-healing workflow | AWS Step Functions + Lambda |
| Audit log | AWS DynamoDB |
| DNS / CDN / Tunnel | Cloudflare (free) |
| CI/CD | GitHub Actions + ghcr.io |
| Frontend | React (Vite) on Vercel |
| Auth | GitHub OAuth |

---

## Project Structure

```
cloudforge/
├── control-plane/       Go API — deploy orchestration, project management
├── detector/            Python — anomaly detection + self-healing playbooks
├── dashboard/           React — developer UI
├── lambdas/             AWS Lambda — kubectl remediation actions
├── infrastructure/      Terraform (AWS) + Kubernetes YAML manifests
├── scripts/             Bootstrap scripts
└── .github/workflows/   CI + deploy pipelines
```

---

## Roadmap

- [x] Repo scaffold + CI skeleton
- [x] Go control plane — projects + deploys API live
- [x] Neon Postgres connected — data persisted
- [x] k3s running on GitHub Codespaces (₹0)
- [ ] GitHub webhook → build → deploy pipeline
- [ ] Prometheus + Grafana observability
- [ ] Self-healing engine (detector + Step Functions + Lambda)

---

## Development

> Runs on GitHub Codespaces — no local setup needed. Open the repo and click **Code → Codespaces → Create codespace on main**.

k3s starts automatically. Control plane connects to Neon Postgres via `DATABASE_URL` secret.

```bash
# Start control plane
cd control-plane && go run ./cmd/server

# Test
curl http://localhost:8080/healthz
curl http://localhost:8080/api/projects
```

---

## License

MIT
