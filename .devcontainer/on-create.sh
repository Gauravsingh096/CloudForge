#!/usr/bin/env bash
# Runs ONCE when the Codespace is first created.
# Installs: k3s, cloudflared, k9s, Python deps, Go tools.
set -euo pipefail

echo "==> [on-create] Starting one-time Codespace setup..."

# ── k3s ────────────────────────────────────────────────────────────────────
echo "==> Installing k3s..."
curl -sfL https://get.k3s.io | \
  INSTALL_K3S_SKIP_ENABLE=true \
  INSTALL_K3S_SKIP_START=true \
  sh -s - \
    --disable=traefik \
    --disable=servicelb \
    --write-kubeconfig-mode=644

# ── cloudflared ────────────────────────────────────────────────────────────
echo "==> Installing cloudflared..."
curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg \
  | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] \
  https://pkg.cloudflare.com/cloudflared $(lsb_release -cs) main" \
  | sudo tee /etc/apt/sources.list.d/cloudflared.list
sudo apt-get update -qq
sudo apt-get install -y -qq cloudflared

# ── k9s ────────────────────────────────────────────────────────────────────
echo "==> Installing k9s..."
K9S_VERSION=$(curl -s https://api.github.com/repos/derailed/k9s/releases/latest \
  | grep '"tag_name"' | cut -d'"' -f4)
curl -fsSL "https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_Linux_amd64.tar.gz" \
  | sudo tar -xz -C /usr/local/bin k9s

# ── Python dependencies ────────────────────────────────────────────────────
echo "==> Installing Python deps for detector..."
pip install --quiet -r detector/requirements.txt

# ── Go tools ──────────────────────────────────────────────────────────────
echo "==> Installing Go tools..."
go install github.com/air-verse/air@latest          # live reload for Go
go install github.com/pressly/goose/v3/cmd/goose@latest  # DB migrations

# ── KUBECONFIG in shell profile ────────────────────────────────────────────
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.profile

echo ""
echo "==> [on-create] Done. Codespace is ready."
echo "    k3s, cloudflared, k9s, Go tools all installed."
echo "    k3s will auto-start on next Codespace boot (post-start.sh)."
