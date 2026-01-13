# Plano de Implementacao: TUI de Configuracao

## Resumo Executivo

Implementar uma interface TUI (Terminal User Interface) interativa para gerenciamento de todas as configuracoes da aplicacao shotgun-cli. O TUI sera aberto quando o usuario executar `shotgun-cli config` (sem subcomandos), proporcionando uma experiencia intuitiva para visualizar, editar e validar configuracoes organizadas por categorias.

## Analise de Requisitos

### Requisitos Funcionais

- [ ] RF01: Abrir TUI quando usuario executar `shotgun-cli config` sem argumentos
- [ ] RF02: Exibir configuracoes organizadas por categorias (Scanner, Context, Template, Output, LLM, Gemini)
- [ ] RF03: Permitir navegacao entre categorias usando teclas de direcao ou Tab
- [ ] RF04: Permitir edicao inline de valores de configuracao
- [ ] RF05: Validar valores em tempo real durante edicao (usando `config.ValidateValue`)
- [ ] RF06: Exibir mensagens de erro claras para valores invalidos
- [ ] RF07: Salvar configuracoes automaticamente ao sair ou sob demanda (Ctrl+S)
- [ ] RF08: Mostrar fonte de cada valor (default, config file, environment)
- [ ] RF09: Permitir reset de valor individual para default
- [ ] RF10: Exibir descricao/help contextual para cada configuracao
- [ ] RF11: Suportar diferentes tipos de input (texto, numero, boolean toggle, select/dropdown)
- [ ] RF12: Manter compatibilidade com subcomandos existentes (`config show`, `config set`)

### Requisitos Nao-Funcionais

- [ ] RNF01: Interface responsiva que se adapta ao tamanho do terminal
- [ ] RNF02: Tempo de inicializacao < 100ms
- [ ] RNF03: Seguir o tema Nord ja estabelecido no projeto (`internal/ui/styles`)
- [ ] RNF04: Acessibilidade via teclado completa (sem dependencia de mouse)
- [ ] RNF05: Feedback visual imediato para acoes do usuario

## Analise Tecnica

### Arquitetura Proposta

```
cmd/config.go                    # Entry point - detecta ausencia de subcomando
    |
    v
internal/ui/config_wizard.go     # Model principal do TUI de config
    |
    +-- internal/ui/screens/config_category.go    # Screen por categoria
    |
    +-- internal/ui/components/config_field.go    # Componente de campo editavel
    |
    +-- internal/ui/components/config_toggle.go   # Toggle para booleans
    |
    +-- internal/ui/components/config_select.go   # Select para enums
    |
    +-- internal/config/metadata.go               # Metadados das configs (tipo, descricao, opcoes)
```

### Componentes Afetados

| Arquivo/Modulo | Tipo de Mudanca | Descricao |
|----------------|-----------------|-----------|
| `cmd/config.go` | Modificar | Adicionar deteccao de ausencia de subcomando para lancar TUI |
| `internal/config/keys.go` | Modificar | Adicionar constantes de categoria |
| `internal/config/metadata.go` | Criar | Metadados de cada config (tipo, descricao, opcoes validas) |
| `internal/ui/config_wizard.go` | Criar | Model principal do TUI de configuracao |
| `internal/ui/screens/config_category.go` | Criar | Screen para exibir/editar configs de uma categoria |
| `internal/ui/components/config_field.go` | Criar | Componente de campo de texto editavel |
| `internal/ui/components/config_toggle.go` | Criar | Componente toggle para booleans |
| `internal/ui/components/config_select.go` | Criar | Componente select para valores enum |
| `internal/ui/styles/theme.go` | Modificar | Adicionar estilos especificos para campos de config |

### Dependencias

- **Internas**: 
  - `internal/config` (validacao, keys)
  - `internal/ui/styles` (tema Nord)
  - `github.com/spf13/viper` (leitura/escrita de config)
  
- **Externas** (ja existentes):
  - `github.com/charmbracelet/bubbletea` (framework TUI)
  - `github.com/charmbracelet/bubbles` (componentes como textinput)
  - `github.com/charmbracelet/lipgloss` (styling)

