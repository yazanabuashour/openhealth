package health

import (
	"context"
	"time"
)

type HistoryFilter struct {
	From  *time.Time
	To    *time.Time
	Limit *int
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
	ResolveLabCollectionTarget(ctx context.Context, target LabCollectionTarget) (LabCollection, string, error)
	CreateLabCollection(ctx context.Context, input LabCollectionInput) (LabCollection, error)
	ReplaceLabCollection(ctx context.Context, id int, input LabCollectionInput) (LabCollection, error)
	DeleteLabCollection(ctx context.Context, id int) error
	ListBodyComposition(ctx context.Context, filter HistoryFilter) ([]BodyCompositionEntry, error)
	ResolveBodyCompositionTarget(ctx context.Context, target BodyCompositionTarget) (BodyCompositionEntry, string, error)
	CreateBodyComposition(ctx context.Context, input BodyCompositionInput) (BodyCompositionEntry, error)
	ReplaceBodyComposition(ctx context.Context, id int, input BodyCompositionInput) (BodyCompositionEntry, error)
	DeleteBodyComposition(ctx context.Context, id int) error
	ListSleep(ctx context.Context, filter HistoryFilter) ([]SleepEntry, error)
	ResolveSleepTarget(ctx context.Context, target SleepTarget) (SleepEntry, string, error)
	UpsertSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error)
	DeleteSleep(ctx context.Context, id int) error
	ListImaging(ctx context.Context, params ImagingListParams) ([]ImagingRecord, error)
	ResolveImagingTarget(ctx context.Context, target ImagingTarget) (ImagingRecord, string, error)
	CreateImaging(ctx context.Context, input ImagingRecordInput) (ImagingRecord, error)
	ReplaceImaging(ctx context.Context, id int, input ImagingRecordInput) (ImagingRecord, error)
	DeleteImaging(ctx context.Context, id int) error
	ResolveMedicationTarget(ctx context.Context, target MedicationTarget) (MedicationCourse, string, error)
}
