package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	WeightTaskActionUpsert   = "upsert_weights"
	WeightTaskActionList     = "list_weights"
	WeightTaskActionValidate = taskActionValidate

	WeightListModeLatest  = listModeLatest
	WeightListModeHistory = listModeHistory
	WeightListModeRange   = listModeRange
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
	Note  *string `json:"note,omitempty"`
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
	Note   *string `json:"note,omitempty"`
	Status string  `json:"status"`
}

type WeightTaskEntry struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Note  *string `json:"note,omitempty"`
}

func RunWeightTask(ctx context.Context, config localruntime.Config, request WeightTaskRequest) (WeightTaskResult, error) {
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

	return withService(ctx, config, func(ctx context.Context, service health.Service) (WeightTaskResult, error) {
		switch normalized.Action {
		case WeightTaskActionUpsert:
			return runWeightUpsert(ctx, service, normalized)
		case WeightTaskActionList:
			return runWeightList(ctx, service, normalized)
		default:
			return WeightTaskResult{}, fmt.Errorf("unsupported weight task action %q", normalized.Action)
		}
	})
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
	Unit       health.WeightUnit
	Note       *string
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
	if rejection := rejectNegativeLimit(request.Limit); rejection != "" {
		return normalizedWeightTaskRequest{}, rejection
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
	list, rejection := normalizeTaskListRequest(taskListRequest{
		ListMode: request.ListMode,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
	}, "weight")
	if rejection != "" {
		return normalizedWeightTaskRequest{}, rejection
	}
	normalized.ListMode = list.ListMode
	normalized.From = list.From
	normalized.To = list.To
	normalized.Limit = list.Limit
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
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedWeightInput{}, rejection
	}
	return normalizedWeightInput{
		RecordedAt: recordedAt,
		Value:      input.Value,
		Unit:       unit,
		Note:       note,
	}, ""
}

func normalizeUnit(value string) (health.WeightUnit, string) {
	switch value {
	case "lb", "lbs", "pound", "pounds":
		return health.WeightUnitLb, ""
	default:
		return "", "unit must be lb"
	}
}

func runWeightUpsert(ctx context.Context, service health.Service, request normalizedWeightTaskRequest) (WeightTaskResult, error) {
	result := WeightTaskResult{}
	for _, weight := range request.Weights {
		written, err := service.UpsertWeight(ctx, health.WeightRecordInput{
			RecordedAt: weight.RecordedAt,
			Value:      weight.Value,
			Unit:       weight.Unit,
			Note:       weight.Note,
		})
		if err != nil {
			return WeightTaskResult{}, err
		}
		result.Writes = append(result.Writes, WeightWrite{
			Date:   written.Entry.RecordedAt.Format(time.DateOnly),
			Value:  written.Entry.Value,
			Unit:   string(written.Entry.Unit),
			Note:   written.Entry.Note,
			Status: string(written.Status),
		})
	}
	limit := 100
	entries, err := service.ListWeight(ctx, health.HistoryFilter{Limit: &limit})
	if err != nil {
		return WeightTaskResult{}, err
	}
	result.Entries = weightTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d weight entries", len(result.Entries))
	return result, nil
}

func runWeightList(ctx context.Context, service health.Service, request normalizedWeightTaskRequest) (WeightTaskResult, error) {
	entries, err := service.ListWeight(ctx, health.HistoryFilter{
		From:  request.From,
		To:    request.To,
		Limit: limitPointer(request.Limit),
	})
	if err != nil {
		return WeightTaskResult{}, err
	}
	return WeightTaskResult{
		Entries: weightTaskEntries(entries),
		Summary: fmt.Sprintf("returned %d weight entries", len(entries)),
	}, nil
}

func weightTaskEntries(entries []health.WeightEntry) []WeightTaskEntry {
	out := make([]WeightTaskEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, WeightTaskEntry{
			Date:  entry.RecordedAt.Format(time.DateOnly),
			Value: entry.Value,
			Unit:  string(entry.Unit),
			Note:  entry.Note,
		})
	}
	return out
}