## Plano de Implementacao

### Fase 1: Infraestrutura de Metadados

**Objetivo**: Criar sistema de metadados que descreve cada configuracao

#### Tarefas:

1. **Criar `internal/config/metadata.go`**
   - Descricao detalhada: Definir estrutura que mapeia cada chave de configuracao para seus metadados (tipo, descricao, opcoes validas, valor default)
   - Arquivos envolvidos: `internal/config/metadata.go`
   
   ```go
   // internal/config/metadata.go
   package config
   
   // ConfigType represents the type of a configuration value
   type ConfigType int
   
   const (
       TypeString ConfigType = iota
       TypeInt
       TypeBool
       TypeSize      // e.g., "10MB", "500KB"
       TypePath
       TypeURL
       TypeEnum      // predefined set of values
       TypeTimeout   // integer with range validation
   )
   
   // ConfigCategory represents a group of related configurations
   type ConfigCategory string
   
   const (
       CategoryScanner  ConfigCategory = "Scanner"
       CategoryContext  ConfigCategory = "Context"
       CategoryTemplate ConfigCategory = "Template"
       CategoryOutput   ConfigCategory = "Output"
       CategoryLLM      ConfigCategory = "LLM Provider"
       CategoryGemini   ConfigCategory = "Gemini Integration"
   )
   
   // ConfigMetadata holds metadata about a configuration key
   type ConfigMetadata struct {
       Key          string
       Category     ConfigCategory
       Type         ConfigType
       Description  string
       DefaultValue interface{}
       EnumOptions  []string     // For TypeEnum
       MinValue     int          // For TypeInt/TypeTimeout
       MaxValue     int          // For TypeInt/TypeTimeout
       Required     bool
   }
   
   // AllConfigMetadata returns metadata for all configuration keys
   func AllConfigMetadata() []ConfigMetadata {
       return []ConfigMetadata{
           // Scanner
           {
               Key:          KeyScannerMaxFiles,
               Category:     CategoryScanner,
               Type:         TypeInt,
               Description:  "Maximum number of files to scan",
               DefaultValue: 10000,
               MinValue:     1,
               MaxValue:     1000000,
           },
           {
               Key:          KeyScannerMaxFileSize,
               Category:     CategoryScanner,
               Type:         TypeSize,
               Description:  "Maximum size per file (e.g., 1MB, 500KB)",
               DefaultValue: "1MB",
           },
           // ... demais configuracoes
       }
   }
   
   // GetMetadata returns metadata for a specific key
   func GetMetadata(key string) (ConfigMetadata, bool) {
       for _, m := range AllConfigMetadata() {
           if m.Key == key {
               return m, true
           }
       }
       return ConfigMetadata{}, false
   }
   
   // GetByCategory returns all configs for a category
   func GetByCategory(category ConfigCategory) []ConfigMetadata {
       var result []ConfigMetadata
       for _, m := range AllConfigMetadata() {
           if m.Category == category {
               result = append(result, m)
           }
       }
       return result
   }
   
   // AllCategories returns all categories in display order
   func AllCategories() []ConfigCategory {
       return []ConfigCategory{
           CategoryScanner,
           CategoryContext,
           CategoryTemplate,
           CategoryOutput,
           CategoryLLM,
           CategoryGemini,
       }
   }
   ```

2. **Adicionar testes para metadata**
   - Arquivos envolvidos: `internal/config/metadata_test.go`
   - Verificar que todas as chaves em `ValidKeys()` tem metadata
   - Verificar consistencia entre metadata e validadores existentes

### Fase 2: Componentes de Input Especializados

**Objetivo**: Criar componentes reutilizaveis para diferentes tipos de configuracao

#### Tarefas:

