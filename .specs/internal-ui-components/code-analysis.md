# Análise de Código — internal/ui/components

| Campo            | Valor                                                        |
|------------------|--------------------------------------------------------------|
| **Nome do Módulo**   | `components`                                                 |
| **Caminho do Pacote** | `github.com/quantmind-br/shotgun-cli/internal/ui/components` |
| **Nível de Detalhe** | Detalhado (`detalhado`)                                      |
| **Arquivos Analisados** | `config_field.go`, `config_field_test.go`, `config_select.go`, `config_select_test.go`, `config_toggle.go`, `config_toggle_test.go`, `progress.go`, `progress_test.go`, `tree.go`, `tree_test.go` |
| **Dependências Internas** | `internal/config` (`ConfigMetadata`, `ValidateValue`, `ConfigType`), `internal/ui/styles` (tema e cores), `internal/core/scanner` (`FileNode`), `internal/core/tokens` (`FormatTokens`) |
| **Dependências Externas** | `github.com/charmbracelet/bubbles/textinput`, `github.com/charmbracelet/bubbles/spinner`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss` |
| **Padrão de Projeto** | Bubble Tea MVU (Model-View-Update) com componentes reutilizáveis |

---

## 1. Visão Geral

O módulo `internal/ui/components` é a camada de **apresentação TUI (Terminal User Interface)** do shotgun-cli. Ele implementa componentes de interface baseados no framework [Bubble Tea](https://github.com/charmbracelet/bubbletea), seguindo o padrão MVU (Model-View-Update).

O módulo fornece **5 componentes principais**, todos estruturados como structs com métodos `Init()`, `Update()`, `View()`:

1. **ConfigFieldModel** — Campo de entrada de texto livre (string, int, path, url, size, timeout)
2. **ConfigSelectModel** — Seletor de opções pré-definidas (enum)
3. **ConfigToggleModel** — Interruptor booleano (true/false)
4. **ProgressModel** — Modal de progresso com spinner e barra de progresso
5. **FileTreeModel** — Árvore de arquivos com navegação, seleção e filtro fuzzy

Os três primeiros componentes (`ConfigField`, `ConfigSelect`, `ConfigToggle`) são **widgets de configuração** usados no wizard de geração de configuração. Eles compartilham uma interface implícita: todos possuem `Init()`, `Focus()`, `Blur()`, `View()`, `Value()`, `SetValue()`, `IsModified()`, `IsFocused()`, `SetWidth()`, e `Metadata()`.

O componente `ProgressModel` é independente — exibe feedback visual durante operações longas (scan, geração de contexto, envio à LLM).

O componente `FileTreeModel` é o mais complexo — implementa virtualização, navegação por teclado, seleção hierárquica, filtro fuzzy, e gerenciamento de estados parciais de seleção.

---

## 2. Análise por Arquivo

### 2.1 `config_field.go` — Campo de Texto Livre

**Responsabilidade:** Renderiza um campo de entrada de texto para configurações de tipo livre (string, int, size, path, url, timeout).

**Estrutura (`ConfigFieldModel`):**
```go
type ConfigFieldModel struct {
    metadata    config.ConfigMetadata  // Metadados da configuração
    value       string                 // Valor atual como string
    input       textinput.Model        // Wrapper do bubble:textinput
    focused     bool                   // Status de foco
    width       int                    // Largura do container
    err         error                  // Erro de validação
    modified    bool                   // Indicador de modificação
    placeholder string                 // Texto de placeholder
}
```

**Dependências externas:** `textinput.Model` do pacote `bubbles/textinput` — este é o **único componente que encapsula outro bubble** (os outros são auto-contidos).

**Fluxo de inicialização (`NewConfigField`):**
1. Cria `textinput.Model` com limite de 256 caracteres e largura 40
2. Aplica estilos: prompt → cor primária, texto → cor de texto, placeholder → estilo definido, cursor → cor primária
3. Define o placeholder com base no tipo da configuração (`getPlaceholder`)
4. Define o valor inicial via `currentValue`

**Placeholder por tipo (`getPlaceholder`):**
| Tipo | Placeholder |
|------|------------|
| `TypeSize` | "e.g., 10MB, 500KB" |
| `TypePath` | "/path/to/file" |
| `TypeURL` | "https://api.example.com" |
| `TypeInt` | "Enter a number" |
| `TypeTimeout` | "Timeout in seconds" |
| `default` | "" (vazio) |

**Fluxo de atualização (`Update`):**
- Se **não focado**: retorna sem efeito
- Se **focado**: delega para `m.input.Update(msg)` do textinput e depois:
  - Compara o novo valor com `m.value`
  - Se diferente: marca `modified = true` e valida via `config.ValidateValue(m.metadata.Key, newValue)`
  - O erro de validação é armazenado em `m.err`

**Fluxo de renderização (`View`):**
```
[Label] (cor primária se focado)
[Descrição] (itálico, cor murcha)
[Input] (borda focada ou murcha, largura ajustada)
[Mensagem de erro] OU [Mensagem de sucesso]
```

**Validação:** A validação é delegada a `config.ValidateValue(key, value)`, que não está no escopo deste módulo mas é referenciada no código.

**Métodos de controle:**
- `Focus()` / `Blur()` — controlam estado de foco e delegam ao textinput
- `SetWidth(width)` — ajusta largura do input (range: 20-60, subtrai 4 do container)
- `Value()` / `SetValue(value)` — get/set com reset de `modified` e `err`
- `Error()` / `IsModified()` / `IsValid()` / `Metadata()` — accessors

**Funções exportadas:** `NewConfigField`
**Funções privadas:** `getPlaceholder`
**Métodos:** `Init`, `Update`, `Focus`, `Blur`, `IsFocused`, `SetWidth`, `Value`, `SetValue`, `Error`, `IsModified`, `IsValid`, `Metadata`, `View`

---

### 2.2 `config_select.go` — Seletor Enum

**Responsabilidade:** Renderiza um seletor de lista suspensa (dropdown) para configurações de tipo `enum`.

**Estrutura (`ConfigSelectModel`):**
```go
type ConfigSelectModel struct {
    metadata config.ConfigMetadata
    options  []string        // Opções disponíveis (de metadata.EnumOptions ou [currentValue])
    selected int             // Índice da opção selecionada
    focused  bool            // Status de foco
    width    int             // Largura do container
    modified bool            // Indicador de modificação
    expanded bool            // Dropdown aberto/fechado
}
```

**Fluxo de inicialização (`NewConfigSelect`):**
1. Usa `metadata.EnumOptions` como opções; se vazio, usa `[currentValue]`
2. Encontra o índice da `currentValue` nas opções
3. Se não encontrada, seleciona índice 0 (primeira opção)

**Fluxo de atualização (`Update`) — Máquina de estados de teclas:**

| Tecla (quando focado) | Ação |
|-----------------------|------|
| `enter` / ` ` (espaço) | Toggle `expanded` (abre/fecha dropdown) |
| `up` / `k` | Move cursor para cima (se `expanded`); wrap-around |
| `down` / `j` | Move cursor para baixo (se `expanded`); wrap-around |
| `esc` | Fecha dropdown se aberto |
| `tab` | Se `expanded`: fecha dropdown. Se não: cicla para próxima opção |

**Notas de implementação:**
- A navegação com `up/down` só funciona quando o dropdown está **expandido**
- O `tab` serve como atalho de ciclo rápido quando **colapsado**
- `init()` retorna `nil` (nenhum comando necessário)

**Fluxo de renderização (`View`):**
```
[Label] (cor primária se focado)
[Descrição] (itálico, cor murcha)
[Borda focada/murcha]
  ┌──────────────────────┐
  │ [Opção atual] ▼      │ (colapsado)
  │   (Enter para expandir) │ (hint quando focado)
  │                      │
  │ ▶ Opção A            │ (expandido — opção selecionada)
  │   Opção B            │
  │   Opção C            │
  └──────────────────────┘
