# TUI Validator — Audit Report

**Application**: `/tmp/shotgun-tui-audit`
**Args**: ``
**Working directory**: `/home/diogo/dev/shotgun-cli`
**Timestamp**: `20260520T185002Z` UTC
**Pipeline**: `tui-validator` skill (tmux + grim + capture-pane)
**Workspace**: `/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z`

---

## 1. Executive Summary

Drove the full 5-step Bubble Tea wizard (File Selection → Template Selection →
Task Input → Rules Input → Review) through tmux + grim, probed all 55
documented keybindings across every context (global, step1-files,
step2-template, modal-preview, step3-task-input, step4-rules-input,
step5-review), pushed Latin-extended / CJK / emoji / combining marks /
bracketed paste through the text areas, then resized the pane across five
geometries (60×20, 80×24, 160×40, 80×50, 200×60) plus two below-minimum
sanity checks (50×18, 40×14). No crashes, no data loss, all three quit paths
(`q`, `Ctrl-Q`, `Ctrl-C`) exit cleanly. Generation flow (`g`) works end-to-end:
prompt file written, clipboard populated. Two **major** findings (Tab
hijacks text-input focus toggle; help screen is unscrollable and clips the
top half at 80×24) should be fixed before the next release; everything else
is polish.

**Severity breakdown:**

| Severity | Count             |
| -------- | ----------------- |
| Blocker  | 0  |
| Major    | 2    |
| Minor    | 5    |
| Cosmetic | 3  |

| Audit stat                     | Value                |
| ------------------------------ | -------------------- |
| Captures (text + ANSI)         | 70         |
| Screenshots                    | 12        |
| Keybindings inventoried        | 55        |
| Initial geometry               | 80 × 24 |
| TERM                           | `xterm-256color`         |

**Verdict**: ship-ready for power users; recommend the two majors land in
the next patch release because they affect documented behavior the help
screen itself promises.

---

## 2. Methodology

### Phases executed

| Phase | What was done | Status |
| ----- | ------------- | ------ |
| 1. Discover  | Located binary at `/tmp/shotgun-tui-audit` (rebuilt from `github.com/quantmind-br/shotgun-cli` because the supplied artifact was a Go `.a` archive, not an ELF). Verified `--help` and confirmed flagless invocation launches the wizard. Read `AGENTS.md`, `TUI_REFACTOR.md` for documented bindings. | ✅ |
| 2. Inventory | Sent `?` → help renders; resized to 100×60 to see the full screen and parsed 55 bindings across 7 contexts into `keybindings.json`. Confirmed status-bar hints on every step too. | ✅ |
| 3. Probe     | Drove every documented binding from Step 1 → Step 5 → modal → back. Classified each; captured before/after pairs as numbered text scrapes. Tested all three quit paths individually (each requires a fresh launch). | ✅ |
| 4. Stress    | Sent Latin-extended (`á é í ó ú ã õ ç ñ ü`), symbols (`€ £ ¥ © ® § ¶`), CJK (`中 日 韓 한`), emoji including ZWJ (`😀 🚀 👨‍👩‍👧`), combining marks (`é+̌ ã+̃`), 3-line bracketed paste, and rapid 20-char typing through the Step 3 text area. Also probed unbound `C-g / C-n / C-p` and undocumented F-keys. | ✅ |
| 5. Visual    | Pixel-screenshotted Step 1 at five sizes (60×20, 80×24, 160×40, 80×50, 200×60) plus two below-minimum (50×18, 40×14). Repeated for Step 5 at three sizes (60×20, 80×24, 160×40). | ✅ |
| 6. Report    | This document. | ✅ |

### Coverage

