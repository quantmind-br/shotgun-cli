# Fluxograma: Registro e Criação de Providers (`Registry`)

**Arquivo fonte:** `registry.go`
**Módulo:** `github.com/quantmind-br/shotgun-cli/internal/core/llm`
**Tipo:** `*llm.Registry` + `ProviderCreator` (factory pattern)

---

## Fluxograma Mermaid

```mermaid
flowchart TD
    Start([Início: Programa inicia]) --> NewReg["llm.NewRegistry()"]
    
    NewReg --> RegCreated["Registry criado:\n  mu: sync.RWMutex{locked: false}\n  creators: map[ProviderType]ProviderCreator{} (vazio)"]

    RegCreated --> Phase1[("FASE 1: Registro\n(init / bootstrap)")]

    Phase1 --> RegOpenAI["r.Register(ProviderOpenAI, factoryOpenAI)\n→ lock.write()\n→ creators[openai] = factoryOpenAI\n→ unlock()"]

    RegOpenAI --> RegAnthropic["r.Register(ProviderAnthropic, factoryAnthropic)\n→ lock.write()\n→ creators[anthropic] = factoryAnthropic\n→ unlock()"]

    RegAnthropic --> RegGemini["r.Register(ProviderGemini, factoryGemini)\n→ lock.write()\n→ creators[gemini] = factoryGemini\n→ unlock()"]

    RegGemini --> RegistryComplete["Registry completo:\n  creators: {\n    openai: → openai.NewClient,\n    anthropic: → anthropic.NewClient,\n    gemini: → geminiapi.NewClient\n  }"]

    RegistryComplete --> Phase2[("FASE 2: Criação\n(runtime)")]

    Phase2 --> UserReq["Usuário requisita LLM\n→ Config{Provider: openai, APIKey: \"...\", ...}"]

    UserReq --> ApplyDefaults["cfg.WithDefaults()\n→ preenche BaseURL, Model, Timeout"]

    ApplyDefaults --> CreateCall["r.Create(cfg)"]

    CreateCall --> RLock["lock.read()"]
    RLock --> Lookup{"creators[cfg.Provider]\nexiste?"}

    Lookup -->|Sim| Found["creator = creators[cfg.Provider]\nr.Unlock()"]
    Lookup -->|Não| NotFound["creator = nil, ok = false\nr.Unlock()"]

    NotFound --> CheckOk{"ok?"}
    CheckOk -->|Não| ErrUnsupported["Erro: 'unsupported provider: <nome>'\nFim: Error"]

    Found --> FactoryCall["creator(cfg)\n→ ex: openai.NewClient(cfg)"]

    FactoryCall --> FactoryErr{"Factory\nfalhou?"}
    FactoryErr -->|Sim| ErrFactory["Erro: <err do factory>\nFim: Error"]
    FactoryErr -->|Não| ProviderCreated([Provider pronto: Client{config}])

    ErrUnsupported --> EndErr([Fim: Error])

    classDef initPhase fill:#E6E6FA,stroke:#333,stroke-width:2px
    classDef runtime fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef success fill:#90EE90,stroke:#333,stroke-width:2px
    classDef error fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef decision fill:#FFD700,stroke:#333,stroke-width:2px
    classDef phase fill:#FFE4B5,stroke:#FF8C00,stroke-width:3px

    class Phase1,Phase2 phase
    class NewReg,RegCreated,RegOpenAI,RegAnthropic,RegGemini,RegistryComplete initPhase
    class UserReq,ApplyDefaults,CreateCall,RLookup,Lookup,Found,FactoryCall runtime
    class ProviderCreated success
    class ErrUnsupported,ErrFactory,ErrOpenAI error
    class CheckOk,FactoryErr decision
```

---

## Detalhamento: Fase 1 — Registro (init / bootstrap)

### Passos de Registro

Cada chamada `Register` segue o mesmo padrão:

```go
r.mu.Lock()           // write lock — exclusive
defer r.mu.Unlock()
r.creators[providerType] = creator
```

