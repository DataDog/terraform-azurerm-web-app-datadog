// Package azure wraps the `az` CLI calls that inspect resources and deploy the
// workload. All calls go through the shared bounded-retry exec helper.
package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	e2eshared "github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/shared"
)

// sharedCfg runs the `az` CLI and retries only transient Azure control-plane failures.
var sharedCfg = e2eshared.Config{
	Command: "az",
	RetryPatterns: []string{
		"AnotherOperationInProgress",
		"Conflict",
		"ECONNRESET",
		"ETIMEDOUT",
		"GatewayTimeout",
		"InternalServerError",
		"Read timed out",
		"ServiceUnavailable",
		"TooManyRequests",
		"temporarily unavailable",
	},
}

const (
	readAttempts = 5
	readDelay    = 15 * time.Second
)

// Client targets a single subscription.
type Client struct {
	SubscriptionID string
}

func New(subscriptionID string) *Client {
	return &Client{SubscriptionID: subscriptionID}
}

// Validate checks that Azure CLI credentials can access the configured subscription.
func (c *Client) Validate(ctx context.Context) error {
	_, err := e2eshared.Run(ctx, sharedCfg, "account", "show", "--subscription", c.SubscriptionID, "--output", "none")
	return err
}

// AppSettings returns the web app's application settings as a name->value map.
func (c *Client) AppSettings(ctx context.Context, rg, app string) (map[string]string, error) {
	res, err := e2eshared.RunWithRetries(ctx, sharedCfg, readAttempts, readDelay, "webapp", "config", "appsettings", "list",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app, "--output", "json")
	if err != nil {
		return nil, err
	}
	var raw []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &raw); err != nil {
		return nil, fmt.Errorf("parse appsettings: %w", err)
	}
	out := make(map[string]string, len(raw))
	for _, s := range raw {
		out[s.Name] = s.Value
	}
	return out, nil
}

// SiteContainer is one entry from the sitecontainers ARM collection.
type SiteContainer struct {
	Name       string `json:"name"`
	Properties struct {
		Image                string `json:"image"`
		IsMain               bool   `json:"isMain"`
		TargetPort           string `json:"targetPort"`
		EnvironmentVariables []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"environmentVariables"`
	} `json:"properties"`
}

// SiteContainers lists the sidecar containers attached to the web app via the
// ARM REST API (the azapi resources the module creates are not visible to the
// `az webapp` surface).
func (c *Client) SiteContainers(ctx context.Context, rg, app string) ([]SiteContainer, error) {
	url := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s/sitecontainers?api-version=2024-11-01",
		c.SubscriptionID, rg, app)
	res, err := e2eshared.RunWithRetries(ctx, sharedCfg, readAttempts, readDelay, "rest", "--method", "get", "--url", url, "--output", "json")
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Value []SiteContainer `json:"value"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &wrap); err != nil {
		return nil, fmt.Errorf("parse sitecontainers: %w", err)
	}
	return wrap.Value, nil
}

// WebApp contains the deployed properties used by the verifier.
type WebApp struct {
	Tags       map[string]string `json:"tags"`
	SiteConfig struct {
		LinuxFxVersion string `json:"linuxFxVersion"`
	} `json:"siteConfig"`
}

// GetWebApp returns the web app's resource tags and runtime configuration.
func (c *Client) GetWebApp(ctx context.Context, rg, app string) (WebApp, error) {
	res, err := e2eshared.RunWithRetries(ctx, sharedCfg, readAttempts, readDelay, "webapp", "show",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app, "--output", "json")
	if err != nil {
		return WebApp{}, err
	}
	var webApp WebApp
	if err := json.Unmarshal([]byte(res.Stdout), &webApp); err != nil {
		return WebApp{}, fmt.Errorf("parse webapp show: %w", err)
	}
	if webApp.Tags == nil {
		webApp.Tags = map[string]string{}
	}
	return webApp, nil
}

// LogSettings contains the platform file-logging configuration.
type LogSettings struct {
	ApplicationLogs struct {
		FileSystem struct {
			Level string `json:"level"`
		} `json:"fileSystem"`
	} `json:"applicationLogs"`
}

// GetLogSettings returns the web app's logging configuration.
func (c *Client) GetLogSettings(ctx context.Context, rg, app string) (LogSettings, error) {
	res, err := e2eshared.RunWithRetries(ctx, sharedCfg, readAttempts, readDelay, "webapp", "log", "show",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app, "--output", "json")
	if err != nil {
		return LogSettings{}, err
	}
	var settings LogSettings
	if err := json.Unmarshal([]byte(res.Stdout), &settings); err != nil {
		return LogSettings{}, fmt.Errorf("parse webapp log settings: %w", err)
	}
	return settings, nil
}

// ResourceGroupExists reports whether the resource group still exists.
func (c *Client) ResourceGroupExists(ctx context.Context, rg string) (bool, error) {
	res, err := e2eshared.RunWithRetries(ctx, sharedCfg, readAttempts, readDelay, "group", "exists",
		"--subscription", c.SubscriptionID, "--name", rg)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(res.Stdout)
}

// DeleteResourceGroup starts deleting the resource group and everything in it.
func (c *Client) DeleteResourceGroup(ctx context.Context, rg string) error {
	_, err := e2eshared.RunWithRetries(ctx, sharedCfg, 3, 5*time.Second, "group", "delete",
		"--subscription", c.SubscriptionID, "--name", rg, "--yes", "--no-wait")
	return err
}

// DeployPrebuiltPackage downloads the prebuilt prod workload package from the
// self-monitoring storage account's `code` container and zip-deploys it. The
// package is pulled, never rebuilt from source (per spec).
func (c *Client) DeployPrebuiltPackage(ctx context.Context, rg, app, storageAccount, blobName string) error {
	dir, err := os.MkdirTemp("", "e2e-workload-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	zipPath := filepath.Join(dir, blobName)

	if _, err := e2eshared.RunWithRetries(ctx, sharedCfg, 3, 5*time.Second, "storage", "blob", "download",
		"--account-name", storageAccount, "--container-name", "code", "--name", blobName,
		"--file", zipPath, "--auth-mode", "login", "--no-progress", "--output", "none"); err != nil {
		return fmt.Errorf("download prebuilt package %s/%s: %w", storageAccount, blobName, err)
	}

	// The published package bundles node_modules, so it runs as-is on the Linux
	// Web App without an SCM/Oryx build (which conflicts with bundled deps).
	return c.DeployLocalZip(ctx, rg, app, zipPath)
}

// DeployLocalZip zip-deploys an already-local workload package. CI uses
// DeployPrebuiltPackage (pulls the published artifact); this is the local
// escape hatch for developers without storage RBAC on the artifact account,
// who can point E2E_WORKLOAD_ZIP at a package reconstructed from the
// self-monitoring source.
func (c *Client) DeployLocalZip(ctx context.Context, rg, app, zipPath string) error {
	// --track-status false pushes the package without gating on the runtime
	// worker-startup poll: under load that poll falsely times out even when the
	// worker comes up fine. The HTTP trigger that follows is the real
	// worker-started signal, so we don't need az to confirm it here.
	if _, err := e2eshared.RunWithRetries(ctx, sharedCfg, 2, 15*time.Second, "webapp", "deploy",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app,
		"--src-path", zipPath, "--type", "zip", "--track-status", "false", "--async", "false", "--output", "none"); err != nil {
		return fmt.Errorf("zip-deploy workload: %w", err)
	}
	return nil
}
