# Análise de Código — Módulo `cmd`

> **Módulo:** `cmd`  
> **Caminho do pacote:** `github.com/quantmind-br/shotgun-cli/cmd`  
> **Nível de detalhe:** detalhado  
> **Data da análise:** 2026-05-20  
> **Gerado por:** reversa-archaeologist

---

## 1. Visão Geral

O módulo `cmd` é a **camada de apresentação** (presentation layer) do shotgun-cli. É o **ponto de composição** (*composition root*) que define toda a árvore de comandos CLI usando a biblioteca [Cobra](https://github.com/spf13/cobra) e integra-se ao Viper para configuração. Todos os comandos delegam a serviços do pacote `internal/app` e infraestrutura de `internal/platform`/`internal/core`.

### Responsabilidades Principais

| Responsabilidade | Arquivo(s) |
|---|---|
| Entrada principal (root command, TUI wizard) | `root.go` |
: | Inicialização de configuração global (Viper + zerolog) |
| Registro de providers LLM | `providers.go` |
| Configuração de LLM | `config_llm.go` |
| Geração de contexto (`context generate`) | `context.go` |
| Envio para LLM (`context send`) | `send.go` |
| Configuração (`config show/set/TUI`) | `config.go` |
| Diagnóstico LLM (`llm status/doctor/list`) | `llm.go` |
| Templates (`template list/render/import/export`) | `template.go` |
| Split de diffs (`diff split`) | `diff.go` |
| Shell completion | `completion.go` |

### Métricas do Módulo

| Métrica | Valor |
|---|---|
| Arquivos `.go` (excl. testes) | 10 |
| Arquivos `_test.go` | 10 |
| Funções/métodos (estimado) | ~55 |
| Tipos definidos | 6 (ProgressMode, ProgressOutput, GenerateConfig, CLIConfig duplicado) |
| Cobra commands definidos | 12 (root + 11 subcomandos) |
| Flags registradas | ~60 |
| Dependências externas (pacotes) | 12 |
| Dependências internas (pacotes) | 8 |

---

## 2. Dependências Internas (pacotes)

| Pacote | Uso no módulo | Tipo |
|---|---|---|
| `github.com/quantmind-br/shotgun-cli/internal/app` | `ContextService`, `GenerateConfig`, `GenerateResult`, `ProgressCallback` | Serviço de aplicação |
| `github.com/quantmind-br/shotgun-cli/internal/config` | Constantes de chaves (`KeyLLMProvider`, etc.), `IsValidKey`, `ValidateValue`, `ConvertValue` | Configuração |
| `github.com/quantmind-br/shotgun-cli/internal/core/llm` | `Registry`, `Provider`, `Config`, `ProviderType` | LLM core |
| `github.com/quantmind-br/shotgun-cli/internal/core/scanner` | `ScanConfig`, `Progress` | Scanner core |
| `github.com/quantmind-br/shotgun-cli/internal/core/template` | `Manager`, `ManagerConfig`, `Template` | Template core |
| `github.com/quantmind-br/shotgun-cli/internal/core/tokens` | `FormatTokens`, `EstimateFromBytes` | Token estimation |
| `github.com/quantmind-br/shotgun-cli/internal/ui` | `Wizard`, `WizardConfig`, `LLMConfig`, `ContextConfig`, `ConfigWizard` | Interface TUI |
| `github.com/quantmind-br/shotgun-cli/internal/ui/styles` | Estilos lipgloss (`TitleStyle`, `StatsLabelStyle`, etc.) | Estilos UI |
| `github.com/quantmind-br/shotgun-cli/internal/utils` | `ParseSize`, `ParseSizeWithDefault`, `FormatBytes` | Utilitários |
| `github.com/quantmind-br/shotgun-cli/internal/platform/anthropic` | `NewClient` | Provider Anthropic |
| `github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi` | `NewClient` | Provider Gemini |
| `github.com/quantmind-br/shotgun-cli/internal/platform/openai` | `NewClient` | Provider OpenAI |

---

## 3. Dependências Externas

| Pacote | Versão / Uso |
|---|---|
| `github.com/spf13/cobra` | Framework de comandos CLI |
| `github.com/spf13/viper` | Configuração (YAML + env vars) |
| `github.com/charmbracelet/bubbletea` | TUI framework (Bubble Tea) |
| `github.com/charmbracelet/lipgloss` | Estilização de terminal |
| `github.com/rs/zerolog` | Logging estruturado |
| `github.com/adrg/xdg` | Diretórios XDG (template import/export) |

---

## 4. Análise Detalhada por Arquivo

### 4.1 `root.go` — Ponto de Entrada Principal

**Arquitetura:** Define `rootCmd` (cobra.Command), `Execute()`, e a lógica de inicialização.

**Principais funções:**
- `Execute()` — Entrada principal chamada por `main.go`. Invoca `rootCmd.Execute()`.
- `runRootCommand()` — Se nenhum subcomando: lança TUI Wizard. Se flags de versão: imprime versão.
- `launchTUIWizard()` — Lê configuração do Viper, configura `WizardConfig`, cria e roda Bubble Tea program com tela cheia e mouse.
- `initConfig()` — Configuração hierárquica: zerolog → Viper (config file → env vars → defaults) → flags → logging.
- `getConfigDir()` — Retorna o diretório de configuração baseado no SO (XDG, AppData, Library/Application Support).
- `setConfigDefaults()` — Define todos os valores padrão do Viper para scanner, context, output, llm, template.
- `updateLoggingLevel()` — Ajusta zerolog para Debug (verbose) ou Error (quiet).
- `init()` — Registra `cobra.OnInitialize(initConfig)` e flags globais.

**Pontos de atenção:**
- 🟡 `MaxMemory` no `ScanConfig` é definido mas não é implementado no scanner (tecnologia debt identificada pelo architect phase).
- O wizard usa `includeIgnored: true` hardcoded em `launchTUIWizard()`, ignorando o valor do config.
- `initConfig()` é chamada por `cobra.OnInitialize`, que é executado durante `rootCmd.Execute()`. A ordem de inicialização depende da hierarquia de comandos Cobra.

### 4.2 `providers.go` — Registro de Providers LLM

**Padrão:** Registry com inicialização em `init()`.

**Funções:**
- `init()` — Cria `providerRegistry` global via `llm.NewRegistry()` e registra OpenAI, Anthropic, Gemini com closures que instanciam os providers.
- `CreateLLMProvider(cfg)` — Fábrica que usa o registry para criar o provider.
- `GetProviderRegistry()` — Retorna o registry singleton para uso externo.

**Pontos de atenção:**
- Singleton global via variável package-level — pode dificultar testes isolados.
- Registry é recriado em `init()`, não em `NewContextService()`. Há duas instâncias potenciais.

### 4.3 `config_llm.go` — Construção de Config LLM

**Funções:**
- `BuildLLMConfig()` — Constrói `llm.Config` a partir do Viper, aplicando defaults do provider.
- `BuildLLMConfigWithOverrides(model, timeout)` — Aplica overrides de flags sobre config base.

**Pontos de atenção:**
- Nenhuma validação de config nesta camada — a validação é delegada a `llm.Config.Validate()` (chamado nos subcomandos).
- A máscara de API key é feita via `cfg.MaskAPIKey()` (definição em `internal/core/llm/config.go`).

### 4.4 `context.go` — Geração de Contexto

**Tipos definidos:**
- `ProgressMode` — `none`, `human`, `json`
- `ProgressOutput` — Estrutura JSON com timestamp, stage, message, current, total, percent
- `GenerateConfig` — Configuração de geração de contexto

**Fluxo principal `generateContextHeadless()`:**
1. Constrói `scannerConfig` a partir de flags + defaults
2. Constrói `templateVars` (TASK, RULES, FILE_STRUCTURE, CURRENT_DATE, custom vars)
3. Carrega template via `template.Manager`
4. Chama `app.NewContextService()` → `Generate()` ou `GenerateWithProgress()`
5. Imprime summary com contagem de arquivos, tamanho, estimativa de tokens

**Pontos de atenção:**
- `buildGenerateConfig()` é complexa (80+ linhas) — viola SRP, deveria ser decomposta.
- Nome de output padrão usa timestamp `YYYYMMDD-HHMMSS` hardcoded.
- O comando `context send` (em `send.go`) usa `BuildLLMConfigWithOverrides` com flags `model` e `timeout` mas a descrição menciona "Gemini" especificamente — é um artefato de refatoração incompleta.

### 4.5 `send.go` — Envio para LLM

**Fluxo `runContextSend()`:**
1. Lê conteúdo de arquivo ou stdin
2. Valida conteúdo não vazio
3. Aplica overrides de modelo/timeout
4. Cria provider via `CreateLLMProvider()`
5. Valida config e disponibilidade
6. Envia via `llmProvider.Send()`
7. Salva resposta em arquivo ou imprime
8. Mostra uso de tokens e duração

**Pontos de atenção:**
- Nome do comando diz "Gemini" mas usa o provider configurado genericamente (abstraçao correta).
- Flag `--raw` usa `result.RawResponse` vs `result.Response` processado.

### 4.6 `config.go` — Configuração

**Comandos:** `config` (parent), `config show`, `config set`, e TUI interativo.

**Fluxo `showCurrentConfig()`:**
1. Obtém e exibe caminho do config file
2. Agrupa todas as chaves Viper por categoria (scanner, context, template, output, llm)
3. Exibe com estilos lipgloss indicando a fonte do valor (config file, environment, flag/default)

**Fluxo `setConfigValue()`:**
1. Converte o valor (tipos) via `config.ConvertValue()`
2. Define no Viper
3. Cria diretório se necessário
4. Escreve via `viper.WriteConfig()` (ou `SafeWriteConfig()` para arquivo novo)

**Fluxo `launchConfigTUI()`:**
1. Cria `ui.NewConfigWizard()`
2. Roda com Bubble Tea (alt screen + mouse)

**Pontos de atenção:**
- `getConfigSource()` é simplista — Viper não expõe a fonte real do valor (file vs env vs flag vs default).
- O TUI de config é separado do wizard principal de geração.

### 4.7 `llm.go` — Gerenciamento de LLM

**Comandos:** `llm` (parent), `llm status`, `llm doctor`, `llm list`.

**Fluxo `runLLMStatus()`:**
1. Constrói config via `BuildLLMConfig()`
2. Exibe provider, modelo, base URL, API key (mascarada), timeout
3. Cria provider e verifica disponibilidade/configuração

**Fluxo `runLLMDoctor()`:**
1. Verifica provider válido
2. Verifica API key configurada
3. Verifica modelo configurado
4. Verifica disponibilidade do provider
5. Verifica configuração completa
6. Exibe orientações específicas por provider

**Fluxo `runLLMList()`:**
1. Lista todos os providers suportados
2. Marca o atual com `*`
3. Mostra URLs de obtenção de API key

### 4.8 `template.go` — Gerenciamento de Templates

**Comandos:** `template` (parent), `template list`, `template render`, `template import`, `template export`.

**Fluxo `renderTemplate()`:**
1. Cria `template.Manager`
2. Valida variáveis obrigatórias (`requiredVars`)
3. Renderiza template com Go templates
4. Escreve para arquivo ou stdout

**Fluxo `templateImportCmd.RunE`:**
1. Lê arquivo do usuário
2. Extrai nome do arquivo (remove extensão e prefixo `prompt_`)
3. Pede confirmação se já existe
4. Salva em `~/.config/shotgun-cli/templates/`

**Fluxo `templateExportCmd.RunE`:**
1. Busca template por nome
2. Pede confirmação se arquivo de destino existe
3. Salva em diretório de destino

**Pontos de atenção:**
- `template import` usa `xdg.ConfigHome` hardcoded ao invés de `getDefaultConfigPath()` — inconsistência de caminhos.

### 4.9 `diff.go` — Split de Diffs

**Comando:** `diff split`.

**Fluxo `splitDiffFile()`:**
1. Cria diretório de saída
2. Lê arquivo diff inteiro em memória
3. Chama `diff.IntelligentSplit()` (do pacote `internal/core/diff`)
4. Escreve chunks com metadados opcionais

**Pontos de atenção:**
- Leitura inteira do arquivo em memória (`allLines`) — não escalável para diffs muito grandes.
- `writeDiffChunk` é uma função dedicada com header condicional.

### 4.10 `completion.go` — Shell Completion

**Comando:** `completion [bash|zsh|fish|powershell]`.

**Completers customizados:**
- `configKeyCompletion()` — Lista todas as chaves de config com descrições
- `boolValueCompletion()` — Retorna `true`/`false` para chaves booleanas, `markdown`/`text` para `output.format`, file completion para `template.custom-path`

**Pontos de atenção:**
- Completers são registrados apenas se `configSetCmd` já foi inicializado (dependência de ordem de `init()`).
- `configKeyTemplateCustomPath` é definido localmente e usado no completer, mas a constante real está em `internal/config/keys.go`.

---

## 5. Padrões Arquiteturais Identificados

### 5.1 Composition Root
O módulo `cmd` é o composition root do aplicativo. Todas as dependências externas são resolvidas aqui: providers LLM, scanner, generators, UI wizards.

### 5.2 Service Delegation
Cada comando delega para `app.NewContextService()` ou `CreateLLMProvider()` — os commands nunca instanciam lógica de negócio diretamente.

### 5.3 Config Hierárquica (Viper)
- Defaults em `setConfigDefaults()`
- Config file YAML (localização multiplataforma)
- Variáveis de ambiente com prefixo `SHOTGUN_`
- Flags de linha de comando
- Override order: flags > env > file > defaults

### 5.4 Progress Reporting
Modo triplo: `none` / `human` / `json`. O modo `human` usa `\r` para overwriting da linha. O modo `json` serializa cada evento como JSON.

### 5.5 TUI com Bubble Tea
Dois wizards: `Wizard` (geração) e `ConfigWizard` (configuração). Ambos usam `tea.WithAltScreen()` e `tea.WithMouseCellMotion()`.

### 5.6 Registry Pattern
`llm.Registry` com `ProviderCreator` closures. Registrado globalmente em `providers.go` via `init()`.

---

## 6. Qualidade de Código

### 6.1 Cobertura de Testes
Todos os 10 arquivos têm arquivos de teste correspondentes. Padrão: tests table-driven com `captureStdout`/`captureStderr`.

### 6.2 Pontos Fracos Identificados

| Issue | Arquivo | Severidade | Descrição |
|---|---|---|---|
| MAX_MEMORY não implementado | `root.go`, `context.go` | Alta | `MaxMemory` é configurado mas o scanner não o respeita |
| Template import path inconsistency | `template.go` | Média | Usa `xdg.ConfigHome` ao invés de `getDefaultConfigPath()` |
| `buildGenerateConfig` complexidade | `context.go` | Média | ~80 linhas, viola SRP |
| Wizard hardcoded `IncludeIgnored: true` | `root.go` | Baixa | Ignora configuração do usuário |
| `getConfigSource` impreciso | `config.go` | Baixa | Viper não expõe fonte real do valor |
| Variável local redeclarando constante | `completion.go` | Baixa | `configKeyTemplateCustomPath` local vs `config.KeyTemplateCustomPath` global |

### 6.3 Pontos Fortes

| Aspecto | Descrição |
|---|---|
| Testes | Cobertura extensiva com table-driven tests e captura de stdout/stderr |
| Delegação | Commands nunca executam lógica de negócio diretamente |
| Validação | `PreRunE` em quase todos os subcomandos com validação antecipada |
| Internacionalização de caminho | `getConfigDir()` respeita XDG, AppData, Application Support |
| Feedback de progresso | Triplo modo (none/human/json) com formatação consistente |
| Template system | Variáveis obrigatórias validadas antes da renderização |
| Provider abstraction | Interface unificada `llm.Provider` para todos os LLMs |

---

## 7. Fluxos Principais (Resumo)

1. **TUI Wizard Flow** — `root` → `launchTUIWizard()` → `ui.NewWizard()` → Bubble Tea loop
2. **Context Generate Flow** — `context generate` → `buildGenerateConfig()` → `app.ContextService.Generate()` → scanner → generator → output
3. **Config Set Flow** — `config set` → `setConfigValue()` → `config.ConvertValue()` → `viper.WriteConfig()`
4. **LLM Doctor Flow** — `llm doctor` → `BuildLLMConfig()` → provider checks → diagnostic report
5. **Template Render Flow** — `template render` → `template.Manager.RenderTemplate()` → var validation → Go template execute
6. **Diff Split Flow** — `diff split` → `splitDiffFile()` → `diff.IntelligentSplit()` → chunk writing
7. **Send to LLM Flow** — `context send` → file/stdin read → provider send → output/save

---

## 8. Gap Analysis

| Gap | Descrição | Impacto |
|---|---|---|
| Nenhum rate limiting | LLM calls sem retry policy ou rate limiting | Crítico — pode causar bans de API |
| Nenhum context cache | Cada geração de contexto reescaneia o filesystem | Alto — desperdício de I/O e tempo |
| `MaxMemory` sem implementação | Configura mas não é aplicado pelo scanner | Alto — risco de OOM em projetos grandes |
| Config source impreciso | `getConfigSource()` não pode distinguir env de file | Médio — `config show` pode mostrar fonte errada |
| Provas de conceito de Gemini | `send.go` menciona Gemini hardcoded na help text | Baixo — pode confundir usuários |
