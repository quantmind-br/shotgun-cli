# Análise de Código — internal/platform/llmbase

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/platform/llmbase`                                    |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/platform/llmbase` |
| **Linguagem**     | Go                                                             |
| **Versão Go**     | 1.24.0                                                         |
| **Arquivos fonte**| `base_client.go`, `sender.go`                                  |
| **Arquivos de test** | Nenhum                                                       |
| **Dep. externas** | `context`, `encoding/json`, `fmt`, `time` (padrão Go)         |
| **Dep. internas** | `internal/core/llm`, `internal/platform/http`                 |
| **Coesão**        | Alta — foca exclusivamente em infraestrutura comum de provedores HTTP-based LLM |
| **Acoplamento**   | Médio — depende de `internal/platform/http` e `internal/core/llm` |

---

## 1. Visão Geral

O pacote `internal/platform/llmbase` é a **camada de infraestrutura compartilhada** para todos os provedores LLM HTTP-based do shotgun-cli (OpenAI, Anthropic, Gemini). Ele existe separadamente de `internal/core/llm` para **evitar ciclos de importação**: os provedores concretos em `internal/platform/` precisam usar a infraestrutura HTTP, mas não podem ser importados por `internal/core/llm` (que é de nível 0).

O pacote implementa dois artefatos principais:

1. **`BaseClient`** (`base_client.go`) — Struct que encapsula toda a lógica comum de comunicação HTTP com APIs de LLMs: criação do cliente HTTP, envio de requisições, parseamento de respostas, tratamento de erros HTTP e callbacks de progresso.
2. **`Sender`** (`sender.go`) — Interface Strategy que provedores concretos implementam para definir comportamento específico de cada API (build de request, parse de response, endpoint, headers).

**Dependência direta externa:** zero pacotes externos — apenas `internal/core/llm` e `internal/platform/http` (ambos internos) e bibliotecas padrão Go.

Este módulo é um **pilar de infraestrutura de nível 1** — é importado exclusivamente pelos provedores concretos em `internal/platform/openai`, `internal/platform/anthropic` e `internal/platform/geminiapi`.

---

## 2. Arquivo `sender.go`

### 2.1 Estrutura

Este arquivo contém **uma única definição**: a interface `Sender`.

#### `Sender` (interface)

Interface Strategy que define as operações provider-specific. BaseClient delega todas as variações (formato de request, parse de response, endpoint, headers, tipo de response) ao concrete sender.

| Método                      | Retorno                                    | Descrição                                          |
|-----------------------------|--------------------------------------------|----------------------------------------------------|
| `BuildRequest(content)`     | `(interface{}, error)`                     | Cria o payload específico do provider               |
| `ParseResponse(response, rawJSON)` | `(*llm.Result, error)`                | Extrai `llm.Result` da resposta crua da API         |
| `GetEndpoint()`             | `string`                                   | Path da API (ex: `/v1/chat/completions`)            |
| `GetHeaders()`              | `map[string]string`                        | Headers HTTP provider-specific                      |
| `NewResponse()`             | `interface{}`                              | Instância vazia do struct de response para unmarshal |
| `GetProviderName()`         | `string`                                   | Nome de exibição (ex: "OpenAI", "Anthropic")        |

### 2.2 Observações de Design

- **Padrão Strategy** — BaseClient implementa o "template" genérico, cada provider implementa `Sender` para customizar o que varia.
- **`ParseResponse` recebe `rawJSON []byte`** — permite logging/debugging da resposta crua além do parsed struct.
- **`NewResponse()` retorna `interface{}`** — o caller faz type assertion. Não há generic porque Go 1.22 ainda não tem generic constraints em métodos.
- **`GetEndpoint()` para Gemini inclui API key na query string** — o método retorna `"/models/{model}:generateContent?key={key}"`. Documentação do comentário confirma que é intencional para Gemini.
- **Sem testes** — nenhum arquivo `*_test.go` neste pacote.

---

## 3. Arquivo `base_client.go`

### 3.1 Estrutura

