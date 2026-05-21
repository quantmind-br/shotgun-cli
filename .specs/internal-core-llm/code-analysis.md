# Análise de Código — internal/core/llm

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/core/llm`                                            |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/core/llm`        |
| **Linguagem**     | Go                                                             |
| **Versão Go**     | 1.24.0                                                         |
| **Arquivos fonte**| `config.go`, `provider.go`, `registry.go`                      |
| **Arquivos de test** | `config_test.go`, `provider_test.go`, `registry_test.go`    |
| **Dep. externas** | `fmt`, `net/url`, `context`, `time`, `sync` (padrão Go)       |
| **Dep. de test**  | `github.com/stretchr/testify`                                  |
| **Coesão**        | Alta — todos os arquivos tratam exclusivamente do domínio LLM   |
| **Acoplamento**   | Baixo — módulo de nível 0 (zero dependências de outros internos) |

---

## 1. Visão Geral

O pacote `internal/core/llm` é o **núcleo de abstração de provedores LLM** do shotgun-cli. Ele define:

1. **Configuração unificada** (`config.go`) — Struct `Config` que abstrai parâmetros comuns a qualquer provedor (provider, API key, base URL, modelo, timeout, tokens máximos, temperatura), funções de validação, máscara de chave de API e aplicação de valores padrão.
2. **Interface Provider + Tipos** (`provider.go`) — Interface `Provider` com métodos `Send`, `SendWithProgress`, `Name`, `IsAvailable`, `IsConfigured`, `ValidateConfig`. Tipos `Result`, `Usage` e a enumeração `ProviderType` com constantes para os três provedores suportados: `openai`, `anthropic`, `gemini`.
3. **Registro de Provedores** (`registry.go`) — `Registry` com `sync.RWMutex` para concorrência segura, usando um `map[ProviderType]ProviderCreator` para lookup e criação de instâncias de provedores via factory pattern.

**Dependência direta externa:** zero. Apenas pacotes da biblioteca padrão Go. **Dependência transitiva:** zero.

Este módulo é um **pilar de infraestrutura**: é importado por `internal/app`, `internal/ui` e possivelmente `cmd/`, mas não importa nenhum pacote interno do shotgun-cli.

---

## 2. Arquivo `provider.go`

### 2.1 Estrutura

Este arquivo é o **centro da abstração de domínio**. Define os tipos e a interface que todas as implementações concretas (OpenAI, Anthropic, Gemini) devem seguir.

#### `Result` (struct)

Resultado de uma chamada LLM.

| Campo         | Tipo            | Obrigatório | Descrição                                       |
|---------------|-----------------|-------------|-------------------------------------------------|
| `Response`    | `string`        | Sim         | Resposta processada/limpa                        |
| `RawResponse` | `string`        | Não         | Resposta bruta da API                           |
| `Model`       | `string`        | Sim         | Nome do modelo usado                             |
| `Provider`    | `string`        | Sim         | Nome do provedor                                 |
| `Duration`    | `time.Duration` | Sim         | Tempo de execução                                |
| `Usage`       | `*Usage`        | Não         | Métricas de uso (tokens)                         |

#### `Usage` (struct)

Métricas de tokens consumidos na chamada.

| Campo              | Tipo   | Obrigatório | Descrição                          |
|--------------------|--------|-------------|------------------------------------|
| `PromptTokens`     | `int`  | Sim         | Tokens no prompt                   |
| `CompletionTokens` | `int`  | Sim         | Tokens na resposta                 |
| `TotalTokens`      | `int`  | Sim         | Total de tokens consumidos         |

#### `Provider` (interface)

Contrato implementado por todos os provedores concretos.

| Método                      | Retorno                              | Descrição                                          |
|-----------------------------|--------------------------------------|----------------------------------------------------|
| `Send(ctx, content)`        | `(*Result, error)`                   | Envia prompt e retorna resposta                    |
| `SendWithProgress(ctx, content, progress)` | `(*Result, error)` | Envio com callback de progresso (TUI)              |
| `Name()`                    | `string`                             | Nome do provedor (ex: "OpenAI")                    |
| `IsAvailable()`             | `bool`                               | Verifica se o provedor está disponível (binário)   |
| `IsConfigured()`            | `bool`                               | Verifica se está configurado (API key presente)    |
| `ValidateConfig()`          | `error`                              | Validação da configuração antes do uso             |

#### `ProviderType` (string typed) + constantes

| Constante        | Valor        | Provedor        |
|------------------|--------------|-----------------|
| `ProviderOpenAI`   | `"openai"`   | OpenAI Chat Completions |
| `ProviderAnthropic`| `"anthropic"`| Anthropic Messages    |
| `ProviderGemini`   | `"gemini"`   | Google Gemini           |

