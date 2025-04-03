package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	namespace       = flag.String("namespace", getEnv("NAMESPACE", "default"), "Kubernetes namespace")
	deploymentName  = flag.String("deployment", getEnv("DEPLOYMENT", ""), "Deployment name to restart")
	memoryThreshold = flag.Int("threshold", getEnvInt("MEMORY_THRESHOLD", 5000), "Memory threshold in Mi")
	kubectlPath     = flag.String("kubectl", getEnv("KUBECTL_PATH", "/usr/local/bin/kubectl"), "Path to kubectl binary")
	verbose         = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	if *deploymentName == "" {
		log.Fatal("Deployment name is required. Use --deployment flag or set DEPLOYMENT environment variable.")
	}

	if *verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)
	}

	output, err := exec.Command(*kubectlPath, "top", "pods", "-n", *namespace).CombinedOutput()
	if err != nil {
		log.Printf("Error executing kubectl top pods: %v\n", err)
		log.Println(string(output))
		return
	}

	if *verbose {
		log.Println("Output from kubectl top pods:")
		log.Println(string(output))
	}

	totalMemory := extractTotalMemory(string(output))
	log.Printf("Total memory used in namespace '%s': %dMi\n", *namespace, totalMemory)

	if totalMemory >= *memoryThreshold {
		log.Printf("Memory usage exceeded threshold (%dMi). Restarting deployment '%s'...\n", *memoryThreshold, *deploymentName)
		err := restartDeployment(*namespace, *deploymentName, *kubectlPath)
		if err != nil {
			log.Printf("Error restarting deployment: %v\n", err)
		}
	} else if *verbose {
		log.Println("Memory usage is within the threshold. No action needed.")
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

func restartDeployment(namespace, deploymentName, kubectlPath string) error {
	cmd := exec.Command(kubectlPath, "rollout", "restart", "deployment/"+deploymentName, "-n", namespace)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing kubectl rollout restart: %v", err)
	}
	log.Println("Deployment successfully restarted.")
	return nil
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
