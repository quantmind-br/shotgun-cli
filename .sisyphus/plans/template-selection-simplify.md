# Simplify Template Selection Screen

## Context

### Original Request
Modificar a tela de seleção de template para remover o preview dos templates. Não exibir descrição ou qualquer outra informação. A tela deverá conter apenas a seleção de templates e deverá estar otimizada para exibição em janelas de terminal pequenas.

### Interview Summary
**Key Discussions**:
- Manter o modal de preview completo (tecla 'v') - usuário deseja manter essa funcionalidade
- Menu estilizado alinhado à esquerda, consistente com o padrão existente
- Atualizar testes para refletir a nova interface simplificada

**Research Findings**:
- Arquivo principal: `internal/ui/screens/template_selection.go` (529 linhas)
- Arquivo de testes: `internal/ui/screens/template_selection_test.go` (640 linhas)
- Layout atual usa duas colunas: lista (35 cols) + detalhes (variável)
- Método `renderTemplateDetails()` precisa ser removido
- Função `countNonEmptyLines()` será removida (usada apenas por `renderTemplateDetails()`)
- Modal e todos os métodos relacionados devem ser MANTIDOS

### Metis Review
**Identified Gaps** (addressed):
- Footer hint: Manter "v: View full" pois o modal continua funcional
- Truncamento de nomes longos: Aplicar truncamento com base na largura disponível
- Lista vazia: Já tratado pelo `checkEarlyReturns()` - sem mudanças necessárias

---

## Work Objectives

### Core Objective
Simplificar a tela de seleção de template removendo o painel de detalhes lateral, mantendo apenas a lista de templates e o modal de preview completo.

### Concrete Deliverables
- `internal/ui/screens/template_selection.go` modificado com layout simplificado
- `internal/ui/screens/template_selection_test.go` atualizado para nova interface

### Definition of Done
- [x] `go test -race ./internal/ui/screens/...` passa sem falhas
- [x] `golangci-lint run ./internal/ui/screens/...` sem erros
- [x] Tela renderiza apenas lista de templates (sem painel de detalhes)
- [x] Modal de preview (tecla 'v') funciona corretamente
- [x] Interface funciona em terminais pequenos (40x10)

### Must Have
- Lista de templates com estilo atual (cursor, checkmark para selecionado)
- Modal de preview completo via tecla 'v'
- Footer com atalhos de teclado
- Scroll da lista quando há muitos templates
- Indicadores de scroll (↑ more above, ↓ more below)

### Must NOT Have (Guardrails)
- Painel de detalhes/descrição lateral
- Preview inline (apenas via modal)
- Mudanças em keybindings ou comportamento de navegação
- Mudanças no wizard flow ou outras telas
- Mudanças na lógica de carregamento/seleção de templates
- Novas dependências ou utilitários compartilhados

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES
- **User wants tests**: YES (update existing)
- **Framework**: go test

### Approach
Atualizar testes existentes para refletir a nova interface. Remover testes que verificam funcionalidades removidas (detalhes, required vars inline).

---

## Task Flow

```
Task 1 (Simplify View) → Task 2 (Remove Dead Code) → Task 3 (Update Tests)
```

## Parallelization

| Task | Depends On | Reason |
|------|------------|--------|
| 1 | - | Primeira mudança no código |
| 2 | 1 | Remove código que foi desvinculado na Task 1 |
| 3 | 2 | Testes devem refletir o estado final do código |

---

## TODOs

- [x] 1. Simplify View() method to render only template list

  **What to do**:
  - Remove two-column layout from `View()` method
  - Remove call to `renderTemplateDetails()`
  - Render `renderTemplateList()` directly without border box
  - Maintain header, list content, and footer structure
  - Keep all modal rendering logic unchanged (`showingFullPreview` branch)

  **Must NOT do**:
  - Change `renderTemplateList()` internal logic
  - Modify modal behavior or keybindings
  - Change footer hints (keep "v: View full")

  **Parallelizable**: NO (first task)

  **References**:

  **Pattern References**:
  - `internal/ui/screens/template_selection.go:133-184` - Current `View()` method to modify
  - `internal/ui/screens/template_selection.go:207-239` - `renderTemplateList()` to keep unchanged
  - `internal/ui/screens/template_selection.go:378-404` - `renderFooter()` to keep unchanged

  **Test References**:
  - `internal/ui/screens/template_selection_test.go:277-296` - `TestTemplateSelectionViewWithTemplates` needs update

  **Acceptance Criteria**:
  - [ ] `View()` returns string without "Description", "Required Variables", or "Preview" sections
  - [ ] `View()` contains template names from the list
  - [ ] Modal still renders when `showingFullPreview` is true
  - [ ] Manual: `go test -v -run TestTemplateSelectionView ./internal/ui/screens/...` → shows which tests need updating

  **Commit**: NO (groups with 2)

