package runner

import (
	"context"
	"fmt"
	"strings"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

const (
	MedicationTaskActionRecord   = "record_medications"
	MedicationTaskActionCorrect  = "correct_medication"
	MedicationTaskActionDelete   = "delete_medication"
	MedicationTaskActionList     = "list_medications"
	MedicationTaskActionValidate = "validate"

	MedicationStatusActive = "active"
	MedicationStatusAll    = "all"
)

type MedicationTaskRequest struct {
	Action      string            `json:"action"`
	Medications []MedicationInput `json:"medications,omitempty"`
	Medication  *MedicationInput  `json:"medication,omitempty"`
	Target      *MedicationTarget `json:"target,omitempty"`
	Status      string            `json:"status,omitempty"`
}

type MedicationInput struct {
	Name       string  `json:"name"`
	DosageText *string `json:"dosage_text,omitempty"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

type MedicationTarget struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
}

type MedicationTaskResult struct {
	Rejected        bool              `json:"rejected"`
	RejectionReason string            `json:"rejection_reason,omitempty"`
	Writes          []MedicationWrite `json:"writes,omitempty"`
	Entries         []MedicationEntry `json:"entries,omitempty"`
	Summary         string            `json:"summary"`
}

type MedicationWrite struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	DosageText *string `json:"dosage_text,omitempty"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
	Status     string  `json:"status"`
}

type MedicationEntry struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	DosageText *string `json:"dosage_text,omitempty"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

func RunMedicationTask(ctx context.Context, config client.LocalConfig, request MedicationTaskRequest) (MedicationTaskResult, error) {
	normalized, rejection := normalizeMedicationTaskRequest(request)
	if rejection != "" {
		return MedicationTaskResult{
			Rejected:        true,
			RejectionReason: rejection,
			Summary:         rejection,
		}, nil
	}

	if normalized.Action == MedicationTaskActionValidate {
		return MedicationTaskResult{Summary: "valid"}, nil
	}

	api, err := client.OpenLocal(config)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	defer func() {
		_ = api.Close()
	}()

	switch normalized.Action {
	case MedicationTaskActionRecord:
		return runMedicationRecord(ctx, api, normalized)
	case MedicationTaskActionCorrect:
		return runMedicationCorrect(ctx, api, normalized)
	case MedicationTaskActionDelete:
		return runMedicationDelete(ctx, api, normalized)
	case MedicationTaskActionList:
		return runMedicationList(ctx, api, normalized)
	default:
		return MedicationTaskResult{}, fmt.Errorf("unsupported medication task action %q", normalized.Action)
	}
}

type normalizedMedicationTaskRequest struct {
	Action      string
	Medications []normalizedMedicationInput
	Medication  normalizedMedicationInput
	Target      normalizedMedicationTarget
	Status      client.MedicationStatus
}

type normalizedMedicationInput struct {
	Name       string
	DosageText *string
	StartDate  string
	EndDate    *string
}

type normalizedMedicationTarget struct {
	ID        int
	Name      string
	StartDate string
}

func normalizeMedicationTaskRequest(request MedicationTaskRequest) (normalizedMedicationTaskRequest, string) {
	action := request.Action
	if action == "" {
		action = MedicationTaskActionValidate
	}

	normalized := normalizedMedicationTaskRequest{
		Action: action,
	}

	switch action {
	case MedicationTaskActionValidate:
		for _, medication := range request.Medications {
			if _, rejection := normalizeMedicationInput(medication); rejection != "" {
				return normalizedMedicationTaskRequest{}, rejection
			}
		}
		if request.Medication != nil {
			if _, rejection := normalizeMedicationInput(*request.Medication); rejection != "" {
				return normalizedMedicationTaskRequest{}, rejection
			}
		}
		if request.Target != nil {
			if _, rejection := normalizeMedicationTarget(*request.Target); rejection != "" {
				return normalizedMedicationTaskRequest{}, rejection
			}
		}
		if _, rejection := normalizeMedicationStatus(request.Status); rejection != "" {
			return normalizedMedicationTaskRequest{}, rejection
		}
		return normalized, ""
	case MedicationTaskActionRecord:
		if len(request.Medications) == 0 {
			return normalizedMedicationTaskRequest{}, "medications are required"
		}
		for _, medication := range request.Medications {
			parsed, rejection := normalizeMedicationInput(medication)
			if rejection != "" {
				return normalizedMedicationTaskRequest{}, rejection
			}
			normalized.Medications = append(normalized.Medications, parsed)
		}
		return normalized, ""
	case MedicationTaskActionCorrect:
		if request.Target == nil {
			return normalizedMedicationTaskRequest{}, "target is required"
		}
		target, rejection := normalizeMedicationTarget(*request.Target)
		if rejection != "" {
			return normalizedMedicationTaskRequest{}, rejection
		}
		if request.Medication == nil {
			return normalizedMedicationTaskRequest{}, "medication is required"
		}
		medication, rejection := normalizeMedicationInput(*request.Medication)
		if rejection != "" {
			return normalizedMedicationTaskRequest{}, rejection
		}
		normalized.Target = target
		normalized.Medication = medication
		return normalized, ""
	case MedicationTaskActionDelete:
		if request.Target == nil {
			return normalizedMedicationTaskRequest{}, "target is required"
		}
		target, rejection := normalizeMedicationTarget(*request.Target)
		if rejection != "" {
			return normalizedMedicationTaskRequest{}, rejection
		}
		normalized.Target = target
		return normalized, ""
	case MedicationTaskActionList:
		status, rejection := normalizeMedicationStatus(request.Status)
		if rejection != "" {
			return normalizedMedicationTaskRequest{}, rejection
		}
		normalized.Status = status
		return normalized, ""
	default:
		return normalizedMedicationTaskRequest{}, fmt.Sprintf("unsupported medication task action %q", action)
	}
}

func normalizeMedicationInput(input MedicationInput) (normalizedMedicationInput, string) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return normalizedMedicationInput{}, "name is required"
	}
	if input.DosageText != nil {
		dosageText := strings.TrimSpace(*input.DosageText)
		if dosageText == "" {
			return normalizedMedicationInput{}, "dosage_text must not be empty"
		}
		input.DosageText = &dosageText
	}
	if _, rejection := parseDateOnly(input.StartDate); rejection != "" {
		return normalizedMedicationInput{}, "start_date must be YYYY-MM-DD"
	}
	if input.EndDate != nil {
		endDate := strings.TrimSpace(*input.EndDate)
		if endDate == "" {
			return normalizedMedicationInput{}, "end_date must be YYYY-MM-DD"
		}
		parsedEnd, rejection := parseDateOnly(endDate)
		if rejection != "" {
			return normalizedMedicationInput{}, "end_date must be YYYY-MM-DD"
		}
		parsedStart, _ := parseDateOnly(input.StartDate)
		if parsedEnd.Before(parsedStart) {
			return normalizedMedicationInput{}, "end_date must be on or after start_date"
		}
		input.EndDate = &endDate
	}
	return normalizedMedicationInput{
		Name:       name,
		DosageText: input.DosageText,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}, ""
}

func normalizeMedicationTarget(target MedicationTarget) (normalizedMedicationTarget, string) {
	if target.ID < 0 {
		return normalizedMedicationTarget{}, "target id must be greater than 0"
	}
	if target.ID > 0 {
		return normalizedMedicationTarget{ID: target.ID}, ""
	}
	name := strings.TrimSpace(target.Name)
	if name == "" || target.StartDate == "" {
		return normalizedMedicationTarget{}, "target id or name and start_date are required"
	}
	if _, rejection := parseDateOnly(target.StartDate); rejection != "" {
		return normalizedMedicationTarget{}, "start_date must be YYYY-MM-DD"
	}
	return normalizedMedicationTarget{
		Name:      name,
		StartDate: target.StartDate,
	}, ""
}

func normalizeMedicationStatus(value string) (client.MedicationStatus, string) {
	if value == "" {
		return client.MedicationStatusActive, ""
	}
	switch value {
	case MedicationStatusActive:
		return client.MedicationStatusActive, ""
	case MedicationStatusAll:
		return client.MedicationStatusAll, ""
	default:
		return "", "status must be active or all"
	}
}

func runMedicationRecord(ctx context.Context, api *client.LocalClient, request normalizedMedicationTaskRequest) (MedicationTaskResult, error) {
	result := MedicationTaskResult{}
	for _, medication := range request.Medications {
		existing, err := matchingMedicationCourses(ctx, api, normalizedMedicationTarget{
			Name:      medication.Name,
			StartDate: medication.StartDate,
		})
		if err != nil {
			return MedicationTaskResult{}, err
		}
		if len(existing) > 0 {
			if medicationMatches(existing[0], medication) && allMedicationMatches(existing, medication) {
				result.Writes = append(result.Writes, medicationWrite(existing[0], "already_exists"))
				continue
			}
			reason := fmt.Sprintf("medication already exists for %s starting %s; use correct_medication", medication.Name, medication.StartDate)
			return MedicationTaskResult{
				Rejected:        true,
				RejectionReason: reason,
				Summary:         reason,
			}, nil
		}
		written, err := api.CreateMedicationCourse(ctx, client.MedicationCourseInput(medication))
		if err != nil {
			return MedicationTaskResult{}, err
		}
		result.Writes = append(result.Writes, medicationWrite(written, "created"))
	}
	entries, err := listMedicationEntries(ctx, api, client.MedicationStatusAll)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	result.Entries = entries
	result.Summary = fmt.Sprintf("stored %d medication entries", len(entries))
	return result, nil
}

func runMedicationCorrect(ctx context.Context, api *client.LocalClient, request normalizedMedicationTaskRequest) (MedicationTaskResult, error) {
	target, rejection, err := medicationTarget(ctx, api, request.Target)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	if rejection != "" {
		return MedicationTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	written, err := api.ReplaceMedicationCourse(ctx, target.ID, client.MedicationCourseInput(request.Medication))
	if err != nil {
		return MedicationTaskResult{}, err
	}
	entries, err := listMedicationEntries(ctx, api, client.MedicationStatusAll)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	return MedicationTaskResult{
		Writes:  []MedicationWrite{medicationWrite(written, "updated")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d medication entries", len(entries)),
	}, nil
}

func runMedicationDelete(ctx context.Context, api *client.LocalClient, request normalizedMedicationTaskRequest) (MedicationTaskResult, error) {
	target, rejection, err := medicationTarget(ctx, api, request.Target)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	if rejection != "" {
		return MedicationTaskResult{Rejected: true, RejectionReason: rejection, Summary: rejection}, nil
	}
	if err := api.DeleteMedicationCourse(ctx, target.ID); err != nil {
		return MedicationTaskResult{}, err
	}
	entries, err := listMedicationEntries(ctx, api, client.MedicationStatusAll)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	return MedicationTaskResult{
		Writes:  []MedicationWrite{medicationWrite(target, "deleted")},
		Entries: entries,
		Summary: fmt.Sprintf("stored %d medication entries", len(entries)),
	}, nil
}

func runMedicationList(ctx context.Context, api *client.LocalClient, request normalizedMedicationTaskRequest) (MedicationTaskResult, error) {
	entries, err := listMedicationEntries(ctx, api, request.Status)
	if err != nil {
		return MedicationTaskResult{}, err
	}
	return MedicationTaskResult{
		Entries: entries,
		Summary: fmt.Sprintf("returned %d medication entries", len(entries)),
	}, nil
}

func medicationTarget(ctx context.Context, api *client.LocalClient, target normalizedMedicationTarget) (client.MedicationCourse, string, error) {
	matches, err := matchingMedicationCourses(ctx, api, target)
	if err != nil {
		return client.MedicationCourse{}, "", err
	}
	switch len(matches) {
	case 0:
		return client.MedicationCourse{}, "no matching medication", nil
	case 1:
		return matches[0], "", nil
	default:
		return client.MedicationCourse{}, "multiple matching medications; target is ambiguous", nil
	}
}

func matchingMedicationCourses(ctx context.Context, api *client.LocalClient, target normalizedMedicationTarget) ([]client.MedicationCourse, error) {
	items, err := api.ListMedicationCourses(ctx, client.MedicationListOptions{Status: client.MedicationStatusAll})
	if err != nil {
		return nil, err
	}
	matches := []client.MedicationCourse{}
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if strings.EqualFold(item.Name, target.Name) && item.StartDate == target.StartDate {
			matches = append(matches, item)
		}
	}
	return matches, nil
}

func listMedicationEntries(ctx context.Context, api *client.LocalClient, status client.MedicationStatus) ([]MedicationEntry, error) {
	items, err := api.ListMedicationCourses(ctx, client.MedicationListOptions{Status: status})
	if err != nil {
		return nil, err
	}
	out := make([]MedicationEntry, 0, len(items))
	for _, item := range items {
		out = append(out, medicationEntry(item))
	}
	return out, nil
}

func allMedicationMatches(items []client.MedicationCourse, medication normalizedMedicationInput) bool {
	for _, item := range items {
		if !medicationMatches(item, medication) {
			return false
		}
	}
	return true
}

func medicationMatches(item client.MedicationCourse, medication normalizedMedicationInput) bool {
	return strings.EqualFold(item.Name, medication.Name) &&
		equalStringPointer(item.DosageText, medication.DosageText) &&
		item.StartDate == medication.StartDate &&
		equalStringPointer(item.EndDate, medication.EndDate)
}

func medicationWrite(item client.MedicationCourse, status string) MedicationWrite {
	return MedicationWrite{
		ID:         item.ID,
		Name:       item.Name,
		DosageText: item.DosageText,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
		Status:     status,
	}
}

func medicationEntry(item client.MedicationCourse) MedicationEntry {
	return MedicationEntry{
		ID:         item.ID,
		Name:       item.Name,
		DosageText: item.DosageText,
		StartDate:  item.StartDate,
		EndDate:    item.EndDate,
	}
}

func equalStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
