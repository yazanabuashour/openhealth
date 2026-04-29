package runner

import (
	"context"
	"fmt"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

func RunLabTask(ctx context.Context, config client.LocalConfig, request LabTaskRequest) (LabTaskResult, error) {
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

	api, err := client.OpenLocal(config)
	if err != nil {
		return LabTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case LabTaskActionRecord:
		return runLabRecord(ctx, api, normalized)
	case LabTaskActionCorrect:
		return runLabCorrect(ctx, api, normalized)
	case LabTaskActionPatch:
		return runLabPatch(ctx, api, normalized)
	case LabTaskActionDelete:
		return runLabDelete(ctx, api, normalized)
	case LabTaskActionList:
		return runLabList(ctx, api, normalized)
	default:
		return LabTaskResult{}, fmt.Errorf("unsupported lab task action %q", normalized.Action)
	}
}
