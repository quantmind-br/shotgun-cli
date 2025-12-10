# Plano de Aumento de Cobertura de Testes

## Resumo Executivo

Este plano detalha a estratégia para aumentar a cobertura de testes de **77.4%** para **80%+** em cada pacote do projeto `shotgun-cli`.

### Cobertura Atual por Pacote

| Pacote | Cobertura Atual | Meta | Gap | Prioridade |
|--------|-----------------|------|-----|------------|
| `internal/ui` | **40.4%** | 80% | 39.6% | 🔴 CRÍTICA |
| `internal/platform/clipboard` | **50.0%** | 80% | 30.0% | 🔴 ALTA |
| `internal/platform/gemini` | **60.0%** | 80% | 20.0% | 🟡 MÉDIA |
| `internal/core/scanner` | **79.3%** | 80% | 0.7% | 🟢 BAIXA |
| `internal/ui/screens` | **80.0%** | 80% | 0% | ✅ OK |
| `internal/ui/styles` | **84.7%** | 80% | 0% | ✅ OK |
| `internal/core/ignore` | **89.6%** | 80% | 0% | ✅ OK |
| `internal/core/context` | **91.4%** | 80% | 0% | ✅ OK |
| `internal/ui/components` | **92.6%** | 80% | 0% | ✅ OK |
| `internal/core/template` | **92.9%** | 80% | 0% | ✅ OK |
| `internal/core/tokens` | **100.0%** | 80% | 0% | ✅ OK |
| `internal/utils` | **100.0%** | 80% | 0% | ✅ OK |

---

## Fase 1: internal/ui (40.4% → 80%)

### 1.1 Funções com 0% de Cobertura

As seguintes funções em `wizard.go` precisam de testes:

| Função | Linha | Complexidade | Estratégia de Teste |
|--------|-------|--------------|---------------------|
| `handleWindowResize` | 369 | Baixa | Simular `tea.WindowSizeMsg` |
| `handleScanError` | 471 | Baixa | Enviar `ScanErrorMsg` |
| `handleGenerationError` | 496 | Baixa | Enviar `GenerationErrorMsg` |
| `handleSendToGemini` | 507 | Média | Mock de estado + `GeminiSendMsg` |
| `handleGeminiProgress` | 541 | Baixa | Enviar `GeminiProgressMsg` |
| `handleGeminiComplete` | 550 | Média | Enviar `GeminiCompleteMsg` |
| `handleGeminiError` | 560 | Baixa | Enviar `GeminiErrorMsg` |
| `sendToGeminiCmd` | 569 | Alta | Mock de executor (difícil) |
| `handleTemplateMessage` | 604 | Baixa | Enviar `TemplateSelectedMsg` |
| `handleRescanRequest` | 645 | Baixa | Simular rescan request |
| `handleStepInput` | 759 | Alta | Testar cada step com input |
| `writeFile` | 833 | Média | Usar temp dir |
| `parseSize` | 849 | Baixa | Testar conversões de tamanho |
| `finalizeGeneration` | 1002 | Alta | Mock de estado completo |
| `validateContentSize` | 1023 | Média | Testar limites de tamanho |
| `saveGeneratedContent` | 1042 | Média | Usar temp dir |

### 1.2 Funções com Baixa Cobertura (<80%)

| Função | Cobertura | Meta | Testes Necessários |
|--------|-----------|------|-------------------|
| `View` | 8.0% | 80% | Testar renderização de cada step |
| `iterativeScanCmd` | 3.7% | 80% | Testar ciclo completo de scan |
| `iterativeGenerateCmd` | 4.5% | 80% | Testar ciclo completo de geração |
| `clipboardCopyCmd` | 33.3% | 80% | Testar sucesso e falha |
| `handleNextStep` | 57.1% | 80% | Testar transições de cada step |
| `getPrevStep` | 58.3% | 80% | Testar navegação reversa |
| `handleKeyPress` | 60.0% | 80% | Testar todas as teclas |
| `Update` | 61.3% | 80% | Testar todos os tipos de mensagem |

### 1.3 Testes a Implementar (`internal/ui/wizard_test.go`)

