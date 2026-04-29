package main

import (
	"context"

	"github.com/yazanabuashour/openhealth/client"
)

func seedScenario(dbPath string, sc scenario) error {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	ctx := context.Background()
	switch sc.ID {
	case "mixed-bounded-range":
		if err := upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		}); err != nil {
			return err
		}
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "mt-mixed-latest-then-correct":
		if err := upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		}); err != nil {
			return err
		}
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "mt-bp-latest-then-correct":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "repeat-add":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "update-existing":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		})
	case "bounded-range", "bounded-range-natural", "latest-only":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "history-limit-two":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "bp-latest-only", "bp-bounded-range", "bp-bounded-range-natural":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "bp-history-limit-two":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "bp-correct-existing", "bp-correct-missing-reject":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		})
	case "bp-correct-ambiguous-reject":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
		})
	case "sleep-latest-only":
		return upsertSleep(ctx, api, []sleepState{
			{Date: "2026-03-28", QualityScore: 2, WakeupCount: intPointer(3)},
			{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(1)},
			{Date: "2026-03-30", QualityScore: 5},
		})
	case "medication-correct", "medication-delete":
		return recordMedications(ctx, api, []medicationState{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		})
	case "lab-range", "lab-latest-analyte", "lab-correct", "lab-delete":
		return recordLabs(ctx, api, []labCollectionState{
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
		})
	case "lab-patch":
		return recordLabs(ctx, api, []labCollectionState{
			{Date: "2026-03-29", Results: []labResultState{
				{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")},
				{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
			}},
		})
	case "imaging-correct", "imaging-delete":
		return recordImaging(ctx, api, []imagingState{
			{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough")},
		})
	default:
		return nil
	}
}

func upsertWeights(ctx context.Context, api *client.LocalClient, weights []weightState) error {
	for _, weight := range weights {
		recordedAt, err := parseDate(weight.Date)
		if err != nil {
			return err
		}
		if _, err := api.UpsertWeight(ctx, client.WeightRecordInput{
			RecordedAt: recordedAt,
			Value:      weight.Value,
			Unit:       client.WeightUnit(weight.Unit),
			Note:       weight.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordBloodPressures(ctx context.Context, api *client.LocalClient, readings []bloodPressureState) error {
	for _, reading := range readings {
		recordedAt, err := parseDate(reading.Date)
		if err != nil {
			return err
		}
		if _, err := api.RecordBloodPressure(ctx, client.BloodPressureRecordInput{
			RecordedAt: recordedAt,
			Systolic:   reading.Systolic,
			Diastolic:  reading.Diastolic,
			Pulse:      reading.Pulse,
			Note:       reading.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func upsertSleep(ctx context.Context, api *client.LocalClient, entries []sleepState) error {
	for _, entry := range entries {
		recordedAt, err := parseDate(entry.Date)
		if err != nil {
			return err
		}
		if _, err := api.UpsertSleep(ctx, client.SleepInput{
			RecordedAt:   recordedAt,
			QualityScore: entry.QualityScore,
			WakeupCount:  entry.WakeupCount,
			Note:         entry.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordMedications(ctx context.Context, api *client.LocalClient, medications []medicationState) error {
	for _, medication := range medications {
		if _, err := api.CreateMedicationCourse(ctx, client.MedicationCourseInput{
			Name:       medication.Name,
			DosageText: medication.DosageText,
			StartDate:  medication.StartDate,
			EndDate:    medication.EndDate,
			Note:       medication.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordLabs(ctx context.Context, api *client.LocalClient, collections []labCollectionState) error {
	for _, collection := range collections {
		collectedAt, err := parseDate(collection.Date)
		if err != nil {
			return err
		}
		results := make([]client.LabResultInput, 0, len(collection.Results))
		for _, result := range collection.Results {
			results = append(results, client.LabResultInput{
				TestName:      result.TestName,
				CanonicalSlug: clientAnalyteSlug(result.CanonicalSlug),
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				Notes:         append([]string(nil), result.Notes...),
			})
		}
		if _, err := api.CreateLabCollection(ctx, client.LabCollectionInput{
			CollectedAt: collectedAt,
			Note:        collection.Note,
			Panels:      []client.LabPanelInput{{PanelName: "Panel", Results: results}},
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordBodyComposition(ctx context.Context, api *client.LocalClient, records []bodyCompositionState) error {
	for _, record := range records {
		recordedAt, err := parseDate(record.Date)
		if err != nil {
			return err
		}
		var unit *client.WeightUnit
		if record.WeightUnit != nil {
			value := client.WeightUnit(*record.WeightUnit)
			unit = &value
		}
		if _, err := api.CreateBodyComposition(ctx, client.BodyCompositionInput{
			RecordedAt:     recordedAt,
			BodyFatPercent: record.BodyFatPercent,
			WeightValue:    record.WeightValue,
			WeightUnit:     unit,
			Method:         record.Method,
			Note:           record.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordImaging(ctx context.Context, api *client.LocalClient, records []imagingState) error {
	for _, record := range records {
		performedAt, err := parseDate(record.Date)
		if err != nil {
			return err
		}
		if _, err := api.CreateImaging(ctx, client.ImagingRecordInput{
			PerformedAt: performedAt,
			Modality:    record.Modality,
			BodySite:    record.BodySite,
			Title:       record.Title,
			Summary:     record.Summary,
			Impression:  record.Impression,
			Note:        record.Note,
			Notes:       append([]string(nil), record.Notes...),
		}); err != nil {
			return err
		}
	}
	return nil
}
