# deploy-infra.ps1
# Automates the EKS cluster bootstrapping and deployment of Helm charts

$ErrorActionPreference = "Stop"

Write-Host "=== EKS Bootstrapping & Deployment Script ===" -ForegroundColor Cyan

# 1. Check prerequisites
foreach ($cmd in @("aws", "kubectl", "helm")) {
    if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
        Write-Error "Prerequisite not found: $cmd. Please install it first."
    }
}

# 2. Configure EKS kubeconfig context
Write-Host "`n[1/5] Configuring kubectl context for envoy-cluster..." -ForegroundColor Yellow
aws eks update-kubeconfig --region ap-south-1 --name envoy-cluster
Write-Host "✅ kubectl configured." -ForegroundColor Green

# 3. Add required Helm repositories
Write-Host "`n[2/5] Adding Helm Repositories..." -ForegroundColor Yellow
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo add external-secrets https://charts.external-secrets.io
helm repo add warpstream https://charts.warpstream.com
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add stakater https://stakater.github.io/Reloader/
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
Write-Host "✅ Helm repos updated." -ForegroundColor Green

# 4. Install Infrastructure System Charts
Write-Host "`n[3/5] Installing Infrastructure Helper Charts..." -ForegroundColor Yellow

Write-Host "-> Installing Ingress Nginx..." -ForegroundColor Gray
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx `
  --namespace ingress-nginx --create-namespace

Write-Host "-> Installing Cert-Manager..." -ForegroundColor Gray
helm upgrade --install cert-manager jetstack/cert-manager `
  --namespace cert-manager --create-namespace `
  --set crds.enabled=true

Write-Host "-> Installing External Secrets Operator..." -ForegroundColor Gray
helm upgrade --install external-secrets external-secrets/external-secrets `
  --namespace external-secrets --create-namespace `
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::711396988882:role/prod-external-secrets-operator-irsa-role"

# Write-Host "-> Installing Prometheus Monitoring..." -ForegroundColor Gray
# helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack `
#   --namespace monitoring --create-namespace

Write-Host "-> Installing Argo CD..." -ForegroundColor Gray
helm upgrade --install argocd argo/argo-cd `
  --namespace argocd --create-namespace

Write-Host "-> Installing Stakater Reloader..." -ForegroundColor Gray
helm upgrade --install reloader stakater/reloader `
  --namespace reloader --create-namespace

Write-Host "-> Waiting for Infrastructure Services to be ready..." -ForegroundColor Gray

Write-Host "   Waiting for Ingress Nginx..." -ForegroundColor Gray
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx --timeout=120s

Write-Host "   Waiting for Cert-Manager..." -ForegroundColor Gray
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cert-manager -n cert-manager --timeout=120s

Write-Host "   Waiting for External Secrets Operator..." -ForegroundColor Gray
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=external-secrets -n external-secrets --timeout=120s

Write-Host "✅ Infrastructure system charts installed." -ForegroundColor Green

# 5. Install WarpStream Infrastructure & Agent
Write-Host "`n[4/5] Deploying WarpStream Streaming Infrastructure..." -ForegroundColor Yellow

Write-Host "-> Deploying WarpStream local configuration..." -ForegroundColor Gray
helm upgrade --install warpstream-infra deployments/helm/warpstream-infra `
  --namespace warpstream --create-namespace

Write-Host "-> Deploying WarpStream Agent..." -ForegroundColor Gray
helm upgrade --install warpstream-agent warpstream/warpstream-agent `
  -f deployments/helm/warpstream-infra/warpstream-values.yaml `
  --namespace warpstream --create-namespace

Write-Host "✅ WarpStream deployed." -ForegroundColor Green

# 6. Deploy Application Charts
Write-Host "`n[5/5] Deploying Applications..." -ForegroundColor Yellow

Write-Host "-> Deploying envoy-backend..." -ForegroundColor Gray
helm upgrade --install envoy-backend deployments/helm/envoy-backend `
  -f deployments/helm/envoy-backend/values-production.yaml `
  --namespace default

# Check if c:\billing-service exists to deploy the billing-service
if (Test-Path "c:\billing-service") {
    Write-Host "-> Deploying billing-service from c:\billing-service..." -ForegroundColor Gray
    Push-Location "c:\billing-service"
    try {
        helm upgrade --install billing-service deployments/helm/billing-service `
          -f deployments/helm/billing-service/values-prod.yaml `
          --namespace default
        Write-Host "✅ billing-service deployed successfully." -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
} else {
    Write-Warning "Could not find c:\billing-service. Skipping billing-service deployment."
}

Write-Host "`n=== All deployments triggered! check status using: kubectl get pods -A ===" -ForegroundColor Green
