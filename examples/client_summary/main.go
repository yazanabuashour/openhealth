package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	api, err := client.OpenLocal(client.LocalConfig{
		DataDir: os.Getenv(client.EnvDataDir),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open local client: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "close local client: %v\n", closeErr)
			os.Exit(1)
		}
	}()

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
		"db=%s active medications=%d weights=%d\n",
		api.Paths.DatabasePath,
		summary.JSON200.ActiveMedicationCount,
		len(weights.JSON200.Items),
	)
}
