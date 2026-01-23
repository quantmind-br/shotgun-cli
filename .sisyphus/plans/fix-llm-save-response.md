# Fix LLM Save Response Configuration

## Context

### Original Request
O usuário reportou que a configuração `gemini.save-response: true` não está funcionando - as respostas da LLM não estão sendo salvas automaticamente.

### Interview Summary
**Key Findings**:
- A API LLM está funcionando corretamente (testado com sucesso)
- A configuração `gemini.save-response: true` está definida no arquivo de config
- O comando `context send` ignora completamente esta configuração

**Root Cause**:
O comando CLI `context send` (`cmd/send.go`) não lê a configuração `gemini.save-response`. Ele só salva quando a flag `-o` é explicitamente passada.

### Metis Review
**Identified Gaps** (addressed):
- Falta de chave `llm.save-response` para consistência com o namespace LLM
- O `cmd/send.go` precisa respeitar as configurações de save-response

---

## Work Objectives

### Core Objective
Fazer o comando `context send` respeitar a configuração `gemini.save-response` (e adicionar `llm.save-response` para consistência), salvando automaticamente as respostas quando configurado.

### Concrete Deliverables
- Nova chave de configuração `llm.save-response` em `internal/config/keys.go`
- Validação atualizada em `internal/config/validator.go`
- Default value em `cmd/root.go`
- Lógica de auto-save em `cmd/send.go`

### Definition of Done
- [ ] `shotgun-cli config set llm.save-response true` funciona
- [ ] `echo "test" | shotgun-cli context send` salva resposta automaticamente quando config está true
- [ ] Todos os testes passam

### Must Have
- Compatibilidade retroativa com `gemini.save-response`
- Auto-geração do nome do arquivo de saída quando não especificado

### Must NOT Have (Guardrails)
- Não quebrar a funcionalidade existente da flag `-o`
- Não alterar comportamento do TUI Wizard (já funciona)
- Não remover suporte a `gemini.save-response`

---

## Verification Strategy (MANDATORY)

### Test Decision
- **Infrastructure exists**: YES
- **User wants tests**: YES (Tests-after)
- **Framework**: go test

---

## Task Flow

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

## Parallelization

| Task | Depends On | Reason |
|------|------------|--------|
| 2 | 1 | Precisa da key definida |
| 3 | 1 | Precisa da key definida |
| 4 | 1, 2, 3 | Usa a key e validação |
| 5 | 4 | Verifica implementação |

---

## TODOs

- [x] 1. Adicionar KeyLLMSaveResponse em internal/config/keys.go

  **What to do**:
  - Adicionar `KeyLLMSaveResponse = "llm.save-response"` na seção LLM

  **Must NOT do**:
  - Alterar outras keys existentes

  **Parallelizable**: NO (base para outros)

  **References**:
  - `internal/config/keys.go:16-21` - Seção LLM onde adicionar a nova key
  - `internal/config/keys.go:30` - Exemplo de key similar: `KeyGeminiSaveResponse`

  **Acceptance Criteria**:
  - [ ] `grep "KeyLLMSaveResponse" internal/config/keys.go` → retorna a linha
  - [ ] `go build ./...` → compila sem erros

  **Commit**: NO (agrupa com 2, 3)

---

- [x] 2. Adicionar validação para llm.save-response em validator.go

  **What to do**:
  - Adicionar `KeyLLMSaveResponse` na lista de `ValidKeys()` (linha ~49)
  - Adicionar `KeyLLMSaveResponse` no switch case de `ValidateValue()` junto com outros booleans (linha ~72)
  - Adicionar `KeyLLMSaveResponse` no switch case de `ConvertValue()` junto com outros booleans (linha ~112)

  **Must NOT do**:
  - Criar nova função de validação (usar `validateBooleanValue` existente)

  **Parallelizable**: NO (depende de 1)

  **References**:
  - `internal/config/validator.go:48` - Onde adicionar na lista ValidKeys
  - `internal/config/validator.go:72` - Case de validação boolean
  - `internal/config/validator.go:112` - Case de conversão boolean

  **Acceptance Criteria**:
  - [ ] `go test ./internal/config/... -run TestValidateValue` → PASS
  - [ ] `go build ./...` → compila sem erros

  **Commit**: NO (agrupa com 1, 3)

---