1. **Criar `internal/ui/components/config_field.go`**
   - Descricao: Componente base para campos de texto editaveis com validacao
   - Arquivos envolvidos: `internal/ui/components/config_field.go`
   
   ```go
   // internal/ui/components/config_field.go
   package components
   
   import (
       "github.com/charmbracelet/bubbles/textinput"
       tea "github.com/charmbracelet/bubbletea"
       "github.com/quantmind-br/shotgun-cli/internal/config"
   )
   
   type ConfigFieldModel struct {
       metadata    config.ConfigMetadata
       input       textinput.Model
       value       string
       originalVal string
       err         error
       focused     bool
       modified    bool
   }
   
   func NewConfigField(meta config.ConfigMetadata, currentValue string) *ConfigFieldModel {
       ti := textinput.New()
       ti.Placeholder = meta.Description
       ti.SetValue(currentValue)
       
       return &ConfigFieldModel{
           metadata:    meta,
           input:       ti,
           value:       currentValue,
           originalVal: currentValue,
       }
   }
   
   func (m *ConfigFieldModel) Focus() tea.Cmd {
       m.focused = true
       return m.input.Focus()
   }
   
   func (m *ConfigFieldModel) Blur() {
       m.focused = false
       m.input.Blur()
   }
   
   func (m *ConfigFieldModel) Update(msg tea.Msg) tea.Cmd {
       var cmd tea.Cmd
       m.input, cmd = m.input.Update(msg)
       
       newVal := m.input.Value()
       if newVal != m.value {
           m.value = newVal
           m.modified = newVal != m.originalVal
           m.err = config.ValidateValue(m.metadata.Key, newVal)
       }
       
       return cmd
   }
   
   func (m *ConfigFieldModel) View() string {
       // Renderizar campo com indicadores de estado
       // - Borda colorida se focado
       // - Indicador de modificado
       // - Mensagem de erro se invalido
   }
   
   func (m *ConfigFieldModel) Value() string { return m.value }
   func (m *ConfigFieldModel) IsValid() bool { return m.err == nil }
   func (m *ConfigFieldModel) IsModified() bool { return m.modified }
   func (m *ConfigFieldModel) Error() error { return m.err }
   ```

2. **Criar `internal/ui/components/config_toggle.go`**
   - Descricao: Toggle visual para valores booleanos
   - Arquivos envolvidos: `internal/ui/components/config_toggle.go`
   
   ```go
   // internal/ui/components/config_toggle.go
   package components
   
   type ConfigToggleModel struct {
       metadata    config.ConfigMetadata
       value       bool
       originalVal bool
       focused     bool
   }
   
   func NewConfigToggle(meta config.ConfigMetadata, currentValue bool) *ConfigToggleModel {
       return &ConfigToggleModel{
           metadata:    meta,
           value:       currentValue,
           originalVal: currentValue,
       }
   }
   
   func (m *ConfigToggleModel) Toggle() {
       m.value = !m.value
   }
   
   func (m *ConfigToggleModel) View() string {
       // Renderizar como: [ON ] ou [OFF]
       // Com cores diferentes para cada estado
       // Indicador visual de modificado se diferente do original
   }
   ```

3. **Criar `internal/ui/components/config_select.go`**
   - Descricao: Dropdown/select para valores enum (providers, formats, etc)
   - Arquivos envolvidos: `internal/ui/components/config_select.go`
   
   ```go
   // internal/ui/components/config_select.go
   package components
   
   type ConfigSelectModel struct {
       metadata    config.ConfigMetadata
       options     []string
       selected    int
       originalIdx int
       expanded    bool
       focused     bool
   }
   
   func NewConfigSelect(meta config.ConfigMetadata, currentValue string) *ConfigSelectModel {
       // Encontrar indice do valor atual nas opcoes
   }
   
   func (m *ConfigSelectModel) View() string {
       // Se nao expandido: mostrar valor atual com indicador de dropdown
       // Se expandido: mostrar lista de opcoes com cursor
   }
   ```

### Fase 3: Screen de Categoria

**Objetivo**: Criar tela que exibe e permite editar todas as configs de uma categoria

#### Tarefas:

