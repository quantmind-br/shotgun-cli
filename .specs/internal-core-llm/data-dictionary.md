# Dicionário de Dados — internal/core/llm

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/core/llm`                                            |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/core/llm`        |
| **Document Language** | Português (pt-br)                                           |

---

## 1. Tipos do Domínio

### 1.1 `ProviderType` (string typed)

Tipo string que identifica qual provedor LLM está sendo utilizado.

| Valor        | Constante        | Provedor                   |
|--------------|------------------|----------------------------|
| `"openai"`   | `ProviderOpenAI`   | OpenAI (Chat Completions API)  |
| `"anthropic"`| `ProviderAnthropic`| Anthropic (Messages API)       |
| `"gemini"`   | `ProviderGemini`   | Google Gemini (Generative Language API) |

**Método**: `func (p ProviderType) String() string` — retorna o valor string diretamente.

**Uso**: Chave do map `Registry.creators`, campo `Config.Provider`, entrada em `AllProviders()`.

### 1.2 `Provider` (interface)

Contrato para todos os provedores LLM.

| Método                      | Retorno                              | Descrição                                          |
|-----------------------------|--------------------------------------|----------------------------------------------------|
| `Send(ctx, content)`        | `(*Result, error)`                   | Envia prompt e retorna resposta                    |
| `SendWithProgress(ctx, content, progress)` | `(*Result, error)` | Envio assíncrono com callback de progresso        |
| `Name()`                    | `string`                             | Nome legível do provedor (ex: "OpenAI")            |
| `IsAvailable()`             | `bool`                               | Verifica se binário/ferramenta está disponível     |
| `IsConfigured()`            | `bool`                               | Verifica se API key está presente                  |
| `ValidateConfig()`          | `error`                              | Valida a configuração do provider antes do uso     |

**Implementações concretas** (fora deste módulo): `openai.Client`, `anthropic.Client`, `geminiapi.Client` (em `internal/platform/`).

---

## 2. Estruturas de Dados

### 2.1 `Config`

Configuração unificada para qualquer provedor LLM.

| Campo         | Tipo       | Obrigatório | Padrão                      | Descrição                                                |
|---------------|------------|-------------|-----------------------------|----------------------------------------------------------|
| `Provider`    | `ProviderType` | Sim       | —                           | Provedor a utilizar (openai, anthropic, gemini)         |
| `APIKey`      | `string`   | Sim         | —                           | Chave de API para autenticação                           |
| `BaseURL`     | `string`   | Não         | `DefaultConfigs()[Provider]`| URL base da API (permite customização/proxy)            |
| `Model`       | `string`   | Sim         | `DefaultConfigs()[Provider]`| Nome do modelo (ex: "gpt-4o", "claude-sonnet-4-20250514") |
| `Timeout`     | `int`      | Sim         | `DefaultConfigs()[Provider]`| Timeout em segundos                                      |
| `MaxTokens`   | `int`      | Não         | `0` (sem limite)            | Número máximo de tokens na resposta                      |
| `Temperature` | `float64`  | Não         | `0.0`                       | Aleatoriedade (0.0 = determinístico, 2.0 = máximo)      |

**Validação** (`Validate()`):
1. `Provider` deve ser non-zero
2. `Provider` deve estar em `AllProviders()`
3. `APIKey` deve ser não vazio
4. `Model` deve ser não vazio
5. `BaseURL` (se não vazio) deve ser URL parseável
6. `Timeout` deve ser > 0

**Comportamento de máscara** (`MaskAPIKey()`):
| Condição                    | Resultado              |
|-----------------------------|------------------------|
| `APIKey == ""`              | `"(not configured)"`   |
| `len(APIKey) <= 8`          | `"***"`                |
| `len(APIKey) > 8`           | `APIKey[:4] + "..." + APIKey[len-4:]` |

---

### 2.2 `Result`

Resultado de uma chamada LLM via `Provider.Send()` ou `Provider.SendWithProgress()`.

