package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	ImagingTaskActionRecord   = "record_imaging"
	ImagingTaskActionCorrect  = "correct_imaging"
	ImagingTaskActionDelete   = "delete_imaging"
	ImagingTaskActionList     = "list_imaging"
	ImagingTaskActionValidate = taskActionValidate

	ImagingListModeLatest  = listModeLatest
	ImagingListModeHistory = listModeHistory
	ImagingListModeRange   = listModeRange
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
	Date       string   `json:"date"`
	Modality   string   `json:"modality"`
	BodySite   *string  `json:"body_site,omitempty"`
	Title      *string  `json:"title,omitempty"`
	Summary    string   `json:"summary"`
	Impression *string  `json:"impression,omitempty"`
	Note       *string  `json:"note,omitempty"`
	Notes      []string `json:"notes,omitempty"`
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
	ID         int      `json:"id"`
	Date       string   `json:"date"`
	Modality   string   `json:"modality"`
	BodySite   *string  `json:"body_site,omitempty"`
	Title      *string  `json:"title,omitempty"`
	Summary    string   `json:"summary"`
	Impression *string  `json:"impression,omitempty"`
	Note       *string  `json:"note,omitempty"`
	Notes      []string `json:"notes,omitempty"`
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
	Notes       []string
}

type normalizedImagingTarget struct {
	ID   int
	Date *time.Time
}

func RunImagingTask(ctx context.Context, config localruntime.Config, request ImagingTaskRequest) (ImagingTaskResult, error) {
	normalized, rejection := normalizeImagingTaskRequest(request)
	if rejection != "" {
		return ImagingTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if normalized.Action == ImagingTaskActionValidate {
		return ImagingTaskResult{Summary: "valid"}, nil
	}

	return withService(ctx, config, func(ctx context.Context, service health.Service) (ImagingTaskResult, error) {
		switch normalized.Action {
		case ImagingTaskActionRecord:
			return runImagingRecord(ctx, service, normalized)
		case ImagingTaskActionCorrect:
			return runImagingCorrect(ctx, service, normalized)
		case ImagingTaskActionDelete:
			return runImagingDelete(ctx, service, normalized)
		case ImagingTaskActionList:
			return runImagingList(ctx, service, normalized)
		default:
			return ImagingTaskResult{}, fmt.Errorf("unsupported imaging task action %q", normalized.Action)
		}
	})
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
	if rejection := rejectNegativeLimit(request.Limit); rejection != "" {
		return normalizedImagingTaskRequest{}, rejection
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
	list, rejection := normalizeTaskListRequest(taskListRequest{
		ListMode: request.ListMode,
		FromDate: request.FromDate,
		ToDate:   request.ToDate,
		Limit:    request.Limit,
	}, "imaging")
	if rejection != "" {
		return normalizedImagingTaskRequest{}, rejection
	}
	normalized.ListMode = list.ListMode
	normalized.From = list.From
	normalized.To = list.To
	normalized.Limit = list.Limit
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
	notes, rejection := normalizeNoteList(input.Notes, "notes")
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
		Notes:       notes,
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

func runImagingRecord(ctx context.Context, service health.Service, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	result := ImagingTaskResult{}
	for _, record := range request.Records {
		existing, err := service.ListImaging(ctx, health.ImagingListParams{})
		if err != nil {
			return ImagingTaskResult{}, err
		}
		if duplicate, ok := matchingImagingRecord(existing, record); ok {
			result.Writes = append(result.Writes, imagingWrite(duplicate, "already_exists"))
			continue
		}
		written, err := service.CreateImaging(ctx, health.ImagingRecordInput(record))
		if err != nil {
			return ImagingTaskResult{}, err
		}
		result.Writes = append(result.Writes, imagingWrite(written, "created"))
	}
	entries, err := listImagingEntries(ctx, service, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d imaging entries", len(entries))
	return result, nil
}

func runImagingCorrect(ctx context.Context, service health.Service, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	target, rejection, err := imagingTarget(ctx, service, request.Target)
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
	if record.Notes == nil {
		record.Notes = append([]string(nil), target.Notes...)
	}
	written, err := service.ReplaceImaging(ctx, target.ID, health.ImagingRecordInput(record))
	if err != nil {
		return ImagingTaskResult{}, err
	}
	entries, err := listImagingEntries(ctx, service, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Writes:  []ImagingWrite{imagingWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d imaging entries", len(entries)),
	}, nil
}

func runImagingDelete(ctx context.Context, service health.Service, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	target, rejection, err := imagingTarget(ctx, service, request.Target)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	if rejection != "" {
		return ImagingTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := service.DeleteImaging(ctx, target.ID); err != nil {
		return ImagingTaskResult{}, err
	}
	entries, err := listImagingEntries(ctx, service, normalizedImagingTaskRequest{ListMode: ImagingListModeHistory, Limit: 100})
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Writes:  []ImagingWrite{imagingWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d imaging entries", len(entries)),
	}, nil
}

func runImagingList(ctx context.Context, service health.Service, request normalizedImagingTaskRequest) (ImagingTaskResult, error) {
	entries, err := listImagingEntries(ctx, service, request)
	if err != nil {
		return ImagingTaskResult{}, err
	}
	return ImagingTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d imaging entries", len(entries)),
	}, nil
}

func imagingTarget(ctx context.Context, service health.Service, target normalizedImagingTarget) (health.ImagingRecord, string, error) {
	return service.ResolveImagingTarget(ctx, health.ImagingTarget{
		ID:          target.ID,
		PerformedAt: target.Date,
	})
}

func listImagingEntries(ctx context.Context, service health.Service, request normalizedImagingTaskRequest) ([]ImagingTaskEntry, error) {
	items, err := service.ListImaging(ctx, health.ImagingListParams{
		HistoryFilter: health.HistoryFilter{
			From:  request.From,
			To:    request.To,
			Limit: limitPointer(request.Limit),
		},
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

func matchingImagingRecord(items []health.ImagingRecord, input normalizedImagingInput) (health.ImagingRecord, bool) {
	for _, item := range items {
		if item.PerformedAt.Format(time.DateOnly) != input.PerformedAt.Format(time.DateOnly) {
			continue
		}
		if strings.EqualFold(item.Modality, input.Modality) &&
			equalStringPointerFold(item.BodySite, input.BodySite) &&
			equalStringPointer(item.Title, input.Title) &&
			item.Summary == input.Summary &&
			equalStringPointer(item.Impression, input.Impression) &&
			equalStringPointer(item.Note, input.Note) &&
			equalStringSlices(item.Notes, input.Notes) {
			return item, true
		}
	}
	return health.ImagingRecord{}, false
}

func imagingWrite(item health.ImagingRecord, status string) ImagingWrite {
	return ImagingWrite{
		ID:       item.ID,
		Date:     item.PerformedAt.Format(time.DateOnly),
		Modality: item.Modality,
		Status:   status,
	}
}

func imagingEntry(item health.ImagingRecord) ImagingTaskEntry {
	return ImagingTaskEntry{
		ID:         item.ID,
		Date:       item.PerformedAt.Format(time.DateOnly),
		Modality:   item.Modality,
		BodySite:   item.BodySite,
		Title:      item.Title,
		Summary:    item.Summary,
		Impression: item.Impression,
		Note:       item.Note,
		Notes:      append([]string(nil), item.Notes...),
	}
}

func equalStringPointerFold(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return strings.EqualFold(*left, *right)
}
