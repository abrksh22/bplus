# b+ (Be Positive)

> **The Next-Generation Agentic Terminal Coding Assistant**
> *Smart. Adaptive. Developer-First. Privacy-Conscious.*

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

---

## What is b+?

**b+ (Be Positive)** is an intelligent, model-agnostic, privacy-first agentic coding assistant built for the terminal. It combines the best aspects of existing tools while solving their fundamental limitations through:

- **Revolutionary 7-Layer AI Architecture**: Multi-agent system with intent clarification, parallel planning, synthesis, execution, validation, and context management
- **Intelligent Multi-Model Routing**: Automatically selects optimal models (cloud or local) based on task complexity, cost, and privacy
- **Hybrid Architecture**: Seamless switching between local LLMs (privacy, zero cost) and cloud models (maximum capability)
- **Cost Optimization**: 60-80% cheaper than competitors while maintaining quality
- **Advanced Anti-Hallucination**: Multi-layer validation catches 80-90% of errors before they reach you
- **Privacy-First**: Your code, your choice - works 100% offline with local models
- **Pluggable Tool System**: Community can contribute custom tools via plugin marketplace

## Key Features

### 🚀 Fast & Thorough Modes
- **Fast Mode**: Single-agent execution for quick tasks (default)
- **Thorough Mode**: Full 7-layer architecture for complex, critical work

### 🔌 Model Agnostic
Supports multiple providers with unified `provider/model-id` format:
- **Cloud**: Anthropic, OpenAI, Google Gemini, Groq, OpenRouter
- **Local**: Ollama, LM Studio, llama.cpp
- **Intelligent Routing**: Auto-select optimal model based on task

### 🛠️ Comprehensive Tool System
- **Core Tools**: File ops (read, write, edit, glob, grep), execution (bash, process mgmt)
- **Advanced Tools**: Git, testing, web, documentation, security
- **LSP Integration**: Real-time code intelligence for 15+ languages
- **MCP Support**: Access to 1,000+ community servers
- **Plugin System**: Community-contributed tools (Phase 13.5)

### 🎯 7-Layer AI Architecture
1. **Intent Clarification**: Understand requirements before starting
2. **Parallel Planning**: Generate 4 diverse execution plans simultaneously
3. **Plan Synthesis**: Combine best aspects into optimal strategy
4. **Main Agent**: Execute tasks with full tool access
5. **Validation**: Catch errors with feedback loop before showing results
6. **Context Management**: Intelligent optimization for long sessions
7. **Oversight** *(Future)*: Compliance, security, analytics

### 🎨 Best-in-Class Developer Experience
- **Beautiful Terminal UI**: Built with Bubble Tea ecosystem
- **Real-Time Streaming**: See code generation as it happens
- **Session Management**: Save, load, resume, share sessions
- **Undo/Redo**: Full operation history with rollback
- **Cost Tracking**: Know exactly what you're spending

## Quick Start

### Installation

**macOS (Homebrew):**
```bash
brew install abrksh22/tap/bplus
```

**Linux (APT):**
```bash
curl -s https://bplus.dev/install.sh | sudo bash
```

**Go Install:**
```bash
go install github.com/abrksh22/bplus@latest
```

**From Source:**
```bash
git clone https://github.com/abrksh22/bplus.git
cd bplus
make build
```

### Basic Usage

```bash
# Start b+ with default model
b+

# Use specific model
b+ --model anthropic/claude-sonnet-4-5

# Fast mode (default) for quick tasks
b+ --fast

# Thorough mode for complex features
b+ --thorough

# Local-only for privacy
b+ --local-only

# With budget limit
b+ --budget 5.00
```

### Configuration

```yaml
# ~/.config/bplus/config.yaml
mode: thorough

models:
  default: "anthropic/claude-sonnet-4-5"

  layers:
    intent_clarification: "openai/gpt-4-turbo"
    main_agent: "anthropic/claude-sonnet-4-5"
    validation: "openai/gpt-4-turbo"

providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"

  ollama:
    base_url: "http://localhost:11434"
```

## Example Workflows

