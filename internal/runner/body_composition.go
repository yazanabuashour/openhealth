package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	BodyCompositionTaskActionRecord   = "record_body_composition"
	BodyCompositionTaskActionCorrect  = "correct_body_composition"
	BodyCompositionTaskActionDelete   = "delete_body_composition"
	BodyCompositionTaskActionList     = "list_body_composition"
	BodyCompositionTaskActionValidate = taskActionValidate

	BodyCompositionListModeLatest  = listModeLatest
	BodyCompositionListModeHistory = listModeHistory
	BodyCompositionListModeRange   = listModeRange
)

type BodyCompositionTaskRequest struct {
	Action   string                 `json:"action"`
	Records  []BodyCompositionInput `json:"records,omitempty"`
	Record   *BodyCompositionInput  `json:"record,omitempty"`
	Target   *BodyCompositionTarget `json:"target,omitempty"`
	ListMode string                 `json:"list_mode,omitempty"`
	FromDate string                 `json:"from_date,omitempty"`
	ToDate   string                 `json:"to_date,omitempty"`
	Limit    int                    `json:"limit,omitempty"`
}

type BodyCompositionInput struct {
	Date           string   `json:"date"`
	BodyFatPercent *float64 `json:"body_fat_percent,omitempty"`
	WeightValue    *float64 `json:"weight_value,omitempty"`
	WeightUnit     *string  `json:"weight_unit,omitempty"`
	Method         *string  `json:"method,omitempty"`
	Note           *string  `json:"note,omitempty"`
}

type BodyCompositionTarget struct {
	ID   int    `json:"id,omitempty"`
	Date string `json:"date,omitempty"`
}

type BodyCompositionTaskResult struct {
	Rejected        bool                       `json:"rejected"`
	RejectionReason string                     `json:"rejection_reason,omitempty"`
	Writes          []BodyCompositionWrite     `json:"writes,omitempty"`
	Entries         []BodyCompositionTaskEntry `json:"entries,omitempty"`
	Summary         string                     `json:"summary"`
}

type BodyCompositionWrite struct {
	ID             int      `json:"id"`
	Date           string   `json:"date"`
	BodyFatPercent *float64 `json:"body_fat_percent,omitempty"`
	WeightValue    *float64 `json:"weight_value,omitempty"`
	WeightUnit     *string  `json:"weight_unit,omitempty"`
	Status         string   `json:"status"`
}

type BodyCompositionTaskEntry struct {
	ID             int      `json:"id"`
	Date           string   `json:"date"`
	BodyFatPercent *float64 `json:"body_fat_percent,omitempty"`
	WeightValue    *float64 `json:"weight_value,omitempty"`
	WeightUnit     *string  `json:"weight_unit,omitempty"`
	Method         *string  `json:"method,omitempty"`
	Note           *string  `json:"note,omitempty"`
}

type normalizedBodyCompositionTaskRequest struct {
	Action   string
	Records  []normalizedBodyCompositionInput
	Record   normalizedBodyCompositionInput
	Target   normalizedBodyCompositionTarget
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
}

type normalizedBodyCompositionInput struct {
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *health.WeightUnit
	Method         *string
	Note           *string
}

type normalizedBodyCompositionTarget struct {
	ID   int
	Date *time.Time
}

