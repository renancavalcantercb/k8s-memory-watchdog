# Kubernetes Memory Watchdog

A service that monitors memory usage of pods in a Kubernetes namespace and automatically restarts deployments when memory usage exceeds a configured threshold.

## Features

- Continuous monitoring of pod memory usage
- Automatic deployment restart when memory limit is exceeded
- Flexible configuration via YAML file or environment variables
- Prometheus metrics support
- Configurable logging
- Graceful shutdown
- Unit tests

## Requirements

- Go 1.16 or higher
- Access to a Kubernetes cluster
- kubectl configured and accessible

## Installation

```bash
go install github.com/your-username/k8s-memory-watchdog@latest
```

## Configuration

The watchdog can be configured through:

1. YAML configuration file
2. Environment variables
3. Command-line flags

### Example usage with flags

```bash
k8s-memory-watchdog --namespace=my-namespace --deployment=my-app --threshold=5000 --interval=5m
```

### Environment variables

- `NAMESPACE`: Kubernetes namespace (default: "default")
- `DEPLOYMENT`: Name of the deployment to monitor
- `MEMORY_THRESHOLD`: Memory threshold in Mi (default: 5000)
- `KUBECTL_PATH`: Path to kubectl binary (default: "/usr/local/bin/kubectl")
- `CHECK_INTERVAL`: Check interval (default: "5m")
- `VERBOSE`: Enable verbose logging (default: false)

### Configuration file

See `config.yaml` for all available configuration options.

## Metrics

The service exposes Prometheus metrics at `/metrics` when enabled:

- `k8s_memory_watchdog_memory_usage`: Current memory usage in Mi
- `k8s_memory_watchdog_deployment_restarts_total`: Total number of restarts
- `k8s_memory_watchdog_checks_total`: Total number of checks

## Logging

Logging can be configured for:

- Level: debug, info, warn, error
- Format: text or json
- Output: stdout or file

## Development

### Run tests

```bash
go test -v ./...
```

### Build binary

```bash
go build -o k8s-memory-watchdog
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## License

MIT