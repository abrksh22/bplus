# Contributing to b+ (Be Positive)

Thank you for your interest in contributing to b+! We welcome contributions from the community.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/abrksh22/bplus/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - System information (OS, Go version, b+ version)
   - Relevant logs or error messages

### Suggesting Features

1. Check [Discussions](https://github.com/abrksh22/bplus/discussions) and [Issues](https://github.com/abrksh22/bplus/issues)
2. Create a new discussion or issue with:
   - Clear use case
   - Proposed solution
   - Alternative solutions considered
   - Impact on existing functionality

### Contributing Code

#### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/bplus.git
   cd bplus
   ```

2. **Install dependencies:**
   ```bash
   make deps
   make tools
   ```

3. **Create a branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. **Make your changes** following the [Development Principles](docs/PLAN.md#development-principles)

2. **Write tests** for new functionality

3. **Run checks:**
   ```bash
   make check  # Runs fmt, vet, lint, test
   ```

4. **Build and test:**
   ```bash
   make build
   make test
   ```

5. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `chore:` - Maintenance tasks
   - `refactor:` - Code refactoring
   - `test:` - Test additions/changes
   - `perf:` - Performance improvements

6. **Push and create a Pull Request:**
   ```bash
   git push origin feature/your-feature-name
   ```

#### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Keep functions small and focused
- Write clear, self-documenting code
- Add comments for complex logic

#### Testing

- Write unit tests for all new code
- Maintain or increase code coverage
- Test edge cases and error conditions
- Run `make test-coverage` to check coverage

### Contributing Plugins

See [PLUGIN_DEVELOPMENT.md](docs/PLUGIN_DEVELOPMENT.md) for creating community plugins.

## Pull Request Process

1. **Ensure your PR:**
   - Has a clear title and description
   - Links to related issues
   - Passes all CI checks
   - Has appropriate tests
   - Updates documentation if needed

2. **Review process:**
   - Maintainers will review your PR
   - Address any feedback or requested changes
   - Once approved, a maintainer will merge

3. **After merge:**
   - Your contribution will be included in the next release
   - You'll be added to the contributors list

## Project Structure

See [PLAN.md](docs/PLAN.md) for detailed architecture and development phases.

```
b+/
├── cmd/bplus/          # CLI entry point
├── internal/           # Core infrastructure
├── layers/             # 7-layer AI system
├── models/             # Provider integrations
├── tools/              # Tool system
├── plugins/            # Plugin system
├── ui/                 # Terminal UI
├── mcp/                # MCP integration
└── lsp/                # LSP integration
```

## Development Phases

We're currently in **Phase 1: Foundation**. Check [PLAN.md](docs/PLAN.md) to see what's being worked on and where you can help.

## Areas Where We Need Help

- **Provider Integrations**: AWS Bedrock, Azure OpenAI, custom providers
- **Tool Implementations**: Database tools, cloud platform tools, etc.
- **Plugin Development**: Community tools and integrations
- **Documentation**: User guides, tutorials, examples
- **Testing**: Unit tests, integration tests, bug reports
- **UI/UX**: Terminal UI improvements, themes
- **Translations**: Internationalization support

## Questions?

- **Discussions**: [GitHub Discussions](https://github.com/abrksh22/bplus/discussions)
- **Issues**: [GitHub Issues](https://github.com/abrksh22/bplus/issues)
- **Discord**: Coming soon

## License

By contributing to b+, you agree that your contributions will be licensed under the MIT License.

---

Thank you for helping make b+ better! 🎉
