#!/usr/bin/env bash
# Runs EVERY TIME the Codespace starts (including after sleep/resume).
# Starts k3s and Cloudflare Tunnel.
set -euo pipefail

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# ── k3s ────────────────────────────────────────────────────────────────────
if pgrep -x k3s > /dev/null; then
  echo "==> k3s already running"
else
  echo "==> Starting k3s..."
  # --snapshotter=native required in Codespaces — overlayfs not supported in nested containers
  sudo nohup k3s server \
    --disable=traefik \
    --disable=servicelb \
    --write-kubeconfig-mode=644 \
    --snapshotter=native \
    > /tmp/k3s.log 2>&1 &

  echo -n "==> Waiting for k3s to be ready"
  until kubectl get nodes 2>/dev/null | grep -q "Ready"; do
    echo -n "."
    sleep 3
  done
  echo " ready!"
fi

# Apply core manifests (namespace + control plane deployment)
echo "==> Applying K8s manifests..."
kubectl apply -f infrastructure/k8s/namespace.yaml 2>/dev/null || true

# ── Cloudflare Tunnel ──────────────────────────────────────────────────────
if [ -z "${CF_TUNNEL_TOKEN:-}" ]; then
  echo "==> WARNING: CF_TUNNEL_TOKEN secret not set."
  echo "    Set it in: github.com/settings/codespaces → Secrets"
  echo "    Skipping Cloudflare Tunnel — ports available via Codespace port forwarding only."
else
  if pgrep -x cloudflared > /dev/null; then
    echo "==> Cloudflare Tunnel already running"
  else
    echo "==> Starting Cloudflare Tunnel..."
    nohup cloudflared tunnel --no-autoupdate run --token "$CF_TUNNEL_TOKEN" \
      > /tmp/cloudflared.log 2>&1 &
    echo "==> Tunnel started. cloudforge.is-a.dev is now live."
  fi
fi

# ── Status summary ─────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║           CloudForge Codespace is ready              ║"
echo "╠══════════════════════════════════════════════════════╣"
kubectl get nodes 2>/dev/null | tail -n +2 | \
  awk '{printf "║  k3s node: %-42s ║\n", $1" ("$2")"}'
echo "╠══════════════════════════════════════════════════════╣"
echo "║  Port 8080  →  Control Plane API                     ║"
echo "║  Port 5173  →  Dashboard (npm run dev)               ║"
echo "║  Port 9090  →  Prometheus                            ║"
echo "║  Port 3001  →  Grafana                               ║"
echo "╠══════════════════════════════════════════════════════╣"
if [ -n "${CF_TUNNEL_TOKEN:-}" ]; then
echo "║  https://cloudforge.is-a.dev  →  LIVE via tunnel     ║"
else
echo "║  CF_TUNNEL_TOKEN not set — use port forwarding only  ║"
fi
echo "╚══════════════════════════════════════════════════════╝"
