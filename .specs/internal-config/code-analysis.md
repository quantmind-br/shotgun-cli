# Análise de Código — internal/config

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/config`                                              |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/config`          |
| **Linguagem**     | Go                                                             |
| **Versão Go**     | 1.24.0                                                         |
| **Arquivos fonte**| `keys.go`, `metadata.go`, `validator.go`                       |
| **Arquivos de test** | `keys_test.go`, `metadata_test.go`, `validator_test.go`     |
| **Dep. externas** | `github.com/quantmind-br/shotgun-cli/internal/utils`           |
| **Testes**        | `github.com/stretchr/testify`                                  |

---

## 1. Visão Geral

O pacote `internal/config` é a **camada de configuração centralizada** do shotgun-cli. Ele fornece três responsabilidades distintas, refletidas em três arquivos de código-fonte:

1. **Constantes de chaves** (`keys.go`) — 22 constantes `const` para todas as chaves de configuração usadas em todo o código-base.
2. **Metadados** (`metadata.go`) — Tipos (`ConfigType`, `ConfigCategory`, `ConfigMetadata`) e uma lista imutável (inicializada via `init()`) que descreve tipo, categoria, descrição, valor padrão, opções enum e limites de faixa para cada chave.
3. **Validação e conversão** (`validator.go`) — Funções `IsValidKey()`, `ValidateValue()`, `ConvertValue()` e validadores privados (ex: `validateSizeFormat`, `validateBooleanValue`, etc.).

**Dependência direta externa:** apenas `internal/utils` (função `ParseSize`). **Dependência transitiva:** zero pacotes externos de terceiros além do `testify` em testes.

---

## 2. Arquivo `keys.go`

### 2.1 Estrutura

- **Uma única declaração `const`** com 22 constantes, organizadas em 6 blocos de comentário:
  - Scanner (9 chaves)
  - LLM (6 chaves)
  - Context (3 chaves)
  - Template (1 chave)
  - Output (2 chaves)
  - Global (2 chaves)

### 2.2 Padrão

- Todas as chaves seguem a convenção de nomeação: `Key<Nome> = "categoria.sub-chave"`.
- As chaves globais (`verbose`, `quiet`) são exceções: não usam prefixo de categoria nem ponto.
- **Nenhuma lógica**, apenas dados.

### 2.3 Testes (`keys_test.go`)

| Teste                          | Propósito                                                  |
|--------------------------------|------------------------------------------------------------|
| `TestNoDuplicateKeyValues`     | Garante que nenhum valor de chave é duplicado              |
| `TestKeyNamingConvention`      | Verifica prefixo `Key` em todos os nomes de constantes     |
| `TestKeyValueFormat`           | Valida ponto `.` nos valores (exceto `verbose`/`quiet`)    |
| `TestAllKeysDocumented`        | Assegura ≥ 22 chaves                                       |
| `TestScannerKeysExist`         | Lista esperada de chaves scanner com prefixo `scanner.`    |
| `TestLLMKeysExist`             | Lista esperada de chaves LLM com prefixo `llm.`            |

---

## 3. Arquivo `metadata.go`

### 3.1 Tipos Definidos

#### `ConfigType` (enum inteiro)

| Constante     | Valor | String  | Descrição                                   |
|---------------|-------|---------|---------------------------------------------|
| `TypeString`  | 0     | `string`| Valor string livre                          |
| `TypeInt`     | 1     | `int`   | Número inteiro                              |
| `TypeBool`    | 2     | `bool`  | Booleano (`true`/`false`)                   |
| `TypeSize`    | 3     | `size`  | Formato de tamanho (`10MB`, `500KB`)        |
| `TypePath`    | 4     | `path`  | Caminho de sistema de arquivos              |
| `TypeURL`     | 5     | `url`   | URL para endpoints HTTP                     |
| `TypeEnum`    | 6     | `enum`  | Valor de um conjunto predefinido            |
| `TypeTimeout` | 7     | `timeout`| Inteiro positivo (segundos) com faixa       |

Método: `func (t ConfigType) String() string` — retorna representação string; caso default retorna `"unknown"`.

