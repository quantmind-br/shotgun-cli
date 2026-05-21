# Dicionário de Dados — Módulo `internal/core/template`

> **Pacote:** `github.com/quantmind-br/shotgun-cli/internal/core/template`
> **Nível:** `detalhado`

---

## 1. Tipos Principais

### 1.1 `Template` (struct)

Estrutura de dados que representa uma template markdown com metadados.

| Campo | Tipo | Visibilidade | JSON Tag | Obrigatório | Descrição |
|-------|------|-------------|----------|-------------|-----------|
| `Name` | `string` | pública | `name` | Sim | Nome identificador da template (sem extensão, sem prefixo `prompt_`) |
| `Description` | `string` | pública | `description` | Não | Descrição extraída do primeiro cabeçalho ou comentário HTML |
| `Content` | `string` | pública | `content` | Sim | Conteúdo bruto da template markdown |
| `RequiredVars` | `[]string` | pública | `required_vars` | Não | Lista de variáveis obrigatórias encontradas no conteúdo |
| `FilePath` | `string` | pública | `file_path` | Não | Caminho do arquivo de origem |
| `IsEmbedded` | `bool` | pública | `is_embedded` | Não | `true` se carregado do filesystem embedded |
| `Source` | `string` | pública | `source` | Não | Origem: `"embedded"`, `"user"`, ou nome do diretório custom |

**Invariants:**
- `Name` ≠ `""`
- `Content` ≠ `""`
- Se `IsEmbedded == true` então `Source == "embedded"`

**Regra de negócio (inferida):** Templates com `prompt_` no nome do arquivo têm o prefixo removido no nome final.

---

### 1.2 `TemplateSource` (interface)

Interface de estratégia para carregamento de templates a partir de diferentes fontes.

| Método | Retorno | Descrição |
|--------|---------|-----------|
| `LoadTemplates()` | `(map[string]*Template, error)` | Carrega todas templates da fonte |
| `GetSourceName()` | `string` | Retorna nome legível da fonte |

**Implementações concretas:** `EmbeddedSource`, `FilesystemSource`

---

### 1.3 `EmbeddedSource` (struct)

Carrega templates do filesystem embedded (`go:embed`).

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `fsys` | `fs.FS` | privada | Filesystem Go embed |

**Regra de negócio:** Aponta para `assets.Templates/templates` (subdiretório do embed).

---

### 1.4 `FilesystemSource` (struct)

Carrega templates de um diretório no filesystem do sistema.

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `path` | `string` | privada | Caminho absoluto do diretório |
| `sourceName` | `string` | privada | Nome da fonte (para metadados) |

**Regra de negócio:** Usa `os.DirFS(path)` para leitura.

---

### 1.5 `Manager` (struct)

Gerenciador principal de templates com thread-safety.

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `templates` | `map[string]*Template` | privada | Mapa principal nome → template |
| `mu` | `sync.RWMutex` | privada | Semaforo de leitura múltipla / escrita única |
| `renderer` | `*Renderer` | privada | Instância de renderização |

---

### 1.6 `ManagerConfig` (struct)

Configuração opcional para `NewManager`.

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `CustomPath` | `string` | pública | Caminho de templates custom (expansão `~` automática) |

---

### 1.7 `Renderer` (struct)

Responsável pela substituição de variáveis em templates.

| Campo | Tipo | Visibilidade | Descrição |
|-------|------|-------------|-----------|
| `defaultVars` | `map[string]string` | privada | Variáveis padrão (atualmente só `CURRENT_DATE`) |

---

## 2. Constantes

### 2.1 Variáveis Padrão

| Constante | Valor | Tipo | Descrição |
|-----------|-------|------|-----------|
| `VarTask` | `"TASK"` | `const` | Variável para descrever a tarefa |
| `VarRules` | `"RULES"` | `const` | Variável para regras de comportamento |
| `VarFileStructure` | `"FILE_STRUCTURE"` | `const` | Variável para estrutura de arquivos |
| `VarCurrentDate` | `"CURRENT_DATE"` | `const` | Variável para data atual (auto-generated) |

### 2.2 Fonte Embedded

| Constante | Valor | Tipo | Descrição |
|-----------|-------|------|-----------|
| `sourceEmbedded` | `"embedded"` | `const` (package-private) | Nome da fonte embedded |

---

## 3. Padrão de Variáveis

### 3.1 RegEx de Variável

```
\{([A-Z_][A-Z0-9_]*)\}
```

