# Plano de Cobertura de Testes - Meta: 90%

**Projeto**: shotgun-cli  
**Framework de Testes**: Go test  
**Data**: Janeiro 2026  
**Autor**: Sisyphus AI Agent

---

## Status Atual

### Cobertura Geral

| Metrica | Valor |
|---------|-------|
| **Cobertura atual** | 73.7% |
| **Meta** | 90% |
| **Gap** | 16.3% |

### Cobertura por Pacote/Modulo

| Pacote | Cobertura Atual | Meta | Gap | Prioridade |
|--------|----------------|------|-----|------------|
| internal/core/tokens | 100.0% | 90% | - | Concluido |
| internal/utils | 100.0% | 90% | - | Concluido |
| internal/core/diff | 98.4% | 90% | - | Concluido |
| internal/core/llm | 96.4% | 90% | - | Concluido |
| internal/ui/components | 94.0% | 90% | - | Concluido |
| internal/core/template | 92.9% | 90% | - | Concluido |
| internal/core/context | 91.2% | 90% | - | Concluido |
| internal/core/ignore | 89.6% | 90% | 0.4% | Baixa |
| internal/app | 87.5% | 90% | 2.5% | Media |
| internal/ui/styles | 84.7% | 90% | 5.3% | Baixa |
| internal/core/scanner | 84.4% | 90% | 5.6% | Media |
| internal/platform/clipboard | 83.3% | 90% | 6.7% | Media |
| internal/ui/screens | 83.1% | 90% | 6.9% | Media |
| internal/platform/openai | 74.2% | 90% | 15.8% | Alta |
| internal/platform/anthropic | 70.0% | 90% | 20.0% | Alta |
| internal/platform/geminiapi | 69.7% | 90% | 20.3% | Alta |
| internal/platform/gemini | 68.7% | 90% | 21.3% | Alta |
| internal/config | 67.0% | 90% | 23.0% | Alta |
| internal/ui (wizard.go) | 56.0% | 90% | 34.0% | Alta |
| cmd | 37.0% | 90% | 53.0% | Alta |
| internal/assets | N/A | N/A | - | N/A (embed only) |
| main.go | 0.0% | N/A | - | N/A (entrypoint) |

---

## Analise Detalhada

### Arquivos Sem Testes

| Arquivo | Razao | Acao Recomendada |
|---------|-------|------------------|
| `main.go` | Entrypoint da aplicacao | Nao requer testes unitarios |
| `internal/assets/embed.go` | Apenas embedding de assets | Nao requer testes unitarios |

### Arquivos com Baixa Cobertura (<60%)

#### 1. `cmd/` - 37.0% Cobertura

**Funcoes com 0% cobertura:**

| Funcao | Arquivo | Impacto | Complexidade | Testabilidade |
|--------|---------|---------|--------------|---------------|
| `showCurrentConfig` | config.go:154 | Medio | Baixa | Alta |
| `getGeminiStatusSummary` | config.go:226 | Baixo | Baixa | Alta |
| `setConfigValue` | config.go:250 | Alto | Media | Media |
| `getDefaultConfigPath` | config.go:331 | Baixo | Baixa | Alta |
| `printGenerationSummary` | context.go:386 | Baixo | Baixa | Alta |
| `sendToGemini` | context.go:397 | Alto | Alta | Baixa |
| `renderProgressHuman` | context.go:453 | Baixo | Baixa | Alta |
| `renderProgressJSON` | context.go:463 | Baixo | Baixa | Alta |
| `renderProgress` | context.go:469 | Baixo | Baixa | Alta |
| `runGeminiStatus` | gemini.go:68 | Medio | Media | Media |
| `printGeminiStatusHuman` | gemini.go:79 | Baixo | Baixa | Alta |
| `printGeminiNextSteps` | gemini.go:139 | Baixo | Baixa | Alta |
| `printGeminiStatusJSON` | gemini.go:157 | Baixo | Baixa | Alta |
| `runGeminiDoctor` | gemini.go:186 | Medio | Media | Media |
| `runLLMStatus` | llm.go:59 | Medio | Media | Media |
| `runLLMDoctor` | llm.go:101 | Medio | Media | Media |
| `runLLMList` | llm.go:193 | Baixo | Baixa | Alta |
| `displayURL` | llm.go:230 | Baixo | Baixa | Alta |
| `GetProviderRegistry` | providers.go:51 | Baixo | Baixa | Alta |
| `Execute` | root.go:125 | Alto | Alta | Baixa |
| `launchTUIWizard` | root.go:61 | Alto | Alta | Baixa |
| `formatDuration` | send.go:157 | Baixo | Baixa | Alta |

