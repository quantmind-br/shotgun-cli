```mermaid
sequenceDiagram
  participant Client
  participant API
  participant Store
  Client->>API: Request data
  API->>Store: Load entity
  Store-->>API: Entity
  API-->>Client: Response
```
