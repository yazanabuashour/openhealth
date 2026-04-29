package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func containsAll(value string, needles []string) bool {
	normalized := strings.ToLower(value)
	for _, needle := range needles {
		if !strings.Contains(normalized, strings.ToLower(needle)) {
			return false
		}
	}
	return true
}

func boundedRangeAssistantPass(message string) bool {
	previous := -1
	for _, date := range []string{"2026-03-30", "2026-03-29"} {
		index := includedDateResultLineIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedDate(message, "2026-03-28")
}

func latestOnlyAssistantPass(message string) bool {
	return mentionsIncludedDate(message, "2026-03-30") &&
		!mentionsIncludedDate(message, "2026-03-29") &&
		!mentionsIncludedDate(message, "2026-03-28")
}

func historyLimitTwoAssistantPass(message string) bool {
	previous := -1
	for _, date := range []string{"2026-03-30", "2026-03-29"} {
		index := includedDateResultLineIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedDate(message, "2026-03-28") &&
		!mentionsIncludedDate(message, "2026-03-27")
}

func bloodPressureBoundedRangeAssistantPass(message string) bool {
	previous := -1
	for _, expected := range []bloodPressureState{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	} {
		index := includedBloodPressureResultLineIndex(message, expected)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func bloodPressureLatestOnlyAssistantPass(message string) bool {
	return includedBloodPressureResultLineIndex(message, bloodPressureState{Date: "2026-03-30", Systolic: 118, Diastolic: 76}) >= 0 &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-29") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func bloodPressureHistoryLimitTwoAssistantPass(message string) bool {
	previous := -1
	for _, expected := range []bloodPressureState{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	} {
		index := includedBloodPressureResultLineIndex(message, expected)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedBloodPressureDate(message, "2026-03-28") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-27")
}

func mixedLatestAssistantPass(message string) bool {
	return dateMentionIndex(message, "2026-03-31") >= 0 &&
		containsAll(message, []string{"150.8", "119/77"}) &&
		containsAny(strings.ToLower(message), []string{"pulse 62", "62"})
}

func mixedLatestSeedAssistantPass(message string) bool {
	return mentionsIncludedDate(message, "2026-03-30") &&
		containsAll(message, []string{"151.6", "118/76"}) &&
		!mentionsIncludedDate(message, "2026-03-29")
}

func mixedBoundedRangeAssistantPass(message string) bool {
	return boundedRangeAssistantPass(message) &&
		bloodPressureBoundedRangeAssistantPass(message) &&
		containsAll(message, []string{"151.6", "152.2"}) &&
		!mentionsIncludedDate(message, "2026-03-28") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func nonISODateRejectAssistantPass(message string) bool {
	return containsAny(strings.ToLower(message), []string{"yyyy-mm-dd", "iso", "invalid", "cannot", "can't", "reject", "unsupported", "format"})
}

func sleepWakeupCountAssistantPass(message string, count int) bool {
	lower := strings.ToLower(message)
	digits := fmt.Sprintf("%d", count)
	words := numberWord(count)
	needles := []string{
		"woke up " + digits,
		digits + " wake",
		"wakeups " + digits,
		"wakeups: " + digits,
		"wakeup count " + digits,
		"wakeup_count\":" + digits,
		"wakeup_count\": " + digits,
		digits + " times",
		digits + " time",
	}
	if words != "" {
		needles = append(needles,
			"woke up "+words,
			words+" wake",
			words+" times",
			words+" time",
		)
	}
	if count == 2 {
		needles = append(needles, "twice")
	}
	return containsAny(lower, needles)
}

func numberWord(value int) string {
	switch value {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	default:
		return ""
	}
}

func mentionsIncludedDate(message string, date string) bool {
	return includedDateResultLineIndex(message, date) >= 0
}

func includedDateResultLineIndex(message string, date string) int {
	offset := 0
	for _, line := range strings.SplitAfter(message, "\n") {
		dateIndex := dateMentionIndex(line, date)
		if dateIndex < 0 {
			offset += len(line)
			continue
		}
		if lineMentionsExclusion(line) {
			offset += len(line)
			continue
		}
		if lineLooksLikeResult(line) {
			return offset + dateIndex
		}
		offset += len(line)
	}
	return -1
}

func includedBloodPressureResultLineIndex(message string, expected bloodPressureState) int {
	offset := 0
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		dateIndex := dateMentionIndex(line, expected.Date)
		if dateIndex < 0 {
			offset += len(line)
			continue
		}
		if lineMentionsExclusion(line) {
			offset += len(line)
			continue
		}
		if lineMentionsBloodPressure(line, expected.Systolic, expected.Diastolic) {
			return offset + dateIndex
		}
		if followingBloodPressureResultLine(lines, i, func(candidate string) bool {
			return lineMentionsBloodPressure(candidate, expected.Systolic, expected.Diastolic)
		}) {
			return offset + dateIndex
		}
		offset += len(line)
	}
	return -1
}

func mentionsIncludedBloodPressureDate(message string, date string) bool {
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		if dateMentionIndex(line, date) >= 0 && !lineMentionsExclusion(line) {
			if lineLooksLikeBloodPressureResult(line) {
				return true
			}
			if followingBloodPressureResultLine(lines, i, lineLooksLikeBloodPressureResult) {
				return true
			}
		}
	}
	return false
}

func followingBloodPressureResultLine(lines []string, start int, matches func(string) bool) bool {
	for i := start + 1; i < len(lines) && i <= start+3; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if lineMentionsExclusion(line) {
			return false
		}
		return matches(line)
	}
	return false
}

func lineMentionsBloodPressure(line string, systolic int, diastolic int) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, fmt.Sprintf("%d/%d", systolic, diastolic)) ||
		(strings.Contains(lower, fmt.Sprintf("%d", systolic)) && strings.Contains(lower, fmt.Sprintf("%d", diastolic)))
}

func lineLooksLikeBloodPressureResult(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return strings.Contains(lower, "/") ||
		strings.Contains(lower, "systolic") ||
		strings.Contains(lower, "diastolic") ||
		strings.Contains(lower, "blood pressure")
}

func lineMentionsExclusion(line string) bool {
	lower := strings.ToLower(line)
	return containsAny(lower, []string{"no entries", "not included", "not include", "excluded", "outside", "do not include"})
}

func lineLooksLikeResult(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
		return true
	}
	if len(trimmed) >= 2 && unicode.IsDigit(rune(trimmed[0])) && (trimmed[1] == '.' || trimmed[1] == ')') {
		return true
	}
	lower := strings.ToLower(trimmed)
	return strings.Contains(lower, " lb") ||
		strings.Contains(lower, "sleep quality") ||
		strings.Contains(lower, "quality_score") ||
		strings.Contains(lower, "wakeups") ||
		strings.Contains(lower, "wakeup_count") ||
		strings.Contains(lower, "glucose") ||
		strings.Contains(lower, "tsh") ||
		strings.Contains(lower, "mg/dl") ||
		strings.Contains(lower, "uiu/ml") ||
		(strings.Contains(lower, `"date"`) &&
			strings.Contains(lower, `"value"`) &&
			(strings.Contains(lower, `"unit":"lb"`) || strings.Contains(lower, `"unit": "lb"`)))
}

