package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type parsedMetrics struct {
	metrics      metrics
	finalMessage string
	sessionID    string
}

func parseMetrics(path string) (parsedMetrics, error) {
	file, err := os.Open(path)
	if err != nil {
		return parsedMetrics{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	out := parsedMetrics{
		metrics: metrics{
			EventTypeCounts:          map[string]int{},
			CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not from OS-level tracing.",
		},
	}
	commandIDs := map[string]struct{}{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		out.metrics.EventTypeCounts[event.Type]++
		if event.Type == "thread.started" && event.ThreadID != "" && out.sessionID == "" {
			out.sessionID = event.ThreadID
		}
		switch event.Item.Type {
		case "agent_message":
			if event.Type == "item.completed" {
				out.metrics.AssistantCalls++
				out.finalMessage = event.Item.Text
			}
		case "command_execution":
			if event.Item.ID != "" {
				commandIDs[event.Item.ID] = struct{}{}
			}
			if event.Type == "item.completed" {
				out.metrics.CommandExecutions++
				if isFileInspectionCommand(event.Item.Command) {
					out.metrics.FileInspectionCommands++
				}
				if isBroadRepoSearchCommand(event.Item.Command) {
					out.metrics.BroadRepoSearch = true
					addMetricEvidence(&out.metrics.BroadRepoSearchEvidence, event.Item.Command)
					if mentionsGeneratedPath(event.Item.AggregatedOutput) {
						out.metrics.GeneratedPathFromBroadSearch = true
						addMetricEvidence(&out.metrics.GeneratedPathFromBroadSearchEvidence, event.Item.Command)
					}
				}
				if inspectsGeneratedFileCommand(event.Item.Command, event.Item.AggregatedOutput) {
					out.metrics.GeneratedFileInspected = true
					addMetricEvidence(&out.metrics.GeneratedFileEvidence, event.Item.Command)
				}
				if inspectsModuleCache(event.Item.Command) {
					out.metrics.ModuleCacheInspected = true
					addMetricEvidence(&out.metrics.ModuleCacheEvidence, event.Item.Command)
				}
				if usesOpenHealthCLI(event.Item.Command) {
					out.metrics.CLIUsed = true
					addMetricEvidence(&out.metrics.CLIUsageEvidence, event.Item.Command)
				}
				if usesDirectSQLite(event.Item.Command) {
					out.metrics.DirectSQLiteAccess = true
					addMetricEvidence(&out.metrics.DirectSQLiteEvidence, event.Item.Command)
				}
			}
		}
		if event.Usage != nil {
			input := event.Usage.InputTokens
			cached := event.Usage.CachedInputTokens
			nonCached := input - cached
			output := event.Usage.OutputTokens
			out.metrics.UsageExposed = true
			out.metrics.InputTokens = &input
			out.metrics.CachedInputTokens = &cached
			out.metrics.NonCachedInputTokens = &nonCached
			out.metrics.OutputTokens = &output
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	out.metrics.ToolCalls = len(commandIDs)
	return out, nil
}
