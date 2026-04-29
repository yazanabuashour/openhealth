package health

import (
	"strings"
	"time"
	"unicode"
)

type AnalyteSlug string

const (
	AnalyteSlugTSH              AnalyteSlug = "tsh"
	AnalyteSlugFreeT4           AnalyteSlug = "free-t4"
	AnalyteSlugCholesterolTotal AnalyteSlug = "cholesterol-total"
	AnalyteSlugLDL              AnalyteSlug = "ldl"
	AnalyteSlugHDL              AnalyteSlug = "hdl"
	AnalyteSlugTriglycerides    AnalyteSlug = "triglycerides"
	AnalyteSlugGlucose          AnalyteSlug = "glucose"
)

func NormalizeAnalyteSlug(value string) (AnalyteSlug, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "", false
	}

	var builder strings.Builder
	previousHyphen := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			previousHyphen = false
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if !previousHyphen {
				builder.WriteByte('-')
				previousHyphen = true
			}
		default:
			return "", false
		}
	}

	normalized := builder.String()
	if normalized == "" || normalized[0] == '-' || normalized[len(normalized)-1] == '-' {
		return "", false
	}
	return AnalyteSlug(normalized), true
}

type LabResult struct {
	ID            int
	PanelID       int
	TestName      string
	CanonicalSlug *AnalyteSlug
	ValueText     string
	ValueNumeric  *float64
	Units         *string
	RangeText     *string
	Flag          *string
	Notes         []string
	DisplayOrder  int
}

type LabResultWithCollection struct {
	LabResult
	CollectedAt  time.Time
	CollectionID int
	PanelName    string
}

type LabPanel struct {
	ID           int
	CollectionID int
	PanelName    string
	DisplayOrder int
	Results      []LabResult
}

type LabCollection struct {
	ID          int
	CollectedAt time.Time
	Note        *string
	Source      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Panels      []LabPanel
}

type AnalyteSummary struct {
	Slug     AnalyteSlug
	Latest   LabResultWithCollection
	Previous *LabResultWithCollection
}

type AnalyteTrend struct {
	Slug     AnalyteSlug
	Latest   *LabResultWithCollection
	Previous *LabResultWithCollection
	Points   []LabResultWithCollection
}

type LabResultInput struct {
	TestName      string
	CanonicalSlug *AnalyteSlug
	ValueText     string
	ValueNumeric  *float64
	Units         *string
	RangeText     *string
	Flag          *string
	Notes         []string
}

type LabPanelInput struct {
	PanelName string
	Results   []LabResultInput
}

type LabCollectionInput struct {
	CollectedAt time.Time
	Note        *string
	Panels      []LabPanelInput
}

type LabResultWriteParams struct {
	TestName      string
	CanonicalSlug *AnalyteSlug
	ValueText     string
	ValueNumeric  *float64
	Units         *string
	RangeText     *string
	Flag          *string
	Notes         []string
	DisplayOrder  int
}

type LabPanelWriteParams struct {
	PanelName    string
	DisplayOrder int
	Results      []LabResultWriteParams
}

type CreateLabCollectionParams struct {
	CollectedAt time.Time
	Note        *string
	Source      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Panels      []LabPanelWriteParams
}

type UpdateLabCollectionParams struct {
	ID          int
	CollectedAt time.Time
	Note        *string
	UpdatedAt   time.Time
	Panels      []LabPanelWriteParams
}

type DeleteLabCollectionParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}
