# Análise de Código — Módulo `internal/core/template`

> **Pacote:** `github.com/quantmind-br/shotgun-cli/internal/core/template`
> **Nível:** `detalhado`
> **Linguagem:** Go 1.21+
> **Total de linhas de código (produção):** 732
> **Total de linhas de teste:** 1590
> **Arquivos fonte:** 4 (`template.go`, `loader.go`, `manager.go`, `renderer.go`)
> **Arquivos de teste:** 4 (`template_struct_test.go`, `loader_test.go`, `manager_test.go`, `renderer_test.go`)

---

## 1. Visão Geral

O módulo `internal/core/template` implementa um sistema completo de gerenciamento e renderização de templates markdown para geração de prompts de IA. Ele segue um design em três camadas:

| Camada | Responsabilidade | Arquivo Principal |
|--------|-----------------|-------------------|
| **Domínio** | Estrutura de dados `Template`, extração de metadados, validação | `template.go` |
| **Carregamento** | Fontes multiplas (embedded, filesystem, user, custom) | `loader.go` |
| **Gerenciamento** | API unificada, merge de fontes, thread-safety | `manager.go` |
| **Renderização** | Substituição de variáveis, sanitização, preview | `renderer.go` |

### Dependências Externas

| Dependência | Uso |
|-------------|-----|
| `github.com/adrg/xdg` | Resolvimento do diretório XDG_CONFIG_HOME para templates do usuário |
| `github.com/quantmind-br/shotgun-cli/internal/assets` | Embedded filesystem com templates embarcados (`//go:embed`) |
| `github.com/stretchr/testify` | Framework de testes |

### Dependências Internas

- `internal/assets` — fornece `assets.Templates` (go:embed fs)

**Zero dependências circulares.** O pacote importa apenas `internal/assets` (que é leaf).

---

## 2. Estrutura de Arquivos e Responsabilidades

### 2.1 `template.go` (200 linhas) — Camada de Domínio

**Tipo principal:** `Template` (struct + métodos)

**Constantes:**
- `VarTask = "TASK"`
- `VarRules = "RULES"`
- `VarFileStructure = "FILE_STRUCTURE"`
- `VarCurrentDate = "CURRENT_DATE"`

**RegEx global:** `variablePattern = regexp.MustCompile(`\{([A-Z_][A-Z0-9_]*)\}`)`

| Função | Tipo | Visibilidade | Responsabilidade |
|--------|------|-------------|-----------------|
| `parseTemplate` | package-private | `func` | Parseia conteúdo `.md` → `*Template` (nome, descrição, vars) |
| `extractTemplateName` | package-private | `func` | Remove `.md` e prefixo `prompt_` do nome |
| `extractDescription` | package-private | `func` | Extrai primeiro `# Header` ou `<!-- comment -->` |
| `extractRequiredVars` | package-private | `func` | Usa `variablePattern` para extrair variáveis únicas |
| `validateTemplateContent` | package-private | `func` | Valida balanceamento de `{}` fora de code blocks |
| `GetVariableNames` | receiver | `*Template` | Retorna variáveis encontradas no conteúdo |
| `HasVariable` | receiver | `*Template` | Verifica presença literal de `{VAR}` no conteúdo |
| `GetVariableCount` | receiver | `*Template` | Conta ocorrências de `{VAR}` |
| `IsValid` | receiver | `*Template` | Valida nome, conteúdo e balanceamento de chaves |

**Observações de qualidade:**
- `extractRequiredVars` tem tag `//nolint:unparam` — reserva erro para futura lógica de validação.
- `validateTemplateContent` faz dupla verificação (balanceamento + regex) por linha.
- `HasVariable` usa `strings.Contains` simples — não valida formato regex.
- `GetVariableNames` ignora erro de `extractRequiredVars` silenciosamente (nil-coalescing pattern).

### 2.2 `loader.go` (122 linhas) — Camada de Carregamento

**Interface principal:** `TemplateSource` (2 métodos)