1. **Criar `internal/ui/screens/config_category.go`**
   - Descricao: Screen que lista e gerencia edicao de configs de uma categoria
   - Arquivos envolvidos: `internal/ui/screens/config_category.go`
   
   ```go
   // internal/ui/screens/config_category.go
   package screens
   
   import (
       tea "github.com/charmbracelet/bubbletea"
       "github.com/quantmind-br/shotgun-cli/internal/config"
       "github.com/quantmind-br/shotgun-cli/internal/ui/components"
   )
   
   type configItem struct {
       metadata config.ConfigMetadata
       field    interface{} // *ConfigFieldModel, *ConfigToggleModel, ou *ConfigSelectModel
   }
   
   type ConfigCategoryModel struct {
       category    config.ConfigCategory
       items       []configItem
       cursor      int
       width       int
       height      int
       scrollY     int
       editMode    bool
   }
   
   func NewConfigCategory(category config.ConfigCategory) *ConfigCategoryModel {
       metas := config.GetByCategory(category)
       items := make([]configItem, len(metas))
       
       for i, meta := range metas {
           currentVal := viper.Get(meta.Key)
           items[i] = configItem{
               metadata: meta,
               field:    createFieldForType(meta, currentVal),
           }
       }
       
       return &ConfigCategoryModel{
           category: category,
           items:    items,
       }
   }
   
   func createFieldForType(meta config.ConfigMetadata, value interface{}) interface{} {
       switch meta.Type {
       case config.TypeBool:
           return components.NewConfigToggle(meta, value.(bool))
       case config.TypeEnum:
           return components.NewConfigSelect(meta, value.(string))
       default:
           return components.NewConfigField(meta, fmt.Sprintf("%v", value))
       }
   }
   
   func (m *ConfigCategoryModel) Update(msg tea.Msg) tea.Cmd {
       // Navegacao com up/down
       // Enter para entrar em modo de edicao
       // Esc para sair de modo de edicao
       // Space para toggle em booleans
       // Tab para proximo campo
   }
   
   func (m *ConfigCategoryModel) View() string {
       // Lista de campos com:
       // - Nome da config
       // - Valor atual (editavel)
       // - Fonte (default/config/env)
       // - Indicador de modificado
       // - Mensagem de erro se invalido
   }
   
   // HasUnsavedChanges returns true if any field was modified
   func (m *ConfigCategoryModel) HasUnsavedChanges() bool
   
   // GetChanges returns map of modified key->value
   func (m *ConfigCategoryModel) GetChanges() map[string]interface{}
   ```

### Fase 4: Wizard Principal de Configuracao

**Objetivo**: Criar o model principal que orquestra a navegacao entre categorias

#### Tarefas:

