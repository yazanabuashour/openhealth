package runner

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	LabTaskActionRecord   = "record_labs"
	LabTaskActionCorrect  = "correct_labs"
	LabTaskActionPatch    = "patch_labs"
	LabTaskActionDelete   = "delete_labs"
	LabTaskActionList     = "list_labs"
	LabTaskActionValidate = "validate"

	LabListModeLatest  = "latest"
	LabListModeHistory = "history"
	LabListModeRange   = "range"
)

type LabTaskRequest struct {
	Action        string                 `json:"action"`
	Collections   []LabCollectionInput   `json:"collections,omitempty"`
	Collection    *LabCollectionInput    `json:"collection,omitempty"`
	ResultUpdates []LabResultUpdateInput `json:"result_updates,omitempty"`
	Target        *LabTarget             `json:"target,omitempty"`
	ListMode      string                 `json:"list_mode,omitempty"`
	FromDate      string                 `json:"from_date,omitempty"`
	ToDate        string                 `json:"to_date,omitempty"`
	Limit         int                    `json:"limit,omitempty"`
	AnalyteSlug   string                 `json:"analyte_slug,omitempty"`
}

type LabCollectionInput struct {
	Date   string          `json:"date"`
	Note   *string         `json:"note,omitempty"`
	Panels []LabPanelInput `json:"panels"`
}

type LabPanelInput struct {
	PanelName string           `json:"panel_name"`
	Results   []LabResultInput `json:"results"`
}

type LabResultInput struct {
	TestName      string   `json:"test_name"`
	CanonicalSlug *string  `json:"canonical_slug,omitempty"`
	ValueText     string   `json:"value_text"`
	ValueNumeric  *float64 `json:"value_numeric,omitempty"`
	Units         *string  `json:"units,omitempty"`
	RangeText     *string  `json:"range_text,omitempty"`
	Flag          *string  `json:"flag,omitempty"`
}

type LabResultUpdateInput struct {
	PanelName string              `json:"panel_name"`
	Match     LabResultMatchInput `json:"match"`
	Result    LabResultInput      `json:"result"`
}

type LabResultMatchInput struct {
	CanonicalSlug string `json:"canonical_slug,omitempty"`
	TestName      string `json:"test_name,omitempty"`
}

type LabTarget struct {
	ID   int    `json:"id,omitempty"`
	Date string `json:"date,omitempty"`
}

type LabTaskResult struct {
	Rejected        bool                 `json:"rejected"`
	RejectionReason string               `json:"rejection_reason,omitempty"`
	Writes          []LabCollectionWrite `json:"writes,omitempty"`
	Entries         []LabCollectionEntry `json:"entries,omitempty"`
	Summary         string               `json:"summary"`
}

type LabCollectionWrite struct {
	ID     int    `json:"id"`
	Date   string `json:"date"`
	Status string `json:"status"`
}

type LabCollectionEntry struct {
	ID     int             `json:"id"`
	Date   string          `json:"date"`
	Note   *string         `json:"note,omitempty"`
	Panels []LabPanelEntry `json:"panels"`
}

type LabPanelEntry struct {
	PanelName string           `json:"panel_name"`
	Results   []LabResultEntry `json:"results"`
}

type LabResultEntry struct {
	TestName      string   `json:"test_name"`
	CanonicalSlug *string  `json:"canonical_slug,omitempty"`
	ValueText     string   `json:"value_text"`
	ValueNumeric  *float64 `json:"value_numeric,omitempty"`
	Units         *string  `json:"units,omitempty"`
	RangeText     *string  `json:"range_text,omitempty"`
	Flag          *string  `json:"flag,omitempty"`
}

