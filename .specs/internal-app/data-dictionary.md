# Dicionário de Dados — Módulo `internal/app`

**Nível de detalhe:** detalhado
**Pacote:** `github.com/quantmind-br/shotgun-cli/internal/app`
**Total de tipos documentados:** 7

---

## 1. Interfaces

### 1.1 `ContextService`

**Localização:** `context.go:10-14`
**Exportada:** ✅
**Pacotes que a implementam:** `DefaultContextService` (mesmo pacote)

| Método | Recebedor | Parâmetros | Retorno | Descrição |
|--------|-----------|------------|---------|-----------|
| `Generate` | `*DefaultContextService` | `ctx context.Context`, `cfg GenerateConfig` | `(*GenerateResult, error)` | Gera contexto de códigobase de forma síncrona |
| `GenerateWithProgress` | `*DefaultContextService` | `ctx context.Context`, `cfg GenerateConfig`, `progress ProgressCallback` | `(*GenerateResult, error)` | Gera contexto com callbacks de progresso |
| `SendToLLM` | `*DefaultContextService` | `ctx context.Context`, `content string`, `provider llm.Provider` | `(*llm.Result, error)` | Envia conteúdo a um provider LLM existente |
| `SendToLLMWithProgress` | `*DefaultContextService` | `ctx context.Context`, `content string`, `cfg LLMSendConfig`, `progress LLMProgressCallback` | `(*llm.Result, error)` | Envia a LLM com criação de provider + progresso |

---

## 2. Structs (Dados de Transferência)

### 2.1 `GenerateConfig`

**Localização:** `context.go:55-65`
**Exportada:** ✅
**Finalidade:** Configuração da operação de geração de contexto de códigobase

| Campo | Tipo | Exportado | Valor Padrão | Obrigatório | Descrição |
|-------|------|-----------|-------------|-------------|-----------|
| `RootPath` | `string` | ✅ | `""` | ✅ | Caminho raiz para escaneamento. Normalizado para absoluto por `Validate()` |
| `ScanConfig` | `*scanner.ScanConfig` | ✅ | `nil` → `scanner.DefaultScanConfig()` | ❌ | Configurações do scanner (tamanho, workers, ignorar) |
| `Selections` | `map[string]bool` | ✅ | `nil` → `scanner.NewSelectAll(tree)` | ❌ | Map de caminhos para incluir/excluir. nil = selecionar tudo |
| `Template` | `string` | ✅ | `""` | ❌ | Nome do template para renderização |
| `TemplateVars` | `map[string]string` | ✅ | `nil` | ❌ | Variáveis do template (ex: `TASK`, `RULES`) |
| `MaxSize` | `int64` | ✅ | `0` (sem limite) | ❌ | Tamanho máximo do conteúdo gerado (bytes) |
| `EnforceLimit` | `bool` | ✅ | `false` | ❌ | Se `true`, erro se conteúdo exceder `MaxSize` |
| `OutputPath` | `string` | ✅ | `""` → gerado com timestamp | ❌ | Caminho do arquivo de saída |
| `CopyToClipboard` | `bool` | ✅ | `false` | ❌ | Copiar conteúdo gerado para clipboard |
| `IncludeTree` | `bool` | ✅ | `false` | ❌ | Incluir árvore de diretórios no output |
| `IncludeSummary` | `bool` | ✅ | `false` | ❌ | Incluir summaries de arquivos |
| `SkipBinary` | `bool` | ✅ | `false` | ❌ | Pular arquivos binários |

**Métodos:**
| Método | Recebedor | Descrição |
|--------|-----------|-----------|
| `Validate()` | `*GenerateConfig` | Valida rootPath (não vazio, existente, é diretório). Normaliza para absoluto. |
| `GenerateOutputPath()` | `*GenerateConfig` | Retorna outputPath configurado ou gera `shotgun-prompt-YYYYMMDD-HHMMSS.md` |

---

### 2.2 `GenerateResult`

**Localização:** `context.go:67-73`
**Exportada:** ✅
**Finalidade:** Resultado de uma operação de geração de contexto

