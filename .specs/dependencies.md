# Mapa de Dependências — shotgun-cli

> Gerado em 2026-05-20 por reversa-scout
> Nível de detalhamento: **detalhado**

---

## 1. Dependências Externas

### 1.1 Dependências Diretas

| Pacote | Versão | Uso |
|--------|--------|-----|
| `github.com/spf13/cobra` | v1.10.2 | Framework CLI (comandos, flags, subcomandos) |
| `github.com/spf13/viper` | v1.21.0 | Gerenciamento de configuração (config files, env vars, defaults) |
| `github.com/charmbracelet/bubbletea` | v1.3.5 | TUI framework (Bubble Tea, MVU pattern) |
| `github.com/charmbracelet/bubbles` | v0.21.0 | Componentes Bubble Tea (viewport, textinput, list, etc.) |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | Estilização terminal (cores, layout, bordas) |
| `github.com/rs/zerolog` | v1.33.0 | Logging estruturado |
| `github.com/sabhiram/go-gitignore` | v0.0.0-20210923224102-525f6e181f06 | Parsing de padrões .gitignore |
| `github.com/adrg/xdg` | v0.5.3 | Diretórios XDG (configuração no sistema) |
| `github.com/atotto/clipboard` | v0.1.4 | Copiar texto para clipboard do sistema |
| `github.com/stretchr/testify` | v1.11.1 | Framework de testes |
| `golang.org/x/text` | v0.32.0 | Utilitários de texto (encoding) |

### 1.2 Dependências Indiretas (principais)

| Pacote | Dependido por | Uso |
|--------|--------------|-----|
| `github.com/fsnotify/fsnotify` | viper | Watch de arquivos de config |
| `github.com/spf13/afero` | viper | Abstração de filesystem |
| `github.com/spf13/pflag` | cobra | Parsing de flags POSIX |
| `github.com/spf13/cast` | viper | Conversão de tipos |
| `github.com/go-viper/mapstructure/v2` | viper | Binding de config |
| `github.com/subosito/gotenv` | viper | Parsing de .env |
| `github.com/pelletier/go-toml/v2` | viper | Suporte a formato TOML |
| `go.yaml.in/yaml/v3` | viper | Suporte a formato YAML |
| `github.com/mattn/go-isatty` | bubbletea | Detecção de TTY |
| `github.com/mattn/go-colorable` | lipgloss | Cores Windows |
| `github.com/lucasb-eyer/go-colorful` | lipgloss | Manipulação de cores |
| `github.com/rivo/uniseg` | lipgloss | Segmentation de unicode |
| `github.com/muesli/termenv` | lipgloss | Detecção de capabilities terminal |
| `github.com/aymanbagabas/go-osc52/v2` | termenv | Protocolo OSC52 (cores terminal) |
| `github.com/charmbracelet/colorprofile` | termenv | Suporte a color profile |
| `github.com/charmbracelet/x/ansi` | lipgloss | Parser ANSI |
| `github.com/charmbracelet/x/cellbuf` | lipgloss | Buffer de células |
| `github.com/charmbracelet/x/term` | lipgloss | Abstração de terminal |
| `github.com/inconshreveable/mousetrap` | cobra | Proteção contra duplo-clique Windows |
| `github.com/mattn/go-runewidth` | lipgloss | Largura de caracteres CJK |
| `github.com/sagikazarmark/locafero` | viper | Fontes de config adicionais |
| `github.com/sourcegraph/conc` | locafero | Utilities concorrentes |
| `github.com/erikgeiser/coninput` | bubbletea | Input Windows |
| `github.com/xo/terminfo` | bubbletea | Terminal capabilities |
| `golang.org/x/sync` | diversos | Semáforos, errgroup |
| `golang.org/x/sys` | diversos | Abstração syscall |
| `github.com/davecgh/go-spew` | testify | Dump de estruturas |
| `github.com/pmezard/go-difflib` | testify | Diferença de strings |

---

## 2. Dependências Internas — Grafo de Imports

### 2.1 Visão do Grafo

