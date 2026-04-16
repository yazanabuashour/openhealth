package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

type MedicationStatus string

const (
	MedicationStatusActive MedicationStatus = "active"
	MedicationStatusAll    MedicationStatus = "all"
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

type BloodPressureRecordInput struct {
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
}

type BloodPressureListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type BloodPressureEntry struct {
	ID               int
	RecordedAt       time.Time
	Systolic         int
	Diastolic        int
	Pulse            *int
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type MedicationCourseInput struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
}

type MedicationListOptions struct {
	Status MedicationStatus
}

type MedicationCourse struct {
	ID         int
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type LabResultInput struct {
	TestName      string
	CanonicalSlug *AnalyteSlug
	ValueText     string
	ValueNumeric  *float64
	Units         *string
	RangeText     *string
	Flag          *string
}

type LabPanelInput struct {
	PanelName string
	Results   []LabResultInput
}

type LabCollectionInput struct {
	CollectedAt time.Time
	Panels      []LabPanelInput
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
	DisplayOrder  int
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
	Source      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Panels      []LabPanel
}

func (c *LocalClient) RecordBloodPressure(ctx context.Context, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BloodPressureEntry{}, err
	}
	entry, err := service.RecordBloodPressure(ctx, toHealthBloodPressureRecordInput(input))
	if err != nil {
		return BloodPressureEntry{}, err
	}
	return fromHealthBloodPressureEntry(entry), nil
}

func (c *LocalClient) ReplaceBloodPressure(ctx context.Context, id int, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BloodPressureEntry{}, err
	}
	entry, err := service.ReplaceBloodPressure(ctx, id, toHealthBloodPressureRecordInput(input))
	if err != nil {
		return BloodPressureEntry{}, err
	}
	return fromHealthBloodPressureEntry(entry), nil
}

func (c *LocalClient) DeleteBloodPressure(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteBloodPressure(ctx, id)
}

func (c *LocalClient) ListBloodPressure(ctx context.Context, options BloodPressureListOptions) ([]BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	entries, err := service.ListBloodPressure(ctx, historyFilterFromBloodPressureOptions(options))
	if err != nil {
		return nil, err
	}
	return fromHealthBloodPressureEntries(entries), nil
}

func (c *LocalClient) CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	item, err := service.CreateMedicationCourse(ctx, toHealthMedicationCourseInput(input))
	if err != nil {
		return MedicationCourse{}, err
	}
	return fromHealthMedicationCourse(item), nil
}

func (c *LocalClient) ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	item, err := service.ReplaceMedicationCourse(ctx, id, toHealthMedicationCourseInput(input))
	if err != nil {
		return MedicationCourse{}, err
	}
	return fromHealthMedicationCourse(item), nil
}

func (c *LocalClient) DeleteMedicationCourse(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteMedicationCourse(ctx, id)
}

func (c *LocalClient) ListMedicationCourses(ctx context.Context, options MedicationListOptions) ([]MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListMedications(ctx, health.MedicationListParams{
		Status: health.MedicationStatus(options.Status),
	})
	if err != nil {
		return nil, err
	}
	return fromHealthMedicationCourses(items), nil
}

func (c *LocalClient) CreateLabCollection(ctx context.Context, input LabCollectionInput) (LabCollection, error) {
	service, err := c.localService()
	if err != nil {
		return LabCollection{}, err
	}
	item, err := service.CreateLabCollection(ctx, toHealthLabCollectionInput(input))
	if err != nil {
		return LabCollection{}, err
	}
	return fromHealthLabCollection(item), nil
}

func (c *LocalClient) ReplaceLabCollection(ctx context.Context, id int, input LabCollectionInput) (LabCollection, error) {
	service, err := c.localService()
	if err != nil {
		return LabCollection{}, err
	}
	item, err := service.ReplaceLabCollection(ctx, id, toHealthLabCollectionInput(input))
	if err != nil {
		return LabCollection{}, err
	}
	return fromHealthLabCollection(item), nil
}

func (c *LocalClient) DeleteLabCollection(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteLabCollection(ctx, id)
}

func (c *LocalClient) ListLabCollections(ctx context.Context) ([]LabCollection, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}
	return fromHealthLabCollections(items), nil
}

func historyFilterFromBloodPressureOptions(options BloodPressureListOptions) health.HistoryFilter {
	filter := health.HistoryFilter{
		From: options.From,
		To:   options.To,
	}
	if options.Limit != 0 {
		limit := options.Limit
		filter.Limit = &limit
	}
	return filter
}