#### `ConfigCategory` (string typed)

| Constante         | Valor          | Descrição                                      |
|-------------------|----------------|------------------------------------------------|
| `CategoryScanner` | `"Scanner"`    | Configuração de varredura de arquivos          |
| `CategoryContext` | `"Context"`    | Configuração de geração de contexto            |
| `CategoryTemplate`| `"Template"`   | Configuração de templates                      |
| `CategoryOutput`  | `"Output"`     | Configuração de saída                          |
| `CategoryLLM`     | `"LLM Provider"`| Configuração do provedor LLM                  |

#### `ConfigMetadata` (struct)

| Campo          | Tipo               | Obrigatório | Descrição                                           |
|----------------|--------------------|-------------|-----------------------------------------------------|
| `Key`          | `string`           | Sim         | Chave (ex: `"scanner.max-files"`)                   |
| `Category`     | `ConfigCategory`   | Sim         | Agrupamento lógico                                  |
| `Type`         | `ConfigType`       | Sim         | Tipo para validação e UI                            |
| `Description`  | `string`           | Sim         | Descrição em linguagem natural                       |
| `DefaultValue` | `interface{}`      | Não         | Valor padrão                                        |
| `EnumOptions`  | `[]string`         | Não         | Opções válidas para `TypeEnum`                      |
| `MinValue`     | `int`              | Não         | Limite inferior para `TypeInt`/`TypeTimeout`        |
| `MaxValue`     | `int`              | Não         | Limite superior para `TypeInt`/`TypeTimeout`        |
| `Required`     | `bool`             | Não         | Indica se a chave deve ter valor não vazio          |

### 3.2 Funções Exportadas

| Função                    | Retorno                        | Descrição                                      |
|---------------------------|--------------------------------|------------------------------------------------|
| `AllConfigMetadata()`     | `[]ConfigMetadata`             | Retorna lista imutável com todos os 21 metadados|
| `GetMetadata(key string)` | `(ConfigMetadata, bool)`       | Busca metadado por chave                        |
| `GetByCategory(category)` | `[]ConfigMetadata`             | Filtra metadados por categoria                  |
| `AllCategories()`         | `[]ConfigCategory`             | Retorna 5 categorias na ordem de exibição       |

### 3.3 Inicialização

- `init()` chama `buildAllMetadata()` que retorna um literal slice de 21 `ConfigMetadata`.
- O slice é armazenado em uma variável pacote `allMetadata` (não exportada).
- **Não é mutável em runtime** — a lista é fixa após `init()`.

### 3.4 Testes (`metadata_test.go`)

| Teste                              | Propósito                                                    |
|------------------------------------|--------------------------------------------------------------|
| `TestConfigType_String`            | Verifica representação string de cada `ConfigType`           |
| `TestAllConfigMetadata_ReturnsAllKeys` | Assegura ≥ 1 metadado                                      |
| `TestAllConfigMetadata_MatchesValidKeys` | Garante correspondência 1:1 com `ValidKeys()`           |
| `TestAllConfigMetadata_NoDuplicateKeys` | Detecta chaves duplicadas                              |
| `TestAllConfigMetadata_AllHaveCategories` | Todas as chaves têm categoria                      |
| `TestAllConfigMetadata_AllHaveDescriptions` | Todas as chaves têm descrição                   |
| `TestAllConfigMetadata_AllHaveTypes` | Todos os tipos estão em `ConfigType` válido               |
| `TestAllConfigMetadata_EnumTypesHaveOptions` | Chaves enum têm `EnumOptions` não vazias             |
| `TestAllConfigMetadata_IntTypesHaveRanges` | Chaves int/timeout têm faixa válida                    |
| `TestGetMetadata_*`                | Busca por chave existente/não existente                     |
| `TestGetByCategory_*`              | Filtragem por categoria                                      |
| `TestAllCategories_*`              | Retorno e cobertura de todas as categorias                   |
| `TestMetadataDefaults_MatchRootDefaults` | Confirma valores padrão com mapa esperado              |
| `TestMetadataEnumOptions_MatchValidators` | Confirma opções enum com validadores                   |
| `TestMetadataRanges_MatchValidators` | Confirma faixas int com validadores                      |

