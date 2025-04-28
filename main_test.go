package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"
)

func TestExtractTotalMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name: "valid input",
			input: `NAME                     CPU(cores)   MEMORY(bytes)
pod-1                    100m         1000Mi
pod-2                    200m         2000Mi`,
			expected: 3000,
		},
		{
			name: "empty input",
			input: `NAME                     CPU(cores)   MEMORY(bytes)`,
			expected: 0,
		},
		{
			name: "invalid memory format",
			input: `NAME                     CPU(cores)   MEMORY(bytes)
pod-1                    100m         invalid`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTotalMemory(tt.input)
			if result != tt.expected {
				t.Errorf("extractTotalMemory() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		fallback time.Duration
		expected time.Duration
	}{
		{
			name:     "valid duration",
			envKey:   "TEST_DURATION",
			envValue: "1m",
			fallback: 5 * time.Minute,
			expected: 1 * time.Minute,
		},
		{
			name:     "invalid duration",
			envKey:   "TEST_DURATION",
			envValue: "invalid",
			fallback: 5 * time.Minute,
			expected: 5 * time.Minute,
		},
		{
			name:     "missing env var",
			envKey:   "NONEXISTENT",
			envValue: "",
			fallback: 5 * time.Minute,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnvDuration(tt.envKey, tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnvDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	// Reset flags before each test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test default values
	config := parseFlags()
	if config.Namespace != "default" {
		t.Errorf("Expected default namespace, got %v", config.Namespace)
	}
	if config.MemoryThreshold != 5000 {
		t.Errorf("Expected default memory threshold 5000, got %v", config.MemoryThreshold)
	}
	if config.CheckInterval != 5*time.Minute {
		t.Errorf("Expected default check interval 5m, got %v", config.CheckInterval)
	}
}

// MockKubernetesClient implements KubernetesClient interface for testing
type MockKubernetesClient struct {
	memoryUsage int
	restartErr  error
}

func (m *MockKubernetesClient) GetPodMemoryUsage(ctx context.Context) (int, error) {
	return m.memoryUsage, nil
}

func (m *MockKubernetesClient) RestartDeployment(ctx context.Context) error {
	return m.restartErr
}

func TestWatchdogRun(t *testing.T) {
	tests := []struct {
		name           string
		memoryUsage    int
		threshold      int
		restartErr     error
		shouldRestart  bool
		checkInterval  time.Duration
		timeout        time.Duration
	}{
		{
			name:          "memory below threshold",
			memoryUsage:   1000,
			threshold:     2000,
			shouldRestart: false,
			checkInterval: 100 * time.Millisecond,
			timeout:       200 * time.Millisecond,
		},
		{
			name:          "memory above threshold",
			memoryUsage:   3000,
			threshold:     2000,
			shouldRestart: true,
			checkInterval: 100 * time.Millisecond,
			timeout:       200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKubernetesClient{
				memoryUsage: tt.memoryUsage,
				restartErr:  tt.restartErr,
			}

			config := Config{
				MemoryThreshold: tt.threshold,
				CheckInterval:   tt.checkInterval,
				Verbose:         true,
			}

			watchdog := NewWatchdog(mockClient, config)
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := watchdog.Run(ctx)
			if err != nil && err != context.DeadlineExceeded {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
} 