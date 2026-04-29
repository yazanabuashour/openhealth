package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yazanabuashour/openhealth/client"
)

func verifyScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	return verifyScenarioTurn(dbPath, sc, len(scenarioTurns(sc)), finalMessage)
}

func verifyScenarioTurn(dbPath string, sc scenario, turnIndex int, finalMessage string) (verificationResult, error) {
	if isMixedScenario(sc.ID) || strings.HasPrefix(sc.ID, "mt-") {
		return verifyMixedOrMultiTurnScenario(dbPath, sc, turnIndex, finalMessage)
	}
	if isBodyCompositionScenario(sc.ID) {
		return verifyBodyCompositionScenario(dbPath, sc, finalMessage)
	}
	if isBloodPressureScenario(sc.ID) {
		return verifyBloodPressureScenario(dbPath, sc, finalMessage)
	}
	if isSleepScenario(sc.ID) {
		return verifySleepScenario(dbPath, sc, finalMessage)
	}
	if isMedicationScenario(sc.ID) {
		return verifyMedicationScenario(dbPath, sc, finalMessage)
	}
	if isLabScenario(sc.ID) {
		return verifyLabScenario(dbPath, sc, finalMessage)
	}
	if isImagingScenario(sc.ID) {
		return verifyImagingScenario(dbPath, sc, finalMessage)
	}

	weights, err := listWeights(dbPath)
	var states []weightState
	listErrorDetail := ""
	if err != nil {
		rawStates, rawErr := listRawWeights(dbPath)
		if rawErr != nil {
			return verificationResult{}, fmt.Errorf("list weights: %w; raw fallback: %w", err, rawErr)
		}
		states = rawStates
		listErrorDetail = fmt.Sprintf(" typed list error: %v.", err)
	} else {
		states = weightStates(weights)
	}

	result := verificationResult{
		Weights: states,
	}
	switch sc.ID {
	case "add-two", "repeat-add":
		expected := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected exactly two newest-first rows; observed %s%s", describeWeights(states), listErrorDetail)
	case "update-existing":
		expected := []weightState{
			{Date: "2026-03-29", Value: 151.6, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected one updated row; observed %s%s", describeWeights(states), listErrorDetail)
	case "bounded-range", "bounded-range-natural":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = boundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed %s%s", describeWeights(states), listErrorDetail)
	case "latest-only":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = latestOnlyAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed %s%s", describeWeights(states), listErrorDetail)
	case "history-limit-two":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = historyLimitTwoAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed %s%s", describeWeights(states), listErrorDetail)
	case "ambiguous-short-date":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"year", "which year", "clarify", "ambiguous"})
		result.Details = fmt.Sprintf("expected no write and a year clarification; observed %s%s", describeWeights(states), listErrorDetail)
	case "invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "unsupported", "positive", "cannot", "can't", "unit", "value", "lb", "pounds"})
		result.Details = fmt.Sprintf("expected no write and an invalid input rejection; observed %s%s", describeWeights(states), listErrorDetail)
	case "non-iso-date-reject":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and a strict YYYY-MM-DD date rejection; observed %s%s", describeWeights(states), listErrorDetail)
	default:
		return verificationResult{}, fmt.Errorf("unknown scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyMixedOrMultiTurnScenario(dbPath string, sc scenario, turnIndex int, finalMessage string) (verificationResult, error) {
	weights, err := listWeights(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list weights: %w", err)
	}
	bloodPressures, err := listBloodPressures(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list blood pressures: %w", err)
	}
	medications, err := listMedications(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list medications: %w", err)
	}
	labs, err := listLabs(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list labs: %w", err)
	}
	bodyComposition, err := listBodyComposition(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list body composition: %w", err)
	}
	imaging, err := listImaging(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list imaging: %w", err)
	}
	weightStates := weightStates(weights)
	bloodPressureStates := bloodPressureStates(bloodPressures)
	medicationStates := medicationStates(medications)
	labStates := labCollectionStates(labs)
	bodyCompositionStates := bodyCompositionStates(bodyComposition)
	imagingStates := imagingStates(imaging)
	result := verificationResult{
		Weights:         weightStates,
		BodyComposition: bodyCompositionStates,
		BloodPressures:  bloodPressureStates,
		Medications:     medicationStates,
		Labs:            labStates,
		Imaging:         imagingStates,
	}

	switch sc.ID {
	case "mixed-add-latest":
		expectedWeights := []weightState{{Date: "2026-03-31", Value: 150.8, Unit: "lb"}}
		expectedReadings := []bloodPressureState{{Date: "2026-03-31", Systolic: 119, Diastolic: 77, Pulse: intPointer(62)}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = mixedLatestAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected latest weight and blood-pressure rows for 2026-03-31; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mixed-bounded-range":
		expectedWeights := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = mixedBoundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mixed-invalid-direct-reject":
		result.DatabasePass = len(weightStates) == 0 && len(bloodPressureStates) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "positive", "unsupported", "cannot", "can't", "reject", "stone", "systolic", "diastolic"})
		result.Details = fmt.Sprintf("expected no mixed-domain writes and a direct invalid input rejection; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mixed-medication-lab":
		expectedMedications := []medicationState{{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"}}
		expectedLabs := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}}}
		result.DatabasePass = medicationsEqual(medicationStates, expectedMedications) && labsEqual(labStates, expectedLabs) && len(weightStates) == 0 && len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "25 mcg", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected one medication and one glucose lab; observed medications %s and labs %s", describeMedications(medicationStates), describeLabs(labStates))
	case "mixed-import-file-coverage":
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb", Note: stringPointer("morning scale")}}
		expectedBody := []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb")}}
		expectedMedications := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}
		expectedLabs := []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}}
		expectedImaging := []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality", Impression: stringPointer("Normal chest radiograph"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) &&
			bodyCompositionEqual(bodyCompositionStates, expectedBody) &&
			medicationsEqual(medicationStates, expectedMedications) &&
			labsEqual(labStates, expectedLabs) &&
			imagingEqual(imagingStates, expectedImaging) &&
			len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"154.2", "18.7", "Glucose", "89", "Semaglutide", "X-ray", "narrative"})
		result.Details = fmt.Sprintf("expected no skipped import-file rows; observed weights %s, body composition %s, medications %s, labs %s, imaging %s", describeWeights(weightStates), describeBodyComposition(bodyCompositionStates), describeMedications(medicationStates), describeLabs(labStates), describeImaging(imagingStates))
	case "mt-weight-clarify-then-add":
		if turnIndex == 1 {
			result.DatabasePass = len(weightStates) == 0 && len(bloodPressureStates) == 0
			result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"year", "which year", "clarify", "ambiguous"})
			result.Details = fmt.Sprintf("expected no first-turn writes and a year clarification; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 152.2, Unit: "lb"}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "152.2"})
		result.Details = fmt.Sprintf("expected second-turn weight write after year clarification with no blood-pressure writes; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mt-mixed-latest-then-correct":
		if turnIndex == 1 {
			expectedWeights := []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			}
			expectedReadings := []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			}
			result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
			result.AssistantPass = mixedLatestSeedAssistantPass(finalMessage)
			result.Details = fmt.Sprintf("expected unchanged seed rows and latest mixed-domain answer; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedWeights := []weightState{
			{Date: "2026-03-30", Value: 151.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-30", "151", "117/75"})
		result.Details = fmt.Sprintf("expected latest mixed-domain corrections on 2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mt-bp-latest-then-correct":
		if turnIndex == 1 {
			expectedReadings := []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			}
			result.DatabasePass = len(weightStates) == 0 && bloodPressuresEqual(bloodPressureStates, expectedReadings)
			result.AssistantPass = bloodPressureLatestOnlyAssistantPass(finalMessage)
			result.Details = fmt.Sprintf("expected unchanged seed rows and latest blood-pressure answer; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = len(weightStates) == 0 && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = includedBloodPressureResultLineIndex(finalMessage, bloodPressureState{Date: "2026-03-30", Systolic: 117, Diastolic: 75}) >= 0
		result.Details = fmt.Sprintf("expected latest blood-pressure correction on 2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	default:
		return verificationResult{}, fmt.Errorf("unknown mixed or multi-turn scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyBodyCompositionScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	weights, err := listWeights(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list weights: %w", err)
	}
	bodyComposition, err := listBodyComposition(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list body composition: %w", err)
	}
	weightStates := weightStates(weights)
	bodyStates := bodyCompositionStates(bodyComposition)
	result := verificationResult{
		Weights:         weightStates,
		BodyComposition: bodyStates,
	}

	switch sc.ID {
	case "body-composition-combined-weight-row":
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb"}}
		expectedBody := []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb"), Method: stringPointer("smart scale")}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bodyCompositionEqual(bodyStates, expectedBody)
		result.AssistantPass = containsAll(finalMessage, []string{"154.2", "18.7"})
		result.Details = fmt.Sprintf("expected combined import row split into weight and body-composition domains; observed weights %s and body composition %s", describeWeights(weightStates), describeBodyComposition(bodyStates))
	default:
		return verificationResult{}, fmt.Errorf("unknown body-composition scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyBloodPressureScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	readings, err := listBloodPressures(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list blood pressures: %w", err)
	}
	states := bloodPressureStates(readings)
	result := verificationResult{
		BloodPressures: states,
	}

	switch sc.ID {
	case "bp-add-two":
		expected := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64), Note: stringPointer("home cuff")},
		}
		result.DatabasePass = bloodPressuresEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected exactly two newest-first blood-pressure rows; observed %s", describeBloodPressures(states))
	case "bp-latest-only":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureLatestOnlyAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed %s", describeBloodPressures(states))
	case "bp-history-limit-two":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureHistoryLimitTwoAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed %s", describeBloodPressures(states))
	case "bp-bounded-range", "bp-bounded-range-natural":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureBoundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed %s", describeBloodPressures(states))
	case "bp-correct-existing":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77, Pulse: intPointer(63)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "121/77"})
		result.Details = fmt.Sprintf("expected corrected 2026-03-29 blood-pressure row with no duplicate; observed %s", describeBloodPressures(states))
	case "bp-correct-missing-reject":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"no existing", "no local", "missing", "not found", "cannot", "can't", "did not", "not updated", "no update"})
		result.Details = fmt.Sprintf("expected unchanged seed row and missing-date correction rejection; observed %s", describeBloodPressures(states))
	case "bp-correct-ambiguous-reject":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"multiple", "ambiguous", "more than one", "cannot", "can't", "did not", "not updated"})
		result.Details = fmt.Sprintf("expected unchanged duplicate same-date rows and ambiguous correction rejection; observed %s", describeBloodPressures(states))
	case "bp-invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "positive", "cannot", "can't", "systolic", "diastolic", "pulse", "reject"})
		result.Details = fmt.Sprintf("expected no write and an invalid blood-pressure rejection; observed %s", describeBloodPressures(states))
	case "bp-invalid-relation":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "systolic", "diastolic", "greater", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and a systolic-greater-than-diastolic rejection; observed %s", describeBloodPressures(states))
	case "bp-non-iso-date-reject":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and a strict YYYY-MM-DD date rejection; observed %s", describeBloodPressures(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown blood-pressure scenario %q", sc.ID)
	}

	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifySleepScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	entries, err := listSleep(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list sleep: %w", err)
	}
	states := sleepStates(entries)
	result := verificationResult{Sleep: states}

	switch sc.ID {
	case "sleep-upsert-natural":
		expected := []sleepState{{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(2), Note: stringPointer("woke up after storm")}}
		result.DatabasePass = sleepEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "4"}) && sleepWakeupCountAssistantPass(finalMessage, 2)
		result.Details = fmt.Sprintf("expected one sleep check-in with quality 4 and two wakeups; observed %s", describeSleep(states))
	case "sleep-latest-only":
		expectedDB := []sleepState{
			{Date: "2026-03-30", QualityScore: 5},
			{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(1)},
			{Date: "2026-03-28", QualityScore: 2, WakeupCount: intPointer(3)},
		}
		result.DatabasePass = sleepEqual(states, expectedDB)
		result.AssistantPass = mentionsIncludedDate(finalMessage, "2026-03-30") &&
			!mentionsIncludedDate(finalMessage, "2026-03-29") &&
			!mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest sleep row 2026-03-30; observed %s", describeSleep(states))
	case "sleep-invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "quality", "1-5", "between 1 and 5", "wakeup", "negative", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and an invalid sleep rejection; observed %s", describeSleep(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown sleep scenario %q", sc.ID)
	}

	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyMedicationScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	medications, err := listMedications(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list medications: %w", err)
	}
	states := medicationStates(medications)
	result := verificationResult{Medications: states}

	switch sc.ID {
	case "medication-add-list":
		expected := []medicationState{
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
		}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "25 mcg"}) && !strings.Contains(strings.ToLower(finalMessage), "vitamin d")
		result.Details = fmt.Sprintf("expected two stored medications and active output limited to Levothyroxine; observed %s", describeMedications(states))
	case "medication-non-oral-dosage":
		expected := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("2.5 mg subcutaneous injection weekly"), StartDate: "2026-02-01"}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Semaglutide", "subcutaneous", "weekly"})
		result.Details = fmt.Sprintf("expected Semaglutide non-oral dosage text; observed %s", describeMedications(states))
	case "medication-note":
		expected := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Semaglutide", "subcutaneous", "coverage approved"})
		result.Details = fmt.Sprintf("expected Semaglutide medication note; observed %s", describeMedications(states))
	case "medication-correct":
		expected := []medicationState{
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
			{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
		}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "50 mcg"})
		result.Details = fmt.Sprintf("expected Levothyroxine dosage correction; observed %s", describeMedications(states))
	case "medication-delete":
		expected := []medicationState{{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no vitamin d"}) && containsAll(finalMessage, []string{"Levothyroxine"})
		result.Details = fmt.Sprintf("expected Vitamin D deleted and Levothyroxine retained; observed %s", describeMedications(states))
	case "medication-invalid-date":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and strict medication date rejection; observed %s", describeMedications(states))
	case "medication-end-before-start":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"end", "start", "before", "invalid", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and end-before-start rejection; observed %s", describeMedications(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown medication scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyLabScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	labs, err := listLabs(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list labs: %w", err)
	}
	states := labCollectionStates(labs)
	result := verificationResult{Labs: states}

	switch sc.ID {
	case "lab-record-list":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected one glucose lab collection; observed %s", describeLabs(states))
	case "lab-arbitrary-slug":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Vitamin D", CanonicalSlug: stringPointer("vitamin-d"), ValueText: "32", ValueNumeric: floatPointer(32), Units: stringPointer("ng/mL")}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Vitamin D", "32"})
		result.Details = fmt.Sprintf("expected one Vitamin D lab collection with arbitrary slug; observed %s", describeLabs(states))
	case "lab-note":
		expected := []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89", "stable", "A1C"})
		result.Details = fmt.Sprintf("expected lab collection note; observed %s", describeLabs(states))
	case "lab-same-day-multiple":
		expected := []labCollectionState{
			{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "TSH", "3.1", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected two distinct same-day lab collections; observed %s", describeLabs(states))
	case "lab-range":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = mentionsDatesInOrder(finalMessage, "2026-03-30", "2026-03-29") && !mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged lab seed rows and output limited to 2026-03-29..2026-03-30; observed %s", describeLabs(states))
	case "lab-latest-analyte":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89"}) && !mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged lab seed rows and latest glucose answer; observed %s", describeLabs(states))
	case "lab-correct":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "TSH", "3.1"})
		result.Details = fmt.Sprintf("expected 2026-03-29 lab correction; observed %s", describeLabs(states))
	case "lab-patch":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{
			{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92", ValueNumeric: floatPointer(92), Units: stringPointer("mg/dL")},
			{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
		}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "92", "HDL", "51"})
		result.Details = fmt.Sprintf("expected one patched glucose result with sibling HDL preserved; observed %s", describeLabs(states))
	case "lab-delete":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no lab"}) && !mentionsIncludedDate(finalMessage, "2026-03-29")
		result.Details = fmt.Sprintf("expected 2026-03-29 lab deleted; observed %s", describeLabs(states))
	case "lab-invalid-slug":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"analyte", "slug", "invalid", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and invalid analyte slug rejection; observed %s", describeLabs(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown lab scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyImagingScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	imaging, err := listImaging(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list imaging: %w", err)
	}
	states := imagingStates(imaging)
	result := verificationResult{Imaging: states}

	switch sc.ID {
	case "imaging-record-list":
		expected := []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality", Impression: stringPointer("Normal chest radiograph"), Note: stringPointer("ordered for cough"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}}
		result.DatabasePass = imagingEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "X-ray", "chest", "narrative"})
		result.Details = fmt.Sprintf("expected one chest X-ray imaging record; observed %s", describeImaging(states))
	case "imaging-correct":
		expected := []imagingState{{Date: "2026-03-29", Modality: "CT", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "Stable small pulmonary nodule.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough")}}
		result.DatabasePass = imagingEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "CT", "Stable"})
		result.Details = fmt.Sprintf("expected corrected CT imaging record; observed %s", describeImaging(states))
	case "imaging-delete":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no imaging"})
		result.Details = fmt.Sprintf("expected imaging record deleted; observed %s", describeImaging(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown imaging scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func listWeights(dbPath string) ([]client.WeightEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListWeights(context.Background(), client.WeightListOptions{Limit: 100})
}

func listBloodPressures(dbPath string) ([]client.BloodPressureEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListBloodPressure(context.Background(), client.BloodPressureListOptions{Limit: 100})
}

func listMedications(dbPath string) ([]client.MedicationCourse, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListMedicationCourses(context.Background(), client.MedicationListOptions{Status: client.MedicationStatusAll})
}

func listLabs(dbPath string) ([]client.LabCollection, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListLabCollections(context.Background())
}

func listBodyComposition(dbPath string) ([]client.BodyCompositionEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListBodyComposition(context.Background(), client.BodyCompositionListOptions{Limit: 100})
}

func listSleep(dbPath string) ([]client.SleepEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListSleep(context.Background(), client.SleepListOptions{Limit: 100})
}

func listImaging(dbPath string) ([]client.ImagingRecord, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListImaging(context.Background(), client.ImagingListOptions{Limit: 100})
}
