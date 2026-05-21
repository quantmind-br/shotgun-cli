# Dicionário de Dados — internal/core/ignore

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/core/ignore`
> **Nível de detalhe:** Detalhado

---

## 1. Tipos Públicos

### IgnoreReason
| Propriedade | Tipo | Descrição |
|---|---|---|
| `IgnoreReason` | `int` (enum) | Motivo pelo qual um caminho foi ignorado. |

**Valores enumerados:**

| Constante | Valor | String() | Semântica |
|---|---|---|---|
| `IgnoreReasonNone` | 0 | `"none"` | Caminho não é ignorado |
| `IgnoreReasonBuiltIn` | 1 | `"built-in"` | Ignorado por padrões embutidos |
| `IgnoreReasonGitignore` | 2 | `"gitignore"` | Ignorado por regras `.gitignore` |
| `IgnoreReasonCustom` | 3 | `"custom"` | Ignorado por regras customizadas |
| `IgnoreReasonExplicit` | 4 | `"explicit"` | Ignorado por exclusão explícita |

**Métodos:**
| Método | Assinatura | Descrição |
|---|---|---|
| `String()` | `func (r IgnoreReason) string` | Retorna representação textual. Fallback `"unknown"` para valores fora do range. |

**Alias deprecated:**
| Alias | Tipo | Descrição |
|---|---|---|
| `Reason` | `IgnoreReason` | Alias mantido para compatibilidade retroativa. `//nolint:revive`. |

---

### IgnoreEngine (Interface)
| Método | Parâmetros | Retorno | Descrição |
|---|---|---|---|
| `ShouldIgnore` | `relPath string` | `(bool, IgnoreReason)` | Verifica se um caminho deve ser ignorado (com motivo). |
| `LoadGitignore` | `rootDir string` | `error` | Carrega regras `.gitignore` recursivamente do diretório raiz. |
| `AddCustomRule` | `pattern string` | `error` | Adiciona um único padrão customizado. |
| `AddCustomRules` | `patterns []string` | `error` | Adiciona múltiplos padrões customizados. Filtra vazios. |
| `AddExplicitExclude` | `pattern string` | `error` | Adiciona um padrão de exclusão explícita (prioridade máxima). |
| `AddExplicitInclude` | `pattern string` | `error` | Adiciona um padrão de inclusão explícita (override). |
| `IsGitignored` | `relPath string` | `bool` | Consulta se o caminho é ignorado apenas por `.gitignore`. |
| `IsCustomIgnored` | `relPath string` | `bool` | Consulta se o caminho é ignorado apenas por custom rules. |
| `LoadShotgunignore` | `rootDir string` | `error` | Carrega regras `.shotgunignore` recursivamente do diretório raiz. |

---

### LayeredIgnoreEngine (Struct)
| Campo | Tipo | Descrição |
|---|---|---|
| `builtInMatcher` | `*gitignore.GitIgnore` | Matcher para padrões embutidos (inicializados no constructor). |
| `gitignoreMatcher` | `*gitignore.GitIgnore` | Matcher para regras `.gitignore`. |
| `customMatcher` | `*gitignore.GitIgnore` | Matcher para padrões customizados acumulados. |
| `explicitExcludes` | `*gitignore.GitIgnore` | Matcher para exclusões explícitas (prioridade máxima). |
| `explicitIncludes` | `*gitignore.GitIgnore` | Matcher para inclusões explícitas (override). |
| `customPatterns` | `[]string` | Lista acumulada de padrões customizados (para recompilação incremental). |
| `explicitExcludePatterns` | `[]string` | Lista acumulada de padrões de exclusão explícita. |
| `explicitIncludePatterns` | `[]string` | Lista acumulada de padrões de inclusão explícita. |

> **Nota:** Todos os campos são privados (camelCase minúsculo). O struct implementa `IgnoreEngine`.

---

## 2. Padrões Built-In (42+ entradas)

Organizados em 11 categorias:

