package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

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

func toHealthImagingRecordInput(input ImagingRecordInput) health.ImagingRecordInput {
	return health.ImagingRecordInput(input)
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
