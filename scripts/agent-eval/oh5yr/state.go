package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/yazanabuashour/openhealth/client"
	storagesqlite "github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

func listRawWeights(dbPath string) ([]weightState, error) {
	db, err := storagesqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(context.Background(), `
SELECT recorded_at, value, unit, note
FROM health_weight_entry
WHERE deleted_at IS NULL
ORDER BY recorded_at DESC, id DESC
`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	states := []weightState{}
	for rows.Next() {
		var state weightState
		var note sql.NullString
		if err := rows.Scan(&state.Date, &state.Value, &state.Unit, &note); err != nil {
			return nil, err
		}
		state.Value = roundWeight(state.Value)
		if note.Valid {
			state.Note = &note.String
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

func weightStates(weights []client.WeightEntry) []weightState {
	states := make([]weightState, 0, len(weights))
	for _, weight := range weights {
		states = append(states, weightState{
			Date:  weight.RecordedAt.Format(time.DateOnly),
			Value: roundWeight(weight.Value),
			Unit:  string(weight.Unit),
			Note:  weight.Note,
		})
	}
	return states
}

func bloodPressureStates(readings []client.BloodPressureEntry) []bloodPressureState {
	states := make([]bloodPressureState, 0, len(readings))
	for _, reading := range readings {
		states = append(states, bloodPressureState{
			Date:      reading.RecordedAt.Format(time.DateOnly),
			Systolic:  reading.Systolic,
			Diastolic: reading.Diastolic,
			Pulse:     reading.Pulse,
			Note:      reading.Note,
		})
	}
	return states
}

func medicationStates(medications []client.MedicationCourse) []medicationState {
	states := make([]medicationState, 0, len(medications))
	for _, medication := range medications {
		states = append(states, medicationState{
			Name:       medication.Name,
			DosageText: medication.DosageText,
			StartDate:  medication.StartDate,
			EndDate:    medication.EndDate,
			Note:       medication.Note,
		})
	}
	return states
}

func labCollectionStates(collections []client.LabCollection) []labCollectionState {
	states := make([]labCollectionState, 0, len(collections))
	for _, collection := range collections {
		state := labCollectionState{
			Date: collection.CollectedAt.Format(time.DateOnly),
			Note: collection.Note,
		}
		for _, panel := range collection.Panels {
			for _, result := range panel.Results {
				var slug *string
				if result.CanonicalSlug != nil {
					value := string(*result.CanonicalSlug)
					slug = &value
				}
				state.Results = append(state.Results, labResultState{
					TestName:      result.TestName,
					CanonicalSlug: slug,
					ValueText:     result.ValueText,
					ValueNumeric:  result.ValueNumeric,
					Units:         result.Units,
					Notes:         append([]string(nil), result.Notes...),
				})
			}
		}
		states = append(states, state)
	}
	return states
}

func bodyCompositionStates(records []client.BodyCompositionEntry) []bodyCompositionState {
	states := make([]bodyCompositionState, 0, len(records))
	for _, record := range records {
		var weightUnit *string
		if record.WeightUnit != nil {
			value := string(*record.WeightUnit)
			weightUnit = &value
		}
		states = append(states, bodyCompositionState{
			Date:           record.RecordedAt.Format(time.DateOnly),
			BodyFatPercent: record.BodyFatPercent,
			WeightValue:    record.WeightValue,
			WeightUnit:     weightUnit,
			Method:         record.Method,
			Note:           record.Note,
		})
	}
	return states
}

func sleepStates(entries []client.SleepEntry) []sleepState {
	states := make([]sleepState, 0, len(entries))
	for _, entry := range entries {
		states = append(states, sleepState{
			Date:         entry.RecordedAt.Format(time.DateOnly),
			QualityScore: entry.QualityScore,
			WakeupCount:  entry.WakeupCount,
			Note:         entry.Note,
		})
	}
	return states
}

func imagingStates(records []client.ImagingRecord) []imagingState {
	states := make([]imagingState, 0, len(records))
	for _, record := range records {
		states = append(states, imagingState{
			Date:       record.PerformedAt.Format(time.DateOnly),
			Modality:   record.Modality,
			BodySite:   record.BodySite,
			Title:      record.Title,
			Summary:    record.Summary,
			Impression: record.Impression,
			Note:       record.Note,
			Notes:      append([]string(nil), record.Notes...),
		})
	}
	return states
}
