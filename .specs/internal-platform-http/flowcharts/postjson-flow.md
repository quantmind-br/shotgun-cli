# Flowchart: `JSONClient.PostJSON`

> **Method:** `(*JSONClient).PostJSON(ctx, path, headers, body, target) error`
> **Location:** `client.go:45`
> **Flow:** Main JSON-over-HTTP POST pipeline

---

## Mermaid diagram

```mermaid
flowchart TD
    A([Start: PostJSON]) --> B[json.Marshal body]
    B --> M1{marshal OK?}
    M1 -- NO --> E1[return error 'failed to marshal request']
    M1 -- YES --> C[Build URL: baseURL + path]
    C --> D[NewRequestWithContext POST]
    D --> M2{request OK?}
    M2 -- NO --> E2[return error 'failed to create request']
    M2 -- YES --> F[Set Content-Type: application/json]
    F --> G[Set headers from map]
    G --> H[httpClient.Do req]
    H --> M3{transport OK?}
    M3 -- NO --> E3[return error 'request failed']
    M3 -- YES --> I[Read resp.Body fully]
    I --> M4{read OK?}
    M4 -- NO --> E4[return error 'failed to read response']
    M4 -- YES --> J{status == 200?}
    J -- NO --> K[return &HTTPError{StatusCode, Body}]
    J -- YES --> L[json.Unmarshal respBody → target]
    L --> M5{unmarshal OK?}
    M5 -- NO --> E5[return error 'failed to parse response']
    M5 -- YES --> Z([return nil ✓])

    E1 --> Z
    E2 --> Z
    E3 --> Z
    E4 --> Z
    K --> Z
    E5 --> Z
```

---

## Step-by-step trace

| # | Step | Input | Output / Branch | Error? |
|---|------|-------|-----------------|--------|
| 1 | `json.Marshal(body)` | `body interface{}` | `[]byte, error` | If err ≠ nil → E1 |
| 2 | URL construction | `baseURL` + `path` | `url string` | — |
| 3 | `NewRequestWithContext` | `ctx`, `POST`, `url`, body reader | `*http.Request, error` | If err ≠ nil → E2 |
| 4 | Set headers | Content-Type + map | — | — |
| 5 | `httpClient.Do(req)` | request | `*http.Response, error` | If err ≠ nil → E3 |
| 6 | `io.ReadAll(resp.Body)` | response body | `[]byte, error` | If err ≠ nil → E4 |
| 7 | Status check | `resp.StatusCode` | 200 → continue; else → K | — |
| 8 | `json.Unmarshal(respBody, target)` | body + target ptr | `error` | If err ≠ nil → E5 |
| 9 | Return | — | `nil` | ✓ success |

---

## Error taxonomy

| Code path | Returned type | Example message |
|-----------|---------------|-----------------|
| E1 | `*fmt.wrapError` | `"failed to marshal request: <underlying>"` |
| E2 | `*fmt.wrapError` | `"failed to create request: <underlying>"` |
| E3 | `*fmt.wrapError` | `"request failed: <underlying>"` |
| E4 | `*fmt.wrapError` | `"failed to read response: <underlying>"` |
| K | `*HTTPError` | `&HTTPError{StatusCode: 500, Body: []byte("internal error")}` |
| E5 | `*fmt.wrapError` | `"failed to parse response: <underlying>"` |

---

## Resource management

- `resp.Body` is **defer-closed** immediately after `Do()` returns.
- The `defer` function uses `_ = resp.Body.Close()` — errors from close are silently discarded (intentional, per Go best practice).
