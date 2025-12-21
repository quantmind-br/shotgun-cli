# Sugestões Consolidadas - shotgun-cli

> Documento unificado a partir de `sugestoes1.md` e `sugestoes2.md`

---

## Resumo por Nível de Prioridade

| Nível | Quantidade | Descrição |
|-------|------------|-----------|
| **P0** | 5 | Alta prioridade - Alto impacto + Escopo contido |
| **P1** | 7 | Média prioridade - Multiplicadores de fluxo de trabalho |
| **P2** | 4 | Baixa prioridade - Apostas maiores / Segurança |
| **Ganhos Rápidos** | 8 | Implementação rápida - Agrupar com releases |
| **Total** | **24** | |

### Classificação Detalhada

#### P0 - Crítico (Fazer Primeiro)
| # | Funcionalidade | Complexidade | Justificativa |
|---|----------------|--------------|---------------|
| 1 | Flags CLI (--template, --task, --rules, --var) | S-M | Paridade modo sem interface/TUI - funcionalidade central incompleta |
| 2 | Ativar include-tree/include-summary | S | Chaves de config existem mas não funcionam - bug/dívida técnica |
| 3 | Configurações do scanner (workers, hidden, ignored) | M | Flexibilidade essencial para repositórios grandes |
| 4 | Progresso modo sem interface (humano + JSON) | M | UX crítica para automação/CI |
| 5 | Gemini status/doctor | S-M | Depuração de integração impossível atualmente |

#### P1 - Importante (Próxima Iteração)
| # | Funcionalidade | Complexidade | Justificativa |
|---|----------------|--------------|---------------|
| 6 | Divisão de diff por tokens | M | Alinhamento com orçamentos reais de LLM |
| 7 | Comando diff apply | M | Completa ciclo de fluxo de trabalho diff |
| 8 | Contexto a partir de diff | M | Fluxo de revisão de PR automatizado |
| 9 | Perfis salvos | M-L | Produtividade para uso repetido |
| 10 | TUI: variáveis customizadas | M | Templates avançados inutilizáveis no TUI |
| 11 | Profundidade máxima da árvore | S | Controle de tokens fácil |
| 12 | Template padrão configurável | S | Conveniência para usuários avançados |

#### P2 - Desejável (Lista de Pendências)
| # | Funcionalidade | Complexidade | Justificativa |
|---|----------------|--------------|---------------|
| 13 | Proteções para dados sensíveis | M-L | Segurança - importante mas escopo grande |
| 14 | Saída JSON + stdout | S-M | Pipelines de CI/CD |
| 15 | Detecção de linguagem aprimorada | M | Qualidade incremental |
| 16 | Temas TUI | M | Acessibilidade/preferência pessoal |

#### Ganhos Rápidos (Agrupar com Releases)
| # | Funcionalidade | Complexidade |
|---|----------------|--------------|
| GR1 | Status Gemini em config show | S |
| GR2 | Autocompletar template import/export | S |
| GR3 | scanner.max-memory | S |
| GR4 | Chaves gemini.* no autocompletar | S |
| GR5 | gemini.enabled consistente | S |
| GR6 | Template para mensagens de commit | S |
| GR7 | TUI: atalhos de teclado para Revisão | S |
| GR8 | TUI: filtro de templates | S |

---

## P0 - Alta Prioridade (Alto Impacto + Escopo Contido)

### 1. Paridade Modo Sem Interface com Assistente TUI (--template, --task, --rules, --var)

**Problema:** O modo sem interface define `TASK="Context generation"` fixo e deixa `RULES` vazio, sem flags para definir template/task/rules como o assistente coleta.

**Solução:** Adicionar flags ao `shotgun-cli context generate`:
- `--template <nome>` - usar sistema de templates existente
- `--task "<texto>"` - descrição da tarefa para o LLM
- `--rules "<texto>"` - regras/restrições
- `--var CHAVE=VALOR` (repetível) - popular variáveis de template customizadas

**Arquivos Impactados:**
- `cmd/context.go` (flags e `generateContextHeadless`)
- `internal/core/context/generator.go` (GenerateConfig)
- `internal/core/template/manager.go`

**Critério de Aceite:**
```bash
shotgun-cli context generate --template prompt_makePlan --task "Revisar código" --rules "Usar PT-BR" --var FOO=bar
```

**Complexidade:** S-M

---

### 2. Ativar Chaves de Configuração Pendentes (include-tree, include-summary)