| Implementação | Uso |
|--------------|-----|
| `EmbeddedSource` | Templates `//go:embed` no binary |
| `FilesystemSource` | Templates em disco (pasta do usuário ou custom path) |

**Função-chave:** `loadTemplatesFromFS` — helper genérico que itera `fs.ReadDir`, filtra `.md`, lê conteúdo, parseia, seta `IsEmbedded` e `Source`.

**Comportamento de tolerância a falhas:**
- Arquivos não-`.md` são ignorados silenciosamente.
- Erros de leitura (`fs.ReadFile`) são ignorados com `continue`.
- Erros de parse (`parseTemplate` com conteúdo vazio) são ignorados com `continue`.
- Erros de directory são propagados.

### 2.3 `manager.go` (187 linhas) — Camada de Gerenciamento

**Interface pública:** `TemplateManager` (6 métodos)

| Método | Responsabilidade |
|--------|-----------------|
| `ListTemplates()` | Lista todas ordenadas alfabeticamente |
| `GetTemplate(name)` | Busca por nome (thread-safe) |
| `RenderTemplate(name, vars)` | Renderiza com substituição |
| `ValidateTemplate(name)` | Valida conteúdo da template |
| `GetRequiredVariables(name)` | Retorna variáveis exigidas (exclui auto) |
| `GetTemplateNames()` | Lista nomes (sem dados) |
| `HasTemplate(name)` | Verifica existência |

**Struct `Manager`:**
```go
type Manager struct {
    templates map[string]*Template  // mapa principal
    mu        sync.RWMutex          // proteção concurrente
    renderer  *Renderer             // instância do renderer
}
```

**Struct `ManagerConfig`:**
```go
type ManagerConfig struct {
    CustomPath string  // caminho custom opcional
}
```

**Ordenação de fontes (prioridade — última vence):**
1. `embedded` (via `assets.Templates/templates`)
2. `user` (via `$XDG_CONFIG_HOME/shotgun-cli/templates`)
3. `custom` (via `cfg.CustomPath`, se fornecido; `~` expandido)

**Thread-safety:** Todos os métodos públicos usam `RLock`/`RUnlock` exceto `loadFromSources` que usa `Lock`/`Unlock`.

### 2.4 `renderer.go` (186 linhas) — Camada de Renderização

**Struct `Renderer`:**
```go
type Renderer struct {
    defaultVars map[string]string  // contém CURRENT_DATE
}
```

**Pipeline de renderização:**
1. `validateVariables` — verifica obrigatórias (exclui auto-generated)
2. `mergeVariables` — defaultVars → override por vars fornecidas → CURRENT_DATE sempre atualizado
3. `substituteVariables` — `strings.ReplaceAll` iterativo por variável
4. `sanitizeVariableValue` — normaliza `\r\n` → `\n`, remove whitespace trailing

**Comportamento especial:**
- `CURRENT_DATE` é **sempre** atualizado no momento da renderização, mesmo que fornecido pelo chamador.
- `isAutoGeneratedVar` filtra `CURRENT_DATE` das variáveis obrigatórias do usuário.
- `PreviewTemplate` gera `[task]`, `[rules]` como valores placeholder.
- `ValidateVariableNames` valida nomes contra `variablePattern`.

---

## 3. Métricas de Qualidade

| Métrica | Valor |
|---------|-------|
| Linhas de código (prod.) | 732 |
| Linhas de código (test) | 1.590 |
| Razão test/prod | 2.17x |
| Funções públicas (interface) | 7 (`TemplateManager`) + 2 (constructors) |
| Funções package-private | ~15 |
| Receivers de método | 5 (`Template`) |
| Tipos principais | 4 (`Template`, `TemplateSource`, `Manager`, `Renderer`, `ManagerConfig`) |
| Interfaces definidas | 2 (`TemplateSource`, `TemplateManager`) |
| Pacotes importados (prod.) | 10 (std: 8, ext: 2) |
| Pacotes importados (test) | 5 (std: 4, ext: 1) |

### Cobertura de Testes (identificada por leitura)

