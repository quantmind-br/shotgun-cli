╭────────────────────────────────────────────────────────────────────────────╮
│                                                                            │
│   diff --git a/internal/ui/screens/file_selection.go                       │
│   b/internal/ui/screens/file_selection.go                                  │
│   index 1a6ef77..6975a89 100644                                            │
│   --- a/internal/ui/screens/file_selection.go                              │
│   +++ b/internal/ui/screens/file_selection.go                              │
│   @@ -107,7 +107,7 @@                                                      │
│   "i: Toggle Ignored",                                                     │
│   "/: Filter",                                                             │
│   "F5: Rescan",                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "F8: Next",                                                          │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "F7: Next",                                                          │
│       "F1: Help",                                                          │
│       "Ctrl+Q: Quit",                                                      │
│     ----------                                                             │
│   }                                                                        │
│   diff --git a/internal/ui/screens/rules_input.go                          │
│   b/internal/ui/screens/rules_input.go                                     │
│   index 9979309..9461f66 100644                                            │
│   --- a/internal/ui/screens/rules_input.go                                 │
│   +++ b/internal/ui/screens/rules_input.go                                 │
│   @@ -134,7 +134,7 @@                                                      │
│   content.WriteString(charCount)shortcuts := []string{                     │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "Esc: Focus/Unfocus",                                                │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "Esc: Edit/Done",                                                    │
│       "F8: Next (Skip)",                                                   │
│       "F10: Back",                                                         │
│       "F1: Help",                                                          │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   diff --git a/internal/ui/screens/task_input.go                           │
│   b/internal/ui/screens/task_input.go                                      │
│   index 5240212..6d47b67 100644                                            │
│   --- a/internal/ui/screens/task_input.go                                  │
│   +++ b/internal/ui/screens/task_input.go                                  │
│   @@ -113,7 +113,7 @@                                                      │
│   }                                                                        │
│                                                                            │
│                                                                            │
│     ----------                                                             │
│     shortcuts := []string{                                                 │
│     ----------                                                             │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "Esc: Focus/Unfocus",                                                │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "Esc: Edit/Done",                                                    │
│       "F8: Next",                                                          │
│       "F10: Back",                                                         │
│       "F1: Help",                                                          │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   diff --git a/internal/ui/wizard.go b/internal/ui/wizard.go               │
│   index ba6a80d..66a87c1 100644                                            │
│   --- a/internal/ui/wizard.go                                              │
│   +++ b/internal/ui/wizard.go                                              │
│   @@ -165,7 +165,7 @@                                                      │
│                                                                            │
│                                                                            │
│     ----------                                                             │
│     // Process navigation shortcuts                                        │
│     switch msg.String() {                                                  │
│     ----------                                                             │
│                                                                            │
│   * case "f8", "ctrl+pgdn":                                                │
│                                                                            │
│   * case "f7", "f8", "ctrl+pgdn":                                          │
│   cmd = m.handleNextStep()                                                 │
│   cmds = append(cmds, cmd)                                                 │
│   case "f9":                                                               │
│   @@ -173,7 +173,7 @@                                                      │
│   cmd = m.handleSendToGemini()                                             │
│   if cmd != nil {                                                          │
│   cmds = append(cmds, cmd)                                                 │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       }                                                                    │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       } // F9 is handled here but not for navigation                       │
│     ----------                                                             │
│   case "f10", "ctrl+pgup":                                                 │
│   cmd = m.handlePrevStep()                                                 │
│   cmds = append(cmds, cmd)                                                 │
│   @@ -188,7 +188,14 @@                                                     │
│   }                                                                        │
│                                                                            │
│   func (m *WizardModel) handleNextStep() tea.Cmd {                         │
│                                                                            │
│   * if m.step < StepReview {                                               │
│                                                                            │
│   * // F8: Next step, but on review screen it should trigger               │
│   generation                                                               │
│   * if m.step == StepReview {                                              │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       return m.generateContext()                                           │
│     ----------                                                             │
│                                                                            │
│   * }                                                                      │
│   *                                                                        │
│   * // For all other steps, proceed if valid                               │
│   * // We only check against StepRulesInput as it is the step before       │
│   Review                                                                   │
│   * if m.step <= StepRulesInput {                                          │
│   if m.canAdvanceStep() {                                                  │
│   m.step++                                                                 │
│   return m.initStep()                                                      │
│   @@ -493,12 +500,12 @@// Footer                                           │
│   shortcuts := []string{                                                   │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "F8: Generate",                                                      │
│     ----------                                                             │
│                                                                            │
│                                                                            │
│   *                                                                        │
│                                                                            │
│     ----------                                                             │
│       "F8: Generate / F7: Next",                                           │
│       "F10: Back",                                                         │
│       "F1: Help",                                                          │
│       "Ctrl+Q: Quit",                                                      │
│     ----------                                                             │
│   }                                                                        │
│                                                                            │
│   * footer := styles.RenderFooter(shortcuts)                               │
│                                                                            │
│   * footer := styles.RenderFooter(shortcuts) // Footer rendering will      │
│   handle F7/F8 distinction                                                 │
│   view.WriteString(footer)return view.String()                             │
│                                                                            │
│                                                                            │
│     ----------                                                             │
│     ----------                                                             │
│                                                                            │
╰────────────────────────────────────────────────────────────────────────────╯