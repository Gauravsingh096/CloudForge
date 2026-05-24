#!/usr/bin/env bash
set -euo pipefail

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# ── k3s ────────────────────────────────────────────────────────────────────
echo "==> Stopping any existing k3s instances..."
sudo pkill -x k3s 2>/dev/null || true
sleep 3

echo "==> Starting k3s..."
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

kubectl apply -f infrastructure/k8s/namespace.yaml 2>/dev/null || true

# ── Cloudflare Tunnel ──────────────────────────────────────────────────────
if [ -z "${CF_TUNNEL_TOKEN:-}" ]; then
  echo "==> WARNING: CF_TUNNEL_TOKEN not set — use port forwarding only."
else
  if pgrep -x cloudflared > /dev/null; then
    echo "==> Cloudflare Tunnel already running"
  else
    echo "==> Starting Cloudflare Tunnel..."
    nohup cloudflared tunnel --no-autoupdate run --token "$CF_TUNNEL_TOKEN" \
      > /tmp/cloudflared.log 2>&1 &
  fi
fi

echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║           CloudForge Codespace is ready              ║"
echo "╚══════════════════════════════════════════════════════╝"