---

## 4. Arquivo `validator.go`

### 4.1 Funções Exportadas

| Função                    | Assinatura                                      | Descrição                                          |
|---------------------------|-------------------------------------------------|----------------------------------------------------|
| `ValidKeys()`             | `() []string`                                   | Retorna fatia com 22 chaves válidas                |
| `IsValidKey(key string)`  | `(bool)`                                        | Verifica se `key` está em `ValidKeys()`            |
| `ValidateValue(key, value string)` | `(error)`                              | Valora string `value` para a `key` dada            |
| `ConvertValue(key, value string)`  | `(interface{}, error)`                    | Converte `value` para tipo Go apropriado           |

### 4.2 Validadores Privados

| Função                  | Chaves Aplicáveis                                | Regra de Validação                                  |
|-------------------------|--------------------------------------------------|-----------------------------------------------------|
| `validateWorkers`       | `scanner.workers`                                | Inteiro positivo, faixa [1, 32]                     |
| `validateMaxFiles`      | `scanner.max-files`                              | Inteiro positivo (>0), rejeita formatos de tamanho   |
| `validateSizeFormat`    | `max-file-size`, `context.max-size`, `max-memory` | Delega para `utils.ParseSize()`                    |
| `validateBooleanValue`  | Todos os `bool` (9 chaves)                       | `"true"` ou `"false"` (case-insensitive)            |
| `validateOutputFormat`  | `output.format`                                  | `"markdown"` ou `"text"`                             |
| `validatePath`          | `template.custom-path`                           | Aceita vazio; expande `~`; verifica parent dir é dir |
| `validateTimeout`       | `llm.timeout`                                    | Inteiro positivo, faixa [1, 3600]                    |
| `validateLLMProvider`   | `llm.provider`                                   | `"openai"`, `"anthropic"`, `"gemini"`               |
| `validateURL`           | `llm.base-url`                                   | Vazio OK; senão, prefixo `http://` ou `https://`    |

### 4.3 Fluxo `ValidateValue`

```
ValidateValue(key, value)
  ├─ switch key
  ├─ case scanner.max-files       → validateMaxFiles(value)
  ├─ case max-file-size, context.max-size, max-memory → validateSizeFormat(value)
  ├─ case 9 chaves bool           → validateBooleanValue(value)
  ├─ case scanner.workers         → validateWorkers(value)
  ├─ case output.format           → validateOutputFormat(value)
  ├─ case template.custom-path    → validatePath(value)
  ├─ case llm.timeout             → validateTimeout(value)
  ├─ case llm.provider            → validateLLMProvider(value)
  ├─ case llm.api-key             → nil (qualquer string é válida)
  ├─ case llm.base-url            → validateURL(value)
  ├─ case llm.model               → nil (validação específica do provedor)
  └─ default                      → nil (chave não reconhecida)
```

### 4.4 Fluxo `ConvertValue`

```
ConvertValue(key, value)
  ├─ case scanner.max-files, workers, llm.timeout → fmt.Sscanf("%d")
  ├─ case 9 chaves bool → strings.ToLower(value) == "true"
  └─ default → retorna value como string
```

### 4.5 Testes (`validator_test.go`)

| Teste                        | Propósito                                                |
|------------------------------|----------------------------------------------------------|
| `TestIsValidKey`             | Chaves válidas/inválidas                                 |
| `TestValidKeys`              | Retorna lista não vazia sem duplicatas                    |
| `TestValidateValue_Workers`  | Casos: 1-32 válidos, 0/33/-1 inválidos, não-numérico     |
| `TestValidateValue_MaxFiles` | Números válidos, zero/inválido, formatos de tamanho rejeitados |
| `TestValidateValue_SizeFormat`| 1MB/500KB/1GB válidos, abc inválido                     |
| `TestValidateValue_Boolean`  | true/false (case-insensitive), yes/no/1/0 inválidos      |
| `TestValidateValue_LLMProvider`| openai/anthropic/gemini válidos, invalid inválido     |
| `TestValidateValue_OutputFormat`| markdown/text válidos, json inválido                 |
| `TestValidateValue_Timeout`  | 1-3600 válidos, 0/3601/inválido                          |
| `TestValidateValue_URL`      | http(s) válidos, ftp/sem-scheme inválidos                |
| `TestConvertValue_Integer`   | Conversão "1000" → 1000                                  |
| `TestConvertValue_Boolean`   | "true"/"TRUE" → true, "false"/"FALSE" → false            |
| `TestConvertValue_String`    | "openai" → "openai"                                      |
| `TestValidatePath`           | Cenarios: vazio, caminhos válidos, parent inexistente     |
| `TestValidatePath_ExistingFile` | Testa com arquivo temporário real                      |
| `TestValidateValue_Path`     | template.custom-path com caminhos válidos                |