- **Keys probed**: 55 documented + 11 undocumented (`C-c`, `C-g`, `C-n`, `C-p`, `F2`, `F3`, `F4`, `F6`, `F10`, `F11`, `F12`, plus 20-key burst of `j`/`k`). Total ~66.
- **Modes tested**: Step 1 (file tree), Step 1 filter mode, Step 1 with `i` (ignored files visible), Step 2 (template list), Template Preview modal (via Enter and `v`), Step 3 (task input, focused + blurred), Step 4 (rules input, focused + blurred), Step 5 (pre-generation, post-generation).
- **Geometries**: 60×20, 80×24, 80×50, 100×60 (help inspection), 160×40, 200×60. Plus below-min: 50×18, 40×14.
- **Not tested (and why)**:
  - **Destructive keys**: skipped by default per skill policy. The audited TUI doesn't expose destructive keys in the documented set (no `d/D/Delete`/`Ctrl-K`) so nothing was intentionally omitted on that account.
  - **Send-to-LLM end-to-end**: no LLM provider is configured in the test environment, so `s` and `F9` were sent and confirmed silent — the actual HTTP path was not exercised.
  - **Config TUI**: only the wizard was audited; `shotgun-cli config` opens a separate sidebar+editor+quit-confirm TUI per `AGENTS.md` and was out of scope for this run.

### Limitations

- **Wayland-only screenshots**: pixel screenshots were taken via `grim` on Hyprland; `xdg-toplevel-icon` warnings printed but didn't affect quality. The compositor was active during the run.
- **`/tmp/shotgun-tui-audit` was a Go `.a` object archive on arrival**, not an executable. Rebuilt with `go build -o /tmp/shotgun-tui-audit .` from the repo root before launching. All findings reference the rebuilt binary.
- The TUI generated a real `shotgun-prompt-20260520-154642.md` file during the `g` probe; it was deleted after the test.
- Rapid-key timing tests used `tmux send-keys` to emit events; this matches a fast-typist key-roll. Lower-level USB-keyboard timing was not simulated.

---

## 3. Keybindings Inventory

55 bindings inventoried across 7 contexts: `global`, `step1-files`,
`step2-template`, `modal-preview`, `step3-task-input`,
`step4-rules-input`, `step5-review`. Every binding listed in the in-app help
screen was probed at least once and confirmed active, with three exceptions
documented as findings F-01 / F-02 / F-04 in §4. All status-bar hints
matched the help screen content except where called out in F-03 / F-08.

The full bindings table is auto-generated at the bottom of this report from
`keybindings.json` (raw file:
`/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/keybindings.json`).

Notable inventory observations:
- `Tab` is bound twice in steps 3-4 (global "next step" + local "toggle
  focus"). The global handler wins — see F-01.
- `q` is contextual: literal character in text inputs (steps 3-4), quit
  everywhere else. Correctly disambiguated.
- `j/k` are bound in step1, step2, and the preview modal. The modal version
  is rate-limited (F-04).
- No undocumented bindings discovered.

---

## 4. Findings

Ten findings total. The two majors (F-01 Tab-in-text-inputs, F-02 unscrollable
help) are the only ones that affect documented user-facing behavior at the
default 80×24 geometry. All others are polish: status-bar gaps,
truncation at the smallest supported size, missing feedback when LLM is
not configured, and a wide-terminal layout that doesn't grow.

The detailed entries are auto-appended below in the `## Findings` section
(rendered from `findings.json`).

---

## 5. Visual Gallery

Twelve PNGs covering Step 1 and Step 5 at multiple geometries plus the help
screen captured at 100×60.

Notable visual observations (the auto-rendered gallery follows below):
- `step1-tiny-60x20.png` — the status bar overlaps the last visible tree row
  at this size (F-05).
- `extreme-50x18.png` / `extreme-40x14.png` — the "Terminal too small (need
  ≥60×16)" guard renders cleanly with the current size displayed. This is a
  positive observation: the TUI does NOT crash at sub-minimum sizes.
- `help-full-100x60.png` — only at this size does the entire help screen
  fit. The default 80×24 cuts off the entire Global Shortcuts section
  (F-02).
- `step1-wide-160x40.png` / `step1-huge-200x60.png` — show the progress bar
  pinned to ~76 columns regardless of pane width (F-10).
- `step5-tiny-60x20.png` — visible text clipping and a dropped bullet
  (F-06).

