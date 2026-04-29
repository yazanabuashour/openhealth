package runner

import (
	"fmt"
	"strings"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

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
	notes, rejection := normalizeNoteList(input.Notes, "notes")
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
		Notes:         notes,
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