```

**Métodos:** `Init`, `Update`, `Focus`, `Blur`, `IsFocused`, `SetWidth`, `Value`, `SetValue`, `IsModified`, `Metadata`, `View`
**Métodos privados:** `renderSelect`

---

### 2.3 `config_toggle.go` — Interruptor Booleano

**Responsabilidade:** Renderiza um interruptor booleano [ENABLED]/[DISABLED] para configurações de tipo `bool`.

**Estrutura (`ConfigToggleModel`):**
```go
type ConfigToggleModel struct {
    metadata config.ConfigMetadata
    value    bool            // Valor booleano
    focused  bool            // Status de foco
    width    int             // Largura do container
    modified bool            // Indicador de modificação
}
```

**Fluxo de inicialização (`NewConfigToggle`):**
- Simples: apenas armazena `metadata` e `currentValue`

**Fluxo de atualização (`Update`) — Máquina de estados de teclas:**

| Tecla (quando focado) | Ação |
|-----------------------|------|
| `enter` / ` ` / `tab` | Inverte `value` (toggle) |
| `y` / `Y` | Define `value = true` (apenas se já for false) |
| `n` / `N` | Define `value = false` (apenas se já for true) |

**Fluxo de renderização (`View`):**
```
[Label] (cor primária se focado)
[Descrição] (itálico, cor murcha)
[Borda focada/murcha]
  ┌──────────────────────┐
  │ [ENABLED] ou [DISABLED] │
  │   (Space/Enter to toggle) │ (hint quando focado)
  └──────────────────────┘