No `magick compare`-based pixel diffs were generated because no screen-pair
showed visually different rendering of the same content (different sizes
produce intentionally different content, not visual regression).

---

## 6. Reproducibility

Every major finding above is reproducible from a fresh launch:

| Finding | Repro from fresh boot? | Steps |
| ------- | ---------------------- | ----- |
| F-01    | yes | Launch → Down×8, Space (pick a file) → F8 → Space, F8 → type any text → press Tab. The wizard jumps to Step 4 instead of toggling focus. Help still says "Tab — Toggle focus (edit/done)". |
| F-02    | yes | Launch at default 80×24 → press `?`. The visible top of the screen begins at "↑/↓ or k/j  Navigate templates" — the Global Shortcuts and File Selection sections are above the viewport. PgUp / Up / Home / `g` / `k` all do nothing. Resize ≥46 rows tall to see the rest. |
| F-04    | yes (timing-dependent) | Launch → Down×8, Space, F8 → Space, Enter (open preview modal). Tap `j` six times quickly (no inter-key delay): the `[Lines N-M of K]` indicator advances only once. Wait 150 ms between `j` presses → it advances per-press. |
| F-05    | yes | Launch → resize pane to 60×20. The "↑/↓: Navigate │ ←/→: Exp" status bar overlaps the last file row in the tree. |
| F-06    | yes | Launch → Down×8, Space → F8 → Space, F8 → type "x" → F8, F8 (lands on Step 5) → resize to 60×20. Right-hand text clips and the "Copy the content to your clipboard" bullet drops. |

---

## 7. Improvement suggestions (non-bugs)

