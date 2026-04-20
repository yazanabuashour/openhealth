package runner

import (
	"context"
	"fmt"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	SleepTaskActionUpsert   = "upsert_sleep"
	SleepTaskActionDelete   = "delete_sleep"
	SleepTaskActionList     = "list_sleep"
	SleepTaskActionValidate = "validate"

	SleepListModeLatest  = "latest"
	SleepListModeHistory = "history"
	SleepListModeRange   = "range"
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

func RunSleepTask(ctx context.Context, config client.LocalConfig, request SleepTaskRequest) (SleepTaskResult, error) {
	normalized, rejection := normalizeSleepTaskRequest(request)
	if rejection != "" {
		return SleepTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if normalized.Action == SleepTaskActionValidate {
		return SleepTaskResult{Summary: "valid"}, nil
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		return SleepTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case SleepTaskActionUpsert:
		return runSleepUpsert(ctx, api, normalized)
	case SleepTaskActionDelete:
		return runSleepDelete(ctx, api, normalized)
	case SleepTaskActionList:
		return runSleepList(ctx, api, normalized)
	default:
		return SleepTaskResult{}, fmt.Errorf("unsupported sleep task action %q", normalized.Action)
	}
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
	if request.Limit < 0 {
		return normalizedSleepTaskRequest{}, "limit must be greater than or equal to 0"
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
	if normalized.ListMode == "" {
		normalized.ListMode = SleepListModeHistory
	}
	switch normalized.ListMode {
	case SleepListModeLatest:
		normalized.Limit = 1
	case SleepListModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case SleepListModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedSleepTaskRequest{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedSleepTaskRequest{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedSleepTaskRequest{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedSleepTaskRequest{}, fmt.Sprintf("unsupported sleep list mode %q", normalized.ListMode)
	}
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

func runSleepUpsert(ctx context.Context, api *client.LocalClient, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	result := SleepTaskResult{}
	for _, entry := range request.Entries {
		written, err := api.UpsertSleep(ctx, client.SleepInput(entry))
		if err != nil {
			return SleepTaskResult{}, err
		}
		result.Writes = append(result.Writes, sleepWrite(written.Entry, string(written.Status)))
	}
	entries, err := listSleepEntries(ctx, api, normalizedSleepTaskRequest{ListMode: SleepListModeHistory, Limit: 100})
	if err != nil {
		return SleepTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d sleep entries", len(entries))
	return result, nil
}

func runSleepDelete(ctx context.Context, api *client.LocalClient, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	target, rejection, err := sleepTarget(ctx, api, request.Target)
	if err != nil {
		return SleepTaskResult{}, err
	}
	if rejection != "" {
		return SleepTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := api.DeleteSleep(ctx, target.ID); err != nil {
		return SleepTaskResult{}, err
	}
	entries, err := listSleepEntries(ctx, api, normalizedSleepTaskRequest{ListMode: SleepListModeHistory, Limit: 100})
	if err != nil {
		return SleepTaskResult{}, err
	}
	return SleepTaskResult{
		Writes:  []SleepWrite{sleepWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d sleep entries", len(entries)),
	}, nil
}

func runSleepList(ctx context.Context, api *client.LocalClient, request normalizedSleepTaskRequest) (SleepTaskResult, error) {
	entries, err := listSleepEntries(ctx, api, request)
	if err != nil {
		return SleepTaskResult{}, err
	}
	return SleepTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d sleep entries", len(entries)),
	}, nil
}

func sleepTarget(ctx context.Context, api *client.LocalClient, target normalizedSleepTarget) (client.SleepEntry, string, error) {
	items, err := api.ListSleep(ctx, client.SleepListOptions{})
	if err != nil {
		return client.SleepEntry{}, "", err
	}
	matches := []client.SleepEntry{}
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if target.Date != nil && item.RecordedAt.Format(time.DateOnly) == target.Date.Format(time.DateOnly) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return client.SleepEntry{}, "no matching sleep entry", nil
	case 1:
		return matches[0], "", nil
	default:
		return client.SleepEntry{}, "multiple matching sleep entries; target is ambiguous", nil
	}
}

func listSleepEntries(ctx context.Context, api *client.LocalClient, request normalizedSleepTaskRequest) ([]SleepTaskEntry, error) {
	items, err := api.ListSleep(ctx, client.SleepListOptions{
		From:  request.From,
		To:    request.To,
		Limit: request.Limit,
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

func sleepWrite(item client.SleepEntry, status string) SleepWrite {
	return SleepWrite{
		ID:           item.ID,
		Date:         item.RecordedAt.Format(time.DateOnly),
		QualityScore: item.QualityScore,
		WakeupCount:  item.WakeupCount,
		Status:       status,
	}
}

func sleepEntry(item client.SleepEntry) SleepTaskEntry {
	return SleepTaskEntry{
		ID:           item.ID,
		Date:         item.RecordedAt.Format(time.DateOnly),
		QualityScore: item.QualityScore,
		WakeupCount:  item.WakeupCount,
		Note:         item.Note,
	}
}
