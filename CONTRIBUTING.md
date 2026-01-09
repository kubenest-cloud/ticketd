# Contributing to TicketD

First off, thank you for considering contributing to TicketD! It's people like you that make TicketD a great tool for the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Guidelines](#coding-guidelines)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Community](#community)

---

## Code of Conduct

This project adheres to a simple code of conduct:
- **Be respectful**: Treat everyone with respect and kindness
- **Be constructive**: Provide helpful feedback and suggestions
- **Be collaborative**: Work together toward common goals
- **Be inclusive**: Welcome contributors of all backgrounds and skill levels

By participating in this project, you agree to uphold these values.

---

## Getting Started

### Prerequisites

Before you begin, ensure you have:
- **Go 1.21+** installed ([download here](https://go.dev/dl/))
- **Git** for version control
- **A C compiler** (required for SQLite):
  - **Linux**: `gcc` (usually pre-installed)
  - **macOS**: Xcode Command Line Tools (`xcode-select --install`)
  - **Windows**: [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/)

### Fork and Clone

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ticketd.git
   cd ticketd
   ```
3. **Add upstream** remote:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/ticketd.git
   ```

---

## Development Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Create a `.env` file

```bash
cat > .env << EOF
TICKETD_ADMIN_USER=admin
TICKETD_ADMIN_PASS=password123
TICKETD_PORT=8080
TICKETD_DB_PATH=ticketd_dev.db
TICKETD_PUBLIC_BASE_URL=http://localhost:8080
EOF
```

### 3. Run the Application

```bash
go run .
```

The admin dashboard will be available at `http://localhost:8080/admin`

### 4. Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 5. Code Formatting

Before committing, ensure your code is properly formatted:

```bash
# Format code
go fmt ./...

# Run linter (install golangci-lint first)
golangci-lint run

# Vet code
go vet ./...
```

---

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

#### 🐛 Bug Reports
- Use the GitHub issue tracker
- Provide a clear title and description
- Include steps to reproduce
- Share relevant logs and error messages
- Mention your Go version and OS

#### 💡 Feature Requests
- Check existing issues first
- Explain the problem you're trying to solve
- Describe your proposed solution
- Consider multiple approaches
- Be open to discussion and alternatives

#### 📝 Documentation Improvements
- Fix typos and grammatical errors
- Improve clarity and examples
- Add missing documentation
- Update outdated information

#### 🔧 Code Contributions
- Fix bugs
- Implement new features
- Refactor code
- Improve performance
- Add tests

---

## Coding Guidelines

TicketD follows Go best practices and conventions. Please adhere to these guidelines:

### Go Style Guide

Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

### Project-Specific Guidelines

#### 1. **Package Organization**

```
ticketd/
├── main.go                    # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── errors/               # Custom error types
│   ├── validator/            # Input validation
│   ├── store/                # Data models and interfaces
│   │   └── sqlite/           # SQLite implementation
│   └── web/                  # HTTP handlers and templates
│       ├── templates/        # HTML templates
│       └── static/           # Static assets
```

#### 2. **Naming Conventions**

- **Packages**: Short, lowercase, single-word names (e.g., `store`, `web`)
- **Files**: Lowercase with underscores (e.g., `handlers_admin.go`)
- **Variables**: camelCase for locals, PascalCase for exports
- **Constants**: PascalCase or UPPER_SNAKE_CASE for enums

```go
// Good
const StatusOpen = "OPEN"
var maxRetries = 3
type Client struct { ... }

// Bad
const status_open = "OPEN"
var MaxRetries = 3
type client struct { ... }
```

#### 3. **Error Handling**

Always handle errors explicitly. Use the custom error types in `internal/errors`:

```go
// Good
client, err := store.GetClient(id)
if err != nil {
    if apperrors.IsNotFound(err) {
        return nil, apperrors.NotFoundError("client", id)
    }
    return nil, apperrors.Wrap(err, "failed to fetch client")
}

// Bad
client, _ := store.GetClient(id)  // Don't ignore errors!
```

#### 4. **Validation**

Use the validator package for all input validation:

```go
// Good
input = validator.TrimSubmissionInput(input)
if err := validator.ValidateSubmission(input); err != nil {
    return err
}

// Bad
if input.Message == "" {  // Don't do ad-hoc validation
    return errors.New("message required")
}
```

#### 5. **Documentation**

All exported functions, types, and packages must have godoc comments:

```go
// Client represents a client organization that can create forms.
// Each client has an allowed domain used for CORS validation.
type Client struct {
    ID            int64
    Name          string
    AllowedDomain string
    CreatedAt     time.Time
}

// CreateClient creates a new client after validating the input.
// Returns an error if validation fails or the database operation fails.
func (s *Store) CreateClient(name, allowedDomain string) (Client, error) {
    // Implementation...
}
```

#### 6. **Database Operations**

- Use parameterized queries (already done)
- Wrap errors with context
- Check rows affected for UPDATE/DELETE
- Use transactions for multi-step operations (when needed)

```go
result, err := s.db.Exec(`UPDATE clients SET name = ? WHERE id = ?`, name, id)
if err != nil {
    return apperrors.Wrapf(err, "failed to update client %d", id)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
    return apperrors.Wrap(err, "failed to check rows affected")
}
if rowsAffected == 0 {
    return apperrors.NotFoundError("client", id)
}
```

#### 7. **Testing**

- Write table-driven tests where appropriate
- Test error cases, not just happy paths
- Use meaningful test names
- Use `t.Helper()` in test helper functions

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"empty email", "", false}, // Optional field
        {"invalid format", "not-an-email", true},
        {"too long", strings.Repeat("a", 300) + "@example.com", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### 8. **Logging**

Use structured logging with `log/slog`:

```go
// Good
slog.Info("Client created", "client_id", client.ID, "name", client.Name)
slog.Error("Failed to create client", "error", err, "name", name)

// Bad
log.Printf("Client created: %d", client.ID)  // Unstructured
```

#### 9. **Performance**

- Avoid premature optimization
- Don't allocate in hot paths unnecessarily
- Use appropriate data structures
- Profile before optimizing

#### 10. **Security**

- Never log sensitive data (passwords, tokens)
- Validate all user input
- Use parameterized queries (already done)
- Sanitize output in templates (Go templates auto-escape)
- Check CORS origins (already implemented)

---

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks, dependency updates

### Examples

```
feat(web): add copy-to-clipboard for embed codes

Add a copy button next to each embed code in the forms page.
Displays a success notification when copied.

Closes #42
```

```
fix(store): handle IN_PROGRESS status correctly

The status field was using spaces instead of underscores,
causing validation errors. Updated to use IN_PROGRESS
consistently.

Fixes #38
```

```
docs: add contributing guidelines and improve README

- Add comprehensive CONTRIBUTING.md
- Improve README with badges and better structure
- Add code of conduct section
```

### Co-authoring Commits

When pair programming or working together:

```
feat(validator): add email validation

Add comprehensive email validation with length checks
and format validation using net/mail.

Co-authored-by: Jane Doe <jane@example.com>
Co-authored-by: Claude Sonnet 4.5 <noreply@anthropic.com>
```

---

## Pull Request Process

### Before Submitting

1. ✅ **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. ✅ **Run tests**:
   ```bash
   go test ./...
   ```

3. ✅ **Format code**:
   ```bash
   go fmt ./...
   golangci-lint run
   ```

4. ✅ **Update documentation** if needed

5. ✅ **Write good commit messages** (see guidelines above)

### Submitting a Pull Request

1. **Push to your fork**:
   ```bash
   git push origin your-feature-branch
   ```

2. **Open a PR** on GitHub with:
   - Clear title following commit message conventions
   - Description explaining what and why
   - Reference to any related issues
   - Screenshots for UI changes

3. **PR Template** (fill this out):

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] All new and existing tests pass locally with my changes
- [ ] I have tested the changes manually

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] Any dependent changes have been merged and published

