# Plano de Implementação: Feedback Visual de Seleção de Arquivos e Pastas

## 1. Visão Geral

Este documento descreve o plano detalhado para implementar feedback visual aprimorado no processo de seleção de arquivos e pastas no TUI wizard do shotgun-cli. A implementação adicionará cores distintas para indicar visualmente o estado de seleção de arquivos e diretórios.

## 2. Problema Atual

Atualmente, o TUI não fornece feedback visual suficiente sobre o estado de seleção de diretórios:

- **Arquivos**: Têm checkboxes `[✓]` (selecionado) ou `[ ]` (não selecionado)
- **Diretórios**: Não possuem indicação visual de seleção, apenas emojis 📁/📂 para expandido/colapsado
- **Limitação**: Usuários precisam expandir diretórios para verificar se seus arquivos estão selecionados

### Impacto na UX

- Dificulta navegação em projetos grandes
- Requer múltiplas operações de expand/collapse para verificar seleções
- Falta de clareza sobre o estado parcial de seleção em diretórios

## 3. Solução Proposta

### 3.1. Estados de Seleção

Definir três estados visuais distintos para arquivos e diretórios:

1. **Não Selecionado** - Nenhum arquivo no nó está selecionado
2. **Totalmente Selecionado** - Todos os arquivos no nó/subárvore estão selecionados
3. **Parcialmente Selecionado** - Alguns (mas não todos) arquivos no nó/subárvore estão selecionados

### 3.2. Esquema de Cores

Baseado na análise do arquivo `internal/ui/styles/theme.go`, o esquema de cores atual utiliza:

```go
PrimaryColor   = "#00ADD8" // Azul ciano - títulos e destaque
SecondaryColor = "#5E81AC" // Azul acinzentado
AccentColor    = "#A3BE8C" // Verde - sucesso e progresso
ErrorColor     = "#BF616A" // Vermelho - erros
WarningColor   = "#EBCB8B" // Amarelo - avisos
SuccessColor   = "#A3BE8C" // Verde - sucesso (mesmo que AccentColor)
MutedColor     = "#5C7E8C" // Azul acinzentado escuro - texto secundário
TreeStyle      = "#ECEFF4" // Branco/cinza claro - árvore de arquivos
```

#### Cores Propostas para Estados de Seleção

| Estado | Cor | Código Hex | Justificativa |
|--------|-----|------------|---------------|
| **Não Selecionado** | MutedColor | `#5C7E8C` | Cor neutra que indica estado inativo, mantém consistência com texto secundário |
| **Totalmente Selecionado** | SuccessColor | `#A3BE8C` | Verde indica ação completa/sucesso, consistente com tema de progresso |
| **Parcialmente Selecionado** | WarningColor | `#EBCB8B` | Amarelo indica estado intermediário/atenção, consistente com avisos |

#### Visualização do Esquema

```
Não Selecionado:         📁 src/              (cor: #5C7E8C - azul acinzentado)
Parcialmente Selecionado: 📁 internal/         (cor: #EBCB8B - amarelo)
Totalmente Selecionado:   📁 cmd/              (cor: #A3BE8C - verde)
```

## 4. Análise Técnica Detalhada

### 4.1. Estrutura de Dados Atual

#### FileNode (`internal/core/scanner/scanner.go`)

```go
type FileNode struct {
    Name            string       // Nome do arquivo/diretório
    Path            string       // Caminho absoluto
    RelPath         string       // Caminho relativo
    IsDir           bool         // É diretório?
    Children        []*FileNode  // Filhos (se diretório)
    Selected        bool         // ❗ CAMPO EXISTENTE MAS NÃO USADO
    IsGitignored    bool         // Ignorado por .gitignore?
    IsCustomIgnored bool         // Ignorado por regras customizadas?
    Size            int64        // Tamanho do arquivo
    Expanded        bool         // Expandido no TUI?
    Parent          *FileNode    // Referência ao pai
}
```

