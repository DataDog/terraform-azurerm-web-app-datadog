// Package azure wraps the `az` CLI for the resource inspection and workload
// deployment the e2e suite needs. It mirrors how the datadog-ci reference impl
// shells out to the cloud CLI (gcloud / az) and parses JSON, rather than
// pulling in a heavy SDK. All calls go through the bounded-retry exec helper.
package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/exec"
)

// Client targets a single subscription.
type Client struct {
	SubscriptionID string
}

func New(subscriptionID string) *Client {
	return &Client{SubscriptionID: subscriptionID}
}

// AppSettings returns the web app's application settings as a name->value map.
func (c *Client) AppSettings(ctx context.Context, rg, app string) (map[string]string, error) {
	res, err := exec.RunWithRetries(ctx, exec.Options{}, "az", "webapp", "config", "appsettings", "list",
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
		Image      string `json:"image"`
		IsMain     bool   `json:"isMain"`
		TargetPort string `json:"targetPort"`
	} `json:"properties"`
}

// SiteContainers lists the sidecar containers attached to the web app via the
// ARM REST API (the azapi resources the module creates are not visible to the
// `az webapp` surface).
func (c *Client) SiteContainers(ctx context.Context, rg, app string) ([]SiteContainer, error) {
	url := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s/sitecontainers?api-version=2024-11-01",
		c.SubscriptionID, rg, app)
	res, err := exec.RunWithRetries(ctx, exec.Options{}, "az", "rest", "--method", "get", "--url", url, "--output", "json")
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

// Tags returns the web app's resource tags.
func (c *Client) Tags(ctx context.Context, rg, app string) (map[string]string, error) {
	res, err := exec.RunWithRetries(ctx, exec.Options{}, "az", "webapp", "show",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app, "--output", "json")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Tags map[string]string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &raw); err != nil {
		return nil, fmt.Errorf("parse webapp show: %w", err)
	}
	if raw.Tags == nil {
		raw.Tags = map[string]string{}
	}
	return raw.Tags, nil
}

// WebAppExists reports whether the web app still exists. Used to assert the
// clean end-state after REMOVE.
func (c *Client) WebAppExists(ctx context.Context, rg, app string) (bool, error) {
	res := exec.Run(ctx, nil, "az", "webapp", "show",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app, "--output", "json")
	if res.ExitCode == 0 {
		return true, nil
	}
	// ResourceNotFound is the expected post-destroy outcome.
	if strings.Contains(res.Stderr, "ResourceNotFound") || strings.Contains(res.Stderr, "could not be found") ||
		strings.Contains(res.Stderr, "was not found") {
		return false, nil
	}
	return false, fmt.Errorf("webapp show failed unexpectedly: %s", res.Stderr)
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

	if _, err := exec.RunWithRetries(ctx, exec.Options{}, "az", "storage", "blob", "download",
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
	// A source-only deploy + SCM build completes in a few minutes; 2 attempts is
	// enough and avoids long hangs if the worker can't start (each az deploy can
	// itself wait ~10 min for worker startup before erroring).
	if _, err := exec.RunWithRetries(ctx, exec.Options{MaxAttempts: 2, DelaySeconds: 15}, "az", "webapp", "deploy",
		"--subscription", c.SubscriptionID, "--resource-group", rg, "--name", app,
		"--src-path", zipPath, "--type", "zip", "--async", "false", "--output", "none"); err != nil {
		return fmt.Errorf("zip-deploy workload: %w", err)
	}
	return nil
}
