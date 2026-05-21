# Dicionário de Dados — internal/config

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/config`                                              |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/config`          |
| **Document Language** | Português (pt-br)                                           |

---

## 1. Tipos do Domínio

### 1.1 `ConfigType` (enum)

Tipo inteiro com 8 valores, representando o tipo de dado de uma configuração.

| Valor | Constante    | String  | Descrição                                          |
|-------|--------------|---------|----------------------------------------------------|
| 0     | `TypeString` | `"string"`      | Texto livre                                        |
| 1     | `TypeInt`    | `"int"`         | Número inteiro                                     |
| 2     | `TypeBool`   | `"bool"`        | Booleano (`true` / `false`)                        |
| 3     | `TypeSize`   | `"size"`        | Formato de tamanho: `<número><unidade>` (ex: `10MB`) |
| 4     | `TypePath`   | `"path"`        | Caminho de sistema de arquivos                     |
| 5     | `TypeURL`    | `"url"`         | Endereço URL                                       |
| 6     | `TypeEnum`   | `"enum"`        | Valor de conjunto predefinido                      |
| 7     | `TypeTimeout`| `"timeout"`     | Inteiro positivo (segundos), com faixa mínima/máxima |

**Método**: `func (t ConfigType) String() string` — mapeia o inteiro para sua representação string. Valor fora do range retorna `"unknown"`.

---

### 1.2 `ConfigCategory` (string typed)

Tipo string para agrupamento lógico das configurações.

| Valor              | Constante          | Descrição                              |
|--------------------|--------------------|----------------------------------------|
| `"Scanner"`        | `CategoryScanner`  | Configuração de varredura de arquivos  |
| `"Context"`        | `CategoryContext`  | Configuração de geração de contexto    |
| `"Template"`       | `CategoryTemplate` | Configuração de templates              |
| `"Output"`         | `CategoryOutput`   | Configuração de saída                  |
| `"LLM Provider"`   | `CategoryLLM`      | Configuração do provedor LLM           |

**Ordem de exibição**: definida em `AllCategories()` — Scanner → Context → Template → Output → LLM Provider.

---

## 2. Estruturas de Dados

### 2.1 `ConfigMetadata`

Descreve todos os atributos de uma única chave de configuração.

| Campo            | Tipo               | Obrigatório | Valores Válidos / Descrição                                      |
|------------------|--------------------|-------------|------------------------------------------------------------------|
| `Key`            | `string`           | Sim         | Identificador único da config (ex: `"scanner.max-files"`)         |
| `Category`       | `ConfigCategory`   | Sim         | Um dos 5 valores de `ConfigCategory`                             |
| `Type`           | `ConfigType`       | Sim         | Um dos 8 valores de `ConfigType`                                 |
| `Description`    | `string`           | Sim         | Texto legível em linguagem natural                               |
| `DefaultValue`   | `interface{}`      | Não         | Valor padrão; tipo depende de `Type` (int, bool, string, size)   |
| `EnumOptions`    | `[]string`         | Não         | Lista de valores válidos para `TypeEnum`. Obrigatório se `Type == TypeEnum` |
| `MinValue`       | `int`              | Não         | Limite inferior; usado para `TypeInt` e `TypeTimeout`. 0 = sem limite |
| `MaxValue`       | `int`              | Não         | Limite superior; usado para `TypeInt` e `TypeTimeout`. 0 = sem limite |
| `Required`       | `bool`             | Não         | Indica obrigatoriedade de valor não vazio. **Não usado na validação atual** |

---

## 3. Dicionário de Chaves de Configuração (22 chaves)

Cada entrada contém: chave, categoria, tipo, descrição, valor padrão, opções de enum (se aplicável), faixa (se aplicável).

### 3.1 Scanner (9 chaves)

| #   | Constante                | Chave                          | Tipo       | Descrição                                     | Padrão    | Faixa / Opções                           |
|-----|--------------------------|--------------------------------|------------|-----------------------------------------------|-----------|------------------------------------------|
| 1   | `KeyScannerMaxFiles`     | `scanner.max-files`            | `TypeInt`  | Número máximo de arquivos a escanear          | `10000`   | [1, 1000000]                             |
| 2   | `KeyScannerMaxFileSize`  | `scanner.max-file-size`        | `TypeSize` | Tamanho máximo por arquivo                    | `1MB`     | Formato: `<n><unidade>` (KB, MB, GB)     |
| 3   | `KeyScannerMaxMemory`    | `scanner.max-memory`           | `TypeSize` | Uso máximo de memória para escaneamento       | `500MB`   | Formato: `<n><unidade>`                    |
| 4   | `KeyScannerSkipBinary`   | `scanner.skip-binary`          | `TypeBool` | Pular arquivos binários durante escaneamento  | `true`    | `true` / `false`                         |
| 5   | `KeyScannerIncludeHidden`| `scanner.include-hidden`       | `TypeBool` | Incluir arquivos ocultos (.`*`)               | `false`   | `true` / `false`                         |
| 6   | `KeyScannerIncludeIgnored`| `scanner.include-ignored`     | `TypeBool` | Incluir arquivos ignorados no git             | `false`   | `true` / `false`                         |
| 7   | `KeyScannerWorkers`      | `scanner.workers`              | `TypeInt`  | Número de workers de escaneamento em paralelo | `1`       | [1, 32]                                  |
| 8   | `KeyScannerRespectGitignore`| `scanner.respect-gitignore` | `TypeBool` | Respeitar arquivos `.gitignore`               | `true`    | `true` / `false`                         |
| 9   | `KeyScannerRespectShotgunignore`| `scanner.respect-shotgunignore` | `TypeBool` | Respeitar arquivos `.shotgunignore`       | `true`    | `true` / `false`                         |

