package client

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

type WeightUnit string

const WeightUnitLb WeightUnit = "lb"

type WeightWriteStatus string

const (
	WeightWriteStatusCreated       WeightWriteStatus = "created"
	WeightWriteStatusAlreadyExists WeightWriteStatus = "already_exists"
	WeightWriteStatusUpdated       WeightWriteStatus = "updated"
)

type WeightRecordInput struct {
	RecordedAt time.Time
	Value      float64
	Unit       WeightUnit
	Note       *string
}

type WeightListOptions struct {
	From  *time.Time
	To    *time.Time
	Limit int
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

type WeightWriteResult struct {
	Entry  WeightEntry
	Status WeightWriteStatus
}

func (c *LocalClient) RecordWeight(ctx context.Context, input WeightRecordInput) (WeightEntry, error) {
	service, err := c.localService()
	if err != nil {
		return WeightEntry{}, err
	}
	entry, err := service.RecordWeight(ctx, toHealthWeightRecordInput(input))
	if err != nil {
		return WeightEntry{}, err
	}
	return fromHealthWeightEntry(entry), nil
}

func (c *LocalClient) UpsertWeight(ctx context.Context, input WeightRecordInput) (WeightWriteResult, error) {
	service, err := c.localService()
	if err != nil {
		return WeightWriteResult{}, err
	}
	result, err := service.UpsertWeight(ctx, toHealthWeightRecordInput(input))
	if err != nil {
		return WeightWriteResult{}, err
	}
	return WeightWriteResult{
		Entry:  fromHealthWeightEntry(result.Entry),
		Status: WeightWriteStatus(result.Status),
	}, nil
}

func (c *LocalClient) ListWeights(ctx context.Context, options WeightListOptions) ([]WeightEntry, error) {
	service, err := c.localService()
	if err != nil {
		return nil, err
	}

	filter := health.HistoryFilter{
		From: options.From,
		To:   options.To,
	}
	if options.Limit != 0 {
		limit := options.Limit
		filter.Limit = &limit
	}

	entries, err := service.ListWeight(ctx, filter)
	if err != nil {
		return nil, err
	}
	return fromHealthWeightEntries(entries), nil
}

func (c *LocalClient) LatestWeight(ctx context.Context) (*WeightEntry, error) {
	entries, err := c.ListWeights(ctx, WeightListOptions{Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return &entries[0], nil
}

func (c *LocalClient) localService() (health.Service, error) {
	if c == nil || c.session == nil || c.session.Service == nil {
		return nil, fmt.Errorf("local OpenHealth client is required")
	}
	return c.session.Service, nil
}

func toHealthWeightRecordInput(input WeightRecordInput) health.WeightRecordInput {
	unit := input.Unit
	if unit == "" {
		unit = WeightUnitLb
	}
	return health.WeightRecordInput{
		RecordedAt: input.RecordedAt,
		Value:      input.Value,
		Unit:       health.WeightUnit(unit),
		Note:       input.Note,
	}
}

func fromHealthWeightEntries(entries []health.WeightEntry) []WeightEntry {
	out := make([]WeightEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, fromHealthWeightEntry(entry))
	}
	return out
}

func fromHealthWeightEntry(entry health.WeightEntry) WeightEntry {
	return WeightEntry{
		ID:               entry.ID,
		RecordedAt:       entry.RecordedAt,
		Value:            entry.Value,
		Unit:             WeightUnit(entry.Unit),
		Source:           entry.Source,
		SourceRecordHash: entry.SourceRecordHash,
		Note:             entry.Note,
		CreatedAt:        entry.CreatedAt,
		UpdatedAt:        entry.UpdatedAt,
		DeletedAt:        entry.DeletedAt,
	}
}