## Related Issues
Fixes #(issue number)
```

### Review Process

1. **Maintainer review**: A maintainer will review your PR
2. **Address feedback**: Make requested changes
3. **Approval**: Once approved, a maintainer will merge

### After Merging

1. **Delete your branch**:
   ```bash
   git branch -d your-feature-branch
   git push origin --delete your-feature-branch
   ```

2. **Sync your fork**:
   ```bash
   git checkout main
   git pull upstream main
   git push origin main
   ```

---

## Testing Guidelines

### Test Coverage

We aim for **80%+ code coverage** overall, with **100% coverage** for:
- Validation logic
- Error handling
- Security-critical code (CORS, authentication)

### Testing Strategy

#### Unit Tests
Test individual functions in isolation:
- `internal/validator/` - All validation functions
- `internal/errors/` - Error type detection
- `internal/config/` - Configuration loading and validation

#### Integration Tests
Test components working together:
- `internal/store/sqlite/` - Database operations with in-memory SQLite
- `internal/web/` - HTTP handlers using `httptest`

#### End-to-End Tests
Test complete workflows:
- Creating a client, form, and submission
- CORS validation
- Authentication flow

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/validator/

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Verbose output
go test -v ./...

# Generate coverage HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Writing Tests

Use table-driven tests:

```go
func TestFormatLimit(t *testing.T) {
    tests := []struct {
        name  string
        input int
        want  int
    }{
        {"zero", 0, 20},
        {"negative", -5, 20},
        {"positive", 10, 10},
        {"large", 1000, 1000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := formatLimit(tt.input)
            if got != tt.want {
                t.Errorf("formatLimit(%d) = %d, want %d", tt.input, got, tt.want)
            }
        })
    }
}
```

---

## Documentation

### Types of Documentation

1. **Code Comments**
   - Godoc for all exported items
   - Inline comments for complex logic

2. **README.md**
   - Quick start guide
   - Installation instructions
   - Basic usage examples

3. **CONTRIBUTING.md** (this file)
   - Development guidelines
   - How to contribute

4. **Architecture Documentation**
   - See `TEMPLATE_IMPROVEMENTS.md` for example
   - Document major design decisions

### Writing Good Documentation

- **Be clear and concise**
- **Use examples**
- **Keep it up to date**
- **Assume beginner-level Go knowledge**
- **Link to relevant resources**

---

## Community

### Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion (if enabled)

### Recognition

Contributors are recognized in:
- Git commit history
- Release notes
- Special thanks in README (for significant contributions)

---

## Development Roadmap

See [GitHub Issues](https://github.com/OWNER/ticketd/issues) for:
- Planned features
- Known bugs
- Enhancement proposals

### Priority Labels

- `priority: high` - Critical bugs, security issues
- `priority: medium` - Important features, significant bugs
- `priority: low` - Nice-to-have features, minor improvements

### Good First Issues

Look for issues labeled `good first issue` - these are great for new contributors!

---

## Questions?

If you have questions not covered in this guide:
1. Check existing [GitHub Issues](https://github.com/OWNER/ticketd/issues)
2. Open a new issue with the `question` label
3. Be specific and provide context

---

## License

By contributing to TicketD, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to TicketD! 🎉

Small, focused, and honest — that's TicketD.