**Funcoes com cobertura parcial (<50%):**

| Funcao | Arquivo | Cobertura | Branches Faltantes |
|--------|---------|-----------|-------------------|
| `completion.init` | completion.go:156 | 37.5% | Completions de shell |
| `getConfigSource` | config.go:289 | 22.2% | Casos de source |
| `generateContextHeadless` | context.go:248 | 48.3% | Error paths |
| `loadTemplateContent` | context.go:365 | 20.0% | File loading errors |
| `clearProgressLine` | context.go:481 | 50.0% | JSON mode |
| `runRootCommand` | root.go:41 | 33.3% | TUI launch path |
| `getConfigDir` | root.go:210 | 27.3% | XDG/fallback paths |

#### 2. `internal/ui/wizard.go` - 56.0% Cobertura

**Funcoes com 0% cobertura:**

| Funcao | Linha | Complexidade | Estrategia de Teste |
|--------|-------|--------------|---------------------|
| `handleSendToGemini` | 590 | Alta | Mock LLM provider |
| `createLLMProvider` | 663 | Media | Mock registry |
| `sendToLLMCmd` | 697 | Alta | Integration test |
| `handleTemplateMessage` | 728 | Baixa | Unit test |
| `handleRescanRequest` | 774 | Media | Unit test |
| `schedulePollScan` | 1005 | Media | Mock tea.Cmd |
| `schedulePollGenerate` | 1011 | Media | Mock tea.Cmd |
| `pollScan` | 1041 | Alta | Mock channels |
| `finishScan` | 1068 | Media | Unit test |
| `pollGenerate` | 1109 | Alta | Mock channels |
| `finishGeneration` | 1138 | Media | Unit test |
| `finalizeGeneration` | 1144 | Media | Unit test |
| `validateContentSize` | 1165 | Baixa | Unit test |
| `saveGeneratedContent` | 1184 | Media | Temp file test |

**Funcoes com cobertura parcial (<50%):**

| Funcao | Cobertura | Branches Faltantes |
|--------|-----------|-------------------|
| `iterativeScanCmd` | 9.1% | Scan iterations |
| `iterativeGenerateCmd` | 7.1% | Generate iterations |
| `handleStepInput` | 33.3% | Step-specific inputs |
| `clipboardCopyCmd` | 33.3% | Error handling |

#### 3. `internal/config/validator.go` - 67.0% Cobertura

**Funcoes com 0% cobertura:**

| Funcao | Razao | Acao |
|--------|-------|------|
| `validatePath` | Valida paths no filesystem | Criar tests com temp dirs |
| `validateGeminiModel` | Valida modelos Gemini | Adicionar table tests |
| `validateBrowserRefresh` | Valida refresh interval | Adicionar boundary tests |

#### 4. `internal/platform/` - ~70% Media

**Funcoes com 0% cobertura por provider:**

| Provider | Funcao | Impacto |
|----------|--------|---------|
| anthropic | `Name()`, `IsAvailable()`, `ValidateConfig()` | Baixo |
| anthropic | `ValidModels()`, `IsKnownModel()` | Baixo |
| geminiapi | `Name()`, `IsAvailable()`, `ValidateConfig()` | Baixo |
| geminiapi | `ValidModels()`, `IsKnownModel()` | Baixo |
| openai | `Name()`, `IsAvailable()`, `ValidateConfig()` | Baixo |
| openai | `ValidModels()`, `IsKnownModel()` | Baixo |
| gemini | **Todas as funcoes em provider.go** | Alto |

### Codigo Critico Nao Testado

#### Alta Prioridade (Core Business Logic)