func RunBodyCompositionTask(ctx context.Context, config localruntime.Config, request BodyCompositionTaskRequest) (BodyCompositionTaskResult, error) {
	normalized, rejection := normalizeBodyCompositionTaskRequest(request)
	if rejection != "" {
		return BodyCompositionTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if normalized.Action == BodyCompositionTaskActionValidate {
		return BodyCompositionTaskResult{Summary: "valid"}, nil
	}

	return withService(ctx, config, func(ctx context.Context, service health.Service) (BodyCompositionTaskResult, error) {
		switch normalized.Action {
		case BodyCompositionTaskActionRecord:
			return runBodyCompositionRecord(ctx, service, normalized)
		case BodyCompositionTaskActionCorrect:
			return runBodyCompositionCorrect(ctx, service, normalized)
		case BodyCompositionTaskActionDelete:
			return runBodyCompositionDelete(ctx, service, normalized)
		case BodyCompositionTaskActionList:
			return runBodyCompositionList(ctx, service, normalized)
		default:
			return BodyCompositionTaskResult{}, fmt.Errorf("unsupported body composition task action %q", normalized.Action)
		}
	})
}

func normalizeBodyCompositionTaskRequest(request BodyCompositionTaskRequest) (normalizedBodyCompositionTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = BodyCompositionTaskActionValidate
	}
	normalized := normalizedBodyCompositionTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}
	if rejection := rejectNegativeLimit(request.Limit); rejection != "" {
		return normalizedBodyCompositionTaskRequest{}, rejection
	}

	switch action {
	case BodyCompositionTaskActionValidate:
		for _, record := range request.Records {
			if _, rejection := normalizeBodyCompositionInput(record); rejection != "" {
				return normalizedBodyCompositionTaskRequest{}, rejection
			}
		}
		if request.Record != nil {
			if _, rejection := normalizeBodyCompositionInput(*request.Record); rejection != "" {
				return normalizedBodyCompositionTaskRequest{}, rejection
			}
		}
		if request.Target != nil {
			if _, rejection := normalizeBodyCompositionTarget(*request.Target); rejection != "" {
				return normalizedBodyCompositionTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case BodyCompositionTaskActionRecord:
		if len(request.Records) == 0 {
			return normalizedBodyCompositionTaskRequest{}, "records are required"
		}
		for _, record := range request.Records {
			parsed, rejection := normalizeBodyCompositionInput(record)
			if rejection != "" {
				return normalizedBodyCompositionTaskRequest{}, rejection
			}
			normalized.Records = append(normalized.Records, parsed)
		}
		return normalized, ""
	case BodyCompositionTaskActionCorrect:
		if request.Target == nil {
			return normalizedBodyCompositionTaskRequest{}, "target is required"
		}
		target, rejection := normalizeBodyCompositionTarget(*request.Target)
		if rejection != "" {
			return normalizedBodyCompositionTaskRequest{}, rejection
		}
		if request.Record == nil {
			return normalizedBodyCompositionTaskRequest{}, "record is required"
		}
		record, rejection := normalizeBodyCompositionInput(*request.Record)
		if rejection != "" {
			return normalizedBodyCompositionTaskRequest{}, rejection
		}
		normalized.Target = target
		normalized.Record = record
		return normalized, ""
	case BodyCompositionTaskActionDelete:
		if request.Target == nil {
			return normalizedBodyCompositionTaskRequest{}, "target is required"
		}
		target, rejection := normalizeBodyCompositionTarget(*request.Target)
		if rejection != "" {
			return normalizedBodyCompositionTaskRequest{}, rejection
		}
		normalized.Target = target
		return normalized, ""
	case BodyCompositionTaskActionList:
		return normalizeBodyCompositionListRequest(normalized, request)
	default:
		return normalizedBodyCompositionTaskRequest{}, fmt.Sprintf("unsupported body composition task action %q", action)
	}
}

func normalizeBodyCompositionListRequest(normalized normalizedBodyCompositionTaskRequest, request BodyCompositionTaskRequest) (normalizedBodyCompositionTaskRequest, string) {
	list, rejection := normalizeTaskListRequest(taskListRequest{
		ListMode: request.ListMode,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
	}, "body composition")
	if rejection != "" {
		return normalizedBodyCompositionTaskRequest{}, rejection
	}
	normalized.ListMode = list.ListMode
	normalized.From = list.From
	normalized.To = list.To
	normalized.Limit = list.Limit
	return normalized, ""
}

func normalizeBodyCompositionInput(input BodyCompositionInput) (normalizedBodyCompositionInput, string) {
	recordedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedBodyCompositionInput{}, rejection
	}
	if input.BodyFatPercent == nil && input.WeightValue == nil {
		return normalizedBodyCompositionInput{}, "at least one body composition measurement is required"
	}
	if input.BodyFatPercent != nil && (*input.BodyFatPercent <= 0 || *input.BodyFatPercent > 100) {
		return normalizedBodyCompositionInput{}, "body_fat_percent must be greater than 0 and less than or equal to 100"
	}
	if (input.WeightValue == nil) != (input.WeightUnit == nil) {
		return normalizedBodyCompositionInput{}, "weight_value and weight_unit must be provided together"
	}
	var weightUnit *health.WeightUnit
	if input.WeightValue != nil {
		if *input.WeightValue <= 0 {
			return normalizedBodyCompositionInput{}, "weight_value must be greater than 0"
		}
		unit, rejection := normalizeUnit(*input.WeightUnit)
		if rejection != "" {
			return normalizedBodyCompositionInput{}, "weight_unit must be lb"
		}
		weightUnit = &unit
	}
	method, rejection := normalizeOptionalLabText(input.Method, "method")
	if rejection != "" {
		return normalizedBodyCompositionInput{}, rejection
	}
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedBodyCompositionInput{}, rejection
	}
	return normalizedBodyCompositionInput{
		RecordedAt:     recordedAt,
		BodyFatPercent: input.BodyFatPercent,
		WeightValue:    input.WeightValue,
		WeightUnit:     weightUnit,
		Method:         method,
		Note:           note,
	}, ""
}

