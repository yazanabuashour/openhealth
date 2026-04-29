package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	BloodPressureTaskActionRecord   = "record_blood_pressure"
	BloodPressureTaskActionCorrect  = "correct_blood_pressure"
	BloodPressureTaskActionList     = "list_blood_pressure"
	BloodPressureTaskActionValidate = taskActionValidate

	BloodPressureListModeLatest  = listModeLatest
	BloodPressureListModeHistory = listModeHistory
	BloodPressureListModeRange   = listModeRange
)

type BloodPressureTaskRequest struct {
	Action   string               `json:"action"`
	Readings []BloodPressureInput `json:"readings,omitempty"`
	ListMode string               `json:"list_mode,omitempty"`
	FromDate string               `json:"from_date,omitempty"`
	ToDate   string               `json:"to_date,omitempty"`
	Limit    int                  `json:"limit,omitempty"`
}

type BloodPressureInput struct {
	Date      string  `json:"date"`
	Systolic  int     `json:"systolic"`
	Diastolic int     `json:"diastolic"`
	Pulse     *int    `json:"pulse,omitempty"`
	Note      *string `json:"note,omitempty"`
}

type BloodPressureTaskResult struct {
	Rejected        bool                 `json:"rejected"`
	RejectionReason string               `json:"rejection_reason,omitempty"`
	Writes          []BloodPressureWrite `json:"writes,omitempty"`
	Entries         []BloodPressureEntry `json:"entries,omitempty"`
	Summary         string               `json:"summary"`
}

type BloodPressureWrite struct {
	Date      string  `json:"date"`
	Systolic  int     `json:"systolic"`
	Diastolic int     `json:"diastolic"`
	Pulse     *int    `json:"pulse,omitempty"`
	Note      *string `json:"note,omitempty"`
	Status    string  `json:"status"`
}

type BloodPressureEntry struct {
	Date      string  `json:"date"`
	Systolic  int     `json:"systolic"`
	Diastolic int     `json:"diastolic"`
	Pulse     *int    `json:"pulse,omitempty"`
	Note      *string `json:"note,omitempty"`
}

func RunBloodPressureTask(ctx context.Context, config localruntime.Config, request BloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	normalized, rejection := normalizeBloodPressureTaskRequest(request)
	if rejection != "" {
		return BloodPressureTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == BloodPressureTaskActionValidate {
		return BloodPressureTaskResult{Summary: "valid"}, nil
	}

	return withService(ctx, config, func(ctx context.Context, service health.Service) (BloodPressureTaskResult, error) {
		switch normalized.Action {
		case BloodPressureTaskActionRecord:
			return runBloodPressureRecord(ctx, service, normalized)
		case BloodPressureTaskActionCorrect:
			return runBloodPressureCorrect(ctx, service, normalized)
		case BloodPressureTaskActionList:
			return runBloodPressureList(ctx, service, normalized)
		default:
			return BloodPressureTaskResult{}, fmt.Errorf("unsupported blood pressure task action %q", normalized.Action)
		}
	})
}

type normalizedBloodPressureTaskRequest struct {
	Action   string
	Readings []normalizedBloodPressureInput
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
}

type normalizedBloodPressureInput struct {
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
	Note       *string
}

func normalizeBloodPressureTaskRequest(request BloodPressureTaskRequest) (normalizedBloodPressureTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = BloodPressureTaskActionValidate
	}
	normalized := normalizedBloodPressureTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}

	if rejection := rejectNegativeLimit(request.Limit); rejection != "" {
		return normalizedBloodPressureTaskRequest{}, rejection
	}

	switch action {
	case BloodPressureTaskActionValidate:
		for _, reading := range request.Readings {
			if _, rejection := normalizeBloodPressureInput(reading); rejection != "" {
				return normalizedBloodPressureTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case BloodPressureTaskActionRecord, BloodPressureTaskActionCorrect:
		if len(request.Readings) == 0 {
			return normalizedBloodPressureTaskRequest{}, "readings are required"
		}
		correctionDates := map[string]struct{}{}
		for _, reading := range request.Readings {
			parsed, rejection := normalizeBloodPressureInput(reading)
			if rejection != "" {
				return normalizedBloodPressureTaskRequest{}, rejection
			}
			if action == BloodPressureTaskActionCorrect {
				date := parsed.RecordedAt.Format(time.DateOnly)
				if _, ok := correctionDates[date]; ok {
					return normalizedBloodPressureTaskRequest{}, fmt.Sprintf("duplicate correction date %s", date)
				}
				correctionDates[date] = struct{}{}
			}
			normalized.Readings = append(normalized.Readings, parsed)
		}
		return normalized, ""
	case BloodPressureTaskActionList:
		return normalizeBloodPressureListRequest(normalized, request)
	default:
		return normalizedBloodPressureTaskRequest{}, fmt.Sprintf("unsupported blood pressure task action %q", action)
	}
}

func normalizeBloodPressureListRequest(normalized normalizedBloodPressureTaskRequest, request BloodPressureTaskRequest) (normalizedBloodPressureTaskRequest, string) {
	list, rejection := normalizeTaskListRequest(taskListRequest{
		ListMode: request.ListMode,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
	}, "blood pressure")
	if rejection != "" {
		return normalizedBloodPressureTaskRequest{}, rejection
	}
	normalized.ListMode = list.ListMode
	normalized.From = list.From
	normalized.To = list.To
	normalized.Limit = list.Limit
	return normalized, ""
}

func normalizeBloodPressureInput(input BloodPressureInput) (normalizedBloodPressureInput, string) {
	recordedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedBloodPressureInput{}, rejection
	}
	if input.Systolic <= 0 {
		return normalizedBloodPressureInput{}, "systolic must be greater than 0"
	}
	if input.Diastolic <= 0 {
		return normalizedBloodPressureInput{}, "diastolic must be greater than 0"
	}
	if input.Systolic <= input.Diastolic {
		return normalizedBloodPressureInput{}, "systolic must be greater than diastolic"
	}
	if input.Pulse != nil && *input.Pulse <= 0 {
		return normalizedBloodPressureInput{}, "pulse must be greater than 0"
	}
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedBloodPressureInput{}, rejection
	}
	return normalizedBloodPressureInput{
		RecordedAt: recordedAt,
		Systolic:   input.Systolic,
		Diastolic:  input.Diastolic,
		Pulse:      input.Pulse,
		Note:       note,
	}, ""
}

