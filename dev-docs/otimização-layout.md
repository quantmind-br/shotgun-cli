# Arquitetura de UIs de Terminal Responsivas: Um Guia Abrangente para Otimização de Layout com Go, Bubble Tea e Lip Gloss

Este guia aborda os princípios e técnicas essenciais para construir interfaces de usuário de terminal (TUIs) responsivas e robustas usando o ecossistema Go, com foco nas bibliotecas `bubbletea` e `lipgloss`.

## A Paradigma Orientada a Estado: Fundamentos da Arquitetura `bubbletea`

Para otimizar o layout, é imperativo compreender a filosofia da `bubbletea`. Construída sobre os princípios da Arquitetura Elm, a `bubbletea` introduz um paradigma onde a UI é uma representação pura e direta do estado da aplicação. A complexidade da UI é transferida para a gestão desse estado.

Esta arquitetura baseia-se em três componentes essenciais:

### 1. O Model
No cerne de qualquer aplicação `bubbletea` está o **Model**. Geralmente uma `struct` em Go, o Model encapsula todo o estado da aplicação. Qualquer dado que possa influenciar a exibição — desde texto em um campo de entrada até as dimensões da janela — deve residir aqui.

### 2. A Função `Update`
A função **`Update`** é o único local onde o estado pode ser modificado. Ela atua como um despachante central, recebendo mensagens (`tea.Msg`) que representam eventos (pressionamentos de tecla, respostas de rede, etc.). Com base na mensagem, a função processa a lógica e retorna uma **nova instância** do Model com o estado atualizado.

### 3. A Função `View`
A função **`View`** é responsável por renderizar a UI. Ela recebe o Model atual como argumento e retorna uma `string` que representa a aparência da tela. A `View` deve ser "burra", simplesmente refletindo o estado atual sem conter lógica de negócios.

## O Desafio Central: Gerenciando o Redimensionamento da Janela (`tea.WindowSizeMsg`)

O evento mais crucial para layouts responsivos é a `tea.WindowSizeMsg`. Esta mensagem é enviada pela `bubbletea` sempre que o tamanho do terminal muda.

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Lógica para lidar com o redimensionamento
        m.width = msg.Width
        m.height = msg.Height
        return m, nil
    //... outros casos de mensagens
    }
    return m, nil
}
```

As dimensões recebidas (`msg.Width` e `msg.Height`) devem ser armazenadas no Model para que a função `View` possa usá-las para calcular o layout. Em aplicações com componentes aninhados, a `tea.WindowSizeMsg` deve ser propagada para todos os filhos que precisam se ajustar.

### O Problema da Inicialização Assíncrona e o Padrão "Ready State"

A `bubbletea` consulta o tamanho da janela de forma assíncrona para não bloquear a inicialização. Isso cria uma condição de corrida: a primeira chamada para `model.View()` pode ocorrer **antes** que a `tea.WindowSizeMsg` inicial seja recebida.

A consequência é que, na primeira renderização, `m.width` e `m.height` podem ser `0`, causando falhas visuais ou até mesmo `panics`.

A solução robusta é o padrão de **"Estado Pronto" (Ready State)**.

```go
const (
    initializing state = iota
    ready
)

type model struct {
    state  state
    width  int
    height int
    //... outros campos do modelo
}