```

**Estilo visual:**
- `ENABLED`: texto verde brilhante (SuccessColor, bold) entre colchetes murchos
- `DISABLED`: texto cinza murch entre colchetes murchos

**Métodos:** `Init`, `Update`, `Focus`, `Blur`, `IsFocused`, `SetWidth`, `Value`, `SetValue`, `IsModified`, `Metadata`, `View`

---

### 2.4 `progress.go` — Modal de Progresso

**Responsabilidade:** Exibe um modal de progresso centralizado na tela com spinner animado, barra de progresso e informações de estágio.

**Estrutura (`ProgressModel`):**
```go
type ProgressModel struct {
    current int64            // Progresso atual
    total   int64            // Total (0 = sem limite, <0 = streaming)
    stage   string           // Nome do estágio atual
    message string           // Mensagem descritiva
    visible bool             // Visibilidade do modal
    width   int              // Largura da tela
    height  int              // Altura da tela
    spinner spinner.Model    // Spinner animado
}
```

**Modos de operação (definidos por `total`):**

| `total` | Significado | Renderização |
|---------|-------------|--------------|
| `total > 0` | Progresso conhecido | Barra de progresso + porcentagem + contagem |
| `total == 0` | Mensagem simples | Spinner + mensagem |
| `total < 0` | Streaming / contador | Spinner + contador ("X files scanned...") |

**Fluxo de inicialização (`NewProgress`):**
1. Cria spinner do tipo `spinner.Dot`
2. Estiliza com cor primária
3. `visible = false` (oculto por padrão)

**Métodos de controle:**
- `Update(current, total, stage, message)` — define estado completo e marca visível
- `UpdateMessage(stage, message)` — redefine current/total para 0, define novo estágio (para troca de fase)
- `Hide()` / `Show()` — controle de visibilidade
- `IsVisible()` — query de visibilidade
- `SetSize(width, height)` — define dimensões da tela
- `GetProgress()` — retorna tuple (current, total, stage, message)

**Fluxo de renderização (`View`):**
```
╭──────────────────────────────╮
│         Processing           │  ← Título
│                              │  ← Espaço vazio
│   <stage>                    │  ← Estágio
│   [████████░░░░]             │  ← Barra de progresso
│   50.0% (50/100)             │  ← Estatística
╰──────────────────────────────╯
```

**Dimensões:** Modal fixo de 60x8, mas ajustado à largura da tela (mínimo 30, máximo `width - 4`).

**Auxiliares de formatação:**
- `centerLine(line)` — centraliza horizontalmente no container
- `padCenter(text, width)` — centraliza e limita largura
- `truncate(text, maxLen)` — truncamento com "..." se exceder
- `visualWidth(text)` — largura visual (delegado a `lipgloss.Width`)
- `renderProgressBar(width)` — barra com █ e espaços vazios

**Método não implementado como tea.Cmd:** `UpdateSpinner(msg tea.Msg)` — atualiza o spinner manualmente. Note que `ProgressModel` não implementa diretamente `Update(tea.Msg)` do bubbletea — ele é atualizado proceduralmente via `Update()`, `UpdateMessage()`, etc. Isso significa que o componente é **controlado imperativamente** pelo chamador, não pelo loop do Bubble Tea.

---

### 2.5 `usage_bar.go` (funções embutidas em progress.go) — Barra de Uso de Contexto

**Nota:** `UsageBar` não é um arquivo separado — está embutido em `progress.go` (linhas iniciais do arquivo, antes de `ProgressModel`).

**Estrutura (`UsageBar`):**
```go
type UsageBar struct {
    CurrentBytes int64
    MaxBytes     int64
    MaxBytesStr  string
    TotalTokens  int
    Width        int
}
```

**Fluxo de renderização (`View`):**
1. Calcula porcentagem de uso
2. Determina ícone e cor: ✅ (≤80%), ⚠️ (>80%), ⛔ (>100%)
3. Renderiza barra com █ (preenchido) e ░ (vazio)
4. Exibe: ícone + tamanho + limite + tokens + barra

**Modo sem limite:** Quando `MaxBytes == 0`, exibe tamanho atual e tokens sem barra.

**Helper:** `formatSizeHelper(bytes)` — formata bytes para KB/MB/GB/TB (base 1024).

---

### 2.6 `tree.go` — Árvore de Arquivos

**Responsabilidade:** Renderiza uma árvore de arquivos interativa com navegação por teclado, seleção hierárquica, filtro fuzzy, scroll, e indicadores de estado de seleção parcial.

**Estrutura (`FileTreeModel`):**
```go
type FileTreeModel struct {
    tree            *scanner.FileNode      // Raiz da árvore de arquivos
    cursor          int                    // Índice do item no cursor
    selections      map[string]bool        // Caminhos selecionados
    selectionStates map[string]SelectionState  // Estados de seleção (cache)
    showIgnored     bool                   // Mostrar arquivos ignorados
    filter          string                 // Texto de filtro
    filterMatches   map[string]bool        // Caminhos que correspondem ao filtro
    expanded        map[string]bool        // Diretórios expandidos
    width           int                    // Largura da tela
    height          int                    // Altura da tela
    visibleItems    []treeItem             // Itens visíveis na viewport
    topIndex        int                    // Índice do primeiro item visível
    lastFilter      string                 // Cache: último filtro computado
    filterCacheValid bool                  // Cache válido?
}
```

**Estrutura auxiliar (`treeItem`):**
```go
type treeItem struct {
    node    *scanner.FileNode  // Nó do scanner
    path    string             // Caminho completo
    depth   int                // Profundidade na árvore
    isLast  bool               // É o último filho do irmão?
    hasNext []bool             // [bool] para cada nível de profundidade
}
```

**Fluxo de inicialização (`NewFileTree`):**
1. Cria mapas vazios para selections, selectionStates, expanded
2. Expande automaticamente o nó raiz
3. Copia seleções iniciais do parâmetro
4. Reconstrói itens visíveis (`rebuildVisibleItems`)
5. Recalcula estados de seleção (`recomputeSelectionStates`)

**Navegação (métodos imperativos, não via bubble tea Update):**

| Método | Ação |
|--------|------|
| `MoveUp()` | Move cursor para cima (se > 0) |
| `MoveDown()` | Move cursor para baixo (se < total-1) |
| `ExpandNode()` | Expande diretório no cursor |
| `CollapseNode()` | Colapsa diretório no cursor |
| `ToggleSelection()` | Seleciona/deseleciona nó atual (arquivo: toggle; diretório: toggle recursivo) |
| `SelectAllVisible()` | Seleciona todos os arquivos visíveis |
| `DeselectAllVisible()` | Deseleciona todos os arquivos visíveis |
| `ToggleShowIgnored()` | Alterna visibilidade de ignorados |
| `SetFilter(filter)` | Define filtro fuzzy |
| `GetFilter()` / `ClearFilter()` | Query/limpeza de filtro |

**Renderização do cabeçalho do item (`renderTreeItem`):**
```
[Prefixo de árvore] [Checkbox] [Ícone Dir] [Nome] [Status Ignore] [Tamanho]
```

| Parte | Detalhes |
|-------|----------|
| **Prefixo** | `│  ` / `   ` por nível + `└──` / `├──` |
| **Checkbox** | `[ ]` / `[✓]` para arquivos; vazio para diretórios |
| **Ícone dir** | 📂 (expandido) / 📁 (colapsado) |
| **Nome** | Nome base + `/` para diretórios; estilizado com cor do estado de seleção |
| **Ignore status** | Indicador de gitignore/custom-ignore (via `styles.RenderIgnoreIndicator`) |
| **Tamanho** | Tamanho formatado em KB/MB (para arquivos não vazios) |
| **Cursor** | `▶ ` em cor primária + fundo selecionado |

**Filtro fuzzy (`computeFilterMatches`):**
1. Verifica cache: se o filtro não mudou e o cache é válido, retorna imediatamente
2. Para cada nó, verifica correspondência fuzzy no nome + substring no caminho relativo
3. Marca todos os ancestrais como visíveis (auto-expansão)
4. Armazena resultados no cache

**Lógica de filtro fuzzy (`fuzzyMatch`):**
- Implementa matching de subsequência: caracteres devem aparecer na ordem, mas não consecutivamente
- Exemplo: "abc" corresponde a "aXbYc"
- Case-insensitive

**Virtualização (`buildVisibleItems`):**
1. Percorre a árvore em pré-ordem
2. Respeita estado de expansão e filtro
3. Filtra nós ignorados (se `showIgnored == false`)
4. Armazena resultados em `visibleItems` (slice linear)
5. Ordena filhos: diretórios primeiro, depois arquivos (alfabético)

**Estados de seleção (`recomputeSelectionStates`):**
- Post-order traversal (bottom-up)
- `SelectionSelected`: todos os filhos visíveis selecionados
- `SelectionUnselected`: nenhum filho visível selecionado
- `SelectionPartial`: alguns filhos visíveis selecionados

**Scroll automático (`adjustScroll`):**
- Mantém cursor visível na viewport
- Rola automaticamente ao mover para cima/baixo

**Estatísticas:**
- `GetVisibleFileCount()` — conta arquivos visíveis (após filtro)
- `GetTotalFileCount()` — conta todos os arquivos (respeita `showIgnored`, ignora filtro)

**Empty states (`renderEmptyState`):**
| Condição | Mensagem |
|----------|----------|
| Filtro definido, sem correspondências | "No files match filter 'X'. Press Ctrl+C to clear." |
| Todos ignorados | "All files are hidden (gitignored or custom-ignored). Press 'i' to show ignored." |
| Vazio mesmo sem filtro | "This directory is empty." |

**Scrollbar:** Renderizado à direita da árvore quando há mais itens que a altura visível. Usa caracteres `█` para o polegar e `│` para a faixa.

---

## 3. Análise de Dependências

### 3.1 Dependências Internas

| Dependência | Uso |
|------------|-----|
| `internal/config.ConfigMetadata` | Metadados de configuração (key, type, description, enum options) |
| `internal/config.ValidateValue` | Validação de valor (referenciado, código não visível neste módulo) |
| `internal/config.ConfigType` | Enum de tipos de configuração (TypeString, TypeInt, etc.) |
| `internal/ui/styles` | Estilos e cores (PrimaryColor, TextColor, ErrorColor, etc.) |
| `internal/core/scanner.FileNode` | Nó raiz da árvore de arquivos |
| `internal/core/tokens.FormatTokens` | Formatação de contagem de tokens |

### 3.2 Dependências Externas

| Dependência | Uso |
|------------|-----|
| `github.com/charmbracelet/bubbles/textinput` | Campo de texto (apenas ConfigFieldModel) |
| `github.com/charmbracelet/bubbles/spinner` | Spinner animado (ProgressModel) |
| `github.com/charmbracelet/bubbletea` | Interface tea.Model (Init/Update/View) |
| `github.com/charmbracelet/lipgloss` | Estilização de texto terminal |

### 3.3 Fluxo de Acoplamento

```
internal/ui/components
  ├── ConfigFieldModel ──→ config.ConfigMetadata (leitura)
  │                         config.ValidateValue (chamada)
  │                         textinput.Model (encapsulado)
  ├── ConfigSelectModel ──→ config.ConfigMetadata (leitura)
  ├── ConfigToggleModel ──→ config.ConfigMetadata (leitura)
  ├── ProgressModel ──────→ spinner.Model (encapsulado)
  │                         (sem dependência direta do config)
  ├── UsageBar ───────────→ tokens.FormatTokens (chamada)
  └── FileTreeModel ──────→ scanner.FileNode (leitura)
                            styles.SelectionState (leitura)
