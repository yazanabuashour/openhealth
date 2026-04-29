package client

import (
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

func historyFilterFromOptions(from *time.Time, to *time.Time, limitValue int) health.HistoryFilter {
	filter := health.HistoryFilter{From: from, To: to}
	if limitValue != 0 {
		limit := limitValue
		filter.Limit = &limit
	}
	return filter
}
