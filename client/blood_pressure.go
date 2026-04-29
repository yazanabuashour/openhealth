package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
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

func historyFilterFromBloodPressureOptions(options BloodPressureListOptions) health.HistoryFilter {
	return historyFilterFromOptions(options.From, options.To, options.Limit)
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
