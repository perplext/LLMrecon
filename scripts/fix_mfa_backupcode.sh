#!/bin/bash

echo "Fixing BackupCode references to MFABackupCode..."

# Find and replace BackupCode with MFABackupCode in MFA module
# But avoid replacing MFABackupCode or BackupCodeConfig
find /Users/nconsolo/claude-code/LLMrecon/src/security/access/mfa -name "*.go" -type f | while read -r file; do
    # Use word boundaries to avoid partial matches
    sed -i '' 's/\bBackupCode\b/MFABackupCode/g' "$file"
    # Fix any double replacements
    sed -i '' 's/MFAMFABackupCode/MFABackupCode/g' "$file"
    # Fix BackupCodeConfig back to its original name
    sed -i '' 's/MFABackupCodeConfig/BackupCodeConfig/g' "$file"
done

echo "BackupCode references fixed!"