---

## 5. Acoplamento e Dependências

### 5.1 Dependência Direta (read)

| Pacote               | Uso                                                        |
|----------------------|------------------------------------------------------------|
| `internal/utils`     | `ParseSize(value)` em `validateSizeFormat()`               |

### 5.2 Pacotes Dependentes (write — quem importa `internal/config`)

| Pacote                         | Uso Principal                                                    |
|--------------------------------|------------------------------------------------------------------|
| `cmd/root.go`                  | `config.AllCategories()`, `config.AllConfigMetadata()`, keys    |
| `cmd/config.go`                | `config.IsValidKey()`, `config.ValidateValue()`, keys           |
| `cmd/config_llm.go`            | `config.KeyLLMProvider` e demais keys LLM                       |
| `cmd/send.go`                  | `config.KeyLLMTimeout`, `config.KeyLLMSaveResponse`             |
| `cmd/context.go`               | `config.KeyScannerWorkers`, `config.KeyContextMaxSize`, etc.    |
| `cmd/llm.go`                   | `config.KeyLLMProvider`, `config.KeyLLMAPIKey`, etc.            |
| `cmd/template.go`              | `config.KeyTemplateCustomPath`                                   |
| `internal/ui/config_wizard.go` | `config.AllConfigMetadata()`, `config.AllCategories()`, `GetMetadata()` |
| `internal/ui/config_wizard_test.go` | Mock de metadados para testes do wizard                    |
| `internal/ui/screens/config_category.go` | Renderização de categorias e campos de config            |
| `internal/ui/screens/config_category_test.go` | Testes da tela de categoria                      |
| `internal/ui/screens/template_selection.go` | `config.KeyTemplateCustomPath`                      |
| `internal/ui/components/config_toggle.go` | `config.GetMetadata()` para validação de toggles         |
| `cmd/llm_test.go`              | `config.KeyLLMAPIKey` no teste                                  |

### 5.3 Grafo Simplificado

```
internal/config
    ├── read ──► internal/utils.ParseSize
    └── exported ◄── cmd/root.go
                   ├── cmd/config.go
                   ├── cmd/config_llm.go
                   ├── cmd/send.go
                   ├── cmd/context.go
                   ├── cmd/llm.go
                   ├── cmd/template.go
                   ├── cmd/llm_test.go
                   ├── internal/ui/config_wizard.go
                   ├── internal/ui/config_wizard_test.go
                   ├── internal/ui/screens/config_category.go
                   ├── internal/ui/screens/config_category_test.go
                   ├── internal/ui/screens/template_selection.go
                   └── internal/ui/components/config_toggle.go
```

---

## 6. Métricas de Código

| Métrica              | Valor |
|----------------------|-------|
| Arquivos Go          | 3     |
| Arquivos de teste    | 3     |
| Constantes `const`   | 22    |
| Tipos definidos      | 3 (`ConfigType`, `ConfigCategory`, `ConfigMetadata`) |
| Funções exportadas   | 7 (`AllConfigMetadata`, `GetMetadata`, `GetByCategory`, `AllCategories`, `ValidKeys`, `IsValidKey`, `ValidateValue`, `ConvertValue`) |
| Funções privadas     | 9 (`validateWorkers`, `validateMaxFiles`, `validateSizeFormat`, `validateBooleanValue`, `validateOutputFormat`, `validatePath`, `validateTimeout`, `validateLLMProvider`, `validateURL`) |
| Total de linhas (est.) | ~470 (código) + ~430 (testes) |
| Dependências externas | 0 (apenas `internal/utils` interna) |
| Testes unitários     | 17+ funções de teste (keys) + 15+ (metadata) + 14+ (validator) |

