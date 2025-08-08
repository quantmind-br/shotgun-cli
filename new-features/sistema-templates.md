# FEATURE: 

Implemente a funcionalidade de "Sistema de Templates Extensível" na ferramenta `shotgun-cli`, escrita em Go.

---

### 1. PRIMARY GOAL

Seu objetivo principal é refatorar o sistema de templates do `shotgun-cli`, transformando-o de um sistema estático e pré-definido para um sistema dinâmico e extensível, permitindo que os usuários adicionem seus próprios templates de prompt.

### 2. CONTEXTO DO PROJETO

O `shotgun-cli` é uma ferramenta de TUI (Terminal User Interface) construída com Go e a biblioteca BubbleTea. Sua função é ajudar desenvolvedores a gerar prompts para LLMs a partir do contexto de suas bases de código.

Atualmente, a ferramenta possui um sistema de templates estático, definido em `internal/core/template.go` e `internal/core/template_simple.go`. A lista de templates disponíveis (`AvailableTemplates`) é fixa no código-fonte, limitando os usuários aos quatro templates padrão. A sua tarefa é remover essa limitação.

### 3. DESCRIÇÃO COMPLETA DA FUNCIONALIDADE

A funcionalidade "Sistema de Templates Extensível" permitirá que o `shotgun-cli` descubra e carregue dinamicamente templates de prompt (`.md`) de um diretório de configuração do usuário. Isso dará aos usuários o poder de criar, personalizar e compartilhar seus próprios templates, adaptando a ferramenta aos seus fluxos de trabalho específicos.

**Benefícios a serem alcançados:**
*   **Personalização:** Usuários poderão criar templates para qualquer linguagem, framework ou necessidade.
*   **Padronização de Equipes:** Equipes poderão compartilhar um conjunto de templates para garantir consistência.
*   **Ecossistema e Comunidade:** A comunidade poderá criar e compartilhar seus próprios templates.
*   **Flexibilidade:** A ferramenta se tornará à prova de futuro, adaptável a novas técnicas de prompting sem precisar de atualizações no código-fonte.

### 4. FLUXO DE FUNCIONAMENTO DETALHADO

1.  **Localização dos Templates:** A ferramenta deverá procurar por templates personalizados em um diretório específico, seguindo o padrão XDG. O caminho será: `[xdg.ConfigHome]/shotgun-cli/templates/`. Se este diretório não existir na inicialização, a aplicação deverá criá-lo.

2.  **Lógica de Carregamento:** Na inicialização, o `SimpleTemplateProcessor` (em `internal/core/template_simple.go`) deverá:
    a. Primeiro, carregar os templates padrão internos como base.
    b. Em seguida, escanear o diretório de templates do usuário (`.../shotgun-cli/templates/`).
    c. Para cada arquivo com a extensão `.md` encontrado, a aplicação tentará carregá-lo como um template personalizado.

3.  **Formato do Template Personalizado:** Cada arquivo `.md` no diretório de templates do usuário deve conter um bloco de metadados **YAML Frontmatter** no início do arquivo, seguido pelo conteúdo do prompt.
    *   O YAML Frontmatter é delimitado por `---`.
    *   Campos obrigatórios no YAML: `key` (string, único), `name` (string, para exibição), `description` (string, para exibição).

4.  **Integração na UI:** A lista de templates na tela de seleção (`TemplateSelection`) deverá ser atualizada para exibir uma lista combinada, agrupando os templates padrão e os personalizados carregados dinamicamente.

### 5. REQUISITOS DE IMPLEMENTAÇÃO

**Modificações no Código Core (`internal/core/`):**

1.  **Dependência:** Adicione uma dependência para parsing de YAML. A biblioteca `gopkg.in/yaml.v3` é a escolha recomendada. Execute `go get gopkg.in/yaml.v3` e `go mod tidy`.

2.  **Modifique `internal/core/template_simple.go`:**
    *   Altere a função `LoadTemplatesFromDirectory` para `LoadTemplates`. Esta função deverá orquestrar o carregamento de templates padrão e personalizados.
    *   Crie uma nova função, `loadCustomTemplates(templatesDir string)`, que será responsável por escanear o diretório do usuário, ler os arquivos `.md`, e parsear o YAML Frontmatter.
    *   A struct `TemplateInfo` (em `internal/core/template.go`) já é adequada para armazenar os metadados. Utilize-a.
    *   A lógica de carregamento deve ser resiliente: se um arquivo `.md` estiver mal formatado ou não contiver o YAML Frontmatter necessário, ele deve ser ignorado e um aviso pode ser logado no console (sem quebrar a aplicação).

