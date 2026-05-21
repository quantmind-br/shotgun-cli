# Dicionário de Dados — Módulo `cmd`

> **Módulo:** `cmd`  
> **Caminho do pacote:** `github.com/quantmind-br/shotgun-cli/cmd`  
> **Nível de detalhe:** detalhado  
> **Data:** 2026-05-20  
> **Gerado por:** reversa-archaeologist

---

## 1. Tipos Definidos no Pacote `cmd`

### 1.1 `ProgressMode` (string type)

| Campo | Tipo | Descrição |
|---|---|---|
| — | `string` | Tipo definido para controlar o formato de saída de progresso |

**Valores permitidos:**

| Valor | Constante | Significado |
|---|---|---|
| `"none"` | `ProgressNone` | Sem saída de progresso |
| `"human"` | `ProgressHuman` | Saída legível por humano (`\r` overwrite) |
| `"json"` | `ProgressJSON` | Saída JSON, uma linha por evento |

**Onde é usado:** Flag `--progress` do comando `context generate`.

---

### 1.2 `ProgressOutput` (struct)

```go
type ProgressOutput struct {
    Timestamp string  `json:"timestamp"`
    Stage     string  `json:"stage"`
    Message   string  `json:"message"`
    Current   int64   `json:"current,omitempty"`
    Total     int64   `json:"total,omitempty"`
    Percent   float64 `json:"percent,omitempty"`
}
```

| Campo | Tipo | JSON | Obrigatório | Descrição |
|---|---|---|---|---|
| `Timestamp` | `string` | `"timestamp"` | Sim | Timestamp RFC3339 do evento |
| `Stage` | `string` | `"stage"` | Sim | Etapa da operação (ex: "scanning", "generating") |
| `Message` | `string` | `"message"` | Sim | Mensagem legível |
| `Current` | `int64` | `"current"` | Não (omitempty) | Itens processados atualmente |
| `Total` | `int64` | `"total"` | Não (omitempty) | Total de itens |
| `Percent` | `float64` | `"percent"` | Não (omitempty) | Porcentagem (0.0–100.0) |

**Estágios conhecidos:** `scanning`, `generating`, `saving`, `complete`, `loading`, `starting`.

---

### 1.3 `GenerateConfig` (struct)

```go
type GenerateConfig struct {
    RootPath     string
    Include      []string
    Exclude      []string
    Output       string
    MaxSize      int64
    EnforceLimit bool
    Template     string
    Task         string
    Rules        string
    CustomVars   map[string]string
    Workers      int
    IncludeHidden  bool
    IncludeIgnored bool
    ProgressMode ProgressMode
}
```

| Campo | Tipo | Padrão | Descrição |
|---|---|---|---|
| `RootPath` | `string` | `"."` | Diretório raiz da varredura (convertido para absoluto) |
| `Include` | `[]string` | `["*"]` | Padrões glob de arquivos para incluir |
| `Exclude` | `[]string` | `[]` | Padrões glob de arquivos para excluir |
| `Output` | `string` | `"shotgun-prompt-YYYYMMDD-HHMMSS.md"` | Arquivo de saída |
| `MaxSize` | `int64` | `10MB` (10485760) | Tamanho máximo do contexto em bytes |
| `EnforceLimit` | `bool` | `true` | Se deve falhar ao exceder o tamanho |
| `Template` | `string` | `""` (usar padrão) | Nome do template a usar |
| `Task` | `string` | `""` (default: "Context generation") | Descrição da tarefa para a LLM |
| `Rules` | `string` | `""` | Regras/constraints para a LLM |
| `CustomVars` | `map[string]string` | `{}` | Variáveis customizadas KEY=VALUE |
| `Workers` | `int` | `0` (usar config) | Override de workers do scanner |
| `IncludeHidden` | `bool` | `false` | Incluir arquivos ocultos |
| `IncludeIgnored` | `bool` | `false` | Incluir arquivos ignorados |
| `ProgressMode` | `ProgressMode` | `ProgressNone` | Modo de saída de progresso |

---

## 2. Chaves de Configuração Utilizadas

### 2.1 Scanner

| Chave Viper | Tipo Padrão | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `scanner.max-files` | `int64` | `10000` | — | Máximo de arquivos para varrer |
| `scanner.max-file-size` | `string` | `"1MB"` | — | Tamanho máximo por arquivo |
| `scanner.max-memory` | `string` | `"500MB"` | — | Limite de memória da varredura |
| `scanner.respect-gitignore` | `bool` | `true` | — | Respeitar `.gitignore` |
| `scanner.skip-binary` | `bool` | `true` | — | Pular arquivos binários |
| `scanner.workers` | `int` | `1` | `--workers` | Workers paralelos |
| `scanner.include-hidden` | `bool` | `false` | `--include-hidden` | Incluir arquivos ocultos |
| `scanner.respect-shotgunignore` | `bool` | `true` | — | Respeitar `.shotgunignore` |

### 2.2 Context

