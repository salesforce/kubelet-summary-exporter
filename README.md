# kubelet-summary-exporter

A prometheus-exporter for stats (ephemeral volumes) from kubelet


Expected to run as a daemonset and pull from `<nodeip>:10255/stats/summary`

### Development

Running locally:

1. Docker Desktop (with kubernetes)
2. skaffold https://skaffold.dev/docs/install/
3. `skaffold dev -p local`
4. `kubectl -n kubelet-summary-exporter port-forward ds/kubelet-summary-exporter 9091`
5. `curl localhost:9091/metrics`

### Deploying

Example manifests are in config: `kubectl kustomize config/base/`

### Configuration

```
Usage: kubelet-summary-exporter

Flags:
  -h, --help                   Show context-sensitive help.
      --prom-listen=":9091"    Address to listen for for Prometheus metrics
      --node-host=STRING       Address to request kubelet's stats/summary from ($NODE_HOST)
      --insecure               Don't validate certificates ($INSECURE)
      --ca=STRING              Certificate location ($CA_CRT)
      --token-path=STRING      Token location ($TOKEN)
      --timeout=5s             Timeout for requests ($TIMEOUT)
      --look-up-hostname       Use api-server to deterimine hostname (assumes in cluster config) ($LOOK_UP_HOSTNAME)
```
