package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return filepath.Abs(strings.TrimSpace(string(out)))
}

func commandOutputWithEnv(env []string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(out))
}

func countNewSessionFiles(marker time.Time, runRoot string) int {
	sessionsDir := filepath.Join(evalCodexHome(runRoot), "sessions")
	count := 0
	_ = filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(marker) && fileContains(path, runRoot) {
			count++
		}
		return nil
	})
	return count
}

func fileContains(path string, needle string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			return true
		}
	}
	return false
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func isWithin(path string, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC), nil
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func floatPointer(value float64) *float64 {
	return &value
}

func clientAnalyteSlug(value *string) *client.AnalyteSlug {
	if value == nil {
		return nil
	}
	slug := client.AnalyteSlug(*value)
	return &slug
}

func equalIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalExpectedStringPointer(got *string, want *string) bool {
	if want == nil {
		return true
	}
	return equalClinicalSentencePointer(got, want)
}

func equalExpectedStringSlice(got []string, want []string) bool {
	if len(want) == 0 {
		return true
	}
	return slices.Equal(got, want)
}

func equalClinicalSentence(left string, right string) bool {
	return normalizeClinicalSentence(left) == normalizeClinicalSentence(right)
}

func equalClinicalSentencePointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return equalClinicalSentence(*left, *right)
}

func normalizeClinicalSentence(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), ".")
}

func equalFloatPointer(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return math.Abs(*left-*right) < 0.001
}

func weightsEqual(got []weightState, want []weightState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Unit != want[i].Unit ||
			math.Abs(got[i].Value-want[i].Value) > 0.001 ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func bloodPressuresEqual(got []bloodPressureState, want []bloodPressureState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Systolic != want[i].Systolic ||
			got[i].Diastolic != want[i].Diastolic ||
			!equalIntPointer(got[i].Pulse, want[i].Pulse) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func medicationsEqual(got []medicationState, want []medicationState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Name != want[i].Name ||
			got[i].StartDate != want[i].StartDate ||
			!equalStringPointer(got[i].DosageText, want[i].DosageText) ||
			!equalStringPointer(got[i].EndDate, want[i].EndDate) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func labsEqual(got []labCollectionState, want []labCollectionState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) ||
			len(got[i].Results) != len(want[i].Results) {
			return false
		}
		for j := range got[i].Results {
			if got[i].Results[j].TestName != want[i].Results[j].TestName ||
				!equalStringPointer(got[i].Results[j].CanonicalSlug, want[i].Results[j].CanonicalSlug) ||
				!labValueTextEqual(got[i].Results[j].ValueText, want[i].Results[j].ValueText, want[i].Results[j].Units) ||
				!equalFloatPointer(got[i].Results[j].ValueNumeric, want[i].Results[j].ValueNumeric) ||
				!equalStringPointer(got[i].Results[j].Units, want[i].Results[j].Units) ||
				!equalExpectedStringSlice(got[i].Results[j].Notes, want[i].Results[j].Notes) {
				return false
			}
		}
	}
	return true
}

func labValueTextEqual(got string, want string, units *string) bool {
	if got == want {
		return true
	}
	if units == nil {
		return false
	}
	return got == strings.TrimSpace(want+" "+*units)
}

func bodyCompositionEqual(got []bodyCompositionState, want []bodyCompositionState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			!equalFloatPointer(got[i].BodyFatPercent, want[i].BodyFatPercent) ||
			!equalFloatPointer(got[i].WeightValue, want[i].WeightValue) ||
			!equalStringPointer(got[i].WeightUnit, want[i].WeightUnit) ||
			!equalStringPointer(got[i].Method, want[i].Method) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func sleepEqual(got []sleepState, want []sleepState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].QualityScore != want[i].QualityScore ||
			!equalIntPointer(got[i].WakeupCount, want[i].WakeupCount) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func imagingEqual(got []imagingState, want []imagingState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Modality != want[i].Modality ||
			!equalClinicalSentence(got[i].Summary, want[i].Summary) ||
			!equalStringPointer(got[i].BodySite, want[i].BodySite) ||
			!equalStringPointer(got[i].Title, want[i].Title) ||
			!equalClinicalSentencePointer(got[i].Impression, want[i].Impression) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) ||
			!equalExpectedStringSlice(got[i].Notes, want[i].Notes) {
			return false
		}
	}
	return true
}