1. **Criar `internal/ui/config_wizard.go`**
   - Descricao: Model principal que gerencia navegacao entre categorias
   - Arquivos envolvidos: `internal/ui/config_wizard.go`
   
   ```go
   // internal/ui/config_wizard.go
   package ui
   
   import (
       tea "github.com/charmbracelet/bubbletea"
       "github.com/quantmind-br/shotgun-cli/internal/config"
       "github.com/quantmind-br/shotgun-cli/internal/ui/screens"
       "github.com/quantmind-br/shotgun-cli/internal/ui/styles"
   )
   
   type ConfigWizardModel struct {
       categories      []config.ConfigCategory
       categoryScreens map[config.ConfigCategory]*screens.ConfigCategoryModel
       activeCategory  int
       width           int
       height          int
       showHelp        bool
       confirmQuit     bool
       savedMessage    string
       savedMsgTimer   int
   }
   
   func NewConfigWizard() *ConfigWizardModel {
       categories := config.AllCategories()
       categoryScreens := make(map[config.ConfigCategory]*screens.ConfigCategoryModel)
       
       for _, cat := range categories {
           categoryScreens[cat] = screens.NewConfigCategory(cat)
       }
       
       return &ConfigWizardModel{
           categories:      categories,
           categoryScreens: categoryScreens,
           activeCategory:  0,
       }
   }
   
   func (m *ConfigWizardModel) Init() tea.Cmd {
       return nil
   }
   
   func (m *ConfigWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.WindowSizeMsg:
           m.width = msg.Width
           m.height = msg.Height
           // Propagar para screens
           
       case tea.KeyMsg:
           return m.handleKeyPress(msg)
       }
       
       // Delegar para screen ativa
       currentScreen := m.categoryScreens[m.categories[m.activeCategory]]
       cmd := currentScreen.Update(msg)
       
       return m, cmd
   }
   
   func (m *ConfigWizardModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
       switch msg.String() {
       case "ctrl+c", "ctrl+q":
           if m.hasUnsavedChanges() {
               m.confirmQuit = true
               return m, nil
           }
           return m, tea.Quit
           
       case "tab", "shift+tab":
           // Navegar entre categorias
           
       case "ctrl+s":
           // Salvar configuracoes
           return m, m.saveChanges()
           
       case "f1":
           m.showHelp = !m.showHelp
           
       case "r":
           // Reset campo atual para default
       }
       
       return m, nil
   }
   
   func (m *ConfigWizardModel) View() string {
       if m.showHelp {
           return m.renderHelp()
       }
       
       if m.confirmQuit {
           return m.renderConfirmQuit()
       }
       
       // Layout: sidebar com categorias + area principal com campos
       sidebar := m.renderSidebar()
       mainContent := m.categoryScreens[m.categories[m.activeCategory]].View()
       footer := m.renderFooter()
       
       return lipgloss.JoinVertical(
           lipgloss.Left,
           m.renderHeader(),
           lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent),
           footer,
       )
   }
   
   func (m *ConfigWizardModel) renderSidebar() string {
       // Lista de categorias com indicador de ativa
       // Mostrar icone se categoria tem mudancas nao salvas
   }
   
   func (m *ConfigWizardModel) renderHeader() string {
       return styles.RenderHeader(0, "Configuration Settings")
   }
   
   func (m *ConfigWizardModel) renderFooter() string {
       shortcuts := []string{
           "Tab: Next Category",
           "Enter: Edit",
           "Ctrl+S: Save",
           "r: Reset to Default",
           "F1: Help",
           "Ctrl+Q: Quit",
       }
       return styles.RenderFooter(shortcuts)
   }
   
   func (m *ConfigWizardModel) saveChanges() tea.Cmd {
       return func() tea.Msg {
           for _, cat := range m.categories {
               screen := m.categoryScreens[cat]
               changes := screen.GetChanges()
               for key, value := range changes {
                   viper.Set(key, value)
               }
           }
           
           if err := viper.WriteConfig(); err != nil {
               return ConfigSaveErrorMsg{Err: err}
           }
           
           return ConfigSavedMsg{}
       }
   }
   
   func (m *ConfigWizardModel) hasUnsavedChanges() bool {
       for _, cat := range m.categories {
           if m.categoryScreens[cat].HasUnsavedChanges() {
               return true
           }
       }
       return false
   }
   ```

### Fase 5: Integracao com CLI

**Objetivo**: Modificar `cmd/config.go` para lancar TUI quando executado sem subcomandos

#### Tarefas:

1. **Modificar `cmd/config.go`**
   - Descricao: Adicionar RunE ao configCmd para detectar ausencia de subcomando
   - Arquivos envolvidos: `cmd/config.go`
   
   ```go
   // Modificar configCmd em cmd/config.go
   var configCmd = &cobra.Command{
       Use:   "config",
       Short: "Configuration management",
       Long:  "Commands for viewing and modifying shotgun-cli configuration",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Se nenhum subcomando foi especificado, lancar TUI
           return launchConfigTUI()
       },
   }
   
   func launchConfigTUI() error {
       wizard := ui.NewConfigWizard()
       
       program := tea.NewProgram(
           wizard,
           tea.WithAltScreen(),
           tea.WithMouseCellMotion(),
       )
       
       if _, err := program.Run(); err != nil {
           return fmt.Errorf("failed to start config TUI: %w", err)
       }
       
       return nil
   }
   ```