**Observação Importante**: O campo `Selected` existe na struct mas **não é utilizado** atualmente. As seleções são gerenciadas em um `map[string]bool` no `FileTreeModel`.

#### FileTreeModel (`internal/ui/components/tree.go`)

```go
type FileTreeModel struct {
    tree         *scanner.FileNode      // Raiz da árvore
    cursor       int                    // Posição do cursor
    selections   map[string]bool        // Mapa path -> selecionado
    showIgnored  bool                   // Mostrar arquivos ignorados?
    filter       string                 // Filtro ativo
    expanded     map[string]bool        // Mapa path -> expandido
    width        int                    // Largura disponível
    height       int                    // Altura disponível
    visibleItems []treeItem             // Itens visíveis (cache)
    topIndex     int                    // Índice do topo (scroll)
}
```

### 4.2. Lógica de Seleção Atual

#### Métodos de Seleção

1. **ToggleSelection()** (linha 96-103)
   - Alterna seleção de arquivo individual
   - Apenas para arquivos (não diretórios)
   - Modifica `selections[path]`

2. **ToggleDirectorySelection()** (linha 105-114)
   - Alterna seleção de todos os arquivos em um diretório
   - Usa `areAllFilesInDirSelected()` para determinar estado atual
   - Chama `setDirectorySelection()` para propagar

3. **areAllFilesInDirSelected()** (linha 338-352)
   - **CRÍTICO**: Verifica se todos os arquivos em um diretório estão selecionados
   - Retorna `true` apenas se todos os arquivos estão selecionados
   - Já faz a lógica necessária para determinar estado "totalmente selecionado"

4. **setDirectorySelection()** (linha 354-364)
   - Define seleção de todos os arquivos em um diretório
   - Percorre recursivamente usando `walkNode()`

### 4.3. Lógica de Renderização Atual

#### renderTreeItem() (linha 170-240)

```go
func (m *FileTreeModel) renderTreeItem(item treeItem, isCursor bool) string {
    // 1. Constrói prefixo da árvore (│, ├, └)
    var prefix strings.Builder
    for d := 0; d < item.depth; d++ {
        if d < len(item.hasNext) && item.hasNext[d] {
            prefix.WriteString("│  ")
        } else {
            prefix.WriteString("   ")
        }
    }

    // 2. Adiciona conector da árvore
    if item.depth > 0 {
        if item.isLast {
            prefix.WriteString("└──")
        } else {
            prefix.WriteString("├──")
        }
    }

    // 3. Checkbox (apenas para arquivos)
    var checkbox string
    if !item.node.IsDir {
        if m.selections[item.path] {
            checkbox = "[✓] "
        } else {
            checkbox = "[ ] "
        }
    }

    // 4. Indicador de diretório
    var dirIndicator string
    if item.node.IsDir {
        if m.expanded[item.path] {
            dirIndicator = "📂 "
        } else {
            dirIndicator = "📁 "
        }
    }

    // 5. Nome do arquivo/diretório
    name := filepath.Base(item.path)
    if item.node.IsDir {
        name += "/"
    }

    // 6. Status de ignore
    var ignoreStatus string
    if item.node.IsGitignored {
        ignoreStatus = " (g)"
    } else if item.node.IsCustomIgnored {
        ignoreStatus = " (c)"
    }

    // 7. Tamanho do arquivo
    var sizeInfo string
    if !item.node.IsDir && item.node.Size > 0 {
        sizeInfo = fmt.Sprintf(" (%s)", formatFileSize(item.node.Size))
    }

    // 8. Combina todas as partes
    line := prefix.String() + checkbox + dirIndicator + name + ignoreStatus + sizeInfo

    // 9. Aplica destaque do cursor
    if isCursor {
        line = styles.SelectedStyle.Render(line)  // ❗ ÚNICO USO DE ESTILO
    }

    return line
}
```