| Campo | Tipo | Exportado | Descrição |
|-------|------|-----------|-----------|
| `Content` | `string` | ✅ | Conteúdo gerado (texto completo) |
| `OutputPath` | `string` | ✅ | Caminho do arquivo onde o conteúdo foi salvo |
| `FileCount` | `int` | ✅ | Número de arquivos incluídos na geração |
| `ContentSize` | `int64` | ✅ | Tamanho do conteúdo em bytes |
| `TokenEstimate` | `int64` | ✅ | Estimativa de tokens (via `tokens.EstimateFromBytes`) |
| `CopiedToClipboard` | `bool` | ✅ | Se o conteúdo foi copiado com sucesso para o clipboard |

---

### 2.3 `LLMSendConfig`

**Localização:** `context.go:34-41`
**Exportada:** ✅
**Finalidade:** Configuração para envio de conteúdo a um provider LLM

| Campo | Tipo | Exportado | Obrigatório | Descrição |
|-------|------|-----------|-------------|-----------|
| `Provider` | `llm.ProviderType` | ✅ | ✅ | Tipo de provider (openai, anthropic, gemini) |
| `APIKey` | `string` | ✅ | ✅* | Chave de API do provider |
| `BaseURL` | `string` | ✅ | ❌ | URL base da API (sobrescreve default do provider) |
| `Model` | `string` | ✅ | ✅* | Modelo LLM a usar |
| `Timeout` | `int` | ✅ | ❌ | Timeout em segundos (default: 300 do provider) |
| `SaveResponse` | `bool` | ✅ | ❌ | Salvar resposta em arquivo |
| `OutputPath` | `string` | ✅ | ❌ | Caminho do arquivo de resposta (só usado se SaveResponse=true) |

*Depende da validação do provider específico.

**Mapeamento para `llm.Config`:**
```go
llm.Config{
    Provider: cfg.Provider,
    APIKey:   cfg.APIKey,
    BaseURL:  cfg.BaseURL,
    Model:    cfg.Model,
    Timeout:  cfg.Timeout,
}
```

---

### 2.4 `CLIConfig`

**Localização:** `config.go:4-26`
**Exportada:** ✅
**Finalidade:** Configuração global de CLI — mapeamento direto de chaves Viper

| Campo | Tipo | Exportado | Descrição |
|-------|------|-----------|-----------|
| `RootPath` | `string` | ✅ | Caminho raiz do projeto |
| `Include` | `[]string` | ✅ | Padrões de arquivos para incluir (glob) |
| `Exclude` | `[]string` | ✅ | Padrões de arquivos para excluir (glob) |
| `Output` | `string` | ✅ | Caminho de arquivo de saída padrão |
| `MaxSize` | `int64` | ✅ | Tamanho máximo de output (bytes) |
| `EnforceLimit` | `bool` | ✅ | Aplicar limite de tamanho |
| `SendGemini` | `bool` | ✅ | 🟡 **INFERIDO:** Flag específica para Gemini como provider |
| `GeminiModel` | `string` | ✅ | 🟡 **INFERIDO:** Modelo padrão do Gemini |
| `GeminiOutput` | `string` | ✅ | 🟡 **INFERIDO:** Caminho de output para resposta Gemini |
| `GeminiTimeout` | `int` | ✅ | 🟡 **INFERIDO:** Timeout para chamadas Gemini |
| `Template` | `string` | ✅ | Nome do template de geração |
| `Task` | `string` | ✅ | Descrição da tarefa para o template |
| `Rules` | `string` | ✅ | Regras para o template |
| `CustomVars` | `map[string]string` | ✅ | Variáveis customizadas do template |
| `Workers` | `int` | ✅ | Número de workers para scanner |
| `IncludeHidden` | `bool` | ✅ | Incluir arquivos ocultos |
| `IncludeIgnored` | `bool` | ✅ | Incluir arquivos ignorados |
| `ProgressMode` | `ProgressMode` | ✅ | Modo de saída de progresso |

**Observação:** `CLIConfig` tem 15 campos, dos quais 4 são específicos de Gemini (`SendGemini`, `GeminiModel`, `GeminiOutput`, `GeminiTimeout`). Isso representa **acoplamento entre a configuração CLI e um provider específico**.

---

### 2.5 `ProgressOutput`

**Localização:** `config.go:37-44`
**Exportada:** ✅
**Finalidade:** Evento estruturado JSON para saída de progresso via CLI