Contém a struct `BaseClient` e seus métodos, além de structs auxiliares de configuração.

#### `BaseClient` (struct)

| Campo         | Tipo                        | Descrição                                            |
|---------------|-----------------------------|------------------------------------------------------|
| `JSONClient`  | `*platformhttp.JSONClient`  | Cliente HTTP reutilizável para todas as requisições   |
| `APIKey`      | `string`                    | Chave de API do provedor                             |
| `Model`       | `string`                    | Nome do modelo LLM                                   |
| `MaxTokens`   | `int`                       | Limite máximo de tokens na resposta                  |
| `ProviderName`| `string`                    | Nome legível do provedor (ex: "OpenAI")              |

#### `Config` (struct)

Configuração minimalista para `NewBaseClient()`.

| Campo   | Tipo           | Obrigatório | Padrão              | Descrição                |
|---------|----------------|-------------|---------------------|--------------------------|
| `APIKey`| `string`       | Sim         | —                   | Chave de API             |
| `BaseURL`| `string`      | Sim         | —                   | URL base da API          |
| `Model` | `string`       | Sim         | —                   | Nome do modelo            |
| `MaxTokens` | `int`     | Não         | `0` (sem limite)    | Tokens máximos           |
| `Timeout`| `time.Duration` | Não       | `300s`              | Timeout da requisição    |

#### `DefaultConfig` (struct)

Valores padrão provider-specific para `NewBaseClientWithDefaults()`.

| Campo   | Tipo           | Descrição                      |
|---------|----------------|--------------------------------|
| `BaseURL`   | `string`  | URL base do provider           |
| `Model`     | `string`  | Modelo padrão do provider      |
| `MaxTokens` | `int`     | Tokens máximos padrão          |
| `Timeout`   | `time.Duration` | Timeout padrão             |

#### Fábricas

| Função | Retorno | Descrição |
|--------|---------|-----------|
| `NewBaseClient(cfg, providerName)` | `*BaseClient` | Cria BaseClient direto com Config. Sem validação de API key. Timeout default 300s. |
| `NewBaseClientWithDefaults(cfg, defaults, providerName)` | `(*BaseClient, error)` | Cria a partir de `llm.Config`, aplica defaults, valida API key. Retorna erro se API key vazia. |

#### Métodos de `BaseClient`

| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Name()` | `string` | Retorna `ProviderName` |
| `IsAvailable()` | `bool` | Sempre `true` — provedores HTTP sempre "disponíveis" se há internet |
| `IsConfigured()` | `bool` | `true` se `APIKey != ""` e `Model != ""` |
| `ValidateConfig()` | `error` | Valida `APIKey` e `Model` não vazios |
| `Send(ctx, content, sender)` | `(*llm.Result, error)` | Fluxo principal: build request → POST → parse response → retorna Result |
| `SendWithProgress(ctx, content, sender, progress)` | `(*llm.Result, error)` | Wrapper de `Send` com callbacks de progresso ("Connecting...", "Response received") |
| `HandleHTTPError(err, parseBody)` | `error` | Converte `*platformhttp.HTTPError` para mensagem legível; passa erro não-HTTP intacto |

### 3.2 Fluxo Detalhado de `Send()`

```
1. startTime := time.Now()
2. reqBody := sender.BuildRequest(content)          ← provider-specific
3. headers := sender.GetHeaders()                     ← provider-specific
4. endpoint := sender.GetEndpoint()                   ← provider-specific
5. response := sender.NewResponse()                   ← provider-specific
6. c.JSONClient.PostJSON(endpoint, headers, reqBody, response)
7. rawJSON := json.Marshal(response)                  ← para ParseResponse
8. result := sender.ParseResponse(response, rawJSON)  ← provider-specific
9. result.Duration = time.Since(startTime)
10. return result, nil
```

Cada etapa pode falhar com erro:
- Etapa 2: erro de build do request
- Etapa 6: erro HTTP (timeout, status != 200, etc.)
- Etapa 8: erro de parse (struct incompatível, campo ausente)

### 3.3 Observações de Design

- **Embedding de `*llmbase.BaseClient`** — os clientes concretos (OpenAI, Anthropic, Gemini) embed BaseClient, herdando implicitamente `Name()`, `IsAvailable()`, `IsConfigured()`, `ValidateConfig()`.
- **`IsAvailable()` sempre `true`** — não há verificação de conectividade real. Isso é uma simplificação que não detecta quedas de rede antes do POST.
- **`json.Marshal` silencioso** — `rawJSON, _ := json.Marshal(response)` ignora erro de marshal. Isso é seguro pois a struct já veio de json.Unmarshal via PostJSON, mas não é ideal.
- **Sem retry / circuit breaker** — nenhuma lógica de retry em `Send()`. Falhas HTTP são propagadas diretamente.
- **HandleHTTPError é parcial** — usa função `parseBody func([]byte) string` para extrair mensagem do body JSON, mas erros não-HTTPError são passados intactos.

### 3.4 Testes

| Arquivo de teste | Status |
|------------------|--------|
| `base_client_test.go` | ❌ Não existe |
| `sender_test.go` | ❌ Não existe |

**Nenhum teste unitário** no pacote `llmbase`. A cobertura é feita implicitamente pelos testes dos providers concretos (openai, anthropic, geminiapi).

---

## 4. Diagrama de Dependências

```
llmbase
├── internal/core/llm        (Result, Usage, Provider)
└── internal/platform/http   (JSONClient, ClientConfig, HTTPError)