func describeWeights(weights []weightState) string {
	if len(weights) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(weights))
	for _, weight := range weights {
		note := ""
		if weight.Note != nil {
			note = " note " + *weight.Note
		}
		parts = append(parts, fmt.Sprintf("%s %.1f %s%s", weight.Date, weight.Value, weight.Unit, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeBloodPressures(readings []bloodPressureState) string {
	if len(readings) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(readings))
	for _, reading := range readings {
		pulse := ""
		if reading.Pulse != nil {
			pulse = fmt.Sprintf(" pulse %d", *reading.Pulse)
		}
		note := ""
		if reading.Note != nil {
			note = " note " + *reading.Note
		}
		parts = append(parts, fmt.Sprintf("%s %d/%d%s%s", reading.Date, reading.Systolic, reading.Diastolic, pulse, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeMedications(medications []medicationState) string {
	if len(medications) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(medications))
	for _, medication := range medications {
		dosage := ""
		if medication.DosageText != nil {
			dosage = " " + *medication.DosageText
		}
		end := ""
		if medication.EndDate != nil {
			end = " to " + *medication.EndDate
		}
		note := ""
		if medication.Note != nil {
			note = " note " + *medication.Note
		}
		parts = append(parts, fmt.Sprintf("%s%s from %s%s%s", medication.Name, dosage, medication.StartDate, end, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeLabs(collections []labCollectionState) string {
	if len(collections) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(collections))
	for _, collection := range collections {
		results := make([]string, 0, len(collection.Results))
		for _, result := range collection.Results {
			note := ""
			if len(result.Notes) > 0 {
				note = " notes " + strings.Join(result.Notes, " / ")
			}
			results = append(results, fmt.Sprintf("%s %s%s", result.TestName, result.ValueText, note))
		}
		note := ""
		if collection.Note != nil {
			note = " note " + *collection.Note
		}
		parts = append(parts, fmt.Sprintf("%s%s (%s)", collection.Date, note, strings.Join(results, ", ")))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeBodyComposition(records []bodyCompositionState) string {
	if len(records) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(records))
	for _, record := range records {
		values := []string{}
		if record.BodyFatPercent != nil {
			values = append(values, fmt.Sprintf("body fat %.1f%%", *record.BodyFatPercent))
		}
		if record.WeightValue != nil && record.WeightUnit != nil {
			values = append(values, fmt.Sprintf("weight %.1f %s", *record.WeightValue, *record.WeightUnit))
		}
		if record.Method != nil {
			values = append(values, "method "+*record.Method)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", record.Date, strings.Join(values, ", ")))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeSleep(entries []sleepState) string {
	if len(entries) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		wakeups := ""
		if entry.WakeupCount != nil {
			wakeups = fmt.Sprintf(" wakeups %d", *entry.WakeupCount)
		}
		note := ""
		if entry.Note != nil {
			note = " note " + *entry.Note
		}
		parts = append(parts, fmt.Sprintf("%s quality %d%s%s", entry.Date, entry.QualityScore, wakeups, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeImaging(records []imagingState) string {
	if len(records) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(records))
	for _, record := range records {
		site := ""
		if record.BodySite != nil {
			site = " " + *record.BodySite
		}
		note := ""
		if len(record.Notes) > 0 {
			note = " notes " + strings.Join(record.Notes, " / ")
		}
		parts = append(parts, fmt.Sprintf("%s %s%s: %s%s", record.Date, record.Modality, site, record.Summary, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