| Campo | Tipo | Exportado | Tag JSON | Descrição |
|-------|------|-----------|----------|-----------|
| `Timestamp` | `string` | ✅ | `timestamp` | Carimbo de data/hora do evento |
| `Stage` | `string` | ✅ | `stage` | Nome da etapa ("scanning", "generating", etc.) |
| `Message` | `string` | ✅ | `message` | Mensagem descritiva |
| `Current` | `int64` | ✅ | `current` | Itens processados atualmente |
| `Total` | `int64` | ✅ | `total` | Total de itens (omitempty) |
| `Percent` | `float64` | ✅ | `percent` | Porcentagem de conclusão (omitempty) |

---

## 3. Tipos de Função (Callbacks)

### 3.1 `ProgressCallback`

**Localização:** `context.go:80`
**Exportada:** ✅
**Assinatura:** `func(stage string, message string, current, total int64)`
**Uso:** Reportar progresso durante scanning e geração de contexto
**Recebedores:** `DefaultContextService.GenerateWithProgress`

---

### 3.2 `LLMProgressCallback`

**Localização:** `context.go:32`
**Exportada:** ✅
**Assinatura:** `func(stage string)`
**Uso:** Reportar progresso durante envio a LLM
**Recebedores:** `DefaultContextService.SendToLLMWithProgress`

---

## 4. Enums / Constantes

### 4.1 `ProgressMode`

**Localização:** `config.go:18`
**Exportada:** ✅
**Tipo base:** `string`

| Constante | Valor | Descrição |
|-----------|-------|-----------|
| `ProgressNone` | `"none"` | Desativa relatórios de progresso |
| `ProgressHuman` | `"human"` | Modo padrão — saída legível por humanos |
| `ProgressJSON` | `"json"` | Saída estruturada JSON (máquina-consumível) |

**Uso:** Campo `CLIConfig.ProgressMode` e possivelmente configurado via flag CLI.

---

## 5. Tipos Não-Exportados (Internos)

### 5.1 `DefaultContextService`

