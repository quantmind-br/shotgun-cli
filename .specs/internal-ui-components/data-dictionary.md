# Dicionário de Dados — internal/ui/components

| Campo            | Valor                                                        |
|------------------|--------------------------------------------------------------|
| **Nome do Módulo**   | `components`                                                 |
| **Nível de Detalhe** | Detalhado (`detalhado`)                                      |
| **Tipos Definidos**  | 6 structs, 1 struct auxiliar local, 1 enum implícito        |
| **Enums Externos Usados** | `config.ConfigType`, `config.ConfigMetadata`, `styles.SelectionState` |
| **Tipos Externos Usados** | `tea.Model`, `textinput.Model`, `spinner.Model`             |

---

## 1. Tipos Definidos no Módulo

### 1.1 `ConfigFieldModel` — Campo de Entrada de Texto

**Arquivo:** `config_field.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `metadata` | `config.ConfigMetadata` | privada | Metadados da configuração (key, type, description) |
| `value` | `string` | privada | Valor atual do campo como string |
| `input` | `textinput.Model` | privada | Instância encapsulada de textinput do Bubble Tea |
| `focused` | `bool` | privada | Indica se o campo está focado |
| `width` | `int` | privada | Largura do container |
| `err` | `error` | privada | Erro de validação atual |
| `modified` | `bool` | privada | Indica se o valor foi modificado pelo usuário |
| `placeholder` | `string` | privada | Texto do placeholder gerado |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Init()` | `tea.Cmd` | Inicializa o campo com blink do input |
| `Update(tea.Msg)` | `(*ConfigFieldModel, tea.Cmd)` | Processa mensagem Bubble Tea; valida valor se modificado |
| `Focus()` | `tea.Cmd` | Ativa foco e delega ao input |
| `Blur()` | — | Desativa foco e delega ao input |
| `IsFocused()` | `bool` | Retorna status de foco |
| `SetWidth(int)` | — | Ajusta largura do input (range: 20-60) |
| `Value()` | `string` | Retorna valor atual |
| `SetValue(string)` | — | Define valor, reseta modified e err |
| `Error()` | `error` | Retorna erro de validação |
| `IsModified()` | `bool` | Retorna se o campo foi modificado |
| `IsValid()` | `bool` | Retorna true se não há erro |
| `Metadata()` | `config.ConfigMetadata` | Retorna metadados da configuração |
| `View()` | `string` | Renderiza HTML-like text para o terminal |

**Função construtora:**
| Nome | Parâmetros | Retorno | Descrição |
|------|-----------|---------|-----------|
| `NewConfigField` | `config.ConfigMetadata`, `currentValue string` | `*ConfigFieldModel` | Cria campo com estilo Bubble Tea |

**Função privada:**
| Nome | Parâmetros | Retorno | Descrição |
|------|-----------|---------|-----------|
| `getPlaceholder` | `config.ConfigMetadata` | `string` | Retorna placeholder baseado no tipo de configuração |

---

### 1.2 `ConfigSelectModel` — Seletor de Opções

**Arquivo:** `config_select.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `metadata` | `config.ConfigMetadata` | privada | Metadados da configuração |
| `options` | `[]string` | privada | Lista de opções disponíveis |
| `selected` | `int` | privada | Índice da opção selecionada |
| `focused` | `bool` | privada | Indica se o componente está focado |
| `width` | `int` | privada | Largura do container |
| `modified` | `bool` | privada | Indica se a seleção foi modificada |
| `expanded` | `bool` | privada | Dropdown aberto ou fechado |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Init()` | `tea.Cmd` | Retorna nil (sem comando) |
| `Update(tea.Msg)` | `(*ConfigSelectModel, tea.Cmd)` | Processa teclas de navegação no dropdown |
| `Focus()` | `tea.Cmd` | Ativa foco |
| `Blur()` | — | Desativa foco e fecha dropdown |
| `IsFocused()` | `bool` | Retorna status de foco |
| `SetWidth(int)` | — | Define largura do container |
| `Value()` | `string` | Retorna opção selecionada (ou "" se inválido) |
| `SetValue(string)` | — | Seleciona opção por valor; reset modified |
| `IsModified()` | `bool` | Retorna se foi modificado |
| `Metadata()` | `config.ConfigMetadata` | Retorna metadados |
| `View()` | `string` | Renderiza dropdown colapsado/expandido |

