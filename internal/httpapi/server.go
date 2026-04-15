package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/yazanabuashour/openhealth/internal/api/generated"
	"github.com/yazanabuashour/openhealth/internal/health"
)

type server struct {
	service health.Service
}

func NewHandler(service health.Service) http.Handler {
	strict := generated.NewStrictHandlerWithOptions(
		&server{service: service},
		nil,
		generated.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		},
	)

	return withCorrelationID(generated.Handler(strict))
}

func (s *server) Health(context.Context, generated.HealthRequestObject) (generated.HealthResponseObject, error) {
	return generated.Health200JSONResponse{
		Ok: true,
	}, nil
}

func (s *server) GetHealthSummary(ctx context.Context, _ generated.GetHealthSummaryRequestObject) (generated.GetHealthSummaryResponseObject, error) {
	summary, err := s.service.Summary(ctx)
	if err != nil {
		return nil, err
	}
	body, err := toGeneratedSummary(summary)
	if err != nil {
		return nil, err
	}
	return generated.GetHealthSummary200JSONResponse(body), nil
}

func (s *server) ListHealthWeight(ctx context.Context, request generated.ListHealthWeightRequestObject) (generated.ListHealthWeightResponseObject, error) {
	items, err := s.service.ListWeight(ctx, historyFilter(request.Params.From, request.Params.To, request.Params.Limit))
	if err != nil {
		return nil, err
	}
	return generated.ListHealthWeight200JSONResponse{
		Items: mapWeightEntries(items),
	}, nil
}

func (s *server) CreateHealthWeight(ctx context.Context, request generated.CreateHealthWeightRequestObject) (generated.CreateHealthWeightResponseObject, error) {
	if request.Body == nil {
		return nil, &health.ValidationError{Message: "request body is required"}
	}
	entry, err := s.service.RecordWeight(ctx, health.WeightRecordInput{
		RecordedAt: request.Body.RecordedAt,
		Value:      float64(request.Body.Value),
		Unit:       health.WeightUnit(request.Body.Unit),
	})
	if err != nil {
		return nil, err
	}
	return generated.CreateHealthWeight201JSONResponse(toGeneratedWeightEntry(entry)), nil
}

func (s *server) UpdateHealthWeight(ctx context.Context, request generated.UpdateHealthWeightRequestObject) (generated.UpdateHealthWeightResponseObject, error) {
	if request.Body == nil {
		return nil, &health.ValidationError{Message: "request body is required"}
	}

	input := health.WeightUpdateInput{}
	if request.Body.RecordedAt != nil {
		recordedAt := request.Body.RecordedAt.UTC()
		input.RecordedAt = &recordedAt
	}
	if request.Body.Value != nil {
		value := float64(*request.Body.Value)
		input.Value = &value
	}
	if request.Body.Unit != nil {
		unit := health.WeightUnit(*request.Body.Unit)
		input.Unit = &unit
	}

	entry, err := s.service.UpdateWeight(ctx, request.Id, input)
	if err != nil {
		return nil, err
	}
	return generated.UpdateHealthWeight200JSONResponse(toGeneratedWeightEntry(entry)), nil
}

func (s *server) DeleteHealthWeight(ctx context.Context, request generated.DeleteHealthWeightRequestObject) (generated.DeleteHealthWeightResponseObject, error) {
	if err := s.service.DeleteWeight(ctx, request.Id); err != nil {
		return nil, err
	}
	return generated.DeleteHealthWeight200JSONResponse{
		Success: true,
	}, nil
}

func (s *server) GetHealthWeightTrend(ctx context.Context, request generated.GetHealthWeightTrendRequestObject) (generated.GetHealthWeightTrendResponseObject, error) {
	params := health.WeightTrendParams{}
	if request.Params.Range != nil {
		params.Range = health.WeightRange(*request.Params.Range)
	}
	trend, err := s.service.WeightTrend(ctx, params)
	if err != nil {
		return nil, err
	}
	return generated.GetHealthWeightTrend200JSONResponse{
		Range:                 generated.HealthWeightRange(trend.Range),
		RawPoints:             mapWeightEntries(trend.RawPoints),
		MovingAveragePoints:   mapMovingAveragePoints(trend.MovingAveragePoints),
		MonthlyAverageBuckets: mapMonthlyAverageBuckets(trend.MonthlyAverageBuckets),
	}, nil
}

