# k8s-memory-watchdog

This tool monitors the total memory usage of pods within a specified Kubernetes namespace.
If the total memory exceeds a configured threshold, it will restart a specified deployment.

## Features

- Monitors memory usage via `kubectl top pods`
- Automatically restarts a deployment when memory usage exceeds a defined threshold
- Configurable via command-line flags or environment variables
- Verbose logging option

## Usage

```bash
go run main.go --namespace=my-namespace --deployment=my-app --threshold=5000
```

Or using environment variables:

```bash
export NAMESPACE=my-namespace
export DEPLOYMENT=my-app
export MEMORY_THRESHOLD=5000
go run main.go --verbose
```

## Flags

| Flag           | Description                                                  |
|----------------|--------------------------------------------------------------|
| `--namespace`  | Kubernetes namespace where pods are running                  |
| `--deployment` | Name of the deployment to restart when memory limit is hit   |
| `--threshold`  | Memory threshold in MiB (e.g., 5000 for 5Gi)                 |
| `--kubectl`    | Path to the `kubectl` binary (default: `/usr/local/bin/kubectl`) |
| `--verbose`    | Enable verbose logging                                       |

## Environment Variables

- `NAMESPACE`
- `DEPLOYMENT`
- `MEMORY_THRESHOLD`
- `KUBECTL_PATH`

These can be used as alternatives to the flags.

## Example Cronjob (Linux)

To run the script every 10 minutes:

```cron
*/10 * * * * /usr/local/go/bin/go run /path/to/main.go --namespace=my-namespace --deployment=my-app --threshold=5000
```