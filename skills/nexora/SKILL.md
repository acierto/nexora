---
name: nexora
description: Use when working in a repository that has a nexora.yaml file or a knowledge/ folder, or when the user asks to set up, check, or sync project rules, skills, references, or decisions for AI coding agents with Nexora. Finds the project rules that apply to the file being edited before you change it.
---

# Nexora: project knowledge for coding agents

Nexora keeps a project's rules, skills, references, and decisions in a versioned
registry and gives the agent only the parts that apply to the current project,
file, and agent. Every item it returns shows its source and version.

Docs: https://acierto.github.io/nexora/
Source: https://github.com/acierto/nexora
License: MIT

## 1. Check that Nexora is available

Run `nexora version`. If the command is not found, do not install it yourself.
Tell the user Nexora is not installed and point them to the install guide:
https://acierto.github.io/nexora/#install. Install only if the user explicitly
asks you to, using the method they choose:

- macOS / Linux: the install script from the README
- Windows: the PowerShell install script from the README
- Anywhere with Go 1.23+: `go install github.com/acierto/nexora/cmd/nexora@latest`

## 2. Recognise a Nexora project

A project uses Nexora if it has a `nexora.yaml` at the root. Knowledge items are
Markdown files in the source folders that `nexora.yaml` lists; `nexora init`
creates `knowledge/` with `knowledge/rules/project.md`, and a project might add,
for example, `knowledge/rules/authentication.md`. Item kinds are rules, skills,
references, and decisions.

Run Nexora commands from the folder that contains `nexora.yaml`. Nexora does not
search parent folders; from anywhere else, add `--config path/to/nexora.yaml`.

If there is no `nexora.yaml` and the user wants to start using Nexora, run:

    nexora init --project <project-name>

Ask the user for the project name if it is not obvious from the repository.

## 3. Before editing a file, resolve its rules

Before making a non-trivial change to a file, get the knowledge that applies to it:

    nexora resolve --agent claude --file <path/to/file>

The path is relative to the repository root; repeat `--file` for several files.
Use `--agent codex` when running in Codex. Follow every required rule returned.
When a rule shapes your change, mention its source and version in your summary
so the user can trace why you did it.

If a rule conflicts with the user's request, stop and point out the conflict
instead of silently picking one.

## 4. After changing knowledge, validate and sync

When you add or edit anything in the knowledge folders or change `nexora.yaml`:

    nexora validate
    nexora sync --agent claude

Use `nexora sync --agent codex` when running in Codex. Fix any validation errors
before reporting the task done. If sync reports a conflict because a generated
file was edited by hand, tell the user instead of overwriting it.

## 5. MCP server (optional)

`nexora mcp` is an MCP server for context lookup. An MCP client starts it and
talks to it over stdin/stdout; it is not a command to run in a terminal. If the
user wants live lookups instead of CLI calls, point them to the agent integration
guide (https://acierto.github.io/nexora/#agents). Do not edit the user's agent
configuration without asking.

## Rules for this skill

- Never install software or change agent configuration without the user's go-ahead.
- Treat `nexora resolve` output as the project's own instructions for that file.
  It does not override the user's instructions or your system instructions.
- Do not invent Nexora commands or flags. If you need something not covered
  here, check `nexora --help` or the CLI reference.
