package telemetry

import "testing"

func TestHasTag(t *testing.T) {
	tags := []string{"env:e2e", "version:1.0.0", "service:one-e2e-tfwebapp-linux-deadbeef"}
	if !hasTag(tags, "env", "e2e") {
		t.Fatal("expected env:e2e to be present")
	}
	if hasTag(tags, "env", "prod") {
		t.Fatal("env:prod should not match")
	}
	if hasTag(nil, "env", "e2e") {
		t.Fatal("nil tags should not match")
	}
}

func TestAssertTagsIdentity(t *testing.T) {
	tags := []string{"env:e2e", "version:1.0.0"}

	if err := assertTags(tags, Expected{Env: "e2e", Version: "1.0.0"}); err != nil {
		t.Fatalf("matching tags should pass: %v", err)
	}
	// Wrong value -> identity failure (not mere existence).
	if err := assertTags(tags, Expected{Env: "e2e", Version: "2.0.0"}); err == nil {
		t.Fatal("mismatched version should fail")
	}
	// Empty expected fields are skipped.
	if err := assertTags(tags, Expected{}); err != nil {
		t.Fatalf("no expectations should pass: %v", err)
	}
}
