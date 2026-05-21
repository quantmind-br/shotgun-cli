# Análise de Código — Módulo `internal/app`

**Nível de detalhe:** detalhado
**Caminho do pacote:** `github.com/quantmind-br/shotgun-cli/internal/app`
**Número de arquivos de código:** 7 (6 fontes + 1 docs)
**Dependências externas:** 0 (apenas packages internos do projeto)
**Dependências internas:** `core/scanner`, `core/contextgen`, `core/llm`, `core/tokens`, `platform/clipboard`

---

## 1. Visão Geral Arquitetural

O módulo `internal/app` é a **camada de serviço de aplicação** — a orquestradora entre a camada de apresentação (CLI/TUI) e a lógica de domínio (core). Ele **não possui dependências de frameworks**, apenas invoca componentes core e platform.

### Responsabilidades

1. **Geração de contexto:** escanear filesystem → aplicar seleções → renderizar template → salvar arquivo (e opcionalmente copiar para clipboard)
2. **Envio a LLMs:** criar provider via registry → validar → enviar conteúdo → salvar resposta (opcional)
3. **Configuração:** mapear `CLIConfig` → `GenerateConfig`/`LLMSendConfig`

### Regras de Fronteira

- ❌ **Não importa** `internal/ui` (TUI) nem `internal/cmd` (CLI) — evita dependência de ciclo
- ❌ **Não usa Viper diretamente** — recebe config via parâmetros
- ❌ **Não cria instâncias globais** de serviço — usa construtor com opções funcionais

---

## 2. Inventário de Arquivos

### 2.1 Arquivos de Produção

| Arquivo | Linhas | Pacote | Função |
|---------|--------|--------|--------|
| `context.go` | ~95 | `app` | Definição da interface `ContextService`, structs de config (`GenerateConfig`, `LLMSendConfig`), struct de resultado (`GenerateResult`), callbacks de progresso |
| `service.go` | ~200 | `app` | Implementação `DefaultContextService` — métodos `Generate`, `GenerateWithProgress`, `SendToLLM`, `SendToLLMWithProgress` |
| `config.go` | ~60 | `app` | `CLIConfig`, `ProgressMode` (constantes), `ProgressOutput` (struct JSON) |
| `providers.go` | ~20 | `app` | `DefaultProviderRegistry` (singleton), `init()` com registro de 3 providers |

### 2.2 Arquivos de Teste

| Arquivo | Linhas | Função |
|---------|--------|--------|
| `service_test.go` | ~170 | Testes unitários de `Generate`, `GenerateWithProgress`, `SendToLLM` |
| `service_llm_test.go` | ~280 | Testes unitários de `SendToLLMWithProgress` — 17 casos de teste |
| `integration_test.go` | ~200 | Testes de fluxo end-to-end de LLM com mocks |
| `config_test.go` | ~40 | Testes de campos de `CLIConfig` e constantes de `ProgressMode` |
| `context_test.go` | ~65 | Testes de validação de `GenerateConfig` e `GenerateOutputPath` |

---

## 3. Análise Detalhada por Arquivo

### 3.1 `context.go` — Definições de Contrato

**Funções/Tipos exportados:** 6

| Tipo | Exportado | Descrição |
|------|-----------|-----------|
| `ContextService` (interface) | ✅ | Contrato público com 4 métodos |
| `LLMSendConfig` (struct) | ✅ | Config para envio a LLM (provider, apiKey, model, saveResponse, outputPath) |
| `LLMProgressCallback` (func) | ✅ | `func(stage string)` |
| `GenerateConfig` (struct) | ✅ | Config para geração de contexto (rootPath, selections, template, maxSizes, output) |
| `GenerateResult` (struct) | ✅ | Resultado da geração (content, fileCount, tokenEstimate, clipboard) |
| `ProgressCallback` (func) | ✅ | `func(stage string, message string, current, total int64)` |

**Métodos:**