### Fase 6: Estilos e Polish Visual

**Objetivo**: Adicionar estilos especificos e refinar a experiencia visual

#### Tarefas:

1. **Adicionar estilos em `internal/ui/styles/theme.go`**
   - Arquivos envolvidos: `internal/ui/styles/theme.go`
   
   ```go
   // Adicionar ao theme.go
   
   // Config field styles
   var (
       ConfigLabelStyle = lipgloss.NewStyle().
           Foreground(Nord9).
           Bold(true)
       
       ConfigValueStyle = lipgloss.NewStyle().
           Foreground(TextColor)
       
       ConfigValueModifiedStyle = lipgloss.NewStyle().
           Foreground(Nord13). // Yellow for modified
           Bold(true)
       
       ConfigSourceStyle = lipgloss.NewStyle().
           Foreground(MutedColor).
           Italic(true)
       
       ConfigToggleOnStyle = lipgloss.NewStyle().
           Foreground(SuccessColor).
           Bold(true)
       
       ConfigToggleOffStyle = lipgloss.NewStyle().
           Foreground(MutedColor)
       
       CategoryActiveStyle = lipgloss.NewStyle().
           Background(Nord10).
           Foreground(Nord6).
           Padding(0, 1)
       
       CategoryInactiveStyle = lipgloss.NewStyle().
           Foreground(TextColor).
           Padding(0, 1)
       
       CategoryModifiedStyle = lipgloss.NewStyle().
           Foreground(WarningColor)
   )
   
   // RenderConfigField renders a configuration field with its value and source
   func RenderConfigField(label, value, source string, isModified, isFocused bool) string {
       // Implementar rendering do campo
   }
   
   // RenderConfigToggle renders a boolean toggle
   func RenderConfigToggle(value bool, isFocused bool) string {
       if value {
           return ConfigToggleOnStyle.Render("[ON ]")
       }
       return ConfigToggleOffStyle.Render("[OFF]")
   }
   ```

### Fase 7: Testes

**Objetivo**: Garantir cobertura de testes adequada (>85%)

#### Tarefas:

1. **Testes de Metadata**
   - Arquivos: `internal/config/metadata_test.go`
   - Casos: 
     - Todas as chaves tem metadata
     - GetMetadata retorna corretamente
     - GetByCategory retorna configs corretas

2. **Testes de Componentes**
   - Arquivos: `internal/ui/components/config_field_test.go`, `config_toggle_test.go`, `config_select_test.go`
   - Casos:
     - Criacao com valores iniciais
     - Atualizacao de valores
     - Validacao em tempo real
     - Deteccao de modificacao

3. **Testes de Config Category Screen**
   - Arquivos: `internal/ui/screens/config_category_test.go`
   - Casos:
     - Navegacao entre campos
     - Modo de edicao
     - Coleta de mudancas

4. **Testes de Config Wizard**
   - Arquivos: `internal/ui/config_wizard_test.go`
   - Casos:
     - Navegacao entre categorias
     - Salvamento de configuracoes
     - Deteccao de mudancas nao salvas
     - Confirmacao de saida

## Estrategia de Testes

### Testes Unitarios

- [ ] `internal/config/metadata_test.go` - Testes de AllConfigMetadata, GetMetadata, GetByCategory
- [ ] `internal/ui/components/config_field_test.go` - Testes de criacao, update, validacao
- [ ] `internal/ui/components/config_toggle_test.go` - Testes de toggle, view
- [ ] `internal/ui/components/config_select_test.go` - Testes de selecao, expansao
- [ ] `internal/ui/screens/config_category_test.go` - Testes de navegacao, edicao
- [ ] `internal/ui/config_wizard_test.go` - Testes de navegacao, save, quit

### Testes de Integracao

- [ ] Testar fluxo completo: abrir TUI -> editar config -> salvar -> verificar arquivo
- [ ] Testar validacao: editar com valor invalido -> verificar mensagem de erro
- [ ] Testar cancel: editar -> sair sem salvar -> verificar confirmacao

