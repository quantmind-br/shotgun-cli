# Custom Template Example

<!--
  INSTRUÇÕES PARA CRIAR SEU TEMPLATE:

  1. A primeira linha com # ou o primeiro comentário HTML define a descrição
  2. Use {VARIAVEL} para placeholders (apenas MAIÚSCULAS, números e _)
  3. Salve em ~/.config/shotgun-cli/templates/ ou configure custom path
  4. Nome do arquivo (sem extensão) será o nome do template

  VARIÁVEIS DISPONÍVEIS:
  - {TASK}            : Tarefa específica fornecida pelo usuário
  - {RULES}           : Regras e restrições do usuário
  - {FILE_STRUCTURE}  : Estrutura de arquivos e conteúdo do projeto
  - {CURRENT_DATE}    : Data atual (gerada automaticamente)

  Você pode criar suas próprias variáveis customizadas também!
-->

## Context

You are an AI assistant helping with software development tasks.

**Current Date:** {CURRENT_DATE}

---

## Task Description

{TASK}

---

## Project Constraints & Rules

{RULES}

---

## Project Structure

{FILE_STRUCTURE}

---

## Instructions

Please analyze the provided information and:

1. Understand the task requirements
2. Review the project structure
3. Consider the specified rules and constraints
4. Provide a detailed solution

## Output Format

[Customize this section based on your needs - markdown, code, analysis, etc.]