| Recebedor | Método | Exportado | Descrição |
|-----------|--------|-----------|-----------|
| `*GenerateConfig` | `Validate()` | ✅ | Valida rootPath (existente, é diretório, resolve para absoluto) |
| `*GenerateConfig` | `GenerateOutputPath()` | ✅ | Retorna outputPath configurado ou gera nome com timestamp |

**Métodos não-exportados:** nenhum (todos via interface)

**Dependências de import:**
- `context` (std)
- `fmt` (std) — erros
- `os` (std) — `os.Stat`
- `path/filepath` (std) — `filepath.Abs`
- `time` (std) — timestamp

**Observações:**
- O `GenerateConfig` modifica o próprio `RootPath` durante validação (normaliza para absoluto). Isso é **side-effect no parâmetro** — potencial fonte de bugs se o mesmo struct for reutilizado.
- A interface `ContextService` é minimalista (4 métodos), mas poderosa — abrange toda a funcionalidade do módulo.

---

### 3.2 `service.go` — Implementação Principal

**Tipo:** `DefaultContextService` (struct, não-exportada)
**Métodos (via interface `ContextService` + 1 extra):**

| Método | Exportado | Linhas | Descrição |
|--------|-----------|--------|-----------|
| `Generate` | ✅ | ~25 | Delega para `GenerateWithProgress` com `nil` callback |
| `GenerateWithProgress` | ✅ | ~85 | **Workflow principal:** valida → escaneia → seleciona → gera → limita → salva → clipboard |
| `SendToLLM` | ✅ | ~20 | Valida provider → delega para `provider.Send` |
| `SendToLLMWithProgress` | ✅ | ~55 | Cria provider via registry → valida → envia com progresso → salva resposta |
| `Scanner()` | ❌ | ~4 | Retorna scanner interno (para testes) |

**Functional Options (3):**

| Opção | Linha | Efeito |
|-------|-------|--------|
| `WithRegistry(*llm.Registry)` | ~30 | Substitui registry padrão |
| `WithScanner(scanner.Scanner)` | ~35 | Substitui scanner padrão |
| `WithGenerator(contextgen.ContextGenerator)` | ~40 | Substitui generator padrão |

**Workflow `GenerateWithProgress` (detalhado):**

```
1. cfg.Validate() → erro se rootPath inválido
2. scanConfig = cfg.ScanConfig ?? scanner.DefaultScanConfig()
3. [progress] → "scanning" (0, 0)
4. Se progress != nil:
   - Cria progressCh (buffered, cap=100)
   - Goroutine: range progressCh → report()
   - s.scanner.ScanWithProgress(rootPath, scanConfig, progressCh)
   - Close(progressCh), <-done
5. Senão:
   - s.scanner.Scan(rootPath, scanConfig)
6. tree, err → erro se falhar scan
7. selections = cfg.Selections ?? scanner.NewSelectAll(tree)
8. [progress] → "generating" (0, 0)
9. Se progress != nil:
   - s.generator.GenerateWithProgressEx(tree, selections, genConfig, adaptedCallback)
10. Senão:
    - s.generator.Generate(tree, selections, genConfig)
11. content, err → erro se falhar geração
12. Se cfg.EnforceLimit && contentSize > cfg.MaxSize → erro
13. [progress] → "saving" (0, 0)
14. os.WriteFile(outputPath, content, 0600)
15. Se cfg.CopyToClipboard → clipboard.Copy(content)
16. [progress] → "complete" (1, 1)
17. Retorna GenerateResult
```

**Workflow `SendToLLMWithProgress` (detalhado):**

```
1. Constrói llm.Config a partir de LLMSendConfig
2. llmCfg.WithDefaults() → aplica defaults por provider
3. provider = s.registry.Create(llmCfg) → erro se provider não registrado
4. provider.IsAvailable() → erro se não disponível
5. provider.ValidateConfig() → erro se config inválida
6. Se progress != nil:
   - provider.SendWithProgress(ctx, content, progress)
7. Senão:
   - provider.Send(ctx, content)
8. result, err → erro se falhar
9. Se cfg.SaveResponse && cfg.OutputPath != "":
   - os.WriteFile(cfg.OutputPath, result.Response, 0600) — erro não bloqueia result
10. Retorna result (pode ter erro de save mas result válido)
```