1. **`cmd/context.go:sendToGemini`** - Integracao com LLM
2. **`cmd/root.go:Execute`** - Entry point da CLI
3. **`internal/ui/wizard.go:handleSendToGemini`** - Workflow de envio LLM na TUI
4. **`internal/platform/gemini/provider.go`** - Provider web nao testado

#### Media Prioridade (User-Facing)

1. **`cmd/config.go:setConfigValue`** - Configuracao do usuario
2. **`cmd/llm.go:runLLMStatus`** - Diagnostico de providers
3. **`cmd/llm.go:runLLMDoctor`** - Verificacao de saude

---

## Estrategia de Implementacao

### Fase 1: Quick Wins (Semana 1-2)

**Objetivo**: +8% cobertura (73.7% -> ~82%)

#### 1.1 Platform Providers - Metodos Simples

**Estimativa**: +3% cobertura global

```go
// internal/platform/anthropic/client_test.go - ADICIONAR
func TestClientName(t *testing.T) {
    client, _ := NewClient(testConfig)
    assert.Equal(t, "anthropic", client.Name())
}

func TestClientIsAvailable(t *testing.T) {
    client, _ := NewClient(testConfig)
    available := client.IsAvailable()
    // Deve retornar false sem API key valida
    assert.False(t, available)
}

func TestValidModels(t *testing.T) {
    models := ValidModels()
    assert.Contains(t, models, "claude-3-opus-20240229")
}

func TestIsKnownModel(t *testing.T) {
    assert.True(t, IsKnownModel("claude-3-opus-20240229"))
    assert.False(t, IsKnownModel("unknown-model"))
}
```

Repetir para:
- `internal/platform/openai/` (+1%)
- `internal/platform/geminiapi/` (+1%)

#### 1.2 Config Validator - Funcoes Simples

**Estimativa**: +2% cobertura global

```go
// internal/config/validator_test.go - ADICIONAR
func TestValidatePath(t *testing.T) {
    tmpDir := t.TempDir()
    tests := []struct {
        name    string
        path    string
        wantErr bool
    }{
        {"valid existing dir", tmpDir, false},
        {"non-existent path", "/path/does/not/exist/12345", true},
        {"empty path", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validatePath(tt.path)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateGeminiModel(t *testing.T) {
    tests := []struct {
        model   string
        wantErr bool
    }{
        {"gemini-pro", false},
        {"gemini-1.5-pro", false},
        {"invalid-model", true},
    }
    // ...
}

func TestValidateBrowserRefresh(t *testing.T) {
    tests := []struct {
        value   string
        wantErr bool
    }{
        {"1000", false},   // 1 segundo
        {"60000", false},  // 1 minuto
        {"0", true},       // invalido
        {"-1", true},      // invalido
        {"abc", true},     // nao numerico
    }
    // ...
}
```

#### 1.3 CMD Helper Functions

**Estimativa**: +3% cobertura global

```go
// cmd/send_test.go - ADICIONAR
func TestFormatDuration(t *testing.T) {
    tests := []struct {
        duration time.Duration
        expected string
    }{
        {500 * time.Millisecond, "500ms"},
        {1500 * time.Millisecond, "1.5s"},
        {65 * time.Second, "1m5s"},
    }
    for _, tt := range tests {
        result := formatDuration(tt.duration)
        assert.Equal(t, tt.expected, result)
    }
}

// cmd/llm_test.go - ADICIONAR
func TestDisplayURL(t *testing.T) {
    tests := []struct {
        url      string
        expected string
    }{
        {"https://api.openai.com/v1", "https://api.openai.com/v1"},
        {"", "(not configured)"},
    }
    // ...
}

// cmd/context_test.go - ADICIONAR
func TestRenderProgressHuman(t *testing.T) {
    // Capture stdout
    // Verify output format
}

func TestRenderProgressJSON(t *testing.T) {
    // Capture stdout
    // Verify JSON structure
}
```

---

### Fase 2: Codigo Critico (Semana 3-4)

**Objetivo**: +5% cobertura (82% -> ~87%)

#### 2.1 CMD Config Operations

**Arquivos**: `cmd/config.go`, `cmd/config_test.go`

