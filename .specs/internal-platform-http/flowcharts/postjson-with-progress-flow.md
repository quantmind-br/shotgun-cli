# Flowchart: `JSONClient.PostJSONWithProgress`

> **Method:** `(*JSONClient).PostJSONWithProgress(ctx, path, headers, body, target, progressFn) ([]byte, error)`
> **Location:** `client.go:88`
> **Flow:** JSON-over-HTTP POST with progress callback on success

---

## Mermaid diagram

```mermaid
flowchart TD
    A([Start: PostJSONWithProgress]) --> B[json.Marshal body]
    B --> M1{marshal OK?}
    M1 -- NO --> E1[return nil, error 'failed to marshal request']
    M1 -- YES --> C[Build URL: baseURL + path]
    C --> D[NewRequestWithContext POST]
    D --> M2{request OK?}
    M2 -- NO --> E2[return nil, error 'failed to create request']
    M2 -- YES --> F[Set Content-Type + headers]
    F --> G[httpClient.Do req]
    G --> M3{transport OK?}
    M3 -- NO --> E3[return nil, error 'request failed']
    M3 -- YES --> H[Read resp.Body]
    H --> M4{read OK?}
    M4 -- NO --> E4[return nil, error 'failed to read response']
    M4 -- YES --> I{status == 200?}
    I -- NO --> K[return respBody, &HTTPError]
    I -- YES --> L[json.Unmarshal → target]
    L --> M5{unmarshal OK?}
    M5 -- NO --> E5[return respBody, error 'failed to parse response']
    M5 -- YES --> P{progressFn != nil?}
    P -- NO --> Z([return respBody, nil ✓])
    P -- YES --> Q[contentSize = len(jsonBody)]
    Q --> R[progressFn 'uploading' 'sending N bytes' N N]
    R --> Z

    E1 --> Z
    E2 --> Z
    E3 --> Z
    E4 --> Z
    K --> Z
    E5 --> Z
```

---

## Step-by-step trace

| # | Step | Input | Output / Branch | Returns |
|---|------|-------|-----------------|---------|
| 1 | `json.Marshal(body)` | `body` | `[]byte, error` | `nil, err` on fail |
| 2 | URL: `baseURL + path` | — | `string` | — |
| 3 | `NewRequestWithContext` | `ctx`, `POST`, `url`, reader | `*req, error` | `nil, err` on fail |
| 4 | Set headers | Content-Type + map | — | — |
| 5 | `httpClient.Do(req)` | request | `*resp, error` | `nil, err` on fail |
| 6 | `io.ReadAll(resp.Body)` | body | `[]byte, error` | `nil, err` on fail |
| 7 | Status ≠ 200? | `resp.StatusCode` | yes → K | `respBody, *HTTPError` |
| 8 | `json.Unmarshal` | body → target | `error` | `respBody, err` on fail |
| 9 | `progressFn != nil`? | callback ref | yes → Q | — |
| 10 | Fire progress event | `len(jsonBody)` | one callback call | — |
| 11 | Return | — | `respBody, nil` | ✓ |

---

## Differences from `PostJSON`

| Aspect | `PostJSON` | `PostJSONWithProgress` |
|--------|-----------|----------------------|
| Return signature | `error` | `([]byte, error)` |
| Return on success | `nil` | `respBody, nil` |
| Return on non-200 | `*HTTPError` | `respBody, *HTTPError` |
| Progress callback | N/A | One event on success path |
| Progress stage | — | `"uploading"` |
| Progress message | — | `"sending <N> bytes"` |
| Progress current/total | — | `len(jsonBody), len(jsonBody)` |

> 🟡 **INFERIDO:** The progress callback reports the **request** payload size (`len(jsonBody)`) rather than the response body size. This is a known semantic simplification — the callback fires once and reports the upload size, not the download size.

---

## Resource management

- Same as `PostJSON`: `resp.Body` is **defer-closed** immediately after `Do()` returns.