| Categoria | Padrão(s) | Exemplos de matching |
|---|---|---|
| Shotgun | `shotgun-prompt*.md` | `shotgun-prompt.md`, `shotgun-prompt-feature.md` |
| VCS | `.git/`, `.svn/`, `.hg/`, `.bzr/` | `.git/config`, `.svn/entries` |
| IDE/Editor | `.vscode/`, `.idea/`, `*.swp`, `*.swo`, `*~`, `.DS_Store`, `Thumbs.db` | `.vscode/settings.json`, `file.swp` |
| Build/Deps | `node_modules/`, `bower_components/`, `vendor/`, `target/`, `build/`, `dist/`, `out/`, `bin/`, `obj/` | `node_modules/package/index.js` |
| Cache/Temp | `__pycache__/`, `*.pyc`, `*.pyo`, `.cache/`, `.tmp/`, `tmp/`, `.pytest_cache/`, `.mypy_cache/` | `__pycache__/module.pyc`, `.cache/data` |
| Imagens/Mídia | `*.png`, `*.jpg`, `*.jpeg`, `*.gif`, `*.ico`, `*.svg`, `*.webp`, `*.mp3`, `*.mp4`, `*.wav`, `*.avi`, `*.mov`, `*.mkv` | `logo.png`, `video.mp4` |
| Fontes | `*.ttf`, `*.otf`, `*.woff`, `*.woff2`, `*.eot` | `font.ttf` |
| Documentos | `*.pdf`, `*.doc`, `*.docx`, `*.xls`, `*.xlsx`, `*.ppt`, `*.pptx` | `report.pdf` |
| Binários | `*.exe`, `*.dll`, `*.so`, `*.dylib` | `app.exe` |
| Bancos | `*.sqlite`, `*.sqlite3`, `*.db` | `data.sqlite` |
| Logs | `*.log`, `logs/` | `app.log`, `logs/error.log` |
| Pacotes | `*.jar`, `*.war`, `*.nar`, `*.ear`, `*.zip`, `*.tar.gz`, `*.rar`, `*.7z` | `app.jar`, `backup.zip` |
| OS files | `.DS_Store?`, `._*`, `.Spotlight-V100`, `.Trashes`, `ehthumbs.db` | `.DS_Store`, `.Trashes` |

---

## 3. Fluxo de Dados — Camadas de Regras

```
Input: relPath (string, relativo ao root)
         │
         ▼
   filepath.ToSlash(relPath)          ← Normalização cross-platform
         │
         ▼
   ┌──────────────────────────────────────┐
   │         5 CAMADAS DE REGRA           │
   │                                      │
   │  1. Explicit Excludes (MAX)          │  ← MatchesPath → (true, Explicit)
   │  2. Explicit Includes (OVERRIDE)     │  ← MatchesPath → (false, None)
   │  3. Built-In Patterns                │  ← MatchesPath → (true, BuiltIn)
   │  4. Gitignore Patterns               │  ← MatchesPath → (true, Gitignore)
   │  5. Custom Patterns (LOW)            │  ← MatchesPath → (true, Custom)
   └──────────────────────────────────────┘
         │
         ▼
   (false, None)                         ← Nenhum padrão correspondeu
```

### Regra de precedência (documentada e testada)
1. **Explicit exclude** > **explicit include** > **built-in** > **gitignore** > **custom**
2. `explicitIncludes` é o **único override positivo** — retorna `false, IgnoreReasonNone`.
3. Se `explicitIncludes` não matchar, a avaliação continua para as camadas restantes.

---

## 4. Estrutura de Arquivos .gitignore / .shotgunignore

Formato compatível com `.gitignore`:
| Elemento | Sintaxe | Exemplo |
|---|---|---|
| Padrão simples | `<pattern>` | `*.tmp` |
| Diretoria | `<pattern>/` | `build/` |
| Prefixo | `/` + `<pattern>` | `/logs/` |
| Negation | `!` + `<pattern>` | `!important.tmp` |
| Comment | `#` + `<text>` | `# ignorar temporários` |
| Linha vazia | (vazio) | (skip) |

**Tratamento de arquivos aninhados:**
- Arquivos `.gitignore` / `.shotgunignore` em subdiretórios são prefixados com `filepath.Join(relDir, pattern)`.
- Negations (`!`) preservadas: `!"!` + prefix + rest.

---

## 5. Metadados do Módulo

| Propriedade | Valor |
|---|---|
| Package path | `github.com/quantmind-br/shotgun-cli/internal/core/ignore` |
| Go version | 1.24.0 |
| Licença | (inferido: MIT, pelo padrão do ecossistema Go) 🟡 INFERIDO |
| Go doc level | `detalhado` |
| Arquivos de código | `engine.go` (326 linhas), `engine_test.go` (456 linhas) |
| Dependência externa | `github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06` |
| Consumidores | `internal/core/scanner` (FileSystemScanner, tests) |
| Thread-safe | **Não** — sem mutex ou lock |
| Interface stability | `IgnoreEngine` expõe 9 métodos — API estável |
