# Kustom Controller Helm Chart

Helm chart for deploying the Kustom Controller which enforces resource limits/requests on Kubernetes pods.

## Installation

### Add Helm Repository
```bash
helm repo add kustom-controller https://raw.githubusercontent.com/vishalanarase/kustom-controller/master/charts
helm repo update
helm search repo kustom-controller
helm install kustom-controller kustom-controller/kustom-controller
```

## Package chart
```bash
cd charts
helm package ../helm/chart/
helm repo index charts --url https://raw.githubusercontent.com/vishalanarase/kustom-controller/master/charts
```