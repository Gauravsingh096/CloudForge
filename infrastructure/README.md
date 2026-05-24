# infrastructure

Terraform for AWS resources + Kubernetes YAML manifests.

## Structure

```
infrastructure/
├── aws/
│   ├── stepfunctions.tf     Step Functions state machines (4 playbooks)
│   ├── lambda.tf            Lambda functions + IAM roles
│   ├── dynamodb.tf          Audit log table
│   └── variables.tf
└── k8s/
    ├── namespace.yaml       cloudforge-system namespace
    ├── ingress.yaml         NGINX Ingress class + wildcard rule
    ├── prometheus-values.yaml   Helm values for kube-prometheus-stack
    └── grafana-values.yaml
```

## Apply

```bash
# AWS resources
cd infrastructure/aws
terraform init
terraform apply

# K8s manifests
kubectl apply -f infrastructure/k8s/namespace.yaml
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -f infrastructure/k8s/ingress.yaml
helm upgrade --install kube-prometheus kube-prometheus-stack/kube-prometheus-stack \
  -f infrastructure/k8s/prometheus-values.yaml
```
