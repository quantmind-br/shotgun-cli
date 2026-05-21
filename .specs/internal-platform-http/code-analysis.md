# Code Analysis — `internal/platform/http`

> **Module:** `internal/platform/http`
> **Package:** `github.com/quantmind-br/shotgun-cli/internal/platform/http`
> **Go package:** `http`
> **Files analysed:** 2 source files (client.go, client_test.go)
> **Doc level:** detalhado

---

## 1. Overview

The `internal/platform/http` package provides a **shared, reusable HTTP client** for making JSON API requests to LLM provider backends (OpenAI, Anthropic, Gemini, …). It is consumed exclusively by `internal/platform/llmbase`, which in turn is used by every concrete provider (`openai`, `anthropic`, `geminiapi`).

The package is intentionally minimal: it exposes a single struct (`JSONClient`), a configuration struct (`ClientConfig`), a callback type (`ProgressCallback`), and a custom error type (`HTTPError`). All public methods live on `JSONClient`.

### Design goals

| Goal | How it's achieved |
|------|-------------------|
| **Reusability** | One `JSONClient` instance per provider, injected by `llmbase.BaseClient`. |
| **JSON-first** | All requests are marshaled to JSON; all responses are unmarshaled from JSON. |
| **Error transparency** | `HTTPError` captures status code + body so callers can inspect or reformat. |
| **Progress support** | Optional `ProgressCallback` parameter in `PostJSONWithProgress`. |
| **Timeout safety** | Default 300 s timeout; configurable via `ClientConfig`. |

---

## 2. Architecture & Dependencies

```
┌──────────────────────────────────────────────────────────┐
│                     callers (llmbase)                    │
│  ┌──────────────┐        ┌───────────────────────────┐   │
│  │ BaseClient   │ ──────►│  JSONClient               │   │
│  │              │        │  .PostJSON(...)            │   │
│  │              │        │  .PostJSONWithProgress(...)│   │
│  └──────────────┘        └───────────────────────────┘   │
│                           ▲                              │
│                           │  uses                        │
│                           ▼                              │
│              net/http  +  encoding/json  +  io  +  bytes  │
└──────────────────────────────────────────────────────────┘
```

### Import graph

| Import | Purpose |
|--------|---------|
| `bytes` | Build `bytes.Reader` for request body |
| `context` | Attach request context (cancellation, deadlines) |
| `encoding/json` | Marshal request body / unmarshal response body |
| `fmt` | Format error messages |
| `io` | `io.ReadAll` to read response body |
| `net/http` (aliased as `nethttp`) | HTTP client, request creation, constants |
| `time` | Timeout duration |

**Internal callers** (reverse deps):
- `internal/platform/llmbase/base_client.go` — calls `NewJSONClient`, `PostJSON`, and inspects `HTTPError`.

**No other internal packages depend directly on this package.**

---

## 3. File-by-file analysis

### 3.1 `client.go` (production code)

**Lines of code:** ~135 (including blank lines and comments).

#### 3.1.1 `JSONClient` struct

```go
type JSONClient struct {
    httpClient *nethttp.Client
    baseURL    string
}
```

- `httpClient`: A standard `net/http.Client` with a pre-set timeout. Never nil after creation.
- `baseURL`: Concatenated (not joined via `url.JoinPath`) with the `path` argument in `PostJSON` / `PostJSONWithProgress`.

> 🟡 **INFERIDO:** The package deliberately avoids `url.JoinPath` or `strings.TrimSuffix` on `baseURL`. Callers are expected to ensure `baseURL` does **not** end with `/` and `path` **does** start with `/`. This was verified by usage in `llmbase` which always passes absolute paths like `"/chat/completions"`.

#### 3.1.2 `ClientConfig` struct

```go
type ClientConfig struct {
    BaseURL string
    Timeout time.Duration
}
```

- `BaseURL`: Mandatory (no zero-value guard).
- `Timeout`: Zero means 300 s default.

#### 3.1.3 `NewJSONClient(cfg ClientConfig) *JSONClient`

Factory function.

1. Applies default timeout (300 s) if `cfg.Timeout == 0`.
2. Creates a new `nethttp.Client` with that timeout.
3. Returns a new `JSONClient`.

#### 3.1.4 `PostJSON(ctx, path, headers, body, target) error`

The primary workhorse method. Sequence:

