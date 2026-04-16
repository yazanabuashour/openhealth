package client_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func TestOpenLocalSupportsGeneratedClientWithoutNetwork(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	api, err := client.OpenLocal(client.LocalConfig{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthResponse, err := api.HealthWithResponse(ctx)
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	if healthResponse.JSON200 == nil || !healthResponse.JSON200.Ok {
		t.Fatalf("unexpected health response: %#v", healthResponse)
	}

	recordedAt := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	createResponse, err := api.CreateHealthWeightWithResponse(ctx, client.CreateHealthWeightJSONRequestBody{
		RecordedAt: recordedAt,
		Unit:       client.CreateHealthWeightRequestUnitLb,
		Value:      149.4,
	})
	if err != nil {
		t.Fatalf("create weight request: %v", err)
	}
	if createResponse.JSON201 == nil {
		t.Fatalf("unexpected create response: %#v", createResponse)
	}

	listResponse, err := api.ListHealthWeightWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("list weight request: %v", err)
	}
	if listResponse.JSON200 == nil || len(listResponse.JSON200.Items) != 1 {
		t.Fatalf("unexpected list response: %#v", listResponse)
	}
	if listResponse.JSON200.Items[0].Value != 149.4 {
		t.Fatalf("weight value = %v, want %v", listResponse.JSON200.Items[0].Value, 149.4)
	}

	databasePath := filepath.Join(dataDir, "openhealth.db")
	if api.Paths.DatabasePath != databasePath {
		t.Fatalf("databasePath = %q, want %q", api.Paths.DatabasePath, databasePath)
	}
	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("stat database path: %v", err)
	}
}

func TestOpenLocalPersistsDataAcrossSessions(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	recordedAt := time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC)

	api, err := client.OpenLocal(client.LocalConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("open first local client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createResponse, err := api.CreateHealthWeightWithResponse(ctx, client.CreateHealthWeightJSONRequestBody{
		RecordedAt: recordedAt,
		Unit:       client.CreateHealthWeightRequestUnitLb,
		Value:      152.1,
	})
	if err != nil {
		t.Fatalf("create first session weight request: %v", err)
	}
	if createResponse.JSON201 == nil {
		t.Fatalf("unexpected create response: %#v", createResponse)
	}
	if err := api.Close(); err != nil {
		t.Fatalf("close first local client: %v", err)
	}

	api, err = client.OpenLocal(client.LocalConfig{DataDir: dataDir})
	if err != nil {
		t.Fatalf("reopen local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close reopened local client: %v", closeErr)
		}
	}()

	listResponse, err := api.ListHealthWeightWithResponse(ctx, nil)
	if err != nil {
		t.Fatalf("list second session weights: %v", err)
	}
	if listResponse.JSON200 == nil || len(listResponse.JSON200.Items) != 1 {
		t.Fatalf("unexpected reopened list response: %#v", listResponse)
	}
	if listResponse.JSON200.Items[0].Value != 152.1 {
		t.Fatalf("weight value after reopen = %v, want %v", listResponse.JSON200.Items[0].Value, 152.1)
	}
}

func TestLocalClientWeightHelpers(t *testing.T) {
	t.Parallel()

	api, err := client.OpenLocal(client.LocalConfig{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	march29 := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	created, err := api.UpsertWeight(ctx, client.WeightRecordInput{
		RecordedAt: march29,
		Value:      152.2,
		Unit:       client.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("upsert weight: %v", err)
	}
	if created.Status != client.WeightWriteStatusCreated || created.Entry.Value != 152.2 {
		t.Fatalf("created weight = %#v", created)
	}

	again, err := api.UpsertWeight(ctx, client.WeightRecordInput{
		RecordedAt: march29,
		Value:      152.2,
		Unit:       client.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("repeat upsert weight: %v", err)
	}
	if again.Status != client.WeightWriteStatusAlreadyExists || again.Entry.ID != created.Entry.ID {
		t.Fatalf("repeat weight = %#v, want already_exists id %d", again, created.Entry.ID)
	}

	march30 := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	recorded, err := api.RecordWeight(ctx, client.WeightRecordInput{
		RecordedAt: march30,
		Value:      151.6,
	})
	if err != nil {
		t.Fatalf("record weight with default unit: %v", err)
	}
	if recorded.Unit != client.WeightUnitLb {
		t.Fatalf("recorded unit = %q, want %q", recorded.Unit, client.WeightUnitLb)
	}

	weights, err := api.ListWeights(ctx, client.WeightListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("list weights: %v", err)
	}
	if len(weights) != 2 {
		t.Fatalf("weight count = %d, want 2", len(weights))
	}
	if weights[0].Value != 151.6 || !weights[0].RecordedAt.Equal(march30) {
		t.Fatalf("newest weight = %#v, want 151.6 on March 30", weights[0])
	}

	latest, err := api.LatestWeight(ctx)
	if err != nil {
		t.Fatalf("latest weight: %v", err)
	}
	if latest == nil || latest.ID != weights[0].ID {
		t.Fatalf("latest = %#v, want id %d", latest, weights[0].ID)
	}
}

func TestOpenLocalClosesRequestBodies(t *testing.T) {
	t.Parallel()

	api, err := client.OpenLocal(client.LocalConfig{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("open local client: %v", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			t.Fatalf("close local client: %v", closeErr)
		}
	}()

	body := &trackingReadCloser{
		data: []byte(`{"recordedAt":"2026-04-15T12:00:00Z","value":149.4,"unit":"lb"}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := api.CreateHealthWeightWithBody(ctx, "application/json", body)
	if err != nil {
		t.Fatalf("create weight with body request: %v", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if !body.closed {
		t.Fatal("expected request body to be closed")
	}
}

type trackingReadCloser struct {
	data   []byte
	closed bool
}

func (r *trackingReadCloser) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}