| Linha | Ação | Lock |
|-------|------|------|
| `registry.go:28` | `r.mu.Lock()` — adquire lock write | Write |
| `registry.go:29` | `defer r.mu.Unlock()` — libera em retorno | Write |
| `registry.go:30` | `r.creators[providerType] = creator` — insere no map | Write |

**Detalhe de concorrência**: Durante o init, não há concorrência — é sequencial. O lock é um **guarda de futuro** para casos onde `Register` é chamado em runtime (ex: plugins, hot-reload).

### Fases de Registro

| Provider | Factory | Arquivo Concreto |
|----------|---------|------------------|
| `openai` | `func(cfg) { return openai.NewClient(cfg) }` | `internal/platform/openai/` |
| `anthropic` | `func(cfg) { return anthropic.NewClient(cfg) }` | `internal/platform/anthropic/` |
| `gemini` | `func(cfg) { return geminiapi.NewClient(cfg) }` | `internal/platform/geminiapi/` |

---

## Detalhamento: Fase 2 — Criação (runtime)

### Fluxo `Create(cfg)`

```go
func (r *Registry) Create(cfg Config) (Provider, error) {
    r.mu.RLock()
    creator, ok := r.creators[cfg.Provider]
    r.mu.RUnlock()
    
    if !ok {
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
    }
    return creator(cfg)
}
```

| Linha | Ação | Detalhe |
|-------|------|---------|
| `registry.go:35` | `r.mu.RLock()` | Read lock — não bloqueia outros readers |
| `registry.go:36` | Lookup no map | `r.creators[cfg.Provider]` |
| `registry.go:37` | `r.mu.RUnlock()` | Libera read lock imediatamente |
| `registry.go:39-40` | Verifica `ok` | Fora do lock — evita deadlock |
| `registry.go:42` | Chama `creator(cfg)` | Factory recebe Config completa |

### Padrão de Lock

| Operação | Lock Tipo | Motivo |
|----------|-----------|--------|
| `Register` | Write (exclusive) | Modifica o map |
| `Create` | Read (shared) | Só lê o map |
| `SupportedProviders` | Read (shared) | Só lê o map |
| `IsRegistered` | Read (shared) | Só lê o map |

**Importante**: O lock é liberado **antes** de chamar o factory (`creator(cfg)`). Isso permite que o factory seja executado sem bloquear outros reads.

---

## Detalhamento: Operações Auxiliares

### `SupportedProviders()`

```go
func (r *Registry) SupportedProviders() []ProviderType {
    r.mu.RLock()
    defer r.mu.RUnlock()
    result := make([]ProviderType, 0, len(r.creators))
    for pt := range r.creators {
        result = append(result, pt)
    }
    return result
}
```

- **Retorna cópia** — nunca expõe o map interno.
- **Ordem não garantida** — iteração de map Go é não determinística.
- **Thread-safe** — read lock protege a iteração.

### `IsRegistered(providerType)`

```go
func (r *Registry) IsRegistered(providerType ProviderType) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    _, ok := r.creators[providerType]
    return ok
}
```

- **Operação O(1)** — lookup direto no map.
- **Só verifica registro** — não verifica se o factory funcionaria com uma config específica.

---

## Fluxos de Erro

| Cenário | Erro | Causa |
|---------|------|-------|
| Provider não registrado | `"unsupported provider: <nome>"` | `Register()` nunca foi chamado para esse provider |
| Factory retorna erro | `<err do factory>` | Problema na criação concreta (ex: config inválida) |

**Nota**: `Create` não valida a config — isso é responsabilidade do chamador (`cfg.WithDefaults().Validate()`).

---

## Observações de Design

- **Factory pattern + Registry** — desacopla a criação de providers do registry. O registry só conhece `ProviderCreator` (interface func), não as implementações concretas.
- **Leitura fora do lock** — `Create` libera o read lock antes de chamar o factory. Isso é correto porque o map não muda entre o lookup e o call (a menos que `Register` seja chamado concorrentemente).
- **Sem `Unregister`** — registry é imutável após init. Isso simplifica a concorrência (não precisa de write lock durante runtime).
