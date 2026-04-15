package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	baseURL := os.Getenv("OPENHEALTH_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	api, err := client.NewDefault(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	summary, err := api.GetHealthSummaryWithResponse(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "summary request failed: %v\n", err)
		os.Exit(1)
	}
	if summary.JSON200 == nil {
		fmt.Fprintf(os.Stderr, "unexpected summary status: %s\n", summary.Status())
		os.Exit(1)
	}

	weights, err := api.ListHealthWeightWithResponse(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "weight request failed: %v\n", err)
		os.Exit(1)
	}
	if weights.JSON200 == nil {
		fmt.Fprintf(os.Stderr, "unexpected weight status: %s\n", weights.Status())
		os.Exit(1)
	}

	fmt.Printf(
		"active medications=%d weights=%d\n",
		summary.JSON200.ActiveMedicationCount,
		len(weights.JSON200.Items),
	)
}