3.  **Modifique `internal/core/types.go`:**
    *   Garanta que a struct `TemplateInfo` contenha os campos `Key`, `Name`, e `Description` para ser usada tanto pelos templates padrão quanto pelos personalizados.

**Modificações na UI (`internal/ui/`):**

1.  **Modifique `internal/ui/app.go` (ou onde a inicialização ocorre):**
    *   Na inicialização do `Model`, chame a nova lógica de carregamento de templates para popular a lista de templates disponíveis.
    *   Certifique-se de que o diretório de templates do usuário seja criado se não existir.

2.  **Modifique `internal/ui/views.go`:**
    *   A função `renderTemplateSelection` deve ser atualizada para renderizar a lista dinâmica de templates.
    *   A lista deve ser apresentada de forma clara, preferencialmente com subtítulos para "[Templates Padrão]" e "[Templates Personalizados]".

### 6. EXEMPLO DE TEMPLATE PERSONALIZADO

O agente deve ser capaz de parsear um arquivo como `~/.config/shotgun-cli/templates/prompt_jest_tests.md` com o seguinte conteúdo:

```markdown
---
key: "jest-generator"
name: "Gerador de Testes (Jest)"
description: "Gera testes unitários em Jest para o código selecionado."
---
## OBJETIVO:
Você é um Engenheiro de Testes de Software especialista em Jest. Sua missão é analisar o código em `{FILE_STRUCTURE}` e gerar testes unitários robustos.

## TAREFA DO USUÁRIO:
{TASK}

## REGRAS ADICIONAIS:
{RULES}

## CÓDIGO-FONTE PARA TESTAR:
{FILE_STRUCTURE}
```

### 7. DESAFIOS E CONSIDERAÇÕES (COMO LIDAR COM ELES)

*   **Conflito de Chaves (`key`):** Se um template personalizado tiver a mesma `key` de um template padrão (ex: "dev"), o template padrão deve ter prioridade. O template personalizado deve ser ignorado e um aviso deve ser logado.
*   **Validação de Templates:** A lógica de carregamento deve validar a presença dos campos `key`, `name`, e `description` no YAML. Se algum estiver faltando, o template é inválido e deve ser ignorado.
*   **Interface do Usuário:** A UI deve lidar graciosamente com um número variável de templates, incluindo o caso de não haver nenhum template personalizado. A navegação por números (1-4) pode ser mantida para os templates padrão, mas a navegação por setas (cima/baixo) deve percorrer toda a lista.

### 8. CRITÉRIOS DE ACEITAÇÃO (DEFINITION OF DONE)

1.  [ ] A aplicação inicia sem erros, mesmo que o diretório de templates personalizados não exista (ele deve ser criado).
2.  [ ] Se arquivos `.md` com o formato YAML correto forem colocados em `~/.config/shotgun-cli/templates/`, eles aparecem na lista de seleção de templates na UI.
3.  [ ] Se um arquivo `.md` no diretório personalizado estiver mal formatado ou sem os metadados necessários, a aplicação não quebra e simplesmente ignora o arquivo.
4.  [ ] A UI distingue claramente entre templates padrão e personalizados.
5.  [ ] A seleção de um template personalizado funciona corretamente, e o conteúdo do prompt é usado na etapa de geração.
6.  [ ] A dependência `gopkg.in/yaml.v3` foi corretamente adicionada ao `go.mod`.

### 9. ESTRUTURA DE ARQUIVOS RELEVANTE (PARA SUA ORIENTAÇÃO)

```
shotgun-cli/
├── go.mod
├── main.go
└── internal/
    ├── core/
    │   ├── template.go         # Contém a struct TemplateInfo e a lista estática
    │   ├── template_simple.go  # Onde a lógica de carregamento e geração reside
    │   └── types.go            # Definições de tipos centrais
    └── ui/
        ├── app.go              # Modelo principal da UI
        └── views.go            # Lógica de renderização das diferentes telas
```

---

**AÇÃO:** Inicie a implementação. Modifique os arquivos necessários, adicione a nova dependência e garanta que todos os critérios de aceitação sejam cumpridos.
