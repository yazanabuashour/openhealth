package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	SleepTaskActionUpsert   = "upsert_sleep"
	SleepTaskActionDelete   = "delete_sleep"
	SleepTaskActionList     = "list_sleep"
	SleepTaskActionValidate = taskActionValidate

	SleepListModeLatest  = listModeLatest
	SleepListModeHistory = listModeHistory
	SleepListModeRange   = listModeRange
)

type SleepTaskRequest struct {
	Action   string       `json:"action"`
	Entries  []SleepInput `json:"entries,omitempty"`
	Target   *SleepTarget `json:"target,omitempty"`
	ListMode string       `json:"list_mode,omitempty"`
	FromDate string       `json:"from_date,omitempty"`
	ToDate   string       `json:"to_date,omitempty"`
	Limit    int          `json:"limit,omitempty"`
}

type SleepInput struct {
	Date         string  `json:"date"`
	QualityScore int     `json:"quality_score"`
	WakeupCount  *int    `json:"wakeup_count,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type SleepTarget struct {
	ID   int    `json:"id,omitempty"`
	Date string `json:"date,omitempty"`
}

type SleepTaskResult struct {
	Rejected        bool             `json:"rejected"`
	RejectionReason string           `json:"rejection_reason,omitempty"`
	Writes          []SleepWrite     `json:"writes,omitempty"`
	Entries         []SleepTaskEntry `json:"entries,omitempty"`
	Summary         string           `json:"summary"`
}

type SleepWrite struct {
	ID           int    `json:"id"`
	Date         string `json:"date"`
	QualityScore int    `json:"quality_score"`
	WakeupCount  *int   `json:"wakeup_count,omitempty"`
	Status       string `json:"status"`
}

type SleepTaskEntry struct {
	ID           int     `json:"id"`
	Date         string  `json:"date"`
	QualityScore int     `json:"quality_score"`
	WakeupCount  *int    `json:"wakeup_count,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type normalizedSleepTaskRequest struct {
	Action   string
	Entries  []normalizedSleepInput
	Target   normalizedSleepTarget
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
}

type normalizedSleepInput struct {
	RecordedAt   time.Time
	QualityScore int
	WakeupCount  *int
	Note         *string
}

type normalizedSleepTarget struct {
	ID   int
	Date *time.Time
}

func RunSleepTask(ctx context.Context, config localruntime.Config, request SleepTaskRequest) (SleepTaskResult, error) {
	normalized, rejection := normalizeSleepTaskRequest(request)
	if rejection != "" {
		return SleepTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if normalized.Action == SleepTaskActionValidate {
		return SleepTaskResult{Summary: "valid"}, nil
	}

	return withService(ctx, config, func(ctx context.Context, service health.Service) (SleepTaskResult, error) {
		switch normalized.Action {
		case SleepTaskActionUpsert:
			return runSleepUpsert(ctx, service, normalized)
		case SleepTaskActionDelete:
			return runSleepDelete(ctx, service, normalized)
		case SleepTaskActionList:
			return runSleepList(ctx, service, normalized)
		default:
			return SleepTaskResult{}, fmt.Errorf("unsupported sleep task action %q", normalized.Action)
		}
	})
}

func normalizeSleepTaskRequest(request SleepTaskRequest) (normalizedSleepTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = SleepTaskActionValidate
	}
	normalized := normalizedSleepTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}
	if rejection := rejectNegativeLimit(request.Limit); rejection != "" {
		return normalizedSleepTaskRequest{}, rejection
	}

	switch action {
	case SleepTaskActionValidate:
		for _, entry := range request.Entries {
			if _, rejection := normalizeSleepInput(entry); rejection != "" {
				return normalizedSleepTaskRequest{}, rejection
			}
		}
		if request.Target != nil {
			if _, rejection := normalizeSleepTarget(*request.Target); rejection != "" {
				return normalizedSleepTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case SleepTaskActionUpsert:
		if len(request.Entries) == 0 {
			return normalizedSleepTaskRequest{}, "entries are required"
		}
		for _, entry := range request.Entries {
			parsed, rejection := normalizeSleepInput(entry)
			if rejection != "" {
				return normalizedSleepTaskRequest{}, rejection
			}
			normalized.Entries = append(normalized.Entries, parsed)
		}
		return normalized, ""
	case SleepTaskActionDelete:
		if request.Target == nil {
			return normalizedSleepTaskRequest{}, "target is required"
		}
		target, rejection := normalizeSleepTarget(*request.Target)
		if rejection != "" {
			return normalizedSleepTaskRequest{}, rejection
		}
		normalized.Target = target
		return normalized, ""
	case SleepTaskActionList:
		return normalizeSleepListRequest(normalized, request)
	default:
		return normalizedSleepTaskRequest{}, fmt.Sprintf("unsupported sleep task action %q", action)
	}
}

func normalizeSleepListRequest(normalized normalizedSleepTaskRequest, request SleepTaskRequest) (normalizedSleepTaskRequest, string) {
	list, rejection := normalizeTaskListRequest(taskListRequest{
		ListMode: request.ListMode,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
	}, "sleep")
	if rejection != "" {
		return normalizedSleepTaskRequest{}, rejection
	}
	normalized.ListMode = list.ListMode
	normalized.From = list.From
	normalized.To = list.To
	normalized.Limit = list.Limit
	return normalized, ""
}

func normalizeSleepInput(input SleepInput) (normalizedSleepInput, string) {
	recordedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedSleepInput{}, rejection
	}
	if input.QualityScore < 1 || input.QualityScore > 5 {
		return normalizedSleepInput{}, "quality_score must be between 1 and 5"
	}
	if input.WakeupCount != nil && *input.WakeupCount < 0 {
		return normalizedSleepInput{}, "wakeup_count must be greater than or equal to 0"
	}
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedSleepInput{}, rejection
	}
	return normalizedSleepInput{
		RecordedAt:   recordedAt,
		QualityScore: input.QualityScore,
		WakeupCount:  input.WakeupCount,
		Note:         note,
	}, ""
}

