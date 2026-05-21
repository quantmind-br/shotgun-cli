# Fluxograma: Validação de Valor de Configuração

| Item       | Detalhe                                                      |
|------------|--------------------------------------------------------------|
| **Módulo** | `internal/config`                                            |
| **Arquivo** | `validator.go` (`ValidateValue`)                             |
| **Fluxo**  | Validação de um valor de configuração para uma chave específica |

---

## 1. Fluxo Principal `ValidateValue`

```mermaid
flowchart TD
    Start(["ValidateValue key, value"]) --> Dispatch["switch key"]
    
    Dispatch --> D1["key == scanner.max-files"]
    D1 --> V1["validateMaxFiles value]"]
    V1 --> R1["Verifica: não é formato de tamanho"]
    R1 --> Sscanf1["fmt.Sscanf '%d'"]
    Sscanf1 --> Pos1["É positivo > 0?"]
    Pos1 -->|Sim| OK1(["nil"])
    Pos1 -->|Não| Err1(["erro: 'must be positive'"])
    Sscanf1 -->|Falha| Err1
    
    Dispatch --> D2["key ∈ max-file-size, context.max-size, max-memory"]
    D2 --> V2["validateSizeFormat value]"]
    V2 --> PSize["utils.ParseSize value)"]
    PSize --> OK2(["nil"])
    PSize -->|Erro| Err2(["erro: 'expected size format'"])
    
    Dispatch --> D3["key ∈ 9 chaves bool"]
    D3 --> V3["validateBooleanValue value]"]
    V3 --> Lower["strings.ToLower"]
    Lower --> BoolOk["É 'true' ou 'false'?"]
    BoolOk -->|Sim| OK3(["nil"])
    BoolOk -->|Não| Err3(["erro: "expected true or false"])
    
    Dispatch --> D4["key == scanner.workers"]
    D4 --> V4["validateWorkers value]"]
    V4 --> Sscanf4["fmt.Sscanf '%d'"]
    Sscanf4 --> Range4["[1, 32]?"]
    Range4 -->|Sim| OK4(["nil"])
    Range4 -->|Não| Err4(["erro: 'must be between 1 and 32'"])
    Sscanf4 -->|Falha| Err4
    
    Dispatch --> D5["key == output.format"]
    D5 --> V5["validateOutputFormat value]"]
    V5 --> Enum5["É 'markdown' ou 'text'?"]
    Enum5 -->|Sim| OK5(["nil"])
    Enum5 -->|Não| Err5(["erro: "expected markdown or text"])
    
    Dispatch --> D6["key == template.custom-path"]
    D6 --> V6["validatePath value]"]
    V6 --> Empty6["vazio?"]
    Empty6 -->|Sim| OK6(["nil"])
    Empty6 -->|Não| Expand["Expande ~/ ?"]
    Expand --> Parent["Dir pai é diretório?"]
    Parent -->|Sim| OK6
    Parent -->|Não| Err6(["erro: "parent is not a directory"])
    
    Dispatch --> D7["key == llm.timeout"]
    D7 --> V7["validateTimeout value]"]
    V7 --> Sscanf7["fmt.Sscanf '%d'"]
    Sscanf7 --> Range7["[1, 3600]?"]
    Range7 -->|Sim| OK7(["nil"])
    Range7 -->|Não| Err7(["erro: 'timeout too large'"])
    Sscanf7 -->|Falha| Err7
    
    Dispatch --> D8["key == llm.provider"]
    D8 --> V8["validateLLMProvider value]"]
    V8 --> Enum8["É 'openai', 'anthropic' ou 'gemini'?"]
    Enum8 -->|Sim| OK8(["nil"])
    Enum8 -->|Não| Err8(["erro: "expected one of: openai, anthropic, gemini"])
    
    Dispatch --> D9["key == llm.api-key"]
    D9 --> OK9(["nil — qualquer string válida"])
    
    Dispatch --> D10["key == llm.base-url"]
    D10 --> V10["validateURL value]"]
    V10 --> Empty10["vazio?"]
    Empty10 -->|Sim| OK10(["nil"])
    Empty10 -->|Não| Prefix["Começa com http:// ou https://?"]
    Prefix -->|Sim| OK10
    Prefix -->|Não| Err10(["erro: "URL must start with http:// or https://"])
    
    Dispatch --> D11["key == llm.model"]
    D11 --> OK11(["nil — validação específica do provedor"])
    
    Dispatch --> Default["default — chave não reconhecida"]
    Default --> OKD(["nil"])
    
    classDef startend fill:#e8f5e9,stroke:#388e3c
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef decision fill:#fff9c4,stroke:#f9a825
    classDef error fill:#ffebee,stroke:#c62828
    class Start,OK1,OK2,OK3,OK4,OK5,OK6,OK7,OK8,OK9,OK10,OK11,OKD startend
    class V1,V2,V3,V4,V5,V6,V7,V8,V10 process
    class R1,Sscanf1,Pos1,Sscanf4,Range4,BoolOk,Enum5,Empty6,Expand,Parent,Sscanf7,Range7,Enum8,Empty10,Prefix,Default decision
    class Err1,Err2,Err3,Err4,Err5,Err6,Err7,Err8,Err10 error
```