func toHealthBloodPressureRecordInput(input BloodPressureRecordInput) health.BloodPressureRecordInput {
	return health.BloodPressureRecordInput{
		RecordedAt: input.RecordedAt,
		Systolic:   input.Systolic,
		Diastolic:  input.Diastolic,
		Pulse:      input.Pulse,
	}
}

func toHealthMedicationCourseInput(input MedicationCourseInput) health.MedicationCourseInput {
	return health.MedicationCourseInput(input)
}

func toHealthLabCollectionInput(input LabCollectionInput) health.LabCollectionInput {
	panels := make([]health.LabPanelInput, 0, len(input.Panels))
	for _, panel := range input.Panels {
		results := make([]health.LabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, toHealthLabResultInput(result))
		}
		panels = append(panels, health.LabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return health.LabCollectionInput{
		CollectedAt: input.CollectedAt,
		Panels:      panels,
	}
}

func toHealthLabResultInput(input LabResultInput) health.LabResultInput {
	out := health.LabResultInput{
		TestName:     input.TestName,
		ValueText:    input.ValueText,
		ValueNumeric: input.ValueNumeric,
		Units:        input.Units,
		RangeText:    input.RangeText,
		Flag:         input.Flag,
	}
	if input.CanonicalSlug != nil {
		slug := health.AnalyteSlug(*input.CanonicalSlug)
		out.CanonicalSlug = &slug
	}
	return out
}

func fromHealthBloodPressureEntries(entries []health.BloodPressureEntry) []BloodPressureEntry {
	out := make([]BloodPressureEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, fromHealthBloodPressureEntry(entry))
	}
	return out
}

func fromHealthBloodPressureEntry(entry health.BloodPressureEntry) BloodPressureEntry {
	return BloodPressureEntry{
		ID:               entry.ID,
		RecordedAt:       entry.RecordedAt,
		Systolic:         entry.Systolic,
		Diastolic:        entry.Diastolic,
		Pulse:            entry.Pulse,
		Source:           entry.Source,
		SourceRecordHash: entry.SourceRecordHash,
		CreatedAt:        entry.CreatedAt,
		UpdatedAt:        entry.UpdatedAt,
		DeletedAt:        entry.DeletedAt,
	}
}

func fromHealthMedicationCourses(items []health.MedicationCourse) []MedicationCourse {
	out := make([]MedicationCourse, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthMedicationCourse(item))
	}
	return out
}

func fromHealthMedicationCourse(item health.MedicationCourse) MedicationCourse {
	return MedicationCourse{
		ID:         item.ID,
		Name:       item.Name,
		DosageText: item.DosageText,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		Source:     item.Source,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
		DeletedAt:  item.DeletedAt,
	}
}

func fromHealthLabCollections(items []health.LabCollection) []LabCollection {
	out := make([]LabCollection, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthLabCollection(item))
	}
	return out
}

func fromHealthLabCollection(item health.LabCollection) LabCollection {
	panels := make([]LabPanel, 0, len(item.Panels))
	for _, panel := range item.Panels {
		panels = append(panels, fromHealthLabPanel(panel))
	}
	return LabCollection{
		ID:          item.ID,
		CollectedAt: item.CollectedAt,
		Source:      item.Source,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
		Panels:      panels,
	}
}

func fromHealthLabPanel(item health.LabPanel) LabPanel {
	results := make([]LabResult, 0, len(item.Results))
	for _, result := range item.Results {
		results = append(results, fromHealthLabResult(result))
	}
	return LabPanel{
		ID:           item.ID,
		CollectionID: item.CollectionID,
		PanelName:    item.PanelName,
		DisplayOrder: item.DisplayOrder,
		Results:      results,
	}
}

func fromHealthLabResult(item health.LabResult) LabResult {
	out := LabResult{
		ID:           item.ID,
		PanelID:      item.PanelID,
		TestName:     item.TestName,
		ValueText:    item.ValueText,
		ValueNumeric: item.ValueNumeric,
		Units:        item.Units,
		RangeText:    item.RangeText,
		Flag:         item.Flag,
		DisplayOrder: item.DisplayOrder,
	}
	if item.CanonicalSlug != nil {
		slug := AnalyteSlug(*item.CanonicalSlug)
		out.CanonicalSlug = &slug
	}
	return out
}
