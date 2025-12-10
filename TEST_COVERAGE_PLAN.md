# Plano de Expansão da Cobertura de Testes - Shotgun-CLI

## Objetivo
Aumentar a cobertura de testes para pelo menos 80% em todos os módulos do projeto.

## Estado Atual da Cobertura

### ✅ Áreas que já atingiram a meta (80%+)
- `internal/core/tokens` - 100%
- `internal/utils` - 100%
- `internal/core/template` - 92.9%
- `internal/core/context` - 91.4%
- `internal/core/ignore` - 89.6%
- `internal/ui/styles` - 84.7%
- `internal/ui/components` - 92.6%
- `internal/ui/screens` - 80%

### 🚨 Áreas que precisam de expansão

#### 1. `cmd/` - 44.7% → 80% (precisa +35.3%)
**Prioridade:** CRÍTICA
- `cmd/send.go` - 0% cobertura
- `cmd/root.go` - `launchTUIWizard` (0%), `Execute` (0%)
- `cmd/context.go` - `sendToGemini` (0%), `countFilesInTree` (0%), `collectAllSelections` (0%)

#### 2. `internal/platform/clipboard` - 50% → 80% (precisa +30%)
**Prioridade:** MÉDIA
- `clipboard.go` - `Copy` (0%)

#### 3. `internal/platform/gemini` - 60% → 80% (precisa +20%)
**Prioridade:** CRÍTICA
- `executor.go` - `Send` (10.3%), `SendWithProgress` (0%)
- `config.go` - `FindBinary` (43.8%)

#### 4. `internal/ui` - 40.4% → 80% (precisa +39.6%)
**Prioridade:** BAIXA
- Alguns componentes já estão acima de 80%

#### 5. `internal/core/scanner` - 79.3% → 80% (precisa +0.7%)
**Prioridade:** BAIXA
- Já está próximo da meta

## Plano de Implementação

### Fase 1: Cmd/Send (Prioridade CRÍTICA)
**Objetivo:** Criar `cmd/send_test.go`

**Testes a implementar:**
1. `TestRunContextSend_FromFile` - envio via arquivo
2. `TestRunContextSend_FromStdin` - envio via stdin
3. `TestRunContextSend_Validation` - validações de entrada
4. `TestRunContextSend_GeminiIntegration` - integração com gemini
5. `TestRunContextSend_OutputHandling` - escrita de arquivo/stdout
6. `TestRunContextSend_ErrorHandling` - tratamento de erros

### Fase 2: Gemini Executor (Prioridade CRÍTICA)
**Objetivo:** Expandir `internal/platform/gemini/gemini_test.go`

**Testes a implementar:**
1. `TestExecutor_Send_Success` - fluxo completo de sucesso
2. `TestExecutor_Send_Timeout` - cenários de timeout
3. `TestExecutor_Send_Error` - tratamento de erros
4. `TestExecutor_SendWithProgress` - progresso do envio
5. `TestConfig_FindBinary_EdgeCases` - casos extremos para FindBinary

### Fase 3: Clipboard Integration (Prioridade MÉDIA)
**Objetivo:** Expandir `internal/platform/clipboard/clipboard_test.go`

**Testes a implementar:**
1. `TestCopy_Success` - cópia bem-sucedida
2. `TestCopy_Failure` - falhas de cópia
3. `TestIsAvailable` - detecção de disponibilidade

### Fase 4: Root Command (Prioridade MÉDIA)
**Objetivo:** Expandir `cmd/root_test.go`

**Testes a implementar:**
1. `TestLaunchTUIWizard` - inicialização do wizard
2. `TestExecute_Flow` - fluxo de execução

### Fase 5: Context Command (Prioridade MÉDIA)
**Objetivo:** Expandir `cmd/context_test.go`

**Testes a implementar:**
1. `TestSendToGemini` - envio para gemini
2. `TestCountFilesInTree` - contagem de arquivos
3. `TestCollectAllSelections` - coleta de seleções

## Comandos de Validação

Após implementar os testes, executar:

```bash
# Verificar cobertura geral
make coverage

# Testes específicos por módulo
go test ./cmd -v -cover
go test ./internal/platform/gemini -v -cover
go test ./internal/platform/clipboard -v -cover
go test ./internal/ui -v -cover

# Meta: Todos os módulos devem ter 80%+ cobertura
```

## Critérios de Aceitação

1. **Cobertura geral:** ≥ 80%
2. **Cobertura cmd/:** ≥ 75%
3. **Cobertura platform/:** ≥ 80%
4. **Cobertura ui/:** ≥ 70%
5. **Todos os testes passando:** `make test`
6. **Sem regressions:** E2E tests passando

## Timeline

- **Fase 1-2:** Crítico - Implementação imediata
- **Fase 3-5:** Média - Implementação sequencial
- **Validação:** Após cada fase