---

## 7. Pontos de Atenção (Code Review)

| # | Observação                                          | Severidade | Localização          |
|---|-----------------------------------------------------|------------|----------------------|
| 1 | `ValidateValue` retorna `nil` para chaves não mapeadas (default) — silenciar erros. | 🟡 Baixa | `validator.go:ValidateValue` |
| 2 | `ValidateValue` retorna `nil` para `llm.api-key` — não há validação de comprimento/formato. | 🟡 Baixa | `validator.go` |
| 3 | `ConvertValue` usa `fmt.Sscanf(value, "%d")` em vez de `strconv.Atoi` — funcional mas menos idiomático. | 🟡 Baixa | `validator.go:ConvertValue` |
| 4 | `validateMaxFiles` detecta formatos de tamanho com heurística manual (sufixos) em vez de delegar a `utils.ParseSize`. | 🟡 Baixa | `validator.go:validateMaxFiles` |
| 5 | `ConfigMetadata.Required` está definido mas nunca usado na validação atual. | 🟡 Baixa | `metadata.go:ConfigMetadata` |
| 6 | `ValidateValue` default retorna `nil` — se uma chave válida for omitida da tabela switch, a validação passa silenciosamente. | 🔴 Média | `validator.go:ValidateValue` |
| 7 | `ValidateValue` tem 22 cases explícitos — manutenção futura requer sincronização manual entre `keys.go`, `validator.go` e `metadata.go`. | 🟡 Baixa | `validator.go:ValidateValue` |
| 8 | `metadata_test.go` referencia `ValidKeys()` de `validator.go` — acoplamento entre metadata e validator (mesmo pacote, aceitável). | ℹ️ Info | `metadata_test.go` |

---

## 8. Cobertura de Testes (por função)

| Função                   | Cobertura Inferida | Comentários                        |
|--------------------------|--------------------|-------------------------------------|
| `keys.go` (const)        | ✅ Alta            | 6 testes de propriedade            |
| `ConfigType.String()`    | ✅ Alta            | 9 casos cobrindo todos os tipos     |
| `AllConfigMetadata()`    | ✅ Alta            | Verifica contagem, duplicatas, tipos|
| `GetMetadata()`          | ✅ Alta            | Chave existente + não existente     |
| `GetByCategory()`        | ✅ Alta            | 5 categorias + categoria inexistente|
| `AllCategories()`        | ✅ Alta            | Ordem, contagem, cobertura completa |
| `ValidKeys()`            | ✅ Alta            | Não vazio, sem duplicatas           |
| `IsValidKey()`           | ✅ Alta            | 6 casos (válidos e inválidos)       |
| `ValidateValue()`        | ✅ Alta            | ~20 casos cobrindo todos os validadores |
| `ConvertValue()`         | ✅ Média           | 3 casos (int, bool, string)         |
| `validatePath()`         | ✅ Alta            | 6 casos incluindo temp file real    |
| `validateSizeFormat()`   | ✅ Alta            | Delegado a `utils.ParseSize`        |

**Cobertura geral estimada: ≥ 90%** 🟢

---

## 9. Observações Gerais

- O pacote `internal/config` segue o princípio de **separação de responsabilidades**: constantes → metadados → validação.
- A inicialização via `init()` em `metadata.go` é uma decisão arquitetural válida para Go: garante que `allMetadata` esteja populado antes de qualquer uso, sem necessidade de chamadas explícitas de inicialização.
- A correspondência entre `ValidKeys()` (validator), `ConfigMetadata` (metadata) e as constantes (keys) é verificada em testes cruzados.
- O default de `llm.provider` é `"gemini"` (conforme `metadata.go`), consistente com a superfície do sistema.
- O pacote é **low-coupling**: depende apenas de `internal/utils` e é consumido por ~13 pacotes em `cmd/` e `internal/ui/`.
