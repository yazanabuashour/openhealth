package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	ImagingTaskActionRecord   = "record_imaging"
	ImagingTaskActionCorrect  = "correct_imaging"
	ImagingTaskActionDelete   = "delete_imaging"
	ImagingTaskActionList     = "list_imaging"
	ImagingTaskActionValidate = "validate"

	ImagingListModeLatest  = "latest"
	ImagingListModeHistory = "history"
	ImagingListModeRange   = "range"
)

type ImagingTaskRequest struct {
	Action   string         `json:"action"`
	Records  []ImagingInput `json:"records,omitempty"`
	Record   *ImagingInput  `json:"record,omitempty"`
	Target   *ImagingTarget `json:"target,omitempty"`
	ListMode string         `json:"list_mode,omitempty"`
	FromDate string         `json:"from_date,omitempty"`
	ToDate   string         `json:"to_date,omitempty"`
	Limit    int            `json:"limit,omitempty"`
	Modality string         `json:"modality,omitempty"`
	BodySite string         `json:"body_site,omitempty"`
}

type ImagingInput struct {
	Date       string  `json:"date"`
	Modality   string  `json:"modality"`
	BodySite   *string `json:"body_site,omitempty"`
	Title      *string `json:"title,omitempty"`
	Summary    string  `json:"summary"`
	Impression *string `json:"impression,omitempty"`
	Note       *string `json:"note,omitempty"`
}

type ImagingTarget struct {
	ID   int    `json:"id,omitempty"`
	Date string `json:"date,omitempty"`
}

type ImagingTaskResult struct {
	Rejected        bool               `json:"rejected"`
	RejectionReason string             `json:"rejection_reason,omitempty"`
	Writes          []ImagingWrite     `json:"writes,omitempty"`
	Entries         []ImagingTaskEntry `json:"entries,omitempty"`
	Summary         string             `json:"summary"`
}

type ImagingWrite struct {
	ID       int    `json:"id"`
	Date     string `json:"date"`
	Modality string `json:"modality"`
	Status   string `json:"status"`
}

type ImagingTaskEntry struct {
	ID         int     `json:"id"`
	Date       string  `json:"date"`
	Modality   string  `json:"modality"`
	BodySite   *string `json:"body_site,omitempty"`
	Title      *string `json:"title,omitempty"`
	Summary    string  `json:"summary"`
	Impression *string `json:"impression,omitempty"`
	Note       *string `json:"note,omitempty"`
}

type normalizedImagingTaskRequest struct {
	Action   string
	Records  []normalizedImagingInput
	Record   normalizedImagingInput
	Target   normalizedImagingTarget
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
	Modality *string
	BodySite *string
}

type normalizedImagingInput struct {
	PerformedAt time.Time
	Modality    string
	BodySite    *string
	Title       *string
	Summary     string
	Impression  *string
	Note        *string
}

type normalizedImagingTarget struct {
	ID   int
	Date *time.Time
}

func RunImagingTask(ctx context.Context, config client.LocalConfig, request ImagingTaskRequest) (ImagingTaskResult, error) {
	normalized, rejection := normalizeImagingTaskRequest(request)
	if rejection != "" {
		return ImagingTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if normalized.Action == ImagingTaskActionValidate {
		return ImagingTaskResult{Summary: "valid"}, nil
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case ImagingTaskActionRecord:
		return runImagingRecord(ctx, api, normalized)
	case ImagingTaskActionCorrect:
		return runImagingCorrect(ctx, api, normalized)
	case ImagingTaskActionDelete:
		return runImagingDelete(ctx, api, normalized)
	case ImagingTaskActionList:
		return runImagingList(ctx, api, normalized)
	default:
		return ImagingTaskResult{}, fmt.Errorf("unsupported imaging task action %q", normalized.Action)
	}
}

func normalizeImagingTaskRequest(request ImagingTaskRequest) (normalizedImagingTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = ImagingTaskActionValidate
	}
	normalized := normalizedImagingTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}
	if request.Limit < 0 {
		return normalizedImagingTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case ImagingTaskActionValidate:
		for _, record := range request.Records {
			if _, rejection := normalizeImagingInput(record); rejection != "" {
				return normalizedImagingTaskRequest{}, rejection
			}
		}
		if request.Record != nil {
			if _, rejection := normalizeImagingInput(*request.Record); rejection != "" {
				return normalizedImagingTaskRequest{}, rejection
			}
		}
		if request.Target != nil {
			if _, rejection := normalizeImagingTarget(*request.Target); rejection != "" {
				return normalizedImagingTaskRequest{}, rejection
			}
		}
		return normalized, ""
	case ImagingTaskActionRecord:
		if len(request.Records) == 0 {
			return normalizedImagingTaskRequest{}, "records are required"
		}
		for _, record := range request.Records {
			parsed, rejection := normalizeImagingInput(record)
			if rejection != "" {
				return normalizedImagingTaskRequest{}, rejection
			}
			normalized.Records = append(normalized.Records, parsed)
		}
		return normalized, ""
	case ImagingTaskActionCorrect:
		if request.Target == nil {
			return normalizedImagingTaskRequest{}, "target is required"
		}
		target, rejection := normalizeImagingTarget(*request.Target)
		if rejection != "" {
			return normalizedImagingTaskRequest{}, rejection
		}
		if request.Record == nil {
			return normalizedImagingTaskRequest{}, "record is required"
		}
		record, rejection := normalizeImagingInput(*request.Record)
		if rejection != "" {
			return normalizedImagingTaskRequest{}, rejection
		}
		normalized.Target = target
		normalized.Record = record
		return normalized, ""
	case ImagingTaskActionDelete:
		if request.Target == nil {
			return normalizedImagingTaskRequest{}, "target is required"
		}
		target, rejection := normalizeImagingTarget(*request.Target)
		if rejection != "" {
			return normalizedImagingTaskRequest{}, rejection
		}
		normalized.Target = target
		return normalized, ""
	case ImagingTaskActionList:
		return normalizeImagingListRequest(normalized, request)
	default:
		return normalizedImagingTaskRequest{}, fmt.Sprintf("unsupported imaging task action %q", action)
	}
}

