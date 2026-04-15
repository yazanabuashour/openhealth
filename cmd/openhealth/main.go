package main

import (
	"fmt"
	"os"

	"github.com/yazanabuashour/openhealth/internal/app"
)

func main() {
	if _, err := fmt.Fprintln(os.Stdout, app.Banner()); err != nil {
		os.Exit(1)
	}
}
