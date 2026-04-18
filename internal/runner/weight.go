package runner

import (
	"context"
	"fmt"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	WeightTaskActionUpsert   = "upsert_weights"
	WeightTaskActionList     = "list_weights"
	WeightTaskActionValidate = "validate"

	WeightListModeLatest  = "latest"
	WeightListModeHistory = "history"
	WeightListModeRange   = "range"
)

type WeightTaskRequest struct {
	Action   string        `json:"action"`
	Weights  []WeightInput `json:"weights,omitempty"`
	ListMode string        `json:"list_mode,omitempty"`
	FromDate string        `json:"from_date,omitempty"`
	ToDate   string        `json:"to_date,omitempty"`
	Limit    int           `json:"limit,omitempty"`
}

type WeightInput struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type WeightTaskResult struct {
	Rejected        bool              `json:"rejected"`
	RejectionReason string            `json:"rejection_reason,omitempty"`
	Writes          []WeightWrite     `json:"writes,omitempty"`
	Entries         []WeightTaskEntry `json:"entries,omitempty"`
	Summary         string            `json:"summary"`
}

type WeightWrite struct {
	Date   string  `json:"date"`
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Status string  `json:"status"`
}

type WeightTaskEntry struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func RunWeightTask(ctx context.Context, config client.LocalConfig, request WeightTaskRequest) (WeightTaskResult, error) {
	normalized, rejection := normalizeWeightTaskRequest(request)
	if rejection != "" {
		return WeightTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == WeightTaskActionValidate {
		return WeightTaskResult{Summary: "valid"}, nil
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		return WeightTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case WeightTaskActionUpsert:
		return runWeightUpsert(ctx, api, normalized)
	case WeightTaskActionList:
		return runWeightList(ctx, api, normalized)
	default:
		return WeightTaskResult{}, fmt.Errorf("unsupported weight task action %q", normalized.Action)
	}
}

type normalizedWeightTaskRequest struct {
	Action   string
	Weights  []normalizedWeightInput
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
}

type normalizedWeightInput struct {
	RecordedAt time.Time
	Value      float64
	Unit       client.WeightUnit
}

func normalizeWeightTaskRequest(request WeightTaskRequest) (normalizedWeightTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = WeightTaskActionValidate
	}
	normalized := normalizedWeightTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}

	if request.Limit < 0 {
		return normalizedWeightTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case WeightTaskActionValidate:
		for _, weight := range request.Weights {
			if _, rejection := normalizeWeightInput(weight); rejection != "" {
				return normalizedWeightTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case WeightTaskActionUpsert:
		if len(request.Weights) == 0 {
			return normalizedWeightTaskRequest{}, "weights are required"
		}
		for _, weight := range request.Weights {
			parsed, rejection := normalizeWeightInput(weight)
			if rejection != "" {
				return normalizedWeightTaskRequest{}, rejection
			}
			normalized.Weights = append(normalized.Weights, parsed)
		}
		return normalized, ""
	case WeightTaskActionList:
		return normalizeWeightListRequest(normalized, request)
	default:
		return normalizedWeightTaskRequest{}, fmt.Sprintf("unsupported weight task action %q", action)
	}
}

func normalizeWeightListRequest(normalized normalizedWeightTaskRequest, request WeightTaskRequest) (normalizedWeightTaskRequest, string) {
	if normalized.ListMode == "" {
		normalized.ListMode = WeightListModeHistory
	}
	switch normalized.ListMode {
	case WeightListModeLatest:
		normalized.Limit = 1
	case WeightListModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case WeightListModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedWeightTaskRequest{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedWeightTaskRequest{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedWeightTaskRequest{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedWeightTaskRequest{}, fmt.Sprintf("unsupported weight list mode %q", normalized.ListMode)
	}
	return normalized, ""
}

func normalizeWeightInput(input WeightInput) (normalizedWeightInput, string) {
	recordedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedWeightInput{}, rejection
	}
	if input.Value <= 0 {
		return normalizedWeightInput{}, "value must be greater than 0"
	}
	unit, rejection := normalizeUnit(input.Unit)
	if rejection != "" {
		return normalizedWeightInput{}, rejection
	}
	return normalizedWeightInput{
		RecordedAt: recordedAt,
		Value:      input.Value,
		Unit:       unit,
	}, ""
}

func parseDateOnly(value string) (time.Time, string) {
	if value == "" {
		return time.Time{}, "date must be YYYY-MM-DD"
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil || parsed.Format(time.DateOnly) != value {
		return time.Time{}, "date must be YYYY-MM-DD"
	}
	return parsed, ""
}

func normalizeUnit(value string) (client.WeightUnit, string) {
	switch value {
	case "lb", "lbs", "pound", "pounds":
		return client.WeightUnitLb, ""
	default:
		return "", "unit must be lb"
	}
}

func runWeightUpsert(ctx context.Context, api *client.LocalClient, request normalizedWeightTaskRequest) (WeightTaskResult, error) {
	result := WeightTaskResult{}
	for _, weight := range request.Weights {
		written, err := api.UpsertWeight(ctx, client.WeightRecordInput{
			RecordedAt: weight.RecordedAt,
			Value:      weight.Value,
			Unit:       weight.Unit,
		})
		if err != nil {
			return WeightTaskResult{}, err
		}
		result.Writes = append(result.Writes, WeightWrite{
			Date:   written.Entry.RecordedAt.Format(time.DateOnly),
			Value:  written.Entry.Value,
			Unit:   string(written.Entry.Unit),
			Status: string(written.Status),
		})
	}
	entries, err := api.ListWeights(ctx, client.WeightListOptions{Limit: 100})
	if err != nil {
		return WeightTaskResult{}, err
	}
	result.Entries = weightTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d weight entries", len(result.Entries))
	return result, nil
}

func runWeightList(ctx context.Context, api *client.LocalClient, request normalizedWeightTaskRequest) (WeightTaskResult, error) {
	entries, err := api.ListWeights(ctx, client.WeightListOptions{
		From:  request.From,
		To:    request.To,
		Limit: request.Limit,
	})
	if err != nil {
		return WeightTaskResult{}, err
	}
	return WeightTaskResult{
		Entries: weightTaskEntries(entries),
		Summary: fmt.Sprintf("returned %d weight entries", len(entries)),
	}, nil
}

func weightTaskEntries(entries []client.WeightEntry) []WeightTaskEntry {
	out := make([]WeightTaskEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, WeightTaskEntry{
			Date:  entry.RecordedAt.Format(time.DateOnly),
			Value: entry.Value,
			Unit:  string(entry.Unit),
		})
	}
	return out
}