func normalizeSleepTarget(target SleepTarget) (normalizedSleepTarget, string) {
	if target.ID < 0 {
		return normalizedSleepTarget{}, "target id must be greater than 0"
	}
	if target.ID > 0 {
		return normalizedSleepTarget{ID: target.ID}, ""
	}
	if target.Date == "" {
		return normalizedSleepTarget{}, "target id or date is required"
	}
	date, rejection := parseDateOnly(target.Date)
	if rejection != "" {
		return normalizedSleepTarget{}, rejection
	}
	return normalizedSleepTarget{Date: &date}, ""
}

func runSleepUpsert(ctx context.Context, service health.Service, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	result := SleepTaskResult{}
	for _, entry := range request.Entries {
		written, err := service.UpsertSleep(ctx, health.SleepInput(entry))
		if err != nil {
			return SleepTaskResult{}, err
		}
		result.Writes = append(result.Writes, sleepWrite(written.Entry, string(written.Status)))
	}
	entries, err := listSleepEntries(ctx, service, normalizedSleepTaskRequest{ListMode: SleepListModeHistory, Limit: 100})
	if err != nil {
		return SleepTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d sleep entries", len(entries))
	return result, nil
}

func runSleepDelete(ctx context.Context, service health.Service, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	target, rejection, err := sleepTarget(ctx, service, request.Target)
	if err != nil {
		return SleepTaskResult{}, err
	}
	if rejection != "" {
		return SleepTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := service.DeleteSleep(ctx, target.ID); err != nil {
		return SleepTaskResult{}, err
	}
	entries, err := listSleepEntries(ctx, service, normalizedSleepTaskRequest{ListMode: SleepListModeHistory, Limit: 100})
	if err != nil {
		return SleepTaskResult{}, err
	}
	return SleepTaskResult{
		Writes:  []SleepWrite{sleepWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d sleep entries", len(entries)),
	}, nil
}

func runSleepList(ctx context.Context, service health.Service, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	entries, err := listSleepEntries(ctx, service, request)
	if err != nil {
		return SleepTaskResult{}, err
	}
	return SleepTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d sleep entries", len(entries)),
	}, nil
}

func sleepTarget(ctx context.Context, service health.Service, target normalizedSleepTarget) (health.SleepEntry, string, error) {
	return service.ResolveSleepTarget(ctx, health.SleepTarget{
		ID:         target.ID,
		RecordedAt: target.Date,
	})
}

func listSleepEntries(ctx context.Context, service health.Service, request normalizedSleepTaskRequest) ([]SleepTaskEntry, error) {
	items, err := service.ListSleep(ctx, health.HistoryFilter{
		From:  request.From,
		To:    request.To,
		Limit: limitPointer(request.Limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]SleepTaskEntry, 0, len(items))
	for _, item := range items {
		out = append(out, sleepEntry(item))
	}
	return out, nil
}

func sleepWrite(item health.SleepEntry, status string) SleepWrite {
	return SleepWrite{
		ID:           item.ID,
		Date:         item.RecordedAt.Format(time.DateOnly),
		QualityScore: item.QualityScore,
		WakeupCount:  item.WakeupCount,
		Status:       status,
	}
}

func sleepEntry(item health.SleepEntry) SleepTaskEntry {
	return SleepTaskEntry{
		ID:           item.ID,
		Date:         item.RecordedAt.Format(time.DateOnly),
		QualityScore: item.QualityScore,
		WakeupCount:  item.WakeupCount,
		Note:         item.Note,
	}
}
