// Package verify checks the deployed instrumentation contract and clean end state.
package verify

import (
	"context"
	"fmt"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/azure"
	e2eshared "github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/shared"
)

const (
	sidecarName       = "datadog-sidecar"
	sidecarTargetPort = "8126"
	moduleTagKey      = "dd_sls_terraform_module"
)

// Expected is the exact instrumentation identity for one run.
type Expected struct {
	Service      string
	Site         string
	Env          string
	Version      string
	Runtime      string
	RunID        string
	CreatedTS    string
	SidecarImage string
	APIKey       string
}

// Instrumented asserts the web app's settings, runtime, tags, and sidecar wiring.
func Instrumented(ctx context.Context, client *azure.Client, resourceGroup, app string, expected Expected) error {
	var violations e2eshared.Violations

	settings, err := client.AppSettings(ctx, resourceGroup, app)
	if err != nil {
		return err
	}
	e2eshared.RequireValues(&violations, "app setting", settings, map[string]string{
		"DD_AAS_INSTANCE_LOGGING_ENABLED":     "false",
		"DD_ENV":                              expected.Env,
		"DD_LOGS_ENABLED":                     "true",
		"DD_LOGS_INJECTION":                   "true",
		"DD_SERVERLESS_LOG_PATH":              "/home/LogFiles/*.log",
		"DD_SERVICE":                          expected.Service,
		"DD_SITE":                             expected.Site,
		"DD_TAGS":                             e2eshared.DefaultRunIDTagKey + ":" + expected.RunID,
		"DD_TRACE_ENABLED":                    "true",
		"DD_VERSION":                          expected.Version,
		"WEBSITES_ENABLE_APP_SERVICE_STORAGE": "true",
	})
	if settings["DD_API_KEY"] == "" {
		violations.Addf("missing app setting DD_API_KEY")
	} else if settings["DD_API_KEY"] != expected.APIKey {
		violations.Addf("app setting DD_API_KEY does not match the configured key")
	}

	containers, err := client.SiteContainers(ctx, resourceGroup, app)
	if err != nil {
		return err
	}
	sidecars := namedContainers(containers, sidecarName)
	if len(sidecars) != 1 {
		violations.Addf("sidecar %q count = %d, want 1", sidecarName, len(sidecars))
	} else {
		sidecar := sidecars[0]
		if sidecar.Properties.Image != expected.SidecarImage {
			violations.Addf("sidecar image = %q, want pinned %q", sidecar.Properties.Image, expected.SidecarImage)
		}
		if sidecar.Properties.IsMain {
			violations.Addf("sidecar is marked as the main container")
		}
		if sidecar.Properties.TargetPort != sidecarTargetPort {
			violations.Addf("sidecar target port = %q, want %q", sidecar.Properties.TargetPort, sidecarTargetPort)
		}

		environment := make(map[string]string, len(sidecar.Properties.EnvironmentVariables))
		for _, variable := range sidecar.Properties.EnvironmentVariables {
			if _, exists := environment[variable.Name]; exists {
				violations.Addf("duplicate sidecar environment variable %s", variable.Name)
			}
			environment[variable.Name] = variable.Value
		}
		for name := range settings {
			if environment[name] != name {
				violations.Addf("sidecar environment variable %s is not wired to app setting %s", name, name)
			}
		}
	}

	webApp, err := client.GetWebApp(ctx, resourceGroup, app)
	if err != nil {
		return err
	}
	if webApp.SiteConfig.LinuxFxVersion != expected.Runtime {
		violations.Addf("runtime = %q, want %q", webApp.SiteConfig.LinuxFxVersion, expected.Runtime)
	}
	e2eshared.RequireValues(&violations, "tag", webApp.Tags, map[string]string{
		"env":                            expected.Env,
		"service":                        expected.Service,
		"version":                        expected.Version,
		e2eshared.DefaultFreshnessTagKey: expected.CreatedTS,
		e2eshared.DefaultRunIDTagKey:     expected.RunID,
	})
	e2eshared.RequirePresent(&violations, "tag", webApp.Tags, moduleTagKey)

	logSettings, err := client.GetLogSettings(ctx, resourceGroup, app)
	if err != nil {
		return err
	}
	if level := logSettings.ApplicationLogs.FileSystem.Level; level != "Information" {
		violations.Addf("application log level = %q, want %q", level, "Information")
	}

	return violations.Err("instrumentation contract violations")
}

// Removed asserts that destroy removed the web app and its instrumentation.
func Removed(ctx context.Context, client *azure.Client, resourceGroup, app string) error {
	exists, err := client.WebAppExists(ctx, resourceGroup, app)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("web app %q still exists after destroy", app)
	}
	return nil
}

func namedContainers(containers []azure.SiteContainer, name string) []azure.SiteContainer {
	matches := make([]azure.SiteContainer, 0, 1)
	for _, container := range containers {
		if container.Name == name {
			matches = append(matches, container)
		}
	}
	return matches
}