**Problema:** As chaves de config `context.include-tree` e `context.include-summary` existem em autocompletar/config mas não têm efeito real - `GenerateConfig` não possui campos correspondentes.

**Solução:**
- Adicionar `IncludeTree bool` e `IncludeSummary bool` ao `internal/core/context.GenerateConfig`
- Passar valores através do gerador para omitir opcionalmente:
  - Bloco de árvore de diretórios
  - Seção de sumários por arquivo

**Arquivos Impactados:**
- `internal/core/context/generator.go`
- `cmd/context.go`
- `cmd/completion.go`

**Critério de Aceite:**
```bash
shotgun-cli config set context.include-tree false
# Resultado: prompts sem seção de árvore
```

**Complexidade:** S

---

### 3. Expor Configurações do Scanner (workers, hidden, ignored, gitignore)

**Problema:** Modo sem interface e assistente definem `IncludeHidden: false` e `Workers: 1` fixos. Assistente define `IncludeIgnored: true` para alternância, enquanto modo sem interface não expõe essa capacidade.

**Solução:** Adicionar chaves de config + flags:
- `scanner.workers` (int) - paralelismo
- `scanner.include-hidden` (bool) - incluir arquivos ocultos
- `scanner.include-ignored` (bool) - incluir arquivos ignorados
- `scanner.respect-gitignore` (bool) - já existe, garantir consistência
- `scanner.respect-shotgunignore` (bool) - **NOVO** - suporte de primeira classe ao `.shotgunignore`

**Arquivos Impactados:**
- `cmd/config.go`, `cmd/root.go` (padrões)
- `internal/core/scanner/filesystem.go`
- `internal/core/ignore/engine.go` (já tem `LoadShotgunignore`)

**Critério de Aceite:**
```bash
shotgun-cli context generate --workers 8 --include-hidden
# Varredura mais rápida e inclui arquivos ocultos
```

**Complexidade:** M

---

### 4. Saída de Progresso no Modo Sem Interface (humano + JSON)

**Problema:** TUI usa canais de progresso para varredura/geração; modo sem interface não fornece progresso apesar do suporte existente.

**Solução:** Adicionar `--progress=none|human|json` ao `context generate`:
- Scanner: usar `ScanWithProgress` para transmissão de `scanner.Progress`
- Gerador: usar `GenerateWithProgressEx` para transmissão de estágios

**Arquivos Impactados:**
- `cmd/context.go` (renderizador de progresso)
- `internal/core/scanner/filesystem.go`
- `internal/core/context/generator.go`

**Critério de Aceite:**
```bash
shotgun-cli context generate --progress human
# Mostra estágios de varredura e geração sem TUI
```

**Complexidade:** M

---

### 5. Comando Gemini Status/Doctor + Comportamento Consistente

**Problema:**
- Padrões definem `gemini.enabled=false`, mas modo sem interface pode enviar automaticamente se `gemini.auto-send=true` sem verificar `gemini.enabled`
- Autocompletar não sugere chaves `gemini.*`
- Auxiliar de status existe mas não está exposto

**Solução:**
- Adicionar `shotgun-cli gemini status` (e opcionalmente `doctor`)
- Mostrar em `config show`: seção "Status do Gemini" usando `internal/platform/gemini.GetStatus()`
- Aplicar regra: só permitir envio automático ou `--send-gemini` quando `gemini.enabled=true`
- Atualizar autocompletar do shell para incluir chaves `gemini.*`

**Arquivos Impactados:**
- `cmd/config.go` (novo subcomando ou seção em show)
- `cmd/context.go`, `cmd/send.go` (validação de habilitado)
- `cmd/completion.go` (chaves gemini.*)
- `internal/platform/gemini/config.go`

**Critério de Aceite:**
```bash
shotgun-cli gemini status
# Saída: geminiweb encontrado? configurado? caminho dos cookies?
```

**Complexidade:** S-M

---

## P1 - Média Prioridade (Multiplicadores de Fluxo de Trabalho)

### 6. Divisão de Diff por Token/Bytes (não apenas linhas)

**Problema:** Divisor de diff usa "linhas aproximadas por pedaço" que é um indicador fraco para orçamentos de LLM/contexto; o projeto já tem estimador de tokens.

**Solução:** Estender `diff split` com:
- `--max-bytes N` e/ou `--max-tokens N`
- Manter pontos de divisão seguros (limites de arquivo/hunks)
- Cabeçalhos incluem bytes + tokens estimados quando habilitados

