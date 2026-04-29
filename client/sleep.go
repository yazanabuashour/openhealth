package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

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