**Pontos Críticos**:
- Atualmente, **apenas o cursor** recebe estilo visual (linha 235-236)
- Nome do arquivo/diretório não tem cor específica baseada em seleção
- Checkbox para arquivos é texto simples (sem cor)

## 5. Implementação Detalhada

### 5.1. Modificações em `internal/ui/styles/theme.go`

#### 5.1.1. Adicionar Tipo e Estilos de Seleção

Adicionar após linha 52 (`TreeStyle`):

```go
type SelectionState int

const (
    SelectionUnselected SelectionState = iota
    SelectionPartial
    SelectionSelected
)

// Selection state colors for file tree
FileUnselectedColor = lipgloss.Color("#5C7E8C") // Muted blue-gray
FileSelectedColor   = lipgloss.Color("#A3BE8C") // Success green
FilePartialColor    = lipgloss.Color("#EBCB8B") // Warning yellow

// Styles for file/directory names based on selection state
UnselectedNameStyle = lipgloss.NewStyle().
    Foreground(FileUnselectedColor)

SelectedNameStyle = lipgloss.NewStyle().
    Foreground(FileSelectedColor).
    Bold(true)

PartialNameStyle = lipgloss.NewStyle().
    Foreground(FilePartialColor).
    Bold(true)
```

**Justificativa**:
- O tipo forte evita typos em strings e melhora mensagens de erro/testes
- `Bold(true)` em Selected e Partial aumenta destaque visual
- Mantém consistência com estilos existentes (ex: `SelectedStyle`, `ErrorStyle`)

#### 5.1.2. Adicionar Função Helper para Renderização de Nomes

Adicionar ao final do arquivo (após linha 246):

```go
// RenderFileName applies color styling to file/directory names based on selection state
func RenderFileName(name string, selectionState SelectionState) string {
    switch selectionState {
    case SelectionSelected:
        return SelectedNameStyle.Render(name)
    case SelectionPartial:
        return PartialNameStyle.Render(name)
    case SelectionUnselected:
        return UnselectedNameStyle.Render(name)
    default:
        return TreeStyle.Render(name)
    }
}
```

**Parâmetros**:
- `name`: Nome do arquivo/diretório a ser renderizado
- `selectionState`: Valor do tipo `SelectionState` (`SelectionUnselected`, `SelectionPartial`, `SelectionSelected`)

**Retorno**: String com estilo Lip Gloss aplicado

### 5.2. Modificações em `internal/ui/components/tree.go`

#### 5.2.1. Introduzir Cache de Estado de Seleção

**Passo 1 — Estrutura**
- Acrescentar campo `selectionStates map[string]styles.SelectionState` à struct `FileTreeModel`
- Inicializar o mapa em `NewFileTree` e executar `model.recomputeSelectionStates()` após copiar o mapa `selections`

**Passo 2 — Recomputar quando necessário**
- Invocar `recomputeSelectionStates()` em todos os pontos que alteram seleção ou visibilidade:
  - Após `ToggleSelection()`
  - Após `ToggleDirectorySelection()` (logo depois de `setDirectorySelection`)
  - Dentro de `setDirectorySelection()` ao final das atualizações
  - No final de `rebuildVisibleItems()` (garante coerência quando filtro/expansão muda)

**Passo 3 — Implementação do cache**
Adicionar o método abaixo ao final do arquivo, antes de `formatFileSize()`:

```go
func (m *FileTreeModel) recomputeSelectionStates() {
    m.selectionStates = make(map[string]styles.SelectionState)
    if m.tree == nil {
        return
    }

    var visit func(node *scanner.FileNode) styles.SelectionState
    visit = func(node *scanner.FileNode) styles.SelectionState {
        if !m.shouldShowNode(node) {
            return styles.SelectionUnselected
        }

        if !node.IsDir {
            state := styles.SelectionUnselected
            if m.selections[node.Path] {
                state = styles.SelectionSelected
            }
            m.selectionStates[node.Path] = state
            return state
        }

        hasSelected := false
        hasUnselected := false

        for _, child := range node.Children {
            childState := visit(child)
            switch childState {
            case styles.SelectionSelected:
                hasSelected = true
            case styles.SelectionUnselected:
                hasUnselected = true
            case styles.SelectionPartial:
                hasSelected = true
                hasUnselected = true
            }
        }

        state := styles.SelectionUnselected
        switch {
        case hasSelected && !hasUnselected:
            state = styles.SelectionSelected
        case hasSelected && hasUnselected:
            state = styles.SelectionPartial
        }

        m.selectionStates[node.Path] = state
        return state
    }

    visit(m.tree)
}
```

