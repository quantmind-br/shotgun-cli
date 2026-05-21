# Inventário do Projeto — shotgun-cli

> Gerado em 2026-05-20 por reversa-scout
> Nível de detalhamento: **detalhado**

---

## 1. Visão Geral

**shotgun-cli** é uma ferramenta CLI em Go para gerar contextos otimizados de codebases para consumo por LLMs. Suporta modo interativo (TUI Bubble Tea) e modo headless via linha de comando.

- **Nome**: `shotgun-cli`
- **Módulo Go**: `github.com/quantmind-br/shotgun-cli`
- **Versão Go**: 1.24.0
- **Binário**: `shotgun-cli`
- **Licença**: não declarada explicitamente

### Principais funcionalidades
- Geração de contexto de codebase (árvore de arquivos + conteúdo) dentro de limites de tamanho/token
- Suporte a templates de prompt personalizáveis
- Envio direto para LLMs (OpenAI, Anthropic, Gemini)
- TUI wizard de 5 etapas (Bubble Tea MVU)
- Wizard de configuração interativo
- Scanner paralelo de filesystem com `.gitignore` e `.shotgunignore`
- Estimativa de tokens heurística
- Copia automática para clipboard

---

## 2. Estrutura de Diretórios

```
shotgun-cli/
├── main.go                          # Ponto de entrada
├── cmd/                             # Camada de apresentação (Cobra CLI)
│   ├── root.go                      # Comando raiz, flags globais, wizard TUI
│   ├── config.go                    # Subcomando `config`
│   ├── config_llm.go                # Helpers de config LLM (BuildLLMConfig, etc.)
│   ├── context.go                   # Subcomando `context generate`
│   ├── diff.go                      # Subcomando `context diff` (se existir)
│   ├── llm.go                       # Subcomando `llm` (status, doctor, list)
│   ├── providers.go                 # Registro global de providers + CreateLLMProvider
│   ├── send.go                      # Subcomando `context send`
│   ├── template.go                  # Subcomando `template` (list, render, import, export)
│   └── *_test.go                    # Testes
├── internal/
│   ├── app/                         # Camada de serviço
│   │   ├── service.go               # DefaultContextService (orquestração principal)
│   │   ├── context.go               # Interface ContextService, GenerateConfig, GenerateResult
│   │   ├── config.go                # CLIConfig, ProgressMode, ProgressOutput
│   │   ├── providers.go             # DefaultProviderRegistry (registro app-scoped)
│   │   └── *_test.go
│   ├── config/                      # Configuração centralizada
│   │   ├── keys.go                  # Chaves de config (KeyScanner*, KeyLLM*, etc.)
│   │   ├── metadata.go              # ConfigMetadata, ConfigCategory
│   │   ├── validator.go             # Validação e conversão de valores
│   │   └── *_test.go
│   ├── core/                        # Lógica de domínio (zero deps externas)
│   │   ├── scanner/                 # Scanner de filesystem
│   │   │   ├── scanner.go           # Interface Scanner, FileNode, ScanConfig, Progress
│   │   │   ├── filesystem.go        # Implementação FileSystemScanner
│   │   │   ├── helpers.go           # Utilitários do scanner
│   │   │   └── *_test.go
│   │   ├── contextgen/              # Geração de contexto
│   │   │   ├── generator.go         # DefaultContextGenerator, ContextData
│   │   │   ├── template.go          # TemplateRenderer
│   │   │   ├── tree.go              # TreeRenderer
│   │   │   ├── content.go           # FileContent, collectFileContents
│   │   │   └── *_test.go
│   │   ├── diff/                    # Utilitários de diff
│   │   │   ├── split.go             # IntelligentSplit, Chunk
│   │   │   └── *_test.go
│   │   ├── ignore/                  # Motor de ignore em camadas
│   │   │   ├── engine.go            # LayeredIgnoreEngine
│   │   │   └── *_test.go
│   │   ├── llm/                     # Interface LLM unificada
│   │   │   ├── provider.go          # Interface Provider, ProviderType, Result, Usage
│   │   │   ├── registry.go          # Provider Registry (factory pattern)
│   │   │   ├── config.go            # Config, DefaultConfigs
│   │   │   └── *_test.go
│   │   ├── template/                # Gerenciamento de templates
│   │   │   ├── loader.go            # TemplateSource (embedded, filesystem)
│   │   │   ├── manager.go           # TemplateManager, Manager
│   │   │   ├── renderer.go          # TemplateRenderer
│   │   │   ├── template.go          # Template struct, parse, validate
│   │   │   └── *_test.go
│   │   ├── tokens/                  # Estimativa de tokens
│   │   │   ├── estimator.go         # Estimate, FormatTokens, ContextFit
│   │   │   └── *_test.go
│   │   └── .keep
│   ├── platform/                    # Adaptadores de infraestrutura
│   │   ├── llmbase/                 # Cliente base HTTP para LLMs
│   │   │   ├── base_client.go       # BaseClient (comum a todos os providers)
│   │   │   ├── sender.go            # Interface Sender
│   │   │   └── *_test.go
│   │   ├── http/                    # Cliente HTTP JSON compartilhado
│   │   │   ├── client.go            # JSONClient
│   │   │   └── *_test.go
│   │   ├── openai/                  # Provider OpenAI
│   │   │   ├── client.go            # Client (implementa llm.Provider)
│   │   │   ├── types.go             # ChatCompletionRequest, ChatCompletionResponse, etc.
│   │   │   ├── models.go            # Structs auxiliares
│   │   │   └── *_test.go
│   │   ├── anthropic/               # Provider Anthropic
│   │   │   ├── client.go            # Client (implementa llm.Provider)
│   │   │   ├── types.go             # MessagesRequest, MessagesResponse, etc.
│   │   │   ├── models.go            # Structs auxiliares
│   │   │   └── *_test.go
│   │   ├── geminiapi/               # Provider Google Gemini
│   │   │   ├── client.go            # Client (implementa llm.Provider)
│   │   │   ├── types.go             # GenerateRequest, GenerateResponse, etc.
│   │   │   ├── models.go            # Structs auxiliares
│   │   │   └── *_test.go
│   │   ├── clipboard/               # Clipboard (cross-platform copy)
│   │   │   ├── clipboard.go         # Copy()
│   │   │   └── *_test.go
│   │   └── .keep
│   ├── ui/                          # Interface de usuário (TUI Bubble Tea)
│   │   ├── wizard.go                # WizardModel (wizard principal 5 etapas)
│   │   ├── config_wizard.go         # ConfigWizardModel (wizard de config)
│   │   ├── scan_coordinator.go      # ScanCoordinator (async scan)
│   │   ├── generate_coordinator.go  # GenerateCoordinator (async generation)
│   │   ├── components/              # Componentes Bubble Tea reutilizáveis
│   │   │   ├── tree.go              # Tree component (file tree)
│   │   │   ├── progress.go          # Progress bar / spinner
│   │   │   ├── config_field.go      # Config field editor
│   │   │   ├── config_select.go     # Config select dropdown
│   │   │   ├── config_toggle.go     # Config toggle (bool)
│   │   │   └── *_test.go
│   │   ├── screens/                 # Telas do wizard
│   │   │   ├── file_selection.go    # FileSelectionModel (step 1)
│   │   │   ├── template_selection.go # TemplateSelectionModel (step 2)
│   │   │   ├── task_input.go        # TaskInputModel (step 3)
│   │   │   ├── rules_input.go       # RulesInputModel (step 4)
│   │   │   ├── review.go            # ReviewModel (step 5)
│   │   │   ├── config_category.go   # ConfigCategoryModel
│   │   │   ├── header_test.go       # Testes de header
│   │   │   └── input_test.go        # Testes de input
│   │   ├── styles/                  # Estilos Lipgloss
│   │   │   ├── theme.go             # Estilos, cores, renders
│   │   │   └── *_test.go
│   │   └── .keep
│   ├── utils/                       # Utilitários
│   │   ├── conversion.go            # ParseSize, FormatBytes, etc.
│   │   └── *_test.go
│   └── assets/                      # Assets embutidos
│       ├── embed.go                 # //go:embed templates/
│       ├── embed_test.go
│       └── templates/               # Templates de prompt embutidos
│           ├── prompt_analyzeBug.md
│           ├── prompt_makeDiffGitFormat.md
│           ├── prompt_makePlan.md
│           └── prompt_projectManager.md
├── custom_templates/                # Templates de exemplo/placeholder
│   ├── chat.md
│   ├── ideation_code_improvements.md
│   ├── ideation_code_quality.md
│   ├── ideation_documentation.md
│   ├── ideation_features.md
│   ├── ideation_performance.md
│   ├── ideation_security.md
│   └── ideation_ui_ux.md
├── test/                            # Testes e fixtures
│   ├── e2e/                         # Testes end-to-end
│   │   ├── cli_test.go
│   │   ├── context_integration_test.go
│   │   └── filestructure_test.go
│   └── fixtures/
│       └── sample-project/          # Projeto de teste simulado (~50 arquivos)
├── .shotgunignore                   # Padrões de ignore padrão
├── go.mod                           # Dependências Go
├── go.sum
├── Makefile                         # Build, test, lint, release
├── Dockerfile                       # Containerização
├── .goreleaser.yaml / goreleaser.yml
└── .golangci.yml / golangci-local.yml
```