Provedores que importam llmbase:
├── internal/platform/openai (Client embeds *BaseClient)
├── internal/platform/anthropic (Client embeds *BaseClient)
└── internal/platform/geminiapi (Client embeds *BaseClient)
```

---

## 5. Resumo de Métricas

| Arquivo         | Linhas (aprox.) | Tipos definidos | Funções/métodos | Testes |
|-----------------|------------------|-----------------|-----------------|--------|
| `sender.go`     | 33               | 1 (Sender interface, 6 métodos) | 0 | 0 |
| `base_client.go`| 138              | 3 (BaseClient, Config, DefaultConfig) | 9 (2 fábricas + 7 métodos) | 0 |
| **Total**       | **171**          | **4**           | **9**           | **0**  |

---

## 6. Dívida Técnica / Observações

| # | Problema | Severidade | Descrição |
|---|----------|------------|-----------|
| 1 | **Sem testes** | crítica | Nenhum teste unitário neste pacote crítico de infraestrutura. Erros aqui afetam todos os 3 providers. |
| 2 | **`IsAvailable()` sempre `true`** | média | Não detecta perda de conectividade antes do POST. Poderia fazer HEAD simples. |
| 3 | **Sem retry / circuit breaker** | crítica | Falhas HTTP (timeout, 5xx) são propagadas imediatamente. Sem tentativa de reconexão. |
| 4 | **Sem rate limiting** | crítica | Não há controle de taxa de chamadas. Risco de throttling pela API. |
| 5 | **`json.Marshal` silencioso** | baixa | `rawJSON, _ := json.Marshal(response)` ignora erro de marshal. Seguro na prática (response vem de unmarshal válido), mas anti-pattern. |
| 6 | **`SendWithProgress` limitado** | baixa | Só tem 2 callbacks: "Connecting..." e "Response received". Não informa progresso do upload/download. |
| 7 | **Gemini API key na URL** | média | `GetEndpoint()` inclui API key no path query. Aparece em logs de HTTP. Anthropic usa header, OpenAI usa header — inconsistência de segurança. |
| 8 | **Config vs DefaultConfig duplicação** | baixa | `Config` (em base_client.go) e `DefaultConfig` têm campos sobrepostos. Podem ser unificados. |
| 9 | **HandleHTTPError pode retornar nil** | baixa | Se o body do erro estiver vazio ou parse falhar, retorna `""` → `fmt.Errorf("API error [%d]: ")`. Mensagem final é truncada. |