**Passo 4 — Função de acesso**

```go
func (m *FileTreeModel) selectionStateFor(path string) styles.SelectionState {
    if state, ok := m.selectionStates[path]; ok {
        return state
    }
    return styles.SelectionUnselected
}
```

#### 5.2.2. Modificar `renderTreeItem()` para Usar Cores

**Localização**: Linha 170-240

**Modificação 1**: Determinar estado de seleção (adicionar após linha 190)

```go
func (m *FileTreeModel) renderTreeItem(item treeItem, isCursor bool) string {
    var prefix strings.Builder

    // ... código existente de construção de prefixo ...

    // ▼ ADICIONAR APÓS LINHA 190 (após construção do prefixo)

    // Determinar estado de seleção para aplicação de cores
    selectionState := m.selectionStateFor(item.path)

    // ... resto do código continua ...
}
```

**Modificação 2**: Aplicar cor ao nome do arquivo/diretório (modificar linhas 211-215)

```go
// CÓDIGO ORIGINAL (linhas 211-215):
// File name
name := filepath.Base(item.path)
if item.node.IsDir {
    name += "/"
}

// ▼ SUBSTITUIR POR:

// File name with color based on selection state
baseName := filepath.Base(item.path)
if item.node.IsDir {
    baseName += "/"
}
name := styles.RenderFileName(baseName, selectionState)
```

**Modificação 3**: Aplicar cor ao checkbox (modificar linhas 192-199)

```go
// CÓDIGO ORIGINAL (linhas 192-199):
var checkbox string
if !item.node.IsDir {
    if m.selections[item.path] {
        checkbox = "[✓] "
    } else {
        checkbox = "[ ] "
    }
}

// ▼ SUBSTITUIR POR:

var checkbox string
if !item.node.IsDir {
    checkboxText := "[ ] "
    if m.selections[item.path] {
        checkboxText = "[✓] "
    }
    // Aplica cor ao checkbox baseado no estado de seleção
    checkbox = styles.RenderFileName(checkboxText, selectionState)
}
```

**Modificação 4**: Manter lógica de cursor highlight (linha 234-237)

```go
// ▼ MANTER CÓDIGO EXISTENTE
if isCursor {
    line = styles.SelectedStyle.Render(line)
}
```

**Observação**: O cursor highlight (`SelectedStyle`) usa background color e sobrescreve foreground, então funcionará corretamente sobre os novos estilos de cor.

### 5.3. Diagrama de Fluxo de Renderização

```
renderTreeItem(item, isCursor)
    │
    ├─> Construir prefixo da árvore (│, ├, └)
    │
    ├─> selectionStateFor(item.path)  ◄── NOVO
    │   │
    │   └─> O estado já foi pré-computado por recomputeSelectionStates()
    │
    ├─> Construir checkbox (se arquivo)
    │   └─> styles.RenderFileName(checkboxText, selectionState)  ◄── MODIFICADO
    │
    ├─> Construir indicador de diretório (📁/📂)
    │
    ├─> Construir nome
    │   └─> styles.RenderFileName(baseName, selectionState)  ◄── MODIFICADO
    │
    ├─> Adicionar ignoreStatus e sizeInfo
    │
    ├─> Combinar todas as partes em `line`
    │
    └─> Se isCursor: aplicar SelectedStyle.Render(line)
        └─> Retornar linha final
```

## 6. Casos de Teste

