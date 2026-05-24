package deploys

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"text/template"
	"time"

	"github.com/Gauravsingh096/cloudforge/control-plane/internal/metrics"
)

type Deployer struct {
	db *sql.DB
}

func New(db *sql.DB) *Deployer {
	return &Deployer{db: db}
}

type AppConfig struct {
	Name      string
	Namespace string
	Image     string
	Subdomain string
	DeployID  string
}

var deploymentTmpl = template.Must(template.New("deploy").Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-{{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
    managed-by: cloudforge
    deploy-id: "{{.DeployID}}"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
        - name: app
          image: {{.Image}}
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: app-{{.Name}}
  namespace: {{.Namespace}}
spec:
  selector:
    app: {{.Name}}
  ports:
    - port: 80
      targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-{{.Name}}
  namespace: {{.Namespace}}
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
    - host: {{.Subdomain}}.cloudforge.is-a.dev
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app-{{.Name}}
                port:
                  number: 80
`))

// Apply generates K8s manifests for the app and runs kubectl apply.
func (d *Deployer) Apply(ctx context.Context, deployID string) error {
	start := time.Now()

	// fetch deploy + project info
	var appName, image, subdomain string
	err := d.db.QueryRowContext(ctx, `
		SELECT p.name, dep.image, p.subdomain
		FROM deploys dep
		JOIN projects p ON p.id = dep.project_id
		WHERE dep.id = $1`, deployID,
	).Scan(&appName, &image, &subdomain)
	if err != nil {
		return fmt.Errorf("fetch deploy: %w", err)
	}

	cfg := AppConfig{
		Name:      appName,
		Namespace: "user-apps",
		Image:     image,
		Subdomain: subdomain,
		DeployID:  deployID,
	}

	// render manifest
	var buf bytes.Buffer
	if err := deploymentTmpl.Execute(&buf, cfg); err != nil {
		return fmt.Errorf("render manifest: %w", err)
	}

	d.setStatus(ctx, deployID, "deploying")

	// ensure namespace exists
	exec.CommandContext(ctx, "kubectl", "create", "namespace", "user-apps",
		"--dry-run=client", "-o", "yaml").Run()
	exec.CommandContext(ctx, "kubectl", "apply", "-f", "-",
		"--dry-run=client").Run()

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = &buf
	out, err := cmd.CombinedOutput()
	if err != nil {
		d.setStatus(ctx, deployID, "failed")
		metrics.DeployApplyDuration.WithLabelValues(appName, "failed").Observe(time.Since(start).Seconds())
		metrics.DeploysTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("kubectl apply: %w\n%s", err, string(out))
	}

	log.Printf("deployer: applied deploy=%s app=%s image=%s\n%s",
		deployID, appName, image, string(out))
	d.setStatus(ctx, deployID, "running")
	metrics.DeployApplyDuration.WithLabelValues(appName, "running").Observe(time.Since(start).Seconds())
	metrics.DeploysTotal.WithLabelValues("running").Inc()
	return nil
}

func (d *Deployer) setStatus(ctx context.Context, deployID, status string) {
	_, err := d.db.ExecContext(ctx,
		`UPDATE deploys SET status = $1 WHERE id = $2`, status, deployID)
	if err != nil {
		log.Printf("deployer: failed to set status %s for %s: %v", status, deployID, err)
	}
}