| Campo         | Tipo            | Obrigatório | Descrição                                       |
|---------------|-----------------|-------------|-------------------------------------------------|
| `Response`    | `string`        | Sim         | Resposta processada/limpa para o usuário         |
| `RawResponse` | `string`        | Não         | Resposta crua da API (para debugging)           |
| `Model`       | `string`        | Sim         | Nome exato do modelo que gerou a resposta       |
| `Provider`    | `string`        | Sim         | Nome legível do provedor                        |
| `Duration`    | `time.Duration` | Sim         | Tempo total da chamada                          |
| `Usage`       | `*Usage`        | Não         | Métricas de tokens (pode ser nil)               |

---

### 2.3 `Usage`

Métricas de uso de tokens da API.

| Campo              | Tipo   | Obrigatório | Descrição                          |
|--------------------|--------|-------------|------------------------------------|
| `PromptTokens`     | `int`  | Sim         | Tokens enviados no prompt          |
| `CompletionTokens` | `int`  | Sim         | Tokens gerados na resposta         |
| `TotalTokens`      | `int`  | Sim         | Soma dos dois anteriores           |

**Nota**: `TotalTokens` deve ser igual a `PromptTokens + CompletionTokens`. Campo é ponteiro (`*Usage`) pois nem todos os providers reportam métricas.

---

### 2.4 `ProviderCreator` (type alias)

```go
type ProviderCreator func(cfg Config) (Provider, error)
```

Função factory registrada no `Registry`. Recebe a `Config` do usuário e retorna uma instância concreta do `Provider` ou erro.

**Responsabilidade**: Construir o objeto provider com a configuração (ex: `openai.NewClient(cfg)`).

---

### 2.5 `Registry`

Armazena e gerencia os factories de providers.

| Campo      | Tipo                              | Obrigatório | Descrição                                          |
|------------|-----------------------------------|-------------|----------------------------------------------------|
| `mu`       | `sync.RWMutex`                    | Sim         | Lock para acesso concorrente ao map                |
| `creators` | `map[ProviderType]ProviderCreator`| Sim         | Map de provider → factory                          |

**Invariantes**:
- `creators` é sempre inicializado (nunca nil) — `NewRegistry()` cria um map vazio, nunca nil.
- Após `Register(pt, creator)`, `IsRegistered(pt)` retorna `true` permanentemente (não há `Unregister`).

---

## 3. Constantes e Valores Padrão

### 3.1 Provider Constants

| Constante        | Valor        | Provider                   |
|------------------|--------------|----------------------------|
| `ProviderOpenAI`   | `"openai"`   | OpenAI Chat Completions    |
| `ProviderAnthropic`| `"anthropic"`| Anthropic Messages         |
| `ProviderGemini`   | `"gemini"`   | Google Gemini              |

### 3.2 Default Configurations

| Provider       | BaseURL                                        | Modelo                      | Timeout |
|----------------|------------------------------------------------|-----------------------------|---------|
| `openai`       | `https://api.openai.com/v1`                   | `gpt-4o`                    | 300s    |
| `anthropic`    | `https://api.anthropic.com`                   | `claude-sonnet-4-20250514`  | 300s    |
| `gemini`       | `https://generativelanguage.googleapis.com/v1beta` | `gemini-2.5-flash`     | 300s    |

### 3.3 AllProviders (lista imutável)

```
[openai, anthropic, gemini]
```

Retornada por `AllProviders()` em ordem fixa (conforme declaração no código).

---

## 4. Resumo de Todos os Tipos

| Tipo         | Categoria | Origem          | Campos/Elementos |
|--------------|-----------|-----------------|------------------|
| `ProviderType`    | Enum      | `provider.go`    | 3 constantes + String() |
| `Provider`       | Interface | `provider.go`    | 6 métodos         |
| `Result`          | Struct    | `provider.go`    | 6 campos (5 obrigatórios) |
| `Usage`           | Struct    | `provider.go`    | 3 campos (todos obrigatórios) |
| `Config`          | Struct    | `config.go`      | 7 campos (5 obrigatórios) |
| `ProviderCreator` | Alias func | `registry.go`   | 1 assinatura      |
| `Registry`        | Struct    | `registry.go`    | 2 campos + 5 métodos |
