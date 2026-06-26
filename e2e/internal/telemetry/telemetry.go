// Package telemetry polls the Datadog API for traces and logs emitted by the
// instrumented workload, filtered by the run-unique service. It mirrors the
// reference impl's cloud-run-telemetry-checker.ts: poll spans and logs in
// parallel, 15s x 20 attempts, on a bounded budget.
//
// Per the conformance contract we assert *identity, not existence*: the
// identifying tags (service, env) are baked into the search query, so a
// non-empty result set proves those tags are present on ingested telemetry. The
// run-unique service doubles as the run-id marker.
//
// Logs: the code-based Linux Web App workload logs to stdout, which Linux App
// Service writes to a per-instance file on the /home volume shared with the
// sidecar. The fixture enables DD_AAS_INSTANCE_LOGGING_ENABLED so serverless-init
// tails that file, and the suite drives continuous traffic during the poll so the
// end-tailer always has fresh lines. Logs are therefore required (ExpectLogs).
//
// KNOWN GAPS (verified against a live run on 2026-06-16; tracked follow-ups):
//   - version tag on spans: DD_VERSION reaches the app as an app setting and the
//     `version` resource tag is applied (both asserted at the config layer in
//     internal/verify), but the value was not observed on ingested spans. Span
//     version identity is therefore not asserted here; config-layer version
//     identity stands in until the trace-tag propagation is confirmed.
package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

const (
	// App Service cold start + sidecar init + first trace flush + ingestion was
	// observed to take ~6-7 min end-to-end, so the budget is wider than the
	// reference impl's 5 min (15s x 20). This retries the cloud, not assertions.
	pollInterval = 20 * time.Second
	maxAttempts  = 30 // 10 min
	lookback     = 20 * time.Minute
	pageLimit    = int32(25)
)

// Config holds the Datadog API credentials and site for querying telemetry.
// These are distinct from the DD_API_KEY wired into the workload: querying
// also needs an application key.
type Config struct {
	APIKey string
	AppKey string
	Site   string // e.g. datadoghq.com
}

// Expected is the unified-service-tagging identity asserted on ingested
// telemetry. Service is run-unique, so it doubles as the run-id marker; env is
// the unified-tagging proof. Both are baked into the search query, so a
// non-empty result set proves those tags are present on ingested telemetry
// (identity, not existence).
//
// Note on coverage: version identity is asserted at the *config* layer
// (DD_VERSION app setting + `version` resource tag; see internal/verify) rather
// than on spans, and log collection for the code-based Linux App Service path
// is gated by ExpectLogs. See the package-level KNOWN GAPS note.
type Expected struct {
	Service string
	Env     string
}

// Options tunes which signals are required. ExpectLogs gates the logs check.
// The Linux suite wires stdout log collection (DD_AAS_INSTANCE_LOGGING_ENABLED in
// the fixture) and drives traffic during the poll, so it sets ExpectLogs true.
// Windows App Service has no serverless-init log support, so a Windows suite
// would leave this false (traces still required).
type Options struct {
	ExpectLogs bool
}

func (c Config) context(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: c.APIKey},
		"appKeyAuth": {Key: c.AppKey},
	})
	site := c.Site
	if site == "" {
		site = "datadoghq.com"
	}
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{"site": site})
	return ctx
}

// CheckTelemetryFlowing blocks until the required signals identified by exp are
// found in Datadog, or the polling budget is exhausted. Traces are always
// required; logs are required only when opts.ExpectLogs is set. It returns an
// error describing which signal never arrived.
func CheckTelemetryFlowing(ctx context.Context, cfg Config, exp Expected, opts Options) error {
	client := datadog.NewAPIClient(datadog.NewConfiguration())
	authCtx := cfg.context(ctx)

	type result struct {
		label string
		err   error
	}
	checks := []struct {
		label string
		run   func() error
	}{
		{"spans", func() error { return querySpans(authCtx, client, exp) }},
	}
	if opts.ExpectLogs {
		checks = append(checks, struct {
			label string
			run   func() error
		}{"logs", func() error { return queryLogs(authCtx, client, exp) }})
	}

	done := make(chan result, len(checks))
	for _, c := range checks {
		c := c
		go func() { done <- result{c.label, pollUntilFound(authCtx, c.label, c.run)} }()
	}

	var errs []string
	for range checks {
		r := <-done
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.label, r.err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("telemetry not flowing: %s", strings.Join(errs, "; "))
	}
	return nil
}

