package e2e

import (
	"os"
	"testing"
)

const (
	defaultLocation       = "eastus2"
	defaultSite           = "datadoghq.com"
	defaultStorageAccount = "smddsvlsprod"
	defaultSidecarImage   = "index.docker.io/datadog/serverless-init@sha256:6fb7637628fdf31d536bc9c49fbe6304371df5e2ecdb15c1c2d5e2d66395c3a0"
)

type config struct {
	subscriptionID string
	location       string
	site           string
	apiKey         string
	appKey         string
	storageAccount string
	sidecarImage   string
	runID          string
}

func loadConfig(t *testing.T) config {
	t.Helper()

	cfg := config{
		subscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
		location:       getEnv("AZURE_LOCATION", defaultLocation),
		site:           getEnv("DD_SITE", defaultSite),
		apiKey:         os.Getenv("DD_API_KEY"),
		appKey:         os.Getenv("DD_APP_KEY"),
		storageAccount: getEnv("E2E_STORAGE_ACCOUNT", defaultStorageAccount),
		sidecarImage:   getEnv("E2E_SIDECAR_IMAGE", defaultSidecarImage),
		runID:          os.Getenv("E2E_RUN_ID"),
	}

	var missing []string
	for _, required := range []struct{ name, value string }{
		{"AZURE_SUBSCRIPTION_ID", cfg.subscriptionID},
		{"DD_API_KEY", cfg.apiKey},
		{"DD_APP_KEY", cfg.appKey},
	} {
		if required.value == "" {
			missing = append(missing, required.name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("missing required e2e configuration: %v", missing)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
