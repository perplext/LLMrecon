# Branch Protection Setup Guide

This guide helps you configure branch protection rules for the LLMrecon repository to enforce code review through Pull Requests.

## Quick Setup via GitHub Web Interface

1. Navigate to your repository: https://github.com/perplext/LLMrecon
2. Go to **Settings** → **Branches**
3. Click **Add rule** under "Branch protection rules"
4. Configure the following settings:

### Branch name pattern
- Enter: `main`

### Protection Settings

#### ✅ Require a pull request before merging
- [x] **Require approvals**: 1 (or more for stricter control)
- [x] **Dismiss stale pull request approvals when new commits are pushed**
- [x] **Require review from CODEOWNERS** (if you have a CODEOWNERS file)

#### ✅ Require status checks to pass before merging
- [x] **Require branches to be up to date before merging**
- Select required status checks:
  - `build`
  - `test`
  - `lint`

#### ✅ Require conversation resolution before merging
- [x] Enable this to ensure all PR comments are addressed

#### ✅ Require signed commits (optional but recommended)
- [x] Enable for enhanced security

#### ✅ Include administrators
- [x] Enable to apply rules even to repository administrators

#### ✅ Restrict who can push to matching branches
- Add specific users or teams who can push directly (for emergencies)

### Additional Recommended Settings

#### ✅ Require linear history
- [x] Prevent merge commits for a cleaner history

#### ✅ Require deployments to succeed before merging
- Configure if you have deployment workflows

#### ❌ Allow force pushes
- Keep this disabled to preserve history

#### ❌ Allow deletions
- Keep this disabled to prevent accidental branch deletion

## Automated Setup via GitHub CLI

```bash
# Install GitHub CLI if not already installed
# macOS: brew install gh
# Linux: See https://github.com/cli/cli/blob/trunk/docs/install_linux.md

# Authenticate
gh auth login

# Set branch protection rules
gh api repos/perplext/LLMrecon/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["build","test"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  --field restrictions=null \
  --field allow_force_pushes=false \
  --field allow_deletions=false \
  --field required_conversation_resolution=true
```

## Setup Script

Save this as `setup-branch-protection.sh`:

```bash
#!/bin/bash

# Configuration
REPO_OWNER="perplext"
REPO_NAME="LLMrecon"
BRANCH="main"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Setting up branch protection for ${REPO_OWNER}/${REPO_NAME} - ${BRANCH} branch${NC}"

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}GitHub CLI (gh) is not installed. Please install it first.${NC}"
    echo "Visit: https://cli.github.com/manual/installation"
    exit 1
fi

# Check authentication
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}Not authenticated. Running 'gh auth login'...${NC}"
    gh auth login
fi

# Apply branch protection rules
echo -e "${GREEN}Applying branch protection rules...${NC}"

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/${REPO_OWNER}/${REPO_NAME}/branches/${BRANCH}/protection" \
  -f required_status_checks='{"strict":true,"contexts":["build","test","lint"]}' \
  -f enforce_admins=true \
  -f required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true,"require_code_owner_reviews":false,"required_approving_review_count":1}' \
  -f restrictions=null \
  -f allow_force_pushes=false \
  -f allow_deletions=false \
  -f required_conversation_resolution=true \
  -f lock_branch=false \
  -f allow_fork_syncing=true

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Branch protection rules applied successfully!${NC}"
else
    echo -e "${RED}✗ Failed to apply branch protection rules${NC}"
    exit 1
fi

# Verify the protection
echo -e "${YELLOW}Verifying branch protection...${NC}"
gh api "/repos/${REPO_OWNER}/${REPO_NAME}/branches/${BRANCH}/protection" | jq '.'

echo -e "${GREEN}Branch protection setup complete!${NC}"
```

Make it executable:
```bash
chmod +x setup-branch-protection.sh
```

## Best Practices for Working with Protected Branches

### 1. Create Feature Branches
```bash
# Create a new feature branch
git checkout -b feature/your-feature-name

# Make your changes
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/your-feature-name
```

### 2. Create Pull Requests
```bash
# Using GitHub CLI
gh pr create --title "feat: add new feature" --body "Description of changes"

# Or use the GitHub web interface
```

### 3. Emergency Bypass (Admin Only)
If you need to push directly in an emergency:
1. Temporarily disable branch protection
2. Make the critical fix
3. Re-enable protection immediately

## Recommended Workflow

1. **Never commit directly to main**
2. **Always create feature branches**:
   - `feature/` - New features
   - `fix/` - Bug fixes
   - `docs/` - Documentation updates
   - `refactor/` - Code refactoring
   - `test/` - Test additions/updates

3. **Pull Request Process**:
   - Create descriptive PR titles
   - Add detailed descriptions
   - Link related issues
   - Request reviews from relevant team members
   - Address all feedback
   - Ensure all checks pass

4. **Commit Message Convention**:
   ```
   type(scope): subject
   
   body
   
   footer
   ```
   
   Types: feat, fix, docs, style, refactor, test, chore

## GitHub Actions for Status Checks

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  pull_request:
    branches: [ main ]
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    - name: Build
      run: go build -v ./...

  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    - name: Test
      run: go test -v ./...

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

## Troubleshooting

### Common Issues

1. **"Protected branch update failed"**
   - Ensure you have admin permissions
   - Check if organization settings override repository settings

2. **"Required status check not found"**
   - Status checks must run at least once before they can be required
   - Create a dummy PR to trigger the checks

3. **"Cannot push to protected branch"**
   - This is expected! Create a feature branch instead
   - Use `git checkout -b feature/my-feature`

### Useful Commands

```bash
# Check current branch protection
gh api repos/perplext/LLMrecon/branches/main/protection

# List all branches
gh api repos/perplext/LLMrecon/branches

# Remove branch protection (emergency only!)
gh api repos/perplext/LLMrecon/branches/main/protection --method DELETE
```

## Additional Security Recommendations

1. **Enable Dependabot**:
   - Go to Settings → Security & analysis
   - Enable Dependabot alerts and security updates

2. **Set up CODEOWNERS**:
   Create `.github/CODEOWNERS`:
   ```
   # Global owners
   * @perplext
   
   # ML components
   /ml/ @perplext @ml-team
   
   # Security components
   /src/security/ @security-team
   ```

3. **Configure Security Policies**:
   Already have `SECURITY.md` ✓

4. **Enable Secret Scanning**:
   - Settings → Security & analysis → Enable secret scanning

---

After setting up branch protection, all code changes will require:
1. Creating a feature branch
2. Making changes
3. Creating a Pull Request
4. Getting approval
5. Passing all status checks
6. Merging to main

This ensures code quality, enables collaboration, and maintains a clean history!