func normalizeBodyCompositionTarget(target BodyCompositionTarget) (normalizedBodyCompositionTarget, string) {
	if target.ID < 0 {
		return normalizedBodyCompositionTarget{}, "target id must be greater than 0"
	}
	if target.ID > 0 {
		return normalizedBodyCompositionTarget{ID: target.ID}, ""
	}
	if target.Date == "" {
		return normalizedBodyCompositionTarget{}, "target id or date is required"
	}
	date, rejection := parseDateOnly(target.Date)
	if rejection != "" {
		return normalizedBodyCompositionTarget{}, rejection
	}
	return normalizedBodyCompositionTarget{Date: &date}, ""
}

func runBodyCompositionRecord(ctx context.Context, service health.Service, request normalizedBodyCompositionTaskRequest) (BodyCompositionTaskResult, error) {
	result := BodyCompositionTaskResult{}
	for _, record := range request.Records {
		existing, err := service.ListBodyComposition(ctx, health.HistoryFilter{})
		if err != nil {
			return BodyCompositionTaskResult{}, err
		}
		if duplicate, ok := matchingBodyComposition(existing, record); ok {
			result.Writes = append(result.Writes, bodyCompositionWrite(duplicate, "already_exists"))
			continue
		}
		written, err := service.CreateBodyComposition(ctx, health.BodyCompositionInput(record))
		if err != nil {
			return BodyCompositionTaskResult{}, err
		}
		result.Writes = append(result.Writes, bodyCompositionWrite(written, "created"))
	}
	entries, err := listBodyCompositionEntries(ctx, service, normalizedBodyCompositionTaskRequest{ListMode: BodyCompositionListModeHistory, Limit: 100})
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d body composition entries", len(entries))
	return result, nil
}

