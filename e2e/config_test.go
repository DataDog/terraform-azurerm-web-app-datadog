package e2e

import "testing"

func TestLoadConfigUsesDdAuthKeysAndDefaults(t *testing.T) {
	t.Setenv("AZURE_SUBSCRIPTION_ID", "subscription")
	t.Setenv("DD_API_KEY", "api-key")
	t.Setenv("DD_APP_KEY", "app-key")
	t.Setenv("DATADOG_API_KEY", "")
	t.Setenv("DATADOG_APP_KEY", "")
	t.Setenv("AZURE_LOCATION", "")
	t.Setenv("DD_SITE", "")
	t.Setenv("E2E_STORAGE_ACCOUNT", "")
	t.Setenv("E2E_SIDECAR_IMAGE", "")

	cfg := loadConfig(t)

	if cfg.apiKey != "api-key" || cfg.appKey != "app-key" {
		t.Fatal("dd-auth keys were not loaded")
	}
	if cfg.location != defaultLocation || cfg.site != defaultSite {
		t.Fatal("location or site default was not applied")
	}
	if cfg.storageAccount != defaultStorageAccount || cfg.sidecarImage != defaultSidecarImage {
		t.Fatal("artifact default was not applied")
	}
}
