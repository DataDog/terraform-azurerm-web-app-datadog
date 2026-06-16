// Package naming implements the shared e2e resource-hygiene convention from
// serverless-ci/e2e/spec.md: every ephemeral resource is created with a
// deterministic name prefix and a freshness tag, both set atomically at
// creation. The prefix is the identity + blast-radius guard the cross-repo
// sweeper keys on; the freshness tag lets the sweeper skip in-flight tests.
package naming

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	// Tool identifies the delivery tool under test (this Terraform module).
	Tool = "tfwebapp"
	// Platform is the canonical runtime/platform for this repo's coverage.
	Platform = "linux"
	// Prefix is the sweeper's blast-radius guard. "one" is the team (dd- implied).
	Prefix = "one-e2e-" + Tool + "-" + Platform
	// FreshnessTagKey is set in the create call with a unix-ts value. Native
	// creation time isn't usable cross-cloud, so we record it ourselves.
	FreshnessTagKey = "one_e2e_created"
	// RunIDTagKey carries the run id so telemetry and resources cross-reference.
	RunIDTagKey = "one_e2e_runid"
)

// NewRunID returns an 8-char hex run id, matching the reference impl's
// crypto.randomBytes(4).toString('hex').
func NewRunID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("naming: cannot read random bytes: %v", err))
	}
	return hex.EncodeToString(b)
}

// WebAppName returns the globally-unique, DNS-valid web app name for a run.
// e.g. one-e2e-tfwebapp-linux-deadbeef (well under the 60-char Web App budget).
func WebAppName(runID string) string {
	return fmt.Sprintf("%s-%s", Prefix, runID)
}

// ResourceGroupName returns the run-scoped resource group name. It shares the
// sweeper prefix so a leaked group is reaped wholesale.
func ResourceGroupName(runID string) string {
	return fmt.Sprintf("%s-%s-rg", Prefix, runID)
}

// Tags returns the hygiene tags to stamp on every created resource: the
// freshness tag (unix seconds) and the run-id marker.
func Tags(runID string, createdUnix int64) map[string]string {
	return map[string]string{
		FreshnessTagKey: fmt.Sprintf("%d", createdUnix),
		RunIDTagKey:     runID,
	}
}