```
1. json.Marshal(body)
   └─ error → return "failed to marshal request: %w"
2. url := c.baseURL + path
3. nethttp.NewRequestWithContext(ctx, POST, url, body)
   └─ error → return "failed to create request: %w"
4. Set Content-Type header
5. Set each header from headers map
6. c.httpClient.Do(req)
   └─ error → return "request failed: %w"
7. defer resp.Body.Close()
8. io.ReadAll(resp.Body)
   └─ error → return "failed to read response: %w"
9. if resp.StatusCode != 200
   └─ return &HTTPError{StatusCode, Body}
10. json.Unmarshal(respBody, target)
    └─ error → return "failed to parse response: %w"
11. return nil  ✓
```

#### 3.1.5 `PostJSONWithProgress(ctx, path, headers, body, target, progressFn) ([]byte, error)`

A variant that:
- Performs exactly the same HTTP flow as `PostJSON`.
- **Returns** `[]byte` (the raw response body) in addition to error (or `nil, []byte, error`).
- After a successful (200) response, if `progressFn != nil`, fires one progress event with `"uploading"` stage and message `"sending <N> bytes"` (where N is the request body size).

> 🟡 **INFERIDO:** The progress callback uses `len(jsonBody)` (the **request** payload size) rather than the response size. This is a minor semantic inaccuracy for progress reporting but sufficient for the current use case (fire-and-forget upload notification).

#### 3.1.6 `ProgressCallback` type

```go
type ProgressCallback func(stage string, message string, current, total int64)
```

A simple function type. The signature mirrors typical progress-bar libraries.

#### 3.1.7 `HTTPError` struct + `Error() string`

```go
type HTTPError struct {
    StatusCode int
    Body       []byte
}
func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, string(e.Body))
}
```

- Implements `error` interface.
- `Error()` renders `"HTTP <code>: <body>"`.
- Used by `llmbase.BaseClient.HandleHTTPError` for type-assertion and formatting.

---

### 3.2 `client_test.go` (tests)

**Lines of code:** ~160.

Uses `table-driven tests` via `httptest`. Covers:

| Test case | What it validates |
|-----------|-------------------|
| `defaults` | `NewJSONClient` with no timeout → 300 s default |
| `custom config` | Custom timeout is respected |
| `success` | Full happy path: JSON round-trip, headers propagated |
| `marshal error` | Non-serializable body → "failed to marshal request" |
| `server error` | HTTP 500 → `HTTPError` with body |
| `invalid response json` | 200 but malformed JSON → "failed to parse response" |
| `connection error` | Dead server → "request failed" |
| `invalid url` | Malformed URL → "failed to create request" |

Test helper types:
- `testRequest { Name string }` — request body shape
- `testResponse { ID, Echo string }` — response body shape

All tests run `t.Parallel()`.

---

## 4. Quality attributes

| Attribute | Rating | Notes |
|-----------|--------|-------|
| **Cyclomatic complexity** | Low | Each method has ≤ 4 branches; `PostJSON` has ~6. |
| **Cohesion** | High | All code serves the single purpose of JSON-over-HTTP API calls. |
| **Coupling** | Low | Depends only on stdlib; no internal circular deps. |
| **Test coverage** | Good | 8 test cases covering success, failure modes, defaults. |
| **Error handling** | Solid | Context propagation, defer close, typed `HTTPError`. |
| **Concurrency safety** | Safe | `*nethttp.Client` is safe for concurrent use. |

---

## 5. Known issues & improvement ideas

| # | Issue | Severity | Recommendation |
|---|-------|----------|----------------|
| 1 | `baseURL + path` string concatenation, no URL validation | Low | Consider `url.JoinPath(c.baseURL, path)` or `net/url.URL` for safety. |
| 2 | `PostJSONWithProgress` progress uses request size, not response size | Low | Either remove the callback from success path (no-op) or measure `len(respBody)`. |
| 3 | No retry logic | Medium | Transient network errors could be retried automatically (configurable). |
| 4 | No request middleware / interceptor pattern | Low | Could be useful for auth token refresh, logging, or metrics. |
| 5 | `client_test.go` does not test `PostJSONWithProgress` | Medium | Add at least one test exercising the progress callback path. |

---

## 6. Call sites (reverse dependencies)

| Caller | Location | Method used | Notes |
|--------|----------|-------------|-------|
| `internal/platform/llmbase` | `base_client.go` | `NewJSONClient`, `PostJSON` | Creates one `JSONClient` per provider; calls `PostJSON` inside `BaseClient.Send()`. |
| `internal/platform/llmbase` | `base_client.go` | `HTTPError` type assertion | `HandleHTTPError` checks `err.(*platformhttp.HTTPError)`. |
