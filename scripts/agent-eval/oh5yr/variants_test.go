package main

import (
	"strings"
	"testing"
)

func TestVariantsProductionOnly(t *testing.T) {
	t.Parallel()

	ids := map[string]bool{}
	for _, variant := range variants() {
		ids[variant.ID] = true
	}
	if len(ids) != 1 || !ids["production"] {
		t.Fatalf("variants() = %#v, want production only", variants())
	}
	for _, retired := range []string{"cli", "generated-client", "runner-code"} {
		if ids[retired] {
			t.Fatalf("variants() includes retired variant %q: %#v", retired, variants())
		}
	}
}

func TestSelectVariantsAndScenarios(t *testing.T) {
	t.Parallel()

	selectedVariants, err := selectVariants("production")
	if err != nil {
		t.Fatalf("selectVariants: %v", err)
	}
	if got := []string{selectedVariants[0].ID}; strings.Join(got, ",") != "production" {
		t.Fatalf("selected variants = %v", got)
	}
	if _, err := selectVariants("cli"); err == nil || !strings.Contains(err.Error(), `unknown variant "cli"`) {
		t.Fatalf("selectVariants(cli) error = %v, want unknown variant", err)
	}
	selectedScenarios, err := selectScenarios("add-two,bounded-range,latest-only")
	if err != nil {
		t.Fatalf("selectScenarios: %v", err)
	}
	if got := []string{selectedScenarios[0].ID, selectedScenarios[1].ID, selectedScenarios[2].ID}; strings.Join(got, ",") != "add-two,bounded-range,latest-only" {
		t.Fatalf("selected scenarios = %v", got)
	}
	if _, err := selectVariants("missing"); err == nil {
		t.Fatal("selectVariants missing id error = nil")
	}
	if _, err := selectScenarios("missing"); err == nil {
		t.Fatal("selectScenarios missing id error = nil")
	}
}