Funções de conveniência:

| Função                   | Retorno                     | Descrição                                    |
|--------------------------|-----------------------------|----------------------------------------------|
| `AllProviders()`         | `[]ProviderType`            | Lista imutável dos 3 provedores suportados    |
| `IsValidProvider(p)`     | `bool`                      | Verifica se string é um provider válido       |
| `(p ProviderType).String()` | `string`                | Conversão explícita para string               |

### 2.2 Observações de Design

- **Interface pura (duck typing)** — não há struct base; cada provider concreto implementa a interface explicitamente.
- **ProviderType como string typed** — permite comparação direta com strings literais e serialização natural.
- **`Usage` é `*Usage` (ponteiro)** — opcional; alguns providers podem não reportar métricas.
- **Sem construtor explícito** — a criação de providers é delegada ao `Registry` (factory pattern).

### 2.3 Testes (`provider_test.go`)

| Teste                     | Propósito                                        |
|---------------------------|--------------------------------------------------|
| `TestIsValidProvider`     | Valida 6 casos: openai, anthropic, gemini, invalid, vazio, case-sensitive |
| `TestAllProviders`        | Garante 3 providers e presença dos 3 esperados   |
| `TestProviderTypeString`  | Verifica conversão string dos 3 provider types   |

Test fixtures: `mockProvider` — struct anônima em `*_test.go` que implementa `Provider` para testes de `Registry`.

---

## 3. Arquivo `config.go`

### 3.1 Estrutura

Contém a struct de configuração unificada e seus métodos auxiliares.

#### `Config` (struct)

| Campo         | Tipo       | Obrigatório | Descrição                                                |
|---------------|------------|-------------|----------------------------------------------------------|
| `Provider`    | `ProviderType` | Sim       | Qual provedor utilizar                                   |
| `APIKey`      | `string`   | Sim         | Chave de API para autenticação                           |
| `BaseURL`     | `string`   | Não         | URL base da API (permite endpoints customizados/proxy)   |
| `Model`       | `string`   | Sim         | Nome do modelo a utilizar                                |
| `Timeout`     | `int`      | Sim         | Timeout em segundos                                      |
| `MaxTokens`   | `int`      | Não         | Máximo de tokens na resposta                             |
| `Temperature` | `float64`  | Não         | Temperatura (0.0–2.0), controla aleatoriedade            |

#### `DefaultConfigs()` — mapa de configs padrão por provider

| Provider       | BaseURL                                        | Modelo                   | Timeout |
|----------------|------------------------------------------------|--------------------------|---------|
| `ProviderOpenAI`  | `https://api.openai.com/v1`              | `gpt-4o`              | 300s    |
| `ProviderAnthropic`| `https://api.anthropic.com`            | `claude-sonnet-4-20250514` | 300s  |
| `ProviderGemini`  | `https://generativelanguage.googleapis.com/v1beta` | `gemini-2.5-flash` | 300s |

#### Métodos de `Config`

| Método              | Retorno    | Descrição                                                  |
|---------------------|------------|------------------------------------------------------------|
| `Validate()`        | `error`    | Validação completa: provider, provider válido, API key, modelo, URL, timeout > 0 |
| `MaskAPIKey()`      | `string`   | Mascara chave de API para exibição segura                   |
| `WithDefaults()`    | `*Config`  | Aplica valores padrão (BaseURL, Model, Timeout) por provider |

#### Lógica de `MaskAPIKey()`

| Condição                         | Saída              |
|----------------------------------|--------------------|
| `APIKey == ""`                   | `"(not configured)"` |
| `len(APIKey) <= 8`               | `"***"`             |
| `len(APIKey) > 8`                | `primeiros4 + "..." + últimos4` |

### 3.2 Observações de Design

- **`Validate()` é estrita** — todos os campos obrigatórios devem estar preenchidos; `BaseURL` é validado com `url.Parse` mas é opcional.
- **`WithDefaults()` é mutativo** — modifica o receiver e retorna `*Config` para chaining.
- **`DefaultConfigs()` retorna um novo mapa a cada chamada** — não há memoização; cada chamada é independente.
- **Temperatura não é validada** — o range 0.0–2.0 é apenas um comentário; não há validação de faixa em `Validate()`.

### 3.3 Testes (`config_test.go`)

