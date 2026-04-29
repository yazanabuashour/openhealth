package runner

import (
	"context"
	"fmt"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

func runLabRecord(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	result := LabTaskResult{}
	for _, collection := range request.Collections {
		existing, err := matchingLabCollections(ctx, api, normalizedLabTarget{Date: &collection.CollectedAt})
		if err != nil {
			return LabTaskResult{}, err
		}
		if duplicate, ok := matchingLabCollection(existing, collection); ok {
			result.Writes = append(result.Writes, labCollectionWrite(duplicate, "already_exists"))
			continue
		}
		written, err := api.CreateLabCollection(ctx, toClientLabCollectionInput(collection))
		if err != nil {
			return LabTaskResult{}, err
		}
		result.Writes = append(result.Writes, labCollectionWrite(written, "created"))
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d lab entries", len(entries))
	return result, nil
}

func runLabCorrect(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	collection := request.Collection
	if collection.Note == nil {
		collection.Note = target.Note
	}
	written, err := api.ReplaceLabCollection(ctx, target.ID, toClientLabCollectionInput(collection))
	if err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabPatch(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	collection := normalizedLabCollectionFromClient(target)
	for _, update := range request.ResultUpdates {
		if rejection := applyLabResultUpdate(&collection, update); rejection != "" {
			return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
		}
	}
	written, err := api.ReplaceLabCollection(ctx, target.ID, toClientLabCollectionInput(collection))
	if err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabDelete(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	target, rejection, err := labTarget(ctx, api, request.Target)
	if err != nil {
		return LabTaskResult{}, err
	}
	if rejection != "" {
		return LabTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := api.DeleteLabCollection(ctx, target.ID); err != nil {
		return LabTaskResult{}, err
	}
	entries, err := listLabEntries(ctx, api, normalizedLabTaskRequest{Limit: 100})
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Writes:  []LabCollectionWrite{labCollectionWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d lab entries", len(entries)),
	}, nil
}

func runLabList(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) (LabTaskResult, error) {
	entries, err := listLabEntries(ctx, api, request)
	if err != nil {
		return LabTaskResult{}, err
	}
	return LabTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d lab entries", len(entries)),
	}, nil
}

func labTarget(ctx context.Context, api *client.LocalClient, target normalizedLabTarget) (client.LabCollection, string, error) {
	matches, err := matchingLabCollections(ctx, api, target)
	if err != nil {
		return client.LabCollection{}, "", err
	}
	switch len(matches) {
	case 0:
		return client.LabCollection{}, "no matching lab collection", nil
	case 1:
		return matches[0], "", nil
	default:
		return client.LabCollection{}, "multiple matching lab collections; target is ambiguous", nil
	}
}

func matchingLabCollections(ctx context.Context, api *client.LocalClient, target normalizedLabTarget) ([]client.LabCollection, error) {
	items, err := api.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}
	matches := []client.LabCollection{}
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if target.Date != nil && item.CollectedAt.Format(time.DateOnly) == target.Date.Format(time.DateOnly) {
			matches = append(matches, item)
		}
	}
	return matches, nil
}

func listLabEntries(ctx context.Context, api *client.LocalClient, request normalizedLabTaskRequest) ([]LabCollectionEntry, error) {
	items, err := api.ListLabCollections(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LabCollectionEntry, 0, len(items))
	for _, item := range items {
		if request.From != nil && item.CollectedAt.Before(*request.From) {
			continue
		}
		if request.To != nil && item.CollectedAt.After(*request.To) {
			continue
		}
		entry, ok := labCollectionEntry(item, request.AnalyteSlug)
		if !ok {
			continue
		}
		out = append(out, entry)
		if request.Limit > 0 && len(out) >= request.Limit {
			break
		}
	}
	return out, nil
}