```go
// ============================================
// GRUPO 1: Testes de Window Resize
// ============================================

func TestWizardHandleWindowResize(t *testing.T) {
    // Testar que dimensões são atualizadas
    // Testar que screens filhos recebem resize
}

// ============================================
// GRUPO 2: Testes de Error Handling
// ============================================

func TestWizardHandleScanError(t *testing.T) {
    // Testar que erro é armazenado no model
    // Testar que progress é ocultado
    // Testar que View mostra erro
}

func TestWizardHandleGenerationError(t *testing.T) {
    // Similar ao scan error
}

// ============================================
// GRUPO 3: Testes de Integração Gemini
// ============================================

func TestWizardGeminiLifecycle(t *testing.T) {
    // Testar: handleSendToGemini → handleGeminiProgress → handleGeminiComplete
}

func TestWizardGeminiError(t *testing.T) {
    // Testar: handleSendToGemini → handleGeminiError
}

func TestWizardGeminiProgressUpdates(t *testing.T) {
    // Testar múltiplas atualizações de progresso
}

// ============================================
// GRUPO 4: Testes de Template Messages
// ============================================

func TestWizardHandleTemplateMessage(t *testing.T) {
    // Testar que template é selecionado
    // Testar transição para próximo step
}

// ============================================
// GRUPO 5: Testes de Step Input
// ============================================

func TestWizardHandleStepInput_FileSelection(t *testing.T) {
    // Testar input no step de seleção de arquivos
}

func TestWizardHandleStepInput_TemplateSelection(t *testing.T) {
    // Testar input no step de template
}

func TestWizardHandleStepInput_TaskInput(t *testing.T) {
    // Testar input no step de task
}

func TestWizardHandleStepInput_RulesInput(t *testing.T) {
    // Testar input no step de rules
}

func TestWizardHandleStepInput_Review(t *testing.T) {
    // Testar input no step de review
}

// ============================================
// GRUPO 6: Testes de File Operations
// ============================================

func TestWizardWriteFile(t *testing.T) {
    // Usar t.TempDir()
    // Testar escrita bem-sucedida
    // Testar erro de permissão (se possível)
}

func TestWizardParseSize(t *testing.T) {
    // Testar: "1KB", "1MB", "1GB", "invalid"
}

func TestWizardValidateContentSize(t *testing.T) {
    // Testar content dentro do limite
    // Testar content acima do limite
    // Testar com limite zero (sem validação)
}

func TestWizardSaveGeneratedContent(t *testing.T) {
    // Testar salvamento em temp dir
}

// ============================================
// GRUPO 7: Testes de Finalization
// ============================================

func TestWizardFinalizeGeneration(t *testing.T) {
    // Testar finalização com conteúdo válido
    // Testar que review screen é atualizado
}

// ============================================
// GRUPO 8: Testes de Rescan
// ============================================

func TestWizardHandleRescanRequest(t *testing.T) {
    // Testar que scan é reiniciado
    // Testar que estado é resetado
}

// ============================================
// GRUPO 9: Testes de View Rendering
// ============================================

func TestWizardViewFileSelectionStep(t *testing.T) {
    // Verificar elementos visuais do step 1
}

func TestWizardViewTemplateSelectionStep(t *testing.T) {
    // Verificar elementos visuais do step 2
}

func TestWizardViewTaskInputStep(t *testing.T) {
    // Verificar elementos visuais do step 3
}

func TestWizardViewRulesInputStep(t *testing.T) {
    // Verificar elementos visuais do step 4
}

func TestWizardViewReviewStep(t *testing.T) {
    // Verificar elementos visuais do step 5
}

func TestWizardViewWithError(t *testing.T) {
    // Verificar que erro é exibido
}

func TestWizardViewWithProgress(t *testing.T) {
    // Verificar que barra de progresso é exibida
}

// ============================================
// GRUPO 10: Testes de Iterative Commands
// ============================================

func TestWizardIterativeScanCmd(t *testing.T) {
    // Testar ciclo completo de scan iterativo
    // Verificar progress updates
    // Verificar completion
}

func TestWizardIterativeGenerateCmd(t *testing.T) {
    // Testar ciclo completo de geração iterativa
    // Verificar progress updates
    // Verificar completion
}
```

