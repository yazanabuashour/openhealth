# OpenHealth Agent Reference

## Agent Quick Start

- Use `client.OpenLocal(client.LocalConfig{})` for live local data. It opens the
  default SQLite database, applies migrations, and serves the generated OpenAPI
  handler in-process.
- Use `client.ResolveLocalPaths(client.LocalConfig{})` when you only need to
  report or verify the default database path.
- Use explicit `client.LocalConfig{DatabasePath: "..."}` or
  `client.LocalConfig{DataDir: "..."}` for tests, fixtures, and throwaway
  examples.
- Prefer generated client methods such as `ListHealthWeightWithResponse` over
  direct SQLite reads, so agent code stays pinned to the OpenAPI contract.
- Use `client.NewDefault(baseURL)` only when the user explicitly points you at
  an HTTP OpenHealth service. The in-process runtime base URL is a
  request-construction placeholder, not a network listener.

## Weight History

- Use `ListHealthWeightWithResponse` for "what is my weight history?" and
  "what is my latest weight?" questions.
- Weight entries are returned newest first. If the list is non-empty, the first
  entry is the latest weight.
- If the list is empty, report the resolved database path from
  `api.Paths.DatabasePath` so the user can tell which local dataset was checked.
- Use `From`, `To`, and `Limit` in `client.ListHealthWeightParams` for scoped
  history queries.
- Use `GetHealthWeightTrendWithResponse` only for trend or chart-style
  questions.

The runnable example at `examples/openhealth/weight_history/main.go` follows
this flow.
