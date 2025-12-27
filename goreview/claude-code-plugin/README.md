# GoReview Plugin for Claude Code

AI-powered code review, commit generation, and code quality tools integrated with Claude Code.

## Features

- **Slash Commands**: Quick access to GoReview functions
- **Specialized Agents**: Security, performance, and testing experts
- **Background Watcher**: Continuous code quality monitoring
- **Auto Hooks**: Automatic review after edits
- **MCP Server**: Full GoReview functionality as tools

## Installation

### From Marketplace

```bash
# Add the GoReview marketplace
/plugin marketplace add JNZader/goreview

# Install the plugin
/plugin install goreview
```

### Manual Installation

```bash
# Clone the plugin
git clone https://github.com/JNZader/goreview.git
cd goreview/claude-code-plugin

# Install locally
/plugin install --local .
```

## Requirements

- Claude Code v2.0.12 or higher
- GoReview CLI installed and in PATH

```bash
# Install GoReview
go install github.com/JNZader/goreview/goreview/cmd/goreview@latest
```

## Usage

### Slash Commands

| Command | Description |
|---------|-------------|
| `/review` | Review staged code changes |
| `/commit-ai` | Generate AI commit message |
| `/fix-issues` | Auto-fix found issues |
| `/changelog` | Generate changelog |
| `/stats` | Show review statistics |
| `/security-scan` | Deep security analysis |

### Subagents

Use specialized agents for focused analysis:

```
Use the security-reviewer to audit this code
Use the perf-reviewer to find bottlenecks
Use the test-reviewer to check test coverage
```

### MCP Tools

Once installed, Claude Code can use GoReview tools directly:

- `goreview_review` - Analyze code changes
- `goreview_commit` - Generate commit messages
- `goreview_fix` - Auto-fix issues
- `goreview_search` - Search review history
- `goreview_stats` - Get statistics
- `goreview_changelog` - Generate changelogs
- `goreview_doc` - Generate documentation

## Configuration

Edit your Claude Code settings to customize:

```json
{
  "goreview.autoReview": true,
  "goreview.mode": "standard",
  "goreview.personality": "senior"
}
```

## Hooks

The plugin includes automatic hooks:

- **PostToolUse (Edit/Write)**: Quick review after code changes
- **SessionStart**: Show project health on startup
- **Stop**: Save review state before session ends
- **PreCompact**: Checkpoint synchronization

## Files

```
claude-code-plugin/
├── .claude-plugin/
│   └── plugin.json          # Plugin metadata
├── commands/                 # Slash commands
│   ├── review.md
│   ├── commit-ai.md
│   ├── fix-issues.md
│   ├── changelog.md
│   ├── stats.md
│   └── security-scan.md
├── agents/                   # Specialized subagents
│   ├── security-reviewer/
│   ├── perf-reviewer/
│   ├── test-reviewer/
│   ├── fix-agent/
│   └── goreview-watcher/
├── skills/                   # Auto-invoked skills
│   ├── goreview-workflow/
│   └── commit-standards/
├── hooks/                    # Automation hooks
│   ├── hooks.json
│   └── checkpoint-sync.json
└── scripts/                  # Utility scripts
    ├── quick-review.sh
    ├── security-scan.sh
    └── pre-commit-check.sh
```

## License

MIT License - See LICENSE for details.
