package client

import (
	"context"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

type WeightRange string

const (
	WeightRange30d WeightRange = "30d"
	WeightRange90d WeightRange = "90d"
	WeightRange1y  WeightRange = "1y"
	WeightRangeAll WeightRange = "all"
)

type MovingAveragePoint struct {
	RecordedAt time.Time
	Value      float64
}

type MonthlyAverageBucket struct {
	Month string
	Value float64
}

type WeightTrendOptions struct {
	Range WeightRange
}

type WeightTrend struct {
	Range                 WeightRange
	RawPoints             []WeightEntry
	MovingAveragePoints   []MovingAveragePoint
	MonthlyAverageBuckets []MonthlyAverageBucket
}

type LabResultWithCollection struct {
	LabResult
	CollectedAt  time.Time
	CollectionID int
	PanelName    string
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

func (c *LocalClient) Summary(ctx context.Context) (Summary, error) {
	service, err := c.localService()
	if err != nil {
		return Summary{}, err
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		return Summary{}, err
	}
	return fromHealthSummary(summary), nil
}

func (c *LocalClient) WeightTrend(ctx context.Context, options WeightTrendOptions) (WeightTrend, error) {
	service, err := c.localService()
	if err != nil {
		return WeightTrend{}, err
	}
	trend, err := service.WeightTrend(ctx, health.WeightTrendParams{
		Range: health.WeightRange(options.Range),
	})
	if err != nil {
		return WeightTrend{}, err
	}
	return fromHealthWeightTrend(trend), nil
}

func fromHealthSummary(summary health.Summary) Summary {
	out := Summary{
		Average7d:             summary.Average7d,
		Delta30d:              summary.Delta30d,
		ActiveMedicationCount: summary.ActiveMedicationCount,
		LatestLabHighlights:   fromHealthLabResultsWithCollection(summary.LatestLabHighlights),
	}
	if summary.LatestWeight != nil {
		latest := fromHealthWeightEntry(*summary.LatestWeight)
		out.LatestWeight = &latest
	}
	if summary.LatestBloodPressure != nil {
		latest := fromHealthBloodPressureEntry(*summary.LatestBloodPressure)
		out.LatestBloodPressure = &latest
	}
	if summary.LatestSleep != nil {
		latest := fromHealthSleepEntry(*summary.LatestSleep)
		out.LatestSleep = &latest
	}
	return out
}

func fromHealthWeightTrend(trend health.WeightTrend) WeightTrend {
	return WeightTrend{
		Range:                 WeightRange(trend.Range),
		RawPoints:             fromHealthWeightEntries(trend.RawPoints),
		MovingAveragePoints:   fromHealthMovingAveragePoints(trend.MovingAveragePoints),
		MonthlyAverageBuckets: fromHealthMonthlyAverageBuckets(trend.MonthlyAverageBuckets),
	}
}

func fromHealthMovingAveragePoints(items []health.MovingAveragePoint) []MovingAveragePoint {
	out := make([]MovingAveragePoint, 0, len(items))
	for _, item := range items {
		out = append(out, MovingAveragePoint(item))
	}
	return out
}

func fromHealthMonthlyAverageBuckets(items []health.MonthlyAverageBucket) []MonthlyAverageBucket {
	out := make([]MonthlyAverageBucket, 0, len(items))
	for _, item := range items {
		out = append(out, MonthlyAverageBucket(item))
	}
	return out
}

func fromHealthLabResultsWithCollection(items []health.LabResultWithCollection) []LabResultWithCollection {
	out := make([]LabResultWithCollection, 0, len(items))
	for _, item := range items {
		out = append(out, fromHealthLabResultWithCollection(item))
	}
	return out
}

func fromHealthLabResultWithCollection(item health.LabResultWithCollection) LabResultWithCollection {
	return LabResultWithCollection{
		LabResult:    fromHealthLabResult(item.LabResult),
		CollectedAt:  item.CollectedAt,
		CollectionID: item.CollectionID,
		PanelName:    item.PanelName,
	}
}
