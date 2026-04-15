package app

import "testing"

func TestBanner(t *testing.T) {
	t.Parallel()

	if got, want := Banner(), banner; got != want {
		t.Fatalf("Banner() = %q, want %q", got, want)
	}
}