func runBodyCompositionCorrect(ctx context.Context, service health.Service, request normalizedBodyCompositionTaskRequest) (BodyCompositionTaskResult, error) {
	target, rejection, err := bodyCompositionTarget(ctx, service, request.Target)
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	if rejection != "" {
		return BodyCompositionTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	record := request.Record
	if record.Method == nil {
		record.Method = target.Method
	}
	if record.Note == nil {
		record.Note = target.Note
	}
	written, err := service.ReplaceBodyComposition(ctx, target.ID, health.BodyCompositionInput(record))
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	entries, err := listBodyCompositionEntries(ctx, service, normalizedBodyCompositionTaskRequest{ListMode: BodyCompositionListModeHistory, Limit: 100})
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	return BodyCompositionTaskResult{
		Writes:  []BodyCompositionWrite{bodyCompositionWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d body composition entries", len(entries)),
	}, nil
}

func runBodyCompositionDelete(ctx context.Context, service health.Service, request normalizedBodyCompositionTaskRequest) (BodyCompositionTaskResult, error) {
	target, rejection, err := bodyCompositionTarget(ctx, service, request.Target)
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	if rejection != "" {
		return BodyCompositionTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := service.DeleteBodyComposition(ctx, target.ID); err != nil {
		return BodyCompositionTaskResult{}, err
	}
	entries, err := listBodyCompositionEntries(ctx, service, normalizedBodyCompositionTaskRequest{ListMode: BodyCompositionListModeHistory, Limit: 100})
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	return BodyCompositionTaskResult{
		Writes:  []BodyCompositionWrite{bodyCompositionWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d body composition entries", len(entries)),
	}, nil
}

func runBodyCompositionList(ctx context.Context, service health.Service, request normalizedBodyCompositionTaskRequest) (BodyCompositionTaskResult, error) {
	entries, err := listBodyCompositionEntries(ctx, service, request)
	if err != nil {
		return BodyCompositionTaskResult{}, err
	}
	return BodyCompositionTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d body composition entries", len(entries)),
	}, nil
}

func bodyCompositionTarget(ctx context.Context, service health.Service, target normalizedBodyCompositionTarget) (health.BodyCompositionEntry, string, error) {
	return service.ResolveBodyCompositionTarget(ctx, health.BodyCompositionTarget{
		ID:         target.ID,
		RecordedAt: target.Date,
	})
}

func listBodyCompositionEntries(ctx context.Context, service health.Service, request normalizedBodyCompositionTaskRequest) ([]BodyCompositionTaskEntry, error) {
	items, err := service.ListBodyComposition(ctx, health.HistoryFilter{
		From:  request.From,
		To:    request.To,
		Limit: limitPointer(request.Limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]BodyCompositionTaskEntry, 0, len(items))
	for _, item := range items {
		out = append(out, bodyCompositionEntry(item))
	}
	return out, nil
}

func matchingBodyComposition(items []health.BodyCompositionEntry, input normalizedBodyCompositionInput) (health.BodyCompositionEntry, bool) {
	for _, item := range items {
		if item.RecordedAt.Format(time.DateOnly) != input.RecordedAt.Format(time.DateOnly) {
			continue
		}
		if equalFloatPointer(item.BodyFatPercent, input.BodyFatPercent) &&
			equalFloatPointer(item.WeightValue, input.WeightValue) &&
			equalWeightUnitPointer(item.WeightUnit, input.WeightUnit) &&
			equalStringPointer(item.Method, input.Method) &&
			equalStringPointer(item.Note, input.Note) {
			return item, true
		}
	}
	return health.BodyCompositionEntry{}, false
}

func bodyCompositionWrite(item health.BodyCompositionEntry, status string) BodyCompositionWrite {
	return BodyCompositionWrite{
		ID:             item.ID,
		Date:           item.RecordedAt.Format(time.DateOnly),
		BodyFatPercent: item.BodyFatPercent,
		WeightValue:    item.WeightValue,
		WeightUnit:     bodyCompositionWeightUnitString(item.WeightUnit),
		Status:         status,
	}
}

func bodyCompositionEntry(item health.BodyCompositionEntry) BodyCompositionTaskEntry {
	return BodyCompositionTaskEntry{
		ID:             item.ID,
		Date:           item.RecordedAt.Format(time.DateOnly),
		BodyFatPercent: item.BodyFatPercent,
		WeightValue:    item.WeightValue,
		WeightUnit:     bodyCompositionWeightUnitString(item.WeightUnit),
		Method:         item.Method,
		Note:           item.Note,
	}
}

func bodyCompositionWeightUnitString(unit *health.WeightUnit) *string {
	if unit == nil {
		return nil
	}
	value := string(*unit)
	return &value
}

func equalWeightUnitPointer(left *health.WeightUnit, right *health.WeightUnit) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