**Arquivos Impactados:**
- `cmd/diff.go` (lógica de divisão)
- `internal/core/tokens/estimator.go`

**Critério de Aceite:**
```bash
shotgun-cli diff split --max-tokens 8000
# Pedaços próximos ao alvo respeitando limites de diff
```

**Complexidade:** M

---

### 7. Comando `diff apply`

**Problema:** Existe divisor de diff e gerador de contexto, mas `diff apply` completaria o ciclo de aplicar pedaços de volta ao repositório.

**Solução:** Implementar `shotgun-cli diff apply [caminho-do-pedaço] [caminho-do-repo]`:
- Wrapper do utilitário `patch` do SO ou implementação Go básica
- Validação de caminhos
- Relatório de erros detalhado

**Arquivos Impactados:**
- `cmd/diff.go` (novo subcomando)
- Potencial novo pacote utilitário para interação com SO

**Critério de Aceite:**
```bash
shotgun-cli diff apply pedaco_1.diff
# Aplica diff ao repositório atual ou falha graciosamente
```

**Questão:** Preferência por wrapper do `patch` do SO ou implementação Go pura multiplataforma?

**Complexidade:** M

---

### 8. Contexto a partir de Diff/Patch (Fluxo de Revisão de PR)

**Problema:** Não há ponte para "gerar contexto apenas para arquivos tocados por este diff".

**Solução:** Novo comando `shotgun-cli context from-diff --input alteracoes.diff [--include-related N]`:
- Analisar cabeçalhos de diff para caminhos de arquivos
- Construir mapa `selections` contendo apenas esses arquivos
- Executar `generator.Generate`

**Arquivos Impactados:**
- `cmd/context.go` ou novo `cmd/context_diff.go`
- `cmd/diff.go` (reutilizar auxiliares de análise)
- `internal/core/context/generator.go`

**Critério de Aceite:**
```bash
shotgun-cli context from-diff --input pr.diff
# Prompt inclui apenas arquivos tocados pelo diff
```

**Complexidade:** M

---

### 9. Perfis Salvos (Execuções Repetíveis)

**Problema:** Assistente captura estado rico (arquivos selecionados, template, tarefa, regras). Re-inserir isso toda vez custa tempo e introduz divergências.

**Solução:**
- `shotgun-cli profile save <nome>` e `shotgun-cli profile load <nome>`
- Ou `context generate --profile <nome>`
- Perfil armazena: padrões incluir/excluir, lista de seleções, nome do template, tarefa, regras, configurações de scanner/contexto

**Arquivos Impactados:**
- Novo `cmd/profile.go`
- `cmd/context.go` (flag --profile)
- `internal/ui/wizard.go` (opção de salvar)
- Sistema de config existente

**Critério de Aceite:**
```bash
# No assistente: salvar perfil
shotgun-cli context generate --profile minha-revisao
# Reproduz mesmo prompt e arquivos selecionados
```

**Complexidade:** M-L

---

### 10. TUI: Entrada de Variáveis Customizadas de Template

**Problema:** Assistente TUI só permite entrada para variáveis embutidas (`TASK`, `RULES`). Templates customizados frequentemente precisam de variáveis extras (ex: `{LANGUAGE}`, `{COMPONENT}`).

**Solução:** Introduzir novo passo no assistente *Passo 3b: Entrada de Variáveis Customizadas*:
- Após `StepTemplateSelection` e antes de `StepTaskInput`
- Apenas se template selecionado tem variáveis além de `TASK`/`RULES`/`FILE_STRUCTURE`

**Arquivos Impactados:**
- `internal/ui/wizard.go` (navegação)
- `internal/ui/screens/variable_input.go` (novo)
- `internal/core/template/renderer.go` (extração de variáveis)

**Critério de Aceite:**
- Ao selecionar template com variáveis não-padrão, aparece tela de entrada para coletá-las

**Complexidade:** M

---

### 11. Limite de Profundidade da Árvore (context.max-depth)

**Problema:** Para repositórios muito grandes, a árvore ASCII pode contribuir significativamente para uso de tokens e carga cognitiva.

**Solução:** Chave de config `context.max-depth` ou `context.tree-max-depth`:
- Controla profundidade do diretório de árvore renderizado
- Usa método existente `TreeRenderer.WithMaxDepth`