**Observações:**

- O `report` helper é uma closure interna que verifica `progress != nil` antes de chamar — padrão seguro.
- O canal `progressCh` é `make(chan scanner.Progress, 100)` — buffer de 100 para evitar blocking no scanner.
- O goroutine de consumo de progresso é **drainado via `close(progressCh) + <-done`** — correto para evitar goroutine leaks.
- O `SaveResponse` tem comportamento de **erro não-bloqueante**: mesmo que falhe, o `result` é retornado com o erro anexado.
- A constante de permissão `0600` é consistente entre os dois `os.WriteFile` — apenas o owner pode ler/escrever.

---

### 3.3 `providers.go` — Registro de Providers

**Variável global:** `DefaultProviderRegistry` do tipo `*llm.Registry`
**Função:** `init()` — registro automático de 3 providers

| Provider | Tipo | Factory | Package de Implementação |
|----------|------|---------|-------------------------|
| OpenAI | `llm.ProviderOpenAI` | `openai.NewClient(cfg)` | `internal/platform/openai` |
| Anthropic | `llm.ProviderAnthropic` | `anthropic.NewClient(cfg)` | `internal/platform/anthropic` |
| Gemini | `llm.ProviderGemini` | `geminiapi.NewClient(cfg)` | `internal/platform/geminiapi` |

**Observações:**
- O registro via `init()` garante que o registry está pronto antes de qualquer uso.
- A factory recebe `llm.Config`, não `LLMSendConfig` — há uma conversão implícita em `SendToLLMWithProgress` (linha ~165 de service.go).
- **Não há mecanismo de un-register** — o registry é imutável após init.

---

### 3.4 `config.go` — Configuração CLI

**Structs:**

| Struct | Campos | Descrição |
|--------|--------|-----------|
| `CLIConfig` | 15 campos | Configuração global de CLI — mapeamento direto de Viper |
| `ProgressOutput` | 5 campos | Evento JSON de progresso — usado para saída JSON |

**Tipos:**

| Tipo | Valores | Descrição |
|------|---------|-----------|
| `ProgressMode` | `ProgressNone`, `ProgressHuman`, `ProgressJSON` | Modo de saída de progresso |

**Observações:**
- `CLIConfig` tem campos específicos de Gemini (`SendGemini`, `GeminiModel`, `GeminiOutput`, `GeminiTimeout`) — **acoplamento com provider específico**.
- `CLIConfig` tem campos de template (`Template`, `Task`, `Rules`, `CustomVars`) — usados apenas na geração de contexto.
- A separação entre `CLIConfig` (entrada CLI) e `GenerateConfig`/`LLMSendConfig` (entrada service) sugere que há uma camada de mapeamento no package `cmd`.

---

## 4. Dependências de Pacote

### 4.1 Dependências Diretas

| Dependência | Origem | Uso |
|-------------|--------|-----|
| `context` | stdlib | Contexto de cancelamento |
| `fmt` | stdlib | Formatação de erros |
| `os` | stdlib | `WriteFile`, `Stat` |
| `path/filepath` | stdlib | `Abs` |
| `time` | stdlib | Timestamp de output |
| `core/contextgen` | interno | Interface `ContextGenerator`, struct `GenerateConfig`, `GenProgress` |
| `core/llm` | interno | Interface `Provider`, struct `Config`, `Result`, `Registry`, `ProviderType` |
| `core/scanner` | interno | Interface `Scanner`, struct `FileNode`, `ScanConfig`, `Progress` |
| `core/tokens` | interno | Função `EstimateFromBytes` |
| `platform/clipboard` | interno | Função `Copy` |

### 4.2 Dependências Transitivas (via init)

