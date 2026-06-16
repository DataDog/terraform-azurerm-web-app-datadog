// Package telemetry polls the Datadog API for traces and logs emitted by the
// instrumented workload, filtered by the run-unique service. It mirrors the
// reference impl's cloud-run-telemetry-checker.ts: poll spans and logs in
// parallel, 15s x 20 attempts, on a bounded budget.
//
// Per the conformance contract we assert *identity, not existence*: the
// identifying tags (service, env, version) are baked into the search query, so
// a non-empty result set proves those tags are present on ingested telemetry.
// We additionally re-read the returned items and confirm the tag values match.
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
	pollInterval = 15 * time.Second
	maxAttempts  = 20
	lookback     = 15 * time.Minute
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
// telemetry.
type Expected struct {
	Service string
	Env     string
	Version string
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

// CheckTelemetryFlowing blocks until both traces and logs identified by exp are
// found in Datadog, or the polling budget is exhausted. It returns an error
// describing which signal never arrived.
func CheckTelemetryFlowing(ctx context.Context, cfg Config, exp Expected) error {
	client := datadog.NewAPIClient(datadog.NewConfiguration())
	authCtx := cfg.context(ctx)

	type result struct {
		label string
		err   error
	}
	done := make(chan result, 2)

	go func() {
		done <- result{"spans", pollUntilFound(authCtx, "spans", func() error {
			return querySpans(authCtx, client, exp)
		})}
	}()
	go func() {
		done <- result{"logs", pollUntilFound(authCtx, "logs", func() error {
			return queryLogs(authCtx, client, exp)
		})}
	}()

	var errs []string
	for i := 0; i < 2; i++ {
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

// querySpans searches APM spans with the identity baked into the query, then
// verifies the returned span carries the expected service/env/version tags.
func querySpans(ctx context.Context, client *datadog.APIClient, exp Expected) error {
	from, to := window()
	q := fmt.Sprintf("service:%s env:%s version:%s", exp.Service, exp.Env, exp.Version)
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
	attrs := resp.Data[0].Attributes
	if attrs == nil {
		return fmt.Errorf("span missing attributes")
	}
	if svc := attrs.GetService(); svc != exp.Service {
		return fmt.Errorf("span service %q != expected %q", svc, exp.Service)
	}
	if err := assertTags(attrs.GetTags(), exp); err != nil {
		return fmt.Errorf("span %w", err)
	}
	return nil
}

// queryLogs searches logs by the run-unique service and env, then verifies the
// returned log carries the expected service/env tags.
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
	attrs := resp.Data[0].Attributes
	if attrs == nil {
		return fmt.Errorf("log missing attributes")
	}
	if svc := attrs.GetService(); svc != exp.Service {
		return fmt.Errorf("log service %q != expected %q", svc, exp.Service)
	}
	if err := assertTags(attrs.GetTags(), Expected{Service: exp.Service, Env: exp.Env}); err != nil {
		return fmt.Errorf("log %w", err)
	}
	return nil
}

// assertTags confirms the unified-service-tagging values are present on the
// ingested item's tag list (identity, not mere existence). Empty fields in exp
// are skipped.
func assertTags(tags []string, exp Expected) error {
	want := map[string]string{}
	if exp.Env != "" {
		want["env"] = exp.Env
	}
	if exp.Version != "" {
		want["version"] = exp.Version
	}
	for k, v := range want {
		if !hasTag(tags, k, v) {
			return fmt.Errorf("missing identifying tag %s:%s (got %v)", k, v, tags)
		}
	}
	return nil
}

func hasTag(tags []string, key, value string) bool {
	target := key + ":" + value
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}
