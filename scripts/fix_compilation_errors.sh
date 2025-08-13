#!/bin/bash

echo "Fixing compilation errors..."

# Fix unused imports in audit trail
sed -i '' '/^[[:space:]]*"strings"$/d' src/security/access/audit/trail/audit_trail.go
sed -i '' '/^[[:space:]]*"sync"$/d' src/security/access/audit/trail/audit_trail.go
sed -i '' '/github.com\/LLMrecon\/LLMrecon\/src\/security\/access\/common/d' src/security/access/audit/trail/manager.go

# Fix keystore issues - handle privateKey properly in key_export_import.go
# This is complex, so we'll just comment out the problematic lines for now
sed -i '' '240s/privateKey/rsaPrivKey/' src/security/keystore/key_export_import.go
sed -i '' '265s/privateKey/ecdsaPrivKey/' src/security/keystore/key_export_import.go
sed -i '' '290s/privateKey/ed25519PrivKey/' src/security/keystore/key_export_import.go

# Fix privateKey references 
sed -i '' '246s/privateKey/rsaKey/' src/security/keystore/key_export_import.go
sed -i '' '256s/privateKey/rsaKey/' src/security/keystore/key_export_import.go

# Fix the MFA duplicates - comment out duplicates temporarily
sed -i '' '1s/^/\/\/ /' src/security/access/mfa/sms.go || true
sed -i '' '1i\
// NOTE: This file has duplicate type declarations with interface.go and needs refactoring\
' src/security/access/mfa/sms.go || true

sed -i '' '1s/^/\/\/ /' src/security/access/mfa/store.go || true
sed -i '' '1i\
// NOTE: This file has duplicate type declarations with manager.go and needs refactoring\
' src/security/access/mfa/store.go || true

sed -i '' '1s/^/\/\/ /' src/security/access/mfa/totp.go || true
sed -i '' '1i\
// NOTE: This file has duplicate type declarations with interface.go and needs refactoring\
' src/security/access/mfa/totp.go || true

echo "Compilation errors partially fixed. Some files may need manual intervention."