- `platform/openai` — registrado via `init()` em `providers.go`
- `platform/anthropic` — registrado via `init()` em `providers.go`
- `platform/geminiapi` — registrado via `init()` em `providers.go`

### 4.3 Matriz de Dependências

```
internal/app
├── internal/core/scanner    (interface Scanner, struct FileNode/ScanConfig/Progress)
├── internal/core/contextgen (interface ContextGenerator, struct GenerateConfig/GenProgress)
├── internal/core/llm        (interface Provider, struct Config/Result/Registry)
├── internal/core/tokens     (func EstimateFromBytes)
└── internal/platform/clipboard (func Copy)
```

**Sem dependências de:** `internal/cmd`, `internal/ui`, `internal/config` — regra intencional para manter a camada de serviço pura.

---

## 5. Padrões de Design Identificados

### 5.1 Functional Options Pattern

`DefaultContextService` usa o padrão Functional Options para composição:

```go
svc := app.NewContextService(
    WithScanner(mockScanner),
    WithGenerator(mockGenerator),
    WithRegistry(customRegistry),
)
```

- ✅ Testabilidade máxima (mockagem fácil)
- ✅ Extensibilidade (novas opções não quebram APIs)
- ✅ Defaults sensatos (NewFileSystemScanner, NewDefaultContextGenerator, DefaultProviderRegistry)

### 5.2 Callback de Progresso (Observer Pattern)

Dois tipos de callback:
- `ProgressCallback`: `func(stage, message string, current, total int64)` — para scanning/geração
- `LLMProgressCallback`: `func(stage string)` — para LLM (mais simples)

### 5.3 Provider Registry (Factory Pattern)

`llm.Registry` é um factory thread-safe com mutex (`sync.RWMutex`):
- `Register(type, creator)` — registra factories
- `Create(cfg)` — cria provider a partir de config
- `SupportedProviders()` / `IsRegistered(type)` — consulta

### 5.4 Interface Segregation

`ContextService` é uma interface pequena (4 métodos) que evita o princípio da segregação de interfaces. Métodos separados para geração e envio a LLM refletem domínios distintos.

---

## 6. Cobertura de Testes

### 6.1 Resumo

| Arquivo de Teste | Casos de Teste | Funções Cobertas |
|------------------|----------------|------------------|
| `service_test.go` | ~15 | `NewContextService`, `Generate`, `GenerateWithProgress`, `SendToLLM` |
| `service_llm_test.go` | 17 | `SendToLLMWithProgress` (todos os caminhos) |
| `integration_test.go` | 9 | Fluxos end-to-end de LLM |
| `config_test.go` | 2 | `CLIConfig` campos, `ProgressMode` constantes |
| `context_test.go` | 8 | `GenerateConfig.Validate`, `GenerateOutputPath`, `GenerateResult` |

**Total estimado de testes:** ~51 casos de teste

### 6.2 Mocks Utilizados

| Mock | Implementa | Arquivo |
|------|-----------|---------|
| `mockScanner` | `scanner.Scanner` | `service_test.go` |
| `mockGenerator` | `contextgen.ContextGenerator` | `service_test.go` |
| `mockProvider` | `llm.Provider` | `service_test.go` |
| `mockLLMProvider` | `llm.Provider` | `service_llm_test.go` |
| `integrationMockProvider` | `llm.Provider` | `integration_test.go` |

### 6.3 Cobertura por Caminho

