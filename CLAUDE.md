# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL RULES - READ FIRST

**These rules override all other instructions and must be followed at all times:**

### 1. Understand User Intent First
- **ALWAYS** carefully analyze the user's request to understand their true intent
- Ask clarifying questions if the request is ambiguous
- Confirm your understanding before starting significant work
- Consider the broader context and goals, not just literal instructions
- **The user's intent is paramount** - implement what they mean, not just what they say

### 2. Documentation Discipline
- **DO NOT create document junk** - no unnecessary files, no redundant documentation
- Only create new documentation files when explicitly requested or absolutely necessary
- Update existing documentation instead of creating new files
- Keep documentation concise, accurate, and up-to-date
- **Update ALL affected documentation when a task is finished** - not before, not partially

### 3. Commit Discipline
- **ONLY commit when a task is 100% complete** - no partial commits
- Ensure all tests pass before committing
- Ensure code quality checks pass (lint, fmt, vet) before committing
- **Write clear, descriptive commit messages** that explain what was done and why
- Follow conventional commit format: `type(scope): description`
  - Examples: `feat(config): add Viper-based configuration system`, `fix(logging): correct zerolog initialization`
- Do NOT commit if there are any outstanding issues or incomplete work

### 4. Quality Standards
- Write tests alongside implementation (aim for >80% coverage)
- Run `make check` before considering a task complete
- Fix all linting errors and warnings
- Ensure all code is properly formatted (`make fmt`)
- Validate that the implementation meets requirements from PLAN.md

### 5. Phase Discipline
- **ONLY implement features from the current phase** (see `docs/VERIFICATION.md`)
- Do NOT skip ahead to future phases
- Do NOT implement features not in the plan
- Complete all critical requirements of a phase before marking it complete
- Update VERIFICATION.md when a phase is complete

---

## Project Overview

**b+ (Be Positive)** is an intelligent, model-agnostic, privacy-first agentic terminal coding assistant built in Go. It features a revolutionary 7-layer AI architecture that separates intent clarification, parallel planning, synthesis, execution, validation, and context management into independent layers.

**Current Status:** Phase 1 (Foundation) complete. Early development stage.

## Development Commands

### Building
```bash
make build          # Build binary to bin/bplus
make run            # Run directly with go run
make install        # Install to $GOPATH/bin
```

### Testing
```bash
make test           # Run all tests
make test-verbose   # Run tests with verbose output
make test-coverage  # Generate coverage report (coverage.html)
make bench          # Run benchmarks
```

### Code Quality
```bash
make lint           # Run golangci-lint (installs if needed)
make fmt            # Format code with go fmt and gofumpt
make vet            # Run go vet
make check          # Run fmt, vet, lint, test (pre-commit check)
```

### Development Workflow
```bash
make dev            # Run with hot reload (requires air)
make ci             # Run all CI checks (lint, test, build)
make deps           # Download and tidy dependencies
make tools          # Install development tools
```

### Cleanup
```bash
make clean          # Remove build artifacts, coverage files
```

## Project Architecture

### 7-Layer AI System

b+ uses a unique multi-layer architecture where each layer serves a specific purpose:

1. **Layer 1 (Intent Clarification)**: Conversational clarification of user requirements before execution
2. **Layer 2 (Parallel Planning)**: Generates 4 diverse execution plans simultaneously using different models/prompts
3. **Layer 3 (Synthesis)**: Combines the best aspects of all 4 plans into one optimal strategy
4. **Layer 4 (Main Agent)**: The execution layer with full tool access - this is "Fast Mode"
5. **Layer 5 (Validation)**: Validates outputs against intent and plan, provides feedback loop to Layer 4
6. **Layer 6 (Context Management)**: Persistent layer managing context optimization and session state
7. **Layer 7 (Oversight)**: Reserved for future compliance, security, and analytics

**Operating Modes:**
- **Fast Mode** (default): Only Layer 4 active - single-agent execution for quick tasks
- **Thorough Mode** (`--thorough`): All layers active for complex, critical work

### Directory Structure

