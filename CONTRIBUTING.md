# Contributing to Shotgun CLI

Thanks for helping improve Shotgun. Small, focused changes are easiest to
review.

## Development setup

Requirements: Go 1.23 or newer, Git, and `make`.

```bash
git clone https://github.com/quantmind-br/shotgun-cli.git
cd shotgun-cli
go mod download
make test
```

Before opening a pull request, run:

```bash
make test
make lint
```

Include tests for behavioral changes. Do not include API keys, generated
contexts, local configuration, or source code from repositories you do not
have permission to redistribute.

## Pull requests

- Describe the user-visible problem and the chosen solution.
- Keep unrelated refactors in separate pull requests.
- Update documentation when commands, flags, or configuration change.
- Confirm that the relevant tests pass on your platform.

For vulnerabilities, follow [SECURITY.md](SECURITY.md) instead of opening a
public issue.
