# b+ Commands & Flags Reference

> **Best-in-class command and flag system combining the best from Claude Code, Gemini CLI, and Crush CLI**

---

## Table of Contents

1. [Command-Line Flags](#command-line-flags)
2. [Slash Commands (In-Session)](#slash-commands-in-session)
3. [Keyboard Shortcuts](#keyboard-shortcuts)
4. [Custom Commands](#custom-commands)
5. [Command Comparison Matrix](#command-comparison-matrix)

---

## Command-Line Flags

### **Core Flags**

#### `--help` / `-h`
Display comprehensive help information including all commands, flags, and usage examples.
```bash
b+ --help
b+ -h
```

#### `--version` / `-v`
Show the installed b+ version and build information.
```bash
b+ --version
b+ -v
```

#### `--debug`
Enable debug mode with verbose logging for troubleshooting.
```bash
b+ --debug
```

---

### **Execution Modes**

#### `--fast` (Default)
Run in Fast Mode - Layer 4 (Main Agent) only for quick tasks.
```bash
b+ --fast
```

#### `--thorough`
Run in Thorough Mode - All 7 layers active for complex, critical tasks.
```bash
b+ --thorough
```

#### `--mode <mode>`
Explicitly set the execution mode.
```bash
b+ --mode fast
b+ --mode thorough
```

---

### **Layer Control**

#### `--layers <layer-list>`
Enable specific layers only (comma-separated). Layer 4 and 6 cannot be disabled.
```bash
b+ --layers 1,2,3,4,5,6          # All layers except 7
b+ --layers 4,6                  # Fast mode equivalent
b+ --layers 1,4,5,6              # Skip parallel planning
```

#### `--disable-layer <layer>`
Disable specific layer(s). Can be specified multiple times.
```bash
b+ --disable-layer 1             # Skip intent clarification
b+ --disable-layer 2 --disable-layer 3  # Skip planning entirely
```

#### `--plans <number>`
Set number of parallel plans in Layer 2 (default: 4, range: 2-8).
```bash
b+ --plans 6                     # Generate 6 diverse plans
```

#### `--validation-iterations <number>`
Set max validation iterations for Layer 5 (default: 3, range: 1-5).
```bash
b+ --validation-iterations 5     # Allow up to 5 validation rounds
```

---

### **Model & Provider Selection**

#### `--model <provider/model-id>`
Set the default model for all layers.
```bash
b+ --model anthropic/claude-sonnet-4-5
b+ --model openai/gpt-4-turbo
b+ --model ollama/deepseek-coder:33b
```

#### `--layer<N>-model <provider/model-id>`
Set model for specific layer.
```bash
b+ --layer1-model openai/gpt-4-turbo
b+ --layer2-model anthropic/claude-opus-4-1
b+ --layer4-model ollama/codellama:34b
```

#### `--local-only`
Force all operations to use local models only (privacy mode).
```bash
b+ --local-only
```

#### `--cloud-only`
Force all operations to use cloud models only (maximum capability).
```bash
b+ --cloud-only
```

#### `--list-models`
List all available models for all configured providers.
```bash
b+ --list-models
b+ --list-models --provider ollama   # List only Ollama models
```

---

### **Session Management**

#### `--session <name>`
Load or create a named session.
```bash
b+ --session my-auth-refactor
b+ --session feature-xyz
```

#### `--resume` / `-r`
Resume the last active session.
```bash
b+ --resume
b+ -r
```

#### `--new-session` / `-n`
Force start a new session (don't resume).
```bash
b+ --new-session
b+ -n
```

---

### **Context & Files**

#### `--add-dir <path>`
Add additional directory to context (outside current working directory).
```bash
b+ --add-dir ../shared-lib
b+ --add-dir ~/Documents/specs
```

#### `--ignore <pattern>`
Add patterns to ignore (in addition to .gitignore).
```bash
b+ --ignore "*.log"
b+ --ignore "test/**"
```

#### `--context-size <tokens>`
Set maximum context window size (default: 200000).
```bash
b+ --context-size 100000         # Smaller context for faster responses
b+ --context-size 1000000        # Use 1M context for large codebases
```


---

### **Output & Formatting**

#### `--output-format <format>`
Set output format for non-interactive mode (json, markdown, text).
```bash
b+ --output-format json
b+ --output-format markdown
```

#### `--quiet` / `-q`
Suppress progress indicators and non-essential output.
```bash
b+ --quiet
b+ -q
```

#### `--verbose`
Show detailed output including tool calls and layer transitions.
```bash
b+ --verbose
```

#### `--no-color`
Disable color output (useful for logging/piping).
```bash
b+ --no-color
```

#### `--theme <theme>`
Set UI theme (dark, light, custom).
```bash
b+ --theme dark
b+ --theme solarized
```

---

### **Non-Interactive Mode**

#### `--prompt <text>` / `-p <text>`
Execute a single prompt and exit (non-interactive).
```bash
b+ -p "Fix all TypeScript errors in src/"
b+ --prompt "Generate unit tests for auth.go"
```

#### `--file <path>`
Read prompt from file.
```bash
b+ --file task.txt
```

#### `--pipe`
Read prompt from stdin (for piping).
```bash
echo "Refactor this function" | b+ --pipe
cat task.md | b+ --pipe --output-format json
```

---

### **Tool & Integration Control**

#### `--tools <tool-list>`
Enable only specific tools (comma-separated).
```bash
b+ --tools read,write,edit,grep
```

#### `--disable-tool <tool>`
Disable specific tool(s). Can be specified multiple times.
```bash
b+ --disable-tool bash
b+ --disable-tool web-fetch --disable-tool web-search
```

#### `--mcp-server <name>`
Enable specific MCP server(s).
```bash
b+ --mcp-server github
b+ --mcp-server slack,database
```

#### `--no-mcp`
Disable all MCP servers for this session.
```bash
b+ --no-mcp
```

#### `--no-lsp`
Disable LSP integration for this session.
```bash
b+ --no-lsp
```

---

### **Security & Permissions**

#### `--yolo`
Skip ALL permission prompts (use with extreme caution).
```bash
b+ --yolo
```

#### `--auto-approve <category>`
Auto-approve specific categories of operations.
```bash
b+ --auto-approve read            # Auto-approve file reads
b+ --auto-approve write           # Auto-approve file writes
b+ --auto-approve bash            # Auto-approve bash commands
```

#### `--require-approval <category>`
Require approval for specific categories (overrides config).
```bash
b+ --require-approval bash
b+ --require-approval web-fetch
```

#### `--dry-run`
Show what would be done without actually executing (validation mode).
```bash
b+ --dry-run
```

#### `--sandbox`
Run in sandboxed mode with restricted permissions.
```bash
b+ --sandbox
```

---

### **Checkpoint & Backup**

#### `--checkpointing` / `-c`
Enable automatic checkpointing before file modifications.
```bash
b+ --checkpointing
b+ -c
```

#### `--checkpoint-interval <seconds>`
Set checkpoint interval (default: 300 seconds).
```bash
b+ --checkpoint-interval 60      # Checkpoint every minute
```

#### `--no-backup`
Disable automatic file backups.
```bash
b+ --no-backup
```

---

### **Performance & Optimization**

#### `--max-parallel <number>`
Set maximum parallel operations (default: 4).
```bash
b+ --max-parallel 8              # Allow 8 parallel operations
```

#### `--cache`
Enable aggressive caching for faster responses.
```bash
b+ --cache
```

#### `--no-cache`
Disable all caching.
```bash
b+ --no-cache
```

#### `--timeout <seconds>`
Set timeout for operations (default: 300 seconds).
```bash
b+ --timeout 600                 # 10-minute timeout
```

---

### **Configuration**

#### `--config <path>`
Use custom configuration file.
```bash
b+ --config ./team-config.yaml
b+ --config ~/.b+/enterprise.yaml
```

#### `--profile <name>`
Load configuration profile.
```bash
b+ --profile work
b+ --profile personal
b+ --profile enterprise
```

#### `--save-config`
Save current flags as default configuration.
```bash
b+ --thorough --model anthropic/claude-opus-4-1 --save-config
```

---

### **Logging & Diagnostics**

#### `--log-file <path>`
Write logs to specified file.
```bash
b+ --log-file ./b+.log
```

#### `--log-level <level>`
Set logging level (debug, info, warn, error).
```bash
b+ --log-level debug
```

#### `--trace`
Enable tracing for performance analysis.
```bash
b+ --trace
```

#### `--metrics`
Enable metrics collection.
```bash
b+ --metrics
```

---

### **AI & Cost Management**

#### `--budget <amount>`
Set maximum cost budget for session (in USD).
```bash
b+ --budget 5.00                 # Max $5 per session
```

#### `--budget-alert <amount>`
Set cost alert threshold (in USD).
```bash
b+ --budget-alert 2.00           # Alert at $2
```

#### `--estimate`
Show cost estimate before execution.
```bash
b+ --estimate
```

#### `--free-only`
Use only free models (local or free tier cloud).
```bash
b+ --free-only
```

---

## Slash Commands (In-Session)

### **Core Commands**

#### `/help`
Display all available slash commands and keyboard shortcuts.
```
/help
/help commands                   # Show only commands
/help shortcuts                  # Show only keyboard shortcuts
```

#### `/clear`
Clear conversation history (does not affect session state).
```
/clear
/clear --confirm                 # Skip confirmation
```

#### `/reset`
Reset entire session (clear history + reset context).
```
/reset
```

#### `/exit` / `/quit`
Exit b+ session.
```
/exit
/quit
```

---

### **Session Management**

#### `/session`
Manage sessions.
```
/session list                    # List all sessions
/session save <name>             # Save current session
/session load <name>             # Load saved session
/session delete <name>           # Delete session
/session rename <old> <new>      # Rename session
/session export <name> <file>    # Export session to file
/session import <file>           # Import session from file
/session share <name>            # Generate shareable link
```

#### `/checkpoint`
Create manual checkpoint.
```
/checkpoint
/checkpoint save <name>          # Named checkpoint
/checkpoint list                 # List checkpoints
/checkpoint restore <name>       # Restore checkpoint
```

#### `/resume`
Resume last session or specific session.
```
/resume
/resume <name>
```

---

### **Mode & Layer Control**

#### `/mode`
Switch execution mode.
```
/mode fast                       # Switch to Fast Mode
/mode thorough                   # Switch to Thorough Mode
/mode status                     # Show current mode
```

#### `/layers`
Manage layer configuration.
```
/layers status                   # Show layer status
/layers enable <layer>           # Enable specific layer
/layers disable <layer>          # Disable specific layer
/layers reset                    # Reset to default
```

#### `/plans`
Configure parallel planning.
```
/plans 6                         # Set to 6 parallel plans
/plans show                      # Show last generated plans
/plans replay                    # Replay plan generation
```

#### `/validate`
Trigger manual validation of last operation.
```
/validate
/validate strict                 # Use strict validation
```

---

### **Model & Provider Management**

#### `/models`
Manage model selection.
```
/models list                     # List all available models
/models list ollama              # List models for specific provider
/models current                  # Show current model configuration
/models set <provider/model>     # Set default model
/models layer<N> <provider/model> # Set model for specific layer
/models test <provider/model>    # Test model connection
/models refresh                  # Refresh model list from providers
```

#### `/providers`
Manage provider configuration.
```
/providers list                  # List all providers
/providers status                # Show provider connection status
/providers test <provider>       # Test provider connection
/providers configure <provider>  # Configure provider (API keys, etc.)
```

---

### **Context & File Management**

#### `/context`
Manage context.
```
/context status                  # Show context statistics
/context optimize                # Force context optimization
/context clear                   # Clear non-essential context
/context export <file>           # Export context to file
/context health                  # Show context health metrics
```

#### `/files`
Manage file context.
```
/files list                      # List files in context
/files add <path>                # Add file to context
/files remove <path>             # Remove file from context
/files watch <path>              # Watch file for changes
/files unwatch <path>            # Stop watching file
/files reload                    # Reload all files
```

#### `/ignore`
Manage ignore patterns.
```
/ignore list                     # List ignore patterns
/ignore add <pattern>            # Add ignore pattern
/ignore remove <pattern>         # Remove ignore pattern
```

---

### **Tools & Integrations**

#### `/tools`
Manage tool configuration.
```
/tools list                      # List all available tools
/tools enable <tool>             # Enable tool
/tools disable <tool>            # Disable tool
/tools status                    # Show tool status
/tools test <tool>               # Test tool functionality
```

#### `/mcp`
Manage MCP servers.
```
/mcp list                        # List configured MCP servers
/mcp enable <server>             # Enable MCP server
/mcp disable <server>            # Disable MCP server
/mcp status                      # Show MCP server status
/mcp add <server>                # Add new MCP server
/mcp remove <server>             # Remove MCP server
/mcp test <server>               # Test MCP server connection
```

#### `/lsp`
Manage LSP integration.
```
/lsp status                      # Show LSP server status
/lsp restart <language>          # Restart LSP server
/lsp logs <language>             # Show LSP server logs
/lsp diagnostics                 # Show all diagnostics
```

---

### **Undo & History**

#### `/undo`
Undo last operation.
```
/undo
/undo 3                          # Undo last 3 operations
/undo all                        # Undo all operations
```

#### `/redo`
Redo last undone operation.
```
/redo
/redo 2                          # Redo last 2 operations
```

#### `/history`
View operation history.
```
/history                         # Show recent history
/history 20                      # Show last 20 operations
/history search <query>          # Search history
/history export <file>           # Export history
```

---

### **Testing & Validation**

#### `/test`
Run tests.
```
/test                            # Run all tests
/test <pattern>                  # Run tests matching pattern
/test generate                   # Generate tests for current context
/test watch                      # Watch mode for tests
```

#### `/lint`
Run linters.
```
/lint                            # Lint all files
/lint <path>                     # Lint specific file
/lint fix                        # Auto-fix linting issues
```

#### `/format`
Format code.
```
/format                          # Format all files
/format <path>                   # Format specific file
```

#### `/security`
Run security scans.
```
/security                        # Basic security scan
/security deep                   # Deep security analysis
/security report                 # Generate security report
```

---

### **Explanation & Debugging**

#### `/explain`
Explain last operation or specific task.
```
/explain                         # Explain last task
/explain <task-id>               # Explain specific task
/explain layers                  # Show layer breakdown
/explain cost                    # Show cost breakdown
/explain context                 # Explain context management
```

#### `/debug`
Debug information.
```
/debug                           # Show debug info
/debug layers                    # Debug layer execution
/debug tools                     # Debug tool calls
/debug context                   # Debug context management
```

#### `/logs`
View logs.
```
/logs                            # Show recent logs (last 100 lines)
/logs 500                        # Show last 500 lines
/logs --tail 100                 # Show last 100 lines
/logs --follow                   # Follow logs in real-time
/logs --level error              # Show only errors
/logs export <file>              # Export logs to file
```

---

### **Git Integration**

#### `/git`
Git operations.
```
/git status                      # Show git status
/git diff                        # Show diff
/git commit                      # Create smart commit
/git commit --message "msg"      # Commit with message
/git branch                      # List branches
/git checkout <branch>           # Checkout branch
/git pull                        # Pull changes
/git push                        # Push changes
```

#### `/pr`
Pull request operations.
```
/pr create                       # Create PR
/pr create --draft               # Create draft PR
/pr list                         # List open PRs
/pr view <number>                # View PR details
/pr review <number>              # Review PR
/pr merge <number>               # Merge PR
```

---

### **Cost & Performance**

#### `/cost`
Show cost information.
```
/cost                            # Show session cost
/cost history                    # Show cost history
/cost breakdown                  # Show detailed breakdown
/cost estimate <task>            # Estimate task cost
```

#### `/metrics`
Show performance metrics.
```
/metrics                         # Show session metrics
/metrics layers                  # Show layer performance
/metrics tools                   # Show tool performance
/metrics export <file>           # Export metrics
```

---

### **Settings & Configuration**

#### `/settings`
Open settings editor.
```
/settings                        # Open interactive settings
/settings get <key>              # Get setting value
/settings set <key> <value>      # Set setting value
/settings reset                  # Reset to defaults
/settings export <file>          # Export settings
/settings import <file>          # Import settings
```

#### `/config`
Configuration management.
```
/config show                     # Show current config
/config edit                     # Edit config in editor
/config reload                   # Reload config from file
/config validate                 # Validate config file
```

---

### **Documentation & Help**

#### `/docs`
Open documentation.
```
/docs                            # Open main docs
/docs commands                   # Command reference
/docs layers                     # Layer architecture docs
/docs tools                      # Tool documentation
/docs search <query>             # Search docs
```

#### `/examples`
Show examples.
```
/examples                        # List all examples
/examples <category>             # Show category examples
/examples search <query>         # Search examples
```

---

### **Custom Commands**

#### `/init`
Initialize project with b+ configuration.
```
/init                            # Interactive initialization
/init --template <name>          # Use template
/init --minimal                  # Minimal config
```

#### `/scaffold`
Generate boilerplate code.
```
/scaffold component <name>       # Generate component
/scaffold api <name>             # Generate API endpoint
/scaffold test <file>            # Generate test file
```

---

## Keyboard Shortcuts

### **Navigation & Focus**

| Shortcut | Action |
|----------|--------|
| `Ctrl+G` | Focus chat input |
| `Ctrl+F` | Focus file browser |
| `Ctrl+S` | Focus sessions panel |
| `Ctrl+T` | Focus tools panel |
| `Ctrl+L` | Focus layers panel |
| `Ctrl+H` | Toggle history panel |

### **Execution & Control**

| Shortcut | Action |
|----------|--------|
| `Ctrl+Enter` | Send message / Execute |
| `Shift+Enter` | New line in input |
| `Ctrl+C` | Cancel current operation |
| `Ctrl+D` | Exit b+ |
| `Ctrl+Z` | Undo last operation |
| `Ctrl+Y` | Redo operation |

### **Mode & Layer Control**

| Shortcut | Action |
|----------|--------|
| `Shift+Tab` (2x) | Toggle Plan Mode |
| `Ctrl+M` | Toggle Fast/Thorough Mode |
| `Ctrl+1` to `Ctrl+7` | Toggle Layer 1-7 |
| `Ctrl+P` | Show parallel plans |

### **View & UI**

| Shortcut | Action |
|----------|--------|
| `Ctrl+/` | Toggle command palette |
| `Ctrl+\` | Toggle sidebar |
| `Ctrl+B` | Toggle file browser |
| `Ctrl+K` | Clear screen |
| `Ctrl+R` | Reload UI |
| `Ctrl++` | Zoom in |
| `Ctrl+-` | Zoom out |
| `Ctrl+0` | Reset zoom |

### **File Operations**

| Shortcut | Action |
|----------|--------|
| `Ctrl+O` | Open file picker |
| `@` | Fuzzy file search |
| `Drag & Drop` | Attach file/image |
| `Ctrl+Shift+A` | Add directory to context |

### **Search & Help**

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+F` | Search across files |
| `Ctrl+Shift+H` | Search history |
| `?` | Show help overlay |
| `F1` | Open documentation |

---

## Custom Commands

### **User-Scoped Commands**

Stored in `~/.b+/commands/` and available across all projects.

**Example: `~/.b+/commands/review.toml`**
```toml
name = "review"
description = "Perform code review"
prompt = """
Please review the code changes in {{args}} and provide:
1. Code quality assessment
2. Potential bugs or issues
3. Suggestions for improvement
4. Security concerns
"""

[settings]
mode = "thorough"
layers = [1, 2, 3, 4, 5, 6]
validation_strict = true
```

**Usage:**
```
/review src/auth.go
```

---

### **Project-Scoped Commands**

Stored in `.b+/commands/` and available only within the project.

**Example: `.b+/commands/test/e2e.toml`**
```toml
name = "test:e2e"
description = "Run E2E tests and fix failures"
prompt = """
1. Run E2E test suite: npm run test:e2e
2. If any tests fail, analyze failures
3. Fix issues in the code
4. Re-run tests to verify
5. Report results
"""

[settings]
tools = ["bash", "read", "write", "edit", "test"]
auto_approve = ["read"]
```

**Usage:**
```
/test:e2e
```

---

### **MCP Prompt Commands**

MCP servers can expose prompts as slash commands automatically.

**Example: GitHub MCP Server**
```
/github:create-issue "Bug: Login fails on Safari"
/github:review-pr 123
```

---

### **Shell Command Integration**

Execute shell commands directly in prompts.

**Example: `.b+/commands/deploy.toml`**
```toml
name = "deploy"
description = "Deploy to staging"
prompt = """
1. Run tests: $(npm test)
2. Build project: $(npm run build)
3. Deploy: $(./deploy.sh staging)
4. Verify deployment
"""
```

---

### **Argument Handling**

Commands support positional and named arguments.

**Example: `.b+/commands/generate/api.toml`**
```toml
name = "generate:api"
description = "Generate API endpoint"
prompt = """
Generate a REST API endpoint for {{resource}} with the following methods: {{methods}}.

Include:
- Route definition
- Controller
- Service layer
- Unit tests
- API documentation

Follow project conventions in {{convention_file}}.
"""

[[args]]
name = "resource"
required = true
description = "Resource name (e.g., 'user', 'product')"

[[args]]
name = "methods"
default = "GET,POST,PUT,DELETE"
description = "HTTP methods to implement"

[[args]]
name = "convention_file"
default = "docs/API_CONVENTIONS.md"
description = "API conventions file"
```

**Usage:**
```
/generate:api user --methods="GET,POST"
/generate:api product --methods="GET,POST,PUT,DELETE" --convention_file="docs/api.md"
```

---

## Command Comparison Matrix

| Feature | Claude Code | Gemini CLI | Crush CLI | **b+** |
|---------|-------------|------------|-----------|--------|
| **Slash Commands** | ✅ Custom | ✅ Custom + Built-in | ❌ Limited | ✅ **Extensive + Custom** |
| **Mode Switching** | ⚠️ Plan Mode only | ❌ No | ❌ No | ✅ **Fast/Thorough + Layer Control** |
| **Model Selection** | ⚠️ CLI flag only | ⚠️ CLI flag only | ⚠️ Config only | ✅ **CLI + In-Session + Per-Layer** |
| **Session Management** | ❌ Limited | ✅ Checkpointing | ✅ Multi-session | ✅ **Advanced (save/load/share)** |
| **Undo/Redo** | ❌ No | ⚠️ Restore only | ✅ Yes | ✅ **Full History** |
| **Context Management** | ❌ No | ⚠️ Memory command | ❌ No | ✅ **Advanced Layer 6** |
| **Tool Control** | ❌ No | ⚠️ Limited | ❌ No | ✅ **Full Control** |
| **MCP Integration** | ✅ Yes | ✅ Yes | ✅ Yes | ✅ **Enhanced** |
| **Custom Commands** | ✅ Markdown files | ✅ TOML files | ❌ No | ✅ **TOML + MCP + Shell** |
| **Cost Management** | ❌ No | ❌ No | ❌ No | ✅ **Budget Alerts** |
| **Layer Visibility** | ❌ N/A | ❌ N/A | ❌ N/A | ✅ **/explain /layers** |
| **Git Integration** | ⚠️ Via tools | ❌ Via MCP | ⚠️ Via tools | ✅ **Native /git /pr** |
| **Testing Integration** | ⚠️ Via tools | ❌ No | ❌ No | ✅ **Native /test** |
| **Non-Interactive Mode** | ✅ Yes | ❌ No | ✅ Yes | ✅ **Enhanced** |
| **Keyboard Shortcuts** | ⚠️ Basic | ❌ No | ✅ Extensive | ✅ **Comprehensive** |
| **Permission Control** | ⚠️ Basic | ❌ No | ⚠️ --yolo only | ✅ **Granular + Categories** |

---

## Best Practices

### **Efficient Command Usage**

1. **Use Fast Mode for simple tasks**
   ```bash
   b+ --fast -p "Fix typo in README"
   ```

2. **Use Thorough Mode for complex features**
   ```bash
   b+ --thorough -p "Implement OAuth 2.0 authentication"
   ```

3. **Save common configurations as profiles**
   ```bash
   b+ --thorough --model anthropic/claude-opus-4-1 --profile work --save-config
   b+ --profile work  # Use later
   ```

4. **Use custom commands for repetitive tasks**
   ```bash
   # Instead of typing full prompts repeatedly
   /deploy staging
   /test:e2e
   /review:security
   ```

5. **Leverage keyboard shortcuts**
   ```
   Shift+Tab x2  → Plan mode
   Ctrl+M        → Toggle mode
   @filename     → Quick file attach
   ```

---

## Flag Priority

When flags conflict, priority is:
1. Command-line flags (highest priority)
2. Environment variables
3. Profile configuration
4. Project configuration (`.b+/config.yaml`)
5. User configuration (`~/.b+/config.yaml`)
6. Default values (lowest priority)

**Example:**
```bash
# Profile says --fast, but CLI overrides
b+ --profile work --thorough  # Uses thorough mode
```

---

**Last Updated:** 2025-10-25
**Version:** 2.0

