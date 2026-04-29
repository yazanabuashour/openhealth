package main

import (
	"fmt"
	"strings"
)

func variants() []variant {
	return []variant{
		{ID: "production", Title: "Production OpenHealth runner skill"},
	}
}

func scenarios() []scenario {
	return []scenario{
		{
			ID:     "add-two",
			Title:  "Add two weights and verify newest-first output",
			Prompt: "I need to update my local OpenHealth weight history. Use the configured local OpenHealth data path, and use year 2026 for these short dates: 03/29 152.2 lbs and 03/30 151.6. Then tell me the newest-first entries you see.",
		},
		{
			ID:     "repeat-add",
			Title:  "Repeat the same add request without creating duplicates",
			Prompt: "I need to record these OpenHealth weights again in the configured local data path, using year 2026 for the short dates: 03/29 152.2 lbs and 03/30 151.6. Tell me what is stored newest-first after you apply the request.",
		},
		{
			ID:     "update-existing",
			Title:  "Correct an existing same-date weight value",
			Prompt: "I need to correct my local OpenHealth weight for 03/29/2026. It should be 151.6 lbs, not 152.2 lbs. Update the configured local data and tell me what is stored now.",
		},
		{
			ID:     "bounded-range",
			Title:  "List a bounded date range newest-first",
			Prompt: "Please show every local OpenHealth weight history entry dated 03/29/2026 or 03/30/2026 only, newest first. Use the configured local data path. Include each entry dated 03/30/2026 and 03/29/2026 if present. In the final answer, mention only entries in that requested range; do not mention excluded dates at all.",
		},
		{
			ID:     "bounded-range-natural",
			Title:  "List a naturally phrased bounded date range",
			Prompt: "Can you show my OpenHealth weights for Mar 29 and Mar 30, 2026, newest first? Use the configured local data path.",
		},
		{
			ID:     "latest-only",
			Title:  "List only the latest weight entry",
			Prompt: "What is my latest local OpenHealth weight entry? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "history-limit-two",
			Title:  "List the two most recent weight entries",
			Prompt: "Please show my two most recent local OpenHealth weight entries, newest first. Use the configured local data path and mention only those two rows.",
		},
		{
			ID:     "ambiguous-short-date",
			Title:  "Clarify an ambiguous short date without writing",
			Prompt: "Please add this local OpenHealth weight: 03/29 152.2 lbs. There is no year context in this conversation or my request.",
		},
		{
			ID:     "invalid-input",
			Title:  "Reject invalid unit and value without writing",
			Prompt: "Please add this local OpenHealth weight entry: 03/31/2026 -5 stone.",
		},
		{
			ID:     "non-iso-date-reject",
			Title:  "Reject non-ISO full date without writing",
			Prompt: "Please add this local OpenHealth weight entry exactly as written: 2026/03/31 152.2 lbs. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "body-composition-combined-weight-row",
			Title:  "Record combined weight and body-fat import row through two domains",
			Prompt: "Use the configured local OpenHealth data path. Import this row: 03/29/2026 weight 154.2 lb and body-fat percentage 18.7% measured by smart scale. Record the scale weight as weight data and the body-fat value as body-composition data, then tell me what is stored.",
		},
		{
			ID:     "bp-add-two",
			Title:  "Record two blood-pressure readings and verify newest-first output",
			Prompt: "I need to update my local OpenHealth blood pressure history. Use the configured local OpenHealth data path, and use year 2026 for these short dates: 03/29 122/78 pulse 64 with note home cuff, and 03/30 118/76. Then tell me the newest-first entries you see.",
		},
		{
			ID:     "bp-latest-only",
			Title:  "List only the latest blood-pressure reading",
			Prompt: "What is my latest local OpenHealth blood pressure reading? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "bp-history-limit-two",
			Title:  "List the two most recent blood-pressure readings",
			Prompt: "Please show my two most recent local OpenHealth blood pressure readings, newest first. Use the configured local data path and mention only those two rows.",
		},
		{
			ID:     "bp-bounded-range",
			Title:  "List a bounded blood-pressure date range newest-first",
			Prompt: "Please show every local OpenHealth blood pressure reading dated 03/29/2026 or 03/30/2026 only, newest first. Use the configured local data path. Include each reading dated 03/30/2026 and 03/29/2026 if present. In the final answer, mention only readings in that requested range; do not mention excluded dates at all.",
		},
		{
			ID:     "bp-bounded-range-natural",
			Title:  "List a naturally phrased bounded blood-pressure date range",
			Prompt: "Can you show my OpenHealth blood pressure readings for Mar 29 and Mar 30, 2026, newest first? Use the configured local data path.",
		},
		{
			ID:     "bp-invalid-input",
			Title:  "Reject invalid blood-pressure values without writing",
			Prompt: "Please add this local OpenHealth blood pressure reading: 03/31/2026 0/-5 pulse 0.",
		},
		{
			ID:     "bp-invalid-relation",
			Title:  "Reject blood-pressure systolic not greater than diastolic",
			Prompt: "Please add this local OpenHealth blood pressure reading: 03/31/2026 78/78. Do not write it if systolic must be greater than diastolic.",
		},
		{
			ID:     "bp-non-iso-date-reject",
			Title:  "Reject non-ISO blood-pressure date without writing",
			Prompt: "Please add this local OpenHealth blood pressure reading exactly as written: 2026/03/31 122/78. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "bp-correct-existing",
			Title:  "Correct an existing same-date blood-pressure reading",
			Prompt: "I need to correct my local OpenHealth blood pressure reading for 03/29/2026. It should be 121/77 pulse 63, not 122/78 pulse 64. Update the configured local data and tell me what is stored now.",
		},
		{
			ID:     "bp-correct-missing-reject",
			Title:  "Reject a blood-pressure correction for a missing date",
			Prompt: "Please correct my local OpenHealth blood pressure reading for 03/31/2026 to 121/77. If there is no reading for that date, do not create a new one; tell me why it was not updated.",
		},
		{
			ID:     "bp-correct-ambiguous-reject",
			Title:  "Reject an ambiguous same-date blood-pressure correction",
			Prompt: "Please correct my local OpenHealth blood pressure reading for 03/29/2026 to 121/77. If more than one reading exists for that date, do not guess; tell me why it was not updated.",
		},
		{
			ID:     "sleep-upsert-natural",
			Title:  "Record a subjective sleep check-in with optional wakeups",
			Prompt: "Use the configured local OpenHealth data path. For my 03/29/2026 wake date, I slept good, woke up 2 times, and the note is woke up after storm. Record that sleep check-in and tell me what is stored.",
		},
		{
			ID:     "sleep-latest-only",
			Title:  "List only the latest sleep check-in",
			Prompt: "What is my latest local OpenHealth sleep check-in? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "sleep-invalid-input",
			Title:  "Reject invalid sleep quality and wakeup count without writing",
			Prompt: "Please add this local OpenHealth sleep check-in for 03/31/2026: quality 6 out of 5 and woke up -1 times.",
		},
		{
			ID:     "mixed-add-latest",
			Title:  "Record weight and blood-pressure readings, then report latest for both",
			Prompt: "Use the configured local OpenHealth data path. Record weight 150.8 lbs and blood pressure 119/77 pulse 62 for 03/31/2026. Then tell me the latest weight and latest blood-pressure entries.",
		},
		{
			ID:     "mixed-bounded-range",
			Title:  "List bounded weight and blood-pressure ranges newest-first",
			Prompt: "Use the configured local OpenHealth data path. Show my OpenHealth weights and blood pressure readings for Mar 29 and Mar 30, 2026 only, newest first in each domain. Do not mention entries outside that requested range.",
		},
		{
			ID:     "mixed-invalid-direct-reject",
			Title:  "Reject invalid mixed-domain values without writing",
			Prompt: "Please add these local OpenHealth entries: weight 03/31/2026 -5 stone and blood pressure 03/31/2026 0/-5 pulse 0.",
		},
		{
			ID:     "medication-add-list",
			Title:  "Record medications and list active courses",
			Prompt: "Use the configured local OpenHealth data path. Record these medications: Levothyroxine 25 mcg starting 01/01/2026 and Vitamin D starting 02/01/2026 ending 03/01/2026. Then list my active medications only.",
		},
		{
			ID:     "medication-non-oral-dosage",
			Title:  "Record a non-oral medication dosage text",
			Prompt: "Use the configured local OpenHealth data path. Record Semaglutide 2.5 mg subcutaneous injection weekly starting 02/01/2026. Then list my active medications.",
		},
		{
			ID:     "medication-note",
			Title:  "Record medication course narrative note",
			Prompt: "Use the configured local OpenHealth data path. Record Semaglutide 0.25 mg subcutaneous injection weekly starting 02/01/2026 with this medication note: coverage approved after prior authorization. Then list my active medications.",
		},
		{
			ID:     "medication-correct",
			Title:  "Correct an existing medication course",
			Prompt: "Use the configured local OpenHealth data path. Correct my Levothyroxine medication that started 01/01/2026 so the dosage is 50 mcg. Tell me what is stored now.",
		},
		{
			ID:     "medication-delete",
			Title:  "Delete an existing medication course",
			Prompt: "Use the configured local OpenHealth data path. Delete the Vitamin D medication course that started 02/01/2026. Then list all medications.",
		},
		{
			ID:     "medication-invalid-date",
			Title:  "Reject an invalid medication date without writing",
			Prompt: "Please add this local OpenHealth medication exactly as written: Levothyroxine 25 mcg starting 2026/01/01. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "medication-end-before-start",
			Title:  "Reject a medication end date before start date",
			Prompt: "Please add this local OpenHealth medication: Levothyroxine 25 mcg starting 01/02/2026 and ending 01/01/2026.",
		},
		{
			ID:     "lab-record-list",
			Title:  "Record labs and list latest collection",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with a Metabolic panel containing Glucose 89 mg/dL, canonical analyte glucose, range 70-99. Then show my latest lab collection.",
		},
		{
			ID:     "lab-arbitrary-slug",
			Title:  "Record and list an arbitrary lab analyte slug",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with a Micronutrients panel containing Vitamin D 32 ng/mL, canonical analyte vitamin-d. Then show my latest Vitamin D lab result.",
		},
		{
			ID:     "lab-note",
			Title:  "Record lab collection with clinician note",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with collection note \"labs look stable, keep moving\" and a Metabolic panel containing Glucose 89 mg/dL, canonical analyte glucose, with result notes \"HIV 4th gen narrative\" and \"A1C context\". Then show my latest lab collection.",
		},
		{
			ID:     "lab-same-day-multiple",
			Title:  "Record multiple distinct same-day lab collections",
			Prompt: "Use the configured local OpenHealth data path. Record two lab collections for 03/29/2026: one Metabolic panel with Glucose 89 mg/dL canonical analyte glucose, and one Thyroid panel with TSH 3.1 uIU/mL canonical analyte tsh. Then list lab collections newest first.",
		},
		{
			ID:     "lab-range",
			Title:  "List a bounded lab date range",
			Prompt: "Use the configured local OpenHealth data path. Show my OpenHealth lab collections for Mar 29 and Mar 30, 2026 only, newest first. Do not mention entries outside that requested range.",
		},
		{
			ID:     "lab-latest-analyte",
			Title:  "List latest lab result for a canonical analyte",
			Prompt: "Use the configured local OpenHealth data path. What is my latest glucose lab result? Mention only the latest matching collection/result.",
		},
		{
			ID:     "lab-correct",
			Title:  "Correct an existing lab collection",
			Prompt: "Use the configured local OpenHealth data path. Correct my lab collection dated 03/29/2026 so it has a Thyroid panel with TSH 3.1 uIU/mL, canonical analyte tsh. Tell me what is stored now.",
		},
		{
			ID:     "lab-patch",
			Title:  "Patch one lab result while preserving sibling results",
			Prompt: "Use the configured local OpenHealth data path. In my existing 03/29/2026 lab collection, correct only the Glucose result to 92 mg/dL and preserve the other lab results. Tell me what is stored now.",
		},
		{
			ID:     "lab-delete",
			Title:  "Delete an existing lab collection",
			Prompt: "Use the configured local OpenHealth data path. Delete my lab collection dated 03/29/2026. Then list lab collections.",
		},
		{
			ID:     "lab-invalid-slug",
			Title:  "Reject an invalid lab analyte slug shape without writing",
			Prompt: "Please add this local OpenHealth lab for 03/29/2026: UnknownTest 1 mg/dL with canonical analyte bad/slug. Do not write it if the analyte slug shape is invalid.",
		},
		{
			ID:     "mixed-medication-lab",
			Title:  "Record medication and lab data, then report latest domain entries",
			Prompt: "Use the configured local OpenHealth data path. Record Levothyroxine 25 mcg starting 01/01/2026 and a 03/29/2026 glucose lab result of 89 mg/dL. Then tell me the active medication and latest lab result.",
		},
		{
			ID:     "imaging-record-list",
			Title:  "Record an imaging summary and list it",
			Prompt: "Use the configured local OpenHealth data path. Record imaging from 03/29/2026: modality X-ray, body site chest, title Chest X-ray, summary No acute cardiopulmonary abnormality, impression Normal chest radiograph, note ordered for cough, result notes \"XR TOE RIGHT narrative\" and \"US Head/Neck findings\". Then list my latest imaging records.",
		},
		{
			ID:     "imaging-correct",
			Title:  "Correct an existing imaging summary",
			Prompt: "Use the configured local OpenHealth data path. Correct my 03/29/2026 imaging record so the modality is CT, body site chest, and the summary is Stable small pulmonary nodule. Tell me what is stored now.",
		},
		{
			ID:     "imaging-delete",
			Title:  "Delete an existing imaging record",
			Prompt: "Use the configured local OpenHealth data path. Delete my imaging record dated 03/29/2026. Then list imaging records.",
		},
		{
			ID:     "mixed-import-file-coverage",
			Title:  "Import file-style data that previously risked skipped rows",
			Prompt: "Use the configured local OpenHealth data path. Import this health-file data and do not skip supported fields: weight 154.2 lb with note morning scale and body-fat 18.7% on 03/29/2026; lab collection 03/29/2026 Glucose 89 mg/dL canonical analyte glucose with collection note \"labs look stable, keep moving\" and result notes \"HIV 4th gen narrative\" and \"A1C context\"; chest X-ray 03/29/2026 summary No acute cardiopulmonary abnormality impression Normal chest radiograph with result notes \"XR TOE RIGHT narrative\" and \"US Head/Neck findings\"; medication Semaglutide 0.25 mg subcutaneous injection weekly starting 02/01/2026 with note coverage approved after prior authorization. Then summarize what is stored.",
		},
		{
			ID:    "mt-weight-clarify-then-add",
			Title: "Clarify missing year, then add weight in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Please add this local OpenHealth weight: 03/29 152.2 lbs. There is no year context in this conversation or my request."},
				{Prompt: "Use 2026 as the year for that weight entry."},
			},
		},
		{
			ID:    "mt-mixed-latest-then-correct",
			Title: "Read latest mixed-domain entries, then correct both in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenHealth data path. What are my latest weight and blood-pressure entries? Mention only the latest row from each domain."},
				{Prompt: "Correct both latest entries for that same date: weight should be 151.0 lbs and blood pressure should be 117/75 pulse 63. Tell me what is stored now."},
			},
		},
		{
			ID:    "mt-bp-latest-then-correct",
			Title: "Read latest blood pressure, then correct it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenHealth data path. What is my latest blood-pressure reading? Mention only the latest row."},
				{Prompt: "Correct that latest reading to 117/75 pulse 63. Tell me what is stored now."},
			},
		},
	}
}