### 1.4 Estimativa de Esforço - Fase 1

| Grupo | Testes | Complexidade | Tempo Estimado |
|-------|--------|--------------|----------------|
| Window Resize | 1 | Baixa | 15 min |
| Error Handling | 2 | Baixa | 30 min |
| Gemini Lifecycle | 3 | Alta | 2 horas |
| Template Messages | 1 | Baixa | 15 min |
| Step Input | 5 | Média | 1.5 horas |
| File Operations | 4 | Média | 1 hora |
| Finalization | 1 | Média | 30 min |
| Rescan | 1 | Baixa | 15 min |
| View Rendering | 7 | Média | 2 horas |
| Iterative Commands | 2 | Alta | 2 horas |

**Total Estimado: ~10 horas**

---

## Fase 2: internal/platform/clipboard (50% → 80%)

### 2.1 Análise de Cobertura

```
clipboard.go:
  - ClipboardError.Error()   ✅ Testado
  - ClipboardError.Unwrap()  ✅ Testado
  - Copy()                   ❌ 0% (depende do sistema)
  - IsAvailable()            ✅ Testado (não falha se indisponível)
```

### 2.2 Estratégia de Teste

O desafio com `clipboard.Copy()` é que depende do sistema operacional. Estratégias:

**Opção A: Teste de Integração Condicional**
```go
func TestCopy(t *testing.T) {
    if !IsAvailable() {
        t.Skip("clipboard not available in this environment")
    }

    content := "test content"
    err := Copy(content)
    if err != nil {
        t.Errorf("Copy failed: %v", err)
    }
}

func TestCopyEmptyString(t *testing.T) {
    if !IsAvailable() {
        t.Skip("clipboard not available")
    }

    err := Copy("")
    // Verificar comportamento com string vazia
}

func TestCopyLargeContent(t *testing.T) {
    if !IsAvailable() {
        t.Skip("clipboard not available")
    }

    largeContent := strings.Repeat("x", 1024*1024) // 1MB
    err := Copy(largeContent)
    // Verificar se suporta conteúdo grande
}
```

**Opção B: Interface para Mock (Refatoração)**
```go
// clipboard.go
type Clipboard interface {
    Copy(content string) error
    IsAvailable() bool
}

type SystemClipboard struct{}

func (s *SystemClipboard) Copy(content string) error {
    return clipboard.WriteAll(content)
}

// Em testes, usar mock
type MockClipboard struct {
    CopyFunc func(string) error
}
```

### 2.3 Testes a Implementar

```go
// clipboard_test.go

func TestCopySuccess(t *testing.T) {
    if !IsAvailable() {
        t.Skip("clipboard not available")
    }

    tests := []struct {
        name    string
        content string
    }{
        {"simple text", "hello world"},
        {"empty string", ""},
        {"unicode", "こんにちは世界"},
        {"multiline", "line1\nline2\nline3"},
        {"special chars", "tab\there\nnewline"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Copy(tt.content)
            if err != nil {
                t.Errorf("Copy(%q) failed: %v", tt.content, err)
            }
        })
    }
}

func TestCopyErrorWrapping(t *testing.T) {
    // Testar que erros são wrappados em ClipboardError
    // Pode requerer mock ou condição de erro forçada
}
```

### 2.4 Estimativa de Esforço - Fase 2

| Tarefa | Tempo Estimado |
|--------|----------------|
| Testes de integração condicional | 30 min |
| Testes de edge cases | 30 min |
| **Total** | **1 hora** |

---

## Fase 3: internal/platform/gemini (60% → 80%)

### 3.1 Funções com Baixa/Zero Cobertura

| Função | Cobertura | Estratégia |
|--------|-----------|------------|
| `SendWithProgress` | 0% | Mock de comando externo |
| `Send` (casos de sucesso) | Parcial | Requer `geminiweb` instalado |

### 3.2 Análise do Código

O maior desafio é que `Send` e `SendWithProgress` executam o binário externo `geminiweb`. Opções:

