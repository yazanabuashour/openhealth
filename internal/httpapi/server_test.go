package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/yazanabuashour/openhealth/client"
	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/httpapi"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
	"github.com/yazanabuashour/openhealth/internal/testutil"
)

func TestServerSupportsGeneratedClientAndErrorEnvelopes(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	testutil.MustExec(t, db, `
INSERT INTO health_weight_entry (recorded_at, value, unit, source, source_record_hash, note, created_at, updated_at)
VALUES (?, ?, 'lb', 'test', 'weight-a', NULL, ?, ?)
`,
		"2026-03-28T13:15:00Z", 150.2, "2026-03-28T13:15:00Z", "2026-03-28T13:15:00Z",
	)

	repo := sqlite.NewRepository(db)
	service := health.NewService(repo)
	server := httptest.NewServer(httpapi.NewHandler(service))
	defer server.Close()

	api, err := client.NewDefault(server.URL)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	summary, err := api.GetHealthSummaryWithResponse(ctx)
	if err != nil {
		t.Fatalf("summary request: %v", err)
	}
	if summary.JSON200 == nil {
		t.Fatalf("summary status = %s", summary.Status())
	}

	weights, err := api.ListHealthWeightWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("list weight request: %v", err)
	}
	if weights.JSON200 == nil || len(weights.JSON200.Items) != 1 {
		t.Fatalf("weight response = %#v", weights)
	}

	bp, err := api.CreateHealthBloodPressureWithResponse(ctx, client.CreateHealthBloodPressureJSONRequestBody{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Systolic:   118,
		Diastolic:  76,
		Pulse:      intPointer(64),
	})
	if err != nil {
		t.Fatalf("create blood pressure: %v", err)
	}
	if bp.JSON201 == nil || bp.JSON201.Systolic != 118 {
		t.Fatalf("blood pressure response = %#v", bp)
	}
	replacedBP, err := api.ReplaceHealthBloodPressureWithResponse(ctx, bp.JSON201.Id, client.ReplaceHealthBloodPressureJSONRequestBody{
		RecordedAt: time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC),
		Systolic:   119,
		Diastolic:  77,
	})
	if err != nil {
		t.Fatalf("replace blood pressure: %v", err)
	}
	if replacedBP.JSON200 == nil || replacedBP.JSON200.Pulse != nil {
		t.Fatalf("replace blood pressure response = %#v", replacedBP)
	}
	deletedBP, err := api.DeleteHealthBloodPressureWithResponse(ctx, bp.JSON201.Id)
	if err != nil {
		t.Fatalf("delete blood pressure: %v", err)
	}
	if deletedBP.JSON200 == nil || !deletedBP.JSON200.Success {
		t.Fatalf("delete blood pressure response = %#v", deletedBP)
	}

	med, err := api.CreateHealthMedicationWithResponse(ctx, client.CreateHealthMedicationJSONRequestBody{
		Name:      "Levothyroxine",
		StartDate: openapi_types.Date{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("create medication: %v", err)
	}
	if med.JSON201 == nil || med.JSON201.Name != "Levothyroxine" {
		t.Fatalf("medication response = %#v", med)
	}

	missingStartResponse, err := http.Post(
		server.URL+"/api/v1/health/medications",
		"application/json",
		strings.NewReader(`{"name":"Missing date"}`),
	)
	if err != nil {
		t.Fatalf("create medication without start date: %v", err)
	}
	defer func() {
		_ = missingStartResponse.Body.Close()
	}()

	var missingStart struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(missingStartResponse.Body).Decode(&missingStart); err != nil {
		t.Fatalf("decode missing start date error: %v", err)
	}
	if missingStartResponse.StatusCode != http.StatusBadRequest ||
		missingStart.Error.Code != "VALIDATION_ERROR" ||
		missingStart.Error.Message != "start_date is required" {
		t.Fatalf("unexpected missing start date response: status=%d body=%#v", missingStartResponse.StatusCode, missingStart)
	}

	replacedMed, err := api.ReplaceHealthMedicationWithResponse(ctx, med.JSON201.Id, client.ReplaceHealthMedicationJSONRequestBody{
		Name:      "Levothyroxine",
		StartDate: openapi_types.Date{Time: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("replace medication: %v", err)
	}
	if replacedMed.JSON200 == nil || replacedMed.JSON200.StartDate.Format(time.DateOnly) != "2026-01-02" {
		t.Fatalf("replace medication response = %#v", replacedMed)
	}

	replaceMissingStartRequest, err := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/health/medications/"+strconv.Itoa(med.JSON201.Id),
		strings.NewReader(`{"name":"Levothyroxine"}`),
	)
	if err != nil {
		t.Fatalf("create missing start date replace request: %v", err)
	}
	replaceMissingStartRequest.Header.Set("Content-Type", "application/json")
	replaceMissingStartResponse, err := http.DefaultClient.Do(replaceMissingStartRequest)
	if err != nil {
		t.Fatalf("replace medication without start date: %v", err)
	}
	defer func() {
		_ = replaceMissingStartResponse.Body.Close()
	}()

	var replaceMissingStart struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(replaceMissingStartResponse.Body).Decode(&replaceMissingStart); err != nil {
		t.Fatalf("decode replace missing start date error: %v", err)
	}
	if replaceMissingStartResponse.StatusCode != http.StatusBadRequest ||
		replaceMissingStart.Error.Code != "VALIDATION_ERROR" ||
		replaceMissingStart.Error.Message != "start_date is required" {
		t.Fatalf("unexpected replace missing start date response: status=%d body=%#v", replaceMissingStartResponse.StatusCode, replaceMissingStart)
	}

	deletedMed, err := api.DeleteHealthMedicationWithResponse(ctx, med.JSON201.Id)
	if err != nil {
		t.Fatalf("delete medication: %v", err)
	}
	if deletedMed.JSON200 == nil || !deletedMed.JSON200.Success {
		t.Fatalf("delete medication response = %#v", deletedMed)
	}

	slug := client.Glucose
	lab, err := api.CreateHealthLabCollectionWithResponse(ctx, client.CreateHealthLabCollectionJSONRequestBody{
		CollectedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Panels: []client.HealthLabPanelWrite{
			{
				PanelName: "Metabolic",
				Results: []client.HealthLabResultWrite{
					{
						TestName:      "Glucose",
						CanonicalSlug: &slug,
						ValueText:     "89",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create lab collection: %v", err)
	}
	if lab.JSON201 == nil || len(lab.JSON201.Panels) != 1 || lab.JSON201.UpdatedAt.IsZero() {
		t.Fatalf("lab collection response = %#v", lab)
	}
	replacedLab, err := api.ReplaceHealthLabCollectionWithResponse(ctx, lab.JSON201.Id, client.ReplaceHealthLabCollectionJSONRequestBody{
		CollectedAt: time.Date(2026, 4, 16, 8, 0, 0, 0, time.UTC),
		Panels: []client.HealthLabPanelWrite{
			{
				PanelName: "Thyroid",
				Results: []client.HealthLabResultWrite{
					{
						TestName:  "TSH",
						ValueText: "3.1",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("replace lab collection: %v", err)
	}
	if replacedLab.JSON200 == nil || replacedLab.JSON200.Panels[0].PanelName != "Thyroid" {
		t.Fatalf("replace lab collection response = %#v", replacedLab)
	}
	deletedLab, err := api.DeleteHealthLabCollectionWithResponse(ctx, lab.JSON201.Id)
	if err != nil {
		t.Fatalf("delete lab collection: %v", err)
	}
	if deletedLab.JSON200 == nil || !deletedLab.JSON200.Success {
		t.Fatalf("delete lab collection response = %#v", deletedLab)
	}

	response, err := http.Post(
		server.URL+"/api/v1/health/weight",
		"application/json",
		strings.NewReader(`{"recordedAt":"2026-03-28T08:15:00-05:00","value":150.2,"unit":"kg"}`),
	)
	if err != nil {
		t.Fatalf("raw create request: %v", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	var badRequest struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&badRequest); err != nil {
		t.Fatalf("decode validation error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest || badRequest.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("unexpected validation response: status=%d body=%#v", response.StatusCode, badRequest)
	}

	request, err := http.NewRequest(http.MethodPatch, server.URL+"/api/v1/health/weight/999", strings.NewReader(`{"value":149.8}`))
	if err != nil {
		t.Fatalf("create patch request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	notFoundResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("patch request: %v", err)
	}
	defer func() {
		_ = notFoundResponse.Body.Close()
	}()

	var notFound struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(notFoundResponse.Body).Decode(&notFound); err != nil {
		t.Fatalf("decode not found response: %v", err)
	}
	if notFoundResponse.StatusCode != http.StatusNotFound || notFound.Error.Code != "NOT_FOUND" {
		t.Fatalf("unexpected not found response: status=%d body=%#v", notFoundResponse.StatusCode, notFound)
	}
}

func intPointer(value int) *int {
	return &value
}