func selectVariants(filter string) ([]variant, error) {
	all := variants()
	if strings.TrimSpace(filter) == "" {
		return all, nil
	}
	selected := []variant{}
	for _, id := range splitFilterIDs(filter) {
		found := false
		for _, candidate := range all {
			if candidate.ID == id {
				selected = append(selected, candidate)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown variant %q", id)
		}
	}
	return selected, nil
}

func selectScenarios(filter string) ([]scenario, error) {
	all := scenarios()
	if strings.TrimSpace(filter) == "" {
		return all, nil
	}
	selected := []scenario{}
	for _, id := range splitFilterIDs(filter) {
		found := false
		for _, candidate := range all {
			if candidate.ID == id {
				selected = append(selected, candidate)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown scenario %q", id)
		}
	}
	return selected, nil
}

func splitFilterIDs(filter string) []string {
	ids := []string{}
	for _, raw := range strings.Split(filter, ",") {
		id := strings.TrimSpace(raw)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func scenarioTurns(sc scenario) []scenarioTurn {
	if len(sc.Turns) > 0 {
		return sc.Turns
	}
	return []scenarioTurn{{Prompt: sc.Prompt}}
}

func isMultiTurnScenario(sc scenario) bool {
	return len(scenarioTurns(sc)) > 1
}

func scenarioByID(id string) (scenario, bool) {
	for _, sc := range scenarios() {
		if sc.ID == id {
			return sc, true
		}
	}
	return scenario{}, false
}
