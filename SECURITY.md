# Security policy

## Reporting a vulnerability

Please report vulnerabilities through GitHub's private vulnerability reporting
for this repository. Do not open a public issue containing exploit details,
credentials, or private source code.

Include the affected version, reproduction steps, impact, and any suggested
mitigation. You can expect an acknowledgement within seven days.

## Sensitive data

Shotgun can read source trees and generate context files for external AI
providers. Review generated context before sharing it, configure ignore rules
for sensitive paths, and keep provider credentials outside version control.
Users are responsible for ensuring that submitted source code may be processed
by the selected provider.
