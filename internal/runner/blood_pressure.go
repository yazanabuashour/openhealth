package runner

import (
	"context"
	"fmt"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	BloodPressureTaskActionRecord   = "record_blood_pressure"
	BloodPressureTaskActionCorrect  = "correct_blood_pressure"
	BloodPressureTaskActionList     = "list_blood_pressure"
	BloodPressureTaskActionValidate = "validate"

	BloodPressureListModeLatest  = "latest"
	BloodPressureListModeHistory = "history"
	BloodPressureListModeRange   = "range"
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

func RunBloodPressureTask(ctx context.Context, config client.LocalConfig, request BloodPressureTaskRequest) (BloodPressureTaskResult, error) {
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

	api, err := client.OpenLocal(config)
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case BloodPressureTaskActionRecord:
		return runBloodPressureRecord(ctx, api, normalized)
	case BloodPressureTaskActionCorrect:
		return runBloodPressureCorrect(ctx, api, normalized)
	case BloodPressureTaskActionList:
		return runBloodPressureList(ctx, api, normalized)
	default:
		return BloodPressureTaskResult{}, fmt.Errorf("unsupported blood pressure task action %q", normalized.Action)
	}
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

	if request.Limit < 0 {
		return normalizedBloodPressureTaskRequest{}, "limit must be greater than or equal to 0"
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
	if normalized.ListMode == "" {
		normalized.ListMode = BloodPressureListModeHistory
	}
	switch normalized.ListMode {
	case BloodPressureListModeLatest:
		normalized.Limit = 1
	case BloodPressureListModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case BloodPressureListModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedBloodPressureTaskRequest{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedBloodPressureTaskRequest{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedBloodPressureTaskRequest{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedBloodPressureTaskRequest{}, fmt.Sprintf("unsupported blood pressure list mode %q", normalized.ListMode)
	}
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

func runBloodPressureRecord(ctx context.Context, api *client.LocalClient, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	result := BloodPressureTaskResult{}
	for _, reading := range request.Readings {
		written, err := api.RecordBloodPressure(ctx, client.BloodPressureRecordInput{
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
	entries, err := api.ListBloodPressure(ctx, client.BloodPressureListOptions{Limit: 100})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	result.Entries = bloodPressureTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d blood pressure entries", len(result.Entries))
	return result, nil
}

type bloodPressureCorrectionTarget struct {
	input    normalizedBloodPressureInput
	existing client.BloodPressureEntry
}

func runBloodPressureCorrect(ctx context.Context, api *client.LocalClient, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	targets, rejection, err := bloodPressureCorrectionTargets(ctx, api, request.Readings)
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
		written, err := api.ReplaceBloodPressure(ctx, target.existing.ID, client.BloodPressureRecordInput{
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
	entries, err := api.ListBloodPressure(ctx, client.BloodPressureListOptions{Limit: 100})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	result.Entries = bloodPressureTaskEntries(entries)
	result.Summary = fmt.Sprintf("stored %d blood pressure entries", len(result.Entries))
	return result, nil
}

func bloodPressureCorrectionTargets(ctx context.Context, api *client.LocalClient, readings []normalizedBloodPressureInput) ([]bloodPressureCorrectionTarget, string, error) {
	targets := make([]bloodPressureCorrectionTarget, 0, len(readings))
	for _, reading := range readings {
		date := reading.RecordedAt.Format(time.DateOnly)
		to := reading.RecordedAt.Add(24*time.Hour - time.Nanosecond)
		existing, err := api.ListBloodPressure(ctx, client.BloodPressureListOptions{
			From:  &reading.RecordedAt,
			To:    &to,
			Limit: 2,
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

func runBloodPressureList(ctx context.Context, api *client.LocalClient, request normalizedBloodPressureTaskRequest) (BloodPressureTaskResult, error) {
	entries, err := api.ListBloodPressure(ctx, client.BloodPressureListOptions{
		From:  request.From,
		To:    request.To,
		Limit: request.Limit,
	})
	if err != nil {
		return BloodPressureTaskResult{}, err
	}
	return BloodPressureTaskResult{
		Entries: bloodPressureTaskEntries(entries),
		Summary: fmt.Sprintf("returned %d blood pressure entries", len(entries)),
	}, nil
}

func bloodPressureTaskEntries(entries []client.BloodPressureEntry) []BloodPressureEntry {
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
