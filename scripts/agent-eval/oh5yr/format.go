package main

import (
	"fmt"
	"math"
	"os"
)

func roundWeight(value float64) float64 {
	return math.Round(value*10) / 10
}

func roundSeconds(value float64) float64 {
	return math.Round(value*100) / 100
}

func passText(value bool) string {
	if value {
		return "pass"
	}
	return "fail"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func formatIntDelta(value *int) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+d", *value)
}

func formatFloatDelta(value *float64) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+.2f", *value)
}

func deref(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