```

**Observação:** `ProgressModel` e `UsageBar` são os únicos componentes que **não dependem** de `internal/config`. Os três widgets de configuração dependem exclusivamente de `ConfigMetadata`.

---

## 4. Padrões Arquiteturais Identificados

### 4.1 Widget Pattern com Interface Implícita

Os três componentes de configuração (`ConfigField`, `ConfigSelect`, `ConfigToggle`) compartilham a mesma assinatura de métodos sem declarar uma interface Go explícita:

```go
Init() tea.Cmd
Update(msg) (*Model, tea.Cmd)
Focus() tea.Cmd
Blur()
Value() T
SetValue(T)
IsModified() bool
IsFocused() bool
SetWidth(int)
Metadata() ConfigMetadata
View() string
```

Esta interface implícita permite que componentes de nível superior tratem todos os três tipos de forma polymórfica (via slice de interfaces ou reflection).

### 4.2 Controle Imperativo vs. Declarativo

- **ConfigField, ConfigSelect, ConfigToggle**: Seguem o padrão Bubble Tea MVU com `Update(tea.Msg)`. Cada um tem sua própria máquina de estados de teclas dentro do `Update`.
- **ProgressModel**: **Não implementa** `Update(tea.Msg)`. É atualizado imperativamente via chamadas de método (`Update()`, `UpdateMessage()`, `Hide()`, `Show()`). Isso sugere que o componente é controlado externamente pelo loop principal do Bubble Tea.
- **FileTreeModel**: **Não implementa** `Update(tea.Msg)`. Todas as interações são via chamadas de método imperativo (`MoveUp()`, `MoveDown()`, `ExpandNode()`, etc.), sugerindo que um componente pai controla a árvore e invoca esses métodos em resposta a eventos.

### 4.3 Virtualização de Lista

`FileTreeModel` implementa virtualização de lista de forma customizada:
- `visibleItems` contém apenas os itens visíveis na viewport (O(height) em vez de O(total_nodes))
- `topIndex` + `height` determinam qual porção é renderizada
- `adjustScroll()` mantém o cursor dentro da área visível
- Scrollbar visual é renderizada dinamicamente

### 4.4 Cache de Filtro

`FileTreeModel` implementa otimização de filtro com três campos:
- `lastFilter`: último filtro computado
- `filterCacheValid`: flag de validade
- `filterMatches`: mapa de caminhos correspondentes

Quando o filtro não muda e o cache é válido, `computeFilterMatches()` retorna imediatamente sem recomputação.

---

## 5. Cobertura de Testes

| Arquivo | Tests | Principais Cenários Cobertos |
|---------|-------|------------------------------|
| `config_field_test.go` | 10 tests | Creation, focus/blur, setValue, setWidth, view, update unfocused, init, error, placeholder, metadata |
| `config_select_test.go` | 13 tests | Creation, empty options, out-of-bounds, focus/blur, setValue, expand/collapse, navigation, tab cycle, view states, empty option display, metadata |
| `config_toggle_test.go` | 9 tests | Creation, focus/blur, setValue, toggle (enter/space/tab), Y/N shortcuts, unfocused update, view states, init, metadata |
| `progress_test.go` | 18 tests | Creation, update, updateMessage, hide/show, isVisible, setSize, view visible/invisible, progressBar (zero/overflow), centerLine, padCenter, truncate, visualWidth, getProgress, init, updateSpinner |
| `tree_test.go` | 30+ tests | Creation (nil/with tree), setSize, navigation (up/down/bounds), expand/collapse, toggle selection (file/directory), showIgnored, filter, getSelections, view, recomputeSelectionStates, areAllFilesInDirSelected, setDirectorySelection, shouldShowNode, adjustScroll, formatFileSize, fuzzyMatch, fuzzyAncestors, empty states, selectAll, visible/total counts, deselectAll |

**Total de testes:** ~80+ funções de teste, cobrindo todos os caminhos principais.

---

## 6. Observações e Inferred Items 🟡

| Item | Arquivo | Inferência |
|------|---------|------------|
| **ValidateValue** | config_field.go | Função `config.ValidateValue(key, value)` é chamada mas não definida neste módulo. Inferido como validador baseado em tipo. |
| **styles.* constantes** | Todos | Referência a `styles.PrimaryColor`, `styles.TextColor`, `styles.MutedColor`, `styles.SuccessColor`, `styles.ErrorColor`, `styles.BorderColor`, `styles.AccentColor`, `styles.Nord10`, `styles.Nord6`, `styles.InputLabelStyle`, `styles.StatsValueStyle`, `styles.HelpStyle`, `styles.FocusedBorderStyle`, `styles.BlurredBorderStyle`, `styles.RenderError`, `styles.RenderSuccess`, `styles.RenderFileName`, `styles.RenderIgnoreIndicator`, `styles.SelectedStyle`, `styles.InputPlaceholderStyle` — todas inferidas do módulo `internal/ui/styles`. |
| **ConfigKey constants** | tests | `config.KeyScannerMaxFiles`, `config.KeyScannerMaxFileSize`, `config.KeyTemplateCustomPath`, `config.KeyOutputFormat`, `config.KeyLLMModel`, `config.KeyScannerSkipBinary` — inferidos dos testes. |
| **SelectionState enum** | tree.go | `styles.SelectionState` com valores `SelectionSelected`, `SelectionUnselected`, `SelectionPartial` — inferido dos testes. |
| **FormatTokens** | progress.go | `tokens.FormatTokens(tokens int)` — inferido da importação `internal/core/tokens`. |
| **Bubble Tea integration** | ProgressModel | `ProgressModel` não implementa `Update(tea.Msg)` diretamente — inferido que é controlado imperativamente por um componente pai. |
| **Bubble Tea integration** | FileTreeModel | `FileTreeModel` não implementa `Update(tea.Msg)` diretamente — inferido que é controlado imperativamente por um componente pai (provavelmente um wizard). |