### 3.2 LLM Provider (6 chaves)

| #   | Constante             | Chave           | Tipo       | Descrição                                    | Padrão | Faixa / Opções                         |
|-----|-----------------------|-----------------|------------|----------------------------------------------|--------|----------------------------------------|
| 10  | `KeyLLMProvider`      | `llm.provider`  | `TypeEnum` | Provedor LLM a utilizar                      | `gemini` | `["openai", "anthropic", "gemini"]`    |
| 11  | `KeyLLMAPIKey`        | `llm.api-key`   | `TypeString` | Chave de API do provedor LLM                | `""`   | Qualquer string                        |
| 12  | `KeyLLMBaseURL`       | `llm.base-url`  | `TypeURL`  | URL base personalizada para requisições API  | `""`   | Vazio ou `http://`/`https://`          |
| 13  | `KeyLLMModel`         | `llm.model`     | `TypeString` | Nome do modelo LLM a utilizar              | `""`   | Qualquer string                        |
| 14  | `KeyLLMTimeout`       | `llm.timeout`   | `TypeTimeout` | Tempo limite de requisição (segundos)     | `300`  | [1, 3600]                              |
| 15  | `KeyLLMSaveResponse`  | `llm.save-response` | `TypeBool` | Salvar resposta do LLM em arquivo         | `false` | `true` / `false`                       |

### 3.3 Context (3 chaves)

| #   | Constante                | Chave                 | Tipo       | Descrição                                  | Padrão | Faixa / Opções                   |
|-----|--------------------------|-----------------------|------------|--------------------------------------------|--------|----------------------------------|
| 16  | `KeyContextIncludeTree`  | `context.include-tree`| `TypeBool` | Incluir árvore de arquivos no contexto     | `true` | `true` / `false`                 |
| 17  | `KeyContextIncludeSummary`| `context.include-summary` | `TypeBool` | Incluir resumo de arquivos no contexto   | `true` | `true` / `false`                 |
| 18  | `KeyContextMaxSize`      | `context.max-size`    | `TypeSize` | Tamanho máximo do contexto gerado          | `10MB` | Formato: `<n><unidade>`          |

### 3.4 Template (1 chave)

| #   | Constante               | Chave                  | Tipo       | Descrição                                 | Padrão | Faixa / Opções               |
|-----|-------------------------|------------------------|------------|-------------------------------------------|--------|------------------------------|
| 19  | `KeyTemplateCustomPath` | `template.custom-path` | `TypePath` | Caminho personalizado do diretório de templates | `""` | Vazio ou caminho válido     |

### 3.5 Output (2 chaves)

| #   | Constante              | Chave                 | Tipo       | Descrição                                  | Padrão  | Faixa / Opções                |
|-----|------------------------|-----------------------|------------|--------------------------------------------|---------|-------------------------------|
| 20  | `KeyOutputFormat`      | `output.format`       | `TypeEnum` | Formato de saída do contexto gerado        | `markdown` | `["markdown", "text"]`       |
| 21  | `KeyOutputClipboard`   | `output.clipboard`    | `TypeBool` | Copiar contexto gerado para a área de transferência | `true` | `true` / `false`            |

### 3.6 Global (2 chaves)

| #   | Constante       | Chave     | Tipo       | Descrição                                | Padrão | Faixa / Opções          |
|-----|-----------------|-----------|------------|------------------------------------------|--------|-------------------------|
| 22  | `KeyVerbose`    | `verbose` | `TypeBool` | Modoverbose (saída detalhada)            | —      | `true` / `false`        |
| 23  | `KeyQuiet`      | `quiet`   | `TypeBool` | Modo silencioso (sem saída)              | —      | `true` / `false`        |

> **Nota 🟡 INFERIDO**: `verbose` e `quiet` são chaves globais sem valor padrão definido em `metadata.go`. A inferência é de que elas são banderas booleanas de nível CLI (mutuamente exclusivas provavelmente). Os valores padrão dependem da configuração do Viper em `cmd/root.go`.

