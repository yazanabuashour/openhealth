package runclient

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	EnvDatabasePath = localruntime.EnvDatabasePath
)

type LocalConfig = localruntime.Config
type LocalPaths = localruntime.Paths

type WeightUnit = health.WeightUnit
type WeightWriteStatus = health.WeightWriteStatus
type WeightRecordInput = health.WeightRecordInput
type WeightEntry = health.WeightEntry
type WeightWriteResult = health.WeightWriteResult

const (
	WeightUnitLb                   = health.WeightUnitLb
	WeightWriteStatusCreated       = health.WeightWriteStatusCreated
	WeightWriteStatusAlreadyExists = health.WeightWriteStatusAlreadyExists
	WeightWriteStatusUpdated       = health.WeightWriteStatusUpdated
)

type MedicationStatus = health.MedicationStatus
type AnalyteSlug = health.AnalyteSlug
type BloodPressureRecordInput = health.BloodPressureRecordInput
type BloodPressureEntry = health.BloodPressureEntry
type MedicationCourseInput = health.MedicationCourseInput
type MedicationCourse = health.MedicationCourse
type LabResultInput = health.LabResultInput
type LabPanelInput = health.LabPanelInput
type LabCollectionInput = health.LabCollectionInput
type LabResult = health.LabResult
type LabPanel = health.LabPanel
type LabCollection = health.LabCollection
type BodyCompositionInput = health.BodyCompositionInput
type BodyCompositionEntry = health.BodyCompositionEntry
type SleepInput = health.SleepInput
type SleepEntry = health.SleepEntry
type SleepWriteStatus = health.SleepWriteStatus
type SleepWriteResult = health.SleepWriteResult
type ImagingRecordInput = health.ImagingRecordInput
type ImagingRecord = health.ImagingRecord

const (
	MedicationStatusActive = health.MedicationStatusActive
	MedicationStatusAll    = health.MedicationStatusAll

	AnalyteSlugTSH              = health.AnalyteSlugTSH
	AnalyteSlugFreeT4           = health.AnalyteSlugFreeT4
	AnalyteSlugCholesterolTotal = health.AnalyteSlugCholesterolTotal
	AnalyteSlugLDL              = health.AnalyteSlugLDL
	AnalyteSlugHDL              = health.AnalyteSlugHDL
	AnalyteSlugTriglycerides    = health.AnalyteSlugTriglycerides
	AnalyteSlugGlucose          = health.AnalyteSlugGlucose

	SleepWriteStatusCreated       = health.SleepWriteStatusCreated
	SleepWriteStatusAlreadyExists = health.SleepWriteStatusAlreadyExists
	SleepWriteStatusUpdated       = health.SleepWriteStatusUpdated
)

func NormalizeAnalyteSlug(value string) (AnalyteSlug, bool) {
	return health.NormalizeAnalyteSlug(value)
}

type WeightListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type BloodPressureListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type MedicationListOptions struct {
	Status MedicationStatus
}

type BodyCompositionListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type SleepListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
}

type ImagingListOptions struct {
	From     *time.Time
	To       *time.Time
	Limit    int
	Modality *string
	BodySite *string
}

type LocalClient struct {
	Paths LocalPaths

	session *localruntime.Session
}

func OpenLocal(config LocalConfig) (*LocalClient, error) {
	session, err := localruntime.Open(localruntime.Config(config))
	if err != nil {
		return nil, err
	}

	return &LocalClient{
		Paths:   LocalPaths(session.Paths),
		session: session,
	}, nil
}

func (c *LocalClient) Close() error {
	if c == nil || c.session == nil {
		return nil
	}
	return c.session.Close()
}

func (c *LocalClient) UpsertWeight(ctx context.Context, input WeightRecordInput) (WeightWriteResult, error) {
	service, err := c.localService()
	if err != nil {
		return WeightWriteResult{}, err
	}
	return service.UpsertWeight(ctx, input)
}