**Arquivos Impactados:**
- `cmd/config.go` (nova chave)
- `cmd/context.go` (ler config)
- `internal/core/context/generator.go`
- `internal/core/context/tree.go`

**Critério de Aceite:**
```bash
shotgun-cli config set context.max-depth 2
# FileStructure mostra apenas até profundidade 2
```

**Questão:** Padrão como valor positivo (ex: 5) ou ilimitado (-1 ou 0)?

**Complexidade:** S

---

### 12. Template Padrão Configurável

**Problema:** Funcionalidade central usa template padrão básico. Permitir configurar nome do template preferido melhoraria customização.

**Solução:** Chave de config `template.default-name`:
- Especifica nome do template (ex: `makePlan`, `analyzeBug`)
- Usado quando nenhum template explícito é especificado

**Arquivos Impactados:**
- `cmd/config.go`
- `cmd/root.go`
- `internal/core/context/generator.go`

**Critério de Aceite:**
```bash
shotgun-cli config set template.default-name makePlan
# Contexto gerado usa prompt_makePlan.md por padrão
```

**Complexidade:** S

---

## P2 - Prioridade Menor (Apostas Maiores / Segurança + Escalabilidade)

### 13. Proteções para Dados Sensíveis/Segredos

**Problema:** A ferramenta envia conteúdo completo de prompt para Gemini sem detecção de segredos.

**Solução:** Scanner de segredos com opção de desativar para suprimir ou bloquear credenciais óbvias:
- Novo módulo `internal/core/security/redact.go`
- Detectores baseados em regex + heurística de alta entropia
- Chaves de config: `security.redact-secrets` (padrão true), `security.fail-on-secrets` (padrão false)
- Relatório de supressões em arquivo auxiliar ou stderr

**Arquivos Impactados:**
- `internal/core/security/redact.go` (novo)
- `cmd/context.go`
- `cmd/send.go`
- `cmd/config.go`

**Critério de Aceite:**
```bash
# Repositório com padrão de chave de API detectado
shotgun-cli context generate
# Avisos emitidos e valores suprimidos por padrão
```

**Complexidade:** M-L

---

### 14. Modo de Saída Estruturada (JSON + stdout)

**Problema:** Modo sem interface sempre escreve arquivo; ContextData já tem tags JSON, sugerindo saída estruturada limpa.

**Solução:**
- `--output -` escreve prompt renderizado para stdout
- `--format json` emite objeto JSON contendo:
  - ContextData (tarefa/regras/data/árvore/arquivos)
  - Estatísticas (bytes, tokens estimados, contagem de arquivos)

**Arquivos Impactados:**
- `cmd/context.go`
- `internal/core/context/generator.go`

**Critério de Aceite:**
```bash
shotgun-cli context generate --output - | shotgun-cli send
# Pipeline funcional para CI
```

**Complexidade:** S-M

---

### 15. Detecção de Linguagem Aprimorada

**Problema:** Detecção de linguagem em `content.go` é baseada em mapa estático e switch pequeno. Frágil para extensões não-padrão.

**Solução:** Integrar biblioteca mais robusta ou expandir verificações heurísticas significativamente.

**Arquivos Impactados:**
- `internal/core/context/content.go`
- `go.mod` (possível nova dependência)

**Critério de Aceite:**
- Arquivos com extensões ambíguas (ex: `.h` para C/C++, `.log`) são corretamente marcados

**Complexidade:** M

---

### 16. TUI: Customização de Tema de Cores

**Problema:** Esquema de cores Nord está fixo em `theme.go`. Permitir seleção de temas alternativos melhoraria acessibilidade.

**Solução:** Chave de config `ui.theme` (ex: `nord`, `solarized`, `gruvbox`):
- Carregar cores dinamicamente em `theme.go`
- Definir structs de tema separados

**Arquivos Impactados:**
- `internal/ui/styles/theme.go`
- `cmd/config.go`
- Todo pacote `internal/ui`

**Questão:** Quais opções de tema padrão priorizar (Solarized, Gruvbox)?

**Complexidade:** M

---

## Ganhos Rápidos (Implementação Rápida)

### GR1. Exibir Status Gemini em `config show`
- Adicionar seção "Status do Gemini" usando `internal/platform/gemini.GetStatus()`
- **Arquivos:** `cmd/config.go`