### 6.1. Cenários de Teste Manual

#### Teste 1: Arquivo Individual

**Setup**:
```
src/
  ├── main.go (não selecionado)
  └── utils.go (selecionado)
```

**Ações**:
1. Navegar até `main.go`
2. Pressionar `Space` para selecionar

**Resultado Esperado**:
- `main.go` deve mudar de cor azul-acinzentada (#5C7E8C) para verde (#A3BE8C)
- Checkbox deve mudar de `[ ]` para `[✓]`
- Ambos (checkbox e nome) devem estar em verde

#### Teste 2: Diretório Totalmente Selecionado

**Setup**:
```
cmd/
  ├── root.go
  ├── context.go
  └── template.go
```

**Ações**:
1. Navegar até `cmd/`
2. Pressionar `d` para selecionar todos os arquivos

**Resultado Esperado**:
- `cmd/` deve estar em verde (#A3BE8C) e bold
- Todos os arquivos filhos devem ter checkboxes `[✓]` em verde

#### Teste 3: Diretório Parcialmente Selecionado

**Setup**:
```
internal/
  ├── core/
  │   ├── scanner.go (selecionado)
  │   └── context.go (não selecionado)
  └── ui/
      └── wizard.go (não selecionado)
```

**Ações**:
1. Selecionar apenas `scanner.go`
2. Colapsar `internal/core/`

**Resultado Esperado**:
- `internal/` deve estar em amarelo (#EBCB8B) e bold (1 de 3 arquivos selecionados)
- `internal/core/` deve estar em amarelo (#EBCB8B) e bold (1 de 2 arquivos selecionados)

#### Teste 4: Transição de Estados

**Setup**:
```
src/
  ├── file1.go (não selecionado)
  ├── file2.go (não selecionado)
  └── file3.go (não selecionado)
```

**Ações**:
1. Estado inicial: `src/` deve estar azul-acinzentada
2. Selecionar `file1.go`: `src/` deve mudar para amarelo
3. Selecionar `file2.go`: `src/` deve continuar amarelo
4. Selecionar `file3.go`: `src/` deve mudar para verde
5. Desselecionar `file3.go`: `src/` deve voltar para amarelo
6. Desselecionar `file2.go`: `src/` deve continuar amarelo
7. Desselecionar `file1.go`: `src/` deve voltar para azul-acinzentada

**Resultado Esperado**: Transições de cor corretas em cada etapa

#### Teste 5: Diretórios Aninhados

**Setup**:
```
project/
  ├── src/
  │   ├── core/
  │   │   ├── a.go (selecionado)
  │   │   └── b.go (selecionado)
  │   └── ui/
  │       ├── x.go (não selecionado)
  │       └── y.go (selecionado)
  └── test/
      └── test.go (não selecionado)
```

**Resultado Esperado**:
- `project/`: Amarelo (3 de 5 arquivos selecionados)
- `src/`: Amarelo (3 de 4 arquivos selecionados)
- `core/`: Verde (2 de 2 arquivos selecionados)
- `ui/`: Amarelo (1 de 2 arquivos selecionados)
- `test/`: Azul-acinzentado (0 de 1 arquivo selecionado)

#### Teste 6: Interação com Cursor

**Ações**:
1. Navegar com setas para diferentes arquivos/diretórios
2. Verificar que cursor highlight (background azul + texto branco) funciona sobre as cores

**Resultado Esperado**:
- Cursor highlight deve sobrescrever cores de seleção
- Ao mover cursor, cores originais devem permanecer

#### Teste 7: Arquivos Ignorados

**Setup**:
```
src/
  ├── main.go (selecionado)
  ├── temp.log (ignorado, não selecionado)
  └── build/ (ignorado)
```

**Ações**:
1. Pressionar `i` para mostrar arquivos ignorados

**Resultado Esperado**:
- `src/` deve estar amarelo (1 de 2 arquivos selecionados, considerando apenas não-ignorados)
- Arquivos ignorados devem manter cores de seleção se aplicável

#### Teste 8: Filtro Ativo

**Setup**: Diretório com múltiplos arquivos

**Ações**:
1. Selecionar alguns arquivos
2. Pressionar `/` e filtrar por extensão (ex: `.go`)

**Resultado Esperado**:
- Cores devem refletir seleções apenas de arquivos visíveis após filtro
- Ao limpar filtro, cores devem atualizar para refletir todas as seleções

### 6.2. Testes de Performance

#### Teste P1: Árvore Grande (1000+ arquivos)

**Métricas a Medir**:
- Tempo de renderização inicial
- Tempo de resposta ao navegar (up/down)
- Uso de memória

**Critério de Aceitação**:
- Renderização inicial < 500ms
- Navegação < 50ms por movimento
- Incremento de memória < 10% em relação ao original

#### Teste P2: Seleção/Deseleção em Diretório Grande

**Setup**: Diretório com 500+ arquivos

**Ações**:
1. Pressionar `d` para selecionar todos
2. Medir tempo de resposta

**Critério de Aceitação**:
- Resposta < 200ms para 500 arquivos
- Resposta < 1s para 5000 arquivos

### 6.3. Testes de Regressão

#### Teste R1: Funcionalidade Existente

**Verificações**:
- [x] Navegação com setas/vim keys continua funcionando
- [x] Seleção de arquivo individual (`Space`) funciona
- [x] Seleção de diretório (`d`) funciona
- [x] Filtro (`/`) funciona
- [x] Toggle ignored (`i`) funciona
- [x] Rescan (`F5`) funciona
- [x] Checkboxes continuam aparecendo para arquivos

#### Teste R2: Outras Telas do Wizard

**Verificações**:
- [x] Template Selection não é afetada
- [x] Task Input não é afetada
- [x] Rules Input não é afetada
- [x] Review screen não é afetada

## 7. Considerações de Performance

### 7.1. Análise de Complexidade

- `recomputeSelectionStates()`: O(n) onde *n* é o número de nós visíveis; executa apenas quando há mudança de seleção, filtro ou expansão relevante.
- `selectionStateFor()`: O(1) por item renderizado (lookup direto no mapa).
- `renderTreeItem()`: Mantém O(1) porque consome apenas o estado já armazenado.
- Memória adicional: O(n) para o mapa `selectionStates`.

### 7.2. Benchmarks a Implementar (Opcional)

```go
// Em internal/ui/components/tree_test.go

func BenchmarkRecomputeSelectionStates_Small(b *testing.B) {
    // Árvore com ~100 nós
}

func BenchmarkRecomputeSelectionStates_Large(b *testing.B) {
    // Árvore com ~10_000 nós (stress test)
}

func BenchmarkRenderTree_WithCachedStates(b *testing.B) {
    // Renderização completa com estados já computados
}

func BenchmarkToggleSelection_Recompute(b *testing.B) {
    // Simula múltiplos toggles sequenciais
}
```

## 8. Plano de Implementação em Etapas

### Fase 1: Preparação (Estimativa: 30 minutos)

**Etapa 1.1: Backup e Branch**
```bash
git checkout -b feature/visual-selection-feedback
git add .
git commit -m "checkpoint: antes de implementar feedback visual"
```

**Etapa 1.2: Revisar Código Atual**
- Ler novamente `internal/ui/styles/theme.go`
- Ler novamente `internal/ui/components/tree.go`
- Identificar linhas exatas a modificar

### Fase 2: Modificações em Styles (Estimativa: 20 minutos)

**Etapa 2.1: Adicionar Cores e Estilos**
- Editar `internal/ui/styles/theme.go`
- Adicionar constantes de cor (linhas após 52)
- Adicionar estilos de seleção (linhas após 52)
- Adicionar função `RenderFileName()` (linha 246+)

**Etapa 2.2: Compilar e Testar**
```bash
make build
# Verificar que compila sem erros
```

### Fase 3: Modificações em Tree Component (Estimativa: 45 minutos)

**Etapa 3.1: Introduzir cache de seleção**
- Atualizar struct `FileTreeModel` adicionando `selectionStates map[string]styles.SelectionState`
- Inicializar o campo em `NewFileTree`
- Implementar `selectionStateFor()`

**Etapa 3.2: Implementar `recomputeSelectionStates()`**
- Criar função pós-ordem conforme seção 5.2.1
- Garantir que respeita `shouldShowNode()`
- Invocar após `ToggleSelection`, `ToggleDirectorySelection`, `setDirectorySelection` e `rebuildVisibleItems`

**Etapa 3.3: Modificar `renderTreeItem()`**
- Usar `selectionStateFor(item.path)` para obter estado
- Colorir checkbox e nome com `styles.RenderFileName`
- Manter destaque do cursor

**Etapa 3.4: Compilar e Testar**
```bash
make build
# Verificar compilação
```

### Fase 4: Testes Manuais (Estimativa: 45 minutos)

**Etapa 4.1: Testes Básicos**
- Executar aplicação: `./build/shotgun-cli`
- Realizar Teste 1 (Arquivo Individual)
- Realizar Teste 2 (Diretório Totalmente Selecionado)
- Realizar Teste 3 (Diretório Parcialmente Selecionado)

**Etapa 4.2: Testes de Transição**
- Realizar Teste 4 (Transição de Estados)
- Realizar Teste 5 (Diretórios Aninhados)

**Etapa 4.3: Testes de Integração**
- Realizar Teste 6 (Interação com Cursor)
- Realizar Teste 7 (Arquivos Ignorados)
- Realizar Teste 8 (Filtro Ativo)

### Fase 5: Testes de Performance (Estimativa: 30 minutos)

**Etapa 5.1: Preparar Projeto de Teste**
```bash
# Clonar projeto grande para teste
git clone https://github.com/kubernetes/kubernetes /tmp/k8s-test
cd /tmp/k8s-test
```

**Etapa 5.2: Executar Testes de Performance**
- Teste P1: Árvore grande
- Teste P2: Seleção/deseleção em massa
- Medir e documentar resultados

**Etapa 5.3: Validar Métricas**
- Confirmar em perfis manuais/benchmarks que `recomputeSelectionStates()` mantém tempo aceitável em árvores médias e grandes
- Registrar resultados relevantes (ex.: tempo médio) na descrição do PR

### Fase 6: Testes de Regressão (Estimativa: 20 minutos)

**Etapa 6.1: Funcionalidade Existente**
- Verificar Teste R1 (todas as funcionalidades)

**Etapa 6.2: Outras Telas**
- Verificar Teste R2 (outras telas do wizard)

### Fase 7: Documentação e Finalização (Estimativa: 20 minutos)

**Etapa 7.1: Atualizar Documentação**
- Atualizar `CLAUDE.md` se necessário
- Adicionar screenshots (opcional)

**Etapa 7.2: Commit Final**
```bash
git add .
git commit -m "feat: improve visual selection feedback

- add SelectionState constants and themed styles
- maintain cached selectionStates with recomputeSelectionStates
- colorize tree checkboxes and names via selectionStateFor
- colors: unselected (#5C7E8C), partial (#EBCB8B), selected (#A3BE8C)"
```

## 9. Estimativa Total de Tempo

| Fase | Descrição | Tempo Estimado |
|------|-----------|----------------|
| 1 | Preparação | 30 minutos |
| 2 | Modificações em Styles | 20 minutos |
| 3 | Modificações em Tree Component | 45 minutos |
| 4 | Testes Manuais | 45 minutos |
| 5 | Testes de Performance | 30 minutos |
| 6 | Testes de Regressão | 20 minutos |
| 7 | Documentação e Finalização | 20 minutos |
| **Total** | | **3h 30min** |

**Nota**: Tempo adicional de 1-2 horas pode ser necessário para ajustes UX ou benchmarks adicionais.

## 10. Riscos e Mitigações

### Risco 1: Performance Degradada em Árvores Grandes

**Probabilidade**: Baixa
**Impacto**: Médio

**Mitigação**:
- Cache de estados já implementado via `recomputeSelectionStates()` e `selectionStateFor()`
- Executar benchmarks da Fase 5 para validar o comportamento em árvores com milhares de nós
- Monitorar consumo de memória do mapa `selectionStates` em projetos extremos

### Risco 2: Conflito Visual com Cursor Highlight

**Probabilidade**: Baixa
**Impacto**: Médio

**Mitigação**:
- Cursor usa background color, sobrescreve foreground de forma previsível
- Testar explicitamente no Teste 6
- Se necessário, ajustar `SelectedStyle` para melhor contraste

### Risco 3: Cores Difíceis de Distinguir em Alguns Terminais

**Probabilidade**: Média
**Impacto**: Médio

**Mitigação**:
- Escolher cores com contraste adequado (já verificado no esquema atual)
- Testar em terminais comuns: iTerm2, Alacritty, Windows Terminal, GNOME Terminal
- Considerar adicionar configuração para desabilitar cores (futuro)

### Risco 4: Comportamento Inesperado com Filtros/Ignored

**Probabilidade**: Baixa
**Impacto**: Médio

**Mitigação**:
- `recomputeSelectionStates()` consulta `shouldShowNode()` antes de registrar cada nó
- Testar explicitamente no Teste 7 e Teste 8
- Documentar comportamento esperado

## 11. Melhorias Futuras

### 11.1. Curto Prazo

1. **Indicadores Adicionais para Diretórios Parciais**
   - Adicionar símbolo visual (ex: `◐` ou `◔`) ao lado de diretórios parcialmente selecionados
   - Não altera cores, apenas adiciona informação extra

2. **Configuração de Cores**
   - Permitir usuário customizar cores via config.yaml
   - Exemplo:
     ```yaml
     ui:
       colors:
         file_selected: "#A3BE8C"
         file_partial: "#EBCB8B"
         file_unselected: "#5C7E8C"
     ```

### 11.2. Médio Prazo

1. **Modo de Alto Contraste**
   - Adicionar flag `--high-contrast` ou config `ui.high_contrast: true`
   - Usar cores com maior diferenciação para acessibilidade

2. **Legenda de Cores**
   - Adicionar legenda no footer da file selection screen
   - Exemplo: `Verde: Selecionado | Amarelo: Parcial | Cinza: Não selecionado`

3. **Animação de Transição**
   - Animação sutil ao mudar estado de seleção
   - Requer atualização para Bubble Tea com suporte a animações

### 11.3. Longo Prazo

1. **Temas Customizáveis**
   - Suporte a múltiplos temas (dark, light, colorblind-friendly)
   - Carregamento de temas de arquivos externos

2. **Estatísticas de Seleção em Tempo Real**
   - Mostrar no header: "3 de 5 arquivos selecionados em cmd/"
   - Atualizar dinamicamente ao navegar

## 12. Conclusão

Este plano detalha a implementação completa de feedback visual de seleção para o TUI do shotgun-cli. A solução proposta:

✅ **Resolve o problema**: Fornece feedback visual imediato sobre estados de seleção
✅ **Mantém consistência**: Usa esquema de cores existente da aplicação
✅ **É performático**: Implementação O(n) com path para otimização se necessário
✅ **É testável**: Conjunto completo de casos de teste definidos
✅ **É extensível**: Base sólida para melhorias futuras

**Próximos Passos**:
1. Revisar e aprovar este plano
2. Executar Fase 1 (Preparação)
3. Implementar Fases 2-7 sequencialmente
4. Realizar code review
5. Merge para branch principal

---

**Última Atualização**: 2025-10-02
**Autor**: Claude Code (com input do usuário)
**Status**: Aguardando Aprovação
