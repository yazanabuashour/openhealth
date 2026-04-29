package health

import (
	"math"
	"strings"
	"time"
)

func validateHistoryFilter(filter HistoryFilter) error {
	if filter.Limit != nil && (*filter.Limit <= 0 || *filter.Limit > 3650) {
		return &ValidationError{Message: "limit must be between 1 and 3650"}
	}
	return nil
}

func normalizeWeightRecordInput(input WeightRecordInput) (WeightRecordInput, error) {
	if err := validateWeightRecordInput(input); err != nil {
		return WeightRecordInput{}, err
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return WeightRecordInput{}, err
	}
	input.Note = note
	input.RecordedAt = input.RecordedAt.UTC()
	return input, nil
}

func validateWeightRecordInput(input WeightRecordInput) error {
	if input.RecordedAt.IsZero() {
		return &ValidationError{Message: "recorded_at is required"}
	}
	if input.Value <= 0 {
		return &ValidationError{Message: "value must be greater than 0"}
	}
	if input.Unit != WeightUnitLb {
		return &ValidationError{Message: "unit must be 'lb'"}
	}
	return nil
}

func normalizeBloodPressureRecordInput(input BloodPressureRecordInput) (BloodPressureRecordInput, error) {
	if input.RecordedAt.IsZero() {
		return BloodPressureRecordInput{}, &ValidationError{Message: "recorded_at is required"}
	}
	if input.Systolic <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "systolic must be greater than 0"}
	}
	if input.Diastolic <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "diastolic must be greater than 0"}
	}
	if input.Systolic <= input.Diastolic {
		return BloodPressureRecordInput{}, &ValidationError{Message: "systolic must be greater than diastolic"}
	}
	if input.Pulse != nil && *input.Pulse <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "pulse must be greater than 0"}
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return BloodPressureRecordInput{}, err
	}
	input.RecordedAt = input.RecordedAt.UTC()
	input.Note = note
	return input, nil
}

func normalizeMedicationCourseInput(input MedicationCourseInput) (MedicationCourseInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return MedicationCourseInput{}, &ValidationError{Message: "name is required"}
	}
	if input.StartDate == "" {
		return MedicationCourseInput{}, &ValidationError{Message: "start_date is required"}
	}
	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return MedicationCourseInput{}, &ValidationError{Message: "start_date must be YYYY-MM-DD"}
	}
	if input.EndDate != nil {
		if *input.EndDate == "" {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be YYYY-MM-DD"}
		}
		endDate, err := time.Parse(time.DateOnly, *input.EndDate)
		if err != nil {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be YYYY-MM-DD"}
		}
		if endDate.Before(startDate) {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be on or after start_date"}
		}
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return MedicationCourseInput{}, err
	}
	input.Note = note
	return input, nil
}

func normalizeLabCollectionInput(input LabCollectionInput) (LabCollectionInput, error) {
	if input.CollectedAt.IsZero() {
		return LabCollectionInput{}, &ValidationError{Message: "collected_at is required"}
	}
	if len(input.Panels) == 0 {
		return LabCollectionInput{}, &ValidationError{Message: "at least one lab panel is required"}
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return LabCollectionInput{}, err
	}
	input.Note = note
	input.CollectedAt = input.CollectedAt.UTC()
	for panelIndex := range input.Panels {
		panel := &input.Panels[panelIndex]
		panel.PanelName = strings.TrimSpace(panel.PanelName)
		if panel.PanelName == "" {
			return LabCollectionInput{}, &ValidationError{Message: "panel_name is required"}
		}
		if len(panel.Results) == 0 {
			return LabCollectionInput{}, &ValidationError{Message: "at least one lab result is required"}
		}
		for resultIndex := range panel.Results {
			result := &panel.Results[resultIndex]
			result.TestName = strings.TrimSpace(result.TestName)
			if result.TestName == "" {
				return LabCollectionInput{}, &ValidationError{Message: "test_name is required"}
			}
			result.ValueText = strings.TrimSpace(result.ValueText)
			if result.ValueText == "" {
				return LabCollectionInput{}, &ValidationError{Message: "value_text is required"}
			}
			if result.CanonicalSlug != nil {
				validSlug, err := validateAnalyteSlug(*result.CanonicalSlug)
				if err != nil {
					return LabCollectionInput{}, err
				}
				result.CanonicalSlug = &validSlug
			}
			notes, err := normalizeNotes(result.Notes, "notes")
			if err != nil {
				return LabCollectionInput{}, err
			}
			result.Notes = notes
		}
	}
	return input, nil
}

