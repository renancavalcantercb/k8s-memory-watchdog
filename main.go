package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Config represents the watchdog configuration
type Config struct {
	Namespace       string
	DeploymentName  string
	MemoryThreshold int
	KubectlPath     string
	Verbose         bool
	CheckInterval   time.Duration
}

// KubernetesClient interface for Kubernetes operations
type KubernetesClient interface {
	GetPodMemoryUsage(ctx context.Context) (int, error)
	RestartDeployment(ctx context.Context) error
}

// KubectlClient implements KubernetesClient interface using kubectl
type KubectlClient struct {
	config Config
}

// NewKubectlClient creates a new instance of KubectlClient
func NewKubectlClient(config Config) *KubectlClient {
	return &KubectlClient{
		config: config,
	}
}

// GetPodMemoryUsage returns the total memory usage of pods
func (k *KubectlClient) GetPodMemoryUsage(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, k.config.KubectlPath, "top", "pods", "-n", k.config.Namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("error executing kubectl top pods: %v: %s", err, string(output))
	}

	return extractTotalMemory(string(output)), nil
}

// RestartDeployment restarts the specified deployment
func (k *KubectlClient) RestartDeployment(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, k.config.KubectlPath, "rollout", "restart", 
		"deployment/"+k.config.DeploymentName, "-n", k.config.Namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error restarting deployment: %v: %s", err, string(output))
	}
	return nil
}

// Watchdog monitors memory usage and restarts deployments when needed
type Watchdog struct {
	client KubernetesClient
	config Config
}

// NewWatchdog creates a new instance of Watchdog
func NewWatchdog(client KubernetesClient, config Config) *Watchdog {
	return &Watchdog{
		client: client,
		config: config,
	}
}

// Run starts the monitoring
func (w *Watchdog) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.checkAndRestart(ctx); err != nil {
				log.Printf("Error during check: %v", err)
			}
		}
	}
}

// checkAndRestart checks memory usage and restarts if necessary
func (w *Watchdog) checkAndRestart(ctx context.Context) error {
	totalMemory, err := w.client.GetPodMemoryUsage(ctx)
	if err != nil {
		return fmt.Errorf("error getting memory usage: %v", err)
	}

	if w.config.Verbose {
		log.Printf("Total memory usage in namespace '%s': %dMi", w.config.Namespace, totalMemory)
	}

	if totalMemory >= w.config.MemoryThreshold {
		log.Printf("Memory usage exceeded threshold (%dMi). Restarting deployment '%s'...", 
			w.config.MemoryThreshold, w.config.DeploymentName)
		if err := w.client.RestartDeployment(ctx); err != nil {
			return fmt.Errorf("error restarting deployment: %v", err)
		}
		log.Println("Deployment successfully restarted.")
	} else if w.config.Verbose {
		log.Println("Memory usage is within threshold. No action needed.")
	}

	return nil
}

func main() {
	config := parseFlags()
	setupLogging(config.Verbose)

	if config.DeploymentName == "" {
		log.Fatal("Deployment name is required. Use --deployment flag or set DEPLOYMENT environment variable.")
	}

	client := NewKubectlClient(config)
	watchdog := NewWatchdog(client, config)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal. Shutting down...")
		cancel()
	}()

	if err := watchdog.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Error during execution: %v", err)
	}
}

func parseFlags() Config {
	checkInterval := flag.Duration("interval", getEnvDuration("CHECK_INTERVAL", 5*time.Minute), 
		"Check interval")
	namespace := flag.String("namespace", getEnv("NAMESPACE", "default"), "Kubernetes namespace")
	deploymentName := flag.String("deployment", getEnv("DEPLOYMENT", ""), "Deployment name to restart")
	memoryThreshold := flag.Int("threshold", getEnvInt("MEMORY_THRESHOLD", 5000), 
		"Memory threshold in Mi")
	kubectlPath := flag.String("kubectl", getEnv("KUBECTL_PATH", "/usr/local/bin/kubectl"), 
		"Path to kubectl binary")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	return Config{
		Namespace:       *namespace,
		DeploymentName:  *deploymentName,
		MemoryThreshold: *memoryThreshold,
		KubectlPath:     *kubectlPath,
		Verbose:         *verbose,
		CheckInterval:   *checkInterval,
	}
}

func setupLogging(verbose bool) {
	if verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)
	}
}

func extractTotalMemory(output string) int {
	lines := strings.Split(output, "\n")
	totalMemory := 0

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) > 2 {
			memoryStr := strings.ReplaceAll(fields[2], "Mi", "")
			memory, err := strconv.Atoi(memoryStr)
			if err == nil {
				totalMemory += memory
			}
		}
	}

	return totalMemory
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
