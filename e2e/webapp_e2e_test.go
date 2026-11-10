package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/azure"
	"github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/internal/verify"
	e2eshared "github.com/DataDog/terraform-azurerm-web-app-datadog/e2e/shared"
)

var sharedCfg = e2eshared.Config{
	Tool:     "tf",
	Platform: "aas-linux",
}

const (
	fixtureDir   = "e2e/fixtures/linux-node"
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

	runID := cfg.runID
	if runID == "" {
		runID = e2eshared.NewRunID()
	}
	created := time.Now().Unix()
	appName := e2eshared.ResourceName(sharedCfg, runID)
	rgName := appName + "-rg"
	t.Logf("run id %s -> web app %s", runID, appName)

	// Run from a copy of the repository so local state never leaks between runs.
	repoCopy, err := files.CopyTerraformFolderToTemp("..", "e2e-")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(repoCopy) })
	tfOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: filepath.Join(repoCopy, fixtureDir),
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
		"another operation is in progress":               "App Service operation in progress; retrying.",
		"Cannot modify this web hosting plan":            "App Service plan busy; retrying.",
		"AnotherOperationInProgress":                     "Concurrent operation; retrying.",
		"(?s)429.*(TooManyRequests|Throttl)":             "ARM throttling; retrying.",
		"No available instances to satisfy this request": "App Service capacity shortage; retrying.",
	} {
		tfOpts.RetryableTerraformErrors[pattern] = message
	}
	tfOpts.MaxRetries = 3
	tfOpts.TimeBetweenRetries = 30 * time.Second

	defer runPhase(t, "teardown", func() {
		if _, err := terraform.DestroyE(t, tfOpts); err != nil {
			t.Errorf("terraform destroy failed: %v", err)
			if err := az.DeleteResourceGroup(ctx, rgName); err != nil {
				t.Errorf("fallback delete of resource group %q failed: %v", rgName, err)
			}
			return
		}
		require.NoError(t, verify.Removed(ctx, az, rgName), "destroy must remove the resource group")
	})

	runPhase(t, "instrumentation deploy", func() { terraform.InitAndApply(t, tfOpts) })

	runPhase(t, "config verification", func() {
		require.NoError(t, verify.Instrumented(ctx, az, rgName, appName, verify.Expected{
			Service:      appName,
			Site:         cfg.site,
			Env:          ddEnv,
			Version:      ddVersion,
			Runtime:      nodeRuntime,
			RunID:        runID,
			CreatedTS:    strconv.FormatInt(created, 10),
			SidecarImage: cfg.sidecarImage,
			APIKey:       cfg.apiKey,
		}))
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

	runPhase(t, "telemetry wait", func() {
		checkTelemetryFlowing(t, ctx, hostname, cfg, e2eshared.IdentityFor(sharedCfg, appName, ddEnv, ddVersion, runID))
	})

	runPhase(t, "idempotency check", func() {
		exitCode := terraform.PlanExitCode(t, tfOpts)
		if exitCode != 0 {
			t.Log(terraform.Plan(t, tfOpts))
		}
		require.Equal(t, 0, exitCode, "re-apply must be a no-op")
	})
}

// checkTelemetryFlowing waits for spans and logs that carry the run's service, env,
// version, and run ID. serverless-init tails log files from the end, so traffic keeps
// fresh lines flowing while the logs poll runs.
func checkTelemetryFlowing(t *testing.T, ctx context.Context, hostname string, cfg config, id e2eshared.Identity) {
	t.Helper()
	client := e2eshared.NewTelemetryClient(cfg.site, cfg.apiKey, cfg.appKey)

	trafficCtx, stopTraffic := context.WithCancel(ctx)
	defer stopTraffic()
	go e2eshared.GenerateTraffic(trafficCtx, "https://"+hostname, 5*time.Second)

	type search struct {
		label string
		run   func(context.Context, string) ([]e2eshared.Event, error)
		query string
		id    e2eshared.Identity
	}
	// The shared parser does not read span tags under `custom`, where spans carry
	// version, so spans filter on @version server-side and match the rest in-process.
	spanID := id
	spanID.Version = ""
	searches := []search{
		{"spans", client.SearchSpans, e2eshared.SpanQuery(id) + " @version:" + id.Version, spanID},
		{"logs", client.SearchLogs, e2eshared.LogQuery(id), id},
	}
	errs := make(chan error, len(searches))
	for _, s := range searches {
		go func() {
			_, err := client.WaitForMatching(ctx, s.label, s.run, s.query, s.id)
			if err != nil {
				err = fmt.Errorf("%w; sample: %s", err, sampleEvent(ctx, s.run, s.query))
			}
			errs <- err
		}()
	}

	var failures []string
	for range searches {
		if err := <-errs; err != nil {
			failures = append(failures, err.Error())
		}
	}
	require.Empty(t, failures, "telemetry did not flow: %s", strings.Join(failures, "; "))
}

// sampleEvent describes one event matching query, so identity mismatches show what
// Datadog actually received.
func sampleEvent(ctx context.Context, search func(context.Context, string) ([]e2eshared.Event, error), query string) string {
	events, err := search(ctx, query)
	switch {
	case err != nil:
		return err.Error()
	case len(events) == 0:
		return "none"
	default:
		return fmt.Sprintf("attrs=%v tags=%v", events[0].Attrs, events[0].Tags)
	}
}

func runPhase(t *testing.T, name string, run func()) {
	t.Helper()
	started := time.Now()
	t.Logf("START: %s", name)

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Logf("RUNNING: %s (%s elapsed)", name, time.Since(started).Round(time.Second))
			case <-stop:
				return
			}
		}
	}()
	defer func() {
		close(stop)
		<-done
		t.Logf("DONE: %s (%s)", name, time.Since(started).Round(time.Second))
	}()

	run()
}

// triggerWorkload waits for the workload to serve a 2xx response.
func triggerWorkload(t *testing.T, hostname string) {
	t.Helper()
	url := "https://" + hostname + "/"
	client := &http.Client{Timeout: 30 * time.Second}

	const maxAttempts = 20
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			t.Logf("invoke attempt %d/%d: %v", attempt, maxAttempts, err)
		} else {
			resp.Body.Close()
			t.Logf("invoke attempt %d/%d: HTTP %d", attempt, maxAttempts, resp.StatusCode)
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				return
			}
		}
		time.Sleep(10 * time.Second)
	}
	t.Fatalf("workload at %s did not return 2xx", url)
}
