package health_test

import (
	"testing"

	"github.com/yazanabuashour/openhealth/internal/health"
)

func TestNormalizeAnalyteSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want health.AnalyteSlug
		ok   bool
	}{
		{
			name: "lowercase kebab case",
			in:   "vitamin-d",
			want: health.AnalyteSlug("vitamin-d"),
			ok:   true,
		},
		{
			name: "spaces and underscores",
			in:   "urine  albumin_creatinine   ratio",
			want: health.AnalyteSlug("urine-albumin-creatinine-ratio"),
			ok:   true,
		},
		{
			name: "mixed case",
			in:   "Hemoglobin A1c",
			want: health.AnalyteSlug("hemoglobin-a1c"),
			ok:   true,
		},
		{
			name: "trailing hyphen",
			in:   "vitamin-d-",
			ok:   false,
		},
		{
			name: "trailing underscore",
			in:   "hemoglobin_a1c_",
			ok:   false,
		},
		{
			name: "leading separator",
			in:   "-vitamin-d",
			ok:   false,
		},
		{
			name: "unsupported punctuation",
			in:   "bad/slug",
			ok:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := health.NormalizeAnalyteSlug(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("slug = %q, want %q", got, tt.want)
			}
		})
	}
}
