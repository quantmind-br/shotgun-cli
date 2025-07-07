# **PRD: shotgun-cli**

## 1. Visão Geral e Objetivos

O `shotgun-cli` é uma ferramenta de linha de comando (CLI) desenvolvida em Go que visa otimizar a interação de desenvolvedores com modelos de linguagem de IA (LLMs). A aplicação permite a criação rápida e padronizada de prompts complexos diretamente do terminal, eliminando a necessidade de formatação manual repetitiva e o uso de APIs.

### Objetivos Principais:
*   **Agilizar a Geração de Prompts:** Reduzir o tempo gasto por desenvolvedores na criação de prompts para tarefas comuns como arquitetura de software, desenvolvimento de código e depuração.
*   **Padronizar a Qualidade:** Garantir que os prompts enviados às IAs sejam consistentes, completos e bem-estruturados, aumentando a qualidade e a precisão das respostas.
*   **Operação Offline e Segura:** Funcionar inteiramente no ambiente local do usuário, lendo os arquivos do projeto sem a necessidade de enviar código para serviços de terceiros ou gerenciar chaves de API.
*   **Facilidade de Uso:** Oferecer uma experiência de usuário fluida e intuitiva no terminal através de uma interface baseada em texto (TUI).

## 2. Personas de Usuário

### Persona Primária: O Desenvolvedor de Software
*   **Descrição:** Um desenvolvedor (backend, frontend, full-stack) que utiliza IAs generativas (como ChatGPT, Claude, etc.) em seu fluxo de trabalho diário para obter ajuda com código, analisar bugs ou projetar novas funcionalidades.
*   **Necessidades:** Precisa de uma forma rápida de fornecer o contexto completo do projeto (estrutura de arquivos, regras, tarefa específica) para a IA.
*   **Frustrações:** Perde tempo formatando prompts repetidamente, esquecendo de incluir arquivos ou regras importantes, o que resulta em respostas genéricas ou incorretas da IA.

## 3. Requisitos Funcionais (RF)

### RF1: Invocação da Aplicação
O usuário poderá iniciar a aplicação digitando `shotgun-cli` em qualquer diretório no seu terminal.

### RF2: Seleção do Tipo de Tarefa (Revisado)
*   Ao iniciar, a aplicação exibirá uma lista de seleção para o tipo de tarefa.
*   As opções serão: `architect`, `dev`, `find bug`, e `docs-sync`.
*   A seleção de uma tarefa determinará qual template de prompt será usado:
    *   **architect**: Usará o template `prompt_makePlan.md`.
    *   **dev**: Usará o template `prompt_makeDiffGitFormat.md`.
    *   **find bug**: Usará o template `prompt_analyzeBug.md`.
    *   **docs-sync**: Usará o template `prompt_projectManager.md`.

### RF3: Entrada da Tarefa Principal (`{TASK}`)
*   Após selecionar o tipo, um campo de texto multiline será aberto para o usuário descrever a tarefa.
*   As linhas inseridas deverão ser numeradas automaticamente.
*   O usuário poderá finalizar a entrada com um atalho de teclado (ex: `alt+d`).

### RF4: Entrada das Regras do Projeto (`{RULES}`)
*   A aplicação perguntará ao usuário se ele deseja adicionar regras específicas para o projeto.
*   Se a resposta for afirmativa, um campo de texto multiline (similar ao RF3) será exibido para a inserção das regras.
*   Se a resposta for negativa, a seção `{RULES}` no prompt final será preenchida com o texto: `1. There's no user rules.`

### RF5: Seleção de Arquivos para o Contexto (`{FILE_STRUCTURE}`)
*   Será exibida uma interface de árvore de arquivos (file tree) do diretório atual.
*   **RF5.1: Exclusão Padrão:** Arquivos e pastas listados no `.gitignore` do projeto virão pré-desselecionados.
*   **RF5.2: Exclusão Configurável:** A aplicação suportará um arquivo de configuração global (ex: `~/.shotgun_cli/config.json`) onde o usuário pode definir padrões de exclusão adicionais (ex: `node_modules/`, `*.log`). Um comando `shotgun-cli --config` permitirá gerenciar essa configuração.
*   **RF5.3: Interação:** O usuário poderá navegar na árvore e usar uma tecla (ex: `espaço`) para selecionar/desselecionar arquivos ou pastas. Desselecionar uma pasta automaticamente desselecionará todos os seus conteúdos.

