package main

import (
	"strings"
	"testing"
)

func TestNonISODateRejectAssistantPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "strict format rejection",
			message: "Invalid date: use YYYY-MM-DD.",
			want:    true,
		},
		{
			name:    "reject wording",
			message: "I can't record 2026/03/31 because that date format is unsupported.",
			want:    true,
		},
		{
			name:    "bare date mention is not a rejection",
			message: "The date is 2026/03/31.",
			want:    false,
		},
		{
			name:    "successful write wording is not a rejection",
			message: "Recorded 2026/03/31 152.2 lb.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := nonISODateRejectAssistantPass(tt.message); got != tt.want {
				t.Fatalf("nonISODateRejectAssistantPass(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestSleepWakeupCountAssistantPassRequiresCountContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "date digit does not satisfy count",
			message: "Stored 2026-03-29 with quality 4.",
			want:    false,
		},
		{
			name:    "woke up digit",
			message: "Stored 2026-03-29 with quality 4 and woke up 2 times.",
			want:    true,
		},
		{
			name:    "word wakeups",
			message: "Stored 2026-03-29 with quality 4 and two wakeups.",
			want:    true,
		},
		{
			name:    "json wakeup count",
			message: `{"date":"2026-03-29","quality_score":4,"wakeup_count":2}`,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := sleepWakeupCountAssistantPass(tt.message, 2); got != tt.want {
				t.Fatalf("sleepWakeupCountAssistantPass(%q, 2) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestMentionsDatesInOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "newest first iso",
			message: "Stored rows: 2026-03-30 151.6 lb, 2026-03-29 152.2 lb.",
			want:    true,
		},
		{
			name:    "oldest first iso",
			message: "Stored rows: 2026-03-29 152.2 lb, 2026-03-30 151.6 lb.",
			want:    false,
		},
		{
			name:    "newest first short dates",
			message: "03/30: 151.6 lb; 03/29: 152.2 lb.",
			want:    true,
		},
		{
			name: "newest first nested lab answer with range heading",
			message: `OpenHealth lab collections for March 29 and March 30, 2026, newest first:
- 2026-03-30
  - TSH: 3.4 uIU/mL
- 2026-03-29
  - Glucose: 89 mg/dL`,
			want: true,
		},
		{
			name:    "missing date",
			message: "Only 2026-03-30 is present.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mentionsDatesInOrder(tt.message, "2026-03-30", "2026-03-29")
			if got != tt.want {
				t.Fatalf("mentionsDatesInOrder(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestBoundedRangeAssistantPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name: "allows excluded date in exclusion sentence",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"No entries from 03/28/2026 are included.",
			}, "\n"),
			want: true,
		},
		{
			name: "ignores natural date order in prose before result rows",
			message: strings.Join([]string{
				"Here are the weights for March 29 and March 30, 2026:",
				"",
				"- 2026-03-30 12:00:00Z: 151.6 lb",
				"- 2026-03-29 12:00:00Z: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name: "allows bounded prose with bullet result rows",
			message: strings.Join([]string{
				"Using the configured local database, the weights for March 29 and March 30, 2026, newest first:",
				"",
				"- 2026-03-30: 151.6 lb",
				"- 2026-03-29: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name:    "allows compact same-line result rows",
			message: "2026-03-30: 151.6 lb; 2026-03-29: 152.2 lb",
			want:    true,
		},
		{
			name: "rejects excluded date as result row",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"3. 2026-03-28 153.0 lb",
			}, "\n"),
			want: false,
		},
		{
			name:    "rejects missing newest date",
			message: "1. 2026-03-29 152.2 lb",
			want:    false,
		},
		{
			name: "rejects oldest first",
			message: strings.Join([]string{
				"1. 2026-03-29 152.2 lb",
				"2. 2026-03-30 151.6 lb",
			}, "\n"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := boundedRangeAssistantPass(tt.message); got != tt.want {
				t.Fatalf("boundedRangeAssistantPass(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestMixedBoundedRangeAssistantPassAllowsJSONAnswer(t *testing.T) {
	t.Parallel()

	message := `{"weights":[{"date":"2026-03-30","value":151.6,"unit":"lb"},{"date":"2026-03-29","value":152.2,"unit":"lb"}],"blood_pressure":[{"date":"2026-03-30","systolic":118,"diastolic":76},{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}`
	if !mixedBoundedRangeAssistantPass(message) {
		t.Fatalf("mixedBoundedRangeAssistantPass rejected JSON answer: %s", message)
	}
}
