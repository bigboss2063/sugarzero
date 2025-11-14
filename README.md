# sugarzero

`sugarzero` is a context-first logging helper for Go applications. It wraps
[zerolog](https://github.com/rs/zerolog) so you get structured logs, caller
information, and consistent formatting, but exposes an API that revolves around
`context.Context`. Attach a logger and fields once, and every downstream call can
log with the exact same metadata. When combined with OpenTelemetry, trace and
span identifiers flow into your logs automatically.

## Features

- Context-aware initialization via `sugarzero.New`, plus a global fallback so
  functions can keep logging even when the context lacks a logger.
- Structured fields pulled from the context using `WithFields`, `WithField`, and
  `FieldsFromContext`, emitted on every log line without extra plumbing.
- OpenTelemetry integration through `WithTracing`, which injects `trace_id` and
  `span_id` whenever a recording span is found.
- Runtime log-level switches (`SetLogLevel`, `GetLogLevel`) that can be wired to
  admin APIs, CLIs, or feature flags.
- Drop-in helpers (`sugarzero.Debug`, `Infof`, `Warn`, etc.) backed by zerolog’s
  high-performance writer and caller annotations.
- Pluggable writers: pass `io.Writer` instances (or multiple writers) to
  `New` to mirror logs to files, sockets, or buffers for testing.

## Installation

```bash
go get github.com/bigboss2063/sugarzero
```

The module targets Go 1.24 (see `go.mod`) and relies only on standard library
packages plus zerolog and OpenTelemetry SDKs.

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/bigboss2063/sugarzero"
)

func main() {
	ctx, err := sugarzero.New(context.Background(), "info")
	if err != nil {
		log.Fatal(err)
	}

	ctx = sugarzero.WithFields(ctx,
		"service", "example-api",
		"env",     "staging",
	)

	sugarzero.Infof(ctx, "server listening on %s", ":8080")
	sugarzero.Warn(ctx, "user quota almost exceeded", "user_id", 42)
}
```

Even if downstream code forgets to pass the context, the global logger created
by `New` is reused and a warning is emitted only once.

### Adding contextual data

```go
func handleRequest(ctx context.Context, req *Request) {
	ctx = sugarzero.WithFields(ctx,
		"request_id", req.ID,
		"user_id",    req.UserID,
	)

	sugarzero.Info(ctx, "processing request")
}
```

To observe the exact fields attached to the current context (useful in tests or
middleware), call `sugarzero.FieldsFromContext(ctx)`.

### Wiring up tracing

`WithTracing` inspects the current OpenTelemetry span and stores its identifiers
so every log event includes `trace_id` and `span_id`:

```go
tracer := otel.Tracer("checkout")

ctx, span := tracer.Start(parentCtx, "chargePayment")
defer span.End()

ctx = sugarzero.WithTracing(ctx)
ctx = sugarzero.WithField(ctx, "order_id", order.ID)

sugarzero.Info(ctx, "starting payment flow")
```

See `examples/otel` for a more complete setup that configures a tracer provider,
propagation, and per-span logging.

### Changing log levels at runtime

Any component with access to a context carrying the logger (or the global
logger) can call:

```go
_ = sugarzero.SetLogLevel(ctx, "debug")
current := sugarzero.GetLogLevel(ctx) // "debug"
```

The `examples/loglevel` sample exposes `GET/POST /log-level` on `localhost:8080`
so you can demo level changes with:

```bash
curl http://localhost:8080/log-level
curl -X POST -H "Content-Type: application/json" \
  -d '{"level":"warn"}' \
  http://localhost:8080/log-level
```

### Custom writers

```go
file, _ := os.Create("app.log")
ctx, _ := sugarzero.New(context.Background(), "info", os.Stdout, file)
```

Passing multiple writers mirrors each structured log line to every target.

## Examples

- `examples/basic`: end-to-end walkthrough of initialization, fields, and
  dynamic level changes.
- `examples/demo`: smallest possible setup that emits logs at every level.
- `examples/otel`: shows how to pair sugarzero with OpenTelemetry traces so logs
  contain `trace_id`/`span_id`.
- `examples/loglevel`: HTTP endpoint that surfaces the current level and lets
  you switch between `trace`, `debug`, `info`, `warn`, `error`, `fatal`, and
  `panic`.

Run any example with `go run ./examples/<name>`.

## Development

- Format code with `gofmt` (CI/lint tooling is not bundled, but standard Go
  formatting keeps diffs minimal).
- Run tests with:

  ```bash
  go test ./...
  ```

- `sugarzero.Reset()` exists strictly for tests; do not invoke it in
  production code.

Issues and pull requests are welcome—open a discussion if you need additional
helpers or integrations.