```go
func TestShowCurrentConfig(t *testing.T) {
    // Setup viper with test values
    viper.Set("scanner.max-files", 1000)
    viper.Set("llm.provider", "openai")
    
    var buf bytes.Buffer
    oldStdout := os.Stdout
    // Capture and verify output
}

func TestSetConfigValue(t *testing.T) {
    // Use temp config file
    tmpFile := filepath.Join(t.TempDir(), "config.yaml")
    viper.SetConfigFile(tmpFile)
    
    tests := []struct {
        key     string
        value   string
        wantErr bool
    }{
        {"scanner.max-files", "500", false},
        {"invalid.key", "value", true},
        {"scanner.max-files", "invalid", true},
    }
    // ...
}

func TestGetConfigSource(t *testing.T) {
    tests := []struct {
        key      string
        expected string
    }{
        // Test default, env, flag, config file sources
    }
}
```

#### 2.2 LLM Status/Doctor Commands

**Arquivos**: `cmd/llm.go`, `cmd/llm_test.go`

```go
func TestRunLLMStatus(t *testing.T) {
    // Mock provider registry
    // Verify status output for each provider
}

func TestRunLLMDoctor(t *testing.T) {
    // Test diagnostic checks
    // API key validation
    // Model availability
}

func TestRunLLMList(t *testing.T) {
    // Verify all registered providers listed
}
```

#### 2.3 Gemini Web Provider

**Arquivos**: `internal/platform/gemini/provider.go`, `internal/platform/gemini/provider_test.go`

```go
func TestNewWebProvider(t *testing.T) {
    cfg := llm.Config{
        Model:   "gemini-pro",
        Timeout: 30 * time.Second,
    }
    provider, err := NewWebProvider(cfg)
    assert.NoError(t, err)
    assert.NotNil(t, provider)
}

func TestWebProviderName(t *testing.T) {
    provider, _ := NewWebProvider(testConfig)
    assert.Equal(t, "gemini-web", provider.Name())
}

func TestWebProviderIsAvailable(t *testing.T) {
    // Test with/without gemini-cli binary
}

func TestWebProviderIsConfigured(t *testing.T) {
    // Test with/without cookies file
}
```

---

### Fase 3: TUI Coverage (Semana 5-6)

**Objetivo**: +5% cobertura (87% -> ~92%)

#### 3.1 Wizard Helper Functions

**Arquivo**: `internal/ui/wizard_test.go`

```go
func TestValidateContentSize(t *testing.T) {
    tests := []struct {
        content   string
        maxSize   int64
        shouldErr bool
    }{
        {"small content", 1024, false},
        {strings.Repeat("x", 2000), 1024, true},
    }
    for _, tt := range tests {
        m := &Model{maxContentSize: tt.maxSize}
        err := m.validateContentSize(tt.content)
        if tt.shouldErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    }
}

func TestParseSize_Extended(t *testing.T) {
    // Testes adicionais para edge cases
    tests := []struct {
        input    string
        expected int64
        wantErr  bool
    }{
        {"0", 0, false},
        {"1", 1, false},
        {"-1", 0, true},
        {"abc", 0, true},
        {"1.5KB", 0, true}, // floats not supported
    }
}

func TestHandleTemplateMessage(t *testing.T) {
    // Test template loading messages
}

func TestFinishScan(t *testing.T) {
    // Test scan completion handling
}

func TestFinishGeneration(t *testing.T) {
    // Test generation completion
}
```

#### 3.2 Wizard State Transitions

```go
func TestWizardStateTransitions(t *testing.T) {
    tests := []struct {
        name          string
        initialStep   WizardStep
        message       tea.Msg
        expectedStep  WizardStep
    }{
        {"next from template", StepTemplateSelection, tea.KeyMsg{Type: tea.KeyEnter}, StepTaskInput},
        {"prev from task", StepTaskInput, tea.KeyMsg{Type: tea.KeyEsc}, StepTemplateSelection},
        // ...
    }
}

func TestIterativeScanCmd(t *testing.T) {
    // Mock scanner
    // Test incremental scan progress
}

func TestIterativeGenerateCmd(t *testing.T) {
    // Mock generator
    // Test incremental generation
}
```