func mentionsDatesInOrder(message string, dates ...string) bool {
	previous := -1
	for _, date := range dates {
		index := includedDateResultLineIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return true
}

func dateMentionIndex(message string, date string) int {
	lower := strings.ToLower(message)
	best := -1
	for _, pattern := range dateMentionPatterns(date) {
		index := strings.Index(lower, pattern)
		if index >= 0 && (best < 0 || index < best) {
			best = index
		}
	}
	return best
}

func dateMentionPatterns(date string) []string {
	patterns := []string{strings.ToLower(date)}
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return patterns
	}
	return append(patterns,
		strings.ToLower(parsed.Format("01/02")),
		strings.ToLower(parsed.Format("January 2, 2006")),
		strings.ToLower(parsed.Format("January 2")),
		strings.ToLower(parsed.Format("Jan 2")),
	)
}

func promptSummary(sc scenario) string {
	switch sc.ID {
	case "add-two":
		return "add 2026-03-29 152.2 lb and 2026-03-30 151.6 lb"
	case "repeat-add":
		return "reapply the same two weights against preseeded rows"
	case "update-existing":
		return "correct 2026-03-29 from 152.2 lb to 151.6 lb"
	case "bounded-range":
		return "list only 2026-03-29 through 2026-03-30 from preseeded rows"
	case "bounded-range-natural":
		return "naturally ask for 2026-03-29 and 2026-03-30 from preseeded rows"
	case "latest-only":
		return "list only the latest row from preseeded rows"
	case "history-limit-two":
		return "list only the two most recent rows from preseeded rows"
	case "ambiguous-short-date":
		return "ask for year before writing 03/29 152.2 lb"
	case "invalid-input":
		return "reject -5 stone for 2026-03-31"
	case "non-iso-date-reject":
		return "reject non-ISO date 2026/03/31"
	case "body-composition-combined-weight-row":
		return "record combined weight and body-fat import row through weight and body-composition"
	case "bp-add-two":
		return "record 2026-03-29 122/78 pulse 64 and 2026-03-30 118/76"
	case "bp-latest-only":
		return "list only the latest blood-pressure row from preseeded rows"
	case "bp-history-limit-two":
		return "list only the two most recent blood-pressure rows from preseeded rows"
	case "bp-bounded-range":
		return "list only 2026-03-29 through 2026-03-30 blood-pressure rows from preseeded rows"
	case "bp-bounded-range-natural":
		return "naturally ask for 2026-03-29 and 2026-03-30 blood-pressure rows from preseeded rows"
	case "bp-invalid-input":
		return "reject invalid 0/-5 pulse 0 blood-pressure reading"
	case "bp-invalid-relation":
		return "reject systolic not greater than diastolic without tools"
	case "bp-non-iso-date-reject":
		return "reject non-ISO blood-pressure date 2026/03/31"
	case "bp-correct-existing":
		return "correct 2026-03-29 blood pressure from 122/78 pulse 64 to 121/77 pulse 63"
	case "bp-correct-missing-reject":
		return "reject correction for missing 2026-03-31 blood-pressure reading without creating one"
	case "bp-correct-ambiguous-reject":
		return "reject ambiguous correction when multiple 2026-03-29 blood-pressure rows exist"
	case "sleep-upsert-natural":
		return "record subjective sleep quality and optional wakeup count"
	case "sleep-latest-only":
		return "list only the latest sleep check-in from preseeded rows"
	case "sleep-invalid-input":
		return "reject invalid sleep quality and wakeup count"
	case "mixed-add-latest":
		return "record one weight and one blood-pressure reading, then report latest for both"
	case "mixed-bounded-range":
		return "list only 2026-03-29 through 2026-03-30 for both domains"
	case "mixed-invalid-direct-reject":
		return "reject invalid mixed weight and blood-pressure values without tools"
	case "medication-add-list":
		return "record two medications, then list active medications"
	case "medication-non-oral-dosage":
		return "record Semaglutide subcutaneous injection dosage text"
	case "medication-note":
		return "record medication course narrative note"
	case "medication-correct":
		return "correct Levothyroxine dosage from 25 mcg to 50 mcg"
	case "medication-delete":
		return "delete Vitamin D and list all medications"
	case "medication-invalid-date":
		return "reject medication date 2026/01/01 without tools"
	case "medication-end-before-start":
		return "reject medication end date before start date without tools"
	case "lab-record-list":
		return "record one glucose lab collection and list latest labs"
	case "lab-arbitrary-slug":
		return "record one Vitamin D lab collection with arbitrary slug"
	case "lab-note":
		return "record one glucose lab collection with a clinician note"
	case "lab-same-day-multiple":
		return "record two distinct same-day lab collections"
	case "lab-range":
		return "list only 2026-03-29 through 2026-03-30 lab collections"
	case "lab-latest-analyte":
		return "list only the latest glucose lab result"
	case "lab-correct":
		return "correct 2026-03-29 lab collection to TSH"
	case "lab-patch":
		return "patch one glucose lab result while preserving HDL"
	case "lab-delete":
		return "delete 2026-03-29 lab collection"
	case "lab-invalid-slug":
		return "reject invalid lab analyte slug without tools"
	case "mixed-medication-lab":
		return "record one medication and one lab result, then report both"
	case "imaging-record-list":
		return "record one chest X-ray imaging summary and list it"
	case "imaging-correct":
		return "correct one seeded imaging summary"
	case "imaging-delete":
		return "delete one seeded imaging record"
	case "mixed-import-file-coverage":
		return "record mixed import-file data that previously risked skipped rows"
	case "mt-weight-clarify-then-add":
		return "ask for missing year, then add 2026-03-29 152.2 lb in a resumed turn"
	case "mt-bp-latest-then-correct":
		return "read latest blood pressure, then correct it in a resumed turn"
	case "mt-mixed-latest-then-correct":
		return "read latest weight and blood pressure, then correct both in a resumed turn"
	default:
		return sc.Title
	}
}
