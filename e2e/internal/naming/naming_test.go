package naming

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestNewRunID(t *testing.T) {
	id := NewRunID()
	if len(id) != 8 {
		t.Fatalf("run id %q: want 8 hex chars, got %d", id, len(id))
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}$`).MatchString(id) {
		t.Fatalf("run id %q is not lowercase hex", id)
	}
	if NewRunID() == id {
		t.Fatalf("run ids should differ between calls")
	}
}

func TestWebAppNameHygiene(t *testing.T) {
	name := WebAppName("deadbeef")
	if !strings.HasPrefix(name, Prefix) {
		t.Fatalf("web app name %q must start with sweeper prefix %q", name, Prefix)
	}
	// Azure Web App names: <= 60 chars, DNS-valid.
	if len(name) > 60 {
		t.Fatalf("web app name %q exceeds 60 chars", name)
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
		t.Fatalf("web app name %q is not DNS-valid", name)
	}
}

func TestTags(t *testing.T) {
	tags := Tags("deadbeef", 1700000000)
	if tags[RunIDTagKey] != "deadbeef" {
		t.Fatalf("run id tag = %q", tags[RunIDTagKey])
	}
	ts, err := strconv.ParseInt(tags[FreshnessTagKey], 10, 64)
	if err != nil || ts != 1700000000 {
		t.Fatalf("freshness tag = %q (err %v)", tags[FreshnessTagKey], err)
	}
}
