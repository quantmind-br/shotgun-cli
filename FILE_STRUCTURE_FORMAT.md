# FILE_STRUCTURE Format

## Overview

O `{FILE_STRUCTURE}` agora gera um formato completo que inclui:
1. **Árvore ASCII** da estrutura de diretórios
2. **Blocos de conteúdo XML-like** com o conteúdo dos arquivos

## Formato de Saída

```
project_root/
|-- dir1/
|   |-- file1.txt
|   `-- file2.js
|-- dir2/
|   `-- subfile.md
`-- main.go

<file path="dir1/file1.txt">
conteúdo de file1.txt...
</file>
<file path="dir1/file2.js">
conteúdo de file2.js...
</file>
<file path="dir2/subfile.md">
conteúdo de subfile.md...
</file>
<file path="main.go">
conteúdo de main.go...
</file>
```

## Características

### Árvore ASCII
- Usa prefixos `├──` e `└──` para conectores
- Diretórios são listados antes de arquivos
- Identação com `│   ` para hierarquia
- Diretórios terminam com `/`
- Arquivos mostram tamanho `[XXB]`

### Blocos de Conteúdo
- Formato XML-like: `<file path="caminho/relativo">`
- Path relativo ao root do projeto
- Conteúdo literal do arquivo
- Tag de fechamento: `</file>`
- Um bloco por arquivo selecionado

## Implementação

### Arquivos Modificados

1. **internal/core/context/content.go**
   - Adicionada função `renderFileContentBlocks()` que gera blocos XML-like

2. **internal/core/context/generator.go**
   - Adicionada função `buildCompleteFileStructure()` que combina árvore + blocos
   - Modificado `GenerateWithProgressEx()` para usar estrutura completa

3. **internal/core/context/filestructure_test.go** (novo)
   - Testes unitários validando formato completo
   - Testes de renderização de blocos XML-like

### Uso em Templates

Templates externos (como `prompt_makePlan.md`) que usam `{FILE_STRUCTURE}` receberão:
- Árvore completa de arquivos
- Conteúdo de todos os arquivos selecionados em blocos XML-like

Exemplo de template:
```markdown
## Estrutura do Projeto

{FILE_STRUCTURE}
```

Resultado:
```
## Estrutura do Projeto

└── myproject/
    ├── src/
    │   └── main.go [123B]
    └── README.md [45B]

<file path="src/main.go">
package main

func main() {
    println("Hello")
}
</file>
<file path="README.md">
# My Project
</file>
```

## Compatibilidade

### Template Padrão
O template padrão interno continua funcionando como antes:
- Usa `{{.FileStructure}}` para árvore
- Usa `{{range .Files}}` para conteúdo separado em blocos markdown

### Templates Customizados
Templates customizados que usam `{FILE_STRUCTURE}` agora recebem:
- Estrutura completa (árvore + conteúdo)
- Formato otimizado para LLMs

## Exemplo Completo

### Input (Projeto)
```
myproject/
├── src/
│   └── main.go
└── README.md
```

### Output (FILE_STRUCTURE)
```
└── myproject/
    ├── src/
    │   └── main.go [77B]
    └── README.md [40B]

<file path="src/main.go">
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
</file>
<file path="README.md">
# Demo Project

This is a test project.
</file>
```

## Testes

Execute os testes para validar:

```bash
# Testes de FILE_STRUCTURE
go test -v ./internal/core/context -run TestFileStructure

# Todos os testes
make test
```

## Notas de Implementação

- **Performance**: Blocos de conteúdo são gerados apenas para arquivos selecionados
- **Encoding**: Conteúdo preserva encoding original (UTF-8)
- **Newlines**: Garante newline antes de `</file>` se conteúdo não terminar com newline
- **Paths**: Usa paths relativos ao root do scan
- **Ordem**: Árvore vem primeiro, depois blocos de conteúdo