**Função construtora:**
| Nome | Parâmetros | Retorno | Descrição |
|------|-----------|---------|-----------|
| `NewConfigSelect` | `config.ConfigMetadata`, `currentValue string` | `*ConfigSelectModel` | Cria seletor com opções do metadata ou [currentValue] |

**Métodos privados:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `renderSelect()` | `string` | Renderiza lista de opções (colapsada ou expandida) |

**Máquina de teclas (dentro de `Update`):**
| Tecla | Condição | Ação |
|-------|----------|------|
| `enter`, ` ` | qualquer | Toggle `expanded` |
| `up`, `k` | `expanded == true` | Cursor para cima (wrap) |
| `down`, `j` | `expanded == true` | Cursor para baixo (wrap) |
| `esc` | `expanded == true` | Fecha dropdown |
| `tab` | `expanded == true` | Fecha dropdown |
| `tab` | `expanded == false` | Cicla para próxima opção |

---

### 1.3 `ConfigToggleModel` — Interruptor Booleano

**Arquivo:** `config_toggle.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `metadata` | `config.ConfigMetadata` | privada | Metadados da configuração |
| `value` | `bool` | privada | Valor booleano (true/false) |
| `focused` | `bool` | privada | Indica se o componente está focado |
| `width` | `int` | privada | Largura do container |
| `modified` | `bool` | privada | Indica se o valor foi modificado |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Init()` | `tea.Cmd` | Retorna nil |
| `Update(tea.Msg)` | `(*ConfigToggleModel, tea.Cmd)` | Processa teclas de toggle |
| `Focus()` | `tea.Cmd` | Ativa foco |
| `Blur()` | — | Desativa foco |
| `IsFocused()` | `bool` | Retorna status de foco |
| `SetWidth(int)` | — | Define largura |
| `Value()` | `bool` | Retorna valor booleano |
| `SetValue(bool)` | — | Define valor, reseta modified |
| `IsModified()` | `bool` | Retorna se foi modificado |
| `Metadata()` | `config.ConfigMetadata` | Retorna metadados |
| `View()` | `string` | Renderiza [ENABLED]/[DISABLED] |

**Função construtora:**
| Nome | Parâmetros | Retorno | Descrição |
|------|-----------|---------|-----------|
| `NewConfigToggle` | `config.ConfigMetadata`, `currentValue bool` | `*ConfigToggleModel` | Cria interruptor |

**Máquina de teclas (dentro de `Update`):**
| Tecla | Condição | Ação |
|-------|----------|------|
| `enter`, ` `, `tab` | qualquer | Inverte `value` |
| `y`, `Y` | `value == false` | Define `true` |
| `n`, `N` | `value == true` | Define `false` |

---

### 1.4 `UsageBar` — Barra de Uso de Contexto

**Arquivo:** `progress.go` (inline)

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `CurrentBytes` | `int64` | pública | Bytes usados atualmente |
| `MaxBytes` | `int64` | pública | Limite de bytes (0 = sem limite) |
| `MaxBytesStr` | `string` | pública | Representação legível do limite (ex: "100MB") |
| `TotalTokens` | `int` | pública | Contagem total de tokens |
| `Width` | `int` | pública | Largura da barra |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `View()` | `string` | Renderiza barra com ícone de status + barra visual |

**Função construtora:**
| Nome | Parâmetros | Retorno | Descrição |
|------|-----------|---------|-----------|
| `NewUsageBar` | `current int64`, `max int64`, `maxStr string`, `tokens int`, `width int` | `UsageBar` | Cria barra de uso |

**Indicadores de status (por porcentagem):**
| Condição | Ícone | Cor |
|----------|-------|-----|
| ≤ 80% | ✅ | SuccessColor |
| > 80% | ⚠️ | WarningColor |
| > 100% | ⛔ | ErrorColor |
| Sem limite (MaxBytes=0) | — | StatsValueStyle |

---

### 1.5 `ProgressModel` — Modal de Progresso