func RunLabTask(ctx context.Context, config client.LocalConfig, request LabTaskRequest) (LabTaskResult, error) {
	normalized, rejection := normalizeLabTaskRequest(request)
	if rejection != "" {
		return LabTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == LabTaskActionValidate {
		return LabTaskResult{Summary: "valid"}, nil
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		return LabTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case LabTaskActionRecord:
		return runLabRecord(ctx, api, normalized)
	case LabTaskActionCorrect:
		return runLabCorrect(ctx, api, normalized)
	case LabTaskActionPatch:
		return runLabPatch(ctx, api, normalized)
	case LabTaskActionDelete:
		return runLabDelete(ctx, api, normalized)
	case LabTaskActionList:
		return runLabList(ctx, api, normalized)
	default:
		return LabTaskResult{}, fmt.Errorf("unsupported lab task action %q", normalized.Action)
	}
}

type normalizedLabTaskRequest struct {
	Action        string
	Collections   []normalizedLabCollectionInput
	Collection    normalizedLabCollectionInput
	ResultUpdates []normalizedLabResultUpdateInput
	Target        normalizedLabTarget
	ListMode      string
	From          *time.Time
	To            *time.Time
	Limit         int
	AnalyteSlug   *client.AnalyteSlug
}

type normalizedLabCollectionInput struct {
	CollectedAt time.Time
	Note        *string
	Panels      []normalizedLabPanelInput
}

type normalizedLabPanelInput struct {
	PanelName string
	Results   []normalizedLabResultInput
}

type normalizedLabResultInput struct {
	TestName      string
	CanonicalSlug *client.AnalyteSlug
	ValueText     string
	ValueNumeric  *float64
	Units         *string
	RangeText     *string
	Flag          *string
}

type normalizedLabResultUpdateInput struct {
	PanelName          string
	MatchCanonicalSlug *client.AnalyteSlug
	MatchTestName      string
	Result             normalizedLabResultInput
}

type normalizedLabTarget struct {
	ID   int
	Date *time.Time
}

func normalizeLabTaskRequest(request LabTaskRequest) (normalizedLabTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = LabTaskActionValidate
	}

	normalized := normalizedLabTaskRequest{
		Action:   action,
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}
	if request.Limit < 0 {
		return normalizedLabTaskRequest{}, "limit must be greater than or equal to 0"
	}

	switch action {
	case LabTaskActionValidate:
		for _, collection := range request.Collections {
			if _, rejection := normalizeLabCollectionInput(collection); rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
		}
		if request.Collection != nil {
			if _, rejection := normalizeLabCollectionInput(*request.Collection); rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
		}
		for _, update := range request.ResultUpdates {
			if _, rejection := normalizeLabResultUpdateInput(update); rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
		}
		if request.Target != nil {
			if _, rejection := normalizeLabTarget(*request.Target); rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
		}
		if _, rejection := normalizeAnalyteSlug(request.AnalyteSlug); rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		return normalized, ""
	case LabTaskActionRecord:
		if len(request.Collections) == 0 {
			return normalizedLabTaskRequest{}, "collections are required"
		}
		for _, collection := range request.Collections {
			parsed, rejection := normalizeLabCollectionInput(collection)
			if rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
			normalized.Collections = append(normalized.Collections, parsed)
		}
		return normalized, ""
	case LabTaskActionCorrect:
		if request.Target == nil {
			return normalizedLabTaskRequest{}, "target is required"
		}
		target, rejection := normalizeLabTarget(*request.Target)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		if request.Collection == nil {
			return normalizedLabTaskRequest{}, "collection is required"
		}
		collection, rejection := normalizeLabCollectionInput(*request.Collection)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		normalized.Target = target
		normalized.Collection = collection
		return normalized, ""
	case LabTaskActionPatch:
		if request.Target == nil {
			return normalizedLabTaskRequest{}, "target is required"
		}
		target, rejection := normalizeLabTarget(*request.Target)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		if len(request.ResultUpdates) == 0 {
			return normalizedLabTaskRequest{}, "result_updates are required"
		}
		for _, update := range request.ResultUpdates {
			parsed, rejection := normalizeLabResultUpdateInput(update)
			if rejection != "" {
				return normalizedLabTaskRequest{}, rejection
			}
			normalized.ResultUpdates = append(normalized.ResultUpdates, parsed)
		}
		normalized.Target = target
		return normalized, ""
	case LabTaskActionDelete:
		if request.Target == nil {
			return normalizedLabTaskRequest{}, "target is required"
		}
		target, rejection := normalizeLabTarget(*request.Target)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		normalized.Target = target
		return normalized, ""
	case LabTaskActionList:
		return normalizeLabListRequest(normalized, request)
	default:
		return normalizedLabTaskRequest{}, fmt.Sprintf("unsupported lab task action %q", action)
	}
}

