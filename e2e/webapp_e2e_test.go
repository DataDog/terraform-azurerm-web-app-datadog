// Package e2e is the end-to-end conformance suite for the terraform-azurerm-web-app
// Datadog module, conforming to serverless-ci/e2e/spec.md.
//
// It drives the full lifecycle against a real Azure subscription:
//
//	APPLY (terraform apply)        -> verify CONFIG present
//	provision workload + trigger   -> verify TELEMETRY (traces + logs) flows
//	re-APPLY (terraform plan)       -> assert idempotent (empty plan)
//	REMOVE (terraform destroy)      -> verify CLEAN end-state (web app gone)
//
// The module is a wrapper: `terraform apply` is the instrumentation mechanism
// and `terraform destroy` is removal. Teardown runs always, even on failure.
//
// Run locally with Azure CLI login + Datadog API/APP keys; see README.md.
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/azure"
	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/telemetry"
	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/verify"
	e2eshared "github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/shared"
)

// sharedCfg parameterizes the shared naming/tag helpers for this module: the
// tool/platform identity that forms the sweeper name prefix. The run-id tag uses
// the shared default (one_e2e_run_id), matching the other e2e suites and the
// cross-repo serverless-e2e index.
var sharedCfg = e2eshared.Config{
	Tool:     "tfwebapp",
	Platform: "linux",
}

const (
	fixtureDir = "fixtures/linux-node"
	location   = "eastus2"
	// workloadBlob is the prebuilt prod Node.js sidecar-flavor package published
	// by serverless-init-self-monitoring to the storage account's `code`
	// container. We pull it, never rebuild from source (per spec).
	workloadBlob = "node-sidecar.zip"
	// defaultStorageAccount is the prod self-monitoring storage account.
	defaultStorageAccount = "smddsvlsprod"
	// defaultSidecarImage pins serverless-init so a telemetry failure blames the
	// module wiring, not an upstream agent regression. Override via E2E_SIDECAR_IMAGE.
	defaultSidecarImage = "index.docker.io/datadog/serverless-init:latest"

	ddEnv     = "e2e"
	ddVersion = "1.0.0"
)