#### 3.3 LLM Integration in TUI

```go
func TestCreateLLMProvider(t *testing.T) {
    // Mock registry
    registry := llm.NewRegistry()
    registry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
        return &mockProvider{}, nil
    })
    
    m := &Model{providerRegistry: registry}
    provider, err := m.createLLMProvider()
    assert.NoError(t, err)
    assert.NotNil(t, provider)
}

func TestHandleSendToGemini(t *testing.T) {
    // Mock provider that returns success
    // Verify state transitions
}
```

---

### Fase 4: CMD Integration & Polish (Semana 7-8)

**Objetivo**: +3% cobertura (92% -> 95%+)

#### 4.1 Root Command

```go
func TestExecute(t *testing.T) {
    // Test CLI execution with various flags
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    
    tests := []struct {
        args     []string
        wantErr  bool
    }{
        {[]string{"shotgun", "--help"}, false},
        {[]string{"shotgun", "context", "--help"}, false},
        {[]string{"shotgun", "invalid-command"}, true},
    }
}

func TestRunRootCommand(t *testing.T) {
    // Test with various flags
    // --json, --verbose, --quiet
}
```

#### 4.2 Context Command Full Path

```go
func TestGenerateContextHeadless_AllPaths(t *testing.T) {
    tests := []struct {
        name     string
        flags    map[string]interface{}
        wantErr  bool
    }{
        {"minimal", map[string]interface{}{"path": "."}, false},
        {"with template", map[string]interface{}{"path": ".", "template": "default"}, false},
        {"with output", map[string]interface{}{"path": ".", "output": "out.md"}, false},
        {"invalid path", map[string]interface{}{"path": "/nonexistent"}, true},
    }
}

func TestLoadTemplateContent(t *testing.T) {
    // Test with embedded template
    // Test with custom template file
    // Test with non-existent template
}
```

#### 4.3 Scanner Edge Cases

```go
// internal/core/scanner/filesystem_test.go - ADICIONAR
func TestHandleCountError(t *testing.T) {
    // Test error handling during item counting
}

func TestHandleWalkError(t *testing.T) {
    // Test error handling during directory walk
}

func TestShouldSkipLargeFile_EdgeCases(t *testing.T) {
    // Test boundary conditions
}

func TestClassifyIgnoreReason_AllCases(t *testing.T) {
    // Test all ignore reason types
}
```

---

## Checklist de Testes por Arquivo

### cmd/config.go

**Cobertura atual**: ~45%

**Funcoes nao testadas**:
- [ ] `showCurrentConfig()` - Cenarios: output normal, JSON mode, empty config
- [ ] `getGeminiStatusSummary()` - Cenarios: configured, not configured
- [ ] `setConfigValue()` - Cenarios: valid key/value, invalid key, invalid value, file write error
- [ ] `getDefaultConfigPath()` - Cenarios: XDG set, XDG not set, fallback

**Branches nao cobertas**:
- [ ] `getConfigSource` linha 289: testar caso "flag"
- [ ] `getConfigSource` linha 289: testar caso "env"
- [ ] `getConfigSource` linha 289: testar caso "config"

### cmd/context.go

**Cobertura atual**: ~55%

**Funcoes nao testadas**:
- [ ] `printGenerationSummary()` - Cenarios: with/without LLM, various sizes
- [ ] `sendToGemini()` - Cenarios: success, provider error, timeout
- [ ] `renderProgressHuman()` - Cenarios: different progress stages
- [ ] `renderProgressJSON()` - Cenarios: different progress stages
- [ ] `renderProgress()` - Cenarios: human mode, JSON mode

**Branches nao cobertas**:
- [ ] `generateContextHeadless` linha 270: template not found error
- [ ] `generateContextHeadless` linha 285: scanner error
- [ ] `loadTemplateContent` linha 375: custom template error

### cmd/llm.go

**Cobertura atual**: ~30%

