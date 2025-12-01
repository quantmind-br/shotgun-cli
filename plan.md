╭────────────────────────────────────────────────────────────────────────────╮
│                                                                            │
│    Refactoring/Design Plan: Otimização e Enxugamento do Código do          │
│   shotgun-cli                                                              │
│                                                                            │
│   ## 1. Executive Summary & Goals                                          │
│                                                                            │
│   Este plano tem como objetivo otimizar o código do  shotgun-cli           │
│   focando em eficiência, remoção de código morto/legado e                  │
│   simplificação da arquitetura para garantir um desempenho superior e      │
│   uma base de código mais enxuta e fácil de manter.                        │
│                                                                            │
│   * Goal 1: Redução do  Dead Code  e  Legacy Code : Eliminar código        │
│   não utilizado ou de compatibilidade que não contribui para a             │
│   funcionalidade principal da versão atual.                                │
│   * Goal 2: Refatoração para Eficiência: Melhorar a performance            │
│   focando na redução de alocações, simplificação de lógica e               │
│   otimização de caminhos críticos (e.g.,  scanner ,  context               │
│   generation).                                                             │
│   * Goal 3: Simplificação da Arquitetura: Consolidar a lógica              │
│   duplicada e remover abstrações desnecessárias para uma aplicação         │
│   mais "enxuta" (lean).                                                    │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 2. Current Situation Analysis                                         │
│                                                                            │
│   O  shotgun-cli  apresenta uma estrutura Go bem definida, separada em     │
│   cmd/  (comandos CLI),  internal/core/  (lógica de negócios) e            │
│   internal/platform/  (integrações externas/OS). A análise revela          │
│   oportunidades claras para enxugamento, especialmente nas áreas de        │
│   clipboard e validação de configuração ( cmd/config.go ,                  │
│   cmd/context.go ).                                                        │
│                                                                            │
│   * Pontos de Dor e Oportunidades de Otimização:                           │
│     * Lógica de  Clipboard  Duplicada/Complexa (                           │
│     internal/platform/clipboard/ ): Há várias implementações de            │
│     ClipboardManager  (Linux, Darwin, Windows, WSL), um  Manager  de       │
│     alto nível, e funções de cópia no pacote de nível superior (           │
│     clipboard.Copy ). Além disso,  cmd/context.go  contém uma função       │
│     de cópia não modular ( clipboard.Copy(content) ), e há uma lógica      │
│     de  CopyLarge  ( internal/platform/clipboard/manager.go ) que          │
│     parece subutilizada ou redundante se a cópia for padronizada. O        │
│     código de  clipboard  parece excessivamente complexo.                  │
│     * Validação de Configuração Redundante e Desatualizada (               │
│     cmd/config.go ,  cmd/context.go ): As funções de validação (           │
│     validateMaxFiles ,  validateSizeFormat , etc.) estão em                │
│     cmd/config.go . A lógica de  parseSize  está duplicada em              │
│     cmd/context.go  e  cmd/config.go . O código de  parseSize  em          │
│     cmd/context.go  também trata de conversão de tamanhos que seria        │
│     melhor movida para um pacote de utilitários central.                   │
│     *  Dead Code  / Código Legado no  Context Generator : O pacote         │
│     internal/core/context/generator.go  possui um campo  MaxSize  na       │
│     sua  GenerateConfig  marcado como  Deprecated  que é mantido           │
│     apenas por compatibilidade (função  validateConfig ).                  │
│     * Estrutura de  DiffSplit  Simples (cmd/diff.go): A lógica de          │
│     diff split  é auto-suficiente, mas poderia ser movida para             │
│     internal/core/  se houver expectativa de reutilização da lógica de     │
│     análise de diff. Por enquanto, a complexidade está bem encapsulada.    │
│     * Abstração Desnecessária de  Scanner : O pacote                       │
│     internal/core/scanner/filesystem.go  ainda mantém código para          │
│     PathMatcher  legado/não utilizado (LoadGitIgnoreMatcher,               │
│     LoadGitIgnore, LoadGitIgnoreAdapter) que foi substituído pelo          │
│     ignore.IgnoreEngine .                                                  │
│                                                                            │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 3. Proposed Solution / Refactoring Strategy                           │
│                                                                            │
│   O foco principal é isolar a lógica de utilidade, remover código          │
│   redundante e simplificar a arquitetura de integração de plataforma (     │
│   clipboard ).                                                             │
│                                                                            │
│   ### 3.1. High-Level Design / Architectural Overview                      │
│                                                                            │
│   * Princípio: Consolidação de utilitários e remoção de abstrações         │
│   legadas.                                                                 │
│   * Ação: Mover as funções de utilidade (e.g.,  parseSize ,                │
│   formatBytes ) para um novo pacote  internal/utils/conversion .           │
│   Padronizar a camada de  clipboard  para usar a biblioteca externa        │
│   atotto/clipboard  diretamente onde possível, ou simplificar              │
│   drasticamente a lógica de  Manager  e  CopyLarge .                       │
│                                                                            │
│   ### 3.2. Key Components / Modules                                        │
│                                                                            │
│    Module              | Change              | Rationale                   │
│   ---------------------+---------------------+----------------------       │
│    NEW:                | Criar um pacote     | Centraliza a lógica,        │
│     internal/utils/con | para funções de     | remove duplicação e         │
│    version             | utilidade de        | torna o código mais         │
│                        | tamanho             | testável.                   │
│                        | ( parseSize ,       |                             │
│     cmd/config.go      | Atualizar           | Elimina a duplicação        │
│                        |  validateMaxFiles   | de lógica de parsing        │
│                        | e                   | de tamanho.                 │
│                        |  validateSizeFormat |                             │
│                        |   para usar         |                             │
│                        |  internal/utils/con |                             │
│     cmd/context.go     | Remover a           | Elimina a                   │
│                        | duplicação da       | duplicação.                 │
│                        | lógica  parseSize   |                             │
│                        | e  formatBytes      |                             │
│                        | (incluindo          |                             │
│                        |  parseConfigSize ). |                             │
│     internal/platform/ | Simplificar/remover | Reduz a complexidade        │
│    clipboard/          |  clipboard.Manager  | da integração, pois         │
│                        | e a lógica de       | o TUI não utiliza o         │
│                        | detecção de         |  Manager  de alto           │
│                        | ferramenta se não   | nível (usa                  │
│                        | for usada pela CLI  |  clipboard.Copy  do         │
│                        | principal. Remover  | pacote raiz, que por        │
│                        |  CopyLarge  se não  | sua vez usa o               │
│                        | for                 |  Manager ).                 │
│                        | usada/necessária.   |                             │
│                        | Remover arquivos de |                             │
│                        | implementação       |                             │
│                        | específicos de OS   |                             │
│                        | ( darwin.go ,       |                             │
│                        |  linux.go ,         |                             │
│                        |  windows.go ,       |                             │
│                        |  wsl.go ) e usar a  |                             │
│                        | biblioteca de       |                             │
│                        | terceiros ou o      |                             │
│                        |  Manager            |                             │
│                        | simplificado.       |                             │
│     internal/core/scan | Remover código      | Eliminar  Dead Code         │
│    ner/filesystem.go   | legado relacionado  | e simplificar o             │
│                        | a  PathMatcher ,    | componente central          │
│                        |  gitIgnoreAdapter , | de scanning.                │
│                        |  LoadGitIgnoreMatch |                             │
│                        | er , e              |                             │
│                        |  LoadGitIgnore  que |                             │
│                        | não é mais          |                             │
│                        | utilizado devido ao |                             │
│                        | uso do              |                             │
│                        |  ignore.IgnoreEngin |                             │
│                        | e .                 |                             │
│     internal/core/cont | Remover o campo     | Eliminar código             │
│    ext/generator.go    |  MaxSize  e a       | legado.                     │
│                        | lógica de           |                             │
│                        | compatibilidade em  |                             │
│                        |  validateConfig .   |                             │
│                                                                            │
│   ### 3.3. Detailed Action Plan / Phases                                   │
│                                                                            │
│   #### Phase 1: Consolidação e Remoção de Código Legado (High Impact,      │
│   High Priority)                                                           │
│                                                                            │
│    Task               | Rationale/Goal      | | Deliverable/Criter…        │
│   --------------------+---------------------+-+---------------------       │
│    1.1: Criar         | Centralizar         | | Novo pacote com            │
│     internal/utils/co |  parseSize ,        | | funções unit-              │
│    nversion           |  formatBytes .      | | tested.                    │
│    1.2: Eliminar      | Remover  parseSize  | |  cmd/context.go  e         │
│     parseSize         | e  parseConfigSize  | |  cmd/config.go             │
│    Duplicado          | de  cmd/context.go  | | usam apenas                │
│                       | e  cmd/config.go .  | |  internal/utils/con        │
│                       | Atualizar chamadas. | | version .                  │
│    1.3: Eliminar      | Remover             | | Removidas ~50              │
│     PathMatcher       |  gitIgnoreAdapter , | | linhas de código           │
│    Legado             |  LoadGitIgnoreMatch | | legado.                    │
│                       | er ,                | |                            │
│                       |  LoadGitIgnore  de  | |                            │
│                       |  internal/core/scan | |                            │
│                       | ner/filesystem.go . | |                            │
│    1.4: Refatorar     | Substituir o uso de | |  content.go  e             │
│     internal/core/con |  parseSize / format | |  content_test.go           │
│    text/content.go    | Bytes  por chamadas | | atualizados.               │
│                       | a                   | |                            │
│                       |  internal/utils/con | |                            │
│                       | version .           | |                            │
│    1.5: Remover       | Remover o campo     | |  context.GenerateCo        │
│     MaxSize           |  MaxSize  de        | | nfig  simplificado.        │
│    Deprecado          |  context.GenerateCo | |                            │
│                       | nfig  e a lógica de | |                            │
│                       | fallback em         | |                            │
│                       |  validateConfig .   | |                            │
│                                                                            │
│   #### Phase 2: Simplificação e Enxugamento do Clipboard (High Impact,     │
│   Medium Priority)                                                         │
│                                                                            │
│    Task               | Rationale/Goal      | | Deliverable/Criter…        │
│   --------------------+---------------------+-+---------------------       │
│    2.1: Simplificar   | Se o uso da         | | Arquivos                   │
│     clipboard.Manager | biblioteca          | |  internal/platform/        │
│                       |  atotto/clipboard   | | clipboard/*.go             │
│                       | for suficiente para | | reduzidos.                 │
│                       | a função            | |                            │
│                       |  Copy(content) ,    | |                            │
│                       | simplificar o       | |                            │
│                       |  Manager  para      | |                            │
│                       | apenas ser uma      | |                            │
│                       | fachada ou removê-  | |                            │
│                       | lo se possível.     | |                            │
│    2.2: Remover       | Excluir arquivos    | | Arquivos                   │
│    Implementações     | como  darwin.go ,   | | específicos de OS          │
│    Específicas de OS  |  linux.go ,         | | deletados.                 │
│                       |  windows.go ,       | |                            │
│                       |  wsl.go  (e seus    | |                            │
│                       | testes) se a        | |                            │
│                       | biblioteca externa  | |                            │
│                       | ou uma              | |                            │
│                       | implementação       | |                            │
│                       | simplificada for    | |                            │
│                       | adotada.            | |                            │
│    2.3: Remover       | Remover funções     | | Código                     │
│    Lógica de          |  CopyLarge  e       | | desnecessário              │
│     CopyLarge         |  copyViaTempFile    | | removido.                  │
│                       | de                  | |                            │
│                       |  internal/platform/ | |                            │
│                       | clipboard/manager.g | |                            │
│                       | o  se a             | |                            │
│                       | funcionalidade for  | |                            │
│                       | considerada         | |                            │
│    2.4: Atualizar     | Garantir que        | | Funcionalidade de          │
│    chamadas CLI       |  cmd/context.go  e  | | cópia                      │
│                       | outras chamadas ao  | | inalterada/melhorad        │
│                       |  clipboard          | | a na CLI.                  │
│                       | utilizem a          | |                            │
│                       | interface de cópia  | |                            │
│                       | simplificada/padron | |                            │
│                       | izada.              | |                            │
│                                                                            │
│   #### Phase 3: Finalização e Testes (Medium Priority)                     │
│                                                                            │
│    Task               | Rationale/Goal     | … | Deliverable/Crite…        │
│   --------------------+--------------------+---+--------------------       │
│    3.1: Revisão de    | Revisar e          | M | Todos os testes           │
│     Test Cases        | atualizar todos os |   | (unitários, de            │
│                       | testes unitários   |   | integração, e2e)          │
│                       | afetados pelas     |   | passam com 100% de        │
│                       | refatorações       |   | cobertura nas             │
│                       | (e.g.,             |   | áreas modificadas.        │
│                       |  cmd/config_test.g |   |                           │
│                       | o ,                |   |                           │
│                       |  cmd/context_test. |   |                           │
│                       | go ,               |   |                           │
│                       |  scanner/filesyste |   |                           │
│                       | m_test.go ).       |   |                           │
│    3.2: Otimização de | Verificar e        | S | Código final              │
│     Imports           | remover imports    |   | limpo.                    │
│                       | não utilizados     |   |                           │
│                       | após o             |   |                           │
│                       | refactoring.       |   |                           │
│                                                                            │
│   ### 3.3. Detailed Action Plan / Phases (continuação)                     │
│                                                                            │
│   (N/A - Continuação nas seções anteriores)                                │
│                                                                            │
│   ### 3.4. Data Model Changes                                              │
│                                                                            │
│   Nenhuma alteração no modelo de dados ou na estrutura de arquivos         │
│   principais (excluindo a remoção de campos legados como  MaxSize  em      │
│   GenerateConfig ). A estrutura de saída do contexto ( <file               │
│   path="..."> ) é mantida.                                                 │
│                                                                            │
│   ### 3.5. API Design / Interface Changes                                  │
│                                                                            │
│   * Modificação de Assinaturas: As funções  parseSize  e  formatBytes      │
│   originais serão movidas, exigindo ajustes nas chamadas internas.         │
│   * Simplificação de Interfaces (Interna): A interface  clipboard.         │
│   ClipboardManager  deve ser simplificada ou removida para reduzir a       │
│   complexidade interna do pacote  clipboard .                              │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 4. Key Considerations & Risk Mitigation                               │
│                                                                            │
│   ### 4.1. Technical Risks & Challenges                                    │
│                                                                            │
│   * Risco: Quebra de Funcionalidade  Clipboard  (Cross-Platform): A        │
│   simplificação da lógica do clipboard pode quebrar a cópia em             │
│   determinadas plataformas ou ambientes de terminal/WSL, já que a          │
│   complexidade na  internal/platform/clipboard  provavelmente visa         │
│   lidar com essas nuances.                                                 │
│     * Mitigação: Testar a funcionalidade de cópia extensivamente em        │
│     pelo menos Linux (xclip/xsel/wl-copy) e macOS (pbcopy) no mínimo,      │
│     e usar o pacote  atotto/clipboard  ou uma abstração                    │
│     comprovadamente minimalista.                                           │
│   * Risco: Inconsistência na Validação de Configuração: Erros na           │
│   migração da lógica de  parseSize  e validação para um novo pacote        │
│   podem levar a configurações incorretas ou falhas de tempo de             │
│   execução.                                                                │
│     * Mitigação: Transferir os testes existentes ( cmd/config_test.go ,    │
│     cmd/context_test.go ) para o novo pacote                               │
│     internal/utils/conversion  e garantir que todos os cenários de         │
│     formato de tamanho (e.g., 1MB, 500KB) sejam cobertos.                  │
│                                                                            │
│                                                                            │
│   ### 4.2. Dependencies                                                    │
│                                                                            │
│   * Interna: As fases 1 e 2 são dependentes uma da outra. A Fase 1         │
│   (Consolidação de Utils) é um pré-requisito para as refatorações da       │
│   Fase 2 (Clipboard/Scanner), pois o código de utilidade será              │
│   referenciado.                                                            │
│   * Externa: Dependência inalterada das bibliotecas principais (           │
│   spf13/cobra ,  spf13/viper ,  charmbracelet/bubbletea , etc.). A         │
│   dependência em  atotto/clipboard  pode se tornar mais direta.            │
│                                                                            │
│   ### 4.3. Non-Functional Requirements (NFRs) Addressed                    │
│                                                                            │
│   * Performance: A remoção de lógica desnecessária e a simplificação       │
│   de caminhos de código contribuem para uma aplicação mais rápida e        │
│   eficiente com menor pegada de memória.                                   │
│   * Maintainability: A consolidação de utilitários e a remoção de          │
│   abstrações e código legado tornam a base de código mais simples,         │
│   mais limpa e mais fácil de entender e manter.                            │
│   * Testability: A separação da lógica de  parseSize  para um pacote       │
│   de utilitários isolado melhora a testabilidade dessas funções.           │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 5. Success Metrics / Validation Criteria                              │
│                                                                            │
│   * Métricas Quantitativas:                                                │
│     * Tamanho do Executável: Redução observável no tamanho final do        │
│     binário (embora modesta).                                              │
│     * Linhas de Código (LOC): Redução líquida nas LOC total do projeto     │
│     (focando na eliminação de arquivos/funções legadas).                   │
│   * Métricas Qualitativas:                                                 │
│     * Testes de Sucesso: Todos os testes unitários e e2e passam            │
│     (principalmente o  TestCLIContextGenerateProducesFile ).               │
│     * Funcionalidade Principal: O fluxo de trabalho principal do TUI       │
│     (seleção de arquivos, geração de contexto) e o comando  context        │
│     generate  CLI funcionam sem regressões.                                │
│     * Revisão de Código: As classes/funções críticas (e.g.,                │
│     cmd/config.go ,  cmd/context.go ,  filesystem.go ) estão livres de     │
│     código legado ou redundante.                                           │
│                                                                            │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 6. Assumptions Made                                                   │
│                                                                            │
│   * A biblioteca  atotto/clipboard  é suficiente para cobrir os            │
│   requisitos de cópia na maioria dos ambientes alvo (Linux/X11/Wayland,    │
│   macOS, Windows).                                                         │
│   * A lógica  parseSize  em  cmd/context.go  e  cmd/config.go  é a         │
│   mesma e pode ser consolidada sem perda de funcionalidade.                │
│   * O campo  MaxSize  na  context.GenerateConfig  é puramente para         │
│   compatibilidade reversa e sua remoção é aceitável, não afetando o        │
│   template rendering que usa  .Config.MaxTotalSize .                       │
│                                                                            │
│   --------                                                                 │
│                                                                            │
│   ## 7. Open Questions / Areas for Further Investigation                   │
│                                                                            │
│   * Estratégia Exata para  Clipboard : Confirmar se o  Manager  é          │
│   estritamente necessário para o  context.go  e se  CopyLarge  é um        │
│   requisito futuro que justifique ser mantido (e refatorado), ou se        │
│   pode ser removido para maximizar o enxugamento.                          │
│   * Abstração de  Diff Split : A funcionalidade de  diff split  em         │
│   cmd/diff.go  deve ser movida para  internal/core/  (e.g.,                │
│   internal/core/diff/ ) se for prevista a reutilização por outros          │
│   comandos ou lógica central. Caso contrário, pode permanecer em  cmd/ .   │
╰────────────────────────────────────────────────────────────────────────────╯