**Opção A: Testes de Integração (se geminiweb disponível)**
```go
func TestSendWithProgress_Integration(t *testing.T) {
    if !IsAvailable() || !IsConfigured() {
        t.Skip("geminiweb not available or configured")
    }

    cfg := DefaultConfig()
    executor := NewExecutor(cfg)

    progressCh := make(chan string, 10)
    ctx := context.Background()

    result, err := executor.SendWithProgress(ctx, "Say hello", progressCh)

    // Verificar resultado
}
```

**Opção B: Mock do Executor (Refatoração Recomendada)**
```go
// Criar interface
type GeminiExecutor interface {
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress chan<- string) (*Result, error)
}

// Mock para testes
type MockExecutor struct {
    SendFunc             func(context.Context, string) (*Result, error)
    SendWithProgressFunc func(context.Context, string, chan<- string) (*Result, error)
}
```

### 3.3 Testes a Implementar

```go
// gemini_test.go

// ============================================
// Testes de SendWithProgress (estruturais)
// ============================================

func TestSendWithProgress_BinaryNotFound(t *testing.T) {
    cfg := Config{BinaryPath: "/nonexistent/path"}
    executor := NewExecutor(cfg)

    progress := make(chan string, 10)
    _, err := executor.SendWithProgress(context.Background(), "test", progress)

    if err == nil {
        t.Error("expected error for nonexistent binary")
    }
}

func TestSendWithProgress_ContextCancellation(t *testing.T) {
    if !IsAvailable() {
        t.Skip("geminiweb not available")
    }

    cfg := DefaultConfig()
    executor := NewExecutor(cfg)

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancelar imediatamente

    progress := make(chan string, 10)
    _, err := executor.SendWithProgress(ctx, "test", progress)

    if err == nil {
        t.Error("expected error on cancelled context")
    }
}

func TestSendWithProgress_ProgressChannelUpdates(t *testing.T) {
    // Testar que canal de progresso recebe atualizações
    // Pode requerer mock ou integração
}

// ============================================
// Testes Adicionais de Config
// ============================================

func TestConfigWithAllOptions(t *testing.T) {
    cfg := Config{
        Model:          "gemini-3.0-pro",
        Timeout:        60,
        BrowserRefresh: "never",
        Verbose:        true,
        BinaryPath:     "/custom/path",
    }

    executor := NewExecutor(cfg)
    args := executor.buildArgs()

    // Verificar todos os args são construídos corretamente
    expectedArgs := []string{"-m", "gemini-3.0-pro", "--browser-refresh", "never"}
    // ... validação
}

func TestConfigFindBinary_InPath(t *testing.T) {
    // Testar que FindBinary encontra binário no PATH
    // quando BinaryPath está vazio
}
```

### 3.4 Estimativa de Esforço - Fase 3

| Tarefa | Tempo Estimado |
|--------|----------------|
| Testes estruturais de SendWithProgress | 1 hora |
| Testes de context cancellation | 30 min |
| Testes adicionais de Config | 30 min |
| **Total** | **2 horas** |

---

## Fase 4: internal/core/scanner (79.3% → 80%)

### 4.1 Funções com 0% Cobertura

| Função | Linha | Motivo | Estratégia |
|--------|-------|--------|------------|
| `handleCountError` | 139 | Só chamada em erro | Forçar erro de permissão |
| `handleWalkError` | 226 | Só chamada em erro | Forçar erro de permissão |

### 4.2 Funções com Baixa Cobertura

| Função | Cobertura | Estratégia |
|--------|-----------|------------|
| `matchesIncludePatterns` | 18.2% | Testar mais padrões |
| `shouldSkipLargeFile` | 40.0% | Testar edge cases |
| `reportProgress` | 50.0% | Testar canal nil |
| `findParentNode` | 50.0% | Testar mais caminhos |
| `classifyIgnoreReason` | 50.0% | Testar todas as razões |

### 4.3 Testes a Implementar