```
main
  └── cmd
        ├── root
        │     ├── internal/config
        │     ├── internal/core/scanner
        │     ├── internal/ui
        │     └── internal/utils
        ├── config
        │     ├── internal/config
        │     ├── internal/ui
        │     └── internal/ui/styles
        ├── config_llm
        │     ├── internal/config
        │     ├── internal/core/llm
        │     ├── internal/platform/openai
        │     ├── internal/platform/anthropic
        │     └── internal/platform/geminiapi
        ├── context
        │     ├── internal/app
        │     ├── internal/config
        │     ├── internal/core/scanner
        │     ├── internal/core/template
        │     ├── internal/core/tokens
        │     └── internal/utils
        ├── diff
        │     └── internal/core/diff
        ├── llm
        │     ├── internal/config
        │     ├── internal/core/llm
        │     └── internal/ui/styles
        ├── providers
        │     ├── internal/core/llm
        │     ├── internal/platform/openai
        │     ├── internal/platform/anthropic
        │     └── internal/platform/geminiapi
        ├── send
        │     ├── internal/config
        │     └── cmd.providers
        ├── template
        │     ├── internal/assets
        │     ├── internal/config
        │     └── internal/core/template
        └── completion
              └── (cobra autocompletion)

internal/app
  ├── service
  │     ├── internal/core/contextgen
  │     ├── internal/core/llm
  │     ├── internal/core/scanner
  │     ├── internal/core/tokens
  │     └── internal/platform/clipboard
  ├── context
  │     ├── internal/core/llm
  │     └── internal/core/scanner
  ├── config
  └── providers
        ├── internal/core/llm
        ├── internal/platform/openai
        ├── internal/platform/anthropic
        └── internal/platform/geminiapi

internal/config
  ├── keys
  ├── metadata
  └── validator
        └── internal/utils

internal/core/scanner
  ├── scanner
  └── internal/core/ignore

internal/core/contextgen
  ├── generator
  │     ├── internal/core/scanner
  │     └── internal/core/template
  ├── template
  │     └── internal/core/template
  ├── tree
  │     └── internal/core/scanner
  └── content
        └── internal/core/scanner

internal/core/diff
  └── split

internal/core/ignore
  └── engine
        └── github.com/sabhiram/go-gitignore

internal/core/llm
  ├── provider
  ├── registry
  └── config

internal/core/template
  ├── loader
  │     └── internal/assets
  ├── manager
  │     ├── internal/assets
  │     └── internal/core/template
  ├── renderer
  │     └── internal/core/template
  └── template

internal/core/tokens
  └── estimator

internal/platform/llmbase
  ├── base_client
  │     └── internal/platform/http
  └── sender

internal/platform/http
  └── client

internal/platform/openai
  ├── client
  │     ├── internal/core/llm
  │     └── internal/platform/llmbase
  ├── types
  └── models

internal/platform/anthropic
  ├── client
  │     ├── internal/core/llm
  │     └── internal/platform/llmbase
  ├── types
  └── models

internal/platform/geminiapi
  ├── client
  │     ├── internal/core/llm
  │     └── internal/platform/llmbase
  ├── types
  └── models

internal/platform/clipboard
  └── clipboard

internal/ui
  ├── wizard
  │     ├── internal/app
  │     ├── internal/core/contextgen
  │     ├── internal/core/llm
  │     ├── internal/core/scanner
  │     ├── internal/core/template
  │     ├── internal/platform/clipboard
  │     ├── internal/ui/components
  │     ├── internal/ui/screens
  │     └── internal/ui/styles
  ├── config_wizard
  │     ├── internal/config
  │     ├── internal/ui/screens
  │     └── internal/ui/styles
  ├── scan_coordinator
  │     └── internal/core/scanner
  └── generate_coordinator
        └── internal/core/contextgen

internal/ui/components
  ├── tree
  ├── progress
  ├── config_field
  ├── config_select
  └── config_toggle

internal/ui/screens
  ├── file_selection
  │     └── internal/core/scanner
  ├── template_selection
  │     └── internal/core/template
  ├── task_input
  ├── rules_input
  ├── review
  │     ├── internal/core/template
  │     └── internal/platform/clipboard
  └── config_category
        └── internal/config

internal/ui/styles
  └── theme

internal/utils
  └── conversion

internal/assets
  └── embed
```

### 2.2 Dependências em Detalhe por Pacote

#### `cmd`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `github.com/spf13/cobra` | Externa | Framework CLI |
| `github.com/spf13/viper` | Externa | Configuração |
| `github.com/rs/zerolog` | Externa | Logging |
| `github.com/charmbracelet/bubbletea` | Externa | TUI |
| `github.com/charmbracelet/lipgloss` | Externa | Estilos |
| `github.com/adrg/xdg` | Externa | Caminhos XDG |
| `internal/config` | Interna | Constantes de config |
| `internal/app` | Interna | ContextService |
| `internal/core/scanner` | Interna | ScanConfig, FileNode |
| `internal/core/template` | Interna | Template, Manager |
| `internal/core/tokens` | Interna | Estimar tokens |
| `internal/platform/clipboard` | Interna | Copiar ao clipboard |
| `internal/platform/openai` | Interna | OpenAI provider |
| `internal/platform/anthropic` | Interna | Anthropic provider |
| `internal/platform/geminiapi` | Interna | Gemini provider |
| `internal/ui` | Interna | Wizard TUI |
| `internal/ui/styles` | Interna | Estilos |
| `internal/utils` | Interna | ParseSize, FormatBytes |