**Regras do nome de variável:**
- Começa com letra MAIÚSCULA (`A-Z`) ou underscore (`_`)
- Seguido por zero ou mais caracteres: maiúsculas, dígitos, underscores
- Delimitado por `{` e `}`
- Minúsculas e mistas não são reconhecidas (ex: `{task}` não match)

### 3.2 Variáveis Auto-geradas

| Variável | Fonte | Substituição |
|----------|-------|-------------|
| `CURRENT_DATE` | `Renderer.mergeVariables` | `time.Now().Format("2006-01-02")` → `"2026-05-20"` |

**Regra de negócio (inferida):** `CURRENT_DATE` é sempre sobrescrito no momento da renderização, mesmo que o chamador forneça um valor diferente.

---

## 4. Mapeamento de Fontes → Caminhos

| Fonte | Caminho | `IsEmbedded` | `Source` |
|-------|---------|-------------|----------|
| Embedded | `assets.Templates/templates/` | `true` | `"embedded"` |
| User | `$XDG_CONFIG_HOME/shotgun-cli/templates/` | `false` | `"user"` |
| Custom | `cfg.CustomPath` (com expansão `~`) | `false` | `filepath.Base(cfg.CustomPath)` |

**Ordem de prioridade (última vence):**
1. Embedded
2. User
3. Custom

---

## 5. Fluxo de Extração de Metadados

### 5.1 `extractTemplateName(fileName)`

| Input | Processamento | Output |
|-------|--------------|--------|
| `"prompt_review.md"` | Remove `.md` → `"prompt_review"` → Remove prefixo `prompt_` | `"review"` |
| `"analyzeBug.md"` | Remove `.md` → `"analyzeBug"` | `"analyzeBug"` |
| `"template.test.md"` | Remove `.md` → `"template.test"` | `"template.test"` |

### 5.2 `extractDescription(content, fileName)`

| Input Início do Conteúdo | Comportamento | Output |
|-------------------------|--------------|--------|
| `# Code Review Template` | Usa primeiro `#` header | `"Code Review Template"` |
| `<!-- Template description -->` | Usa comentário HTML | `"Template description"` |
| `## Secondary Title` | Usa primeiro `#` ou `##` header | `"Secondary Title"` |
| (nenhum header/comentário) | Fallback para nome do arquivo | `"Template for <name>"` |

### 5.3 `extractRequiredVars(content)`

Varredura regex sobre o conteúdo completo → mapa deduplicado → array de strings.

---

## 6. Sanitização de Variáveis

| Operação | Detalhes |
|----------|---------|
| `\r\n` → `\n` | Normaliza CRLF para LF |
| `\r` → `\n` | Normaliza CR antigo para LF |
| Trimming trailing | Remove espaços/tabs no fim de cada linha |
| Preserva leading | Mantém indentação |
| Não faz | Remoção de caracteres nulos, escape HTML, sanitização XSS |

---

## 7. Mapeamento de Arquivo → Template (Exemplo)

Dado o arquivo `assets/templates/prompt_makePlan.md`:

| Campo | Valor |
|-------|-------|
| `Name` | `"makePlan"` (prefixo `prompt_` removido) |
| `IsEmbedded` | `true` |
| `Source` | `"embedded"` |
| `FilePath` | `"templates/prompt_makePlan.md"` |
| `Description` | Primeiro `# Header` ou `<!-- -->` no arquivo |
| `RequiredVars` | Array deduplicado de variáveis match regex |

---

## 8. Interface `TemplateManager` — Contrato

| Método | Parâmetros | Retorno | Invariants / Regras |
|--------|-----------|---------|---------------------|
| `ListTemplates()` | — | `([]Template, error)` | Retorna ordenado alfabeticamente por `Name` |
| `GetTemplate(name)` | `name string` | `(*Template, error)` | Retorna erro se template não existe |
| `RenderTemplate(name, vars)` | `name, vars` | `(string, error)` | Valida vars obrigatórias, substitui todas, retorna texto |
| `ValidateTemplate(name)` | `name` | `error` | Retorna nil se válido; erro com detalhes se não |
| `GetRequiredVariables(name)` | `name` | `([]string, error)` | Exclui variáveis auto-generated do resultado |
| `GetTemplateNames()` | — | `[]string` | Lista apenas nomes (sem dados completos) |
| `HasTemplate(name)` | `name` | `bool` | `true` se template existe |