func normalizeBodyCompositionInput(input BodyCompositionInput) (BodyCompositionInput, error) {
	if input.RecordedAt.IsZero() {
		return BodyCompositionInput{}, &ValidationError{Message: "recorded_at is required"}
	}
	if input.BodyFatPercent == nil && input.WeightValue == nil {
		return BodyCompositionInput{}, &ValidationError{Message: "at least one body composition measurement is required"}
	}
	if input.BodyFatPercent != nil && (*input.BodyFatPercent <= 0 || *input.BodyFatPercent > 100) {
		return BodyCompositionInput{}, &ValidationError{Message: "body_fat_percent must be greater than 0 and less than or equal to 100"}
	}
	if (input.WeightValue == nil) != (input.WeightUnit == nil) {
		return BodyCompositionInput{}, &ValidationError{Message: "weight_value and weight_unit must be provided together"}
	}
	if input.WeightValue != nil && *input.WeightValue <= 0 {
		return BodyCompositionInput{}, &ValidationError{Message: "weight_value must be greater than 0"}
	}
	if input.WeightUnit != nil && *input.WeightUnit != WeightUnitLb {
		return BodyCompositionInput{}, &ValidationError{Message: "weight_unit must be 'lb'"}
	}
	method, err := normalizeOptionalText(input.Method, "method")
	if err != nil {
		return BodyCompositionInput{}, err
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return BodyCompositionInput{}, err
	}
	input.RecordedAt = input.RecordedAt.UTC()
	input.Method = method
	input.Note = note
	return input, nil
}

func normalizeSleepInput(input SleepInput) (SleepInput, error) {
	if input.RecordedAt.IsZero() {
		return SleepInput{}, &ValidationError{Message: "recorded_at is required"}
	}
	if input.QualityScore < 1 || input.QualityScore > 5 {
		return SleepInput{}, &ValidationError{Message: "quality_score must be between 1 and 5"}
	}
	if input.WakeupCount != nil && *input.WakeupCount < 0 {
		return SleepInput{}, &ValidationError{Message: "wakeup_count must be greater than or equal to 0"}
	}
	note, err := normalizeOptionalText(input.Note, "note")
	if err != nil {
		return SleepInput{}, err
	}
	input.RecordedAt = input.RecordedAt.UTC()
	input.Note = note
	return input, nil
}

func normalizeImagingRecordInput(input ImagingRecordInput) (ImagingRecordInput, error) {
	if input.PerformedAt.IsZero() {
		return ImagingRecordInput{}, &ValidationError{Message: "performed_at is required"}
	}
	input.Modality = strings.TrimSpace(input.Modality)
	if input.Modality == "" {
		return ImagingRecordInput{}, &ValidationError{Message: "modality is required"}
	}
	input.Summary = strings.TrimSpace(input.Summary)
	if input.Summary == "" {
		return ImagingRecordInput{}, &ValidationError{Message: "summary is required"}
	}
	var err error
	if input.BodySite, err = normalizeOptionalText(input.BodySite, "body_site"); err != nil {
		return ImagingRecordInput{}, err
	}
	if input.Title, err = normalizeOptionalText(input.Title, "title"); err != nil {
		return ImagingRecordInput{}, err
	}
	if input.Impression, err = normalizeOptionalText(input.Impression, "impression"); err != nil {
		return ImagingRecordInput{}, err
	}
	if input.Note, err = normalizeOptionalText(input.Note, "note"); err != nil {
		return ImagingRecordInput{}, err
	}
	if input.Notes, err = normalizeNotes(input.Notes, "notes"); err != nil {
		return ImagingRecordInput{}, err
	}
	input.PerformedAt = input.PerformedAt.UTC()
	return input, nil
}

func normalizeNotes(values []string, field string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	notes := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, &ValidationError{Message: field + " must not contain empty values"}
		}
		notes = append(notes, trimmed)
	}
	return notes, nil
}

func normalizeOptionalText(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, &ValidationError{Message: field + " must not be empty"}
	}
	return &trimmed, nil
}

func equalWeightValue(left float64, right float64) bool {
	return math.Abs(left-right) < 0.000000001
}

func validateWeightUpdateInput(input WeightUpdateInput) error {
	if input.RecordedAt == nil && input.Value == nil && input.Unit == nil && input.Note == nil {
		return &ValidationError{Message: "At least one field must be provided"}
	}
	if input.Value != nil && *input.Value <= 0 {
		return &ValidationError{Message: "value must be greater than 0"}
	}
	if input.Unit != nil && *input.Unit != WeightUnitLb {
		return &ValidationError{Message: "unit must be 'lb'"}
	}
	if _, err := normalizeOptionalText(input.Note, "note"); err != nil {
		return err
	}
	return nil
}

func equalStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func validateWeightID(id int) error {
	return validateRecordID(id)
}

func validateRecordID(id int) error {
	if id <= 0 {
		return &ValidationError{Message: "id must be greater than 0"}
	}
	return nil
}

func normalizeWeightRange(value WeightRange) (WeightRange, error) {
	if value == "" {
		return WeightRange90d, nil
	}
	switch value {
	case WeightRange30d, WeightRange90d, WeightRange1y, WeightRangeAll:
		return value, nil
	default:
		return "", &ValidationError{Message: "range must be one of 30d, 90d, 1y, all"}
	}
}

func normalizeMedicationStatus(value MedicationStatus) (MedicationStatus, error) {
	if value == "" {
		return MedicationStatusActive, nil
	}
	switch value {
	case MedicationStatusActive, MedicationStatusAll:
		return value, nil
	default:
		return "", &ValidationError{Message: "status must be 'active' or 'all'"}
	}
}

func validateAnalyteSlug(value AnalyteSlug) (AnalyteSlug, error) {
	slug, ok := NormalizeAnalyteSlug(string(value))
	if !ok {
		return "", &ValidationError{Message: "slug must be a valid analyte slug"}
	}
	return slug, nil
}