---

- [x] 2. Remove dead code: renderTemplateDetails and countNonEmptyLines

  **What to do**:
  - Remove `renderTemplateDetails()` method (lines 289-376)
  - Remove `countNonEmptyLines()` function (lines 472-480)
  - Verify no other code references these functions

  **Must NOT do**:
  - Remove any modal-related methods (`renderFullPreviewModal`, scroll methods, etc.)
  - Remove `showingFullPreview` or `previewScrollY` fields

  **Parallelizable**: NO (depends on Task 1)

  **References**:

  **Pattern References**:
  - `internal/ui/screens/template_selection.go:289-376` - `renderTemplateDetails()` to remove
  - `internal/ui/screens/template_selection.go:472-480` - `countNonEmptyLines()` to remove

  **Verification Steps**:
  - [ ] `grep -n "renderTemplateDetails" internal/ui/screens/template_selection.go` → no results
  - [ ] `grep -n "countNonEmptyLines" internal/ui/screens/template_selection.go` → no results
  - [ ] `go build ./internal/ui/screens/...` → compiles successfully

  **Acceptance Criteria**:
  - [ ] `go build ./internal/ui/screens/...` → SUCCESS (no compilation errors)
  - [ ] Methods `renderTemplateDetails` and `countNonEmptyLines` não existem no arquivo
  - [ ] All modal methods still present: `renderFullPreviewModal`, `handleModalKeyPress`, scroll methods

  **Commit**: YES
  - Message: `refactor(ui): simplify template selection screen layout`
  - Files: `internal/ui/screens/template_selection.go`
  - Pre-commit: `go build ./internal/ui/screens/...`

---

- [x] 3. Update tests to reflect simplified interface

  **What to do**:
  - Update `TestTemplateSelectionViewWithTemplates`: Remove assertions for "Description" and "desc1"
  - Update or remove `TestTemplateSelectionViewWithRequiredVars`: Remove assertions for "Required Variables"
  - Remove `TestCountNonEmptyLines`: Function no longer exists
  - Keep all modal-related tests unchanged
  - Keep all navigation/selection tests unchanged

  **Must NOT do**:
  - Remove or modify modal tests (`TestTemplateModalState`, `TestTemplateModalScrolling`, etc.)
  - Change test structure for navigation tests
  - Add new test dependencies

  **Parallelizable**: NO (depends on Task 2)

  **References**:

  **Test References**:
  - `internal/ui/screens/template_selection_test.go:277-296` - `TestTemplateSelectionViewWithTemplates` to update
  - `internal/ui/screens/template_selection_test.go:298-318` - `TestTemplateSelectionViewWithRequiredVars` to remove
  - `internal/ui/screens/template_selection_test.go:604-622` - `TestCountNonEmptyLines` to remove

  **Tests to Keep Unchanged**:
  - `internal/ui/screens/template_selection_test.go:394-456` - Modal state tests
  - `internal/ui/screens/template_selection_test.go:458-518` - Modal scrolling tests
  - `internal/ui/screens/template_selection_test.go:520-567` - Modal rendering tests

  **Acceptance Criteria**:
  - [ ] `go test -race -v ./internal/ui/screens/...` → ALL PASS
  - [ ] Test `TestTemplateSelectionViewWithTemplates` verifies only template names appear
  - [ ] Test `TestTemplateSelectionViewWithRequiredVars` removed
  - [ ] Test `TestCountNonEmptyLines` removed
  - [ ] All modal tests still present and passing

  **Commit**: YES
  - Message: `test(ui): update template selection tests for simplified layout`
  - Files: `internal/ui/screens/template_selection_test.go`
  - Pre-commit: `go test -race ./internal/ui/screens/...`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 2 | `refactor(ui): simplify template selection screen layout` | template_selection.go | `go build ./internal/ui/screens/...` |
| 3 | `test(ui): update template selection tests for simplified layout` | template_selection_test.go | `go test -race ./internal/ui/screens/...` |

---

## Success Criteria

### Verification Commands
```bash
go test -race -v ./internal/ui/screens/...  # All tests pass
golangci-lint run ./internal/ui/screens/...  # No linting errors
```

### Final Checklist
- [ ] All "Must Have" present (lista, modal, footer, scroll)
- [ ] All "Must NOT Have" absent (detalhes, preview inline, mudanças de behavior)
- [ ] All tests pass
- [ ] Code compiles without errors