func (m model) View() string {
    if m.state == initializing {
        return "Initializing..."
    }
    // Renderização da UI completa usando m.width e m.height
    return "UI is ready!"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.state = ready // Transiciona para o estado 'ready'
        //...
    }
    return m, nil
}
```
Neste padrão, a `View` renderiza uma mensagem de carregamento até que o estado mude para `ready`, garantindo que as dimensões da janela sejam válidas.

## Layouts Declarativos e Dinâmicos com `lipgloss`

Com as dimensões da janela gerenciadas, a biblioteca `lipgloss` é usada para construir os layouts. Ela oferece uma API declarativa, inspirada em CSS, para estilizar e compor texto.

Para compor layouts, `lipgloss` fornece `lipgloss.JoinVertical` e `lipgloss.JoinHorizontal`.

### O Problema da Aritmética Estática vs. Medição Dinâmica

Uma armadilha comum é usar aritmética estática para calcular o tamanho dos componentes.

**Exemplo Frágil:**
```go
func (m model) View() string {
    headerStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
    // A linha abaixo assume que o rodapé e o cabeçalho têm altura 1
    contentStyle := lipgloss.NewStyle().Width(m.width).Height(m.height - 2)
    footerStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)

    headerView := headerStyle.Render("Header")
    contentView := contentStyle.Render("Main Content...")
    footerView := footerStyle.Render("Footer")

    return lipgloss.JoinVertical(lipgloss.Top, headerView, contentView, footerView)
}
```
Se uma borda for adicionada ao `headerStyle`, sua altura muda e o cálculo `m.height - 2` se torna incorreto.

A abordagem robusta é o padrão **"Renderizar, Medir, Calcular"**.

**Exemplo Resiliente:**
```go
func (m model) View() string {
    // 1. Renderizar componentes com tamanhos variáveis
    headerStyle := lipgloss.NewStyle().
        Width(m.width).
        Border(lipgloss.NormalBorder(), true)
    headerView := headerStyle.Render("Header")

    footerStyle := lipgloss.NewStyle().Width(m.width)
    footerView := footerStyle.Render("Footer")

    // 2. Medir a altura real dos componentes renderizados
    headerHeight := lipgloss.Height(headerView)
    footerHeight := lipgloss.Height(footerView)

    // 3. Calcular o espaço restante dinamicamente
    contentHeight := m.height - headerHeight - footerHeight
    contentStyle := lipgloss.NewStyle().
        Width(m.width).
        Height(contentHeight).
        Align(lipgloss.Center, lipgloss.Center)
    contentView := contentStyle.Render("Main Content...")

    return lipgloss.JoinVertical(lipgloss.Top, headerView, contentView, footerView)
}
```
Ao usar `lipgloss.Height()`, obtemos a altura real do componente, tornando o layout dinâmico e adaptável.

## Arquitetura Avançada: Componentização e Gerenciamento de Foco

Para aplicações complexas, a solução é a **componentização**: uma "árvore de modelos" onde um modelo raiz orquestra múltiplos filhos. O modelo raiz atua como um roteador de mensagens:
- **Mensagens Globais** (`tea.KeyCtrlC`): Tratadas pelo raiz.
- **Mensagens de Broadcast** (`tea.WindowSizeMsg`): Propagadas para todos os filhos.
- **Mensagens Focadas** (Entrada de usuário): Roteadas para o componente ativo.

### Gerenciamento de Foco
A aplicação deve saber qual componente está "em foco" para direcionar a entrada. Isso é gerenciado como parte do estado do Model.

```go
// Exemplo de delegação de foco do kancli
case tea.KeyMsg:
    switch {
    //...
    case key.Matches(msg, keys.Left):
        m.cols[m.focused].Blur()
        m.focused = m.focused.getPrev()
        m.cols[m.focused].Focus()
    case key.Matches(msg, keys.Right):
        m.cols[m.focused].Blur()
        m.focused = m.focused.getNext()
        m.cols[m.focused].Focus()
    }