```go
// scanner_test.go

func TestHandleCountError(t *testing.T) {
    // Criar diretório sem permissão de leitura
    tempDir := t.TempDir()
    noReadDir := filepath.Join(tempDir, "no-read")
    os.Mkdir(noReadDir, 0000)
    defer os.Chmod(noReadDir, 0755)

    scanner := NewFileSystemScanner()
    config := DefaultScanConfig()

    _, err := scanner.Scan(noReadDir, config)
    // Verificar handling de erro
}

func TestHandleWalkError(t *testing.T) {
    // Similar ao acima, mas durante walk
}

func TestMatchesIncludePatterns(t *testing.T) {
    tests := []struct {
        patterns []string
        path     string
        expected bool
    }{
        {[]string{"*.go"}, "main.go", true},
        {[]string{"*.go"}, "main.txt", false},
        {[]string{"src/**"}, "src/main.go", true},
        {[]string{"**/*.test.go"}, "pkg/foo.test.go", true},
        {[]string{}, "anything.go", true}, // Empty = include all
    }

    for _, tt := range tests {
        // Testar cada caso
    }
}

func TestShouldSkipLargeFile(t *testing.T) {
    tests := []struct {
        fileSize    int64
        maxFileSize int64
        expected    bool
    }{
        {100, 1000, false},   // Dentro do limite
        {1000, 1000, false},  // Exatamente no limite
        {1001, 1000, true},   // Acima do limite
        {100, 0, false},      // Sem limite (0 = ilimitado)
    }

    for _, tt := range tests {
        // Testar cada caso
    }
}

func TestReportProgress_NilChannel(t *testing.T) {
    // Verificar que não há panic com canal nil
}

func TestFindParentNode_DeepNesting(t *testing.T) {
    // Testar com estrutura profunda de diretórios
}

func TestClassifyIgnoreReason_AllReasons(t *testing.T) {
    // Testar cada tipo de IgnoreReason
}
```

### 4.4 Estimativa de Esforço - Fase 4

| Tarefa | Tempo Estimado |
|--------|----------------|
| Testes de error handling | 45 min |
| Testes de matchesIncludePatterns | 30 min |
| Testes de shouldSkipLargeFile | 15 min |
| Outros testes | 30 min |
| **Total** | **2 horas** |

---

## Fase 5: Pacotes Secundários

### 5.1 internal/ui/screens (80% - manter)

Funções com 0% que podem ser testadas:

```go
// review_test.go
func TestSetGeminiSending(t *testing.T) {}
func TestSetGeminiComplete(t *testing.T) {}
func TestSetGeminiError(t *testing.T) {}
func TestFormatDuration(t *testing.T) {}
func TestParseSize(t *testing.T) {}

// task_input_test.go
func TestSetWillSkipToReview(t *testing.T) {}
```

### 5.2 internal/ui/styles (84.7% - manter)

Funções com 0% que podem ser testadas:

```go
// theme_test.go
func TestRenderInfo(t *testing.T) {}
func TestRenderIgnoreIndicator(t *testing.T) {}
func TestRenderTokenStats(t *testing.T) {}
func TestRenderStepIndicator(t *testing.T) {}
func TestRenderSpinnerFrame(t *testing.T) {}
```

### 5.3 internal/ui/components (92.6% - manter)

Funções com 0%:

```go
// progress_test.go
func TestProgressModelInit(t *testing.T) {}
func TestProgressModelUpdateSpinner(t *testing.T) {}
```

### 5.4 Estimativa de Esforço - Fase 5

| Pacote | Tempo Estimado |
|--------|----------------|
| ui/screens | 1 hora |
| ui/styles | 45 min |
| ui/components | 30 min |
| **Total** | **2.25 horas** |

---

## Cronograma de Implementação

### Semana 1: Prioridades Críticas

| Dia | Tarefa | Meta |
|-----|--------|------|
| Dia 1 | Fase 1 - Grupos 1-4 (wizard basics) | +10% ui |
| Dia 2 | Fase 1 - Grupos 5-7 (wizard advanced) | +15% ui |
| Dia 3 | Fase 1 - Grupos 8-10 (wizard complete) | +15% ui → 80% |
| Dia 4 | Fase 2 (clipboard) + Fase 4 (scanner) | 80%+ ambos |
| Dia 5 | Fase 3 (gemini) | 80% gemini |

