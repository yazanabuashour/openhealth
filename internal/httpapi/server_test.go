package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
