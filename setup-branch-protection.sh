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
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["build", "test", "lint"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Branch protection rules applied successfully!${NC}"
else
    echo -e "${RED}✗ Failed to apply branch protection rules${NC}"
    exit 1
fi

# Verify the protection
echo -e "${YELLOW}Verifying branch protection...${NC}"
gh api "/repos/${REPO_OWNER}/${REPO_NAME}/branches/${BRANCH}/protection" --jq '{
  required_status_checks: .required_status_checks.contexts,
  required_reviews: .required_pull_request_reviews.required_approving_review_count,
  enforce_admins: .enforce_admins.enabled,
  restrictions: .restrictions,
  allow_force_pushes: .allow_force_pushes.enabled,
  allow_deletions: .allow_deletions.enabled
}'

echo -e "${GREEN}Branch protection setup complete!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. All future changes must be made through Pull Requests"
echo "2. Create feature branches for new work: git checkout -b feature/your-feature"
echo "3. Push changes and create PR: gh pr create"
echo "4. Get approval before merging"