| Teste                      | Propósito                                                        |
|----------------------------|------------------------------------------------------------------|
| `TestConfigValidate`       | 10 cenários: 3 válidos (openai, anthropic, gemini), 5 de erro (missing provider, invalid provider, missing api-key, missing model, invalid timeout), 2 de timeout negativo |
| `TestConfigMaskAPIKey`     | 5 cenários: vazio, curto (3 chars), 8 chars, normal, longo |
| `TestDefaultConfigs`       | Garante 3 entries e valores específicos de cada provider         |
| `TestConfigWithDefaults`   | Verifica que `WithDefaults()` preenche BaseURL, Model e Timeout quando vazios |

---

## 4. Arquivo `registry.go`

### 4.1 Estrutura

Implementa o registry de provedores com factory pattern e segurança concorrente.

#### `ProviderCreator` (type alias)

```go
type ProviderCreator func(cfg Config) (Provider, error)
```

Função factory que recebe `Config` e retorna uma `Provider` concreta ou erro.

#### `Registry` (struct)

| Campo      | Tipo                              | Descrição                                          |
|------------|-----------------------------------|----------------------------------------------------|
| `mu`       | `sync.RWMutex`                    | Lock para acesso concorrente ao map de criadores   |
| `creators` | `map[ProviderType]ProviderCreator`| Map provider → factory function                   |

#### Métodos de `Registry`

| Método                  | Retorno                          | Descrição                                                    |
|-------------------------|----------------------------------|--------------------------------------------------------------|
| `NewRegistry()`         | `*Registry`                     | Cria registry vazio com map inicializado                      |
| `Register(pt, creator)` | `void`                           | Registra um factory para um provider (write lock)            |
| `Create(cfg)`           | `(Provider, error)`             | Busca factory por `cfg.Provider` e a invoca (read lock)     |
| `SupportedProviders()`  | `[]ProviderType`                 | Retorna lista de providers registrados (read lock)            |
| `IsRegistered(pt)`      | `bool`                           | Verifica se provider está registrado (read lock)              |

### 4.2 Observações de Design

- **Concorrência segura** — `sync.RWMutex` com locks diferenciados: `Register` usa write lock, todas as outras usam read lock.
- **`Create` falha silenciosamente** — se `cfg.Provider` não estiver registrado, retorna `(nil, error)` com mensagem `"unsupported provider: <nome>"`.
- **`SupportedProviders` retorna cópia** — não expõe o map internamente, retornando um novo `[]ProviderType` com os keys do map.
- **Sem cleanup** — não há método `Unregister` ou `Clear`; providers são registrados apenas via `Register`.

### 4.3 Testes (`registry_test.go`)

| Teste                      | Propósito                                                        |
|----------------------------|------------------------------------------------------------------|
| `TestRegistry`             | 1) vazio → 2) registra 1 provider → 3) cria provider → 4) cria unregistered (erro) |
| `TestRegistryMultipleProviders` | Registra 2 providers, verifica `SupportedProviders()`, `IsRegistered` e `IsRegistered(false)` |

Test fixtures: `mockProvider` — implementa `Provider` com stub, usado para testes de criação via Registry.

---

## 5. Resumo de Métricas

| Arquivo        | Linhas (aprox.) | Tipos definidos | Funções/métodos | Testes | Cobertura implícita |
|----------------|------------------|-----------------|-----------------|--------|---------------------|
| `provider.go`  | 62               | 5 (Result, Usage, Provider, ProviderType + consts) | 4 (AllProviders, IsValidProvider, String, Provider interface 6 métodos) | 3 | 100% (3 funções cobertas) |
| `config.go`    | 74               | 1 (Config)      | 4 (DefaultConfigs, Validate, MaskAPIKey, WithDefaults) | 4 | 100% (todos os métodos cobertos) |
| `registry.go`  | 53               | 2 (ProviderCreator, Registry) | 5 (NewRegistry, Register, Create, SupportedProviders, IsRegistered) | 2 | 100% (todos os métodos cobertos) |
| **Total**      | **189**          | **8**           | **13**          | **9**  | **100%** |

---

## 6. Debt Técnico / Observações

| # | Problema | Severidade | Descrição |
|---|----------|------------|-----------|
| 1 | **Temperatura sem validação** | média | `Validate()` não verifica se `Temperature` está no range 0.0–2.0 |
| 2 | **MaxTokens sem validação** | baixa | `MaxTokens` é opcional e sem validação mínima/máxima |
| 3 | **Sem retry / circuit breaker** | crítica | Nenhum mecanismo de retry ou fallback em `Provider.Send` |
| 4 | **Sem rate limiting** | crítica | Nenhum controle de rate limiting nas chamadas LLM |
| 5 | **`DefaultConfigs` sem memoização** | baixa | Mapa recriado a cada chamada; poderia ser `sync.Once` ou `var` pacotes |
| 6 | **Sem método `Unregister`** | baixa | Registry é imutável após registro; não suporta hot-reload |