---

## 3. Inventário Detalhado por Pacote

### 3.1 `main` (root)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `main.go` | ~20 | Entry point: inicializa zerolog e executa `cmd.Execute()` |

### 3.2 `cmd` (camada de apresentação)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `root.go` | ~250 | Comando raiz `shotgun-cli`. Configuração Viper, flags, wizard TUI |
| `config.go` | ~250 | Subcomando `config` (show, set, TUI wizard) |
| `config_llm.go` | ~50 | Helpers: `BuildLLMConfig()`, `BuildLLMConfigWithOverrides()` |
| `context.go` | ~300 | Subcomando `context generate` (headless mode) |
| `diff.go` | ~50 | Subcomando `context diff` (processamento de diffs) |
| `llm.go` | ~150 | Subcomando `llm` (status, doctor, list) |
| `providers.go` | ~50 | Registro global de providers + `CreateLLMProvider()` |
| `send.go` | ~150 | Subcomando `context send` (enviar arquivo/stdin para LLM) |
| `template.go` | ~250 | Subcomando `template` (list, render, import, export) |
| `*_test.go` | ~variável | Testes unitários para cada comando |

### 3.3 `internal/app` (camada de serviço)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `service.go` | ~200 | `DefaultContextService` — orquestra scanner → generator → output |
| `context.go` | ~100 | Interface `ContextService`, structs `GenerateConfig`, `GenerateResult`, `LLMSendConfig` |
| `config.go` | ~80 | `CLIConfig`, `ProgressMode`, `ProgressOutput` |
| `providers.go` | ~20 | `DefaultProviderRegistry` (registra OpenAI, Anthropic, Gemini) |
| `*_test.go` | ~variável | Testes incluindo `integration_test.go` e `service_llm_test.go` |

