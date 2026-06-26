// Package verify implements the config side of the conformance contract for the
// Linux Web App module: it asserts the instrumentation the module is supposed to
// apply is present with the expected *values* (identity, not existence), and
// that after REMOVE no residue remains.
//
// It mirrors the structure of datadog-ci's aas-verifier.ts (verifyInstrumented
// / verifyUninstrumented) but is reimplemented in Go against this module's
// mapping: a `datadog-sidecar` sitecontainer + DD_* app settings + DD tags.
package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/azure"
)

const (
	sidecarName       = "datadog-sidecar"
	sidecarTargetPort = "8126"
	moduleTagKey      = "dd_sls_terraform_module"
)

// Expected is the identity the module should have applied.
type Expected struct {
	Service      string
	Site         string
	Env          string
	Version      string
	SidecarImage string // pinned image; the running sidecar must match exactly
}

// Instrumented asserts the module produced the expected config on the web app:
// required DD_* app settings with matching values, the datadog-sidecar
// sitecontainer pinned to the expected image on port 8126, and DD resource tags
// with matching values.
func Instrumented(ctx context.Context, c *azure.Client, rg, app string, exp Expected) error {
	settings, err := c.AppSettings(ctx, rg, app)
	if err != nil {
		return err
	}
	// API key wiring: present and non-empty (value is sensitive, not asserted).
	if settings["DD_API_KEY"] == "" {
		return fmt.Errorf("DD_API_KEY missing or empty")
	}
	// Identity assertions on the remaining required settings.
	for key, want := range map[string]string{
		"DD_SITE":                             exp.Site,
		"DD_SERVICE":                          exp.Service,
		"DD_ENV":                              exp.Env,
		"DD_VERSION":                          exp.Version,
		"WEBSITES_ENABLE_APP_SERVICE_STORAGE": "true",
	} {
		got, ok := settings[key]
		if !ok {
			return fmt.Errorf("app setting %s not set", key)
		}
		if got != want {
			return fmt.Errorf("app setting %s = %q, want %q", key, got, want)
		}
	}

	// Sidecar sitecontainer present, pinned, on the agent port.
	containers, err := c.SiteContainers(ctx, rg, app)
	if err != nil {
		return err
	}
	sidecar := findContainer(containers, sidecarName)
	if sidecar == nil {
		return fmt.Errorf("sidecar %q not found among %d sitecontainers", sidecarName, len(containers))
	}
	if !strings.Contains(sidecar.Properties.Image, "serverless-init") {
		return fmt.Errorf("sidecar image %q is not a serverless-init image", sidecar.Properties.Image)
	}
	if exp.SidecarImage != "" && sidecar.Properties.Image != exp.SidecarImage {
		return fmt.Errorf("sidecar image %q != pinned %q", sidecar.Properties.Image, exp.SidecarImage)
	}
	if sidecar.Properties.TargetPort != sidecarTargetPort {
		return fmt.Errorf("sidecar targetPort %q != %q", sidecar.Properties.TargetPort, sidecarTargetPort)
	}

	// DD resource tags with matching values.
	tags, err := c.Tags(ctx, rg, app)
	if err != nil {
		return err
	}
	for key, want := range map[string]string{
		"service": exp.Service,
		"env":     exp.Env,
		"version": exp.Version,
	} {
		got, ok := tags[key]
		if !ok {
			return fmt.Errorf("tag %s not set", key)
		}
		if got != want {
			return fmt.Errorf("tag %s = %q, want %q", key, got, want)
		}
	}
	if _, ok := tags[moduleTagKey]; !ok {
		return fmt.Errorf("module marker tag %s not set", moduleTagKey)
	}
	return nil
}

// Removed asserts the clean end-state after REMOVE: the wrapper owns the web
// app, so `terraform destroy` removes it entirely. We assert the web app (and
// thus its sidecar, DD_* settings and tags) no longer exists -- explicit
// absence, the wrapper-module form of "uninstrumented".
func Removed(ctx context.Context, c *azure.Client, rg, app string) error {
	exists, err := c.WebAppExists(ctx, rg, app)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("web app %q still exists after destroy; instrumentation residue remains", app)
	}
	return nil
}

func findContainer(containers []azure.SiteContainer, name string) *azure.SiteContainer {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}
