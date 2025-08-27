# Contributing to LLMrecon

Thank you for your interest in contributing to LLMrecon! This document provides guidelines and instructions for contributing to the project.

## Branch Protection

The `main` branch is now protected and requires:
- Pull requests for all changes
- At least 1 approving review
- Passing status checks (Go Security Check)
- Resolved conversations before merging
- Up-to-date branches with main

## Development Workflow

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Write clean, well-documented code
   - Follow Go best practices
   - Ensure all tests pass
   - Fix any security issues identified by gosec

3. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: Add your feature description"
   ```

4. **Push to GitHub**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request**
   - Use `gh pr create` or create via GitHub web interface
   - Provide a clear description of changes
   - Reference any related issues
   - Ensure all checks pass

## Commit Message Format

We follow conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Maintenance tasks
- `security:` - Security fixes

## Pull Request Review Process

1. **Automated Checks**
   - Go Security Check must pass
   - Build must succeed
   - Tests must pass

2. **Code Review**
   - At least one maintainer approval required
   - Address all review comments
   - Resolve all conversations

3. **Merging**
   - Squash and merge is preferred
   - Delete feature branch after merge

## Getting Help

- Check existing issues and discussions
- Review the documentation
- Ask questions in pull request comments

Thank you for contributing to LLMrecon!