func (c *LocalClient) ListWeights(ctx context.Context, options WeightListOptions) ([]WeightEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	return service.ListWeight(ctx, historyFilter(options.From, options.To, options.Limit))
}

func (c *LocalClient) RecordBloodPressure(ctx context.Context, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BloodPressureEntry{}, err
	}
	return service.RecordBloodPressure(ctx, input)
}

func (c *LocalClient) ReplaceBloodPressure(ctx context.Context, id int, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BloodPressureEntry{}, err
	}
	return service.ReplaceBloodPressure(ctx, id, input)
}

func (c *LocalClient) ListBloodPressure(ctx context.Context, options BloodPressureListOptions) ([]BloodPressureEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}
	return service.ListBloodPressure(ctx, historyFilter(options.From, options.To, options.Limit))
}

func (c *LocalClient) CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	return service.CreateMedicationCourse(ctx, input)
}

func (c *LocalClient) ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error) {
	service, err := c.localService()
	if err != nil {
		return MedicationCourse{}, err
	}
	return service.ReplaceMedicationCourse(ctx, id, input)
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
	return service.ListMedications(ctx, health.MedicationListParams{Status: health.MedicationStatus(options.Status)})
}

func (c *LocalClient) CreateLabCollection(ctx context.Context, input LabCollectionInput) (LabCollection, error) {
	service, err := c.localService()
	if err != nil {
		return LabCollection{}, err
	}
	return service.CreateLabCollection(ctx, input)
}

func (c *LocalClient) ReplaceLabCollection(ctx context.Context, id int, input LabCollectionInput) (LabCollection, error) {
	service, err := c.localService()
	if err != nil {
		return LabCollection{}, err
	}
	return service.ReplaceLabCollection(ctx, id, input)
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
	return service.ListLabCollections(ctx)
}

func (c *LocalClient) CreateBodyComposition(ctx context.Context, input BodyCompositionInput) (BodyCompositionEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	return service.CreateBodyComposition(ctx, input)
}

func (c *LocalClient) ReplaceBodyComposition(ctx context.Context, id int, input BodyCompositionInput) (BodyCompositionEntry, error) {
	service, err := c.localService()
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	return service.ReplaceBodyComposition(ctx, id, input)
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
	return service.ListBodyComposition(ctx, historyFilter(options.From, options.To, options.Limit))
}

func (c *LocalClient) UpsertSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error) {
	service, err := c.localService()
	if err != nil {
		return SleepWriteResult{}, err
	}
	return service.UpsertSleep(ctx, input)
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
	return service.ListSleep(ctx, historyFilter(options.From, options.To, options.Limit))
}

func (c *LocalClient) CreateImaging(ctx context.Context, input ImagingRecordInput) (ImagingRecord, error) {
	service, err := c.localService()
	if err != nil {
		return ImagingRecord{}, err
	}
	return service.CreateImaging(ctx, input)
}

func (c *LocalClient) ReplaceImaging(ctx context.Context, id int, input ImagingRecordInput) (ImagingRecord, error) {
	service, err := c.localService()
	if err != nil {
		return ImagingRecord{}, err
	}
	return service.ReplaceImaging(ctx, id, input)
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
	return service.ListImaging(ctx, health.ImagingListParams{
		HistoryFilter: historyFilter(options.From, options.To, options.Limit),
		Modality:      options.Modality,
		BodySite:      options.BodySite,
	})
}

func (c *LocalClient) localService() (health.Service, error) {
	if c == nil || c.session == nil || c.session.Service == nil {
		return nil, fmt.Errorf("local OpenHealth runtime is required")
	}
	return c.session.Service, nil
}

func historyFilter(from *time.Time, to *time.Time, limit int) health.HistoryFilter {
	filter := health.HistoryFilter{From: from, To: to}
	if limit != 0 {
		filter.Limit = &limit
	}
	return filter
}