| Arquivo | Tests | Cases cobertos |
|---------|-------|----------------|
| `template_struct_test.go` | 9 testes | Nome extração, descrição, variáveis, parse, validação, helpers |
| `loader_test.go` | 8 testes | Embedded, filesystem, empty dir, invalid dir, malformed, name extraction, source name |
| `manager_test.go` | 11+ testes | List ordenado, GetTemplate, RenderTemplate (table-driven), ValidateTemplate, GetRequiredVariables, multi-source, priority, graceful degradation, benchmark |
| `renderer_test.go` | 11+ testes | NewRenderer, RenderTemplate (table-driven), Sanitize, Merge, IsAutoGenerated, GetRequiredVars, Preview, ValidateNames, Substitute, ValidateVariables, GetDefaultVariables |

**Total estimado de casos de teste:** ~40+ cases com cobertura abrangente.

---

## 4. Análise de Padrões de Design

| Padrão | Onde aparece | Descrição |
|--------|-------------|-----------|
| **Strategy** | `TemplateSource` interface | Polimorfismo para fontes de templates |
| **Template Method** | `NewManager` → `loadFromSources` | Construtor orquestra carregamento multi-fonte |
| **Factory** | `NewEmbeddedSource`, `NewFilesystemSource`, `NewRenderer` | Construtores coerentes Go idiomatic |
| **Singleton-like** | `NewManager`, `NewRenderer` | Instâncias por chamada (não singletons reais) |
| **Null Object** | `validateVariables` com nil vars | Converte nil para map vazio |
| **Decorator** | `validateVariables` + `mergeVariables` | Pipeline de pré-processamento |
| **Thread-safety** | `sync.RWMutex` em `Manager` | Leitura múltipla, escrita única |

---

## 5. Problemas Identificados / Tech Debt

| ID | Severidade | Descrição | Arquivo | Linha |
|----|-----------|-----------|---------|-------|
| T1 | **Baixa** | `getRequiredVars` ignora erro silenciosamente via `vars, _ :=` | `template.go:168` | `GetVariableNames` |
| T2 | **Baixa** | `extractRequiredVars` tem `//nolint:unparam` — erro reservado não implementado | `loader.go:88` | — |
| T3 | **Média** | `validateTemplateContent` faz verificação redundante (balanceamento + regex match) | `template.go:114-132` | Linha dupla |
| T4 | **Baixa** | `HasVariable` usa `strings.Contains` — pode false-positive se `{TASK}` estiver dentro de `{TASKS}` | `template.go:178` | — |
| T5 | **Baixa** | `CURRENT_DATE` é sempre sobrescrito no `mergeVariables`, mesmo se usuário fornecer valor | `renderer.go:112` | — |
| T6 | **Média** | `loadTemplatesFromFS` silencia erros de leitura individualmente — pode mascaram templates faltando | `loader.go:105` | `continue` |
| T7 | **Baixa** | `renderVariableValue` sanitização não remove caracteres nulos ou `\000` | `renderer.go:148` | — |

---

## 6. Fluxo de Inicialização (NewManager)

```
NewManager(cfg)
  ├─ fs.Sub(assets.Templates, "templates")  → EmbeddedSource
  ├─ os.MkdirAll(XDG_CONFIG_HOME/shotgun-cli/templates) → FilesystemSource("user")
  ├─ cfg.CustomPath != "" → FilesystemSource(baseName)
  └─ loadFromSources([embedded, user, custom])
       ├─ LoadTemplates() para cada fonte
       ├─ Merge por nome (última vence)
       └─ Retorno *Manager
```

---

## 7. Fluxo de Renderização

```
Manager.RenderTemplate(name, vars)
  ├─ GetTemplate(name)
  └─ Renderer.RenderTemplate(template, vars)
       ├─ validateVariables(template, vars)
       │    └─ Verifica RequiredVars (exclui auto-generated)
       ├─ mergeVariables(vars)
       │    ├─ Copia defaultVars → merged
       │    ├─ Override por vars
       │    └─ Refresh CURRENT_DATE
       └─ substituteVariables(content, merged)
            └─ strings.ReplaceAll por cada (varName, sanitizedValue)
```