### 3.4 `internal/config` (configuração centralizada)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `keys.go` | ~40 | Constantes de chaves: `KeyScanner*`, `KeyLLM*`, `KeyContext*`, etc. |
| `metadata.go` | ~80 | `ConfigMetadata`, `ConfigCategory`, `AllCategories()` |
| `validator.go` | ~200 | Validação de valores: tipos, ranges, formatos, URLs |
| `*_test.go` | ~variável | Testes de validação |

### 3.5 `internal/core/scanner` (scanner de filesystem)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `scanner.go` | ~150 | Interface `Scanner`, `FileNode`, `ScanConfig`, `Progress` |
| `filesystem.go` | ~200 | Implementação `FileSystemScanner` (scan recursivo, paralelo) |
| `helpers.go` | ~80 | Utilitários: verificação de binário, detecção de tipo de arquivo |
| `*_test.go` | ~variável | Testes unitários |

### 3.6 `internal/core/contextgen` (gerador de contexto)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `generator.go` | ~200 | `DefaultContextGenerator`, `ContextData`, geração sequencial |
| `template.go` | ~100 | `TemplateRenderer` — substituição de variáveis `{VAR}` em templates |
| `tree.go` | ~100 | `TreeRenderer` — renderização de árvore ASCII |
| `content.go` | ~80 | `FileContent` struct, `collectFileContents()` |
| `*_test.go` | ~variável | Testes unitários |

