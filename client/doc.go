// Package client exposes the local OpenHealth SDK for direct local-first use.
//
// Prefer OpenLocal(LocalConfig{}) when running on a developer machine or inside
// an agent-managed Go workspace. It opens SQLite locally, applies migrations,
// and exposes typed helper methods over the local health service.
package client