#### `internal/app`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/contextgen` | Interna | Gerar contexto |
| `internal/core/llm` | Interna | Interface Provider |
| `internal/core/scanner` | Interna | Interface Scanner |
| `internal/core/tokens` | Interna | Estimativa |
| `internal/platform/clipboard` | Interna | Copiar |
| `internal/platform/openai` | Interna | Registrar provider |
| `internal/platform/anthropic` | Interna | Registrar provider |
| `internal/platform/geminiapi` | Interna | Registrar provider |

#### `internal/config`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/utils` | Interna | ParseSize para validação |

#### `internal/core/scanner`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/ignore` | Interna | Ignorar arquivos |

#### `internal/core/contextgen`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/scanner` | Interna | FileNode, ScanConfig |
| `internal/core/template` | Interna | TemplateRenderer, Template |

#### `internal/core/template`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/assets` | Interna | //go:embed templates |
| `github.com/adrg/xdg` | Externa | XDG config path |

#### `internal/platform/llmbase`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/platform/http` | Interna | JSONClient |

#### `internal/platform/openai`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/llm` | Interna | Interface Provider, Result |
| `internal/platform/llmbase` | Interna | BaseClient |

#### `internal/platform/anthropic`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/llm` | Interna | Interface Provider, Result |
| `internal/platform/llmbase` | Interna | BaseClient |

#### `internal/platform/geminiapi`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/llm` | Interna | Interface Provider, Result |
| `internal/platform/llmbase` | Interna | BaseClient |

#### `internal/ui`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/app` | Interna | ContextService, GenerateConfig |
| `internal/core/contextgen` | Interna | ContextGenerator |
| `internal/core/llm` | Interna | ProviderType |
| `internal/core/scanner` | Interna | FileNode |
| `internal/core/template` | Interna | Template |
| `internal/platform/clipboard` | Interna | Copy |
| `internal/ui/components` | Interna | Progress, Tree |
| `internal/ui/screens` | Interna | FileSelection, TemplateSelection, etc. |
| `internal/ui/styles` | Interna | Cores, renders |

#### `internal/ui/screens`
| Dependência | Tipo | Motivo |
|-------------|------|--------|
| `internal/core/scanner` | Interna | FileNode |
| `internal/core/template` | Interna | Template |
| `internal/platform/clipboard` | Interna | Copy |
| `internal/config` | Interna | ConfigCategory |
| `internal/ui/styles` | Interna | Estilos |
| `github.com/charmbracelet/bubbletea` | Externa | MVU |
| `github.com/charmbracelet/lipgloss` | Externa | Estilos |

---

## 3. Fluxos de Dados Principais

### 3.1 Fluxo: Geração de Contexto (CLI Headless)

```
cmd/context.go
  → buildGenerateConfig()
  → buildScannerConfig()
  → loadTemplateContent()
    → template.NewManager() → internal/assets (templates embed)
    → Manager.GetTemplate()
  → app.NewContextService()
  → svc.Generate(ctx, svcCfg)
    → scanner.Scan()
      → ignore.LoadGitignore()
      → ignore.LoadShotgunignore()
      → FileSystemScanner.scanRecursive()
    → generator.Generate(tree, selections, config)
      → TreeRenderer.RenderTree()
      → collectFileContents()
      → TemplateRenderer.RenderTemplate()
    → os.WriteFile()
    → clipboard.Copy()
  → printGenerationSummary()
```

### 3.2 Fluxo: Geração de Contexto (TUI Wizard)

```
cmd/root.go → launchTUIWizard()
  → ui.NewWizard(rootPath, scanConfig, wizardConfig, svc)
  → tea.NewProgram(wizard)
  → wizard.Init() → scanDirectoryCmd()
  → wizard.Update()
    → handleStartScan() → scanCoordinator.Start()
    → ScanCompleteMsg → handleScanComplete() → fileSelection.SetFileTree()
    → handleNextStep() → Step 2 → templateSelection.LoadTemplates()
    → TemplateSelectedMsg → handleTemplateMessage()
    → handleNextStep() → Step 3 → taskInput
    → handleNextStep() → Step 4 → rulesInput (condicional)
    → handleNextStep() → Step 5 → review
    → generateContext() → generateContextCmd()
    → startGenerationMsg → generateCoordinator.Start()
    → GenerationCompleteMsg → handleGenerationComplete()
    → clipboardCopyCmd() → clipboard.Copy()
    → handleSendToLLM() → sendToLLMCmd()
      → svc.SendToLLMWithProgress()
        → registry.Create(llmConfig)
          → openai.NewClient() / anthropic.NewClient() / geminiapi.NewClient()
          → baseClient.Send() → JSONClient.PostJSON()
        → ParseResponse()
    → LLMCompleteMsg → handleLLMComplete()
```