### GR2. Autocompletar para Comandos `template`
- Implementar `ValidArgsFunction` para `template import` (completar arquivos `.md`)
- Completar nomes de templates existentes para `template export`
- **Arquivos:** `cmd/template.go`

### GR3. Implementar `scanner.max-memory`
- Campo `MaxMemory` existe em `ScanConfig` mas não é usado
- Adicionar config e verificação simples para abortar varredura se exceder limite
- **Arquivos:** `cmd/config.go`, `internal/core/scanner/filesystem.go`

### GR4. Adicionar Chaves `gemini.*` ao Autocompletar
- Atualizar listas `configKeyCompletion` e `boolValueCompletion`
- **Arquivos:** `cmd/completion.go`

### GR5. Garantir `gemini.enabled` Consistente
- Verificar em `context generate --send-gemini` e `context send`
- **Arquivos:** `cmd/context.go`, `cmd/send.go`

### GR6. Template para Mensagens de Commit
- Criar `prompt_makeCommitMessage.md`
- Usa `{TASK}` para instrução e conteúdo do diff
- **Arquivos:** `internal/assets/templates/`

### GR7. TUI: Atalhos de Teclado para Revisão
- `c` ou `Ctrl+C` para re-copiar para área de transferência
- `F9` para enviar para Gemini
- **Arquivos:** `internal/ui/wizard.go`, `internal/ui/screens/review.go`

### GR8. TUI: Filtro de Templates
- Entrada interativa para filtrar lista de templates por nome/descrição
- Adaptar lógica de `file_selection.go`
- **Arquivos:** `internal/ui/screens/template_selection.go`

---

## Questões em Aberto

1. **diff apply:** Preferência por wrapper do utilitário `patch` do SO ou implementação Go pura multiplataforma?

2. **Temas TUI:** Quais opções de tema padrão priorizar (Solarized, Gruvbox, outros)?

3. **context.max-depth:** Padrão como valor positivo (ex: 5) ou ilimitado (-1 ou 0)?

---

## Matriz de Priorização Completa

### Por Impacto × Complexidade

```
                    COMPLEXIDADE
                 S         M         L
            ┌─────────┬─────────┬─────────┐
       Alto │ #2      │ #1,#3   │ #13     │
            │         │ #4      │         │
    I       ├─────────┼─────────┼─────────┤
    M  Médio│ #11,#12 │ #6,#7,#8│ #9      │
    P       │ GR1-8   │ #10,#14 │         │
    A       ├─────────┼─────────┼─────────┤
    C  Baixo│         │ #15,#16 │         │
    T       │         │         │         │
    O       └─────────┴─────────┴─────────┘
```

### Ordem de Execução Recomendada

**Ciclo 1 (Ganhos Rápidos + P0 Simples)**
1. GR4 - Chaves gemini.* no autocompletar
2. GR5 - gemini.enabled consistente
3. #2 - Ativar include-tree/include-summary
4. GR1 - Status Gemini em config show
5. #5 - Gemini status/doctor

**Ciclo 2 (P0 Central)**
1. #1 - Flags CLI (--template, --task, --rules, --var)
2. #3 - Configurações do scanner
3. #4 - Progresso modo sem interface

**Ciclo 3 (P1 Fáceis + Ganhos Rápidos)**
1. #11 - Profundidade máxima da árvore
2. #12 - Template padrão configurável
3. GR6 - Template mensagens de commit
4. GR2 - Autocompletar template
5. GR3 - scanner.max-memory

**Ciclo 4 (P1 Fluxo de Trabalho)**
1. #6 - Divisão de diff por tokens
2. #7 - Comando diff apply
3. #8 - Contexto a partir de diff

**Ciclo 5 (P1 TUI + P2)**
1. GR7 - TUI atalhos de Revisão
2. GR8 - TUI filtro de templates
3. #10 - TUI variáveis customizadas
4. #14 - Saída JSON + stdout

**Lista de Pendências (P2 + Futuro)**
1. #9 - Perfis salvos
2. #13 - Proteções para segredos
3. #15 - Detecção de linguagem
4. #16 - Temas TUI

---

## Métricas de Esforço

| Complexidade | Estimativa | Quantidade |
|--------------|------------|------------|
| S (Pequena) | 1-2 dias | 12 |
| M (Média) | 3-5 dias | 10 |
| L (Grande) | 1-2 semanas | 2 |

**Esforço Total Estimado:** ~45-60 dias de desenvolvimento

---

*Documento consolidado em: 15/12/2025*