---

## 4. Tabelas de Lookup

### 4.1 Mapeamento Chave → Tipo

```
scanner.max-files           → TypeInt
scanner.max-file-size       → TypeSize
scanner.max-memory          → TypeSize
scanner.skip-binary         → TypeBool
scanner.include-hidden      → TypeBool
scanner.include-ignored     → TypeBool
scanner.workers             → TypeInt
scanner.respect-gitignore   → TypeBool
scanner.respect-shotgunignore → TypeBool
llm.provider                → TypeEnum
llm.api-key                 → TypeString
llm.base-url                → TypeURL
llm.model                   → TypeString
llm.timeout                 → TypeTimeout
llm.save-response           → TypeBool
context.include-tree        → TypeBool
context.include-summary     → TypeBool
context.max-size            → TypeSize
template.custom-path        → TypePath
output.format               → TypeEnum
output.clipboard            → TypeBool
```

### 4.2 Mapeamento Chave → Categoria

```
Scanner:  scanner.max-files, scanner.max-file-size, scanner.max-memory,
          scanner.skip-binary, scanner.include-hidden, scanner.include-ignored,
          scanner.workers, scanner.respect-gitignore, scanner.respect-shotgunignore

Context:  context.include-tree, context.include-summary, context.max-size

Template: template.custom-path

Output:   output.format, output.clipboard

LLM Provider: llm.provider, llm.api-key, llm.base-url, llm.model,
              llm.timeout, llm.save-response
```

### 4.3 Chaves com Opções Enum

| Chave              | Tipo       | Opções Válidas                  |
|--------------------|------------|---------------------------------|
| `output.format`    | `TypeEnum` | `markdown`, `text`              |
| `llm.provider`     | `TypeEnum` | `openai`, `anthropic`, `gemini` |

### 4.4 Chaves com Faixa Numérica

| Chave                  | Tipo        | Mín   | Máx    |
|------------------------|-------------|-------|--------|
| `scanner.max-files`    | `TypeInt`   | 1     | 1000000|
| `scanner.workers`      | `TypeInt`   | 1     | 32     |
| `llm.timeout`          | `TypeTimeout`| 1    | 3600   |

---

## 5. Regras de Validação por Tipo

| Tipo          | Regra de Validação                                          | Regra de Conversão                     |
|---------------|-------------------------------------------------------------|----------------------------------------|
| `TypeInt`     | `fmt.Sscanf("%d")` + verificação positiva                   | `fmt.Sscanf("%d", &intVal)`           |
| `TypeBool`    | `strings.ToLower` deve ser `"true"` ou `"false"`            | `strings.ToLower == "true"`            |
| `TypeSize`    | Delega para `utils.ParseSize()`                             | — (string mantida)                     |
| `TypePath`    | Vazio OK; `~/` expandido para home; parent deve ser dir     | — (string mantida)                     |
| `TypeURL`     | Vazio OK; senão, prefixo `http://` ou `https://`            | — (string mantida)                     |
| `TypeEnum`    | Valor deve estar em `EnumOptions`                           | — (string mantida)                     |
| `TypeTimeout` | `fmt.Sscanf("%d")` + faixa [1, 3600]                        | `fmt.Sscanf("%d", &intVal)`           |
| `TypeString`  | Sem validação específica                                      | — (string mantida)                     |

---

## 6. Funções de Consulta do Dicionário

| Função                        | Retorna                          | Descrição                                      |
|-------------------------------|----------------------------------|------------------------------------------------|
| `AllConfigMetadata()`         | `[]ConfigMetadata`               | Todos os 21 metadados (Scanner+LLM+Context+Template+Output) |
| `GetMetadata(key string)`     | `(ConfigMetadata, bool)`         | Busca por chave exata                          |
| `GetByCategory(category)`     | `[]ConfigMetadata`               | Filtra por categoria                           |
| `AllCategories()`             | `[]ConfigCategory`               | Lista ordenada das 5 categorias                |
| `ValidKeys()`                 | `[]string`                       | Lista plana das 22 chaves válidas              |
| `IsValidKey(key string)`      | `bool`                           | Verifica existência de chave                   |

---

## 7. Resumo Estatístico

| Métrica                         | Valor |
|---------------------------------|-------|
| Total de chaves de configuração | 22    |
| Categorias                      | 5     |
| Tipos de valor                  | 8     |
| Chaves com valor padrão explícito | 20  |
| Chaves sem padrão definido      | 2 (`verbose`, `quiet`) |
| Chaves com faixa numérica       | 3     |
| Chaves com opções enum          | 2     |
| Chaves booleanas                | 11    |
| Chaves de string livre          | 2 (`llm.api-key`, `llm.model`) |