func normalizeLabListRequest(normalized normalizedLabTaskRequest, request LabTaskRequest) (normalizedLabTaskRequest, string) {
	if normalized.ListMode == "" {
		normalized.ListMode = LabListModeHistory
	}
	switch normalized.ListMode {
	case LabListModeLatest:
		normalized.Limit = 1
	case LabListModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case LabListModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedLabTaskRequest{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedLabTaskRequest{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedLabTaskRequest{}, fmt.Sprintf("unsupported lab list mode %q", normalized.ListMode)
	}
	slug, rejection := normalizeAnalyteSlug(request.AnalyteSlug)
	if rejection != "" {
		return normalizedLabTaskRequest{}, rejection
	}
	normalized.AnalyteSlug = slug
	return normalized, ""
}

func normalizeLabCollectionInput(input LabCollectionInput) (normalizedLabCollectionInput, string) {
	collectedAt, rejection := parseDateOnly(input.Date)
	if rejection != "" {
		return normalizedLabCollectionInput{}, rejection
	}
	if len(input.Panels) == 0 {
		return normalizedLabCollectionInput{}, "at least one lab panel is required"
	}
	note, rejection := normalizeOptionalLabText(input.Note, "note")
	if rejection != "" {
		return normalizedLabCollectionInput{}, rejection
	}
	normalized := normalizedLabCollectionInput{CollectedAt: collectedAt, Note: note}
	for _, panel := range input.Panels {
		parsed, rejection := normalizeLabPanelInput(panel)
		if rejection != "" {
			return normalizedLabCollectionInput{}, rejection
		}
		normalized.Panels = append(normalized.Panels, parsed)
	}
	return normalized, ""
}

func normalizeLabPanelInput(input LabPanelInput) (normalizedLabPanelInput, string) {
	panelName := strings.TrimSpace(input.PanelName)
	if panelName == "" {
		return normalizedLabPanelInput{}, "panel_name is required"
	}
	if len(input.Results) == 0 {
		return normalizedLabPanelInput{}, "at least one lab result is required"
	}
	normalized := normalizedLabPanelInput{PanelName: panelName}
	for _, result := range input.Results {
		parsed, rejection := normalizeLabResultInput(result)
		if rejection != "" {
			return normalizedLabPanelInput{}, rejection
		}
		normalized.Results = append(normalized.Results, parsed)
	}
	return normalized, ""
}

func normalizeLabResultInput(input LabResultInput) (normalizedLabResultInput, string) {
	testName := strings.TrimSpace(input.TestName)
	if testName == "" {
		return normalizedLabResultInput{}, "test_name is required"
	}
	valueText := strings.TrimSpace(input.ValueText)
	if valueText == "" {
		return normalizedLabResultInput{}, "value_text is required"
	}
	slug, rejection := normalizeOptionalAnalyteSlug(input.CanonicalSlug)
	if rejection != "" {
		return normalizedLabResultInput{}, rejection
	}
	units, rejection := normalizeOptionalLabText(input.Units, "units")
	if rejection != "" {
		return normalizedLabResultInput{}, rejection
	}
	rangeText, rejection := normalizeOptionalLabText(input.RangeText, "range_text")
	if rejection != "" {
		return normalizedLabResultInput{}, rejection
	}
	flag, rejection := normalizeOptionalLabText(input.Flag, "flag")
	if rejection != "" {
		return normalizedLabResultInput{}, rejection
	}
	return normalizedLabResultInput{
		TestName:      testName,
		CanonicalSlug: slug,
		ValueText:     valueText,
		ValueNumeric:  input.ValueNumeric,
		Units:         units,
		RangeText:     rangeText,
		Flag:          flag,
	}, ""
}

func normalizeLabResultUpdateInput(input LabResultUpdateInput) (normalizedLabResultUpdateInput, string) {
	panelName := strings.TrimSpace(input.PanelName)
	if panelName == "" {
		return normalizedLabResultUpdateInput{}, "panel_name is required"
	}
	matchSlug := strings.TrimSpace(input.Match.CanonicalSlug)
	matchTestName := strings.TrimSpace(input.Match.TestName)
	if (matchSlug == "") == (matchTestName == "") {
		return normalizedLabResultUpdateInput{}, "match must include exactly one of canonical_slug or test_name"
	}
	var normalizedSlug *client.AnalyteSlug
	if matchSlug != "" {
		slug, rejection := normalizeAnalyteSlug(matchSlug)
		if rejection != "" {
			return normalizedLabResultUpdateInput{}, "match canonical_slug must be a valid analyte slug"
		}
		normalizedSlug = slug
	}
	result, rejection := normalizeLabResultInput(input.Result)
	if rejection != "" {
		return normalizedLabResultUpdateInput{}, rejection
	}
	return normalizedLabResultUpdateInput{
		PanelName:          panelName,
		MatchCanonicalSlug: normalizedSlug,
		MatchTestName:      matchTestName,
		Result:             result,
	}, ""
}

func normalizeLabTarget(target LabTarget) (normalizedLabTarget, string) {
	if target.ID < 0 {
		return normalizedLabTarget{}, "target id must be greater than 0"
	}
	if target.ID > 0 {
		return normalizedLabTarget{ID: target.ID}, ""
	}
	if target.Date == "" {
		return normalizedLabTarget{}, "target id or date is required"
	}
	date, rejection := parseDateOnly(target.Date)
	if rejection != "" {
		return normalizedLabTarget{}, rejection
	}
	return normalizedLabTarget{Date: &date}, ""
}

func normalizeOptionalLabText(value *string, field string) (*string, string) {
	if value == nil {
		return nil, ""
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, field + " must not be empty"
	}
	return &trimmed, ""
}

func normalizeOptionalAnalyteSlug(value *string) (*client.AnalyteSlug, string) {
	if value == nil {
		return nil, ""
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, "canonical_slug must be a valid analyte slug"
	}
	slug, rejection := normalizeAnalyteSlug(trimmed)
	if rejection != "" {
		return nil, "canonical_slug must be a valid analyte slug"
	}
	return slug, ""
}

func normalizeAnalyteSlug(value string) (*client.AnalyteSlug, string) {
	if value == "" {
		return nil, ""
	}
	slug, ok := client.NormalizeAnalyteSlug(value)
	if !ok {
		return nil, "analyte_slug must be a valid analyte slug"
	}
	return &slug, ""
}

func runLabRecord(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	result := LabTaskResult{}
	for _, collection := range request.Collections {
		existing, err := matchingLabCollections(ctx, api, normalizedLabTarget{Date: &collection.CollectedAt})
		if err != nil {
			return LabTaskResult{}, err
		}
		if duplicate, ok := matchingLabCollection(existing, collection); ok {
			result.Writes = append(result.Writes, labCollectionWrite(duplicate, "already_exists"))
			continue
		}
		written, err := api.CreateLabCollection(ctx, toClientLabCollectionInput(collection))
		if err != nil {
			return LabTaskResult{}, err
		}
		result.Writes = append(result.Writes, labCollectionWrite(written, "created"))
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d lab entries", len(entries))
	return result, nil
}

func runLabCorrect(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	collection := request.Collection
	if collection.Note == nil {
		collection.Note = target.Note
	}
	written, err := api.ReplaceLabCollection(ctx, target.ID, toClientLabCollectionInput(collection))
	if err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabPatch(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	collection := normalizedLabCollectionFromClient(target)
	for _, update := range request.ResultUpdates {
		if rejection := applyLabResultUpdate(&collection, update); rejection != "" {
			return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
		}
	}
	written, err := api.ReplaceLabCollection(ctx, target.ID, toClientLabCollectionInput(collection))
	if err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabDelete(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := api.DeleteLabCollection(ctx, target.ID); err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabList(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	entries, err := listLabEntries(ctx, api, request)
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d lab entries", len(entries)),
	}, nil
}

func labTarget(ctx context.Context, api *client.LocalClient, target normalizedLabTarget) (client.LabCollection, string, error) {
	matches, err := matchingLabCollections(ctx, api, target)
	if err != nil {
		return client.LabCollection{}, "", err
	}
	switch len(matches) {
	case 0:
		return client.LabCollection{}, "no matching lab collection", nil
	case 1:
		return matches[0], "", nil
	default:
		return client.LabCollection{}, "multiple matching lab collections; target is ambiguous", nil
	}
}

func matchingLabCollections(ctx context.Context, api *client.LocalClient, target normalizedLabTarget) ([]client.LabCollection, error) {
	items, err := api.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}
	matches := []client.LabCollection{}
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if target.Date != nil && item.CollectedAt.Format(time.DateOnly) == target.Date.Format(time.DateOnly) {
			matches = append(matches, item)
		}
	}
	return matches, nil
}

func listLabEntries(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) ([]LabCollectionEntry, error) {
	items, err := api.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LabCollectionEntry, 0, len(items))
	for _, item := range items {
		if request.From != nil && item.CollectedAt.Before(*request.From) {
			continue
		}
		if request.To != nil && item.CollectedAt.After(*request.To) {
			continue
		}
		entry, ok := labCollectionEntry(item, request.AnalyteSlug)
		if !ok {
			continue
		}
		out = append(out, entry)
		if request.Limit > 0 && len(out) >= request.Limit {
			break
		}
	}
	return out, nil
}

func normalizedLabCollectionFromClient(item client.LabCollection) normalizedLabCollectionInput {
	panels := make([]normalizedLabPanelInput, 0, len(item.Panels))
	for _, panel := range item.Panels {
		results := make([]normalizedLabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, normalizedLabResultInput{
				TestName:      result.TestName,
				CanonicalSlug: result.CanonicalSlug,
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				RangeText:     result.RangeText,
				Flag:          result.Flag,
			})
		}
		panels = append(panels, normalizedLabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return normalizedLabCollectionInput{
		CollectedAt: item.CollectedAt,
		Note:        item.Note,
		Panels:      panels,
	}
}

func applyLabResultUpdate(collection *normalizedLabCollectionInput, update normalizedLabResultUpdateInput) string {
	panelIndex := -1
	for i, panel := range collection.Panels {
		if strings.EqualFold(strings.TrimSpace(panel.PanelName), update.PanelName) {
			if panelIndex != -1 {
				return "multiple matching lab panels; patch is ambiguous"
			}
			panelIndex = i
		}
	}
	if panelIndex == -1 {
		return "no matching lab panel"
	}

	resultIndex := -1
	for i, result := range collection.Panels[panelIndex].Results {
		if !labResultMatchesPatch(result, update) {
			continue
		}
		if resultIndex != -1 {
			return "multiple matching lab results; patch is ambiguous"
		}
		resultIndex = i
	}
	if resultIndex == -1 {
		return "no matching lab result"
	}

	collection.Panels[panelIndex].Results[resultIndex] = update.Result
	return ""
}

func labResultMatchesPatch(result normalizedLabResultInput, update normalizedLabResultUpdateInput) bool {
	if update.MatchCanonicalSlug != nil {
		return result.CanonicalSlug != nil && *result.CanonicalSlug == *update.MatchCanonicalSlug
	}
	return strings.EqualFold(strings.TrimSpace(result.TestName), update.MatchTestName)
}

func toClientLabCollectionInput(input normalizedLabCollectionInput) client.LabCollectionInput {
	panels := make([]client.LabPanelInput, 0, len(input.Panels))
	for _, panel := range input.Panels {
		results := make([]client.LabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, client.LabResultInput(result))
		}
		panels = append(panels, client.LabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return client.LabCollectionInput{
		CollectedAt: input.CollectedAt,
		Note:        input.Note,
		Panels:      panels,
	}
}

func matchingLabCollection(items []client.LabCollection, input normalizedLabCollectionInput) (client.LabCollection, bool) {
	for _, item := range items {
		if labCollectionMatches(item, input) {
			return item, true
		}
	}
	return client.LabCollection{}, false
}

func labCollectionMatches(item client.LabCollection, input normalizedLabCollectionInput) bool {
	if item.CollectedAt.Format(time.DateOnly) != input.CollectedAt.Format(time.DateOnly) ||
		!equalStringPointer(item.Note, input.Note) ||
		len(item.Panels) != len(input.Panels) {
		return false
	}
	for i, panel := range item.Panels {
		if panel.PanelName != input.Panels[i].PanelName || len(panel.Results) != len(input.Panels[i].Results) {
			return false
		}
		for j, result := range panel.Results {
			if !labResultMatches(result, input.Panels[i].Results[j]) {
				return false
			}
		}
	}
	return true
}

func labResultMatches(item client.LabResult, input normalizedLabResultInput) bool {
	if item.TestName != input.TestName ||
		!equalAnalyteSlugPointer(item.CanonicalSlug, input.CanonicalSlug) ||
		item.ValueText != input.ValueText ||
		!equalFloatPointer(item.ValueNumeric, input.ValueNumeric) ||
		!equalStringPointer(item.Units, input.Units) ||
		!equalStringPointer(item.RangeText, input.RangeText) ||
		!equalStringPointer(item.Flag, input.Flag) {
		return false
	}
	return true
}

func labCollectionWrite(item client.LabCollection, status string) LabCollectionWrite {
	return LabCollectionWrite{
		ID:     item.ID,
		Date:   item.CollectedAt.Format(time.DateOnly),
		Status: status,
	}
}

func labCollectionEntry(item client.LabCollection, analyteSlug *client.AnalyteSlug) (LabCollectionEntry, bool) {
	entry := LabCollectionEntry{
		ID:   item.ID,
		Date: item.CollectedAt.Format(time.DateOnly),
		Note: item.Note,
	}
	for _, panel := range item.Panels {
		panelEntry := LabPanelEntry{PanelName: panel.PanelName}
		for _, result := range panel.Results {
			if analyteSlug != nil && (result.CanonicalSlug == nil || *result.CanonicalSlug != *analyteSlug) {
				continue
			}
			panelEntry.Results = append(panelEntry.Results, labResultEntry(result))
		}
		if len(panelEntry.Results) > 0 {
			entry.Panels = append(entry.Panels, panelEntry)
		}
	}
	if len(entry.Panels) == 0 {
		return LabCollectionEntry{}, false
	}
	return entry, true
}

func labResultEntry(item client.LabResult) LabResultEntry {
	var slug *string
	if item.CanonicalSlug != nil {
		value := string(*item.CanonicalSlug)
		slug = &value
	}
	return LabResultEntry{
		TestName:      item.TestName,
		CanonicalSlug: slug,
		ValueText:     item.ValueText,
		ValueNumeric:  item.ValueNumeric,
		Units:         item.Units,
		RangeText:     item.RangeText,
		Flag:          item.Flag,
	}
}

func equalAnalyteSlugPointer(left *client.AnalyteSlug, right *client.AnalyteSlug) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalFloatPointer(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return math.Abs(*left-*right) < 0.000000001
}
