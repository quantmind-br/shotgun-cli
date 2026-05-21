# Data Dictionary — `internal/platform/http`

> **Module:** `internal/platform/http`
> **Package:** `github.com/quantmind-br/shotgun-cli/internal/platform/http`
> **Doc level:** detalhado

---

## A. Types

### A.1 `JSONClient`

| Property | Value |
|----------|-------|
| **Kind** | Struct (exported) |
| **Location** | `client.go:17` |
| **Fields** | see §A.1.1 |
| **Methods** | see §B.1–B.2 |
| **Receiver** | pointer (`*JSONClient`) |
| **Thread-safe** | Yes — `*net/http.Client` is safe for concurrent use |

#### A.1.1 Fields

| Field | Type | Visibility | Description |
|-------|------|------------|-------------|
| `httpClient` | `*nethttp.Client` | unexported | Underlying HTTP client with timeout configured. Never nil after `NewJSONClient`. |
| `baseURL` | `string` | unexported | Base URL concatenated with the request path. No trailing `/` expected. |

---

### A.2 `ClientConfig`

| Property | Value |
|----------|-------|
| **Kind** | Struct (exported) |
| **Location** | `client.go:22` |
| **Purpose** | Configuration parameter for `NewJSONClient` |
| **Fields** | see §A.2.1 |

#### A.2.1 Fields

| Field | Type | Exported | Description | Zero-value semantics |
|-------|------|----------|-------------|---------------------|
| `BaseURL` | `string` | ✅ | Base URL for the API (e.g. `"https://api.openai.com/v1"`) | `"0"` — no default; must be provided by caller |
| `Timeout` | `time.Duration` | ✅ | Maximum duration for a single HTTP request | `0` → defaults to `300 * time.Second` |

---

### A.3 `ProgressCallback`

| Property | Value |
|----------|-------|
| **Kind** | Function type (exported) |
| **Location** | `client.go:109` |
| **Signature** | `func(stage string, message string, current, total int64)` |
| **Purpose** | Callback invoked during `PostJSONWithProgress` to report operation progress |
| **Parameters** | |

| Param | Type | Description |
|-------|------|-------------|
| `stage` | `string` | Current operation stage, e.g. `"uploading"`, `"processing"` |
| `message` | `string` | Human-readable description of the progress step |
| `current` | `int64` | Current progress count |
| `total` | `int64` | Total expected count (or `-1` if unknown) |

---

### A.4 `HTTPError`

| Property | Value |
|----------|-------|
| **Kind** | Struct (exported) |
| **Location** | `client.go:114` |
| **Implements** | `error` interface (via `Error() string`) |
| **Purpose** | Carries HTTP error details (status code + response body) |
| **Fields** | see §A.4.1 |

#### A.4.1 Fields

| Field | Type | Exported | Description |
|-------|------|----------|-------------|
| `StatusCode` | `int` | ✅ | The HTTP status code returned by the server (e.g. 400, 500) |
| `Body` | `[]byte` | ✅ | The raw response body bytes |

#### A.4.2 Method

| Method | Receiver | Return type | Description |
|--------|----------|-------------|-------------|
| `Error() string` | `*HTTPError` | `string` | Returns `"HTTP <code>: <body>"` |

---

## B. Functions & Methods

### B.1 `NewJSONClient(cfg ClientConfig) *JSONClient`

| Property | Value |
|----------|-------|
| **Kind** | Constructor function (exported) |
| **Location** | `client.go:29` |
| **Parameters** | |

| Name | Type | Description |
|------|------|-------------|
| `cfg` | `ClientConfig` | Configuration for the client |

| Return | Type | Description |
|--------|------|-------------|
| — | `*JSONClient` | A new, configured JSONClient |

**Behavior:**
1. If `cfg.Timeout == 0`, replace with `300 * time.Second`.
2. Create `&nethttp.Client{Timeout: timeout}`.
3. Return `&JSONClient{httpClient: ..., baseURL: cfg.BaseURL}`.

---

### B.2 `JSONClient.PostJSON(ctx, path, headers, body, target) error`

| Property | Value |
|----------|-------|
| **Kind** | Method (exported) |
| **Receiver** | `*JSONClient` |
| **Location** | `client.go:45` |
| **Purpose** | Send a POST request with a JSON body and unmarshal the JSON response |

#### Parameters

| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Request context (cancellation, deadline) |
| `path` | `string` | API endpoint path appended to `c.baseURL` |
| `headers` | `map[string]string` | Additional HTTP headers (overridden by `Content-Type`) |
| `body` | `interface{}` | Request body — must be JSON-marshallable |
| `target` | `interface{}` | Pointer to struct for JSON unmarshalling |

#### Return

| Type | Description |
|------|-------------|
| `error` | `nil` on success; otherwise a wrapped error |

**Error variants:**
| Scenario | Error message prefix |
|----------|---------------------|
| JSON marshal fails | `"failed to marshal request"` |
| Request creation fails | `"failed to create request"` |
| HTTP transport fails | `"request failed"` |
| Response body read fails | `"failed to read response"` |
| Non-200 status code | `*HTTPError` (status + body) |
| Response JSON unmarshal fails | `"failed to parse response"` |

---

### B.3 `JSONClient.PostJSONWithProgress(ctx, path, headers, body, target, progressFn) ([]byte, error)`

| Property | Value |
|----------|-------|
| **Kind** | Method (exported) |
| **Receiver** | `*JSONClient` |
| **Location** | `client.go:88` |
| **Purpose** | Same as `PostJSON` but returns raw response bytes and invokes a progress callback |

#### Parameters

| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Request context |
| `path` | `string` | API endpoint path |
| `headers` | `map[string]string` | Additional HTTP headers |
| `body` | `interface{}` | Request body |
| `target` | `interface{}` | Response target pointer |
| `progressFn` | `ProgressCallback` | Optional progress callback (may be nil) |

#### Return

| # | Type | Description |
|---|------|-------------|
| 1 | `[]byte` | Raw response body (always returned, even on error) |
| 2 | `error` | Same semantics as `PostJSON`; `nil` on success |

**Behavior difference from `PostJSON`:**
- On success, if `progressFn != nil`, fires one event: `progressFn("uploading", "sending <N> bytes", N, N)` where N = `len(jsonBody)`.

---

## C. Test-only Types (not exported in production)

| Type | Location | Purpose |
|------|----------|---------|
| `testRequest` | `client_test.go:16` | Mock request body for test cases |
| `testResponse` | `client_test.go:21` | Mock response body for test cases |

---

## D. Constants (implicit)

| Constant | Value | Source |
|----------|-------|--------|
| Default timeout | `300 * time.Second` | `NewJSONClient` |
| HTTP status OK | `nethttp.StatusOK` (200) | `PostJSON`, `PostJSONWithProgress` |
| HTTP method | `nethttp.MethodPost` ("POST") | `PostJSON`, `PostJSONWithProgress` |
| Content-Type | `"application/json"` | `PostJSON`, `PostJSONWithProgress` |
