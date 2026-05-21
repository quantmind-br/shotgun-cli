# Fluxo: Coleta de Conteúdo de Arquivos

| Campo | Valor |
|-------|-------|
| **Módulo** | `internal/core/contextgen` |
| **Arquivo fonte** | `content.go` |
| **Funções envolvidas** | `collectFileContents`, `walkSelectedNodes`, `peekFileHeader`, `readFileContent`, `isTextFile`, `detectLanguage`, `shouldSkipFile` |
| **Nível de detalhe** | detalhado |

---

## Mermaid

```mermaid
flowchart TD
    Start([Início: collectFileContents]) --> Init["files = [], totalSize = 0, fileCount = 0"]
    Init --> Walk["walkSelectedNodes(root, fn)"]

    subgraph DFS [DFS Recursivo sobre FileNode]
        Walk --> ProcessNode["fn(node) chamado"]
        ProcessNode --> IsDir["node.IsDir?"]

        IsDir -->|Sim| Children
        IsDir -->|Não| CheckSelection

        Children["for child in node.Children: walkSelectedNodes(child, fn)"] --> EndDFS

        CheckSelection["selections != nil && !selections[node.Path]?"] -->|Sim| Skip["skip (não selecionado)"]
        CheckSelection -->|Não| CheckMaxFiles

        CheckMaxFiles["fileCount >= config.MaxFiles?"] -->|Sim| ErrFileCount["ERROR: 'maximum file count exceeded: %d'"]
        CheckMaxFiles -->|Não| CheckFileSize

        CheckFileSize["shouldSkipFile(node, config)?\nnode.IsDir || node.Size > MaxFileSize"] -->|Sim| Skip
        CheckFileSize -->|Não| CheckBinary

        CheckBinary["config.SkipBinary?"] -->|Sim| Peek
        CheckBinary -->|Não| ReadFile

        Peek["peekFileHeader(path) → header\n→ isTextFile(string(header))"] --> IsText{"isTextFile?\n(0 bytes? → true)\n(1024 bytes? truncado)\n(contém 0x00? → false)\n(UTF-8 válido? → false)"}

        IsText -->|Sim| ReadFile
        IsText -->|Não, binário| Skip

        ReadFile["readFileContent(path) → string\nos.Open + io.ReadAll"] --> CheckTotalSize

        Skip --> NextNode
        NextNode["próximo nó no DFS"] --> Children

        CheckTotalSize["totalSize + len(content) > MaxTotalSize?"] -->|Sim| ErrTotalSize["ERROR: 'cumulative content size exceeds total size limit: %d + %d > %d'"]
        CheckTotalSize -->|Não| BuildContent

        BuildContent["FileContent{\n  Path = node.Path,\n  RelPath = filepath.Rel(root.Path, node.Path),\n  Language = detectLanguage(node.Name),\n  Content = content,\n  Size = int64(len(content))\n}"] --> Append["files = append(files, FileContent)\ntotalSize += FileContent.Size\nfileCount++"]
        Append --> NextNode
    end

    ErrFileCount --> EndErr([Erro retornado])
    ErrTotalSize --> EndErr

    EndDFS --> CheckWalkError["err != nil?"]
    CheckWalkError -->|Sim| EndErr
    CheckWalkError -->|Não| ReturnFiles["return files, nil"]

    ReturnFiles --> EndOk([Fim: []FileContent])

    style Start fill:#e1f5fe
    style EndOk fill:#c8e6c9
    style EndErr fill:#ffcdd2
    style ErrFileCount fill:#ffcdd2
    style ErrTotalSize fill:#ffcdd2
```

---

## Detalhamento das Etapas

### Etapa A: Inicialização
- `files` = fat slice vazia de `FileContent`
- `totalSize` = 0 (acumulador de bytes)
- `fileCount` = 0 (contador)

### Etapa B: DFS sobre a Árvore
`walkSelectedNodes` é uma função recursiva que:
1. Chama `fn(node)` para processar o nó atual.
2. Se `node.IsDir`, itera `node.Children` e chama recursivamente para cada filho.

### Etapa C: Filtros Aplicados a Cada Arquivo

#### C1. Seleção
Se o mapa `selections` não é `nil`, cada arquivo deve ter seu caminho (`node.Path`) como chave com valor `true`. Arquivos sem entrada no mapa são pulados.

Se `selections == nil`, **todos os arquivos não-ignorados são considerados selecionados**.

#### C2. Limite de Quantidade de Arquivos
Se `fileCount >= config.MaxFiles`, o DFS é interrompido com erro.

#### C3. Limite de Tamanho do Arquivo
`shouldSkipFile` verifica:
- Se `node.IsDir` → skip (já que o DFS só chama `fn` para nós não-dir, esta é redundante mas segura).
- Se `node.Size > config.MaxFileSize` → skip (o arquivo é muito grande).

#### C4. Detecção de Arquivo Binário
Se `config.SkipBinary == true`:
1. `peekFileHeader(path)` abre o arquivo e lê os primeiros **1024 bytes**.
2. `isTextFile(string(header))` verifica:
   - Se content vazio → `true`.
   - Se os primeiros 1024 bytes (ou menos) contêm byte `0x00` → `false`.
   - Se não é UTF-8 válido → `false`.
   - Caso contrário → `true`.
3. Se binário → skip sem ler conteúdo completo.

#### C5. Leitura do Arquivo
`readFileContent(path)`:
1. `os.Open(path)`.
2. `io.ReadAll(file)` → `[]byte`.
3. `string(content)`.

#### C6. Limite de Tamanho Total
Se `totalSize + len(content) > config.MaxTotalSize`, erro é retornado imediatamente.

#### C7. Construção de FileContent
Para cada arquivo que passa nos filtros:
1. `RelPath` calculado via `filepath.Rel(root.Path, node.Path)`.
2. `Language` detectado via `detectLanguage(node.Name)`.
3. `FileContent` construído com `Path`, `RelPath`, `Language`, `Content`, `Size`.
4. Adicionado ao slice `files`.
5. `totalSize` e `fileCount` incrementados.

### Etapa D: Detecção de Linguagem
`detectLanguage(filename)` tenta duas abordagens:
1. **Por basename**: verifica o nome base do arquivo (sem extensão) contra uma lista fixa: `dockerfile`, `makefile`, `rakefile`, `gemfile`, `package.json`, `composer.json`, `cargo.toml`, `go.mod`, `requirements.txt`, etc.
2. **Por extensão**: verifica a extensão do arquivo contra um mapa de 50+ entradas. Fallback: `"text"`.

### Etapa E: Retorno
Se o DFS completa sem erro: `return files, nil`.
Se qualquer ponto de erro é atingido: `return nil, err`.

---

## Função `isTextFile` — Detalhes

| Condição | Resultado | Razão |
|----------|-----------|-------|
| `content` vazio (`len == 0`) | `true` | Arquivo vazio é considerado texto |
| Contém byte `0x00` | `false` | Byte nulo = binário |
| Não é UTF-8 válido | `false` | Não é texto legível |
| Nenhum byte nulo + UTF-8 válido | `true` | Texto legítimo |

**Tamanho de amostra:** máximo 1024 bytes. Se o conteúdo for maior, apenas os primeiros 1024 bytes são verificados.
