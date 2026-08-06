package e2e

import (
	"os"
	"sort"
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
}

func loadConfig(t *testing.T) config {
	t.Helper()

	cfg := config{
		subscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
		location:       firstNonEmpty(os.Getenv("AZURE_LOCATION"), defaultLocation),
		site:           firstNonEmpty(os.Getenv("DD_SITE"), defaultSite),
		apiKey:         firstNonEmpty(os.Getenv("DATADOG_API_KEY"), os.Getenv("DD_API_KEY")),
		appKey:         firstNonEmpty(os.Getenv("DATADOG_APP_KEY"), os.Getenv("DD_APP_KEY")),
		storageAccount: firstNonEmpty(os.Getenv("E2E_STORAGE_ACCOUNT"), defaultStorageAccount),
		sidecarImage:   firstNonEmpty(os.Getenv("E2E_SIDECAR_IMAGE"), defaultSidecarImage),
	}

	missing := make([]string, 0, 3)
	for name, value := range map[string]string{
		"AZURE_SUBSCRIPTION_ID":      cfg.subscriptionID,
		"DATADOG_API_KEY/DD_API_KEY": cfg.apiKey,
		"DATADOG_APP_KEY/DD_APP_KEY": cfg.appKey,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("missing required e2e configuration: %v", missing)
	}

	return cfg
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
