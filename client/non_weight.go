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
	Note       *string
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
	Note             *string
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
	Note       *string
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
	Note       *string
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

type BodyCompositionInput struct {
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *WeightUnit
	Method         *string
	Note           *string
}

type BodyCompositionListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type BodyCompositionEntry struct {
	ID               int
	RecordedAt       time.Time
	BodyFatPercent   *float64
	WeightValue      *float64
	WeightUnit       *WeightUnit
	Method           *string
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type SleepInput struct {
	RecordedAt   time.Time
	QualityScore int
	WakeupCount  *int
	Note         *string
}

type SleepListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type SleepEntry struct {
	ID               int
	RecordedAt       time.Time
	QualityScore     int
	WakeupCount      *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type SleepWriteStatus string

const (
	SleepWriteStatusCreated       SleepWriteStatus = "created"
	SleepWriteStatusAlreadyExists SleepWriteStatus = "already_exists"
	SleepWriteStatusUpdated       SleepWriteStatus = "updated"
)

type SleepWriteResult struct {
	Entry  SleepEntry
	Status SleepWriteStatus
}

type ImagingRecordInput struct {
	PerformedAt time.Time
	Modality    string
	BodySite    *string
	Title       *string
	Summary     string
	Impression  *string
	Note        *string
	Notes       []string
}

type ImagingListOptions struct {
	From     *time.Time
	To       *time.Time
	Limit    int
	Modality *string
	BodySite *string
}

type ImagingRecord struct {
	ID               int
	PerformedAt      time.Time
	Modality         string
	BodySite         *string
	Title            *string
	Summary          string
	Impression       *string
	Note             *string
	Notes            []string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
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

func (c *LocalClient) CreateBodyComposition(ctx context.Context, input BodyCompositionInput) (BodyCompositionEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	item, err := service.CreateBodyComposition(ctx, toHealthBodyCompositionInput(input))
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	return fromHealthBodyCompositionEntry(item), nil
}

func (c *LocalClient) ReplaceBodyComposition(ctx context.Context, id int, input BodyCompositionInput) (BodyCompositionEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	item, err := service.ReplaceBodyComposition(ctx, id, toHealthBodyCompositionInput(input))
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	return fromHealthBodyCompositionEntry(item), nil
}

func (c *LocalClient) DeleteBodyComposition(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteBodyComposition(ctx, id)
}

func (c *LocalClient) ListBodyComposition(ctx context.Context, options BodyCompositionListOptions) ([]BodyCompositionEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListBodyComposition(ctx, historyFilterFromOptions(options.From, options.To, options.Limit))
	if err != nil {
		return nil, err
	}
	return fromHealthBodyCompositionEntries(items), nil
}

func (c *LocalClient) UpsertSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error) {
	service, err := c.localService()
	if err != nil {
		return SleepWriteResult{}, err
	}
	result, err := service.UpsertSleep(ctx, health.SleepInput(input))
	if err != nil {
		return SleepWriteResult{}, err
	}
	return fromHealthSleepWriteResult(result), nil
}

func (c *LocalClient) DeleteSleep(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteSleep(ctx, id)
}

func (c *LocalClient) ListSleep(ctx context.Context, options SleepListOptions) ([]SleepEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListSleep(ctx, historyFilterFromOptions(options.From, options.To, options.Limit))
	if err != nil {
		return nil, err
	}
	return fromHealthSleepEntries(items), nil
}

func (c *LocalClient) CreateImaging(ctx context.Context, input ImagingRecordInput) (ImagingRecord, error) {
	service, err := c.localService()
	if err != nil {
		return ImagingRecord{}, err
	}
	item, err := service.CreateImaging(ctx, toHealthImagingRecordInput(input))
	if err != nil {
		return ImagingRecord{}, err
	}
	return fromHealthImagingRecord(item), nil
}

func (c *LocalClient) ReplaceImaging(ctx context.Context, id int, input ImagingRecordInput) (ImagingRecord, error) {
	service, err := c.localService()
	if err != nil {
		return ImagingRecord{}, err
	}
	item, err := service.ReplaceImaging(ctx, id, toHealthImagingRecordInput(input))
	if err != nil {
		return ImagingRecord{}, err
	}
	return fromHealthImagingRecord(item), nil
}

func (c *LocalClient) DeleteImaging(ctx context.Context, id int) error {
	service, err := c.localService()
	if err != nil {
		return err
	}
	return service.DeleteImaging(ctx, id)
}

func (c *LocalClient) ListImaging(ctx context.Context, options ImagingListOptions) ([]ImagingRecord, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	items, err := service.ListImaging(ctx, health.ImagingListParams{
		HistoryFilter: historyFilterFromOptions(options.From, options.To, options.Limit),
		Modality:      options.Modality,
		BodySite:      options.BodySite,
	})
	if err != nil {
		return nil, err
	}
	return fromHealthImagingRecords(items), nil
}

func historyFilterFromBloodPressureOptions(options BloodPressureListOptions) health.HistoryFilter {
	return historyFilterFromOptions(options.From, options.To, options.Limit)
}

func historyFilterFromOptions(from *time.Time, to *time.Time, limitValue int) health.HistoryFilter {
	filter := health.HistoryFilter{From: from, To: to}
	if limitValue != 0 {
		limit := limitValue
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
		Note:       input.Note,
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
		Note:        input.Note,
		Panels:      panels,
	}
}

func toHealthBodyCompositionInput(input BodyCompositionInput) health.BodyCompositionInput {
	var weightUnit *health.WeightUnit
	if input.WeightUnit != nil {
		unit := health.WeightUnit(*input.WeightUnit)
		weightUnit = &unit
	}
	return health.BodyCompositionInput{
		RecordedAt:     input.RecordedAt,
		BodyFatPercent: input.BodyFatPercent,
		WeightValue:    input.WeightValue,
		WeightUnit:     weightUnit,
		Method:         input.Method,
		Note:           input.Note,
	}
}

func toHealthImagingRecordInput(input ImagingRecordInput) health.ImagingRecordInput {
	return health.ImagingRecordInput(input)
}

func toHealthLabResultInput(input LabResultInput) health.LabResultInput {
	out := health.LabResultInput{
		TestName:     input.TestName,
		ValueText:    input.ValueText,
		ValueNumeric: input.ValueNumeric,
		Units:        input.Units,
		RangeText:    input.RangeText,
		Flag:         input.Flag,
		Notes:        input.Notes,
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
		Note:             entry.Note,
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
		Note:       item.Note,
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
		Note:        item.Note,
		Source:      item.Source,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
		Panels:      panels,
	}
}

func fromHealthBodyCompositionEntries(items []health.BodyCompositionEntry) []BodyCompositionEntry {
	out := make([]BodyCompositionEntry, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthBodyCompositionEntry(item))
	}
	return out
}

func fromHealthBodyCompositionEntry(item health.BodyCompositionEntry) BodyCompositionEntry {
	var weightUnit *WeightUnit
	if item.WeightUnit != nil {
		unit := WeightUnit(*item.WeightUnit)
		weightUnit = &unit
	}
	return BodyCompositionEntry{
		ID:               item.ID,
		RecordedAt:       item.RecordedAt,
		BodyFatPercent:   item.BodyFatPercent,
		WeightValue:      item.WeightValue,
		WeightUnit:       weightUnit,
		Method:           item.Method,
		Note:             item.Note,
		Source:           item.Source,
		SourceRecordHash: item.SourceRecordHash,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		DeletedAt:        item.DeletedAt,
	}
}

func fromHealthSleepEntries(items []health.SleepEntry) []SleepEntry {
	out := make([]SleepEntry, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthSleepEntry(item))
	}
	return out
}

