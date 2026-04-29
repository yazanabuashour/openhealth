package runner

import (
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
	Notes         []string `json:"notes,omitempty"`
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
	Notes         []string `json:"notes,omitempty"`
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
	Notes         []string
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