**Funcoes nao testadas**:
- [ ] `runLLMStatus()` - Cenarios: all providers, single provider, no providers
- [ ] `runLLMDoctor()` - Cenarios: all checks pass, some fail, all fail
- [ ] `runLLMList()` - Cenarios: with/without custom URL
- [ ] `displayURL()` - Cenarios: with URL, empty URL

### internal/ui/wizard.go

**Cobertura atual**: 56%

**Funcoes nao testadas**:
- [ ] `handleSendToGemini()` - Cenarios: success, error, timeout, cancel
- [ ] `createLLMProvider()` - Cenarios: valid config, invalid config, no registry
- [ ] `sendToLLMCmd()` - Cenarios: streaming response, error
- [ ] `handleTemplateMessage()` - Cenarios: valid template, error
- [ ] `handleRescanRequest()` - Cenarios: trigger rescan
- [ ] `schedulePollScan()` - Cenarios: timing
- [ ] `schedulePollGenerate()` - Cenarios: timing
- [ ] `pollScan()` - Cenarios: progress, complete, error
- [ ] `finishScan()` - Cenarios: with results, empty
- [ ] `pollGenerate()` - Cenarios: progress, complete, error
- [ ] `finishGeneration()` - Cenarios: success
- [ ] `finalizeGeneration()` - Cenarios: save, copy, send
- [ ] `validateContentSize()` - Cenarios: valid, too large
- [ ] `saveGeneratedContent()` - Cenarios: success, write error

**Branches nao cobertas**:
- [ ] `iterativeScanCmd` linhas 1017-1040: scan loop iterations
- [ ] `iterativeGenerateCmd` linhas 1077-1108: generate loop iterations
- [ ] `handleStepInput` linhas 892-927: all step types
- [ ] `clipboardCopyCmd` linha 965: error path

### internal/platform/gemini/provider.go

**Cobertura atual**: 0%

**Funcoes nao testadas**:
- [ ] `NewWebProvider()` - Cenarios: valid config, invalid config
- [ ] `Send()` - Cenarios: success, error, timeout
- [ ] `SendWithProgress()` - Cenarios: streaming, error
- [ ] `Name()` - Cenarios: verificar retorno
- [ ] `IsAvailable()` - Cenarios: with/without gemini-cli
- [ ] `IsConfigured()` - Cenarios: with/without cookies
- [ ] `ValidateConfig()` - Cenarios: valid, missing fields

---

## Recomendacoes Tecnicas

### Setup de Infraestrutura de Testes

- [x] Coverage reports configurados (`go test -coverprofile`)
- [ ] Integrar coverage no CI/CD (GitHub Actions)
- [ ] Definir threshold minimo de 85% para PRs
- [ ] Adicionar coverage badge no README

**Configuracao sugerida para CI:**

```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: |
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out | grep total

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: coverage.out
    fail_ci_if_error: true
    threshold: 85%
```

### Boas Praticas

1. **Table-Driven Tests**: Usar para todas as funcoes com multiplos casos
2. **Test Helpers**: Criar helpers em `testutil` package para setup comum
3. **Mock Interfaces**: Usar interfaces para injecao de dependencias
4. **Temp Directories**: Usar `t.TempDir()` para testes de filesystem
5. **Parallel Tests**: Usar `t.Parallel()` quando possivel

### Estrutura de Arquivos de Teste

```
internal/
  core/
    scanner/
      scanner.go
      scanner_test.go      # Unit tests
      filesystem.go
      filesystem_test.go   # Unit tests
  platform/
    anthropic/
      client.go
      client_test.go       # Unit tests
      models.go
      models_test.go       # Unit tests
cmd/
  config.go
  config_test.go           # Unit + integration
  context.go
  context_test.go          # Unit + integration
test/
  e2e/                     # End-to-end tests
    cli_test.go
    context_integration_test.go
  fixtures/                # Test data
    sample-project/
```

### Ferramentas Sugeridas

| Ferramenta | Proposito | Uso |
|------------|-----------|-----|
| `testify/assert` | Assertions | Ja em uso |
| `testify/mock` | Mocking | Para interfaces |
| `gomock` | Alternativa mocking | Para interfaces complexas |
| `httptest` | HTTP mocking | Para API clients |
| `afero` | Filesystem mocking | Para scanner tests |

---