### 3.7 `internal/core/diff` (utilitários de diff)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `split.go` | ~150 | `IntelligentSplit`, `Chunk`, divisão de diffs respeitando limites de linha |
| `*_test.go` | ~variável | Testes unitários |

### 3.8 `internal/core/ignore` (motor de ignore em camadas)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `engine.go` | ~300 | `LayeredIgnoreEngine`: built-in → gitignore → custom → explicit |
| `*_test.go` | ~variável | Testes unitários |

### 3.9 `internal/core/llm` (interface LLM unificada)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `provider.go` | ~80 | Interface `Provider` (Send, SendWithProgress, Name, etc.), `ProviderType`, `Result`, `Usage` |
| `registry.go` | ~60 | `Registry` — factory pattern para criação de providers |
| `config.go` | ~80 | `Config`, `DefaultConfigs()`, `MaskAPIKey()` |
| `*_test.go` | ~variável | Testes unitários |

### 3.10 `internal/core/template` (gerenciamento de templates)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `loader.go` | ~100 | `TemplateSource` interface, `EmbeddedSource`, `FilesystemSource` |
| `manager.go` | ~120 | `Manager`, `TemplateManager` interface, carregamento multi-fonte |
| `renderer.go` | ~100 | `Renderer`, renderização de templates com variáveis |
| `template.go` | ~120 | `Template` struct, parse, extração de variáveis, validação |
| `*_test.go` | ~variável | Testes unitários |

### 3.11 `internal/core/tokens` (estimativa de tokens)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `estimator.go` | ~100 | Heurística 1 token ≈ 4 bytes, `Estimate`, `FormatTokens`, `ContextFit` |
| `*_test.go` | ~variável | Testes unitários |

### 3.12 `internal/platform/llmbase` (cliente base HTTP)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `base_client.go` | ~130 | `BaseClient` — lógica comum: construção de request, parsing de response |
| `sender.go` | ~30 | `Sender` interface — estratégia para construir/parsear requests específicos |
| `*_test.go` | ~variável | Testes unitários |

### 3.13 `internal/platform/http` (cliente HTTP JSON)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `client.go` | ~100 | `JSONClient` — POST JSON com timeout, `HTTPError` |
| `*_test.go` | ~variável | Testes unitários |

### 3.14 `internal/platform/openai` (provider OpenAI)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `client.go` | ~80 | `Client` — implementa `llm.Provider` |
| `types.go` | ~60 | `ChatCompletionRequest`, `Message`, `ChatCompletionResponse`, `ErrorResponse` |
| `models.go` | ~40 | Structs auxiliares |
| `*_test.go` | ~variável | Testes unitários |

### 3.15 `internal/platform/anthropic` (provider Anthropic)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `client.go` | ~100 | `Client` — implementa `llm.Provider` |
| `types.go` | ~80 | `MessagesRequest`, `Message`, `MessagesResponse`, `ErrorResponse` |
| `models.go` | ~30 | Structs auxiliares |
| `*_test.go` | ~variável | Testes unitários |

### 3.16 `internal/platform/geminiapi` (provider Gemini)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `client.go` | ~100 | `Client` — implementa `llm.Provider` |
| `types.go` | ~80 | `GenerateRequest`, `Content`, `Part`, `GenerateResponse`, `UsageMetadata` |
| `models.go` | ~40 | Structs auxiliares, testes de modelos |
| `*_test.go` | ~variável | Testes unitários |

### 3.17 `internal/platform/clipboard` (clipboard)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `clipboard.go` | ~50 | `Copy()` — cross-platform copy to system clipboard |
| `*_test.go` | ~variável | Testes unitários |