func TestLinuxNodeWebAppE2E(t *testing.T) {
	t.Parallel()

	if os.Getenv("SKIP_AAS_TESTS") == "true" {
		t.Skip("SKIP_AAS_TESTS=true")
	}

	subscriptionID := mustEnv(t, "AZURE_SUBSCRIPTION_ID")
	ddAPIKey := mustEnv(t, "DD_API_KEY")       // wired into the workload by the module
	telAPIKey := mustEnv(t, "DATADOG_API_KEY") // queries the Datadog API
	telAppKey := mustEnv(t, "DATADOG_APP_KEY")
	ddSite := envOr("DD_SITE", "datadoghq.com")
	storageAccount := envOr("E2E_STORAGE_ACCOUNT", defaultStorageAccount)
	sidecarImage := envOr("E2E_SIDECAR_IMAGE", defaultSidecarImage)

	runID := e2eshared.NewRunID()
	created := time.Now().Unix()
	// one-e2e-tfwebapp-linux-<runid>: globally-unique, DNS-valid, well under the
	// 60-char Web App budget; the prefix is the sweeper's blast-radius guard.
	appName := e2eshared.ResourceName(sharedCfg, runID)
	rgName := appName + "-rg" // run-scoped group shares the sweeper prefix
	service := appName        // run-unique service => telemetry is filterable by run id

	ctx := context.Background()
	az := azure.New(subscriptionID)

	tfOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: fixtureDir,
		Vars: map[string]interface{}{
			"subscription_id":     subscriptionID,
			"datadog_api_key":     ddAPIKey,
			"datadog_site":        ddSite,
			"datadog_service":     service,
			"datadog_env":         ddEnv,
			"datadog_version":     ddVersion,
			"sidecar_image":       sidecarImage,
			"name":                appName,
			"resource_group_name": rgName,
			"location":            location,
			"tags":                e2eshared.Tags(sharedCfg, runID, created),
			// DD_TAGS stamps the standard run-id marker onto ingested telemetry so
			// the shared serverless-e2e index (filtered on one_e2e_run_id:*) captures
			// this run's spans and logs; the run-unique service makes telemetry
			// filterable per run on top of that.
			"datadog_tags": e2eshared.DefaultRunIDTagKey + ":" + runID,
		},
		EnvVars: map[string]string{"ARM_SUBSCRIPTION_ID": subscriptionID},
		NoColor: true,
	})
	// Retry the cloud, not the assertions: App Service control-plane returns
	// transient 409s ("another operation is in progress") and 429 throttling
	// under load. Add them to the default retryable set so apply/destroy ride
	// through instead of failing the run.
	for re, msg := range map[string]string{
		"another operation is in progress":    "App Service operation in progress; retrying.",
		"Cannot modify this web hosting plan": "App Service plan busy; retrying.",
		"AnotherOperationInProgress":          "Concurrent operation; retrying.",
		"(?s)429.*(TooManyRequests|Throttl)":  "ARM throttling; retrying.",
		"(?s)unexpected status 409":           "ARM 409 conflict; retrying.",
		"RetryableError":                      "Retryable cloud error; retrying.",
	} {
		tfOpts.RetryableTerraformErrors[re] = msg
	}
	tfOpts.MaxRetries = 6
	tfOpts.TimeBetweenRetries = 30 * time.Second

	// REMOVE + verify CLEAN end-state. Teardown always runs, even on failure.
	defer func() {
		terraform.Destroy(t, tfOpts)
		require.NoError(t, verify.Removed(ctx, az, rgName, appName),
			"REMOVE must leave no residue")
	}()

	// APPLY: the wrapper module creates the web app already instrumented.
	terraform.InitAndApply(t, tfOpts)

	// verify CONFIG present (identity, not existence).
	require.NoError(t, verify.Instrumented(ctx, az, rgName, appName, verify.Expected{
		Service:      service,
		Site:         ddSite,
		Env:          ddEnv,
		Version:      ddVersion,
		SidecarImage: sidecarImage,
	}))

	// Provision the workload app and trigger it. CI pulls the prebuilt prod
	// package from the artifact storage account; a developer without storage
	// RBAC can set E2E_WORKLOAD_ZIP to a local package reconstructed from the
	// self-monitoring source (see README).
	if localZip := os.Getenv("E2E_WORKLOAD_ZIP"); localZip != "" {
		require.NoError(t, az.DeployLocalZip(ctx, rgName, appName, localZip))
	} else {
		require.NoError(t, az.DeployPrebuiltPackage(ctx, rgName, appName, storageAccount, workloadBlob))
	}
	hostname := terraform.Output(t, tfOpts, "default_hostname")
	triggerWorkload(t, hostname)

	// Drive continuous traffic for the duration of the telemetry poll. The
	// serverless-init sidecar tails Azure's per-instance stdout log file from the
	// END, so lines written before its tailer attached sit behind the offset.
	// Without fresh requests during the poll no new lines are forwarded and logs
	// never arrive (spans are unaffected -- the tracer ships over HTTP). Stopped
	// once the poll returns.
	tctx, stopTraffic := context.WithCancel(ctx)
	defer stopTraffic()
	go e2eshared.GenerateTraffic(tctx, "https://"+hostname, 5*time.Second)

	// verify TELEMETRY: traces and logs filtered by the run-unique service + env
	// identity. Logs are now collected end-to-end (the workload's stdout is tailed
	// via DD_AAS_INSTANCE_LOGGING_ENABLED; see the fixture), so they are required
	// alongside traces.
	require.NoError(t, telemetry.CheckTelemetryFlowing(ctx,
		telemetry.Config{APIKey: telAPIKey, AppKey: telAppKey, Site: ddSite},
		telemetry.Expected{Service: service, Env: ddEnv},
		telemetry.Options{ExpectLogs: true}))

	// re-APPLY idempotent: a fresh plan must report no changes.
	exitCode := terraform.PlanExitCode(t, tfOpts)
	require.Equal(t, 0, exitCode, "re-apply must be a no-op (terraform plan reported changes)")
}

// triggerWorkload hits the workload's HTTP endpoint until it warms up and
// returns success, generating the trace + log the telemetry check expects. It
// retries the cloud (cold start / DNS propagation), not the assertion.
func triggerWorkload(t *testing.T, hostname string) {
	t.Helper()
	url := fmt.Sprintf("https://%s/", hostname)
	client := &http.Client{Timeout: 30 * time.Second}

	const maxAttempts = 40 // ~10 min; the deploy no longer gates on worker start
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				t.Logf("workload responded %d on attempt %d", resp.StatusCode, attempt)
				// A few more hits to produce a healthy span/log sample.
				for i := 0; i < 3; i++ {
					if r, e := client.Get(url); e == nil {
						r.Body.Close()
					}
					time.Sleep(2 * time.Second)
				}
				return
			}
		}
		t.Logf("workload not ready (attempt %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(15 * time.Second)
	}
	t.Fatalf("workload at %s never became reachable", url)
}

func mustEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("required env var %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
