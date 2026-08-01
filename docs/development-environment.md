# Development Environment

## Repository

| Item | Detected value |
|---|---|
| Repository name | `grounded-search` |
| Local root | `C:/Users/admin/Documents/GroundedSearch` |
| Current branch | `phase-1-foundation` |
| Expected Phase 1 branch | `phase-1-foundation` |
| Origin remote | `https://github.com/sakthi-kr/grounded-search.git` |
| Expected origin remote | `https://github.com/sakthi-kr/grounded-search.git` |
| Working tree before audit | Clean before audit |
| Git `core.autocrlf` | `true` |
| Audit timestamp | `2026-07-30 19:38:31 +0200` |

## Operating environment

| Item | Detected value |
|---|---|
| Shell and platform | `MINGW64_NT-10.0-26200 Sakthi-UW 3.6.9-b4195d69.x86_64 2026-06-06 17:49 UTC x86_64 Msys` |
| WSL status | `Default Version: 2` |

## Required and optional tools

| Tool | Detected version or status | Executable path |
|---|---|---|
| Git | `git version 2.55.0.windows.1` | `/mingw64/bin/git` |
| Go | `Not installed` | `Not installed` |
| Python | `Python 3.10.5` | `/c/Program Files/Python310/python` |
| Python 3 command | `Not installed` | `Not installed` |
| Windows Python launcher | `Python 3.14.5` | `/c/Windows/py` |
| pip through `python -m pip` | `pip 24.3.1 from C:\Users\admin\AppData\Roaming\Python\Python310\site-packages\pip (python 3.10)` | Uses the detected `python` executable |
| Docker CLI | `Not installed` | `Not installed` |
| Docker Compose | `Not installed` | Docker CLI plugin |
| Docker daemon | `Not installed` | Docker Desktop or compatible daemon |
| GNU Make | `Not installed` | `Not installed` |
| GitHub CLI, optional | `gh version 2.96.0 (2026-07-02)` | `/c/Program Files/GitHub CLI/gh` |
| curl | `curl 8.21.0 (x86_64-w64-mingw32) libcurl/8.21.0 Schannel zlib/1.3.2 brotli/1.2.0 zstd/1.5.7 libidn2/2.3.8 libpsl/0.21.5 libssh2/1.11.1 WinLDAP` | `/mingw64/bin/curl` |

## Phase 1 tool requirements

The Phase 1 foundation requires:

- Git;
- Go;
- Python;
- pip associated with the selected Python interpreter;
- Docker Desktop or another Docker-compatible daemon;
- Docker Compose;
- GNU Make, or equivalent verified scripts if Make is unavailable.

GitHub CLI is optional because Git and the GitHub website can complete the
required workflow.

## Windows development rules

- Run repository commands from Git Bash unless a later step explicitly requires
  PowerShell or WSL.
- Keep repository text files in UTF-8 with LF line endings.
- Use `python -m pip` rather than an unqualified `pip` command.
- Do not use the Windows `py` launcher for the project virtual environment
  unless the selected Python version is explicitly verified.
- Docker Desktop must be running before Docker or Compose integration tests.
- Local secrets belong in ignored files such as `.env`, never in tracked files.

## Audit interpretation

This file records detected state only. A listed tool is not automatically
approved for the project merely because it is installed. Version selection and
installation corrections will be made after reviewing this audit.

## Verification commands

```bash
git branch --show-current
git status --short
git remote -v
git --version
go version
python --version
python -m pip --version
docker --version
docker compose version
docker info
make --version
```