func (s *server) ListHealthBloodPressure(ctx context.Context, request generated.ListHealthBloodPressureRequestObject) (generated.ListHealthBloodPressureResponseObject, error) {
	items, err := s.service.ListBloodPressure(ctx, historyFilter(request.Params.From, request.Params.To, request.Params.Limit))
	if err != nil {
		return nil, err
	}
	return generated.ListHealthBloodPressure200JSONResponse{
		Items: mapBloodPressureEntries(items),
	}, nil
}

func (s *server) GetHealthBloodPressureTrend(ctx context.Context, request generated.GetHealthBloodPressureTrendRequestObject) (generated.GetHealthBloodPressureTrendResponseObject, error) {
	items, err := s.service.BloodPressureTrend(ctx, historyFilter(request.Params.From, request.Params.To, request.Params.Limit))
	if err != nil {
		return nil, err
	}
	return generated.GetHealthBloodPressureTrend200JSONResponse{
		Items: mapBloodPressureEntries(items),
	}, nil
}

func (s *server) ListHealthMedications(ctx context.Context, request generated.ListHealthMedicationsRequestObject) (generated.ListHealthMedicationsResponseObject, error) {
	params := health.MedicationListParams{}
	if request.Params.Status != nil {
		params.Status = health.MedicationStatus(*request.Params.Status)
	}
	items, err := s.service.ListMedications(ctx, params)
	if err != nil {
		return nil, err
	}

	body, err := mapMedicationCourses(items)
	if err != nil {
		return nil, err
	}
	return generated.ListHealthMedications200JSONResponse{
		Items: body,
	}, nil
}

func (s *server) ListHealthLabAnalytes(ctx context.Context, _ generated.ListHealthLabAnalytesRequestObject) (generated.ListHealthLabAnalytesResponseObject, error) {
	items, err := s.service.ListAnalytes(ctx)
	if err != nil {
		return nil, err
	}
	return generated.ListHealthLabAnalytes200JSONResponse{
		Items: mapAnalyteSummaries(items),
	}, nil
}

func (s *server) GetHealthLabAnalyteTrend(ctx context.Context, request generated.GetHealthLabAnalyteTrendRequestObject) (generated.GetHealthLabAnalyteTrendResponseObject, error) {
	trend, err := s.service.AnalyteTrend(ctx, health.AnalyteSlug(request.Slug))
	if err != nil {
		return nil, err
	}
	body := generated.HealthAnalyteTrend{
		Slug:   generated.HealthAnalyteSlug(trend.Slug),
		Points: mapLabResultsWithCollection(trend.Points),
	}
	if trend.Latest != nil {
		latest := toGeneratedLabResultWithCollection(*trend.Latest)
		body.Latest = &latest
	}
	if trend.Previous != nil {
		previous := toGeneratedLabResultWithCollection(*trend.Previous)
		body.Previous = &previous
	}
	return generated.GetHealthLabAnalyteTrend200JSONResponse(body), nil
}

func (s *server) ListHealthLabCollections(ctx context.Context, _ generated.ListHealthLabCollectionsRequestObject) (generated.ListHealthLabCollectionsResponseObject, error) {
	items, err := s.service.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}

	body, err := mapLabCollections(items)
	if err != nil {
		return nil, err
	}
	return generated.ListHealthLabCollections200JSONResponse{
		Items: body,
	}, nil
}

func historyFilter(from *time.Time, to *time.Time, limit *int) health.HistoryFilter {
	return health.HistoryFilter{
		From:  from,
		To:    to,
		Limit: limit,
	}
}