func normalizeImagingListRequest(normalized normalizedImagingTaskRequest, request ImagingTaskRequest) (normalizedImagingTaskRequest, string) {
	if normalized.ListMode == "" {
		normalized.ListMode = ImagingListModeHistory
	}
	switch normalized.ListMode {
	case ImagingListModeLatest:
		normalized.Limit = 1
	case ImagingListModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case ImagingListModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedImagingTaskRequest{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedImagingTaskRequest{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedImagingTaskRequest{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedImagingTaskRequest{}, fmt.Sprintf("unsupported imaging list mode %q", normalized.ListMode)
	}
	modality, rejection := normalizeOptionalFilter(request.Modality, "modality")
	if rejection != "" {
		return normalizedImagingTaskRequest{}, rejection
	}
	bodySite, rejection := normalizeOptionalFilter(request.BodySite, "body_site")
	if rejection != "" {
		return normalizedImagingTaskRequest{}, rejection
	}
	normalized.Modality = modality
	normalized.BodySite = bodySite
	return normalized, ""
}

func normalizeImagingInput(input ImagingInput) (normalizedImagingInput, string) {
	performedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedImagingInput{}, rejection
	}
	modality := strings.TrimSpace(input.Modality)
	if modality == "" {
		return normalizedImagingInput{}, "modality is required"
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return normalizedImagingInput{}, "summary is required"
	}
	bodySite, rejection := normalizeOptionalLabText(input.BodySite, "body_site")
	if rejection != "" {
		return normalizedImagingInput{}, rejection
	}
	title, rejection := normalizeOptionalLabText(input.Title, "title")
	if rejection != "" {
		return normalizedImagingInput{}, rejection
	}
	impression, rejection := normalizeOptionalLabText(input.Impression, "impression")
	if rejection != "" {
		return normalizedImagingInput{}, rejection
	}
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedImagingInput{}, rejection
	}
	return normalizedImagingInput{
		PerformedAt: performedAt,
		Modality:    modality,
		BodySite:    bodySite,
		Title:       title,
		Summary:     summary,
		Impression:  impression,
		Note:        note,
	}, ""
}

func normalizeImagingTarget(target ImagingTarget) (normalizedImagingTarget, string) {
	if target.ID < 0 {
		return normalizedImagingTarget{}, "target id must be greater than 0"
	}
	if target.ID > 0 {
		return normalizedImagingTarget{ID: target.ID}, ""
	}
	if target.Date == "" {
		return normalizedImagingTarget{}, "target id or date is required"
	}
	date, rejection := parseDateOnly(target.Date)
	if rejection != "" {
		return normalizedImagingTarget{}, rejection
	}
	return normalizedImagingTarget{Date: &date}, ""
}

func normalizeOptionalFilter(value string, field string) (*string, string) {
	if value == "" {
		return nil, ""
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, field + " must not be empty"
	}
	return &trimmed, ""
}

func runImagingRecord(ctx context.Context, api *client.LocalClient, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	result := ImagingTaskResult{}
	for _, record := range request.Records {
		existing, err := api.ListImaging(ctx, client.ImagingListOptions{})
		if err != nil {
			return ImagingTaskResult{}, err
		}
		if duplicate, ok := matchingImagingRecord(existing, record); ok {
			result.Writes = append(result.Writes, imagingWrite(duplicate, "already_exists"))
			continue
		}
		written, err := api.CreateImaging(ctx, client.ImagingRecordInput(record))
		if err != nil {
			return ImagingTaskResult{}, err
		}
		result.Writes = append(result.Writes, imagingWrite(written, "created"))
	}
	entries, err := listImagingEntries(ctx, api, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d imaging entries", len(entries))
	return result, nil
}

func runImagingCorrect(ctx context.Context, api *client.LocalClient, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	target, rejection, err := imagingTarget(ctx, api, request.Target)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	if rejection != "" {
		return ImagingTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	record := request.Record
	if record.BodySite == nil {
		record.BodySite = target.BodySite
	}
	if record.Title == nil {
		record.Title = target.Title
	}
	if record.Impression == nil {
		record.Impression = target.Impression
	}
	if record.Note == nil {
		record.Note = target.Note
	}
	written, err := api.ReplaceImaging(ctx, target.ID, client.ImagingRecordInput(record))
	if err != nil {
		return ImagingTaskResult{}, err
	}
	entries, err := listImagingEntries(ctx, api, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Writes:  []ImagingWrite{imagingWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d imaging entries", len(entries)),
	}, nil
}

func runImagingDelete(ctx context.Context, api *client.LocalClient, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	target, rejection, err := imagingTarget(ctx, api, request.Target)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	if rejection != "" {
		return ImagingTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := api.DeleteImaging(ctx, target.ID); err != nil {
		return ImagingTaskResult{}, err
	}
	entries, err := listImagingEntries(ctx, api, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Writes:  []ImagingWrite{imagingWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d imaging entries", len(entries)),
	}, nil
}

func runImagingList(ctx context.Context, api *client.LocalClient, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	entries, err := listImagingEntries(ctx, api, request)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d imaging entries", len(entries)),
	}, nil
}

func imagingTarget(ctx context.Context, api *client.LocalClient, target normalizedImagingTarget) (client.ImagingRecord, string, error) {
	items, err := api.ListImaging(ctx, client.ImagingListOptions{})
	if err != nil {
		return client.ImagingRecord{}, "", err
	}
	matches := []client.ImagingRecord{}
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if target.Date != nil && item.PerformedAt.Format(time.DateOnly) == target.Date.Format(time.DateOnly) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return client.ImagingRecord{}, "no matching imaging record", nil
	case 1:
		return matches[0], "", nil
	default:
		return client.ImagingRecord{}, "multiple matching imaging records; target is ambiguous", nil
	}
}

func listImagingEntries(ctx context.Context, api *client.LocalClient, request normalizedImagingTaskRequest) ([]ImagingTaskEntry, error) {
	items, err := api.ListImaging(ctx, client.ImagingListOptions{
		From:     request.From,
		To:       request.To,
		Limit:    request.Limit,
		Modality: request.Modality,
		BodySite: request.BodySite,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ImagingTaskEntry, 0, len(items))
	for _, item := range items {
		out = append(out, imagingEntry(item))
	}
	return out, nil
}

func matchingImagingRecord(items []client.ImagingRecord, input normalizedImagingInput) (client.ImagingRecord, bool) {
	for _, item := range items {
		if item.PerformedAt.Format(time.DateOnly) != input.PerformedAt.Format(time.DateOnly) {
			continue
		}
		if strings.EqualFold(item.Modality, input.Modality) &&
			equalStringPointerFold(item.BodySite, input.BodySite) &&
			equalStringPointer(item.Title, input.Title) &&
			item.Summary == input.Summary &&
			equalStringPointer(item.Impression, input.Impression) &&
			equalStringPointer(item.Note, input.Note) {
			return item, true
		}
	}
	return client.ImagingRecord{}, false
}

func imagingWrite(item client.ImagingRecord, status string) ImagingWrite {
	return ImagingWrite{
		ID:       item.ID,
		Date:     item.PerformedAt.Format(time.DateOnly),
		Modality: item.Modality,
		Status:   status,
	}
}

func imagingEntry(item client.ImagingRecord) ImagingTaskEntry {
	return ImagingTaskEntry{
		ID:         item.ID,
		Date:       item.PerformedAt.Format(time.DateOnly),
		Modality:   item.Modality,
		BodySite:   item.BodySite,
		Title:      item.Title,
		Summary:    item.Summary,
		Impression: item.Impression,
		Note:       item.Note,
	}
}

func equalStringPointerFold(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return strings.EqualFold(*left, *right)
}
