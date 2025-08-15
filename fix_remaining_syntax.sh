#!/bin/bash

# Fix remaining syntax errors in auth.go and other files
echo "Fixing remaining syntax errors..."

# Fix auth.go remaining issues
sed -i '' '
s/^}[\t]*}$/}/g
s/^}[\t]*\([^}]\)/\1/g
s/}[\t]*\(.*UserAgent.*\)/\t\t\1/g
s/}[\t]*ctx = WithMFAStatus/\tctx = WithMFAStatus/g
s/}[\t]*\/\/ Enforce boundaries/\t\/\/ Enforce boundaries/g
s/}[\t]*if err := m\.boundaryEnforcer/\tif err := m.boundaryEnforcer/g
s/}[\t]*m\.auditLogger\.LogAudit/\tm.auditLogger.LogAudit/g
s/}[\t]*session\.LastActivity/\tsession.LastActivity/g
s/}[\t]*\/\/ Check if username/\t\/\/ Check if username/g
s/}[\t]*if _, err := m\.userStore\.GetUserByUsername/\tif _, err := m.userStore.GetUserByUsername/g
s/}[\t]*if _, err := m\.userStore\.GetUserByEmail/\tif _, err := m.userStore.GetUserByEmail/g
s/}[\t]*\/\/ Create user/\t\/\/ Create user/g
s/}[\t]*user := &User/\tuser := \&User/g
s/}[\t]*\/\/ Get user/\t\/\/ Get user/g
s/}[\t]*user, err := m\.userStore\.GetUserByID/\tuser, err := m.userStore.GetUserByID/g
s/}[\t]*\/\/ Update user/\t\/\/ Update user/g
s/}[\t]*user\.PasswordHash/\tuser.PasswordHash/g
s/}[\t]*\/\/ Invalidate all sessions/\t\/\/ Invalidate all sessions/g
s/}[\t]*\/\/ Check if MFA is already enabled/\t\/\/ Check if MFA is already enabled/g
s/}[\t]*\/\/ Add method/\t\t\/\/ Add method/g
s/}[\t]*\/\/ Update user/\t\/\/ Update user/g
s/}[\t]*if policy\.RequireUppercase/\tif policy.RequireUppercase/g
s/}[\t]*if !hasUpper/\t\tif !hasUpper/g
s/}[\t]*\/\/ Enhanced MFA verification/\t\/\/ Enhanced MFA verification/g
s/}[\t]*\/\/ Check uppercase/\t\/\/ Check uppercase/g
s/[\t]*tokenExpiration = m\.config\.SessionPolicy\.TokenExpiration/\t\ttokenExpiration = m.config.SessionPolicy.TokenExpiration/g
s/[\t]*UserID:      user\.ID,/\t\tUserID:      user.ID,/g
/^\t\t\treturn nil, err$/d
' src/security/access/auth.go

# Fix totp.go syntax errors
sed -i '' '
s/^for i := -window; i <= window; i++ {$/\tfor i := -window; i <= window; i++ {/
s/^}[\t]*for i := -window; i <= window; i++ {$/\t\tfor i := -window; i <= window; i++ {/
/^for i := -window; i <= window; i++ {$/d
s/secret := os\.Getenv("SECRET_KEY"), 8-missingPadding)/secret = secret + strings.Repeat("=", 8-missingPadding)/
s/secret = strings\.TrimRight(secret, "=")/secretToUse := config.Secret\
\tsecret := strings.TrimRight(secretToUse, "=")/
s/\t\tfor i := -window; i <= window; i++ {/\t\tif err != nil {/
s/\t\t\tcontinue/\t\t\treturn "", err/
' src/security/access/mfa/totp.go

# Fix schema.go syntax errors  
sed -i '' '
s/^}[\t]*\/\/ ValidateManifestFile/\/\/ ValidateManifestFile/
s/^}[\t]*if !result\.Valid()/\tif !result.Valid()/
s/^}[\t]*\/\/ Convert the manifest/\t\/\/ Convert the manifest/
s/^}[\t]*manifestJSON, err := json\.Marshal/\tmanifestJSON, err := json.Marshal/
s/^}[\t]*\/\/ Get the executable/\t\/\/ Get the executable/
s/^}}[\t]*schemaPath := filepath\.Join/\tschemaPath := filepath.Join/
s/^}[\t]*\/\/ Try to find the schema/\t\t\/\/ Try to find the schema/
s/^}[\t]*\/\/ If the manifest is not valid/\t\/\/ If the manifest is not valid/
s/^}}$/}/
s/^}[\t]*Manifest:/\t\t\tManifest:/
s/^}[\t]*Checksum:/\t\t\t\tChecksum:/
' src/bundle/schema.go

echo "Syntax fixes applied!"