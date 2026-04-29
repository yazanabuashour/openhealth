package runner

import (
	"testing"
	"time"
)

func TestNormalizeTaskListRequestSharedBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		request   taskListRequest
		domain    string
		wantMode  string
		wantLimit int
		wantFrom  string
		wantTo    string
		wantError string
	}{
		{
			name:      "default history",
			request:   taskListRequest{},
			domain:    "weight",
			wantMode:  listModeHistory,
			wantLimit: 25,
		},
		{
			name:      "latest forces limit one",
			request:   taskListRequest{ListMode: listModeLatest, Limit: 9},
			domain:    "sleep",
			wantMode:  listModeLatest,
			wantLimit: 1,
		},
		{
			name:      "range is inclusive through end date",
			request:   taskListRequest{ListMode: listModeRange, FromDate: "2026-03-29", ToDate: "2026-03-30"},
			domain:    "lab",
			wantMode:  listModeRange,
			wantFrom:  "2026-03-29T00:00:00Z",
			wantTo:    "2026-03-30T23:59:59.999999999Z",
			wantLimit: 0,
		},
		{
			name:      "negative limit rejected",
			request:   taskListRequest{Limit: -1},
			domain:    "imaging",
			wantError: "limit must be greater than or equal to 0",
		},
		{
			name:      "domain appears in unsupported mode rejection",
			request:   taskListRequest{ListMode: "recent"},
			domain:    "blood pressure",
			wantError: `unsupported blood pressure list mode "recent"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, rejection := normalizeTaskListRequest(tt.request, tt.domain)
			if rejection != tt.wantError {
				t.Fatalf("rejection = %q, want %q", rejection, tt.wantError)
			}
			if tt.wantError != "" {
				return
			}
			if got.ListMode != tt.wantMode || got.Limit != tt.wantLimit {
				t.Fatalf("got mode=%q limit=%d, want mode=%q limit=%d", got.ListMode, got.Limit, tt.wantMode, tt.wantLimit)
			}
			if tt.wantFrom != "" && got.From.Format(time.RFC3339Nano) != tt.wantFrom {
				t.Fatalf("from = %s, want %s", got.From.Format(time.RFC3339Nano), tt.wantFrom)
			}
			if tt.wantTo != "" && got.To.Format(time.RFC3339Nano) != tt.wantTo {
				t.Fatalf("to = %s, want %s", got.To.Format(time.RFC3339Nano), tt.wantTo)
			}
		})
	}
}
