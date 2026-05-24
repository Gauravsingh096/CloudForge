# scripts

One-time bootstrap scripts for Oracle VM and k3s cluster setup.

## Scripts

| Script | When to run | What it does |
|---|---|---|
| `bootstrap-vm.sh` | Once, after Oracle VM is provisioned | Installs k3s, Helm, k9s, cloudflared, opens iptables ports |
| `install-cluster.sh` | Once, after bootstrap | Installs NGINX Ingress, cert-manager, Prometheus stack via Helm |

## Usage

```bash
# On Oracle VM (via SSH)
chmod +x scripts/bootstrap-vm.sh
./scripts/bootstrap-vm.sh

# After bootstrap completes
chmod +x scripts/install-cluster.sh
./scripts/install-cluster.sh
```
