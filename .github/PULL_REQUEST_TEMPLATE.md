## Description

<!-- Provide a clear and concise description of your changes -->

## Type of Change

<!-- Mark the relevant option with an 'x' -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Style/formatting changes (no functional changes)
- [ ] ♻️ Code refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] ✅ Test updates

## Related Issues

<!-- Link to related issues using #issue_number -->

Fixes #(issue)
Relates to #(issue)

## Changes Made

<!-- List the specific changes you made -->

- Change 1
- Change 2
- Change 3

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Manual Testing

- [ ] Tested locally with `go run .`
- [ ] Tested build process
- [ ] Tested on multiple browsers (if UI changes)
- [ ] Tested edge cases

### Automated Testing

- [ ] Added new tests for new functionality
- [ ] Updated existing tests
- [ ] All tests pass locally (`go test ./...`)
- [ ] No race conditions (`go test -race ./...`)

### Test Coverage

<!-- If you added tests, mention the coverage -->

Current coverage: XX%

## Screenshots/Videos

<!-- If applicable, add screenshots or videos to demonstrate the changes -->

### Before

<!-- Screenshots/description of behavior before changes -->

### After

<!-- Screenshots/description of behavior after changes -->

## Checklist

<!-- Mark completed items with an 'x' -->

### Code Quality

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new compiler warnings
- [ ] I have run `go fmt ./...`
- [ ] I have run `go vet ./...`
- [ ] I have run `golangci-lint run` (if available)

### Documentation

- [ ] I have made corresponding changes to the documentation
- [ ] I have updated godoc comments for exported functions
- [ ] I have updated the README if needed
- [ ] I have updated CONTRIBUTING.md if workflow changed

### Testing

- [ ] I have added tests that prove my fix is effective or feature works
- [ ] New and existing unit tests pass locally
- [ ] I have tested the changes manually

### Security

- [ ] My changes don't introduce security vulnerabilities
- [ ] I haven't hardcoded sensitive information
- [ ] I have validated all user inputs
- [ ] I have considered error handling and edge cases

### Dependencies

- [ ] I haven't added new dependencies without discussion
- [ ] If I added dependencies, I've documented why
- [ ] I've run `go mod tidy`

## Performance Impact

<!-- Describe any performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance may be slightly slower (acceptable tradeoff)
- [ ] Needs performance review

## Breaking Changes

<!-- If this is a breaking change, describe what users need to do -->

⚠️ **This PR contains breaking changes:**

- [ ] Database schema changes (migration needed)
- [ ] Configuration changes (users need to update env vars)
- [ ] API changes (handlers or endpoints modified)
- [ ] Template changes (custom templates may break)

**Migration Guide:**

```bash
# Steps users need to take
```

## Additional Notes

<!-- Any additional information for reviewers -->

## Reviewer Checklist

<!-- For maintainers reviewing this PR -->

- [ ] Code follows project conventions
- [ ] Tests are adequate and pass
- [ ] Documentation is clear and complete
- [ ] Commit messages follow conventional commits
- [ ] No sensitive data in code or commits
- [ ] Ready to merge

---

**By submitting this pull request, I confirm that my contribution is made under the terms of the MIT License.**