```
b+/
├── cmd/bplus/          # CLI entry point - main.go
├── internal/           # Private core infrastructure
│   ├── config/         # Configuration system (Viper-based)
│   ├── storage/        # Database layer (SQLite + bbolt)
│   ├── logging/        # Structured logging (zerolog)
│   └── errors/         # Custom error types
├── layers/             # 7-layer AI implementation
│   ├── intent/         # Layer 1: Intent clarification
│   ├── planning/       # Layer 2: Parallel planning
│   ├── synthesis/      # Layer 3: Plan synthesis
│   ├── execution/      # Layer 4: Main agent (Fast Mode)
│   ├── validation/     # Layer 5: Output validation
│   ├── context/        # Layer 6: Context management
│   └── oversight/      # Layer 7: Future/reserved
├── models/             # LLM provider system
│   ├── providers/      # Provider implementations (Anthropic, OpenAI, Gemini, Ollama, etc.)
│   └── router/         # Intelligent model selection and routing
├── tools/              # Pluggable tool system (25+ planned)
│   ├── file/           # Read, Write, Edit, Glob, Grep
│   ├── exec/           # Bash, process management
│   ├── git/            # Git operations
│   ├── test/           # Testing and validation tools
│   ├── web/            # WebFetch, WebSearch
│   ├── docs/           # Documentation generation
│   └── security/       # Security scanning
├── ui/                 # Terminal UI (Bubble Tea)
│   ├── components/     # Reusable UI components
│   ├── themes/         # Color schemes
│   └── views/          # Main view implementations
├── lsp/                # Language Server Protocol integration
├── mcp/                # Model Context Protocol integration
├── plugins/            # Plugin system for community tools
├── prompts/            # System prompts for each layer
├── commands/           # Slash command system
├── security/           # Permissions and sandboxing
└── docs/               # Architecture and planning docs
```

### Key Design Principles

1. **Incremental Development**: Each phase delivers working, testable functionality
2. **Interface-First**: Define interfaces before implementations for modularity
3. **Configuration Over Code**: Use YAML config files, not hardcoded values
4. **Local-First**: Prioritize local models for privacy and zero-cost operation
5. **Model-Agnostic**: Support multiple providers (cloud and local) with `provider/model-id` format

### Model Provider System

Models are specified using the format: `provider/model-id`

**Supported Providers:**
- **Cloud**: `anthropic/model-id`, `openai/model-id`, `gemini/model-id`, `openrouter/model-id`
- **Local**: `ollama/model-id`, `lmstudio/model-id`

Each layer can use a different model, configured in `~/.config/bplus/config.yaml` or `.b+/config.yaml`.

### Tool System Architecture