## Metricas de Acompanhamento

### Objetivos Semanais

| Semana | Cobertura Meta | Pacotes/Modulos Focus |
|--------|----------------|----------------------|
| 1 | 76% | platform/anthropic, platform/openai, platform/geminiapi |
| 2 | 80% | config/validator, cmd helpers (formatDuration, displayURL) |
| 3 | 84% | cmd/config, cmd/llm |
| 4 | 87% | platform/gemini/provider, cmd/context |
| 5 | 89% | internal/ui/wizard (helpers) |
| 6 | 91% | internal/ui/wizard (state transitions) |
| 7 | 93% | cmd/root, scanner edge cases |
| 8 | 95% | Polish, edge cases, integration tests |

### KPIs

| Metrica | Atual | Meta S4 | Meta Final |
|---------|-------|---------|------------|
| Cobertura Total | 73.7% | 87% | 90%+ |
| Pacotes >=90% | 7/22 | 15/22 | 20/22 |
| Funcoes 0% coverage | ~50 | 15 | <5 |
| Branch coverage | ~70% | 85% | 90% |

---

## Riscos e Mitigacoes

### Riscos Identificados

1. **Risco**: TUI testing complexity (Bubble Tea)
   - **Impacto**: Alto
   - **Mitigacao**: Extrair logica de negocio para funcoes puras testaveis; usar `tea.Batch` e message-based testing

2. **Risco**: External dependencies (gemini-cli binary)
   - **Impacto**: Medio
   - **Mitigacao**: Mock exec.Command; criar interface para executor

3. **Risco**: Tempo insuficiente para atingir 90%
   - **Impacto**: Medio
   - **Mitigacao**: Priorizar codigo critico; aceitar 85% em pacotes dificeis (ui/wizard)

4. **Risco**: Testes frageis devido a output formatting
   - **Impacto**: Baixo
   - **Mitigacao**: Testar estrutura, nao strings exatas; usar regex quando necessario

5. **Risco**: Race conditions em testes paralelos
   - **Impacto**: Medio
   - **Mitigacao**: Usar `t.Parallel()` com cuidado; isolar estado global (viper)

### Decisoes de Exclusao

Os seguintes arquivos/funcoes estao **excluidos** da meta de 90%:

| Item | Razao |
|------|-------|
| `main.go` | Entrypoint, testado via e2e |
| `internal/assets/embed.go` | Apenas declaracao de embed |
| `cmd/*.init()` | Registracao de comandos Cobra |
| `internal/ui/wizard.go` polling functions | Dificulta teste unitario, coberto por e2e |

---

## Conclusao

### Resumo Executivo

A cobertura atual de **73.7%** esta razoavelmente boa, mas longe da meta de **90%**. Os principais gaps estao em:

1. **cmd package (37%)** - Muitas funcoes de output e integracao nao testadas
2. **internal/ui/wizard.go (56%)** - Logica de TUI complexa
3. **platform providers (~70%)** - Metodos simples nao testados

### Proximos Passos Imediatos

1. **Esta semana**: Implementar testes para platform providers (quick wins)
   - `anthropic/models.go` - ValidModels, IsKnownModel
   - `openai/models.go` - ValidModels, IsKnownModel
   - `geminiapi/models.go` - ValidModels, IsKnownModel
   - Client methods: Name(), IsAvailable()

2. **Proxima semana**: Config validator e cmd helpers
   - `validatePath`, `validateGeminiModel`, `validateBrowserRefresh`
   - `formatDuration`, `displayURL`

3. **Configurar CI**: Adicionar coverage check no GitHub Actions

### Estimativa de Esforco

| Fase | Horas Estimadas | Cobertura Esperada |
|------|-----------------|-------------------|
| Fase 1 (Quick Wins) | 8-12h | +8% (82%) |
| Fase 2 (Critico) | 12-16h | +5% (87%) |
| Fase 3 (TUI) | 16-20h | +5% (92%) |
| Fase 4 (Polish) | 8-12h | +3% (95%) |
| **Total** | **44-60h** | **95%** |

---

*Documento gerado automaticamente. Ultima atualizacao: Janeiro 2026*
