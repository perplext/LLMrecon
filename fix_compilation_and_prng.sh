#!/bin/bash

# Fix compilation errors and weak PRNG instances
echo "Fixing compilation errors and weak PRNG instances..."

# Fix malformed import statements and missing imports
echo "Fixing import statements..."

# Fix src/security/access/mfa/utils.go - fix syntax errors
sed -i '' '
s/secret := os\.Getenv("SECRET_KEY")/secret = strings.TrimRight(secret, "=")/
s/secret := os\.Getenv("SECRET_KEY"))/secret = strings.TrimRight(secret, "=")/
' src/security/access/mfa/utils.go

# Add missing imports to utils.go if not present
if ! grep -q '"os"' src/security/access/mfa/utils.go; then
    sed -i '' '/import (/a\
	"os"\
	"time"
' src/security/access/mfa/utils.go
fi

# Fix src/security/access/mfa/totp.go - fix multiple syntax errors
sed -i '' '
/if _, err := rand\.Read(bytes); err != nil {/,/^}$/{
    s/if err != nil {//
    s/}[\t]*return "", err/    return "", err/
}
s/secret := os\.Getenv("SECRET_KEY"))/secret = strings.TrimRight(secret, "=")/
s/if err != nil {$/for i := -window; i <= window; i++ {/
/^}[\t]*for i := -window; i <= window; i++ {$/d
s/secret := os\.Getenv("SECRET_KEY"))/secret = strings.TrimRight(secret, "=")/
s/if err != nil {$/    secret = secret + strings.Repeat("=", 8-missingPadding)/
/^}[\t]*secret := os\.Getenv("SECRET_KEY"), 8-missingPadding)$/d
' src/security/access/mfa/totp.go

# Add missing imports to totp.go
if ! grep -q '"os"' src/security/access/mfa/totp.go; then
    sed -i '' '/import (/a\
	"os"\
	"time"\
	"crypto/sha1"
' src/security/access/mfa/totp.go
fi

# Fix src/security/access/auth.go - clean up error handling blocks
sed -i '' '
/^if err != nil {$/d
/^}$/N
/^}\nif err != nil {$/d
s/if err := m\.userStore\.UpdateUser(ctx, user); err != nil {/    if err := m.userStore.UpdateUser(ctx, user); err != nil {/
s/}[\t]*return nil, err/        return nil, err/
' src/security/access/auth.go

# Add missing time import to auth.go
if ! grep -q '"time"' src/security/access/auth.go; then
    sed -i '' '/import (/a\
	"time"\
	"os"
' src/security/access/auth.go
fi

# Fix secret handling in auth.go
sed -i '' '
s/if secret := os\.Getenv("SECRET_KEY") {/if secret == "" {/
s/secret := os\.Getenv("SECRET_KEY"))/secret = strings.TrimRight(secret, "=")/
' src/security/access/auth.go

# Fix src/bundle/schema.go - clean up error handling and missing imports
sed -i '' '
/^if err != nil {$/d
/^}$/N
/^}\nif err != nil {$/d
s/if _, err := os\.Stat(schemaPath); os\.IsNotExist(err) {/    if _, err := os.Stat(schemaPath); os.IsNotExist(err) {/
s/if strings\.Contains(path, "\.\.") {/    if strings.Contains(schemaPath, "..") {/
s/return "", fmt\.Errorf("path traversal detected")/        return nil, fmt.Errorf("path traversal detected")/
s/path = filepath\.Clean(path)/    schemaPath = filepath.Clean(schemaPath)/
' src/bundle/schema.go

# Add missing imports to schema.go
if ! grep -q '"os"' src/bundle/schema.go; then
    sed -i '' '/import (/a\
	"os"\
	"path/filepath"\
	"time"
' src/bundle/schema.go
fi

# Fix weak PRNG instances - replace math/rand with crypto/rand in security contexts
echo "Fixing weak PRNG instances..."

find src -name "*.go" -type f -exec grep -l "math/rand" {} \; | while read file; do
    echo "Fixing PRNG in $file"
    
    # Replace math/rand import with crypto/rand
    sed -i '' 's/"math\/rand"/"crypto\/rand"/g' "$file"
    
    # Replace rand.Intn with crypto rand equivalent for security contexts
    sed -i '' '
    s/rand\.Intn(\([^)]*\))/randInt(\1)/g
    s/rand\.Int63n(\([^)]*\))/randInt64(\1)/g
    s/rand\.Float64()/randFloat64()/g
    ' "$file"
    
    # Add helper functions if they don't exist
    if ! grep -q "func randInt" "$file"; then
        cat >> "$file" << 'EOF'

// Secure random number generation helpers
func randInt(max int) int {
    n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
    if err != nil {
        panic(err)
    }
    return int(n.Int64())
}

func randInt64(max int64) int64 {
    n, err := rand.Int(rand.Reader, big.NewInt(max))
    if err != nil {
        panic(err)
    }
    return n.Int64()
}

func randFloat64() float64 {
    bytes := make([]byte, 8)
    rand.Read(bytes)
    return float64(binary.BigEndian.Uint64(bytes)) / float64(1<<64)
}
EOF
    fi
    
    # Add required imports
    if ! grep -q '"math/big"' "$file"; then
        sed -i '' '/import (/a\
	"math/big"\
	"encoding/binary"
' "$file"
    fi
done

echo "Compilation and PRNG fixes applied!"
echo "Summary:"
echo "- Fixed syntax errors in MFA files"
echo "- Fixed import statements and missing imports"
echo "- Replaced math/rand with crypto/rand in security contexts"
echo "- Added secure random number generation helpers"
echo "- Fixed path traversal validation"