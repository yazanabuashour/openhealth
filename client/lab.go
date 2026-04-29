package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
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
