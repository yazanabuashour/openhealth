# OpenHealth Agent Reference

The production agent path is the ergonomic Go SDK facade in
`github.com/yazanabuashour/openhealth/client`.

## Agent Quick Start

- Use `client.OpenLocal(client.LocalConfig{})` for live local data. It opens the
  default SQLite database, applies migrations, and serves the OpenAPI handler
  in-process.
- Use `client.ResolveLocalPaths(client.LocalConfig{})` only when you need to
  report or verify the default database path.
- Use explicit `client.LocalConfig{DatabasePath: "..."}` or
  `client.LocalConfig{DataDir: "..."}` for tests, fixtures, and throwaway
  examples.
- Use `UpsertWeight`, `RecordWeight`, `ListWeights`, and `LatestWeight` for
  routine weight writes and reads.
- Use generated OpenAPI methods only for endpoints not covered by the SDK
  facade or when the user explicitly needs raw API-contract behavior.

## Weight History

- Use `LatestWeight` for "what is my latest weight?"
- Use `ListWeights` for "what is my weight history?"
- Weight entries are returned newest first.
- If the list is empty, report the resolved database path from
  `api.Paths.DatabasePath` so the user can tell which local dataset was checked.
- Use `WeightListOptions{From: ..., To: ..., Limit: ...}` for scoped history
  queries.

Copyable weight task snippets live at [weights.md](weights.md).
