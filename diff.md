╭───────────────────────────────────────────────────────────────────────────────╮
│                                                                               │
│                                                                               │
│                                                                               │
│     ----------                                                                │
│     --- a/internal/ui/screens/file_selection.go                               │
│     +++ b/internal/ui/screens/file_selection.go                               │
│     @@ -75,9 +75,7 @@                                                         │
│          case "right", "l":                                                   │
│              m.tree.ExpandNode()                                              │
│     -    case " ":                                                            │
│     +    case " ", "d":                                                       │
│              m.tree.ToggleSelection()                                         │
│              m.updateSelections(selections)                                   │
│     -    case "d":                                                            │
│     -        m.tree.ToggleDirectorySelection()                                │
│     -        m.updateSelections(selections)                                   │
│          case "i":                                                            │
│              m.tree.ToggleShowIgnored()                                       │
│          case "/":                                                            │
│     @@ -124,7 +122,7 @@                                                       │
│              "↑/↓: Navigate",                                                 │
│              "←/→: Expand/Collapse",                                          │
│              "Space: Select File",                                            │
│     -        "d: Select Directory",                                           │
│     +        "d: Select Dir/File",                                            │
│              "i: Toggle Ignored",                                             │
│              "/: Filter",                                                     │
│              "F5: Rescan",                                                    │
│     ----------                                                                │
│                                                                               │
╰───────────────────────────────────────────────────────────────────────────────╯