### Semana 2: Consolidação

| Dia | Tarefa | Meta |
|-----|--------|------|
| Dia 1 | Fase 5 - testes secundários | Manter 80%+ |
| Dia 2 | Revisão e ajustes | Cobertura total 80%+ |
| Dia 3 | Documentação de testes | - |

---

## Comandos de Verificação

```bash
# Verificar cobertura total
make coverage

# Verificar cobertura por pacote
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out

# Verificar cobertura de pacote específico
go test -coverprofile=coverage.out ./internal/ui/...
go tool cover -func=coverage.out | grep "internal/ui/"

# Gerar relatório HTML
go tool cover -html=coverage.out -o coverage.html

# Rodar testes com verbose
go test -v -cover ./internal/ui/...

# Rodar teste específico
go test -v -run TestWizardHandleWindowResize ./internal/ui/
```

---

## Métricas de Sucesso

### Critérios de Aceitação

| Métrica | Valor Atual | Meta |
|---------|-------------|------|
| Cobertura Total | 77.4% | ≥80% |
| `internal/ui` | 40.4% | ≥80% |
| `internal/platform/clipboard` | 50.0% | ≥80% |
| `internal/platform/gemini` | 60.0% | ≥80% |
| `internal/core/scanner` | 79.3% | ≥80% |
| Todos os outros pacotes | ≥80% | Manter ≥80% |

### Critérios de Qualidade

- [ ] Todos os testes passam (`make test`)
- [ ] Nenhum teste flaky (rodar 3x sem falha)
- [ ] Race condition detector passa (`make test-race`)
- [ ] Testes E2E passam (`make test-e2e`)
- [ ] Cobertura de cada pacote ≥80%

---

## Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| Testes de TUI são complexos | Alta | Médio | Usar helpers de teste do Bubble Tea |
| Clipboard indisponível em CI | Média | Baixo | Skip condicional |
| geminiweb não disponível | Alta | Médio | Testes estruturais + skip |
| Testes lentos | Média | Baixo | Paralelizar com t.Parallel() |
| Cobertura não aumenta | Baixa | Alto | Revisar funções não-testáveis |

---

## Anexo A: Helpers de Teste Recomendados

### Helper para Testes de TUI (wizard)

```go
// test_helpers_test.go

func setupWizardWithTree(t *testing.T) *WizardModel {
    t.Helper()
    wizard := NewWizard("/workspace", &scanner.ScanConfig{})
    wizard.fileTree = &scanner.FileNode{
        Name:  "root",
        Path:  "/workspace",
        IsDir: true,
    }
    return wizard
}

func setupWizardForReview(t *testing.T) *WizardModel {
    t.Helper()
    wizard := setupWizardWithTree(t)
    wizard.selectedFiles["main.go"] = true
    wizard.template = &template.Template{Name: "basic"}
    wizard.taskDesc = "Test task"
    wizard.step = StepReview
    wizard.review = screens.NewReview(
        wizard.selectedFiles,
        wizard.fileTree,
        wizard.template,
        wizard.taskDesc,
        "",
    )
    return wizard
}

func assertNoError(t *testing.T, wizard *WizardModel) {
    t.Helper()
    if wizard.error != nil {
        t.Fatalf("unexpected error in wizard: %v", wizard.error)
    }
}
```

### Helper para Testes com Temp Files

```go
func createTempDirWithFiles(t *testing.T, files map[string]string) string {
    t.Helper()
    dir := t.TempDir()

    for path, content := range files {
        fullPath := filepath.Join(dir, path)
        os.MkdirAll(filepath.Dir(fullPath), 0755)
        os.WriteFile(fullPath, []byte(content), 0644)
    }

    return dir
}
```

---

## Anexo B: Funções Não-Testáveis (Excluídas da Meta)

Algumas funções podem ser consideradas não-testáveis ou de baixo valor:

1. **Funções de logging puro** - Não afetam lógica
2. **Funções que apenas delegam** - Testadas indiretamente
3. **Código de inicialização one-time** - Difícil de isolar

Essas funções podem ser marcadas com `//nolint:` se necessário.
