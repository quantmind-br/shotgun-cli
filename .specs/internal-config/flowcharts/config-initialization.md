# Fluxograma: Inicialização de Configuração e Metadados

| Item       | Detalhe                                              |
|------------|------------------------------------------------------|
| **Módulo** | `internal/config`                                    |
| **Arquivo** | `metadata.go` (`init()`)                            |
| **Fluxo**  | Inicialização e carregamento dos metadados de configuração |

---

```mermaid
flowchart TD
    Start(["Início: pacote importado"]) --> Init["init() chamado pelo runtime Go"]
    Init --> Build["buildAllMetadata()"]
    
    Build --> Scan9["Cria 9 metadados Scanner"]
    Scan9 --> Scan9A["scanner.max-files: int, 10000, [1, 1000000]"]
    Scan9 --> Scan9B["scanner.max-file-size: size, '1MB'"]
    Scan9 --> Scan9C["scanner.max-memory: size, '500MB'"]
    Scan9 --> Scan9D["scanner.skip-binary: bool, true"]
    Scan9 --> Scan9E["scanner.include-hidden: bool, false"]
    Scan9 --> Scan9F["scanner.include-ignored: bool, false"]
    Scan9 --> Scan9G["scanner.workers: int, 1, [1, 32]"]
    Scan9 --> Scan9H["scanner.respect-gitignore: bool, true"]
    Scan9 --> Scan9I["scanner.respect-shotgunignore: bool, true"]
    
    Scan9I --> Ctx3["Cria 3 metadados Context"]
    Ctx3 --> Ctx3A["context.include-tree: bool, true"]
    Ctx3 --> Ctx3B["context.include-summary: bool, true"]
    Ctx3 --> Ctx3C["context.max-size: size, '10MB'"]
    
    Ctx3C --> Tmpl1["Cria 1 metadado Template"]
    Tmpl1 --> Tmpl1A["template.custom-path: path, ''"]
    
    Tmpl1A --> Out2["Cria 2 metadados Output"]
    Out2 --> Out2A["output.format: enum, 'markdown', [markdown,text]"]
    Out2 --> Out2B["output.clipboard: bool, true"]
    
    Out2B --> Llm6["Cria 6 metadados LLM Provider"]
    Llm6 --> Llm6A["llm.provider: enum, 'gemini', [openai,anthropic,gemini]"]
    Llm6 --> Llm6B["llm.api-key: string, ''"]
    Llm6 --> Llm6C["llm.base-url: url, ''"]
    Llm6 --> Llm6D["llm.model: string, ''"]
    Llm6 --> Llm6E["llm.timeout: timeout, 300, [1, 3600]"]
    Llm6 --> Llm6F["llm.save-response: bool, false"]
    
    Llm6F --> Store["armazena em allMetadata []ConfigMetadata (variável pacote)"]
    Store --> Done(["Fim: allMetadata populado, pronto para consulta"])
    
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef data fill:#f3e5f5,stroke:#7b1fa2
    classDef terminal fill:#e8f5e9,stroke:#388e3c
    class Init,Build,Scan9,Scan9A,Scan9B,Scan9C,Scan9D,Scan9E,Scan9F,Scan9G,Scan9H,Scan9I,Ctx3,Ctx3A,Ctx3B,Ctx3C,Tmpl1,Tmpl1A,Out2,Out2A,Out2B,Llm6,Llm6A,Llm6B,Llm6C,Llm6D,Llm6E,Llm6F,Store process
    class Start,Done terminal
```

---

## 2. Consulta a Metadados

```mermaid
flowchart TD
    Start(["Chamada: GetMetadata(key)"]) --> Search["Itera sobre allMetadata []ConfigMetadata"]
    Search --> Match["key == m.Key ?"]
    Match -->|Sim| Found["Retorna (m, true)"]
    Match -->|Não| Next["Próximo item?"]
    Next -->|Sim| Search
    Next -->|Não| NotFound["Retorna (ConfigMetadata{}, false)"]
    
    Found --> End(["Fim"])
    NotFound --> End
    
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef decision fill:#fff9c4,stroke:#f9a825
    classDef terminal fill:#e8f5e9,stroke:#388e3c
    class Start,End terminal
    class Match decision
    class Search,Found,NotFound,Next process
```

```mermaid
flowchart TD
    Start(["Chamada: GetByCategory(category)"]) --> Iterate["Itera sobre allMetadata"]
    Iterate --> CatMatch["m.Category == category ?"]
    CatMatch -->|Sim| Append["adiciona m ao resultado"]
    CatMatch -->|Não| Next["Próximo item?"]
    Append --> Next
    Next -->|Sim| Iterate
    Next -->|Não| Return["Retorna []ConfigMetadata resultado"]
    
    Return --> End(["Fim"])
    
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef decision fill:#fff9c4,stroke:#f9a825
    classDef terminal fill:#e8f5e9,stroke:#388e3c
    class Start,End terminal
    class CatMatch,Next decision
    class Iterate,Append,Return process
```

```mermaid
flowchart LR
    Start(["Chamada: AllCategories()"]) --> Return["Retorna []ConfigCategory"]
    Return --> Order["[CategoryScanner, CategoryContext, CategoryTemplate, CategoryOutput, CategoryLLM]"]
    Order --> End(["Fim: lista ordenada para exibição na UI"])
    
    classDef process fill:#e1f5fe,stroke:#0288d1
    classDef terminal fill:#e8f5e9,stroke:#388e3c
    class Start,End terminal
    class Return,Order process
```

---

## 3. Fluxo de Inicialização (Diagrama de Sequência)

```mermaid
sequenceDiagram
    participant Main as main.go / cmd/*.go
    participant Runtime as Go Runtime
    participant Meta as metadata.go init()
    participant Builder as buildAllMetadata()
    participant Store as allMetadata (pacote)

    Main->>Runtime: import "shotgun-cli/internal/config"
    Runtime->>Meta: chama init()
    Meta->>Builder: buildAllMetadata()
    Builder->>Store: retorna []ConfigMetadata (21 itens)
    Store->>Meta: armazena em variável pacote
    Meta-->>Runtime: init() concluído
    Runtime-->>Main: pacote pronto para uso
```

---

## 4. Notas

- O fluxo é **unidirecional e imutável** após `init()`: `allMetadata` é um slice populado uma vez e nunca modificado.
- A ordem de construção segue a ordem: Scanner (9) → Context (3) → Template (1) → Output (2) → LLM Provider (6).
- Não há dependência de arquivos externos (YAML, JSON, etc.) para metadados — tudo é código Go puro.
- A UI (`internal/ui/`) depende desse fluxo: `config_wizard.go` chama `AllCategories()` e `AllConfigMetadata()` para renderizar campos.