### 3.3 Fluxo: Envio de Contexto para LLM (CLI `context send`)

```
cmd/send.go → runContextSend()
  → Read content (file or stdin)
  → CreateLLMProvider(cfg) → cmd/providers.go
    → providerRegistry.Create(cfg) → core/llm/registry.go
      → openai.NewClient() / anthropic.NewClient() / geminiapi.NewClient()
  → provider.Send(ctx, content)
    → BuildRequest() → provider-specific request
    → Send() → baseClient.Send() → JSONClient.PostJSON()
    → ParseResponse() → provider-specific response
  → Output (file or stdout)
```

### 3.4 Fluxo: Configuração (`config set`)

```
cmd/config.go → configSetCmd.RunE()
  → BuildLLMConfig() → cmd/config_llm.go
  → config.IsValidKey(key)
  → config.ValidateValue(key, value)
  → config.ConvertValue(key, value)
  → viper.Set(key, convertedValue)
  → viper.WriteConfig() → filepath to config file
```

---

## 4. Camadas de Arquitetura

```
┌─────────────────────────────────────────────┐
│              Presentation Layer              │
│  cmd/  │  internal/ui/                      │
│  (CLI) │  (TUI Bubble Tea MVU)              │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              Service Layer                   │
│  internal/app/                               │
│  (DefaultContextService, ContextService)     │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│             Domain Layer                     │
│  internal/core/                              │
│  ├── scanner/    (filesystem scanning)       │
│  ├── contextgen/ (context generation)        │
│  ├── diff/       (diff splitting)            │
│  ├── ignore/     (layered ignore engine)     │
│  ├── llm/        (LLM interface/registry)    │
│  ├── template/   (template management)       │
│  └── tokens/     (token estimation)          │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│          Infrastructure Layer                │
│  internal/platform/                          │
│  ├── llmbase/     (shared HTTP client base)  │
│  ├── http/        (JSON HTTP client)         │
│  ├── openai/      (OpenAI provider impl)     │
│  ├── anthropic/   (Anthropic provider impl)  │
│  ├── geminiapi/   (Gemini provider impl)     │
│  └── clipboard/   (system clipboard)         │
└─────────────────────────────────────────────┘

Configuração: internal/config/ (shared across layers)
Utilitários:  internal/utils/ (shared across layers)
Assets:       internal/assets/ (embedded templates)
```

---

## 5. Interfaces e Contratos

| Interface | Pacote | Implementações |
|-----------|--------|----------------|
| `scanner.Scanner` | `internal/core/scanner` | `FileSystemScanner` (implicit) |
| `contextgen.ContextGenerator` | `internal/core/contextgen` | `DefaultContextGenerator` |
| `llm.Provider` | `internal/core/llm` | `openai.Client`, `anthropic.Client`, `geminiapi.Client` |
| `llm.ProviderCreator` | `internal/core/llm` | Funções em `app/providers.go`, `cmd/providers.go` |
| `llm.Registry` | `internal/core/llm` | `DefaultProviderRegistry` (app), `providerRegistry` (cmd) |
| `template.TemplateManager` | `internal/core/template` | `Manager` |
| `template.TemplateSource` | `internal/core/template` | `EmbeddedSource`, `FilesystemSource` |
| `ignore.IgnoreEngine` | `internal/core/ignore` | `LayeredIgnoreEngine` |
| `Sender` | `internal/platform/llmbase` | `*openai.Client`, `*anthropic.Client`, `*geminiapi.Client` (recebedor) |
| `ContextService` | `internal/app` | `DefaultContextService` |

---

## 6. Observações

- **Círculo de dependências**: nenhum ciclo detectado. O fluxo é estritamente em camadas: `cmd → app → core → platform`.
- **Dependência de `internal/assets`** via `//go:embed` permite que templates sejam compilados no binário.
- **Dois registries de providers**: um em `internal/app/providers.go` (para o service layer) e outro em `cmd/providers.go` (para os comandos CLI). São instâncias independentes mas registradas com os mesmos providers.
- **`go-gitignore`** é a única dependência externa usada no core (camada de domínio). Todos os outros pacotes externos são restritos à camada de apresentação e plataforma.
- **Bubble Tea MVU**: a TUI inteira usa o padrão Model-View-Update do Bubble Tea, com mensagens tipadas (`tea.Msg`) para comunicação assíncrona entre componentes.
- **Zero dependências do core**: os pacotes `internal/core/*` não dependem de nenhuma biblioteca externa, apenas de pacotes stdlib Go.
