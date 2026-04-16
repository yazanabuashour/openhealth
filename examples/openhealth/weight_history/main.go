package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	defaultPaths, err := client.ResolveLocalPaths(client.LocalConfig{})
	if err != nil {
		return fmt.Errorf("resolve default local paths: %w", err)
	}

	fs := flag.NewFlagSet("weight_history", flag.ContinueOnError)
	databasePath := fs.String("db", defaultPaths.DatabasePath, "SQLite database path")
	fromRaw := fs.String("from", "", "start timestamp, RFC3339")
	toRaw := fs.String("to", "", "end timestamp, RFC3339")
	limitRaw := fs.Int("limit", 25, "maximum number of weight entries to print")
	includeTrend := fs.Bool("trend", false, "print weight trend counts")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if *limitRaw < 1 {
		return fmt.Errorf("-limit must be at least 1")
	}

	params := client.ListHealthWeightParams{}
	if *fromRaw != "" {
		from, err := time.Parse(time.RFC3339, *fromRaw)
		if err != nil {
			return fmt.Errorf("parse -from: %w", err)
		}
		params.From = &from
	}
	if *toRaw != "" {
		to, err := time.Parse(time.RFC3339, *toRaw)
		if err != nil {
			return fmt.Errorf("parse -to: %w", err)
		}
		params.To = &to
	}
	limit := client.Limit(*limitRaw)
	params.Limit = &limit

	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: *databasePath})
	if err != nil {
		return fmt.Errorf("open local client: %w", err)
	}
	defer func() {
		if closeErr := api.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "close local client: %v\n", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	weights, err := api.ListHealthWeightWithResponse(ctx, &params)
	if err != nil {
		return fmt.Errorf("list weight history: %w", err)
	}
	if weights.JSON200 == nil {
		return fmt.Errorf("list weight history returned %s", weights.Status())
	}

	fmt.Printf("db=%s\n", api.Paths.DatabasePath)
	if len(weights.JSON200.Items) == 0 {
		fmt.Printf("no weight history found in %s\n", api.Paths.DatabasePath)
	} else {
		latest := weights.JSON200.Items[0]
		fmt.Printf("latest=%.1f %s recorded_at=%s\n", latest.Value, latest.Unit, latest.RecordedAt.Format(time.RFC3339))
		fmt.Printf("weight_history count=%d\n", len(weights.JSON200.Items))
		for _, item := range weights.JSON200.Items {
			fmt.Printf(
				"%s %.1f %s source=%s id=%d\n",
				item.RecordedAt.Format(time.RFC3339),
				item.Value,
				item.Unit,
				item.Source,
				item.Id,
			)
		}
	}

	if *includeTrend {
		weightRange := client.HealthWeightRangeAll
		trend, err := api.GetHealthWeightTrendWithResponse(ctx, &client.GetHealthWeightTrendParams{
			Range: &weightRange,
		})
		if err != nil {
			return fmt.Errorf("get weight trend: %w", err)
		}
		if trend.JSON200 == nil {
			return fmt.Errorf("get weight trend returned %s", trend.Status())
		}
		fmt.Printf(
			"trend range=%s raw_points=%d moving_average_points=%d monthly_average_buckets=%d\n",
			trend.JSON200.Range,
			len(trend.JSON200.RawPoints),
			len(trend.JSON200.MovingAveragePoints),
			len(trend.JSON200.MonthlyAverageBuckets),
		)
	}

	return nil
}
