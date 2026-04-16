// Package client exposes the generated OpenHealth API client plus a local
// in-process runtime for default local-first use.
//
// Prefer OpenLocal(LocalConfig{}) when running on a developer machine or inside
// an agent-managed Go workspace. It opens SQLite locally, applies migrations,
// and routes requests to the generated OpenAPI handler without binding a port.
// For common local weight tasks, use the LocalClient helper methods
// RecordWeight, UpsertWeight, ListWeights, and LatestWeight before falling back
// to generated OpenAPI method names.
//
// Use NewDefault(baseURL) only when you intentionally want to talk to an
// explicit HTTP server.
package client