func toGeneratedSummary(summary health.Summary) (generated.HealthSummary, error) {
	body := generated.HealthSummary{
		ActiveMedicationCount: summary.ActiveMedicationCount,
		LatestLabHighlights:   mapLabResultsWithCollection(summary.LatestLabHighlights),
	}
	if summary.LatestWeight != nil {
		latestWeight := toGeneratedWeightEntry(*summary.LatestWeight)
		body.LatestWeight = &latestWeight
	}
	if summary.Average7d != nil {
		value := float32(*summary.Average7d)
		body.Average7d = &value
	}
	if summary.Delta30d != nil {
		value := float32(*summary.Delta30d)
		body.Delta30d = &value
	}
	if summary.LatestBloodPressure != nil {
		latestBloodPressure := toGeneratedBloodPressureEntry(*summary.LatestBloodPressure)
		body.LatestBloodPressure = &latestBloodPressure
	}
	return body, nil
}

func mapWeightEntries(items []health.WeightEntry) []generated.HealthWeightEntry {
	out := make([]generated.HealthWeightEntry, 0, len(items))
	for _, item := range items {
		out = append(out, toGeneratedWeightEntry(item))
	}
	return out
}

func toGeneratedWeightEntry(item health.WeightEntry) generated.HealthWeightEntry {
	return generated.HealthWeightEntry{
		CreatedAt:        item.CreatedAt,
		DeletedAt:        item.DeletedAt,
		Id:               item.ID,
		Note:             item.Note,
		RecordedAt:       item.RecordedAt,
		Source:           item.Source,
		SourceRecordHash: item.SourceRecordHash,
		Unit:             generated.HealthWeightEntryUnit(item.Unit),
		UpdatedAt:        item.UpdatedAt,
		Value:            float32(item.Value),
	}
}

func mapMovingAveragePoints(items []health.MovingAveragePoint) []generated.HealthMovingAveragePoint {
	out := make([]generated.HealthMovingAveragePoint, 0, len(items))
	for _, item := range items {
		out = append(out, generated.HealthMovingAveragePoint{
			RecordedAt: item.RecordedAt,
			Value:      float32(item.Value),
		})
	}
	return out
}

func mapMonthlyAverageBuckets(items []health.MonthlyAverageBucket) []generated.HealthMonthlyAverageBucket {
	out := make([]generated.HealthMonthlyAverageBucket, 0, len(items))
	for _, item := range items {
		out = append(out, generated.HealthMonthlyAverageBucket{
			Month: item.Month,
			Value: float32(item.Value),
		})
	}
	return out
}

func mapBloodPressureEntries(items []health.BloodPressureEntry) []generated.HealthBloodPressureEntry {
	out := make([]generated.HealthBloodPressureEntry, 0, len(items))
	for _, item := range items {
		out = append(out, toGeneratedBloodPressureEntry(item))
	}
	return out
}

func toGeneratedBloodPressureEntry(item health.BloodPressureEntry) generated.HealthBloodPressureEntry {
	return generated.HealthBloodPressureEntry{
		CreatedAt:        item.CreatedAt,
		DeletedAt:        item.DeletedAt,
		Diastolic:        item.Diastolic,
		Id:               item.ID,
		Pulse:            item.Pulse,
		RecordedAt:       item.RecordedAt,
		Source:           item.Source,
		SourceRecordHash: item.SourceRecordHash,
		Systolic:         item.Systolic,
		UpdatedAt:        item.UpdatedAt,
	}
}

