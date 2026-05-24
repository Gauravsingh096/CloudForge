# control-plane

Go API — orchestrates builds, deploys, and project management on k3s.

## Endpoints (planned)

| Method | Path | Description |
|---|---|---|
| GET | /healthz | Health check |
| GET | /api/projects | List all projects |
| POST | /api/projects | Create a new project |
| GET | /api/projects/:id | Get project details |
| GET | /api/projects/:id/deploys | List deploys for a project |
| POST | /api/deploys | Trigger a deploy |
| GET | /api/deploys/:id | Get deploy status |
| POST | /api/webhook/github | Receive GitHub push webhook |

## Structure

```
control-plane/
├── cmd/server/main.go       Entry point
├── internal/
│   ├── api/                 HTTP handlers
│   ├── auth/                GitHub OAuth
│   ├── deploys/             Deploy orchestration logic
│   └── db/                  Postgres queries
├── Dockerfile
└── go.mod
```
