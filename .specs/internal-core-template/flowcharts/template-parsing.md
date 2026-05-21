# Fluxo: Parse e Validação de Template

> **Funções:** `parseTemplate()`, `extractDescription()`, `extractRequiredVars()`, `validateTemplateContent()`
> **Arquivo:** `template.go`

## Fluxograma: parseTemplate

```mermaid
flowchart TD
    Start(["🟢 parseTemplate(content, fileName, filePath)"]) --> EmptyCheck{"content == ''?"}
    EmptyCheck -->|✅ sim| ReturnErr["❌ return nil, error 'template content is empty'"]
    EmptyCheck -->|❌ não| InitTmpl["📦 template := &Template{\n  Name: extractTemplateName(fileName),\n  Content: content,\n  FilePath: filePath,\n  IsEmbedded: true}"]
    InitTmpl --> ExtractDesc["📝 template.Description = extractDescription(content, fileName)"]
    ExtractDesc --> ExtractVars["🔤 vars, err = extractRequiredVars(content)"]
    ExtractVars --> VarsErr{"err != nil?"}
    VarsErr -->|✅ sim| ReturnVarsErr["❌ return nil, error\n'failed to extract required variables'"]
    VarsErr -->|❌ não| SetVars["template.RequiredVars = vars"]
    SetVars --> ReturnOk["✅ return template"]
    ReturnErr --> End(["🔵 FIM"])
    ReturnVarsErr --> End
    ReturnOk --> End
```

## Fluxograma: extractDescription

```mermaid
flowchart TD
    Start(["🟢 extractDescription(content, fileName)"]) --> Split["lines = Split(content, '\\n')"]
    Split --> ForLine["🔁 Para cada line em lines"]
    ForLine --> Trim["line = TrimSpace(line)"]
    Trim --> CheckHeader{"line starts with '#'?"}
    CheckHeader -->|✅ sim| ExtractHeader["description = TrimLeft(line, '#')"]
    ExtractHeader --> CheckDesc{"description != ''?"}
    CheckDesc -->|✅ sim| ReturnDesc["✅ retorna description"]
    CheckDesc -->|❌ não| ForLine

    CheckHeader -->|❌ não| CheckHTML{"line starts with '<!--'"\n&& ends with '-->'?"}
    CheckHTML -->|✅ sim| ExtractHTML["description = Trim(line[4:len(line)-3])"]
    ExtractHTML --> ReturnDesc

    CheckHTML -->|❌ não| CheckBreak{"line != '' &&\n!startsWith('#') &&\n!startsWith('<!--')?"}
    CheckBreak -->|✅ sim| CheckFallback["⏭️ break (sem header encontrado)"]
    CheckBreak -->|❌ não| ForLine

    CheckFallback --> Fallback["📛 default: 'Template for '+extractTemplateName(fileName)"]
    Fallback --> ReturnDefault["✅ retorna fallback"]
    ReturnDesc --> End(["🔵 FIM"])
    ReturnDefault --> End
```

## Fluxograma: extractRequiredVars

```mermaid
flowchart TD
    Start(["🟢 extractRequiredVars(content)"]) --> Matches["matches = variablePattern.FindAllStringSubmatch(content, -1)"]
    Matches --> InitMap["varSet := make(map[string]bool)"]
    InitMap --> ForMatch["🔁 Para cada match em matches"]
    ForMatch --> AddSet{"len(match) > 1?"}
    AddSet -->|✅ sim| varSet["varSet[match[1]] = true"]
    AddSet -->|❌ não| Skip
    varSet --> Skip["⏭️ continue"]

    Skip --> MoreMatches{"mais matches?"}
    MoreMatches -->|✅ sim| ForMatch
    MoreMatches -->|❌ não| Convert["vars := make([]string, 0, len(varSet))"]
    Convert --> ForVar["🔁 Para cada variable em varSet:\n   vars = append(vars, variable)"]
    ForVar --> ReturnVars["✅ retorna vars"]
    ReturnVars --> End(["🔵 FIM"])
```

## Fluxograma: validateTemplateContent

```mermaid
flowchart TD
    Start(["🟢 validateTemplateContent(content)"]) --> EmptyCheck{"content == ''?"}
    EmptyCheck -->|✅ sim| ReturnEmptyErr["❌ error 'template content is empty'"]
    EmptyCheck -->|❌ não| Split["lines = Split(content, '\\n')"]
    Split --> InitCodeBlock["inCodeBlock := false"]
    InitCodeBlock --> ForLine["🔁 Para cada line em lines"]
    ForLine --> Trim["trimmed = TrimSpace(line)"]
    Trim --> CheckCodeBlock{"startsWith('```')?"}
    CheckCodeBlock -->|✅ sim| ToggleCodeBlock["inCodeBlock = !inCodeBlock\n⏭️ continue"]

    CheckCodeBlock -->|❌ não| InBlockCheck{inCodeBlock?"}
    InBlockCheck -->|✅ sim| ContinueBlock["⏭️ continue (ignora code blocks)"]

    InBlockCheck -->|❌ não| CountBraces["open = Count(trimmed, '{')\nclose = Count(trimmed, '}')"]
    CountBraces --> Balanced{"open != close?"}
    Balanced -->|✅ sim| ReturnUnbalanced["❌ error 'unmatched braces on line N'"]
    Balanced -->|❌ não| HasOpen{"contains '{'?"}
    HasOpen -->|✅ sim| ValidatePattern["matches = variablePattern.FindAllString(line)"]
    ValidatePattern --> ForMatchCheck["🔁 Para cada match:\n   !variablePattern.MatchString(match)?"]
    ForMatchCheck -->|✅ sim| ReturnMalformed["❌ error 'malformed variable pattern'"]
    ForMatchCheck -->|❌ não| Continue
    HasOpen -->|❌ não| Continue

    ToggleCodeBlock --> Continue
    ContinueBlock --> Continue
    ReturnUnbalanced --> End(["🔵 FIM"])
    ReturnMalformed --> End
    Continue --> MoreLines{"mais lines?"}
    MoreLines -->|✅ sim| ForLine
    MoreLines -->|❌ não| ReturnOk["✅ return nil"]
    ReturnEmptyErr --> End
    ReturnOk --> End
```

## Regras de Validação de Conteúdo

| Regra | Detalhe |
|-------|---------|
| Conteúdo vazio | Erro imediato |
| Balanceamento de `{}` | Verificado linha a linha (fora de code blocks) |
| Code blocks | `\`\`\`` e `\`\`\`` alternam flag `inCodeBlock`; conteúdo interno é ignorado |
| Padrão regex | Cada `{...}` fora de code blocks deve match `variablePattern` |
| Chaves JSON/Go | `{"key": "value"}` passa (chaves balanceadas) |
| Função `HasVariable` | Usa `strings.Contains` simples — pode false-positivo com sobreposição |
