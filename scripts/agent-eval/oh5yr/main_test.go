package main

import (
	"strings"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func testMetrics(toolCalls int, nonCachedTokens int) metrics {
	input := nonCachedTokens
	cached := 0
	output := 10
	return metrics{
		AssistantCalls:       1,
		ToolCalls:            toolCalls,
		CommandExecutions:    toolCalls,
		UsageExposed:         true,
		InputTokens:          &input,
		CachedInputTokens:    &cached,
		NonCachedInputTokens: &input,
		OutputTokens:         &output,
		EventTypeCounts:      map[string]int{},
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func containsArgPair(args []string, key string, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

func openEvalTestClient(t *testing.T, databasePath string) *client.LocalClient {
	t.Helper()

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
	if err != nil {
		t.Fatalf("OpenLocal: %v", err)
	}
	return api
}

func TestParseRunOptionsDefaultsParallelismAndCacheMode(t *testing.T) {
	t.Parallel()

	options, err := parseRunOptions(nil)
	if err != nil {
		t.Fatalf("parseRunOptions: %v", err)
	}
	if options.Parallelism != defaultRunParallelism {
		t.Fatalf("parallelism = %d, want %d", options.Parallelism, defaultRunParallelism)
	}
	if options.CacheMode != cacheModeShared {
		t.Fatalf("cache mode = %q, want %q", options.CacheMode, cacheModeShared)
	}

	options, err = parseRunOptions([]string{"--parallel", "1", "--cache-mode", "isolated"})
	if err != nil {
		t.Fatalf("parseRunOptions --parallel 1 --cache-mode isolated: %v", err)
	}
	if options.Parallelism != 1 {
		t.Fatalf("parallelism = %d, want 1", options.Parallelism)
	}
	if options.CacheMode != cacheModeIsolated {
		t.Fatalf("cache mode = %q, want %q", options.CacheMode, cacheModeIsolated)
	}

	if _, err := parseRunOptions([]string{"--parallel", "0"}); err == nil || !strings.Contains(err.Error(), "parallel must be greater than or equal to 1") {
		t.Fatalf("parseRunOptions --parallel 0 error = %v, want validation error", err)
	}
	if _, err := parseRunOptions([]string{"--cache-mode", "bad"}); err == nil || !strings.Contains(err.Error(), "cache-mode must be") {
		t.Fatalf("parseRunOptions --cache-mode bad error = %v, want validation error", err)
	}
}
