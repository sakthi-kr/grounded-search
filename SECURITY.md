# Security Policy

## Supported versions

Security fixes are applied to the latest development state and the most recent
tagged release.

## Reporting a vulnerability

Do not open a public GitHub issue containing vulnerability details, credentials,
private data, or exploit instructions.

Use the repository's private GitHub vulnerability-reporting feature when it is
available. Otherwise, contact the maintainer privately through the GitHub
profile associated with this repository.

Include the affected component and commit, reproduction steps, expected and
observed behaviour, potential impact, and redacted logs where useful.

## Project security priorities

GroundedSearch treats unauthorised document disclosure as a release-blocking
defect. ACL bypass, cross-user cache leakage, unauthorised model input, prompt
leakage, and invalid citations receive the highest priority.

## Secrets

Never submit real API tokens, passwords, private keys, `.env` files, database
dumps, or confidential documents. Revoke exposed credentials immediately.