**Localização:** `service.go:12-16`
**Exportada:** ❌ (não-exportada, mas acessível via interface `ContextService`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `scanner` | `scanner.Scanner` | Implementação do scanner de filesystem |
| `generator` | `contextgen.ContextGenerator` | Gerador de contexto (template rendering) |
| `registry` | `*llm.Registry` | Registry de providers LLM |

**Métodos não-exportados:**
| Método | Descrição |
|--------|-----------|
| `Scanner()` | Retorna o scanner interno (utilitário para testes) |

**Functional Options (não-exportadas):**
| Função | Parâmetro | Descrição |
|--------|-----------|-----------|
| `ServiceOption` | `func(*DefaultContextService)` | Tipo-função para opções |
| `WithRegistry` | `*llm.Registry` | Substitui o registry padrão |
| `WithScanner` | `scanner.Scanner` | Substitui o scanner padrão |
| `WithGenerator` | `contextgen.ContextGenerator` | Substitui o generator padrão |

---

### 5.2 `DefaultProviderRegistry`

**Localização:** `providers.go:7`
**Exportada:** ❌ (variável, mas não-exportada pelo nome — mas acessível via `DefaultProviderRegistry` que é exportada por convenção de nome em Go)

**Tipo:** `*llm.Registry`
**Inicialização:** Via `init()` — registra OpenAI, Anthropic, Gemini

---

## 6. Fluxo de Mapeamento entre Configs

### 6.1 CLIConfig → GenerateConfig

```
CLIConfig.RootPath     → GenerateConfig.RootPath
CLIConfig.Output       → GenerateConfig.OutputPath
CLIConfig.MaxSize      → GenerateConfig.MaxSize
CLIConfig.EnforceLimit → GenerateConfig.EnforceLimit
CLIConfig.Template     → GenerateConfig.Template
CLIConfig.Task         → GenerateConfig.TemplateVars["TASK"]
CLIConfig.Rules        → GenerateConfig.TemplateVars["RULES"]
CLIConfig.CustomVars   → GenerateConfig.TemplateVars (merge)
CLIConfig.Include      → scanner.ScanConfig.IncludePatterns
CLIConfig.Exclude      → scanner.ScanConfig.IgnorePatterns
CLIConfig.Workers      → scanner.ScanConfig.Workers
CLIConfig.IncludeHidden→ scanner.ScanConfig.IncludeHidden
CLIConfig.IncludeIgnored→ scanner.ScanConfig.IncludeIgnored
CLIConfig.CopyToClipboard → GenerateConfig.CopyToClipboard
CLIConfig.IncludeTree  → GenerateConfig.IncludeTree
CLIConfig.IncludeSummary → GenerateConfig.IncludeSummary
CLIConfig.SkipBinary   → GenerateConfig.SkipBinary
```

**🟡 INFERIDO:** O mapeamento exato é feito no package `internal/cmd` (fora do escopo). Os campos acima são os mais óbvios baseados na semântica.

### 6.2 CLIConfig → LLMSendConfig

```
CLIConfig.SendGemini   → LLMSendConfig.Provider (se true, llm.ProviderGemini)
CLIConfig.GeminiModel  → LLMSendConfig.Model
CLIConfig.GeminiOutput → LLMSendConfig.OutputPath
CLIConfig.GeminiTimeout → LLMSendConfig.Timeout
```

**🟡 INFERIDO:** Apenas Gemini tem mapeamento direto na CLIConfig. Outros providers (OpenAI, Anthropic) provavelmente usam flags genéricas como `--provider`, `--model`, `--timeout`.

### 6.3 LLMSendConfig → llm.Config

```go
llm.Config{
    Provider: cfg.Provider,
    APIKey:   cfg.APIKey,
    BaseURL:  cfg.BaseURL,
    Model:    cfg.Model,
    Timeout:  cfg.Timeout,
}
```
Após construção: `llmCfg.WithDefaults()` aplica defaults do provider.

---

## 7. Relação entre Tipos do Package `app` e Tipos Internos

| Tipo em `app` | Tipo em `core/scanner` | Relação |
|---------------|------------------------|---------|
| `GenerateConfig.ScanConfig` | `*scanner.ScanConfig` | Composição direta |
| `GenerateConfig.Selections` | `map[string]bool` | Consumido por `scanner.NewSelectAll(tree)` |
| `GenerateResult.FileCount` | `scanner.FileNode.CountFiles()` | Extraído do tree |
| `ProgressCallback` | `scanner.Progress` | Canal `chan scanner.Progress` usado internamente |

| Tipo em `app` | Tipo em `core/contextgen` | Relação |
|---------------|---------------------------|---------|
| `GenerateConfig.Template` | `contextgen.GenerateConfig.Template` | Composição direta |
| `GenerateConfig.MaxSize` | `contextgen.GenerateConfig.MaxTotalSize` | Mapeado em `GenerateWithProgress` |
| `ProgressCallback` | `contextgen.GenProgress` | Adaptado via closure |

| Tipo em `app` | Tipo em `core/llm` | Relação |
|---------------|---------------------|---------|
| `LLMSendConfig` | `llm.Config` | Mapeado em `SendToLLMWithProgress` |
| `LLMProgressCallback` | `func(stage string)` | Direto no `SendWithProgress` |
| `SendToLLM` | `llm.Provider` | Recebe provider já instanciado |
| `GenerateResult.TokenEstimate` | `tokens.EstimateFromBytes()` | Chamado com `contentSize` |

---

## 8. Fluxo de Dados Completo

### 8.1 Geração de Contexto

```
CLIConfig (Viper flags)
    ↓ [mapeamento em cmd]
GenerateConfig
    ↓ Validate()
GenerateConfig (rootPath absoluto)
    ↓ GenerateWithProgress()
    ├── scanner.Scan(rootPath, ScanConfig)
    │   → scanner.FileNode (tree)
    ├── scanner.NewSelectAll(tree) → selections
    ├── generator.Generate(tree, selections, config)
    │   → string (content)
    ├── os.WriteFile(outputPath, content)
    └── clipboard.Copy(content)
    → GenerateResult {Content, FileCount, ContentSize, TokenEstimate, ...}
```

### 8.2 Envio a LLM

```
CLIConfig → LLMSendConfig (mapeamento em cmd)
    ↓ SendToLLMWithProgress()
    ├── llm.Config {Provider, APIKey, BaseURL, Model, Timeout}
    ├── llmCfg.WithDefaults()
    ├── registry.Create(llmCfg)
    │   → llm.Provider (ex: openai.Client)
    ├── provider.ValidateConfig()
    ├── provider.Send(ctx, content)
    │   → *llm.Result {Response, Model, Duration, ...}
    └── os.WriteFile(outputPath, result.Response)
    → *llm.Result
```
