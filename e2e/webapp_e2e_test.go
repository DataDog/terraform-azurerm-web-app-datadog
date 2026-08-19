package e2e

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/azure"
	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/telemetry"
	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/verify"
	e2eshared "github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/shared"
)

var sharedCfg = e2eshared.Config{
	Tool:     "tfwebapp",
	Platform: "linux",
}

const (
	fixtureDir   = "fixtures/linux-node"
	workloadBlob = "node-sidecar.zip"
	ddEnv        = "e2e"
	ddVersion    = "1.0.0"
	nodeRuntime  = "NODE|22-lts"
)

func TestLinuxNodeWebAppE2E(t *testing.T) {
	cfg := loadConfig(t)
	ctx := context.Background()
	az := azure.New(cfg.subscriptionID)

	runPhase(t, "configuration preflight", func() {
		require.NoError(t, az.Validate(ctx), "validate Azure credentials and subscription")
	})

	runID := e2eshared.NewRunID()
	created := time.Now().Unix()
	createdTS := strconv.FormatInt(created, 10)
	appName := e2eshared.ResourceName(sharedCfg, runID)
	rgName := appName + "-rg"
	t.Logf("run id %s -> web app %s", runID, appName)

	tfOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: fixtureDir,
		Vars: map[string]interface{}{
			"subscription_id":     cfg.subscriptionID,
			"datadog_site":        cfg.site,
			"datadog_service":     appName,
			"datadog_env":         ddEnv,
			"datadog_version":     ddVersion,
			"sidecar_image":       cfg.sidecarImage,
			"name":                appName,
			"resource_group_name": rgName,
			"location":            cfg.location,
			"tags":                e2eshared.Tags(sharedCfg, runID, created),
			"datadog_tags":        e2eshared.DefaultRunIDTagKey + ":" + runID,
		},
		EnvVars: map[string]string{
			"ARM_SUBSCRIPTION_ID":    cfg.subscriptionID,
			"TF_VAR_datadog_api_key": cfg.apiKey,
		},
		NoColor: true,
	})
	for pattern, message := range map[string]string{
		"another operation is in progress":    "App Service operation in progress; retrying.",
		"Cannot modify this web hosting plan": "App Service plan busy; retrying.",
		"AnotherOperationInProgress":          "Concurrent operation; retrying.",
		"(?s)429.*(TooManyRequests|Throttl)":  "ARM throttling; retrying.",
		"(?s)unexpected status 409":           "ARM 409 conflict; retrying.",
		"RetryableError":                      "Retryable cloud error; retrying.",
	} {
		tfOpts.RetryableTerraformErrors[pattern] = message
	}
	tfOpts.MaxRetries = 6
	tfOpts.TimeBetweenRetries = 30 * time.Second

	cleaned := false
	cleanup := func() {
		runPhase(t, "destroy", func() { terraform.Destroy(t, tfOpts) })
		runPhase(t, "cleanup verification", func() {
			require.NoError(t, verify.Removed(ctx, az, rgName, appName), "destroy must leave no web app")
		})
	}
	defer func() {
		if !cleaned {
			cleaned = true
			cleanup()
		}
	}()

	runPhase(t, "instrumentation deploy", func() { terraform.InitAndApply(t, tfOpts) })

	expected := verify.Expected{
		Service:      appName,
		Site:         cfg.site,
		Env:          ddEnv,
		Version:      ddVersion,
		Runtime:      nodeRuntime,
		RunID:        runID,
		CreatedTS:    createdTS,
		SidecarImage: cfg.sidecarImage,
		APIKey:       cfg.apiKey,
	}
	runPhase(t, "config verification", func() {
		require.NoError(t, verify.Instrumented(ctx, az, rgName, appName, expected))
	})

	runPhase(t, "workload bundle deploy", func() {
		if localZip := os.Getenv("E2E_WORKLOAD_ZIP"); localZip != "" {
			require.NoError(t, az.DeployLocalZip(ctx, rgName, appName, localZip))
			return
		}
		require.NoError(t, az.DeployPrebuiltPackage(ctx, rgName, appName, cfg.storageAccount, workloadBlob))
	})

	hostname := terraform.Output(t, tfOpts, "default_hostname")
	runPhase(t, "invoke", func() { triggerWorkload(t, hostname) })

	trafficCtx, stopTraffic := context.WithCancel(ctx)
	defer stopTraffic()
	go e2eshared.GenerateTraffic(trafficCtx, "https://"+hostname, 5*time.Second)

	telemetryCtx, cancelTelemetry := context.WithTimeout(ctx, 12*time.Minute)
	defer cancelTelemetry()
	runPhase(t, "telemetry wait", func() {
		require.NoError(t, telemetry.CheckTelemetryFlowing(telemetryCtx,
			telemetry.Config{APIKey: cfg.apiKey, AppKey: cfg.appKey, Site: cfg.site},
			telemetry.Expected{Service: appName, Env: ddEnv},
			telemetry.Options{ExpectLogs: true}))
	})
	stopTraffic()

	runPhase(t, "idempotency check", func() {
		require.Equal(t, 0, terraform.PlanExitCode(t, tfOpts), "re-apply must be a no-op")
	})

	cleaned = true
	cleanup()
}

func runPhase(t *testing.T, name string, run func()) {
	t.Helper()
	done := logProgress(t, name)
	defer done()
	run()
}

func logProgress(t *testing.T, phase string) func() {
	t.Helper()
	started := time.Now()
	t.Logf("START: %s", phase)

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Logf("RUNNING: %s (%s elapsed)", phase, time.Since(started).Round(time.Second))
			case <-stop:
				return
			}
		}
	}()

	return func() {
		close(stop)
		<-done
		t.Logf("DONE: %s (%s)", phase, time.Since(started).Round(time.Second))
	}
}

func triggerWorkload(t *testing.T, hostname string) {
	t.Helper()
	url := "https://" + hostname + "/"
	client := &http.Client{Timeout: 30 * time.Second}

	const maxAttempts = 40
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				t.Logf("invoke attempt %d/%d: HTTP %d", attempt, maxAttempts, resp.StatusCode)
				for i := 0; i < 3; i++ {
					if response, requestErr := client.Get(url); requestErr == nil {
						response.Body.Close()
					}
					time.Sleep(2 * time.Second)
				}
				return
			}
		}
		if err != nil {
			t.Logf("invoke attempt %d/%d: %v", attempt, maxAttempts, err)
		} else {
			t.Logf("invoke attempt %d/%d: HTTP %d", attempt, maxAttempts, resp.StatusCode)
		}
		time.Sleep(15 * time.Second)
	}
	t.Fatalf("workload at %s did not return 2xx", url)
}