- [x] 3. Adicionar default value para llm.save-response em cmd/root.go

  **What to do**:
  - Adicionar `viper.SetDefault(config.KeyLLMSaveResponse, true)` após linha 254

  **Must NOT do**:
  - Alterar default de `gemini.save-response`

  **Parallelizable**: NO (depende de 1)

  **References**:
  - `cmd/root.go:254` - Onde adicionar: após `viper.SetDefault(config.KeyLLMTimeout, 300)`
  - `cmd/root.go:262` - Referência: `viper.SetDefault(config.KeyGeminiSaveResponse, true)`

  **Acceptance Criteria**:
  - [ ] `grep "KeyLLMSaveResponse" cmd/root.go` → retorna a linha
  - [ ] `go build ./...` → compila sem erros

  **Commit**: NO (agrupa com 1, 2)

---

- [x] 4. Atualizar cmd/send.go para respeitar configuração save-response

  **What to do**:
  - Após linha 95 (onde pega `outputFile`), adicionar lógica:
    ```go
    // Check save-response config if no output file specified
    saveResponse := viper.GetBool(config.KeyLLMSaveResponse) || viper.GetBool(config.KeyGeminiSaveResponse)
    if outputFile == "" && saveResponse {
        // Auto-generate output filename
        timestamp := time.Now().Format("20060102-150405")
        outputFile = fmt.Sprintf("llm-response-%s.md", timestamp)
    }
    ```
  - Adicionar imports necessários: `"time"` (já existe), verificar se `config` está importado

  **Must NOT do**:
  - Alterar comportamento quando `-o` é explicitamente passado
  - Alterar lógica de output existente (linhas 137-144)

  **Parallelizable**: NO (depende de 1, 2, 3)

  **References**:
  - `cmd/send.go:95` - Onde `outputFile` é lido da flag
  - `cmd/send.go:137-144` - Lógica de save existente (não alterar)
  - `cmd/send.go:15` - Imports existentes (verificar se `config` está presente)

  **Acceptance Criteria**:
  - [ ] `go build ./...` → compila sem erros
  - [ ] Teste manual: `echo "test" | shotgun-cli context send` com `llm.save-response: true`
    - Deve criar arquivo `llm-response-YYYYMMDD-HHMMSS.md`
  - [ ] Teste manual: `echo "test" | shotgun-cli context send -o custom.md`
    - Deve criar arquivo `custom.md` (comportamento existente preservado)

  **Commit**: YES
  - Message: `fix(cli): respect save-response config in context send command`
  - Files: `internal/config/keys.go`, `internal/config/validator.go`, `cmd/root.go`, `cmd/send.go`
  - Pre-commit: `go test ./... && go build ./...`

---

- [x] 5. Executar testes completos e verificar

  **What to do**:
  - Executar `go test -race ./...`
  - Executar `golangci-lint run ./...`
  - Teste manual de integração

  **Must NOT do**:
  - Ignorar falhas de teste

  **Parallelizable**: NO (depende de 4)

  **References**:
  - `AGENTS.md` - Comandos de build e test

  **Acceptance Criteria**:
  - [ ] `go test -race ./...` → PASS (0 failures)
  - [ ] `golangci-lint run ./...` → 0 errors
  - [ ] Teste end-to-end:
    ```bash
    # Configurar
    shotgun-cli config set llm.save-response true
    
    # Testar auto-save
    echo "Hello, respond OK" | shotgun-cli context send
    # Deve mostrar "Response saved to: llm-response-*.md"
    
    # Verificar arquivo existe
    ls llm-response-*.md
    ```

  **Commit**: NO (se houver correções, fazer novo commit)

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 4 | `fix(cli): respect save-response config in context send command` | keys.go, validator.go, root.go, send.go | go test ./... |

---

## Success Criteria

### Verification Commands
```bash
# Build
go build ./...

# Tests
go test -race ./...

# Lint
golangci-lint run ./...

# Integration test
shotgun-cli config set llm.save-response true
echo "test" | shotgun-cli context send
# Expected: "Response saved to: llm-response-*.md"
```

### Final Checklist
- [ ] Nova config key `llm.save-response` funciona
- [ ] Compatibilidade com `gemini.save-response` mantida
- [ ] Flag `-o` continua funcionando
- [ ] Todos os testes passam
