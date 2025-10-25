# b+ Implementation Plan

> **Phase-by-phase development roadmap**
>
> Each phase builds upon previously completed work. Phases are designed to deliver incrementally functional components that can be tested and validated before moving forward.

---

## Table of Contents

1. [Development Principles](#development-principles)
2. [Phase 1: Foundation & Project Setup](#phase-1-foundation--project-setup)
3. [Phase 2: Core Infrastructure](#phase-2-core-infrastructure)
4. [Phase 3: Terminal UI Foundation](#phase-3-terminal-ui-foundation)
5. [Phase 4: Provider System](#phase-4-provider-system)
6. [Phase 5: Tool System Foundation](#phase-5-tool-system-foundation)
7. [Phase 6: Layer 4 - Main Agent (Fast Mode MVP)](#phase-6-layer-4---main-agent-fast-mode-mvp)
8. [Phase 7: Layer 6 - Context Management](#phase-7-layer-6---context-management)
9. [Phase 8: LSP Integration](#phase-8-lsp-integration)
10. [Phase 9: Layer 1 - Intent Clarification](#phase-9-layer-1---intent-clarification)
11. [Phase 10: Layer 2 - Parallel Planning](#phase-10-layer-2---parallel-planning)
12. [Phase 11: Layer 3 - Plan Synthesis](#phase-11-layer-3---plan-synthesis)
13. [Phase 12: Layer 5 - Validation](#phase-12-layer-5---validation)
14. [Phase 13: Advanced Tool System](#phase-13-advanced-tool-system)
15. [Phase 13.5: Community Plugin System for Tools](#phase-135-community-plugin-system-for-tools)
16. [Phase 14: MCP Integration](#phase-14-mcp-integration)
17. [Phase 15: Advanced UI Features](#phase-15-advanced-ui-features)
18. [Phase 16: Command System](#phase-16-command-system)
19. [Phase 17: Advanced Session Management](#phase-17-advanced-session-management)
20. [Phase 18: Cost & Budget Management](#phase-18-cost--budget-management)
21. [Phase 19: Security & Permissions](#phase-19-security--permissions)
22. [Phase 20: Testing Infrastructure](#phase-20-testing-infrastructure)
23. [Phase 21: Documentation & Examples](#phase-21-documentation--examples)
24. [Phase 22: Distribution & Release](#phase-22-distribution--release)

---

## Development Principles

### **Incremental Delivery**
Each phase delivers a working, testable component. No phase should leave the system in a broken state.

### **Interface-First Design**
Define interfaces before implementations. This allows parallel development and easy swapping of implementations.

### **Test as You Build**
Write tests alongside implementation. Don't defer testing to later phases.

### **Configuration Over Code**
Make behaviors configurable. Use YAML/TOML for settings, not hardcoded values.

### **Fail Fast, Fail Clearly**
Provide clear error messages. Validate inputs early. Don't let errors propagate silently.

### **Performance from Day One**
Profile early and often. Optimize hot paths. Use goroutines for parallelism.

---

## Phase 1: Foundation & Project Setup

**Goal:** Establish project structure, development environment, and basic tooling.

### **1.1 Project Initialization**
- [ ] Initialize Go module: `go mod init github.com/abrksh22/bplus`
- [ ] Create directory structure matching [VISION.md](docs/VISION.md) architecture
- [ ] Set up `.gitignore` for Go projects
- [ ] Create `README.md` with project overview
- [ ] Set up `LICENSE` file (MIT for open core)

### **1.2 Development Environment**
- [ ] Configure `go.mod` with Go 1.21+
- [ ] Set up Makefile for common tasks (build, test, lint, run)
- [ ] Configure golangci-lint with comprehensive linters
- [ ] Set up pre-commit hooks (gofmt, golangci-lint)
- [ ] Configure VSCode/GoLand settings (optional)

### **1.3 CI/CD Foundation**
- [ ] Create GitHub Actions workflow for tests
- [ ] Configure lint checks on PRs
- [ ] Set up build matrix (macOS, Linux, Windows)
- [ ] Configure code coverage reporting

### **1.4 Dependency Management**
- [ ] Lock dependency versions in `go.mod`
- [ ] Document all major dependencies and rationale
- [ ] Set up Dependabot for security updates

**Deliverable:** Project compiles with `go build ./...` and basic structure is in place.

---

## Phase 2: Core Infrastructure

**Goal:** Build foundational utilities used throughout the application.

### **2.1 Configuration System**
- [ ] Create `config/` package
- [ ] Implement configuration struct with all settings from [VISION.md](docs/VISION.md)
- [ ] Use Viper for YAML/TOML/JSON support
- [ ] Support environment variables with `${VAR}` substitution
- [ ] Implement XDG Base Directory support (`~/.config/bplus/`, `~/.local/share/bplus/`)
- [ ] Implement config file discovery (user config → project config → defaults)
- [ ] Create default configuration template
- [ ] Implement configuration validation
- [ ] Add config merging logic (CLI flags > env vars > profile > project config > user config > defaults)

**Config File Structure:**
```go
type Config struct {
    Mode         string           // "fast" or "thorough"
    Models       ModelConfig      // Default and per-layer models
    Providers    ProviderConfig   // API keys, base URLs
    Layers       LayerConfig      // Enable/disable, models, settings
    Tools        ToolConfig       // Enabled tools, permissions
    UI           UIConfig         // Theme, shortcuts, layout
    Session      SessionConfig    // Persistence, checkpoints
    Security     SecurityConfig   // Permissions, sandboxing
    Cost         CostConfig       // Budget, alerts
}
```

### **2.2 Logging System**
- [ ] Create `logging/` package using zerolog
- [ ] Implement structured logging with context
- [ ] Support log levels (debug, info, warn, error)
- [ ] Implement log file rotation
- [ ] Add JSON output for machine parsing
- [ ] Create logger middleware for function tracing

### **2.3 Database Layer**
- [ ] Create `storage/` package
- [ ] Implement SQLite wrapper with schema versioning
- [ ] Create database schemas:
  - Sessions table (id, name, created_at, updated_at, context_snapshot)
  - Messages table (id, session_id, role, content, timestamp, tokens, cost)
  - Files table (id, session_id, path, content_hash, modified_at)
  - Checkpoints table (id, session_id, name, state_snapshot, created_at)
  - Operations table (id, session_id, type, details, timestamp)
  - Metrics table (id, session_id, metric_type, value, timestamp)
- [ ] Implement SQLite FTS5 for full-text search
- [ ] Implement bbolt key-value store wrapper
- [ ] Create database migration system
- [ ] Add database backup and restore utilities

### **2.4 Error Handling**
- [ ] Create `errors/` package with custom error types
- [ ] Implement error wrapping with context
- [ ] Create error codes for categorization
- [ ] Implement user-friendly error messages
- [ ] Add error reporting utilities

### **2.5 Utilities**
- [ ] File system utilities (safe write, atomic replace, backup)
- [ ] String utilities (truncate, format, sanitize)
- [ ] Time utilities (duration formatting, timeout context)
- [ ] Crypto utilities (hash, encrypt/decrypt with age)
- [ ] Network utilities (retry with backoff, timeout)

**Deliverable:** Core infrastructure packages that can be imported and used. All functions have tests.

---

## Phase 3: Terminal UI Foundation

**Goal:** Build the Bubble Tea-based terminal interface foundation.

### **3.1 Bubble Tea Setup**
- [ ] Create `ui/` package
- [ ] Implement main application model (implements tea.Model)
- [ ] Set up message passing system
- [ ] Implement update loop
- [ ] Implement view rendering
- [ ] Add window size handling

### **3.2 Core UI Components**
- [ ] Create `ui/components/` package
- [ ] **Input Component**: Text input with history, autocomplete
- [ ] **Output Component**: Scrollable message display with syntax highlighting
- [ ] **Status Bar**: Show current mode, model, cost, token count
- [ ] **Spinner**: Loading indicator for async operations
- [ ] **Progress Bar**: Show progress for long operations
- [ ] **Modal Dialog**: For confirmations and prompts
- [ ] **List Selector**: Dropdown/menu for selections
- [ ] **Split Pane**: Horizontal/vertical panes with resizing

### **3.3 Layout System**
- [ ] Implement flexible layout engine using Lip Gloss
- [ ] Create layout presets (chat, side-by-side, full-screen)
- [ ] Implement pane resizing with keyboard shortcuts
- [ ] Add layout persistence (save/restore)

### **3.4 Theme System**
- [ ] Create `ui/themes/` package
- [ ] Implement theme struct with all colors and styles
- [ ] Create built-in themes:
  - Dark (default)
  - Light
  - Solarized Dark
  - Solarized Light
  - Nord
  - Dracula
- [ ] Add theme switching at runtime
- [ ] Support custom user themes via config

### **3.5 Keyboard Shortcuts**
- [ ] Create keyboard handler
- [ ] Implement key bindings from [COMMANDS_FLAGS.md](docs/COMMANDS_FLAGS.md):
  - `Ctrl+G`: Focus chat
  - `Ctrl+C`: Cancel operation
  - `Ctrl+D`: Exit
  - `Ctrl+K`: Clear screen
  - `Ctrl+/`: Command palette
  - `Shift+Tab` (2x): Toggle plan mode
  - etc.
- [ ] Add configurable key bindings
- [ ] Implement key binding help overlay

### **3.6 Basic UI Flow**
- [ ] Implement startup screen with b+ logo
- [ ] Create main chat interface
- [ ] Add message streaming (show tokens as they arrive)
- [ ] Implement basic error display
- [ ] Add help overlay (`?` key)

**Deliverable:** A working terminal UI that can display messages, accept input, show status, and respond to keyboard shortcuts. No AI integration yet—just the UI shell.

---

## Phase 4: Provider System

**Goal:** Implement model provider abstraction and integrations.

### **4.1 Provider Interface**
- [ ] Create `models/` package
- [ ] Define `Provider` interface:
```go
type Provider interface {
    Name() string
    ListModels(ctx context.Context) ([]Model, error)
    CreateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, req CompletionRequest) (<-chan Token, error)
    TestConnection(ctx context.Context) error
    GetModelInfo(modelID string) (*ModelInfo, error)
}
```
- [ ] Define `Model` struct (ID, Name, ContextWindow, Pricing, etc.)
- [ ] Define `CompletionRequest` and `CompletionResponse` structs
- [ ] Implement provider registry

### **4.2 API Provider Implementations**

#### **4.2.1 Anthropic Provider**
- [ ] Create `models/providers/anthropic/` package
- [ ] Implement Anthropic API client
- [ ] Support streaming completions
- [ ] Handle rate limiting and retries
- [ ] Map Anthropic models to b+ model format
- [ ] Implement error handling for Anthropic-specific errors
- [ ] Add comprehensive tests with mock server

#### **4.2.2 OpenAI Provider**
- [ ] Create `models/providers/openai/` package
- [ ] Implement OpenAI API client
- [ ] Support both chat and completion endpoints
- [ ] Support streaming
- [ ] Handle function calling format (for tool use)
- [ ] Map OpenAI models to b+ format
- [ ] Add tests

#### **4.2.3 Gemini Provider**
- [ ] Create `models/providers/gemini/` package
- [ ] Implement Google AI API client
- [ ] Handle Gemini-specific request format
- [ ] Support streaming
- [ ] Map Gemini models to b+ format
- [ ] Add tests

#### **4.2.4 Groq Provider**
- [ ] Create `models/providers/groq/` package
- [ ] Implement Groq API client (OpenAI-compatible)
- [ ] Focus on speed optimizations
- [ ] Add tests

#### **4.2.5 OpenRouter Provider**
- [ ] Create `models/providers/openrouter/` package
- [ ] Implement OpenRouter API client
- [ ] Handle model routing to 100+ models
- [ ] Fetch available models dynamically
- [ ] Add tests

### **4.3 Local Provider Implementations**

#### **4.3.1 Ollama Provider**
- [ ] Create `models/providers/ollama/` package
- [ ] Implement Ollama API client
- [ ] Support model listing via `/api/tags`
- [ ] Support model pulling
- [ ] Handle streaming responses
- [ ] Implement health checks
- [ ] Add connection retry logic
- [ ] Add tests

#### **4.3.2 LM Studio Provider**
- [ ] Create `models/providers/lmstudio/` package
- [ ] Implement LM Studio OpenAI-compatible client
- [ ] Support dynamic model discovery
- [ ] Add tests

#### **4.3.3 llama.cpp Provider (Optional)**
- [ ] Create `models/providers/llamacpp/` package
- [ ] Implement direct GGUF loading via CGo
- [ ] Support for air-gapped deployments
- [ ] Add tests

### **4.4 Model Router**
- [ ] Create `models/router/` package
- [ ] Implement intelligent routing based on rules
- [ ] Support routing conditions:
  - Task complexity estimation
  - File count
  - Privacy level
  - Cost constraints
  - Model availability
- [ ] Implement fallback chains
- [ ] Add routing decision logging

### **4.5 Provider Configuration**
- [ ] Implement provider configuration loading
- [ ] Support API key management:
  - Environment variables
  - Config file (encrypted at rest)
  - Interactive prompt
- [ ] Implement base URL customization
- [ ] Add timeout and retry settings
- [ ] Implement provider health monitoring

### **4.6 Model Selection System**
- [ ] Implement model naming: `provider/model-id`
- [ ] Create model parser and validator
- [ ] Implement per-layer model configuration
- [ ] Add default model fallbacks
- [ ] Implement model capability checking

**Deliverable:** A complete provider system where you can configure API keys, list models, and send test requests to any supported provider. Includes comprehensive tests for all providers.

---

## Phase 5: Tool System Foundation

**Goal:** Build the tool execution framework and core file operation tools with a pluggable architecture designed for future community contributions.

### **5.1 Tool Interface (Design for Pluggability)**
- [ ] Create `tools/` package
- [ ] Define `Tool` interface with extensibility in mind:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() []Parameter
    RequiresPermission() bool
    Execute(ctx context.Context, params map[string]interface{}) (*Result, error)

    // Future plugin support hooks
    Category() string          // "file", "git", "web", "custom", etc.
    Version() string           // Semantic versioning
    IsExternal() bool          // true if loaded from plugin
}
```
- [ ] Define `Parameter` struct (name, type, required, description, validation)
- [ ] Define `Result` struct (success, output, error, metadata)
- [ ] Implement tool registry with plugin support hooks:
  - Registry should support dynamic tool registration (for future plugins)
  - Namespace tools to prevent conflicts (e.g., "core.read", "plugin.custom_tool")
  - Support tool versioning and compatibility checks

### **5.2 Permission System**
- [ ] Create `security/permissions/` package
- [ ] Define permission categories (read, write, execute, network)
- [ ] Implement permission checker
- [ ] Create permission prompt UI component
- [ ] Support `--yolo` and `--auto-approve` flags
- [ ] Implement permission audit logging

### **5.3 Core File Tools**

#### **5.3.1 Read Tool**
- [ ] Create `tools/file/read.go`
- [ ] Support reading files with line offset and limit
- [ ] Support multiple file formats:
  - Text files with line numbers
  - Images (display as base64 or description)
  - PDFs (text extraction)
  - Jupyter notebooks (.ipynb)
- [ ] Handle large files (streaming, truncation)
- [ ] Add caching for recently read files
- [ ] Implement tests

#### **5.3.2 Write Tool**
- [ ] Create `tools/file/write.go`
- [ ] Implement safe file writing (atomic writes)
- [ ] Create automatic backups before overwrite
- [ ] Support creating parent directories
- [ ] Validate file paths (prevent path traversal)
- [ ] Add tests

#### **5.3.3 Edit Tool**
- [ ] Create `tools/file/edit.go`
- [ ] Implement exact string replacement
- [ ] Support multiple edit strategies:
  - Exact match (default)
  - Regex replacement
  - Line-based edits
  - Diff-based edits
- [ ] Implement `replace_all` flag
- [ ] Show diffs before applying
- [ ] Add undo capability
- [ ] Add tests

#### **5.3.4 Glob Tool**
- [ ] Create `tools/file/glob.go`
- [ ] Implement glob pattern matching
- [ ] Support multiple patterns
- [ ] Respect `.gitignore` and `.bplusignore`
- [ ] Sort results by modification time
- [ ] Implement caching
- [ ] Add tests

#### **5.3.5 Grep Tool**
- [ ] Create `tools/file/grep.go`
- [ ] Use ripgrep (rg) as backend
- [ ] Support regex patterns
- [ ] Support context lines (-A, -B, -C)
- [ ] Support output modes (content, files_with_matches, count)
- [ ] Support file type filtering
- [ ] Implement multiline search
- [ ] Add tests

### **5.4 Execution Tools**

#### **5.4.1 Bash Tool**
- [ ] Create `tools/exec/bash.go`
- [ ] Implement command execution with timeout
- [ ] Support background execution
- [ ] Capture stdout and stderr separately
- [ ] Implement streaming output
- [ ] Support shell selection (bash, zsh, fish, pwsh)
- [ ] Add safety checks (dangerous commands)
- [ ] Implement tests

#### **5.4.2 Process Management**
- [ ] Create `tools/exec/process.go`
- [ ] Implement background process tracking
- [ ] Support process listing (BashOutput equivalent)
- [ ] Support process killing (KillShell equivalent)
- [ ] Add output filtering with regex
- [ ] Implement tests

### **5.5 Tool Execution Context**
- [ ] Create execution context with:
  - Working directory
  - Environment variables
  - Timeout settings
  - Permission grants
  - Audit trail
- [ ] Implement context propagation
- [ ] Add execution history

**Deliverable:** Working tool system with file and execution tools. Can execute tools from code with proper permissions and error handling.

---

## Phase 6: Layer 4 - Main Agent (Fast Mode MVP)

**Goal:** Build the core agentic loop that executes tasks using tools. This is Fast Mode—the minimal viable b+ that works end-to-end.

### **6.1 Agent Core**
- [ ] Create `layers/execution/` package
- [ ] Implement agent loop:
  1. Receive user message
  2. Send to LLM with system prompt and available tools
  3. Parse LLM response (text or tool calls)
  4. Execute tool calls with permission checks
  5. Send tool results back to LLM
  6. Repeat until task complete
- [ ] Implement streaming output to terminal
- [ ] Handle errors and retries
- [ ] Implement cancellation (Ctrl+C)

### **6.2 System Prompts**
- [ ] Create `prompts/` package
- [ ] Design Layer 4 (Main Agent) system prompt:
  - Role definition
  - Available tools and usage instructions
  - Output formatting
  - Error handling guidance
  - Best practices
- [ ] Support prompt templates with variables
- [ ] Implement prompt versioning

### **6.3 Tool Calling Format**
- [ ] Implement function calling format compatible with multiple providers:
  - OpenAI function calling format
  - Anthropic tool use format
  - Gemini function calling format
  - Fallback XML-based format
- [ ] Parse tool calls from LLM responses
- [ ] Format tool results for LLM
- [ ] Handle malformed tool calls gracefully

### **6.4 Token and Cost Tracking**
- [ ] Track tokens used per message
- [ ] Calculate cost based on provider pricing
- [ ] Display running cost in status bar
- [ ] Log costs to database
- [ ] Implement cost warnings

### **6.5 Error Recovery**
- [ ] Implement automatic retry for transient errors (rate limits, timeouts)
- [ ] Handle tool execution errors gracefully
- [ ] Provide error context to LLM for recovery
- [ ] Implement circuit breakers for failing tools
- [ ] Log all errors for debugging

### **6.6 Session Basics**
- [ ] Save messages to database
- [ ] Load conversation history on resume
- [ ] Implement basic session list command
- [ ] Add session deletion

### **6.7 CLI Integration**
- [ ] Implement main CLI entry point
- [ ] Parse command-line flags (--model, --fast, etc.)
- [ ] Initialize agent with configuration
- [ ] Start UI and agent loop
- [ ] Handle graceful shutdown

**Deliverable:** A working b+ that can execute tasks in Fast Mode. User types a message, agent uses tools, completes task, shows output. This is the first end-to-end milestone.

**Test Scenario:**
```bash
$ b+ --model anthropic/claude-sonnet-4-5

Welcome to b+ (Be Positive) v0.1.0

> Create a new file called hello.go with a simple "Hello, World!" program

[Agent uses Write tool to create hello.go]
[Shows diff of new file]

✓ Created hello.go with Hello World program

> Run the program

[Agent uses Bash tool to run: go run hello.go]

Output:
Hello, World!

✓ Program executed successfully
```

---

## Phase 7: Layer 6 - Context Management

**Goal:** Implement intelligent context optimization and session persistence.

### **7.1 Context Optimizer**
- [ ] Create `layers/context/` package
- [ ] Implement context optimization strategies:
  - **Summarization**: Compress old messages while preserving key info
  - **Semantic chunking**: Keep related information together
  - **Selective pruning**: Remove verbose tool outputs, keep summaries
- [ ] Implement tiered storage:
  - **Hot context**: In-memory, immediately available
  - **Warm context**: Database, fast retrieval
  - **Cold context**: Database, loaded only when needed
- [ ] Calculate context size in tokens
- [ ] Implement context health metrics

### **7.2 Summarization**
- [ ] Implement message summarization using LLM
- [ ] Summarize tool outputs (keep essential info only)
- [ ] Compress conversation history
- [ ] Preserve critical information:
  - User's original intent
  - File modifications
  - Validation results
  - Architecture decisions

### **7.3 Session Management**
- [ ] Implement session save (manual and automatic)
- [ ] Implement session load/resume
- [ ] Create session snapshots with:
  - Full conversation history
  - File states
  - Context state
  - Active configuration
- [ ] Implement session export (to file)
- [ ] Implement session import (from file)
- [ ] Add session metadata (name, description, tags)

### **7.4 Checkpoint System**
- [ ] Implement checkpoint creation
- [ ] Store checkpoint state:
  - Message history up to checkpoint
  - File states
  - Context snapshot
- [ ] Implement checkpoint restore
- [ ] Add automatic checkpointing (before destructive operations)
- [ ] Implement checkpoint cleanup (old checkpoints)

### **7.5 Context Health Monitoring**
- [ ] Calculate context metrics:
  - Current size (tokens)
  - Efficiency (optimization %)
  - Accuracy (validated against actual state)
  - Staleness (outdated files)
- [ ] Display context status in UI
- [ ] Warn when context is getting large
- [ ] Trigger automatic optimization at thresholds

**Deliverable:** Context management system that keeps sessions running smoothly even with long conversations. Sessions can be saved, loaded, and resumed.

---

## Phase 8: LSP Integration

**Goal:** Integrate Language Server Protocol for code intelligence.

### **8.1 LSP Client**
- [ ] Create `lsp/` package
- [ ] Implement LSP client using go-lsp library
- [ ] Support LSP initialization and handshake
- [ ] Implement LSP message handling (JSON-RPC)
- [ ] Add connection management (stdio process spawning)

### **8.2 Language Server Management**
- [ ] Create language server registry
- [ ] Implement auto-detection for common languages:
  - **Go**: gopls
  - **TypeScript**: typescript-language-server
  - **Python**: pyright or pylsp
  - **Rust**: rust-analyzer
  - **Java**: jdtls
  - etc.
- [ ] Implement server lifecycle:
  - Start server when needed
  - Keep alive during session
  - Shutdown gracefully
  - Restart on crashes
- [ ] Support custom server configuration

### **8.3 LSP Features**

#### **8.3.1 Diagnostics**
- [ ] Implement diagnostic collection (errors, warnings, hints)
- [ ] Display diagnostics in UI
- [ ] Provide diagnostics to agent for fixing issues

#### **8.3.2 Code Intelligence**
- [ ] Implement go-to-definition
- [ ] Implement find-references
- [ ] Implement symbol search
- [ ] Provide semantic information to agent

#### **8.3.3 Code Actions**
- [ ] Implement quick fixes
- [ ] Implement refactoring support
- [ ] Expose to agent as tools

### **8.4 LSP Tools**
- [ ] Create LSP-specific tools:
  - **LSPDiagnostics**: Get diagnostics for files
  - **LSPSymbols**: Search symbols in codebase
  - **LSPDefinition**: Get definition location
  - **LSPReferences**: Find references
- [ ] Integrate with existing tool system
- [ ] Add permission checks

**Deliverable:** LSP integration that provides real-time diagnostics and code intelligence to the agent. Agent can use LSP to understand code better.

---

## Phase 9: Layer 1 - Intent Clarification

**Goal:** Implement conversational intent clarification before task execution.

### **9.1 Intent Layer Core**
- [ ] Create `layers/intent/` package
- [ ] Implement intent clarification loop:
  1. Receive user's initial message
  2. Analyze for ambiguity or missing requirements
  3. Ask clarifying questions
  4. Collect user responses
  5. Repeat until intent is clear
  6. Finalize intent message
- [ ] Design intent clarification prompt
- [ ] Support multi-turn conversation
- [ ] Implement session-based lifecycle (closes after forwarding)

### **9.2 Clarification Strategy**
- [ ] Implement ambiguity detection:
  - Vague requirements
  - Missing constraints
  - Undefined scope
  - Unclear goals
- [ ] Generate targeted clarifying questions:
  - Multiple choice when possible
  - Open-ended when necessary
- [ ] Support user responses in various formats
- [ ] Detect when intent is sufficiently clear

### **9.3 Intent Finalization**
- [ ] Confirm intent with user before forwarding
- [ ] Show formatted intent summary
- [ ] Allow user to modify before proceeding
- [ ] Support `/done` command to finalize
- [ ] Format finalized intent for downstream layers

### **9.4 UI Integration**
- [ ] Indicate when in intent clarification mode
- [ ] Show clarification questions clearly
- [ ] Display intent summary before finalization
- [ ] Allow skipping intent clarification (--skip-intent flag)

**Deliverable:** Intent clarification layer that engages with user to understand requirements before starting work.

---

## Phase 10: Layer 2 - Parallel Planning

**Goal:** Implement parallel plan generation with diverse strategies.

### **10.1 Planning Layer Core**
- [ ] Create `layers/planning/` package
- [ ] Implement parallel session spawning:
  - Spawn 4 goroutines (configurable 2-8)
  - Each session independent
  - Each session uses different LLM/prompt
  - All sessions read-only (Glob, Grep, Read only)
- [ ] Collect all plans when complete
- [ ] Handle timeouts and failures gracefully

### **10.2 Planning Prompts**
- [ ] Design 4 diverse planning prompts:
  - **Plan A**: Speed and simplicity focused
  - **Plan B**: Robustness and error handling focused
  - **Plan C**: Maintainability and extensibility focused
  - **Plan D**: Performance and optimization focused
- [ ] Each prompt emphasizes different priorities
- [ ] All prompts request same output format

### **10.3 Plan Format**
- [ ] Define plan structure:
```go
type Plan struct {
    ID               string          // "A", "B", "C", "D"
    ModelUsed        string          // e.g., "anthropic/claude-opus-4-1"
    Strategy         string          // e.g., "speed_and_simplicity"
    Approach         string          // High-level description
    Steps            []string        // Ordered steps
    EstimatedTime    string          // e.g., "45 minutes"
    EstimatedComplexity string       // "low", "medium", "high"
    Risks            []string        // Potential issues
    Benefits         []string        // Advantages
    FilesToModify    []string        // File paths
}
```
- [ ] Implement plan parsing from LLM JSON output
- [ ] Validate plan structure
- [ ] Handle malformed plans

### **10.4 Model Selection for Planning**
- [ ] Support per-session model configuration
- [ ] Allow mixing cloud and local models
- [ ] Implement fallback if a model fails
- [ ] Track which model generated which plan

### **10.5 UI Integration**
- [ ] Show "Generating execution plans..." with spinner
- [ ] Display progress (e.g., "3/4 plans complete")
- [ ] Show summary of generated plans
- [ ] Allow user to view full plans if desired

**Deliverable:** Parallel planning system that generates 4 diverse plans simultaneously.

---

## Phase 11: Layer 3 - Plan Synthesis

**Goal:** Combine multiple plans into one optimized strategy.

### **11.1 Synthesis Layer Core**
- [ ] Create `layers/synthesis/` package
- [ ] Implement plan synthesis algorithm:
  1. Receive all 4 plans from Layer 2
  2. Analyze plans for common patterns
  3. Identify unique benefits from each plan
  4. Assess risks across all plans
  5. Merge steps into optimized sequence
  6. Check for completeness
  7. Generate synthesized plan

### **11.2 Synthesis Prompt**
- [ ] Design synthesis prompt that:
  - Takes all 4 plans as input
  - Identifies commonalities
  - Recognizes unique insights
  - Combines best aspects
  - Produces coherent synthesis
- [ ] Request structured output (markdown with sections)

### **11.3 Synthesized Plan Format**
- [ ] Define synthesized plan structure:
```markdown
## Synthesized Execution Plan

### Approach
[High-level approach combining insights from all plans]

### Steps
1. [Step with attribution to source plans]
2. [Step with attribution]
...

### Estimated Effort
- Time: X-Y minutes
- Complexity: [low/medium/high]
- Files: N modified, M new

### Risk Mitigation
- [Risk from plans] → [Mitigation strategy]

### Benefits
✓ [Benefit from Plan A]
✓ [Benefit from Plan B]
...
```

### **11.4 Plan Quality Checks**
- [ ] Validate synthesized plan has all required sections
- [ ] Check for missing steps
- [ ] Ensure risk mitigation is addressed
- [ ] Verify file list is complete

### **11.5 UI Integration**
- [ ] Show "Synthesizing optimal plan..." with spinner
- [ ] Display synthesized plan to user
- [ ] Allow user to review before execution
- [ ] Support manual plan editing

**Deliverable:** Plan synthesis that combines 4 diverse plans into one optimal strategy.

---

## Phase 12: Layer 5 - Validation

**Goal:** Implement validation layer with feedback loop to Main Agent.

### **12.1 Validation Layer Core**
- [ ] Create `layers/validation/` package
- [ ] Implement validation loop:
  1. Receive main agent's completion summary
  2. Receive original intent and synthesized plan
  3. Validate against checklist
  4. If issues found, send feedback to Layer 4
  5. Wait for Layer 4 to fix issues
  6. Re-validate
  7. Repeat up to max iterations (default: 3)
  8. Send final summary to user

### **12.2 Validation Checklist**
- [ ] Implement configurable validation checklist:
```yaml
validation_guidelines:
  intent_alignment:
    - Does solution address original user request?
    - Are all requirements met?
    - Are there any scope deviations?

  plan_adherence:
    - Were all critical steps completed?
    - Are deviations justified?
    - Is anything missing?

  code_quality:
    - Do all tests pass?
    - Are there linting errors?
    - Is code properly formatted?
    - Are there security concerns?

  performance:
    - Are performance targets met?
    - Were performance tests run?
    - Are there regressions?

  completeness:
    - Is documentation updated?
    - Are error cases handled?
    - Is logging/monitoring added?
```
- [ ] Allow user to customize checklist
- [ ] Support strict vs. lenient validation modes

### **12.3 Validation Execution**
- [ ] Implement each validation category:
  - **Intent alignment**: Compare output to original intent
  - **Plan adherence**: Check all steps completed
  - **Code quality**: Run linters, formatters, type checkers
  - **Performance**: Run benchmarks if specified
  - **Completeness**: Check for missing components
- [ ] Collect validation results
- [ ] Generate detailed validation report

### **12.4 Feedback Generation**
- [ ] When issues found, generate specific feedback:
  - What's wrong
  - Why it's a problem
  - How to fix it
- [ ] Send feedback to Layer 4
- [ ] Track feedback sent (for max iteration check)

### **12.5 Validation Notes**
- [ ] Maintain validation notes in context:
  - Issues found in each iteration
  - Feedback sent
  - Resolutions
- [ ] Remove validation notes when issues resolved
- [ ] Persist notes for user reference

### **12.6 UI Integration**
- [ ] Show "Validating..." indicator during validation
- [ ] Display validation results
- [ ] Show iteration count (e.g., "Validation iteration 2/3")
- [ ] Display final validation summary with ✓/⚠️ markers

**Deliverable:** Validation layer that catches issues before presenting to user, with feedback loop to Main Agent for fixes.

---

## Phase 13: Advanced Tool System

**Goal:** Expand tool system with specialized tools for git, testing, web, and documentation.

### **13.1 Git Tools**
- [ ] Create `tools/git/` package
- [ ] Implement Git tools:
  - **GitStatus**: Show git status
  - **GitDiff**: Show diff (staged, unstaged, or specific files)
  - **GitLog**: Show commit history
  - **GitCommit**: Create commit with generated message
  - **GitBranch**: List, create, checkout branches
  - **GitPull**: Pull changes
  - **GitPush**: Push changes
  - **GitStash**: Stash/pop changes
- [ ] Use go-git library or shell git
- [ ] Implement smart commit message generation
- [ ] Add tests

### **13.2 PR Tools**
- [ ] Create `tools/git/pr.go`
- [ ] Implement PR tools using GitHub CLI (gh):
  - **PRCreate**: Create PR with generated title/body
  - **PRList**: List open PRs
  - **PRView**: View PR details
  - **PRReview**: Add review comments
  - **PRMerge**: Merge PR
- [ ] Support GitLab via glab CLI
- [ ] Add tests

### **13.3 Testing Tools**
- [ ] Create `tools/test/` package
- [ ] Implement testing tools:
  - **TestRun**: Run test suite (detect and run: go test, npm test, pytest, cargo test, etc.)
  - **TestGenerate**: Generate unit tests for functions/classes
  - **Lint**: Run linters (golangci-lint, eslint, pylint, etc.)
  - **Format**: Run formatters (gofmt, prettier, black, etc.)
  - **TypeCheck**: Run type checkers (mypy, tsc, etc.)
- [ ] Auto-detect test frameworks
- [ ] Parse test output
- [ ] Add tests

### **13.4 Web Tools**
- [ ] Create `tools/web/` package
- [ ] Implement web tools:
  - **WebFetch**: Fetch URL content, convert HTML to markdown
  - **WebSearch**: Search web with provider selection (Google, Bing, DuckDuckGo)
- [ ] Handle redirects and errors
- [ ] Implement rate limiting
- [ ] Add caching
- [ ] Add tests

### **13.5 Documentation Tools**
- [ ] Create `tools/docs/` package
- [ ] Implement documentation tools:
  - **DocGenerate**: Generate documentation from code (godoc, JSDoc, etc.)
  - **DocQuery**: Answer questions about codebase using search
  - **DiagramCreate**: Generate mermaid diagrams from code structure
- [ ] Add tests

### **13.6 Security Tools**
- [ ] Create `tools/security/` package
- [ ] Implement basic security scanning:
  - **SecurityScan**: Run security scanners (gosec, npm audit, etc.)
  - **DependencyCheck**: Check for vulnerable dependencies
- [ ] Parse scanner output
- [ ] Add tests

**Deliverable:** Comprehensive tool suite covering git, testing, web, documentation, and security.

---

## Phase 13.5: Community Plugin System for Tools

**Goal:** Make tools pluggable so community can contribute custom tools through a plugin system.

### **13.5.1 Plugin Architecture**
- [ ] Create `plugins/` package
- [ ] Define plugin interface:
```go
type Plugin interface {
    Name() string
    Version() string
    Author() string
    Description() string
    Tools() []Tool  // Returns tools provided by this plugin
    Initialize(ctx context.Context, config map[string]interface{}) error
    Shutdown(ctx context.Context) error
}
```
- [ ] Define plugin manifest format (TOML/YAML):
```toml
[plugin]
name = "my-custom-tools"
version = "1.0.0"
author = "Community Member"
description = "Custom tools for XYZ"
entry_point = "./plugin.so"  # For Go plugins
# OR
entry_point = "./plugin"     # For executable plugins

[dependencies]
bplus_min_version = "0.1.0"

[[tools]]
name = "CustomTool"
description = "Does something custom"
permissions = ["read", "write"]
```

### **13.5.2 Plugin Discovery and Loading**
- [ ] Implement plugin discovery:
  - Search `~/.b+/plugins/` (user plugins)
  - Search `.b+/plugins/` (project plugins)
  - Search `/usr/local/lib/bplus/plugins/` (system plugins)
- [ ] Implement plugin loader supporting multiple formats:
  - **Go plugins** (`.so` shared libraries using Go's plugin system)
  - **Executable plugins** (standalone binaries communicating via JSON-RPC)
  - **WASM plugins** (WebAssembly for sandboxed execution)
- [ ] Validate plugin manifests
- [ ] Check plugin compatibility with current b+ version
- [ ] Implement plugin dependency resolution

### **13.5.3 Plugin Sandbox and Security**
- [ ] Implement plugin sandboxing:
  - Limit filesystem access
  - Restrict network access
  - Control resource usage (CPU, memory, time)
- [ ] Implement plugin permission system:
  - Plugins declare required permissions
  - User approves permissions on first load
  - Store approval in config
- [ ] Create plugin security scanning:
  - Checksum verification
  - Code signing support (optional)
  - Malware detection (basic)

### **13.5.4 Plugin Registry and Marketplace**
- [ ] Create plugin registry format:
```yaml
# ~/.b+/plugins/registry.yaml
plugins:
  - name: "awesome-tools"
    author: "community"
    repository: "https://github.com/user/bplus-awesome-tools"
    version: "1.2.0"
    verified: true
    downloads: 1234
    rating: 4.5
```
- [ ] Implement plugin marketplace CLI:
  - `b+ plugin search <query>` - Search for plugins
  - `b+ plugin install <name>` - Install plugin from registry
  - `b+ plugin update <name>` - Update plugin
  - `b+ plugin remove <name>` - Remove plugin
  - `b+ plugin list` - List installed plugins
  - `b+ plugin info <name>` - Show plugin details
- [ ] Create plugin repository index (GitHub-hosted JSON)
- [ ] Implement plugin verification system (community verified badge)

### **13.5.5 Plugin Development Kit (PDK)**
- [ ] Create plugin development kit with:
  - Plugin template generator: `b+ plugin init <name>`
  - Example plugins (3-5 well-documented examples)
  - Plugin testing framework
  - Plugin packaging tools
  - Documentation for plugin developers
- [ ] Create Go library for plugin development:
```go
// github.com/abrksh22/bplus/sdk
package sdk

type PluginBuilder struct {
    name        string
    version     string
    tools       []Tool
}

func NewPlugin(name, version string) *PluginBuilder
func (p *PluginBuilder) AddTool(tool Tool) *PluginBuilder
func (p *PluginBuilder) Build() Plugin
```
- [ ] Support multiple plugin languages:
  - **Go** (native, fastest)
  - **Python** (via executable plugin interface)
  - **JavaScript/TypeScript** (via Node.js executable)
  - **Rust** (via executable or WASM)

### **13.5.6 Plugin Communication Protocol**
For executable plugins (non-Go), implement JSON-RPC communication:
- [ ] Define plugin communication protocol:
```json
// Request from b+ to plugin
{
  "jsonrpc": "2.0",
  "method": "tool.execute",
  "params": {
    "tool_name": "CustomTool",
    "parameters": {
      "param1": "value1"
    }
  },
  "id": 1
}

// Response from plugin to b+
{
  "jsonrpc": "2.0",
  "result": {
    "success": true,
    "output": "...",
    "metadata": {}
  },
  "id": 1
}
```
- [ ] Implement plugin stdio communication handler
- [ ] Add timeout and error handling
- [ ] Support plugin lifecycle methods (init, shutdown, health check)

### **13.5.7 Plugin Management UI**
- [ ] Add plugin management to settings UI
- [ ] Show installed plugins with status
- [ ] Display plugin permissions
- [ ] Allow enable/disable plugins
- [ ] Show plugin resource usage
- [ ] Implement plugin update notifications

### **13.5.8 Plugin Examples and Templates**
Create example plugins:
- [ ] **Example 1**: Simple file hash tool (Go plugin)
- [ ] **Example 2**: Docker container management (executable plugin)
- [ ] **Example 3**: Jira integration (Python plugin)
- [ ] **Example 4**: Custom code generator (WASM plugin for safety)
- [ ] Document each example thoroughly

### **13.5.9 Plugin Testing and Validation**
- [ ] Implement plugin testing framework
- [ ] Test plugin loading and unloading
- [ ] Test plugin tool execution
- [ ] Test plugin error handling
- [ ] Test plugin security restrictions
- [ ] Test plugin updates
- [ ] Create automated plugin validation

### **13.5.10 Plugin Documentation**
- [ ] Write plugin developer guide:
  - How to create a plugin
  - Plugin architecture overview
  - Tool interface documentation
  - Best practices
  - Security guidelines
  - Publishing guidelines
- [ ] Create plugin submission guide
- [ ] Document plugin marketplace process
- [ ] Create plugin API reference

**Deliverable:** Complete plugin system allowing community to create and share custom tools. Includes plugin marketplace, security sandboxing, and comprehensive developer documentation.

---

## Phase 14: MCP Integration

**Goal:** Integrate Model Context Protocol for extensibility.

### **14.1 MCP Client**
- [ ] Create `mcp/` package
- [ ] Implement MCP client following specification
- [ ] Support MCP protocol messages:
  - Initialize
  - List tools
  - Call tool
  - List prompts
  - Get prompt
- [ ] Handle JSON-RPC communication

### **14.2 MCP Transports**
- [ ] Implement stdio transport:
  - Spawn MCP server process
  - Communicate via stdin/stdout
  - Handle process lifecycle
- [ ] Implement HTTP transport:
  - Connect to HTTP MCP server
  - Send requests, receive responses
- [ ] Implement SSE (Server-Sent Events) transport:
  - Connect to SSE endpoint
  - Handle streaming events

### **14.3 MCP Server Management**
- [ ] Implement MCP server registry
- [ ] Load MCP servers from config:
```yaml
mcp_servers:
  github:
    command: "github-mcp-server"
    args: []
    env:
      GITHUB_TOKEN: "${GITHUB_TOKEN}"
    transport: "stdio"

  slack:
    url: "http://localhost:3000/mcp"
    transport: "http"
```
- [ ] Start MCP servers on demand
- [ ] Keep servers alive during session
- [ ] Shutdown gracefully

### **14.4 MCP Tool Integration**
- [ ] Expose MCP tools as b+ tools
- [ ] Convert MCP tool definitions to b+ tool format
- [ ] Handle tool execution via MCP
- [ ] Parse MCP tool results
- [ ] Add permission checks for MCP tools

### **14.5 MCP Prompt Commands**
- [ ] Expose MCP prompts as slash commands
- [ ] Support prompt arguments
- [ ] Handle prompt execution
- [ ] Integrate with command system

### **14.6 Built-in MCP Servers**
- [ ] Bundle common MCP servers:
  - **GitHub**: Issues, PRs, repos
  - **File System**: Enhanced file operations
  - **Database**: SQL queries (Postgres, MySQL)
- [ ] Document how to add custom MCP servers

**Deliverable:** Full MCP integration allowing b+ to connect to 1,000+ community MCP servers and custom servers.

---

## Phase 15: Advanced UI Features

**Goal:** Build advanced UI components for settings, model selection, and visualization.

### **15.1 Settings UI**
- [ ] Create interactive settings editor
- [ ] Implement provider selection flow:
  - **Step 1**: Provider dropdown (Anthropic, OpenAI, Gemini, Ollama, etc.)
  - **Step 2**: Model list (predefined for API, dynamic for local)
  - **Step 3**: Configuration target (which layer, scope)
- [ ] Implement dynamic model fetching for Ollama/LM Studio
- [ ] Show model details (context window, pricing, status)
- [ ] Support per-layer model configuration
- [ ] Implement settings save (session, user, project)

### **15.2 Provider Configuration UI**
- [ ] Create provider configuration screens
- [ ] For API providers:
  - API key input (masked)
  - Base URL (optional)
  - Timeout, retries
  - Test connection button
- [ ] For local providers:
  - Base URL
  - GPU settings
  - Memory limits
  - Installed models list
  - Pull new model button

### **15.3 Model Testing UI**
- [ ] Create model testing interface
- [ ] Show test results:
  - Connection status
  - API key validity
  - Model availability
  - Test prompt response time
  - Context window size
  - Performance metrics (for local models)

### **15.4 Layer Visualization**
- [ ] Create layer status panel
- [ ] Show which layers are active
- [ ] Display current layer being executed
- [ ] Show layer configuration (model, settings)
- [ ] Indicate layer enable/disable state
- [ ] Display layer performance metrics

### **15.5 Cost Tracking Display**
- [ ] Create cost panel in status bar
- [ ] Show:
  - Current session cost
  - Cost per layer
  - Token usage
  - Budget remaining (if set)
  - Cost warnings
- [ ] Implement cost history graph (optional)

### **15.6 Plan Visualization**
- [ ] Create plan display component
- [ ] Show all 4 parallel plans side-by-side
- [ ] Highlight differences between plans
- [ ] Show synthesized plan
- [ ] Allow plan selection/editing

### **15.7 File Browser**
- [ ] Create file browser panel (Ctrl+B)
- [ ] Show project file tree
- [ ] Highlight modified files
- [ ] Support file selection for context
- [ ] Integrate with @ fuzzy search

### **15.8 Session Management UI**
- [ ] Create session panel (Ctrl+S)
- [ ] List all sessions
- [ ] Show session metadata (name, date, messages, cost)
- [ ] Support session actions (load, delete, export, share)
- [ ] Implement session search

**Deliverable:** Polished UI with settings editor, model selection, layer visualization, and cost tracking.

---

## Phase 16: Command System

**Goal:** Implement comprehensive slash command system and custom commands.

### **16.1 Command Parser**
- [ ] Create `commands/` package
- [ ] Implement slash command parser
- [ ] Support arguments (positional and named)
- [ ] Parse flags (e.g., `/logs --follow`)
- [ ] Handle command completion
- [ ] Add command history

### **16.2 Built-in Commands**
Implement all commands from [COMMANDS_FLAGS.md](docs/COMMANDS_FLAGS.md):

#### **16.2.1 Core Commands**
- [ ] `/help` - Show all commands
- [ ] `/clear` - Clear conversation
- [ ] `/reset` - Reset session
- [ ] `/exit`, `/quit` - Exit b+

#### **16.2.2 Session Commands**
- [ ] `/session list|save|load|delete|rename|export|import|share`
- [ ] `/checkpoint save|list|restore`
- [ ] `/resume [name]`

#### **16.2.3 Mode & Layer Commands**
- [ ] `/mode fast|thorough|status`
- [ ] `/layers status|enable|disable|reset`
- [ ] `/plans [N]|show|replay`
- [ ] `/validate [strict]`

#### **16.2.4 Model Commands**
- [ ] `/models list|current|set|test|refresh`
- [ ] `/models layer<N> <provider/model-id>`
- [ ] `/providers list|status|test|configure`

#### **16.2.5 Context Commands**
- [ ] `/context status|optimize|clear|export|health`
- [ ] `/files list|add|remove|watch|unwatch|reload`
- [ ] `/ignore list|add|remove`

#### **16.2.6 Tool Commands**
- [ ] `/tools list|enable|disable|status|test`
- [ ] `/mcp list|enable|disable|status|add|remove|test`
- [ ] `/lsp status|restart|logs|diagnostics`

#### **16.2.7 History Commands**
- [ ] `/undo [N|all]`
- [ ] `/redo [N]`
- [ ] `/history [N]|search|export`

#### **16.2.8 Testing Commands**
- [ ] `/test [pattern]|generate|watch`
- [ ] `/lint [path]|fix`
- [ ] `/format [path]`
- [ ] `/security [deep]|report`

#### **16.2.9 Explanation Commands**
- [ ] `/explain [task-id]|layers|cost|context`
- [ ] `/debug [layers|tools|context]`
- [ ] `/logs [N]|--tail|--follow|--level|export`

#### **16.2.10 Git Commands**
- [ ] `/git status|diff|commit|branch|checkout|pull|push`
- [ ] `/pr create|list|view|review|merge`

#### **16.2.11 Cost Commands**
- [ ] `/cost [history|breakdown|estimate]`
- [ ] `/metrics [layers|tools|export]`

#### **16.2.12 Settings Commands**
- [ ] `/settings [get|set|reset|export|import]`
- [ ] `/config show|edit|reload|validate`

#### **16.2.13 Documentation Commands**
- [ ] `/docs [commands|layers|tools|search]`
- [ ] `/examples [category|search]`

#### **16.2.14 Utility Commands**
- [ ] `/init [--template|--minimal]`
- [ ] `/scaffold component|api|test`

### **16.3 Custom Command System**
- [ ] Implement TOML command definition parser
- [ ] Support user-scoped commands (`~/.b+/commands/`)
- [ ] Support project-scoped commands (`.b+/commands/`)
- [ ] Support namespaced commands (`:` separator)
- [ ] Implement argument handling (`{{args}}`, `{{arg_name}}`)
- [ ] Support shell command execution in prompts
- [ ] Integrate MCP prompt commands
- [ ] Add command validation

### **16.4 Command Palette**
- [ ] Create command palette UI (Ctrl+/)
- [ ] Fuzzy search for commands
- [ ] Show command descriptions
- [ ] Show keyboard shortcuts
- [ ] Execute commands from palette

**Deliverable:** Full command system with 40+ built-in commands and custom command support.

---

## Phase 17: Advanced Session Management

**Goal:** Implement advanced session features like checkpoints, undo/redo, and sharing.

### **17.1 Checkpoint System**
- [ ] Implement automatic checkpointing:
  - Before file modifications
  - Before destructive operations
  - At regular intervals (configurable)
- [ ] Implement manual checkpointing (`/checkpoint save`)
- [ ] Store checkpoint state:
  - Messages up to checkpoint
  - File states (content snapshots)
  - Context snapshot
  - Configuration
- [ ] Implement checkpoint restoration
- [ ] Add checkpoint browsing
- [ ] Implement checkpoint cleanup (keep last N)

### **17.2 Undo/Redo System**
- [ ] Create operation history
- [ ] Track undoable operations:
  - File modifications
  - File deletions
  - Command executions
- [ ] Implement `/undo` command
- [ ] Implement `/redo` command
- [ ] Support undo count (`/undo 3`)
- [ ] Show operation being undone
- [ ] Implement undo all (`/undo all`)

### **17.3 Session Export**
- [ ] Implement session export to file
- [ ] Export format (JSON or compressed):
```json
{
  "version": "1.0",
  "session": {
    "id": "...",
    "name": "...",
    "created_at": "...",
    "metadata": {...}
  },
  "messages": [...],
  "files": [...],
  "context": {...},
  "configuration": {...}
}
```
- [ ] Support export with/without file contents
- [ ] Implement session import
- [ ] Validate imported sessions

### **17.4 Session Sharing**
- [ ] Implement read-only session snapshots
- [ ] Generate shareable links (via API or file)
- [ ] Create session viewer (web or terminal)
- [ ] Support sharing options:
  - Public (anyone with link)
  - Team (authenticated users)
  - Private (specific users)
- [ ] Implement session anonymization (remove sensitive data)

### **17.5 Multi-Session Support**
- [ ] Allow multiple sessions open simultaneously
- [ ] Implement session switching
- [ ] Show active session in status bar
- [ ] Support per-session configuration
- [ ] Implement session tabs or list view

**Deliverable:** Advanced session management with checkpoints, undo/redo, export/import, and sharing.

---

## Phase 18: Cost & Budget Management

**Goal:** Implement comprehensive cost tracking and budget controls.

### **18.1 Cost Tracking**
- [ ] Track costs per operation:
  - LLM API calls (input and output tokens)
  - Tool executions (if applicable)
- [ ] Store costs in database
- [ ] Calculate running totals (session, day, month)
- [ ] Implement provider-specific pricing:
```go
type Pricing struct {
    InputTokens  float64 // Cost per 1K input tokens
    OutputTokens float64 // Cost per 1K output tokens
    Minimum      float64 // Minimum charge per request
}
```
- [ ] Update pricing with each b+ release

### **18.2 Cost Display**
- [ ] Show current session cost in status bar
- [ ] Implement `/cost` command with:
  - Session cost breakdown
  - Cost per layer
  - Cost per tool
  - Historical costs
- [ ] Create cost visualization (graph optional)

### **18.3 Budget Controls**
- [ ] Implement `--budget` flag (session limit)
- [ ] Implement `--budget-alert` flag (warning threshold)
- [ ] Check budget before expensive operations
- [ ] Prompt user when approaching limit
- [ ] Stop execution when budget exceeded
- [ ] Support daily/monthly budget limits

### **18.4 Cost Estimation**
- [ ] Implement cost estimation for tasks:
  - Estimate tokens from prompt length
  - Factor in expected tool calls
  - Calculate expected cost range
- [ ] Show estimate before execution (if `--estimate` flag)
- [ ] Update estimate during execution

### **18.5 Cost Optimization**
- [ ] Suggest cheaper model alternatives
- [ ] Warn about expensive operations
- [ ] Implement `--free-only` flag (local models only)
- [ ] Show savings when using local models

**Deliverable:** Complete cost tracking and budget management system.

---

## Phase 19: Security & Permissions

**Goal:** Implement comprehensive security and permission system.

### **19.1 Permission Categories**
- [ ] Define permission categories:
  - **read**: File reading
  - **write**: File writing
  - **execute**: Command execution
  - **network**: Network access
  - **mcp**: MCP tool execution
- [ ] Implement permission checking
- [ ] Support permission inheritance

### **19.2 Permission Prompts**
- [ ] Implement permission prompt UI
- [ ] Show what operation requires permission
- [ ] Show operation details (file path, command, etc.)
- [ ] Support options:
  - Allow once
  - Allow for session
  - Allow always (save to config)
  - Deny
- [ ] Implement timeout for prompts

### **19.3 Auto-Approval**
- [ ] Implement `--auto-approve` flag per category
- [ ] Support approval rules in config:
```yaml
permissions:
  auto_approve:
    - read
    - write: ["src/**"]
  require_approval:
    - execute
    - network
```
- [ ] Respect `.bplusignore` for sensitive files

### **19.4 Audit Logging**
- [ ] Log all operations with permissions
- [ ] Store audit logs in database
- [ ] Include:
  - Timestamp
  - Operation type
  - Permission requested
  - Approval status
  - User identity
  - Operation details
- [ ] Implement audit log export
- [ ] Create audit log viewer

### **19.5 Sandboxing**
- [ ] Implement `--sandbox` flag
- [ ] Restrict operations in sandbox mode:
  - No file writes outside project
  - No command execution
  - No network access
- [ ] Display sandbox indicator

### **19.6 Data Encryption**
- [ ] Implement encryption for sensitive data:
  - API keys in config
  - Session exports
  - Audit logs (optional)
- [ ] Use age encryption library
- [ ] Support key management
- [ ] Implement secure key storage

### **19.7 Dangerous Command Detection**
- [ ] Detect potentially dangerous commands:
  - `rm -rf`
  - `dd`
  - `chmod 777`
  - etc.
- [ ] Warn user before execution
- [ ] Require explicit confirmation
- [ ] Log dangerous commands

**Deliverable:** Comprehensive security system with permissions, audit logs, sandboxing, and encryption.

---

## Phase 20: Testing Infrastructure

**Goal:** Build comprehensive testing suite.

### **20.1 Unit Tests**
- [ ] Write unit tests for all packages:
  - Config loading and merging
  - Provider implementations
  - Tool executions
  - Layer implementations
  - Context optimization
  - Command parsing
  - etc.
- [ ] Aim for >80% code coverage
- [ ] Use testify for assertions and mocking
- [ ] Mock external dependencies (LLM APIs, file system, etc.)

### **20.2 Integration Tests**
- [ ] Create integration test suite:
  - End-to-end agent execution
  - Multi-layer workflows
  - Provider switching
  - Session save/load
  - MCP integration
- [ ] Use test fixtures (sample projects)
- [ ] Test common workflows

### **20.3 CLI Tests**
- [ ] Test CLI flags and combinations
- [ ] Test command-line parsing
- [ ] Test configuration loading from different sources
- [ ] Test exit codes

### **20.4 UI Tests**
- [ ] Test Bubble Tea components
- [ ] Test keyboard shortcuts
- [ ] Test rendering
- [ ] Test state management

### **20.5 Performance Tests**
- [ ] Write benchmarks for critical paths:
  - Context optimization
  - File operations
  - LLM request/response
  - Database queries
- [ ] Use `benchstat` for comparing performance
- [ ] Set performance regression thresholds

### **20.6 Test Automation**
- [ ] Set up GitHub Actions for CI:
  - Run tests on every PR
  - Run tests on multiple platforms
  - Generate coverage reports
  - Run linters
- [ ] Set up nightly builds
- [ ] Set up performance regression detection

**Deliverable:** Comprehensive test suite with >80% coverage, integration tests, and CI automation.

---

## Phase 21: Documentation & Examples

**Goal:** Create comprehensive documentation and examples.

### **21.1 User Documentation**
- [ ] Write user guide:
  - Installation
  - Getting started
  - Configuration
  - Basic usage
  - Advanced features
- [ ] Create command reference (from [COMMANDS_FLAGS.md](docs/COMMANDS_FLAGS.md))
- [ ] Write layer architecture explanation
- [ ] Create troubleshooting guide
- [ ] Write FAQs

### **21.2 Developer Documentation**
- [ ] Write architecture documentation
- [ ] Document package structure
- [ ] Create API documentation (godoc)
- [ ] Write contribution guidelines
- [ ] Create development setup guide
- [ ] Document testing procedures

### **21.3 Example Configurations**
- [ ] Create example configs:
  - Minimal config
  - Solo developer config
  - Team config
  - Enterprise config
- [ ] Create example custom commands
- [ ] Create example MCP server configs

### **21.4 Example Workflows**
- [ ] Document common workflows:
  - "Fix all TypeScript errors"
  - "Implement new API endpoint"
  - "Refactor authentication system"
  - "Generate tests for module"
  - "Create PR from issue"
- [ ] Create video tutorials (optional)
- [ ] Create interactive examples

### **21.5 Migration Guides**
- [ ] Write migration guide from Claude Code
- [ ] Write migration guide from Gemini CLI
- [ ] Write migration guide from Cursor/other tools
- [ ] Include config conversion tools

### **21.6 Documentation Website**
- [ ] Create documentation website (optional):
  - Use static site generator (Hugo, Docusaurus, etc.)
  - Host on GitHub Pages or similar
  - Include search functionality
  - Include dark mode
- [ ] Or: Maintain docs in GitHub repo with good organization

**Deliverable:** Complete documentation for users and developers, with examples and migration guides.

---

## Phase 22: Distribution & Release

**Goal:** Set up distribution channels and release process.

### **22.1 GoReleaser Configuration**
- [ ] Set up GoReleaser config (`.goreleaser.yml`)
- [ ] Configure for multiple platforms:
  - macOS (amd64, arm64)
  - Linux (amd64, arm64, 386)
  - Windows (amd64, arm64, 386)
  - BSD (FreeBSD, OpenBSD, NetBSD)
- [ ] Configure binary naming
- [ ] Configure archive formats (tar.gz, zip)
- [ ] Generate checksums
- [ ] Sign binaries (optional)

### **22.2 Package Managers**

#### **22.2.1 Homebrew**
- [ ] Create Homebrew tap repository
- [ ] Generate formula automatically with GoReleaser
- [ ] Test installation: `brew install abrksh22/tap/bplus`
- [ ] Submit to homebrew-core (after stability)

#### **22.2.2 Linux Packages**
- [ ] Create APT repository for Debian/Ubuntu
- [ ] Create YUM repository for RHEL/CentOS/Fedora
- [ ] Generate DEB packages with GoReleaser
- [ ] Generate RPM packages with GoReleaser
- [ ] Host repository on GitHub Pages or similar

#### **22.2.3 Scoop (Windows)**
- [ ] Create Scoop manifest
- [ ] Submit to Scoop bucket
- [ ] Test installation: `scoop install bplus`

#### **22.2.4 Arch Linux**
- [ ] Create AUR package
- [ ] Test installation: `yay -S bplus` or `paru -S bplus`

#### **22.2.5 Go Install**
- [ ] Ensure `go install github.com/abrksh22/bplus@latest` works
- [ ] Document in README

### **22.3 Docker**
- [ ] Create Dockerfile
- [ ] Build multi-arch images (amd64, arm64)
- [ ] Push to Docker Hub
- [ ] Push to GitHub Container Registry
- [ ] Document Docker usage

### **22.4 Auto-Update**
- [ ] Implement update checking
- [ ] Check for new releases on GitHub
- [ ] Notify user of updates
- [ ] Implement self-update command (`b+ update`)
- [ ] Support update channels (stable, beta, nightly)

### **22.5 Release Process**
- [ ] Document release process:
  1. Update version in code
  2. Update CHANGELOG.md
  3. Create git tag
  4. Push tag (triggers GoReleaser via GitHub Actions)
  5. GoReleaser builds and publishes
  6. Update documentation
  7. Announce on social media, Reddit, Hacker News
- [ ] Create release checklist
- [ ] Set up release GitHub Action

### **22.6 Analytics & Telemetry (Optional)**
- [ ] Implement privacy-respecting telemetry:
  - Version information
  - Platform information
  - Feature usage (anonymous)
  - Error reports (opt-in)
- [ ] Make telemetry opt-in (or opt-out with clear consent)
- [ ] Document what data is collected
- [ ] Provide opt-out mechanism

**Deliverable:** Complete distribution setup. Users can install b+ via Homebrew, apt, yum, Scoop, Docker, or direct download.

---

## Development Best Practices

### **Code Organization**
```
b+/
├── cmd/
│   └── bplus/           # Main entry point
│       └── main.go
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── storage/         # Database layer
│   ├── logging/         # Logging utilities
│   └── errors/          # Error handling
├── layers/              # AI layer implementations
│   ├── intent/          # Layer 1
│   ├── planning/        # Layer 2
│   ├── synthesis/       # Layer 3
│   ├── execution/       # Layer 4
│   ├── validation/      # Layer 5
│   └── context/         # Layer 6
├── models/              # LLM provider system
│   ├── providers/       # Provider implementations
│   └── router/          # Model routing
├── tools/               # Tool system
│   ├── file/            # File operations
│   ├── exec/            # Command execution
│   ├── git/             # Git operations
│   ├── test/            # Testing tools
│   └── web/             # Web tools
├── ui/                  # Terminal UI
│   ├── components/      # UI components
│   ├── themes/          # Themes
│   └── views/           # Views
├── mcp/                 # MCP integration
├── lsp/                 # LSP integration
├── commands/            # Command system
├── security/            # Security & permissions
├── prompts/             # AI prompts
├── docs/                # Documentation
├── examples/            # Example configs
└── tests/               # Test suites
```

### **Naming Conventions**
- **Packages**: lowercase, single word (avoid underscores)
- **Files**: snake_case
- **Functions/Methods**: PascalCase (exported), camelCase (private)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE
- **Interfaces**: End with `er` suffix (e.g., `Provider`, `Executer`)

### **Error Handling**
- Always wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return errors, don't panic (except in init or truly unrecoverable situations)
- Log errors at appropriate levels

### **Testing**
- Test file naming: `*_test.go`
- Table-driven tests when possible
- Use subtests for multiple cases: `t.Run("case name", func(t *testing.T) {...})`
- Mock external dependencies
- Test both success and failure paths

### **Logging**
- Use structured logging (zerolog)
- Log levels: Debug, Info, Warn, Error
- Include relevant context in logs
- Don't log sensitive data (API keys, tokens)

### **Configuration**
- Use environment variables for secrets
- Use config files for non-sensitive settings
- Support multiple config locations
- Validate configuration early

### **Performance**
- Use goroutines for parallel operations
- Use buffered channels for producer-consumer
- Profile regularly: `go test -bench=. -cpuprofile=cpu.prof`
- Optimize database queries
- Cache expensive computations

---

## Success Criteria

### **Phase 6 (MVP) Success:**
- User can install b+
- User can configure API key
- User can send a message and get a response
- Agent can use tools (Read, Write, Edit, Glob, Grep, Bash)
- Agent completes a simple task (e.g., "Create a Hello World program")

### **Phase 12 (Thorough Mode) Success:**
- User can enable Thorough Mode
- All 7 layers execute successfully
- Intent clarification works
- 4 parallel plans are generated
- Plans are synthesized
- Main agent executes task
- Validation catches errors and provides feedback
- Context is optimized

### **Phase 22 (Release) Success:**
- b+ is installable via Homebrew (macOS) and apt (Linux)
- Documentation is complete and clear
- Users can successfully complete common workflows
- Test suite passes on all platforms
- CI/CD is automated
- Community is engaged (GitHub stars, issues, PRs)

---

## Notes

- **Incremental Development**: Each phase should leave the system in a working state
- **Testing**: Write tests as you build, not afterward
- **Documentation**: Update docs with each phase
- **Performance**: Profile and optimize regularly, don't wait until the end
- **User Feedback**: Get feedback early and often, even before all features are complete
- **Flexibility**: This plan may evolve based on learnings and feedback

---

**Plan Version:** 1.0
**Date:** 2025-10-25
**Status:** Ready for Implementation

