# Agentlink

Keep your AI instruction files in sync with **zero magic** — just symlinks.

Different tools want different files at project root: `AGENTS.md` (OpenAI/Codex, OpenCode), `CLAUDE.md` (Claude Code), `GEMINI.md`, etc. There's no standard, and I'm not waiting for one. **Agentlink** solves the basic need: keep your **personal** instruction files (in `~`) and your **project** instruction files in sync **without generators**. Edit one, they all reflect it.

Creating instruction files is easy with `/init` commands, but keeping them up to date is the hard part — and expensive too. Good instruction files are often crucial and make a huge difference when using agentic tools. Since they're so important, these files are typically generated with expensive models. Why pay repeatedly to regenerate similar content across different tools?

**Future-proof by design:** We don't know what tomorrow brings in the AI tooling space, but agentlink is ready. New tool expects `.newtool/ai-config.md`? Just add it to your config. Complex nested structure like `workspace/ai/tools/newframework/instructions.md`? No problem. Agentlink automatically creates the directories and symlinks without any code changes needed.

> Scope: **instruction files only**. No MCP `.mcp.json` or chain configs. Simple on purpose.

---

## Why Agentlink?

- **One real file, many aliases** — pick a *source* (`CLAUDE.md` or `AGENTS.md` or whatever), symlink the rest.
- **No codegen** — no templates, no transforms, no surprise diffs.
- **Project + global** — works in repos *and* under `~/.config/…`.
- **Idempotent** — re-run safely; it fixes broken/misdirected links.
- **Portable** — works on macOS and Linux.
- **Future-ready** — handles any directory structure, automatically creates paths. Tomorrow's AI tool? Just add its path.

---

## How it works

You tell Agentlink which file is the **source**, and which other files should **link** to it. Agentlink creates/fixes symlinks accordingly.

```yaml
# .agentlink.yaml (in project root)
source: CLAUDE.md
links:
  - AGENTS.md                              # OpenCode, Codex
  - .github/copilot-instructions.md       # GitHub Copilot  
  - .cursorrules                           # Cursor AI
  - GEMINI.md                              # Gemini CLI
```

Result:
```
./CLAUDE.md                              # real file you edit
./AGENTS.md                           -> CLAUDE.md  (symlink)
./.github/copilot-instructions.md     -> ../CLAUDE.md  (symlink)
./.cursorrules                        -> CLAUDE.md  (symlink)
./GEMINI.md                           -> CLAUDE.md  (symlink)
```

Global mode (in HOME) is the same idea:

```yaml
# ~/.config/agentlink/config.yaml
source: ~/.config/claude/CLAUDE.md
links:
  - ~/.config/opencode/AGENTS.md
  - ~/.config/some-tool/INSTRUCTIONS.md
```

---

## Install

Planned:

- **Homebrew (tap)**  
  ```bash
  brew tap <you>/agentlink
  brew install agentlink
  ```

- **AUR (`agentlink-bin`)**  
  ```bash
  yay -S agentlink-bin
  ```

- **Direct download (GitHub Releases)**  
  Single static binary.

> We’ll wire this up via GoReleaser.

---

## Usage

### Getting started

```bash
# Initialize in your project
agentlink init

# Edit the created .agentlink.yaml to match your needs
# Create your source file (e.g., CLAUDE.md)

# Sync to create symlinks
agentlink sync
```

### Commands

```bash
agentlink init               # create .agentlink.yaml in current directory
agentlink sync               # create/fix symlinks based on config
agentlink check              # print status and problems
agentlink clean              # remove managed symlinks (non-destructive)
agentlink doctor             # environment + permissions sanity checks
```

### Helpful flags

```bash
agentlink sync --dry-run     # show what would change
agentlink sync --force       # replace wrong/missing links (or -f)
agentlink --verbose          # detailed output for any command (or -v)
```

### Without init (auto-config)

```bash
# In a project with .agentlink.yaml
agentlink sync
```

What it does:
- Reads `.agentlink.yaml` in CWD.
- Creates/fixes symlinks listed under `links:` so they point to `source`.

If there's **no** `.agentlink.yaml` in CWD:
- Falls back to `~/.config/agentlink/config.yaml` (global).
- If missing, it **auto-creates** a sane default and tells you.

---

## Config

### Project config (recommended)

Place a single file at repo root:

`.agentlink.yaml`
```yaml
source: CLAUDE.md
links:
  - AGENTS.md
  - OPENCODE.md
```

Notes:
- **`source` must be a real file**, not a symlink (Agentlink warns if it is).
- Paths in `links` are relative to the project root.

### Global config

`~/.config/agentlink/config.yaml`
```yaml
source: ~/.config/claude/CLAUDE.md
links:
  - ~/.config/opencode/AGENTS.md
```

---

## Platform notes

- **macOS + Linux**: standard POSIX symlinks (`ln -s`) — works the same.
- **Git**: symlinks are stored as links (not file copies). That’s fine; teams who dislike that can add them to `.gitignore`.
- **Editors/IDEs**: most follow symlinks transparently.

---

## FAQ

**Why not templates or generators?**  
Because 90% of the time the files **should be identical**. When they’re not, this tool isn’t the right fit (or add a second source and stop linking that one).

**What if my source differs per project?**  
Perfect—put a `.agentlink.yaml` in each repo and choose the source you actually edit there.

**Can the source be `AGENTS.md` instead of `CLAUDE.md`?**  
Yes. The source is *whatever you want to edit*. The others link to it.

**What happens when a new AI tool comes out?**  
Just add its expected path to your config. If "SuperCoder AI" expects `.supercoder/prompts/main.md`, add that path and run `agentlink sync`. Directories are created automatically, symlink points to your source file. Zero code changes, zero updates needed.

**MCP / `.mcp.json`?**  
Out of scope. Formats differ between tools; symlinking a single JSON to multiple consumers usually doesn't make sense.

