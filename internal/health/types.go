package health

import (
	"context"
	"strings"
	"time"
	"unicode"
)

type WeightUnit string

const WeightUnitLb WeightUnit = "lb"

type WeightRange string

const (
	WeightRange30d WeightRange = "30d"
	WeightRange90d WeightRange = "90d"
	WeightRange1y  WeightRange = "1y"
	WeightRangeAll WeightRange = "all"
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

type HistoryFilter struct {
	From  *time.Time
	To    *time.Time
	Limit *int
}

type WeightEntry struct {
	ID               int
	RecordedAt       time.Time
	Value            float64
	Unit             WeightUnit
	Source           string
	SourceRecordHash string
	Note             *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type WeightWriteStatus string

const (
	WeightWriteStatusCreated       WeightWriteStatus = "created"
	WeightWriteStatusAlreadyExists WeightWriteStatus = "already_exists"
	WeightWriteStatusUpdated       WeightWriteStatus = "updated"
)

type WeightWriteResult struct {
	Entry  WeightEntry
	Status WeightWriteStatus
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

type MovingAveragePoint struct {
	RecordedAt time.Time
	Value      float64
}

type MonthlyAverageBucket struct {
	Month string
	Value float64
}

type WeightTrend struct {
	Range                 WeightRange
	RawPoints             []WeightEntry
	MovingAveragePoints   []MovingAveragePoint
	MonthlyAverageBuckets []MonthlyAverageBucket
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

type Summary struct {
	LatestWeight          *WeightEntry
	Average7d             *float64
	Delta30d              *float64
	LatestBloodPressure   *BloodPressureEntry
	LatestSleep           *SleepEntry
	ActiveMedicationCount int
	LatestLabHighlights   []LabResultWithCollection
}

type WeightRecordInput struct {
	RecordedAt time.Time
	Value      float64
	Unit       WeightUnit
	Note       *string
}

type BloodPressureRecordInput struct {
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
	Note       *string
}

type MedicationCourseInput struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
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

type BodyCompositionInput struct {
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *WeightUnit
	Method         *string
	Note           *string
}

type SleepInput struct {
	RecordedAt   time.Time
	QualityScore int
	WakeupCount  *int
	Note         *string
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

type WeightUpdateInput struct {
	RecordedAt *time.Time
	Value      *float64
	Unit       *WeightUnit
	Note       *string
}

type WeightTrendParams struct {
	Range WeightRange
}

type MedicationListParams struct {
	Status MedicationStatus
}

type ImagingListParams struct {
	HistoryFilter
	Modality *string
	BodySite *string
}

type CreateWeightEntryParams struct {
	RecordedAt       time.Time
	Value            float64
	Unit             WeightUnit
	Source           string
	SourceRecordHash string
	Note             *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateWeightEntryParams struct {
	ID         int
	RecordedAt *time.Time
	Value      *float64
	Unit       *WeightUnit
	Note       *string
	UpdatedAt  time.Time
}

type DeleteWeightEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type CreateBloodPressureEntryParams struct {
	RecordedAt       time.Time
	Systolic         int
	Diastolic        int
	Pulse            *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateBloodPressureEntryParams struct {
	ID         int
	RecordedAt time.Time
	Systolic   int
	Diastolic  int
	Pulse      *int
	Note       *string
	UpdatedAt  time.Time
}

type DeleteBloodPressureEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type CreateMedicationCourseParams struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
	Source     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UpdateMedicationCourseParams struct {
	ID         int
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
	Note       *string
	UpdatedAt  time.Time
}

type DeleteMedicationCourseParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
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

type CreateBodyCompositionEntryParams struct {
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
}

type UpdateBodyCompositionEntryParams struct {
	ID             int
	RecordedAt     time.Time
	BodyFatPercent *float64
	WeightValue    *float64
	WeightUnit     *WeightUnit
	Method         *string
	Note           *string
	UpdatedAt      time.Time
}

type DeleteBodyCompositionEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type CreateSleepEntryParams struct {
	RecordedAt       time.Time
	QualityScore     int
	WakeupCount      *int
	Note             *string
	Source           string
	SourceRecordHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateSleepEntryParams struct {
	ID           int
	RecordedAt   *time.Time
	QualityScore *int
	WakeupCount  *int
	Note         *string
	UpdatedAt    time.Time
}

type DeleteSleepEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type CreateImagingRecordParams struct {
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
}

type UpdateImagingRecordParams struct {
	ID          int
	PerformedAt time.Time
	Modality    string
	BodySite    *string
	Title       *string
	Summary     string
	Impression  *string
	Note        *string
	Notes       []string
	UpdatedAt   time.Time
}

type DeleteImagingRecordParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type DeleteLabCollectionParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type FindManualWeightEntryParams struct {
	RecordedAt time.Time
	Unit       WeightUnit
}

type FindManualSleepEntryParams struct {
	RecordedAt time.Time
}

type Repository interface {
	ListWeightEntries(ctx context.Context, filter HistoryFilter) ([]WeightEntry, error)
	FindManualWeightEntry(ctx context.Context, params FindManualWeightEntryParams) (*WeightEntry, error)
	CreateWeightEntry(ctx context.Context, params CreateWeightEntryParams) (WeightEntry, error)
	UpdateWeightEntry(ctx context.Context, params UpdateWeightEntryParams) (WeightEntry, error)
	DeleteWeightEntry(ctx context.Context, params DeleteWeightEntryParams) error
	ListBloodPressureEntries(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error)
	CreateBloodPressureEntry(ctx context.Context, params CreateBloodPressureEntryParams) (BloodPressureEntry, error)
	UpdateBloodPressureEntry(ctx context.Context, params UpdateBloodPressureEntryParams) (BloodPressureEntry, error)
	DeleteBloodPressureEntry(ctx context.Context, params DeleteBloodPressureEntryParams) error
	ListMedicationCourses(ctx context.Context, status MedicationStatus, today string) ([]MedicationCourse, error)
	CreateMedicationCourse(ctx context.Context, params CreateMedicationCourseParams) (MedicationCourse, error)
	UpdateMedicationCourse(ctx context.Context, params UpdateMedicationCourseParams) (MedicationCourse, error)
	DeleteMedicationCourse(ctx context.Context, params DeleteMedicationCourseParams) error
	CountActiveMedicationCourses(ctx context.Context, today string) (int, error)
	ListLabCollections(ctx context.Context) ([]LabCollection, error)
	CreateLabCollection(ctx context.Context, params CreateLabCollectionParams) (LabCollection, error)
	UpdateLabCollection(ctx context.Context, params UpdateLabCollectionParams) (LabCollection, error)
	DeleteLabCollection(ctx context.Context, params DeleteLabCollectionParams) error
	ListLabResultsWithCollection(ctx context.Context) ([]LabResultWithCollection, error)
	ListBodyCompositionEntries(ctx context.Context, filter HistoryFilter) ([]BodyCompositionEntry, error)
	CreateBodyCompositionEntry(ctx context.Context, params CreateBodyCompositionEntryParams) (BodyCompositionEntry, error)
	UpdateBodyCompositionEntry(ctx context.Context, params UpdateBodyCompositionEntryParams) (BodyCompositionEntry, error)
	DeleteBodyCompositionEntry(ctx context.Context, params DeleteBodyCompositionEntryParams) error
	ListSleepEntries(ctx context.Context, filter HistoryFilter) ([]SleepEntry, error)
	FindManualSleepEntry(ctx context.Context, params FindManualSleepEntryParams) (*SleepEntry, error)
	CreateSleepEntry(ctx context.Context, params CreateSleepEntryParams) (SleepEntry, error)
	UpdateSleepEntry(ctx context.Context, params UpdateSleepEntryParams) (SleepEntry, error)
	DeleteSleepEntry(ctx context.Context, params DeleteSleepEntryParams) error
	ListImagingRecords(ctx context.Context, params ImagingListParams) ([]ImagingRecord, error)
	CreateImagingRecord(ctx context.Context, params CreateImagingRecordParams) (ImagingRecord, error)
	UpdateImagingRecord(ctx context.Context, params UpdateImagingRecordParams) (ImagingRecord, error)
	DeleteImagingRecord(ctx context.Context, params DeleteImagingRecordParams) error
}

type Service interface {
	Summary(ctx context.Context) (Summary, error)
	ListWeight(ctx context.Context, filter HistoryFilter) ([]WeightEntry, error)
	RecordWeight(ctx context.Context, input WeightRecordInput) (WeightEntry, error)
	UpsertWeight(ctx context.Context, input WeightRecordInput) (WeightWriteResult, error)
	UpdateWeight(ctx context.Context, id int, input WeightUpdateInput) (WeightEntry, error)
	DeleteWeight(ctx context.Context, id int) error
	WeightTrend(ctx context.Context, params WeightTrendParams) (WeightTrend, error)
	ListBloodPressure(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error)
	RecordBloodPressure(ctx context.Context, input BloodPressureRecordInput) (BloodPressureEntry, error)
	ReplaceBloodPressure(ctx context.Context, id int, input BloodPressureRecordInput) (BloodPressureEntry, error)
	DeleteBloodPressure(ctx context.Context, id int) error
	BloodPressureTrend(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error)
	ListMedications(ctx context.Context, params MedicationListParams) ([]MedicationCourse, error)
	CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error)
	ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error)
	DeleteMedicationCourse(ctx context.Context, id int) error
	ListAnalytes(ctx context.Context) ([]AnalyteSummary, error)
	AnalyteTrend(ctx context.Context, slug AnalyteSlug) (AnalyteTrend, error)
	ListLabCollections(ctx context.Context) ([]LabCollection, error)
	CreateLabCollection(ctx context.Context, input LabCollectionInput) (LabCollection, error)
	ReplaceLabCollection(ctx context.Context, id int, input LabCollectionInput) (LabCollection, error)
	DeleteLabCollection(ctx context.Context, id int) error
	ListBodyComposition(ctx context.Context, filter HistoryFilter) ([]BodyCompositionEntry, error)
	CreateBodyComposition(ctx context.Context, input BodyCompositionInput) (BodyCompositionEntry, error)
	ReplaceBodyComposition(ctx context.Context, id int, input BodyCompositionInput) (BodyCompositionEntry, error)
	DeleteBodyComposition(ctx context.Context, id int) error
	ListSleep(ctx context.Context, filter HistoryFilter) ([]SleepEntry, error)
	UpsertSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error)
	DeleteSleep(ctx context.Context, id int) error
	ListImaging(ctx context.Context, params ImagingListParams) ([]ImagingRecord, error)
	CreateImaging(ctx context.Context, input ImagingRecordInput) (ImagingRecord, error)
	ReplaceImaging(ctx context.Context, id int, input ImagingRecordInput) (ImagingRecord, error)
	DeleteImaging(ctx context.Context, id int) error
}