### Casos de Teste Especificos

| ID | Cenario | Input | Output Esperado |
|----|---------|-------|-----------------|
| TC01 | Abrir TUI sem argumentos | `shotgun-cli config` | TUI abre com primeira categoria selecionada |
| TC02 | Navegar categorias | Tab | Proxima categoria selecionada |
| TC03 | Editar campo texto | Enter + digitar + Enter | Valor atualizado, indicador de modificado |
| TC04 | Toggle boolean | Space em campo bool | Valor invertido |
| TC05 | Validacao invalida | Digitar "abc" em campo numerico | Mensagem de erro exibida |
| TC06 | Salvar mudancas | Ctrl+S | Config salva, mensagem de sucesso |
| TC07 | Sair com mudancas | Ctrl+Q com mudancas | Modal de confirmacao aparece |
| TC08 | Reset para default | 'r' em campo | Valor volta ao default |
| TC09 | Subcomando show | `shotgun-cli config show` | Comportamento existente mantido |
| TC10 | Subcomando set | `shotgun-cli config set key val` | Comportamento existente mantido |

## Riscos e Mitigacoes

| Risco | Probabilidade | Impacto | Mitigacao |
|-------|---------------|---------|-----------|
| Complexidade de componentes custom | Media | Alto | Usar bubbles existentes como base (textinput) |
| Conflito com subcomandos existentes | Baixa | Alto | Testar exaustivamente compatibilidade |
| Performance com muitas configs | Baixa | Medio | Lazy loading de categorias nao visíveis |
| Dificuldade em testes de TUI | Media | Medio | Usar pattern de Model isolado testavel |
| Inconsistencia visual com wizard existente | Media | Baixo | Reutilizar styles e patterns existentes |

## Checklist de Conclusao

- [ ] Codigo implementado seguindo Clean Architecture
- [ ] Testes escritos e passando (>85% cobertura)
- [ ] Lint passando (`golangci-lint run`)
- [ ] Documentacao do README atualizada
- [ ] Help text do comando atualizado
- [ ] Code review realizado
- [ ] Feature testada manualmente em diferentes tamanhos de terminal

## Notas Adicionais

### Keyboard Shortcuts Propostos

| Tecla | Acao |
|-------|------|
| Tab / Shift+Tab | Navegar entre categorias |
| Up / Down / j / k | Navegar entre campos |
| Enter | Entrar em modo de edicao |
| Esc | Sair de modo de edicao / Cancelar |
| Space | Toggle para booleans |
| Ctrl+S | Salvar todas as mudancas |
| r | Reset campo atual para default |
| F1 | Mostrar ajuda |
| Ctrl+Q | Sair (com confirmacao se houver mudancas) |

### Layout Visual Proposto

```
+------------------------------------------------------------------+
| Configuration Settings                                            |
+------------------------------------------------------------------+
|                                                                   |
| [Categories]          | [Scanner Settings]                       |
| > Scanner        *    |                                          |
|   Context             | scanner.max-files                        |
|   Template            | [10000                    ] (default)    |
|   Output              | Maximum number of files to scan          |
|   LLM Provider        |                                          |
|   Gemini              | scanner.max-file-size                    |
|                       | [1MB                      ] (config)     |
|                       | Maximum size per file                    |
|                       |                                          |
|                       | scanner.skip-binary                      |
|                       | [ON ]                       (default)    |
|                       | Skip binary files during scanning        |
|                       |                                          |
+------------------------------------------------------------------+
| Tab: Category | Enter: Edit | Ctrl+S: Save | F1: Help | Ctrl+Q: Quit |
+------------------------------------------------------------------+
```

### Proximos Passos Recomendados

1. Revisar este plano com a equipe
2. Criar issue/epic para tracking
3. Iniciar pela Fase 1 (Metadata) que e prerequisito para as demais
4. Desenvolver componentes (Fase 2) em paralelo com screen (Fase 3)
5. Integrar e testar end-to-end
