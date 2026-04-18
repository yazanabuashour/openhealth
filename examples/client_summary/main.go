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

	summary, err := api.Summary(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "summary request failed: %v\n", err)
		os.Exit(1)
	}

	weights, err := api.ListWeights(ctx, client.WeightListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "weight request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"db=%s active medications=%d weights=%d\n",
		api.Paths.DatabasePath,
		summary.ActiveMedicationCount,
		len(weights),
	)
}