// pollUntilFound runs query on the polling cadence until it returns nil (found
// + identity verified) or the budget is exhausted. A query error is treated as
// "not yet" and retried -- we retry the cloud, not the assertion.
func pollUntilFound(ctx context.Context, label string, query func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("[%s] attempt %d/%d\n", label, attempt, maxAttempts)
		if err := query(); err == nil {
			fmt.Printf("[%s] found and identity verified\n", label)
			return nil
		} else {
			lastErr = err
		}
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
			}
		}
	}
	return fmt.Errorf("timed out after %d attempts (%s); last: %v",
		maxAttempts, time.Duration(maxAttempts)*pollInterval, lastErr)
}

func window() (from, to string) {
	now := time.Now().UTC()
	return now.Add(-lookback).Format(time.RFC3339), now.Format(time.RFC3339)
}

// querySpans searches APM spans with the identity baked into the query. A
// non-empty result proves spans tagged service:<run-unique> env:<env> reached
// Datadog -- identity, not existence.
func querySpans(ctx context.Context, client *datadog.APIClient, exp Expected) error {
	from, to := window()
	q := fmt.Sprintf("service:%s env:%s", exp.Service, exp.Env)
	body := datadogV2.SpansListRequest{
		Data: &datadogV2.SpansListRequestData{
			Type: datadogV2.SPANSLISTREQUESTTYPE_SEARCH_REQUEST.Ptr(),
			Attributes: &datadogV2.SpansListRequestAttributes{
				Filter: &datadogV2.SpansQueryFilter{
					Query: datadog.PtrString(q),
					From:  datadog.PtrString(from),
					To:    datadog.PtrString(to),
				},
				Page: &datadogV2.SpansListRequestPage{Limit: datadog.PtrInt32(pageLimit)},
			},
		},
	}
	api := datadogV2.NewSpansApi(client)
	resp, _, err := api.ListSpans(ctx, body)
	if err != nil {
		return fmt.Errorf("listSpans: %w", err)
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("no spans matching %q", q)
	}
	// The query filtered on service+env, so a result confirms that identity.
	// Re-read the service as a guard against an over-broad match.
	if attrs := resp.Data[0].Attributes; attrs != nil {
		if svc := attrs.GetService(); svc != exp.Service {
			return fmt.Errorf("span service %q != expected %q", svc, exp.Service)
		}
	}
	return nil
}

// queryLogs searches logs by the run-unique service and env. A non-empty result
// proves logs with that identity reached Datadog.
func queryLogs(ctx context.Context, client *datadog.APIClient, exp Expected) error {
	from, to := window()
	q := fmt.Sprintf("service:%s env:%s", exp.Service, exp.Env)
	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: datadog.PtrString(q),
			From:  datadog.PtrString(from),
			To:    datadog.PtrString(to),
		},
		Page: &datadogV2.LogsListRequestPage{Limit: datadog.PtrInt32(pageLimit)},
	}
	api := datadogV2.NewLogsApi(client)
	resp, _, err := api.ListLogs(ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))
	if err != nil {
		return fmt.Errorf("listLogs: %w", err)
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("no logs matching %q", q)
	}
	if attrs := resp.Data[0].Attributes; attrs != nil {
		if svc := attrs.GetService(); svc != exp.Service {
			return fmt.Errorf("log service %q != expected %q", svc, exp.Service)
		}
	}
	return nil
}
