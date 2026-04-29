package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

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
