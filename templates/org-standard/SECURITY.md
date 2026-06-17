# Security Policy

## Reporting a vulnerability

**Do not** open public GitHub issues for security vulnerabilities.

Report privately via **[GitHub Private Vulnerability Reporting](https://docs.github.com/en/code-security/security-advisories/working-with-repository-security-advisories/configuring-private-vulnerability-reporting-for-a-repository)**:

1. Open this repository on GitHub
2. Go to **Security** → **Report a vulnerability**

Repository maintainers should enable private vulnerability reporting in **Settings → Security → Private vulnerability reporting**.

## What to include

- Affected version or commit SHA
- Steps to reproduce
- Impact assessment (data exposure, auth bypass, RCE, etc.)

## Response SLA

| Stage | Target |
|-------|--------|
| Acknowledgment | 72 hours |
| Initial triage | 7 days |
| Fix or mitigation plan | 30 days (severity-dependent) |

## Supported versions

| Version | Supported |
|---------|-----------|
| Latest release tag | Yes |
| `main` / `master` | Yes |
| Older tags | Best effort |

## Scope

In scope: this repository and its official release artifacts.

Out of scope: third-party dependencies (report to upstream); social engineering.