**Arquivo:** `progress.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `current` | `int64` | privada | Progresso atual |
| `total` | `int64` | privada | Total (0 = sem limite, <0 = streaming) |
| `stage` | `string` | privada | Nome do estágio atual |
| `message` | `string` | privada | Mensagem descritiva |
| `visible` | `bool` | privada | Visibilidade do modal |
| `width` | `int` | privada | Largura da tela |
| `height` | `int` | privada | Altura da tela |
| `spinner` | `spinner.Model` | privada | Spinner animado (dot) |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Init()` | `tea.Cmd` | Retorna spinner.Tick |
| `UpdateSpinner(tea.Msg)` | `(*ProgressModel, tea.Cmd)` | Atualiza spinner manualmente |
| `Update(int64, int64, string, string)` | — | Define estado completo (current, total, stage, message) |
| `UpdateMessage(string, string)` | — | Troca de estágio (reseta current/total) |
| `Hide()` | — | Oculta modal |
| `Show()` | — | Exibe modal |
| `IsVisible()` | `bool` | Retorna visibilidade |
| `SetSize(int, int)` | — | Define dimensões da tela |
| `GetProgress()` | `(int64, int64, string, string)` | Retorna estado como tupla |
| `View()` | `string` | Renderiza modal com spinner, barra, estatísticas |

**Métodos privados:**
| Método | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `renderProgressBar(int)` | `width` | `string` | Renderiza barra de progresso █░ |
| `centerLine(string)` | `line` | `string` | Centraliza linha no container |
| `padCenter(string, int)` | `text, width` | `string` | Centraliza e limita largura |
| `truncate(string, int)` | `text, maxLen` | `string` | Trunca com "..." |
| `visualWidth(string)` | `text` | `int` | Largura visual (lipgloss) |

**Modos de renderização (baseados em `total`):**
| `total` | Renderização |
|---------|-------------|
| `> 0` | Barra de progresso + porcentagem + contagem (X/Y) |
| `== 0` | Spinner + mensagem |
| `< 0` | Spinner + contador (streaming) |
| `<= 0 && total == 0 && message == ""` | Apenas "Processing" |

---

### 1.6 `FileTreeModel` — Árvore de Arquivos