### 3.18 `internal/ui` (TUI Bubble Tea)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `wizard.go` | ~450 | `WizardModel` — wizard principal de 5 etapas (Bubble Tea MVU) |
| `config_wizard.go` | ~300 | `ConfigWizardModel` — wizard de configuração |
| `scan_coordinator.go` | ~80 | `ScanCoordinator` — coordena scan assíncrono |
| `generate_coordinator.go` | ~80 | `GenerateCoordinator` — coordena geração assíncrona |
| `*_test.go` | ~variável | Testes unitários |

### 3.19 `internal/ui/components` (componentes Bubble Tea)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `tree.go` | ~100 | Tree view de arquivos (navegação, seleção) |
| `progress.go` | ~80 | Progress bar e spinner |
| `config_field.go` | ~60 | Editor de campo de config |
| `config_select.go` | ~60 | Dropdown de seleção de config |
| `config_toggle.go` | ~40 | Toggle boolean (on/off) |
| `*_test.go` | ~variável | Testes unitários |

### 3.20 `internal/ui/screens` (telas do wizard)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `file_selection.go` | ~120 | `FileSelectionModel` — Step 1: seleção de arquivos |
| `template_selection.go` | ~100 | `TemplateSelectionModel` — Step 2: seleção de template |
| `task_input.go` | ~60 | `TaskInputModel` — Step 3: descrição da tarefa |
| `rules_input.go` | ~60 | `RulesInputModel` — Step 4: regras/constraints |
| `review.go` | ~150 | `ReviewModel` — Step 5: revisão, generate, send to LLM |
| `config_category.go` | ~80 | `ConfigCategoryModel` — tela de categoria de config |
| `*_test.go` | ~variável | Testes unitários |

### 3.21 `internal/ui/styles` (estilos Lipgloss)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `theme.go` | ~150 | Cores, fontes, renders de UI (header, footer, error, success, etc.) |
| `*_test.go` | ~variável | Testes unitários |

### 3.22 `internal/utils` (utilitários)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `conversion.go` | ~80 | `ParseSize()`, `FormatBytes()`, `ParseSizeWithDefault()` |
| `*_test.go` | ~variável | Testes unitários |

### 3.23 `internal/assets` (assets embutidos)
| Arquivo | Linhas estimadas | Descrição |
|---------|-----------------|-----------|
| `embed.go` | ~20 | `//go:embed templates/` — embutimento de templates |
| `embed_test.go` | ~20 | Testes de embed |
| `templates/*.md` | ~variável | Templates de prompt embutidos (analyzeBug, makeDiffGitFormat, makePlan, projectManager) |

---

## 4. Contagem Total

| Métrica | Valor |
|---------|-------|
| **Pacotes Go** | 25+ |
| **Arquivos `.go`** | ~80+ |
| **Arquivos de teste** | ~40+ |
| **Templates embutidos** | 4 |
| **Templates custom (exemplo)** | 8 |
| **Providers suportados** | 3 (OpenAI, Anthropic, Gemini) |
| **Subcomandos CLI** | 8+ (root, config, context, llm, template) |
| **Dependências diretas** | 10 |
| **Dependências indiretas** | 35+ |

---

## 5. Arquivos de Configuração e Build

| Arquivo | Descrição |
|---------|-----------|
| `go.mod` | Declaração de módulo e dependências |
| `go.sum` | Soma de verificação das dependências |
| `Makefile` | Regras: build, test, lint, release, docker |
| `Dockerfile` | Multi-stage build (alpine + builder) |
| `.goreleaser.yaml` / `.goreleaser.yml` | Configuração de release com GoReleaser |
| `.golangci.yml` | Configuração do linter golangci-lint |
| `.golangci-local.yml` | Configuração local do linter |
| `.dockerignore` | Arquivos ignorados no build Docker |
| `.gitignore` | Padrões de ignorância do Git |
| `.shotgunignore` | Padrões padrão de ignorância do shotgun-cli |
| `.gitattributes` | Configurações de atributos Git |