- **Add an "LLM not configured" toast on `s`/`F9`** so users know whether
  the binding fired (covers F-03's silent-no-op observation).
- **Surface F-key cheat strip** at the bottom of the status bar: e.g.
  `F1 Help · F5 Rescan · F7 Back · F8 Next · F9 Send` — addresses F-07.
- **Persistent filter indicator** at the top of the Step 1 tree when a
  filter is active (`Filter: AGE  [x to clear]`) — addresses F-09.
- **Wide-mode layout**: at ≥120 columns, render the file tree and a
  selection-summary side panel in a two-column layout (currently only the
  tree uses the extra width, awkwardly).
- **Step header consistency**: print "Step X/5 • Name" as a dedicated row
  on every step including 3 and 4 — addresses F-08.

---

## 8. Prioritized recommendations

| Priority | Item | Resolves |
| -------- | ---- | -------- |
| P0       | Fix Tab in steps 3–4 to either toggle focus (preferred) or update help text | F-01 |
| P0       | Make the help screen scrollable (reuse the Template Preview viewport) | F-02 |
| P1       | Add `c: Copy │ s/F9: Send to LLM` to the Step 5 status bar pre- and post-generation; show LLM-not-configured warning | F-03 |
| P1       | Investigate j/k rapid-press coalescing in the Template Preview modal — match arrow-key behavior | F-04 |
| P1       | Reserve 2 status-bar rows in the Step 1 tree viewport calculation | F-05 |
| P2       | Wrap-or-ellipsize Step 5 panel content at 60-column width | F-06 |
| P2       | Add "Step X/5 • Name" header to Steps 3 and 4 | F-08 |
| P2       | Persistent filter indicator on Step 1 | F-09 |
| P2       | Make progress bar width scale with pane width | F-10 |
| P3       | F-key cheat strip in status bar | F-07 |

---

## 9. Workspace

```
/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/
├── meta.json
├── keybindings.json
├── findings.json
├── captures/      (70 text + ANSI scrapes)
└── screenshots/   (12 PNGs)
```

---

## 10. Appendix — environment

- **TERM**: `xterm-256color`
- **Initial geometry**: 80 × 24
- **Binary**: `/tmp/shotgun-tui-audit`
- **Args**: ``
- **CWD**: `/home/diogo/dev/shotgun-cli`
- **OS / compositor**: Linux + Hyprland (Wayland)
- **Screenshot tool**: `grim` (with `aha → chromium` fallback never needed)
- **Source repo**: `github.com/quantmind-br/shotgun-cli`, branch `main`,
  built fresh with `go build -o /tmp/shotgun-tui-audit .`
- **Bubble Tea**: v1.3.5 (per project AGENTS.md)
- **Test fixture for file scan**: the shotgun-cli repo itself
  (`/home/diogo/dev/shotgun-cli`).
- **Sessions used**: 4 tmux sessions across the audit (each quit-path test
  required a fresh launch; visual phase restarted once for clean baseline).
- **Generated artifact (deleted after audit)**:
  `shotgun-prompt-20260520-154642.md`.

---

## Keybindings inventory

| Key | Context | Description | Source |
| --- | --- | --- | --- |
| `F1` | global | Toggle help screen | documented+observed |
| `?` | global | Toggle help (alias of F1) | documented+observed |
| `F7` | global | Previous step | documented |
| `M-Left` | global | Previous step (Alt+←) | documented |
| `F8` | global | Next step | documented |
| `M-Right` | global | Next step (Alt+→) | documented |
| `Tab` | global | Next step alias (also toggles focus in text inputs) | documented |
| `BTab` | global | Previous step alias (Shift+Tab) | documented |
| `q` | global | Quit application | documented |
| `C-q` | global | Quit (Ctrl+Q) | documented |
| `C-c` | global | Hard quit (SIGINT, not documented in help) | inferred |
| `Up` | step1-files | Navigate up | documented |
| `Down` | step1-files | Navigate down | documented |
| `k` | step1-files | Navigate up (vim) | documented |
| `j` | step1-files | Navigate down (vim) | documented |
| `Left` | step1-files | Collapse directory | documented |
| `Right` | step1-files | Expand directory | documented |
| `h` | step1-files | Collapse (vim) | documented |
| `l` | step1-files | Expand (vim) | documented |
| `Space` | step1-files | Toggle file/directory selection | documented |
| `a` | step1-files | Select all visible files | documented |
| `A` | step1-files | Deselect all visible files | documented |
| `i` | step1-files | Toggle showing ignored files | documented |
| `/` | step1-files | Enter filter mode (fuzzy search) | documented |
| `x` | step1-files | Clear filter | documented |
| `F5` | step1-files | Rescan directory | documented |
| `r` | step1-files | Rescan directory (alias of F5) | documented |
| `Up` | step2-template | Navigate templates up | documented |
| `Down` | step2-template | Navigate templates down | documented |
| `k` | step2-template | Navigate up (vim) | documented |
| `j` | step2-template | Navigate down (vim) | documented |
| `Space` | step2-template | Select template | documented |
| `Enter` | step2-template | Open preview modal | documented |
| `v` | step2-template | View full template (modal) | documented |
| `j` | modal-preview | Scroll down | documented |
| `k` | modal-preview | Scroll up | documented |
| `PgUp` | modal-preview | Page scroll up | documented |
| `PgDn` | modal-preview | Page scroll down | documented |
| `g` | modal-preview | Jump to top | documented |
| `G` | modal-preview | Jump to bottom | documented |
| `Escape` | modal-preview | Close modal | documented |
| `q` | modal-preview | Close modal (alias) | documented |
| `Tab` | step3-task-input | Toggle focus (edit/done) | documented |
| `Escape` | step3-task-input | Cancel / leave input | documented |
| `Enter` | step3-task-input | New line | documented |
| `BSpace` | step3-task-input | Delete character | documented |
| `Tab` | step4-rules-input | Toggle focus (edit/done) | documented |
| `Escape` | step4-rules-input | Cancel / leave input | documented |
| `Enter` | step4-rules-input | New line | documented |
| `BSpace` | step4-rules-input | Delete character | documented |
| `F8` | step5-review | Generate context | documented |
| `g` | step5-review | Generate context (alias) | documented |
| `c` | step5-review | Copy to clipboard | documented |
| `F9` | step5-review | Send to LLM (if configured) | documented |
| `s` | step5-review | Send to LLM (alias) | documented |

## Findings

### [MAJOR] Tab in Step 3/4 text inputs always advances to next step instead of toggling focus (docs say it should toggle)

**Phase:** 3  
**Evidence:** captures/0037-step3-typed.txt → 0038-step3-tab-blurred.txt: focus='Editing', user types 'Test task', presses Tab → wizard jumps to Step 4. Repeated test from blurred state ('Press Tab to edit'): captures/0039-step3-escape-blurred-state.txt → 0040-step3-tab-blurred-state.txt also advances to Step 4. Help screen (captures/0004-help-full-100x60.txt) explicitly documents 'Text Input (Steps 3-4): Tab — Toggle focus (edit/done)'.  

The help screen documents Tab as a focus toggle inside the text-input steps, but the global 'Tab = next step alias' handler always wins. Users typing in the Task or Rules input cannot use Tab to leave the field — pressing Tab navigates them to the next step (potentially submitting an incomplete task description). The only way to blur the field is Escape (which works).

**Suggested fix:** Either (a) consume Tab inside the focused text-area component and only delegate to the wizard's next-step handler when the field is blurred, or (b) update the help text to drop the 'Tab = Toggle focus' line and explain that Escape is the only blur key. Option (a) preserves the more discoverable convention; option (b) matches current behavior with a one-line doc fix.

---

### [MAJOR] Help screen is taller than 24 rows and is not scrollable — top half hidden at the documented minimum geometry

**Phase:** 3  
**Evidence:** captures/0002-help-question.txt (80x24): the file shows the screen starts at '↑/↓ or k/j  Navigate templates' (i.e. the Template Selection section), with no 'Global Shortcuts' / 'File Selection' sections visible. captures/0003-help-top.txt (80x24, after sending g, Home, PgUp, Up*10) is identical → none of those keys scroll the help. captures/0004-help-full-100x60.txt at 100x60 shows the complete content starting with 'Help - Keyboard Shortcuts / Global Shortcuts / F1 / ?  Toggle this help screen ...'. At the default 80x24 size, F1/F7/F8/Tab/Shift+Tab/q/Ctrl+Q are documented but unreachable from inside the help.  

The help text needs ~46 rows to render in full but the help screen has no scrollback handling — PgUp, Up, Home, k, g all return the same view. Users on the documented minimum 80x24 will only see the last ~17 rows of help, missing the entire Global Shortcuts and File Selection sections. The Template Preview modal (which is scrollable, has its own j/k/g/G/PgUp/PgDn bindings) only proves the design pattern is available — the help screen just doesn't use it.

**Suggested fix:** Wrap the help content in the same scrollable viewport used by the Template Preview modal (j/k/g/G/PgUp/PgDn), or add a 'Page X/Y' indicator. As a short-term fix, reorder so 'Global Shortcuts' (the most important section) renders at the bottom so it survives truncation at 80x24.

---

### [MINOR] Step 5 status bar omits documented `c`, `s`, `F9` shortcuts before generation; review modal status footer disagrees with help

**Phase:** 3  
**Evidence:** captures/0042-step5-initial.txt (before generation): footer shows '↑/↓: Scroll │ g: Generate │ F8: Generate │ F7: Back / F1/?: Help │ q: Quit' — no mention of c/s/F9. Help (captures/0004-help-full-100x60.txt) lists 'Review (Step 5): F8 / g  Generate context, c  Copy to clipboard, F9 / s  Send to LLM (if configured)'. After generation (captures/0043-step5-after-g.txt) the footer becomes '↑/↓: Scroll │ c: Copy / F1/?: Help │ q: Exit' — still no s/F9. Probe confirmed s and F9 work pre-generation (s/F9 require a generated context to send) but produced no visible feedback either way.  

Users have to consult the help screen to discover c/s/F9 — they are documented but not surfaced in the contextual status bar. After generation the new bar appears but still hides s/F9. Worse: s and F9 silently no-op when no LLM is configured (no toast / error / status line change), so the user can't tell whether the key fired and the LLM is misconfigured, vs. the key not being bound at all.

**Suggested fix:** Add `c: Copy │ s/F9: Send to LLM` to the Step 5 status bar (both before and after generation). When s/F9 is pressed but no LLM is configured, show an inline message like `⚠ No LLM configured — run 'shotgun-cli config'`. Bonus: dim/grey-out s/F9 in the status bar when not configured.

---

### [MINOR] Rapid `j`/`k` presses in Template Preview modal coalesce — only the first event advances the scroll

**Phase:** 3  
**Evidence:** captures/0031..0033 with j/k batched (no inter-key delay): sending `j j j j j j` only moved one line (Lines 1-16 → 2-17 then stayed). Re-test with 150 ms delays between presses (per inline log in transcript) scrolled correctly one line per press (8-23 → 9-24 → 10-25 → 11-26 → 12-27). Arrow keys (Down, Down*5) advance correctly on every press regardless of delay.  

When j/k arrive faster than the modal's redraw cycle, only the first key event is consumed; the rest are dropped. The arrow keys do not exhibit this — they're processed per-event. The combined effect is that vim-style scrolling feels broken to anyone who key-rolls.

**Suggested fix:** Make sure the modal's KeyMsg handler for j/k follows the same path as Up/Down (queue per-event instead of debouncing). Or, if debouncing is intentional, apply it uniformly so the inconsistency disappears.

---

### [MINOR] At 60×20 (still inside the supported geometry range) the status bar overlaps the last visible file row

**Phase:** 5  
**Evidence:** captures/0061-step1-tiny-60x20.txt: bottom of the tree shows `   ├──[ ] TUI_REFACTOR.md (3.2 KB)│ ↑/↓: Navigate │ ←/→: Exp` — the file row text and the status bar share the same line, with the status bar overlay starting mid-row. Same overlap visible in screenshots/step1-tiny-60x20.png. At 80×24 (captures/0001-initial.txt) and larger, the status bar lives on its own row.  

The TUI claims a minimum of 60×16 (see captures/0066-extreme-50x18.txt 'Terminal too small (need ≥60x16)') and 60×20 is well inside that band, but the file-tree panel doesn't reserve enough rows for the status bar, producing visible text overlap on the row above the bar.

**Suggested fix:** Subtract status-bar height (2 rows) from the tree viewport when computing how many entries to render, or pin the status bar to the bottom and let the tree clip with `...` rather than overlap.

---

### [MINOR] Step 5 truncates non-fatally at 60×20 — closing `)` after token count and 'Copy clipboard' bullet drop

**Phase:** 5  
**Evidence:** captures/0070-step5-tiny.txt at 60×20: line `📁 Selected Files: Selected: 1 files (13.7 KB / ~3.5K tokens` ends without `)`; the 'Save it as shotgun-prompt-YYYYMMDD-HHMMSS.md' line is truncated mid-name; the 'Copy the content to your clipboard' bullet is missing entirely.  

Step 5 renders fine at 80×24 and larger but at 60×20 (still inside the supported range) the right-hand text clips mid-character and a bullet line is dropped because the panel ran out of rows. No misleading information is shown, but the rendering is visibly broken.

**Suggested fix:** Either wrap long lines inside the Step 5 panel, or render an ellipsis (`…`) when clipping. The dropped bullet is harder — pick a minimum that fits all bullets (≥22 rows) or stack them tighter.

---

### [MINOR] F2/F3/F4/F6/F10/F11/F12 are silently unbound — could be intentional, but no feedback distinguishes 'no binding' from 'binding broken'

**Phase:** 3  
**Evidence:** captures/0048-step5-after-undef.txt: pane hash before sending = after sending F2/F3/F4/F6/F10/F11/F12 (md5 444bfdd…). Same is true on Step 1 and Step 2 (verified during probe).  

Pressing any unbound F-key gives zero feedback. This is conventional, but in a TUI that uses F1/F5/F7/F8/F9 it's reasonable for a user to try F2/F3/F4/F6 looking for adjacent actions (e.g. F2 = Rename in many file managers). A faint 'unbound' flash or no-op silently is fine, but currently there's no way to discover what F-keys do exist short of opening the help.

**Suggested fix:** Optional: add a one-line cheat (`F1 Help · F5 Rescan · F7 Back · F8 Next · F9 Send`) at the very bottom of the status bar on Step 1/2 — it would make F-keys discoverable without opening help.

---

### [COSMETIC] Status bar header for Step 3/4 is missing the 'Step 3/5 • …' title at 80×24

**Phase:** 3  
**Evidence:** captures/0036-step3-initial.txt: line 1 is the description text 'Enter a detailed description of what you want to accomplish. Be specific about r' (the line is also truncated mid-word at column 79). Compare captures/0024-step2-via-F8.txt and captures/0042-step5-initial.txt where line 1 reads 'Step 2/5 • Choose Template' and 'Step 5/5 • Review & Generate' respectively.  

Steps 1, 2 and 5 print a 'Step X/5 • <Name>' header on line 1. Steps 3 and 4 drop that header in favor of the description text. Result: there's no visible cue that you are on Step 3 vs Step 4 (the OPTIONAL marker on Step 4's description is the only hint). Once you've Tab-jumped past a step, you can't tell where you are at a glance.