| Chave Viper | Tipo Padrão | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `context.include-tree` | `bool` | `true` | — | Incluir árvore de diretórios |
| `context.include-summary` | `bool` | `true` | — | Incluir sumários de arquivos |
| `context.max-size` | `string` | `"10MB"` | `--max-size` | Tamanho máximo do contexto |

### 2.3 LLM

| Chave Viper | Tipo Padrão | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `llm.provider` | `string` | `"openai"` | — | Provider: openai, anthropic, gemini |
| `llm.api-key` | `string` | `""` | — | Chave de API |
| `llm.base-url` | `string` | `""` (default do provider) | — | URL customizada da API |
| `llm.model` | `string` | `""` (default do provider) | `--model` | Modelo a usar |
| `llm.timeout` | `int` | `300` | `--timeout` | Timeout em segundos |
| `llm.save-response` | `bool` | `true` | — | Salvar resposta automaticamente |

### 2.4 Template

| Chave Viper | Tipo Padrão | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `template.custom-path` | `string` | `""` | — | Caminho para templates customizados |

### 2.5 Output

| Chave Viper | Tipo Padrão | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `output.format` | `string` | `"markdown"` | — | Formato de saída: markdown, text |
| `output.clipboard` | `bool` | `true` | — | Copiar para área de transferência |

### 2.6 Global

| Chave Viper | Tipo | Valor Padrão | Flag CLI | Descrição |
|---|---|---|---|---|
| `verbose` | `bool` | `false` | `-v` / `--verbose` | Saída verbosa (Debug) |
| `quiet` | `bool` | `false` | `-q` / `--quiet` | Saída silenciosa (Error apenas) |

---

## 3. Configurações de Template (Built-in)

### 3.1 Template Variables

| Variável | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `TASK` | `string` | Sim | Descrição da tarefa para a LLM |
| `RULES` | `string` | Não | Regras/constraints para a LLM |
| `FILE_STRUCTURE` | `string` | Não | Estrutura de arquivos (gerado automaticamente) |
| `CURRENT_DATE` | `string` | Não | Data atual (formato `YYYY-MM-DD`) |
| `{custom}` | `string` | Condiicional | Variáveis customizadas via `--var KEY=VALUE` |

### 3.2 Templates Embutidos

Os templates são carregados via `template.NewManager()` e listados via `template list`. Os nomes seguem o padrão embutido em `internal/assets/`.

---

## 4. Providers LLM

### 4.1 ProviderTypes

| Tipo | Constante | Nome Exibido | Modelo Padrão | URL Padrão | URL API Key |
|---|---|---|---|---|---|
| `"openai"` | `ProviderOpenAI` | OpenAI | `gpt-4o` | `https://api.openai.com/v1` | `https://platform.openai.com/api-keys` |
| `"anthropic"` | `ProviderAnthropic` | Anthropic | `claude-sonnet-4-20250514` | `https://api.anthropic.com` | `https://console.anthropic.com/settings/keys` |
| `"gemini"` | `ProviderGemini` | Google Gemini | `gemini-2.5-flash` | `https://generativelanguage.googleapis.com/v1beta` | `https://aistudio.google.com/app/apikey` |

### 4.2 Interface `llm.Provider`

| Método | Retorno | Descrição |
|---|---|---|
| `Send(ctx, content)` | `*Result, error` | Envia prompt, recebe resposta |
| `SendWithProgress(ctx, content, progress)` | `*Result, error` | Mesmo com callback de progresso |
| `Name()` | `string` | Nome exibido (e.g., "OpenAI") |
| `IsAvailable()` | `bool` | Provider está disponível? |
| `IsConfigured()` | `bool` | Provider está configurado? |
| `ValidateConfig()` | `error` | Validação de config |

### 4.3 Result `llm.Result`

| Campo | Tipo | Descrição |
|---|---|---|
| `Response` | `string` | Resposta processada/limpa |
| `RawResponse` | `string` | Resposta bruta da API |
| `Model` | `string` | Modelo usado |
| `Provider` | `string` | Nome do provider |
| `Duration` | `time.Duration` | Tempo de execução |
| `Usage` | `*Usage` | Métricas de uso de tokens |

### 4.4 Usage `llm.Usage`

| Campo | Tipo | Descrição |
|---|---|---|
| `PromptTokens` | `int` | Tokens no prompt |
| `CompletionTokens` | `int` | Tokens na resposta |
| `TotalTokens` | `int` | Total de tokens |

---

## 5. Tipos de Dados de Entrada (Flags)

### 5.1 `context generate`