Tools are designed to be pluggable from day one:

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() []Parameter
    RequiresPermission() bool
    Execute(ctx context.Context, params map[string]interface{}) (*Result, error)

    // Plugin support
    Category() string
    Version() string
    IsExternal() bool
}
```

**Core Tools** (file, exec, git, test, web, docs, security)
**MCP Tools** (access to 1,000+ community MCP servers)
**Plugin Tools** (community-contributed via plugin marketplace - Phase 13.5)

### Configuration System

Configuration precedence (highest to lowest):
1. CLI flags
2. Environment variables
3. Project config (`.b+/config.yaml`)
4. User config (`~/.config/bplus/config.yaml`)
5. System defaults

Configuration uses Viper with support for YAML, TOML, and JSON formats.

## Development Guidelines

### Code Style

- Follow standard Go conventions
- Use `gofmt` and `gofumpt` for formatting
- Run `golangci-lint` before commits (configured in `.golangci.yml`)
- Write tests alongside implementation (aim for >80% coverage)
- Use table-driven tests with subtests

### Error Handling

- Always wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return errors rather than panic (except in init or unrecoverable situations)
- Use custom error types from `internal/errors/` for categorization
- Provide user-friendly error messages

### Logging

- Use structured logging via zerolog
- Log levels: Debug, Info, Warn, Error
- Never log sensitive data (API keys, tokens)
- Include relevant context in all logs

### Testing

- Test file naming: `*_test.go`
- Use `testify` for assertions and mocking
- Mock external dependencies (LLM APIs, filesystem, network)
- Test both success and failure paths
- Write benchmarks for performance-critical code

### Naming Conventions

- **Packages**: lowercase, single word (avoid underscores)
- **Files**: snake_case (e.g., `model_router.go`)
- **Functions/Methods**: PascalCase (exported), camelCase (unexported)
- **Interfaces**: End with `er` suffix (e.g., `Provider`, `Executor`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE

### Concurrency

- Use goroutines for parallel operations (especially in Layer 2 planning)
- Use buffered channels for producer-consumer patterns
- Always use context for cancellation and timeouts
- Profile with `go test -bench=. -cpuprofile=cpu.prof`

## Important Context

### Phase-Based Development

Development follows a phased approach (see `docs/PLAN.md`):

- **Phase 1** (✅ Complete): Foundation & project setup
- **Phase 2** (Current): Core infrastructure (config, logging, storage)
- **Phase 6**: Layer 4 MVP (Fast Mode) - first end-to-end working version
- **Phase 12**: Thorough Mode (all 7 layers)
- **Phase 13.5**: Community plugin system
- **Phase 22**: Public release v1.0

Always check the current phase before implementing features from later phases.

### Technology Stack

- **Language**: Go 1.21+ (chosen for performance, concurrency, and cross-platform support)
- **Terminal UI**: Bubble Tea ecosystem (bubbletea, lipgloss, bubbles, glamour)
- **Database**: SQLite (sessions, FTS5 search) + bbolt (real-time KV store)
- **LLM Integration**: Custom provider abstractions supporting OpenAI, Anthropic, Google, Ollama
- **LSP**: go-lsp for Language Server Protocol integration
- **MCP**: Custom implementation of Model Context Protocol

### Critical Files

- `docs/VISION.md`: Complete vision, architecture, and competitive analysis
- `docs/PLAN.md`: Detailed 22-phase implementation plan
- `docs/VERIFICATION.md`: Phase completion tracking and verification
- `docs/COMMANDS_FLAGS.md`: Complete command and flag reference
- `Makefile`: All development commands
- `.golangci.yml`: Comprehensive linter configuration

### Anti-Hallucination Architecture

b+ is designed to minimize hallucinations through:

1. **Layer 5 Validation**: Multi-iteration feedback loop catches 80-90% of errors
2. **Tool Execution Verification**: Results are validated before being shown to users
3. **LSP Integration**: Real-time code intelligence prevents invalid code generation
4. **Test Generation**: Automatic test creation validates generated code

### Cost Optimization

The 7-layer architecture is designed for cost efficiency:

- Fast Mode: Only Layer 4 active (~$0.50 per task)
- Thorough Mode: All layers (~$2.40 per task vs Claude Code's $4.80)
- Local model support for zero-cost operation on simple tasks
- Intelligent routing between cheap and expensive models based on complexity

## Common Development Tasks

### Adding a New Provider

1. Create package in `models/providers/<name>/`
2. Implement the `Provider` interface
3. Add provider configuration to config schema
4. Register provider in provider registry
5. Add tests with mock server
6. Document in configuration guide

### Adding a New Tool

1. Create tool in appropriate `tools/<category>/` package
2. Implement the `Tool` interface
3. Define parameters and validation
4. Add permission requirements
5. Write comprehensive tests
6. Register tool in tool registry
7. Update tool documentation

### Adding a New Layer Enhancement

1. Navigate to appropriate `layers/<name>/` package
2. Review layer interface and responsibilities (see `docs/VISION.md`)
3. Maintain layer isolation - communicate only through defined interfaces
4. Ensure layer can be independently enabled/disabled
5. Add layer-specific configuration options
6. Test layer in isolation and integrated

### Running in Development

```bash
# Fast iteration with hot reload
make dev

# Or manual run with specific settings
go run cmd/bplus/main.go --model anthropic/claude-sonnet-4-5 --fast

# Build and test
make build
./bin/bplus
```

## Documentation

- See `docs/VISION.md` for the complete vision and 7-layer architecture details
- See `docs/PLAN.md` for the 22-phase implementation roadmap
- See `docs/VERIFICATION.md` for current phase completion status
- See `README.md` for user-facing overview and quick start

## Project Goals

1. **Quality**: Achieve 75-85% SWE-bench accuracy (comparable to Claude Code's 72.5%)
2. **Cost**: 60-80% cheaper than cloud-only solutions through local model routing
3. **Privacy**: 100% local operation option with air-gapped mode
4. **Speed**: Complete tasks in <45 minutes (vs Claude's 77 min, Gemini's 120+ min)
5. **Community**: Open source (MIT), plugin marketplace, active contributor community

## License

MIT License - see `LICENSE` file for details.
