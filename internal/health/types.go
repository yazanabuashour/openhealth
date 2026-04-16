package health

import (
	"context"
	"time"
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

var validAnalyteSlugs = map[AnalyteSlug]struct{}{
	AnalyteSlugTSH:              {},
	AnalyteSlugFreeT4:           {},
	AnalyteSlugCholesterolTotal: {},
	AnalyteSlugLDL:              {},
	AnalyteSlugHDL:              {},
	AnalyteSlugTriglycerides:    {},
	AnalyteSlugGlucose:          {},
}

func NormalizeAnalyteSlug(value string) (AnalyteSlug, bool) {
	slug := AnalyteSlug(value)
	_, ok := validAnalyteSlugs[slug]
	return slug, ok
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
	Source      string
	CreatedAt   time.Time
	Panels      []LabPanel
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
	ActiveMedicationCount int
	LatestLabHighlights   []LabResultWithCollection
}

type WeightRecordInput struct {
	RecordedAt time.Time
	Value      float64
	Unit       WeightUnit
}

type WeightUpdateInput struct {
	RecordedAt *time.Time
	Value      *float64
	Unit       *WeightUnit
}

type WeightTrendParams struct {
	Range WeightRange
}

type MedicationListParams struct {
	Status MedicationStatus
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
	UpdatedAt  time.Time
}

type DeleteWeightEntryParams struct {
	ID        int
	DeletedAt time.Time
	UpdatedAt time.Time
}

type FindManualWeightEntryParams struct {
	RecordedAt time.Time
	Unit       WeightUnit
}

type Repository interface {
	ListWeightEntries(ctx context.Context, filter HistoryFilter) ([]WeightEntry, error)
	FindManualWeightEntry(ctx context.Context, params FindManualWeightEntryParams) (*WeightEntry, error)
	CreateWeightEntry(ctx context.Context, params CreateWeightEntryParams) (WeightEntry, error)
	UpdateWeightEntry(ctx context.Context, params UpdateWeightEntryParams) (WeightEntry, error)
	DeleteWeightEntry(ctx context.Context, params DeleteWeightEntryParams) error
	ListBloodPressureEntries(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error)
	ListMedicationCourses(ctx context.Context, status MedicationStatus, today string) ([]MedicationCourse, error)
	CountActiveMedicationCourses(ctx context.Context, today string) (int, error)
	ListLabCollections(ctx context.Context) ([]LabCollection, error)
	ListLabResultsWithCollection(ctx context.Context) ([]LabResultWithCollection, error)
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
	BloodPressureTrend(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error)
	ListMedications(ctx context.Context, params MedicationListParams) ([]MedicationCourse, error)
	ListAnalytes(ctx context.Context) ([]AnalyteSummary, error)
	AnalyteTrend(ctx context.Context, slug AnalyteSlug) (AnalyteTrend, error)
	ListLabCollections(ctx context.Context) ([]LabCollection, error)
}