---

## 2. Fluxo Detalhado: `validateMaxFiles`

```mermaid
flowchart TD
    Start(["validateMaxFiles value]"]) --> Upper["strings.ToUpper value)"]
    Upper --> Trim["strings.TrimSpace"]
    Trim --> Suffix["Verifica sufixo: GB, MB, KB?"]
    Suffix -->|Sim| SizeErr(["erro: "expected a number, got size format"])
    Suffix -->|Não| Last2["Verifica penúltimo char é dígito + sufixo 'B'?"]
    Last2 -->|Sim| SizeErr
    Last2 -->|Não| Sscanf["fmt.Sscanf '%d'"]
    Sscanf -->|Falha| ParseErr(["erro: "expected a positive integer"])
    Sscanf -->|Sucesso| Positive["> 0?"]
    Positive -->|Não| ZeroErr(["erro: "must be positive"])
    Positive -->|Sim| OK(["nil"])
    
    classDef startend fill:#e8f5e9,stroke:#388e3c
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef decision fill:#fff9c4,stroke:#f9a825
    classDef error fill:#ffebee,stroke:#c62828
    class Start OK startend
    class Upper,Trim,Suffix,Last2,Sscanf process
    class Suffix,Last2,Positive decision
    class SizeErr,ParseErr,ZeroErr error
```

---

## 3. Fluxo Detalhado: `validatePath`

```mermaid
flowchart TD
    Start(["validatePath value]"]) --> Empty["vazio?"]
    Empty -->|Sim| OK(["nil"])
    Empty -->|Não| Home["Começa com ~/?"]
    Home -->|Sim| Expand["os.UserHomeDir()"]
    Expand --> Join["filepath.Join home, rest]"]
    Home -->|Não| Join
    Join --> Dir["filepath.Dir expanded]"]
    Dir --> Special["Dir é '.' ou '/'?"]
    Special -->|Sim| OK
    Special -->|Não| Stat["os.Stat Dir]"]
    Stat -->|Erro| OK
    Stat -->|Sucesso| IsDir["info.IsDir()?"]
    IsDir -->|Sim| OK
    IsDir -->|Não| NotDirErr(["erro: "parent path exists but is not a directory"])
    
    classDef startend fill:#e8f5e9,stroke:#388e3c
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef decision fill:#fff9c4,stroke:#f9a825
    classDef error fill:#ffebee,stroke:#c62828
    class Start OK startend
    class Empty,Home,Expand,Join,Dir,Special,Stat,IsDir process
    class Empty,Home,Special,IsDir decision
    class NotDirErr error
```

---

## 4. Fluxo Detalhado: `validateSizeFormat` (delegado)

```mermaid
flowchart TD
    Start(["validateSizeFormat value]"]) --> Delegate["utils.ParseSize value)"]
    Delegate --> Result["retorna int64, error"]
    Result -->|Erro| Err(["erro: "expected size format"])
    Result -->|Sucesso| OK(["nil"])
    
    classDef startend fill:#e8f5e9,stroke:#388e3c
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef error fill:#ffebee,stroke:#c62828
    class Start OK startend
    class Delegate,Result process
    class Err error
```

---

## 5. Resumo dos Validadores

| Validador              | Chave(s) Aplicáveis                                  | Regra Principal                                         |
|------------------------|------------------------------------------------------|---------------------------------------------------------|
| `validateWorkers`      | `scanner.workers`                                    | Inteiro, [1, 32]                                        |
| `validateMaxFiles`     | `scanner.max-files`                                  | Inteiro positivo, rejeita sufixos de tamanho              |
| `validateSizeFormat`   | `max-file-size`, `context.max-size`, `max-memory`    | Delega `utils.ParseSize()`                               |
| `validateBooleanValue` | 9 chaves booleanas                                    | Case-insensitive `true`/`false`                          |
| `validateOutputFormat` | `output.format`                                      | `markdown` ou `text`                                    |
| `validatePath`         | `template.custom-path`                               | Vazio OK; `~` expandido; parent deve ser dir             |
| `validateTimeout`      | `llm.timeout`                                        | Inteiro, [1, 3600]                                      |
| `validateLLMProvider`  | `llm.provider`                                       | `openai`, `anthropic`, `gemini`                         |
| `validateURL`          | `llm.base-url`                                       | Vazio ou prefixo `http://`/`https://`                    |
| (sem validação)        | `llm.api-key`, `llm.model`                           | Qualquer string                                          |