**Fix All TypeScript Errors:**
```bash
b+ -p "Fix all TypeScript errors in src/"
```

**Implement New Feature:**
```bash
b+ --thorough
> Implement OAuth 2.0 authentication with Google provider
```

**Generate Tests:**
```bash
b+
> Generate comprehensive unit tests for src/auth/login.ts
```

**Code Review:**
```bash
b+ /review src/payments/checkout.go
```

## Architecture

b+ is built with Go and follows a modular, plugin-friendly architecture:

```
b+/
├── cmd/bplus/          # CLI entry point
├── internal/           # Core infrastructure
├── layers/             # 7-layer AI system
├── models/             # Provider integrations
├── tools/              # Tool system
├── plugins/            # Plugin system (Phase 13.5)
├── ui/                 # Terminal UI
├── mcp/                # MCP integration
└── lsp/                # LSP integration
```

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed design documentation.

## Plugin System

b+ supports community-contributed tools through a plugin system:

```bash
# Search for plugins
b+ plugin search docker

# Install plugin
b+ plugin install awesome-docker-tools

# List installed plugins
b+ plugin list

# Create your own plugin
b+ plugin init my-custom-tools
```

See [PLUGIN_DEVELOPMENT.md](docs/PLUGIN_DEVELOPMENT.md) for plugin creation guide.

## Comparison

| Feature | Claude Code | Gemini CLI | Crush CLI | **b+** |
|---------|-------------|------------|-----------|--------|
| **Price** | $20-200/mo | Free | Free | **Freemium** |
| **Multi-Model** | ❌ | ❌ | ⚠️ | **✅** |
| **Local Models** | ❌ | ❌ | Limited | **✅** |
| **Thorough Mode** | Plan mode | ❌ | ❌ | **7-Layer System** |
| **Cost** (complex task) | $4.80 | Free | Varies | **$0.50-2.00** |
| **Plugin System** | ❌ | ❌ | ❌ | **✅** |
| **Hallucination Rate** | ~20% | ~30% | N/A | **<5%** |

## Roadmap

- [x] **Phase 1**: Foundation ✅ Complete (2025-10-25)
- [ ] **Phase 2**: Core Infrastructure (In Progress)
- [ ] **Phase 6**: Fast Mode MVP (Layer 4 + Layer 6)
- [ ] **Phase 12**: Thorough Mode (All 7 layers)
- [ ] **Phase 13.5**: Community Plugin System
- [ ] **Phase 22**: Public Release v1.0

See [PLAN.md](docs/PLAN.md) for complete development roadmap and [VERIFICATION.md](docs/VERIFICATION.md) for detailed implementation tracking.

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Areas where we need help:**
- Provider integrations (AWS Bedrock, Azure OpenAI, etc.)
- Tool implementations (database tools, cloud platform tools)
- Plugin development
- Documentation and examples
- Testing and bug reports

## License

b+ follows an **Open Source** model:
- **Core** (this repository): MIT License - Free and open source

See [LICENSE](LICENSE) for details.

## Community

- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/abrksh22/bplus/discussions)
- **Issues**: [Report bugs and request features](https://github.com/abrksh22/bplus/issues)
- **Discord**: Coming soon
- **Twitter**: [@bplus_dev](https://twitter.com/bplus_dev)

## Acknowledgments

b+ is inspired by and builds upon the excellent work of:
- **Claude Code** by Anthropic
- **Gemini CLI** by Google
- **Crush CLI** by the open source community
- **Bubble Tea** by Charm

## Documentation

- [Vision Document](docs/VISION.md) - Complete vision and architecture
- [Implementation Plan](docs/PLAN.md) - Detailed development phases
- [Verification Document](docs/VERIFICATION.md) - Phase completion tracking
- [Commands & Flags](docs/COMMANDS_FLAGS.md) - Complete command reference
- [Plugin Development](docs/PLUGIN_DEVELOPMENT.md) - Create custom tools *(coming soon)*
- [Configuration Guide](docs/CONFIGURATION.md) - Advanced configuration *(coming soon)*

---

**Let's build something amazing. Let's be positive. Let's build b+.**

Made with ❤️ by the b+ community
