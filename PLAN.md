# Plano de Correção: Busy-Loop no iterativeScanCmd

## Problema Identificado

A aplicação trava ao abrir em diretórios com muitos arquivos devido a um busy-loop na função `iterativeScanCmd` (`internal/ui/wizard.go:872-940`).

### Causa Raiz

```go
// Linha 935-936
default:
    // No progress yet, re-enqueue
    return m.iterativeScanCmd()()
```

**Fluxo problemático:**

1. `Init()` chama `scanDirectoryCmd()` → envia `startScanMsg`
2. `handleStartScan()` cria `scanState` e chama `iterativeScanCmd()`
3. `iterativeScanCmd()` inicia goroutine com `ScanWithProgress()`
4. `ScanWithProgress()` executa `countItems()` (primeira passagem) - **sem enviar progresso**
5. `iterativeScanCmd()` tenta ler do canal de progresso via `select`
6. Canal vazio → `default` case executa chamada recursiva **síncrona**
7. Loop infinito bloqueia o event loop do Bubble Tea

### Por que não há progresso durante `countItems()`?

- `countItems()` (`filesystem.go:107-141`) não envia nenhum progresso
- `reportProgress()` só é chamado em `walkAndBuild()` (segunda passagem)
- Progresso só é enviado a cada 100 itens (`current%100 == 0`)

## Análise de Soluções

### Solução A: Delay no default case (Mínima)
```go
default:
    time.Sleep(10 * time.Millisecond)
    return m.iterativeScanCmd()()
```

**Prós:** Simples, resolve o problema imediato
**Contras:** Latência artificial, não mostra progresso real

### Solução B: Usar tea.Tick (Idiomática Bubble Tea)
```go
default:
    return tea.Tick(10*time.Millisecond, func(t time.Time) tea.Msg {
        return pollScanProgressMsg{}
    })
```

**Prós:** Idiomático, não bloqueia
**Contras:** Requer novo tipo de mensagem e handler

### Solução C: Progresso durante countItems (Completa)
Adicionar envio de progresso durante a fase de contagem.

**Prós:** UX melhor, mostra "Counting files..."
**Contras:** Mais mudanças, afeta o scanner

### Solução Escolhida: A + C (Híbrida)

1. **Fix imediato (A):** Adicionar delay no default case
2. **Melhoria UX (C):** Enviar mensagem inicial durante countItems

## Plano de Implementação

### Fase 1: Fix do Busy-Loop (Crítico)

**Arquivo:** `internal/ui/wizard.go`

**Mudança 1.1:** Adicionar delay no default case de `iterativeScanCmd()`

```go
// Antes (linha 935-936):
default:
    // No progress yet, re-enqueue
    return m.iterativeScanCmd()()

// Depois:
default:
    // No progress yet, yield to event loop and re-check
    time.Sleep(10 * time.Millisecond)
    return m.iterativeScanCmd()()
```

### Fase 2: Melhoria de UX (Recomendado)

**Arquivo:** `internal/core/scanner/filesystem.go`

**Mudança 2.1:** Enviar progresso inicial antes de `countItems()`

```go
// Em ScanWithProgress(), antes de countItems():
if progress != nil {
    progress <- Progress{
        Current:   0,
        Total:     0,
        Stage:     "counting",
        Message:   "Counting files...",
        Timestamp: time.Now(),
    }
}
```

**Mudança 2.2:** Enviar progresso periódico durante `countItems()` (opcional)

Adicionar parâmetro de progress channel a `countItems()` e enviar atualizações a cada 500 arquivos.

### Fase 3: Testes

1. Testar em diretório pequeno (~100 arquivos)
2. Testar em diretório médio (~1000 arquivos)
3. Testar em diretório grande (~10000+ arquivos) - `~/dev/aistudio-build-proxy-all`
4. Verificar que a UI responde durante o scan
5. Verificar que Ctrl+C funciona durante o scan

## Ordem de Execução

| # | Tarefa | Arquivo | Risco |
|---|--------|---------|-------|
| 1 | Adicionar delay no default case | `wizard.go` | Baixo |
| 2 | Adicionar progresso "counting" | `filesystem.go` | Baixo |
| 3 | Testar manualmente | - | - |
| 4 | Executar testes unitários | - | - |

## Validação

```bash
# Compilar
make build

# Testar em diretório grande
cd ~/dev/aistudio-build-proxy-all && ~/dev/shotgun-cli/build/shotgun-cli

# Executar testes
make test
```

## Rollback

Se houver problemas, reverter as mudanças em:
- `internal/ui/wizard.go` (linha ~935)
- `internal/core/scanner/filesystem.go` (se Fase 2 aplicada)

## Arquivos Afetados

1. `internal/ui/wizard.go` - Fix do busy-loop
2. `internal/core/scanner/filesystem.go` - Melhoria de UX (opcional)

## Estimativa de Impacto

- **Linhas modificadas:** ~5-15
- **Risco de regressão:** Baixo
- **Testes afetados:** Nenhum (comportamento de polling interno)
