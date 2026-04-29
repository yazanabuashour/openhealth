package runner

import (
	"context"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

func RunLabTask(ctx context.Context, config localruntime.Config, request LabTaskRequest) (LabTaskResult, error) {
	normalized, rejection := normalizeLabTaskRequest(request)
	if rejection != "" {
		return LabTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == LabTaskActionValidate {
		return LabTaskResult{Summary: "valid"}, nil
	}

	return withService(ctx, config, func(ctx context.Context, service health.Service) (LabTaskResult, error) {
		switch normalized.Action {
		case LabTaskActionRecord:
			return runLabRecord(ctx, service, normalized)
		case LabTaskActionCorrect:
			return runLabCorrect(ctx, service, normalized)
		case LabTaskActionPatch:
			return runLabPatch(ctx, service, normalized)
		case LabTaskActionDelete:
			return runLabDelete(ctx, service, normalized)
		case LabTaskActionList:
			return runLabList(ctx, service, normalized)
		default:
			return LabTaskResult{}, fmt.Errorf("unsupported lab task action %q", normalized.Action)
		}
	})
}