func mapMedicationCourses(items []health.MedicationCourse) ([]generated.HealthMedicationCourse, error) {
	out := make([]generated.HealthMedicationCourse, 0, len(items))
	for _, item := range items {
		converted, err := toGeneratedMedicationCourse(item)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

func toGeneratedMedicationCourse(item health.MedicationCourse) (generated.HealthMedicationCourse, error) {
	startDate, err := toDate(item.StartDate)
	if err != nil {
		return generated.HealthMedicationCourse{}, err
	}
	endDate, err := toOptionalDate(item.EndDate)
	if err != nil {
		return generated.HealthMedicationCourse{}, err
	}
	return generated.HealthMedicationCourse{
		CreatedAt:  item.CreatedAt,
		DeletedAt:  item.DeletedAt,
		DosageText: item.DosageText,
		EndDate:    endDate,
		Id:         item.ID,
		Name:       item.Name,
		Source:     item.Source,
		StartDate:  startDate,
		UpdatedAt:  item.UpdatedAt,
	}, nil
}

func mapAnalyteSummaries(items []health.AnalyteSummary) []generated.HealthAnalyteSummary {
	out := make([]generated.HealthAnalyteSummary, 0, len(items))
	for _, item := range items {
		summary := generated.HealthAnalyteSummary{
			Slug:   generated.HealthAnalyteSlug(item.Slug),
			Latest: toGeneratedLabResultWithCollection(item.Latest),
		}
		if item.Previous != nil {
			previous := toGeneratedLabResultWithCollection(*item.Previous)
			summary.Previous = &previous
		}
		out = append(out, summary)
	}
	return out
}

func mapLabCollections(items []health.LabCollection) ([]generated.HealthLabCollection, error) {
	out := make([]generated.HealthLabCollection, 0, len(items))
	for _, item := range items {
		converted, err := toGeneratedLabCollection(item)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

func toGeneratedLabCollection(item health.LabCollection) (generated.HealthLabCollection, error) {
	panels := make([]generated.HealthLabPanel, 0, len(item.Panels))
	for _, panel := range item.Panels {
		panels = append(panels, generated.HealthLabPanel{
			CollectionId: panel.CollectionID,
			DisplayOrder: panel.DisplayOrder,
			Id:           panel.ID,
			PanelName:    panel.PanelName,
			Results:      mapLabResults(panel.Results),
		})
	}
	return generated.HealthLabCollection{
		CollectedAt: item.CollectedAt,
		CreatedAt:   item.CreatedAt,
		Id:          item.ID,
		Panels:      panels,
		Source:      item.Source,
	}, nil
}

func mapLabResults(items []health.LabResult) []generated.HealthLabResult {
	out := make([]generated.HealthLabResult, 0, len(items))
	for _, item := range items {
		out = append(out, toGeneratedLabResult(item))
	}
	return out
}

func mapLabResultsWithCollection(items []health.LabResultWithCollection) []generated.HealthLabResultWithCollection {
	out := make([]generated.HealthLabResultWithCollection, 0, len(items))
	for _, item := range items {
		out = append(out, toGeneratedLabResultWithCollection(item))
	}
	return out
}

func toGeneratedLabResult(item health.LabResult) generated.HealthLabResult {
	body := generated.HealthLabResult{
		DisplayOrder: item.DisplayOrder,
		Id:           item.ID,
		PanelId:      item.PanelID,
		TestName:     item.TestName,
		Units:        item.Units,
		ValueText:    item.ValueText,
	}
	if item.CanonicalSlug != nil {
		slug := generated.HealthAnalyteSlug(*item.CanonicalSlug)
		body.CanonicalSlug = &slug
	}
	if item.ValueNumeric != nil {
		value := float32(*item.ValueNumeric)
		body.ValueNumeric = &value
	}
	body.RangeText = item.RangeText
	body.Flag = item.Flag
	return body
}

func toGeneratedLabResultWithCollection(item health.LabResultWithCollection) generated.HealthLabResultWithCollection {
	body := generated.HealthLabResultWithCollection{
		CollectedAt:  item.CollectedAt,
		CollectionId: item.CollectionID,
		DisplayOrder: item.DisplayOrder,
		Id:           item.ID,
		PanelId:      item.PanelID,
		PanelName:    item.PanelName,
		TestName:     item.TestName,
		Units:        item.Units,
		ValueText:    item.ValueText,
	}
	if item.CanonicalSlug != nil {
		slug := generated.HealthAnalyteSlug(*item.CanonicalSlug)
		body.CanonicalSlug = &slug
	}
	if item.ValueNumeric != nil {
		value := float32(*item.ValueNumeric)
		body.ValueNumeric = &value
	}
	body.RangeText = item.RangeText
	body.Flag = item.Flag
	return body
}

func toDate(value string) (openapi_types.Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return openapi_types.Date{}, &health.DatabaseError{
			Message: "stored date is invalid",
			Cause:   err,
		}
	}
	return openapi_types.Date{Time: parsed}, nil
}

func toOptionalDate(value *string) (*openapi_types.Date, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := toDate(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (s *server) String() string {
	return fmt.Sprintf("%T", s)
}
