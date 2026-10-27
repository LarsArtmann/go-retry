# Security Policy

## Security Practices

- **Dependency minimality** — `go-error-family` is the sole runtime dependency;
  the nested `tools/` module (actionlint, govulncheck) is never downloaded by
  consumers.
- **Automated vulnerability scanning** — CI runs `govulncheck` on every push
  and pull request via the official GitHub action.
- **Pinned supply chain** — GitHub Actions are pinned by commit SHA with their
  inputs guarded by an allowlist test; Dependabot owns dependency updates.
- **No secrets in the repository** — the library needs no credentials, and CI
  requires none.

## Supported versions

Security fixes are applied to the latest release on the `master` branch. Older
versions may not receive security updates.

## Reporting a vulnerability

Please report suspected vulnerabilities privately through
[GitHub Security Advisories](https://github.com/LarsArtmann/go-retry/security/advisories/new).
Include the affected version, a concise description, reproduction steps, and
any relevant logs or stack traces. Do not include secrets or sensitive user
data.

If a private advisory cannot be created, open a GitHub issue requesting a
security contact without disclosing vulnerability details.

Reports are acknowledged within two business days and triaged within five.
Please allow reasonable time for investigation and coordinated disclosure.
