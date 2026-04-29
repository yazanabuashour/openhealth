package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	taskActionValidate = "validate"

	listModeLatest  = "latest"
	listModeHistory = "history"
	listModeRange   = "range"
)

type taskListRequest struct {
	ListMode string
	FromDate string
	ToDate   string
	Limit    int
}

type normalizedTaskList struct {
	ListMode string
	From     *time.Time
	To       *time.Time
	Limit    int
}

func withService[T any](
	ctx context.Context,
	config localruntime.Config,
	run func(context.Context, health.Service) (T, error),
) (T, error) {
	session, err := localruntime.Open(config)
	if err != nil {
		var zero T
		return zero, err
	}
	defer func() {
		_ = session.Close()
	}()
	return run(ctx, session.Service)
}

func normalizeTaskListRequest(request taskListRequest, domain string) (normalizedTaskList, string) {
	normalized := normalizedTaskList{
		ListMode: request.ListMode,
		Limit:    request.Limit,
	}
	if request.Limit < 0 {
		return normalizedTaskList{}, "limit must be greater than or equal to 0"
	}
	if normalized.ListMode == "" {
		normalized.ListMode = listModeHistory
	}
	switch normalized.ListMode {
	case listModeLatest:
		normalized.Limit = 1
	case listModeHistory:
		if normalized.Limit == 0 {
			normalized.Limit = 25
		}
	case listModeRange:
		if request.FromDate == "" || request.ToDate == "" {
			return normalizedTaskList{}, "from_date and to_date are required for range"
		}
		from, rejection := parseDateOnly(request.FromDate)
		if rejection != "" {
			return normalizedTaskList{}, rejection
		}
		toDate, rejection := parseDateOnly(request.ToDate)
		if rejection != "" {
			return normalizedTaskList{}, rejection
		}
		toEnd := toDate.Add(24*time.Hour - time.Nanosecond)
		normalized.From = &from
		normalized.To = &toEnd
	default:
		return normalizedTaskList{}, fmt.Sprintf("unsupported %s list mode %q", domain, normalized.ListMode)
	}
	return normalized, ""
}

func rejectNegativeLimit(limit int) string {
	if limit < 0 {
		return "limit must be greater than or equal to 0"
	}
	return ""
}

func limitPointer(limit int) *int {
	if limit == 0 {
		return nil
	}
	return &limit
}

func parseDateOnly(value string) (time.Time, string) {
	if value == "" {
		return time.Time{}, "date must be YYYY-MM-DD"
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil || parsed.Format(time.DateOnly) != value {
		return time.Time{}, "date must be YYYY-MM-DD"
	}
	return parsed, ""
}
