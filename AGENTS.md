# Architecture

`cmd/avito-mcp` starts the MCP stdio server. `internal/avito` owns Avito HTTP and OAuth2 client credentials. `internal/server` exposes focused MCP tools.

Use Go 1.27.1. Keep stdout reserved for MCP messages. Never print credentials or access tokens. Add tests for changes to authentication, request construction, or tool behavior.