**Arquivo:** `tree.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `tree` | `*scanner.FileNode` | privada | Raiz da árvore de arquivos |
| `cursor` | `int` | privada | Índice do item no cursor |
| `selections` | `map[string]bool` | privada | Caminhos selecionados (path → selected) |
| `selectionStates` | `map[string]SelectionState` | privada | Estados de seleção cacheados (path → state) |
| `showIgnored` | `bool` | privada | Mostrar arquivos ignorados |
| `filter` | `string` | privada | Texto de filtro fuzzy |
| `filterMatches` | `map[string]bool` | privada | Caminhos correspondentes ao filtro |
| `expanded` | `map[string]bool` | privada | Diretórios expandidos (path → expanded) |
| `width` | `int` | privada | Largura da tela |
| `height` | `int` | privada | Altura da tela |
| `visibleItems` | `[]treeItem` | privada | Itens visíveis na viewport |
| `topIndex` | `int` | privada | Índice do primeiro item visível |
| `lastFilter` | `string` | privada | Cache: último filtro computado |
| `filterCacheValid` | `bool` | privada | Cache de filtro é válido? |

**Métodos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `SetSize(int, int)` | — | Define dimensões da tela |
| `MoveUp()` | — | Move cursor para cima |
| `MoveDown()` | — | Move cursor para baixo |
| `ExpandNode()` | — | Expande diretório no cursor |
| `CollapseNode()` | — | Colapsa diretório no cursor |
| `ToggleSelection()` | — | Toggle seleção (arquivo: individual; dir: recursivo) |
| `SelectAllVisible()` | — | Seleciona todos os arquivos visíveis |
| `DeselectAllVisible()` | — | Deseleciona todos os arquivos visíveis |
| `ToggleShowIgnored()` | — | Alterna visibilidade de ignorados |
| `SetFilter(string)` | — | Define filtro fuzzy |
| `GetFilter()` | `string` | Retorna filtro atual |
| `ClearFilter()` | — | Limpa filtro |
| `GetSelections()` | `map[string]bool` | Retorna mapa de seleções |
| `View()` | `string` | Renderiza árvore com scrollbar |

**Métodos auxiliares (visíveis para testes, mas não parte da interface pública):**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `rebuildVisibleItems()` | — | Reconstrói lista de itens visíveis |
| `recomputeSelectionStates()` | — | Recalcula estados de seleção (bottom-up) |
| `adjustScroll()` | — | Ajusta scroll para manter cursor visível |
| `areAllFilesInDirSelected(*FileNode)` | `bool` | Verifica se todos os arquivos do dir estão selecionados |
| `setDirectorySelection(*FileNode, bool)` | — | Seleciona/deseleciona recursivamente um diretório |
| `walkNode(*FileNode, func)` | — | Recorre árvore aplicando função a cada nó |
| `selectionStateFor(string)` | `SelectionState` | Retorna estado de seleção cacheado |
| `GetVisibleFileCount()` | `int` | Conta arquivos visíveis (após filtro) |
| `GetTotalFileCount()` | `int` | Conta todos os arquivos (ignora filtro) |
| `buildVisibleItems(...)` | — | Pré-ordem traversal com virtualização |
| `shouldShowNode(*FileNode)` | `bool` | Decide se nó deve ser mostrado |
| `computeFilterMatches()` | — | Computa correspondências fuzzy com cache |
| `markAncestorsVisible(*FileNode)` | — | Marca ancestrais como visíveis + auto-expande |
| `renderEmptyState()` | `string` | Renderiza mensagem de estado vazio |
| `renderTreeItem(treeItem, bool)` | `string` | Renderiza linha individual do tree |
| `buildTreePrefix(treeItem)` | `string` | Prefixo de árvore (│ ├ └) |
| `renderCheckbox(treeItem, SelectionState)` | `string` | Checkbox [ ] / [✓] |
| `renderDirIndicator(treeItem)` | `string` | 📂 / 📁 |
| `renderItemName(treeItem, SelectionState)` | `string` | Nome do arquivo/diretório |
| `renderIgnoreStatus(treeItem)` | `string` | Indicador de ignore |
| `renderSizeInfo(treeItem)` | `string` | Tamanho formatado |

**Estrutura auxiliar local:**
| Campo | Tipo | Descrição |
|-------|------|-----------|
| `node` | `*scanner.FileNode` | Nó do scanner |
| `path` | `string` | Caminho completo |
| `depth` | `int` | Profundidade na árvore |
| `isLast` | `bool` | É o último filho? |
| `hasNext` | `[]bool` | Padrão hasNext para cada nível |

---

### 1.7 `treeItem` — Estrutura Auxiliar da Árvore

**Arquivo:** `tree.go`

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `node` | `*scanner.FileNode` | privada | Referência ao nó original do scanner |
| `path` | `string` | privada | Caminho absoluto do nó |
| `depth` | `int` | privada | Nível de profundidade na árvore (0 = raiz) |
| `isLast` | `bool` | privada | Indica se este é o último filho no mesmo nível |
| `hasNext` | `[]bool` | privada | Array de booleans: para cada nível de profundidade, indica se há mais irmãos após este |

---

## 2. Funções Privadas do Módulo

| Função | Arquivo | Parâmetros | Retorno | Descrição |
|--------|---------|-----------|---------|-----------|
| `getPlaceholder` | config_field.go | `ConfigMetadata` | `string` | Mapeia tipo de config para placeholder |
| `renderSelect` | config_select.go | *(receptor ConfigSelectModel)* | `string` | Renderiza opções do dropdown |
| `renderToggle` | config_toggle.go | *(receptor ConfigToggleModel)* | `string` | Renderiza [ENABLED]/[DISABLED] |
| `formatSizeHelper` | progress.go | `int64` | `string` | Formata bytes para KB/MB/GB/TB |
| `renderProgressBar` | progress.go | `int` | `string` | Renderiza barra █░ |
| `centerLine` | progress.go | `string` | `string` | Centraliza horizontalmente |
| `padCenter` | progress.go | `string, int` | `string` | Centraliza com padding |
| `truncate` | progress.go | `string, int` | `string` | Trunca com "..." |
| `visualWidth` | progress.go | `string` | `int` | Largura visual do texto |
| `formatFileSize` | tree.go | `int64` | `string` | Formata bytes para KB/MB/GB/TB |
| `fuzzyMatch` | tree.go | `string, string` | `bool` | Fuzzy matching de subsequência |
| `rebuildVisibleItems` | tree.go | *(receptor FileTreeModel)* | — | Reconstrói lista linear de visíveis |
| `recomputeSelectionStates` | tree.go | *(receptor FileTreeModel)* | — | Recalcula estados de seleção |
| `adjustScroll` | tree.go | *(receptor FileTreeModel)* | — | Ajusta scroll automático |
| `areAllFilesInDirSelected` | tree.go | *(receptor FileTreeModel, *FileNode)* | `bool` | Verifica se todos os arquivos estão selecionados |
| `setDirectorySelection` | tree.go | *(receptor FileTreeModel, *FileNode, bool)* | — | Seleciona/deseleciona diretório recursivamente |
| `walkNode` | tree.go | *(receptor FileTreeModel, *FileNode, func)* | — | Recorre árvore |
| `selectionStateFor` | tree.go | *(receptor FileTreeModel, string)* | `SelectionState` | Lookup de estado de seleção |
| `countFilesInNode` | tree.go | *(receptor FileTreeModel, *FileNode)* | `int` | Conta arquivos recursivamente |
| `shouldShowNode` | tree.go | *(receptor FileTreeModel, *FileNode)* | `bool` | Decisão de visibilidade de nó |
| `buildVisibleItems` | tree.go | *(receptor FileTreeModel, *FileNode, string, int, bool, []bool)* | — | Pré-ordem traversal com virtualização |
| `computeFilterMatches` | tree.go | *(receptor FileTreeModel)* | — | Computa correspondências fuzzy com cache |
| `markAncestorsVisible` | tree.go | *(receptor FileTreeModel, *FileNode)* | — | Marca ancestrais + auto-expande |
| `renderEmptyState` | tree.go | *(receptor FileTreeModel)* | `string` | Mensagem de estado vazio |
| `renderTreeItem` | tree.go | *(receptor FileTreeModel, treeItem, bool)* | `string` | Renderiza linha individual |
| `buildTreePrefix` | tree.go | *(receptor FileTreeModel, treeItem)* | `string` | Prefixo de árvore │├└ |
| `renderCheckbox` | tree.go | *(receptor FileTreeModel, treeItem, SelectionState)* | `string` | Checkbox [ ]/[✓] |
| `renderDirIndicator` | tree.go | *(receptor FileTreeModel, treeItem)* | `string` | 📂/📁 |
| `renderItemName` | tree.go | *(receptor FileTreeModel, treeItem, SelectionState)* | `string` | Nome estilizado |
| `renderIgnoreStatus` | tree.go | *(receptor FileTreeModel, treeItem)* | `string` | Indicador de ignore |
| `renderSizeInfo` | tree.go | *(receptor FileTreeModel, treeItem)* | `string` | Tamanho formatado |

---

## 3. Tipos Externos Referenciados

### 3.1 `config.ConfigMetadata` (pacote `internal/config`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `Key` | `string` | Chave de configuração (ex: "scanner.max-files") |
| `Category` | `ConfigCategory` | Grupo lógico (Scanner, Context, Template, Output, LLM Provider) |
| `Type` | `ConfigType` | Tipo de valor (string, int, bool, size, path, url, enum, timeout) |
| `Description` | `string` | Descrição legível |
| `DefaultValue` | `interface{}` | Valor padrão |
| `EnumOptions` | `[]string` | Opções válidas para tipo enum |
| `MinValue` | `int` | Valor mínimo (int/timeout) |
| `MaxValue` | `int` | Valor máximo (int/timeout) |
| `Required` | `bool` | Se o valor é obrigatório |

### 3.2 `config.ConfigType` (enum, pacote `internal/config`)

| Constante | Valor (iota) | Representação |
|-----------|-------------|---------------|
| `TypeString` | 0 | "string" |
| `TypeInt` | 1 | "int" |
| `TypeBool` | 2 | "bool" |
| `TypeSize` | 3 | "size" |
| `TypePath` | 4 | "path" |
| `TypeURL` | 5 | "url" |
| `TypeEnum` | 6 | "enum" |
| `TypeTimeout` | 7 | "timeout" |

### 3.3 `styles.SelectionState` (pacote `internal/ui/styles`)

| Constante | Descrição |
|-----------|-----------|
| `SelectionSelected` | Todos os filhos visíveis selecionados |
| `SelectionUnselected` | Nenhum filho visível selecionado |
| `SelectionPartial` | Alguns filhos visíveis selecionados |

### 3.4 `scanner.FileNode` (pacote `internal/core/scanner`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `Name` | `string` | Nome base do arquivo/diretório |
| `Path` | `string` | Caminho absoluto |
| `RelPath` | `string` | Caminho relativo da raiz de scan |
| `IsDir` | `bool` | É diretório? |
| `Children` | `[]*FileNode` | Filhos (apenas para diretórios) |
| `IsGitignored` | `bool` | Ignorado por .gitignore? |
| `IsCustomIgnored` | `bool` | Ignorado por regras customizadas? |
| `Size` | `int64` | Tamanho em bytes (0 para diretórios) |
| `Expanded` | `bool` | Expandido na UI? |
| `Parent` | `*FileNode` | Referência ao pai (não serializado) |