func fromHealthSleepEntry(item health.SleepEntry) SleepEntry {
	return SleepEntry{
		ID:               item.ID,
		RecordedAt:       item.RecordedAt,
		QualityScore:     item.QualityScore,
		WakeupCount:      item.WakeupCount,
		Note:             item.Note,
		Source:           item.Source,
		SourceRecordHash: item.SourceRecordHash,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		DeletedAt:        item.DeletedAt,
	}
}

func fromHealthSleepWriteResult(item health.SleepWriteResult) SleepWriteResult {
	return SleepWriteResult{
		Entry:  fromHealthSleepEntry(item.Entry),
		Status: SleepWriteStatus(item.Status),
	}
}

func fromHealthImagingRecords(items []health.ImagingRecord) []ImagingRecord {
	out := make([]ImagingRecord, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthImagingRecord(item))
	}
	return out
}

func fromHealthImagingRecord(item health.ImagingRecord) ImagingRecord {
	return ImagingRecord{
		ID:               item.ID,
		PerformedAt:      item.PerformedAt,
		Modality:         item.Modality,
		BodySite:         item.BodySite,
		Title:            item.Title,
		Summary:          item.Summary,
		Impression:       item.Impression,
		Note:             item.Note,
		Notes:            append([]string(nil), item.Notes...),
		Source:           item.Source,
		SourceRecordHash: item.SourceRecordHash,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		DeletedAt:        item.DeletedAt,
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
		Notes:        append([]string(nil), item.Notes...),
		DisplayOrder: item.DisplayOrder,
	}
	if item.CanonicalSlug != nil {
		slug := AnalyteSlug(*item.CanonicalSlug)
		out.CanonicalSlug = &slug
	}
	return out
}