```

### Layouts Baseados em Grade vs. Baseados em Fluxo
Para layouts complexos como dashboards, bibliotecas como `bubblelayout` oferecem sistemas de grade mais flexíveis.

| Característica      | `lipgloss.Join*`                                                    | `bubblelayout`                                                              |
| ------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| **Paradigma**       | Baseado em Fluxo (Flow-based)                                       | Baseado em Grade (Grid-based)                                               |
| **Complexidade**    | Baixa                                                               | Média a Alta                                                                |
| **Flexibilidade**   | Limitada a arranjos lineares                                        | Alta, com suporte a vãos e ancoragem                                        |
| **Caso de Uso**     | UIs simples, seções empilhadas                                      | Dashboards complexos, layouts de múltiplos painéis, UIs semelhantes a IDEs. |
| **Primitivas Chave**| `lipgloss.JoinVertical(...)`, `lipgloss.JoinHorizontal(...)`        | `bl.New()`, `layout.Add(...)`, `layout.Resize(...)`, `msg.Size(id)`         |

## TUIs de Alta Performance: Concorrência e Otimização

Operações de I/O (rede, disco) devem ser executadas de forma assíncrona para não bloquear a UI. Isso é feito através de `tea.Cmd`, que são funções que executam tarefas e retornam uma `tea.Msg` ao final.

### Padrões Comuns de `tea.Cmd`

- **Comando Simples**: Para tarefas sem argumentos.
  ```go
  func checkServer() tea.Msg {
      //... faz uma requisição HTTP
      if err != nil {
          return errMsg{err}
      }
      return serverUpMsg{}
  }
  ```
- **Comando com Argumentos**: Usa um fechamento (closure) para passar dados.
  ```go
  func fetchUserData(id int) tea.Cmd {
      return func() tea.Msg {
          user, err := api.FetchUser(id)
          if err != nil {
              return userErrMsg{err}
          }
          return userLoadedMsg{user}
      }
  }
  ```
- **Execução em Lote (`tea.Batch`)**: Executa múltiplos comandos concorrentemente.
  ```go
  cmd := tea.Batch(
      fetchUserData(123),
      fetchSiteStatus("example.com"),
  )
  ```
- **Execução em Sequência (`tea.Sequence`)**: Executa comandos em ordem, um após o outro.
  ```go
  cmd := tea.Sequence(
      saveToDiskCmd,
      tea.Quit, // Executado apenas após saveToDiskCmd terminar
  )
  ```
- **Injeção de Mensagens Externas (`program.Send`)**: Comunica com a aplicação a partir de uma goroutine externa.

## Aplicação Prática: Gerenciando Conteúdo em Componentes Adaptativos

O contrato é simples: **o pai dita os limites, o filho gerencia o conteúdo**.

### APIs de Componentes para Redimensionamento
- `list.Model`: Expõe `SetSize(width, height)`, que deve ser chamado pelo pai.
- `viewport.Model`: Possui campos `Width` e `Height` que são definidos pelo pai.

### Tratamento de Overflow de Conteúdo: Truncamento de Texto
Quando o conteúdo excede o espaço, ele precisa ser truncado de forma inteligente, respeitando caracteres multibyte.

```go
import "unicode/utf8"

// TruncateWithEllipsis trunca uma string 's' para um comprimento máximo 'maxLen'
// de runas, adicionando "..." no final se o truncamento ocorrer.
func TruncateWithEllipsis(s string, maxLen int) string {
    if maxLen <= 0 {
        return ""
    }
    if utf8.RuneCountInString(s) <= maxLen {
        return s
    }
    if maxLen <= 3 {
        return string([]rune(s)[:maxLen])
    }

    runes := []rune(s)
    return string(runes[:maxLen-3]) + "..."
}
```
Esta função garante que o conteúdo textual se adapte graciosamente aos limites de largura definidos pelo pai.

## Conclusão: Princípios para o Desenvolvimento Robusto de TUIs

A construção de TUIs eficazes com `bubbletea` depende da adesão a um conjunto de princípios:

- **Estado em Primeiro Lugar**: A UI é uma função pura do estado. A complexidade pertence ao `Model`.
- **Abrace a Assincronicidade**: Use o padrão de "estado pronto" para lidar com a inicialização assíncrona do tamanho da janela.
- **Meça, Não Adivinhe**: Adote o padrão "renderizar, medir, calcular" com `lipgloss` para criar layouts resilientes.
- **Componha e Delegue**: Estruture aplicações como uma árvore de modelos, onde o raiz orquestra os filhos.
- **Mantenha a UI Fluida**: Nunca bloqueie a `Update` ou a `View`. Delegue operações de I/O para `tea.Cmd`.
- **O Pai Dita os Limites**: Componentes filhos devem ser agnósticos quanto ao seu tamanho externo, que é ditado pelo pai.

A adesão a esses princípios permite a criação de aplicações de linha de comando sofisticadas, responsivas e de fácil manutenção.