**Suggested fix:** Always render the 'Step X/5 • Name' header on its own line, and put the description on the line below.

---

### [COSMETIC] Filter mode applied state is not visually distinct from the unfiltered tree (filter input field disappears once Enter is pressed)

**Phase:** 3  
**Evidence:** captures/0017-filter-mode-open.txt: shows `Filter: _` input field. captures/0019-filter-applied-AGE.txt: tree is narrowed (only AGE-matching files visible), but the 'Filter: AGE' indicator is gone — only the narrowed result list signals that a filter is active.  

Once Enter is pressed to apply a filter, the visual filter indicator vanishes. New users (especially after `x` to clear) may wonder why some files are missing — there's no persistent indicator like 'Filtered: AGE' near the top.

**Suggested fix:** Keep a small 'Filter: AGE  [x to clear]' line at the top of the tree whenever a filter is active.

---

### [COSMETIC] Progress bar pinned to ~76 cells; doesn't scale on wide terminals

**Phase:** 5  
**Evidence:** captures/0063-step1-wide-160x40.txt and 0065-step1-huge-200x60.txt: the `░░░…` progress bar fills only the first ~76 columns out of 160/200. Centered or not, it leaves a large right-hand gutter empty.  

At 160×40 and 200×60 the progress bar (showing context-size budget) doesn't grow with the available width, so the right-hand part of the screen is mostly whitespace. Functional, but wastes wide-monitor real estate.

**Suggested fix:** Compute progress-bar width as a function of `pane_width - margin` rather than a fixed constant.

## Visual gallery

#### extreme-40x14

![extreme-40x14](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/extreme-40x14.png)

#### extreme-50x18

![extreme-50x18](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/extreme-50x18.png)

#### help-full-100x60

![help-full-100x60](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/help-full-100x60.png)

#### initial-step1

![initial-step1](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/initial-step1.png)

#### step1-default-80x24

![step1-default-80x24](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step1-default-80x24.png)

#### step1-huge-200x60

![step1-huge-200x60](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step1-huge-200x60.png)

#### step1-tall-80x50

![step1-tall-80x50](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step1-tall-80x50.png)

#### step1-tiny-60x20

![step1-tiny-60x20](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step1-tiny-60x20.png)

#### step1-wide-160x40

![step1-wide-160x40](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step1-wide-160x40.png)

#### step5-default-80x24

![step5-default-80x24](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step5-default-80x24.png)

#### step5-tiny-60x20

![step5-tiny-60x20](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step5-tiny-60x20.png)

#### step5-wide-160x40

![step5-wide-160x40](/home/diogo/.cache/tui-validator/shotgun-tui-audit/20260520T185002Z/screenshots/step5-wide-160x40.png)