| Flag | Tipo | Curto | Padrão | Descrição |
|---|---|---|---|---|
| `root` | `string` | `-r` | `"."` | Diretório raiz |
| `include` | `[]string` | `-i` | `["*"]` | Padrões de inclusão |
| `exclude` | `[]string` | `-e` | `[]` | Padrões de exclusão |
| `output` | `string` | `-o` | Auto | Arquivo de saída |
| `max-size` | `string` | — | `"10MB"` | Tamanho máximo |
| `enforce-limit` | `bool` | — | `true` | Aplicar limite |
| `template` | `string` | `-t` | `""` | Nome do template |
| `task` | `string` | — | `""` | Descrição da tarefa |
| `rules` | `string` | — | `""` | Regras para LLM |
| `var` | `[]string` | `-V` | `[]` | Variáveis KEY=VALUE |
| `workers` | `int` | — | `0` | Workers do scanner |
| `include-hidden` | `bool` | — | `false` | Incluir ocultos |
| `include-ignored` | `bool` | — | `false` | Incluir ignorados |
| `progress` | `string` | — | `"none"` | Modo: none/human/json |

### 5.2 `context send`

| Flag | Tipo | Curto | Padrão | Descrição |
|---|---|---|---|---|
| `output` | `string` | `-o` | `""` | Arquivo de resposta |
| `model` | `string` | `-m` | `""` | Modelo (override) |
| `timeout` | `int` | — | `0` | Timeout (override) |
| `raw` | `bool` | — | `false` | Resposta bruta |

### 5.3 `config set`

| Argumento | Tipo | Descrição |
|---|---|---|
| `[key]` | `string` | Chave de configuração (validada) |
| `[value]` | `string` | Valor (convertido ao tipo) |

### 5.4 `diff split`

| Flag | Tipo | Curto | Padrão | Obrigatório | Descrição |
|---|---|---|---|---|---|
| `input` | `string` | `-i` | — | Sim | Arquivo diff de entrada |
| `output-dir` | `string` | `-o` | `"chunks"` | Não | Diretório de saída |
| `approx-lines` | `int` | — | `500` | Não | Linhas aproximadas por chunk |
| `no-header` | `bool` | — | `false` | Não | Omitir headers |

### 5.5 `template render`

| Flag | Tipo | Curto | Padrão | Descrição |
|---|---|---|---|---|
| `[template-name]` | `string` | — | (positional) | Nome do template |
| `var` | `map[string]string` | — | `{}` | Variáveis KEY=VALUE |
| `output` | `string` | `-o` | `""` | Arquivo de saída (stdout se vazio) |

---

## 6. Caminhos de Arquivo

### 6.1 Configuração

| Plataforma | Caminho Padrão | Variável |
|---|---|---|
| Linux | `~/.config/shotgun-cli/config.yaml` | `XDG_CONFIG_HOME` |
| macOS | `~/Library/Application Support/shotgun-cli/config.yaml` | — |
| Windows | `%APPDATA%\shotgun-cli\config.yaml` | `APPDATA` |

### 6.2 Templates Customizados

| Caminho | Descrição |
|---|---|
| `~/.config/shotgun-cli/templates/` | Templates importados pelo usuário (Linux) |
| `{custom-path}` | Caminho customizado via `template.custom-path` |

### 6.3 Output Padrão

| Item | Formato | Exemplo |
|---|---|---|
| Contexto gerado | `shotgun-prompt-YYYYMMDD-HHMMSS.md` | `shotgun-prompt-20260520-143022.md` |
| Resposta LLM | `llm-response-YYYYMMDD-HHMMSS.md` | `llm-response-20260520-143022.md` |
| Chunk diff | `{input-name}-chunk-NN.diff` | `changes-chunk-01.diff` |

---

## 7. Formato de Progresso JSON

Cada evento de progresso é serializado como JSON:

```json
{
    "timestamp": "2026-05-20T14:30:22Z",
    "stage": "scanning",
    "message": "Processing files",
    "current": 50,
    "total": 100,
    "percent": 50.0
}
```

Campos `current`, `total`, `percent` são omitidos quando zero/nil (omitempty).

---

## 8. Tabela de Mapeamento Config → ScanConfig

| Chave Viper | Campo `scanner.ScanConfig` | Conversão |
|---|---|---|
| `scanner.max-files` | `MaxFiles` | Direta (`int64`) |
| `scanner.max-file-size` | `MaxFileSize` | `utils.ParseSizeWithDefault()` |
| `scanner.max-memory` | `MaxMemory` | `utils.ParseSizeWithDefault()` |
| `scanner.skip-binary` | `SkipBinary` | Direta (`bool`) |
| `scanner.include-hidden` | `IncludeHidden` | Direta (`bool`) |
| `scanner.include-ignored` | `IncludeIgnored` | Direta (`bool`) |
| `scanner.workers` | `Workers` | Direta (`int`) |
| `scanner.respect-gitignore` | `RespectGitignore` | Direta (`bool`) |
| `scanner.respect-shotgunignore` | `RespectShotgunignore` | Direta (`bool`) |
| `--include` flag | `IncludePatterns` | Direta (`[]string`) |
| `--exclude` flag | `IgnorePatterns` | Direta (`[]string`) |
| `--workers` flag | `Workers` | Override (se > 0) |
| `--include-hidden` flag | `IncludeHidden` | Override (se true) |
| `--include-ignored` flag | `IncludeIgnored` | Override (se true) |