| Camho de Execução | Cobertura |
|-------------------|-----------|
| `Generate` com config inválida | ✅ |
| `Generate` com erro de scan | ✅ |
| `Generate` com erro de geração | ✅ |
| `Generate` com limite excedido | ✅ |
| `Generate` sucesso | ✅ |
| `GenerateWithProgress` com callbacks | ✅ |
| `Generate` com seleções customizadas | ✅ |
| `SendToLLM` provider indisponível | ✅ |
| `SendToLLM` sucesso | ✅ |
| `SendToLLM` erro | ✅ |
| `SendToLLMWithProgress` sucesso | ✅ |
| `SendToLLMWithProgress` com callback | ✅ |
| `SendToLLMWithProgress` nil callback | ✅ |
| `SendToLLMWithProgress` save response | ✅ |
| `SendToLLMWithProgress` sem save | ✅ |
| `SendToLLMWithProgress` provider creation fails | ✅ |
| `SendToLLMWithProgress` unsupported provider | ✅ |
| `SendToLLMWithProgress` not available | ✅ |
| `SendToLLMWithProgress` validate fails | ✅ |
| `SendToLLMWithProgress` send fails | ✅ |
| `SendToLLMWithProgress` save fails (retorna result) | ✅ |
| Config propagation ao provider | ✅ |
| Todos os 3 tipos de provider | ✅ |
| Contexto cancelado | ✅ |

**Caminho não testado explicitamente:**
- 🟡 **Clipboard copy com erro** — não há teste que simule falha no `clipboard.Copy()`
- 🟡 **`GenerateOutputPath` com timestamp** — teste existe mas não verifica formato exato do timestamp (apenas `Contains`)

---

## 7. Questões Técnicas e Debt

### 7.1 Side-Effect em Validate

`GenerateConfig.Validate()` modifica `c.RootPath` para caminho absoluto. Se o mesmo `GenerateConfig` for reutilizado em múltiplas chamadas, o caminho já estará absoluto na segunda chamada. Não é bug atual, mas é **fragilidade**.

### 7.2 CLIConfig Acoplado ao Gemini

Campos `SendGemini`, `GeminiModel`, `GeminiOutput`, `GeminiTimeout` em `CLIConfig` sugerem **hard-coding do Gemini na camada de configuração**, mesmo que a factory de providers seja genérica. Isso limita a extensão para outros providers na configuração CLI.

**🟡 INFERIDO:** Talvez o Gemini seja o provider padrão/primário e por isso tem campos explícitos. Não há evidência de que outros providers tenham campos equivalentes.

### 7.3 MaxMemory Não Implementado no Scanner

O `ScanConfig.MaxMemory` existe mas **não é aplicado** no `filesystem.Scanner`. Documentado como "critical" pelo phase anterior.

### 7.4 Sem Retry Policy

Nenhum mecanismo de retry para chamadas LLM. Se a rede falhar, a chamada é perdida.

### 7.5 TokenEstimate Simplificado

`tokens.EstimateFromBytes()` é chamado com `contentSize` (bytes) para estimar tokens. A conversão é uma aproximação (assumindo ~4 bytes/token para inglês/ASCII). Sem evidência de precisão.

---

## 8. Métricas de Código

| Métrica | Valor |
|---------|-------|
| Linhas de código (excluindo testes) | ~350 |
| Linhas de código (testes) | ~755 |
| Ratio teste:código | ~2.16x |
| Tipos definidos | 7 (1 interface, 6 structs, 1 tipo func) |
| Constantes | 3 (ProgressMode) |
| Variáveis globais | 1 (DefaultProviderRegistry) |
| Funções init | 1 |
| Funções não-método | 0 |
| Métodos total | ~12 (6 via interface + 2 options + 2 struct methods) |
| Importações stdlib | 6 |
| Importações internas | 5 |
| Importações externas de terceiros | 0 |

---

## 9. Resumo Executativo

O módulo `internal/app` é a **camada de orquestração central** do shotgun-cli. Implementa dois fluxos principais:

1. **Geração de contexto:** filesystem scan → content collection → template rendering → output file
2. **Envio a LLM:** provider factory → validation → API call → response save

**Pontos fortes:**
- Arquitetura limpa com interface-segregation
- Functional options para testabilidade
- Progresso assíncrono com goroutines e canais
- Cobertura de testes robusta (~51 casos)

**Pontos de atenção:**
- `CLIConfig` tem campos específicos de Gemini (acoplamento)
- `Validate()` tem side-effect no parâmetro
- Sem retry para LLMs
- `MaxMemory` não implementado
