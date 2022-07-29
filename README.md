# kubelet-summary-exporter

A prometheus-exporter for stats (ephemeral volumes) from kubelet


Expected to run as a daemonset and pull from `<nodeip>:10255/stats/summary`

### Development

Running locally:

1. Docker Desktop (with kubernetes)
2. skaffold https://skaffold.dev/docs/install/
3. `skaffold dev`
4. `kubectl -n kubelet-summary-exporter port-forward ds/kubelet-summary-exporter 9091`
5. `curl localhost:9091/metrics`

### Deploying

Example manifests are in config: `kubectl kustomize config/base/`
