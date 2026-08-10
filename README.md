# mcpopenapi

[![Go Reference](https://pkg.go.dev/badge/github.com/yaronf/mcpopenapi.svg)](https://pkg.go.dev/github.com/yaronf/mcpopenapi)
[![CI](https://github.com/yaronf/mcpopenapi/actions/workflows/test.yml/badge.svg)](https://github.com/yaronf/mcpopenapi/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Turn an **OpenAPI 3** document into **MCP** tools over **Streamable HTTP**, or into portable **tool schemas** for other runtimes (e.g. OpenAI function calling).

Each operation becomes a tool (`operationId`). On MCP `tools/call`, the module rebuilds the HTTP request and forwards it to an in-process `http.Handler` — no second network hop, no embedded auth. `ParseToolSchemas` exposes the same OpenAPI→schema mapping without standing up MCP.

Born for [tripmap](https://github.com/yaronf/tripmap) (`/mcp`), but usable anywhere you already have an OpenAPI-described `http.Handler`.

## Why

OpenAPI is a natural description of an HTTP API; MCP is the native tool surface for Codex / Cursor / ChatGPT. If you already maintain OpenAPI for your API, this bridge avoids hand-writing a parallel tool schema.

## Install

```bash
go get github.com/yaronf/mcpopenapi@v0.2.0
```

Dependency: [`github.com/modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) v1.7+.

## Quick start

```go
package main

import (
	"log"
	"net/http"

	"github.com/yaronf/mcpopenapi"
)

func main() {
	api := http.NewServeMux()
	api.HandleFunc("GET /trips", listTrips)
	// …

	mcpHandler, err := mcpopenapi.NewHandler(mcpopenapi.Config{
		Name:         "my-api",
		Version:      "1.0.0",
		Instructions: "Prefer PATCH over full replace. List before edit.",
		OpenAPIYAML:  openAPIBytes, // valid OpenAPI 3 YAML or JSON
		Upstream:     http.StripPrefix("/api", api),
		PathPrefix:   "/api", // only expose these paths as tools
	})
	if err != nil {
		log.Fatal(err)
	}

	// Auth stays at the edge — the module does not implement Bearer/OAuth.
	http.Handle("/mcp", bearerAuth(token, mcpHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Clients (Codex, Cursor, ChatGPT) connect with Streamable HTTP to `https://host/mcp` and a Bearer token (or whatever your edge requires).

## Mapping rules

| OpenAPI | MCP tool |
|---------|----------|
| `operationId` | Tool name |
| `summary` + `description` | Tool description |
| Path / query / header parameters | Top-level input properties |
| `application/json` request body | Body properties flattened into the same object |
| `text/plain` request body | Single `body` string argument |
| `GET` / `HEAD` | `annotations.readOnlyHint: true` |
| Explicit `security: []` | Skipped (public ops stay off the tool list) |
| Missing `operationId` | Skipped |

`$ref` under `components/schemas` is resolved when building input schemas.

### Idempotency-Key

If an operation declares a header parameter named `Idempotency-Key`, it is **optional** in the tool schema. When the client omits it, the bridge generates a random key before calling Upstream.

### Path filter

Set `PathPrefix` (e.g. `"/api/agent"`) to expose only that subtree. Useful when the same OpenAPI document also documents `/health` or other public routes.

## Transport

`NewHandler` returns a **stateless** Streamable HTTP handler (`StreamableHTTPOptions.Stateless`). Each request is independent; no `Mcp-Session-Id` bookkeeping. That matches simple tool servers and works well behind load balancers.

Protocol support comes from the official Go MCP SDK (including recent protocol versions such as `2025-11-25` / `2026-07-28` as negotiated by the client).

## Auth (deliberately out of scope)

The handler does **not** check credentials. Mount it behind your own middleware (Bearer, mTLS, etc.). Upstream is typically the **internal** API mux — already past auth — so tool calls do not re-send the Bearer token in-process.

## Config reference

```go
type Config struct {
    Name         string       // MCP serverInfo.name
    Version      string       // MCP serverInfo.version
    Instructions string       // server-wide guidance (keep first ~512 chars useful)
    OpenAPIYAML  []byte       // OpenAPI 3 document
    Upstream     http.Handler // required; receives reconstructed requests
    PathPrefix   string       // optional path filter
}
```

`Instructions` is advertised on initialize. Codex/ChatGPT use it as cross-tool guidance — put workflow constraints there, not only in individual operation descriptions.

## Schema-only export

When you need the OpenAPI→tool mapping without MCP (for example OpenAI Responses / Chat Completions function tools):

```go
tools, err := mcpopenapi.ParseToolSchemas(openAPIBytes, "/api/agent")
// tools[i].Name, Description, InputSchema, ReadOnly
```

## Limitations (v1)

- Only `application/json` and `text/plain` request bodies
- No cookie / form / multipart mapping
- No OpenAPI link / callback / webhook support
- Body property names that collide with path/query params are skipped (path/query win)

## Tests

```bash
go test ./...
```

## License

[MIT](LICENSE)