func runBloodPressureRecord(ctx context.Context, service health.Service, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	result := BloodPressureTaskResult{}
	for _, reading := range request.Readings {
		written, err := service.RecordBloodPressure(ctx, health.BloodPressureRecordInput{
			RecordedAt: reading.RecordedAt,
			Systolic:   reading.Systolic,
			Diastolic:  reading.Diastolic,
			Pulse:      reading.Pulse,
			Note:       reading.Note,
		})
		if err != nil {
			return BloodPressureTaskResult{}, err
		}
		result.Writes = append(result.Writes, BloodPressureWrite{
			Date:      written.RecordedAt.Format(time.DateOnly),
			Systolic:  written.Systolic,
			Diastolic: written.Diastolic,
			Pulse:     written.Pulse,
			Note:      written.Note,
			Status:    "created",
		})
	}
	limit := 100
	entries, err := service.ListBloodPressure(ctx, health.HistoryFilter{Limit: &limit})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	result.Entries = bloodPressureTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d blood pressure entries", len(result.Entries))
	return result, nil
}

type bloodPressureCorrectionTarget struct {
	input    normalizedBloodPressureInput
	existing health.BloodPressureEntry
}

func runBloodPressureCorrect(ctx context.Context, service health.Service, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	targets, rejection, err := bloodPressureCorrectionTargets(ctx, service, request.Readings)
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	if rejection != "" {
		return BloodPressureTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	result := BloodPressureTaskResult{}
	for _, target := range targets {
		note := target.input.Note
		if note == nil {
			note = target.existing.Note
		}
		written, err := service.ReplaceBloodPressure(ctx, target.existing.ID, health.BloodPressureRecordInput{
			RecordedAt: target.existing.RecordedAt,
			Systolic:   target.input.Systolic,
			Diastolic:  target.input.Diastolic,
			Pulse:      target.input.Pulse,
			Note:       note,
		})
		if err != nil {
			return BloodPressureTaskResult{}, err
		}
		result.Writes = append(result.Writes, BloodPressureWrite{
			Date:      written.RecordedAt.Format(time.DateOnly),
			Systolic:  written.Systolic,
			Diastolic: written.Diastolic,
			Pulse:     written.Pulse,
			Note:      written.Note,
			Status:    "updated",
		})
	}
	limit := 100
	entries, err := service.ListBloodPressure(ctx, health.HistoryFilter{Limit: &limit})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	result.Entries = bloodPressureTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d blood pressure entries", len(result.Entries))
	return result, nil
}

func bloodPressureCorrectionTargets(ctx context.Context, service health.Service, readings []normalizedBloodPressureInput) ([]bloodPressureCorrectionTarget, string, error) {
	targets := make([]bloodPressureCorrectionTarget, 0, len(readings))
	for _, reading := range readings {
		date := reading.RecordedAt.Format(time.DateOnly)
		to := reading.RecordedAt.Add(24*time.Hour - time.Nanosecond)
		limit := 2
		existing, err := service.ListBloodPressure(ctx, health.HistoryFilter{
			From:  &reading.RecordedAt,
			To:    &to,
			Limit: &limit,
		})
		if err != nil {
			return nil, "", err
		}
		switch len(existing) {
		case 0:
			return nil, fmt.Sprintf("no existing blood pressure reading for %s", date), nil
		case 1:
			targets = append(targets, bloodPressureCorrectionTarget{
				input:    reading,
				existing: existing[0],
			})
		default:
			return nil, fmt.Sprintf("multiple blood pressure readings for %s; correction is ambiguous", date), nil
		}
	}
	return targets, "", nil
}

func runBloodPressureList(ctx context.Context, service health.Service, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	entries, err := service.ListBloodPressure(ctx, health.HistoryFilter{
		From:  request.From,
		To:    request.To,
		Limit: limitPointer(request.Limit),
	})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	return BloodPressureTaskResult{
		Entries: bloodPressureTaskEntries(entries),
		Summary: fmt.Sprintf("returned %d blood pressure entries", len(entries)),
	}, nil
}

func bloodPressureTaskEntries(entries []health.BloodPressureEntry) []BloodPressureEntry {
	out := make([]BloodPressureEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, BloodPressureEntry{
			Date:      entry.RecordedAt.Format(time.DateOnly),
			Systolic:  entry.Systolic,
			Diastolic: entry.Diastolic,
			Pulse:     entry.Pulse,
			Note:      entry.Note,
		})
	}
	return out
}