### RF6: Geração do Prompt Final
*   Após a seleção dos arquivos, a aplicação irá:
    1.  Ler o conteúdo do template de prompt correspondente (RF2).
    2.  Substituir o placeholder `{TASK}` com o conteúdo do RF3.
    3.  Substituir o placeholder `{RULES}` com o conteúdo do RF4.
    4.  Gerar a string do `{FILE_STRUCTURE}` contendo a árvore de arquivos selecionados seguida pelo conteúdo de cada um desses arquivos.
    5.  Salvar o resultado em um novo arquivo no diretório atual (ex: `shotgun_prompt_20250707_121530.md`).

## 4. Requisitos Não-Funcionais (RNF)

*   **RNF1: Usabilidade:** A interface no terminal (TUI) deve ser intuitiva, com instruções claras em cada etapa.
*   **RNF2: Desempenho:** A aplicação deve ser rápida para iniciar e para processar os arquivos, mesmo em projetos de médio porte.
*   **RNF3: Compatibilidade:** O binário final deve ser compatível com os principais sistemas operacionais (Windows, macOS, Linux).
*   **RNF4: Instalação:** A instalação deve ser simplificada através do `npm`.

## 5. Plano Técnico de Alto Nível

### Fluxo da Aplicação (Revisado)
```mermaid
flowchart TD
    A[Start: usuário executa `shotgun-cli`] --> B{Seleção de Tarefa <br> (architect, dev, find bug, docs-sync)};
    B --> C[Entrada da Tarefa <br> (campo multiline)];
    C --> D{Adicionar Regras?};
    D -- Sim --> E[Entrada das Regras <br> (campo multiline)];
    D -- Não --> F[Usa regra padrão];
    E --> G[Seleção de Arquivos <br> (File Tree com exclusões)];
    F --> G;
    G --> H[Processamento: <br> 1. Ler template <br> 2. Ler arquivos selecionados <br> 3. Montar prompt final];
    H --> I[End: Salva `prompt_final.md` no disco];
```

### Tecnologias e Bibliotecas
*   **Linguagem:** **Go**, pela sua performance e capacidade de gerar binários compilados.
*   **TUI Framework:** **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**, juntamente com seu ecossistema (`lipgloss` para estilo, `bubbles` para componentes). É uma escolha robusta e popular para TUIs em Go.
*   **Parsing de .gitignore:** Uma biblioteca como **[go-gitignore](https://github.com/sabhiram/go-gitignore)** para processar as regras de exclusão.
*   **Geração da Estrutura de Arquivos:** Podemos investigar as bibliotecas sugeridas (`gitingest`, `files2prompt`) ou implementar uma função nativa em Go para percorrer o diretório e ler os arquivos selecionados.

### Desafio: Distribuição via NPM
Distribuir um binário Go via `npm` é um desafio interessante. A abordagem padrão é usar o `npm` como um "gerenciador de versão" e "instalador" para o binário.

1.  **Compilação Cruzada:** O projeto Go será configurado para gerar binários para diferentes arquiteturas (windows-amd64, linux-amd64, darwin-arm64, etc.).
2.  **GitHub Releases:** A cada nova versão, os binários compilados serão publicados como assets em uma Release no GitHub.
3.  **Pacote NPM:** O pacote `npm` em si não conterá o binário. Em vez disso, ele terá um script de `postinstall`.
4.  **Script `postinstall`:** Este script detecta o SO e a arquitetura do usuário (`process.platform`, `process.arch`) e baixa o binário correspondente da Release do GitHub, colocando-o em um local acessível (`node_modules/.bin/shotgun-cli`).

Exemplo de `package.json`:
```json
{
  "name": "shotgun-cli",
  "version": "0.1.0",
  "bin": {
    "shotgun-cli": "./bin/shotgun-cli"
  },
  "scripts": {
    "postinstall": "node install.js"
  }
}
```
O arquivo `install.js` conteria a lógica para baixar o binário correto.

## 6. Escopo

### Dentro do Escopo:
*   Todas as funcionalidades descritas nos Requisitos Funcionais.
*   CLI com TUI baseada em texto.
*   Configuração de exclusão de arquivos via `.gitignore` e um arquivo de configuração global.

### Fora do Escopo (para a v1.0):
*   Uma interface gráfica (GUI).
*   Integração direta com APIs de LLMs.
*   Sincronização de configurações na nuvem.
*   Criação ou edição de templates de prompt via CLI.