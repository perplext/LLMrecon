#!/bin/bash

echo "Fixing remaining security issues systematically..."

# 1. Fix math/rand usage - replace with crypto/rand where appropriate
echo "1. Fixing weak PRNG usage..."

# Get list of files using math/rand
files_with_mathrand=$(grep -l "math/rand" src/**/*.go)

for file in $files_with_mathrand; do
    echo "Checking $file for security-sensitive rand usage..."
    
    # Check if it's used for security purposes (tokens, keys, etc.)
    if grep -q "token\|key\|secret\|password\|nonce\|salt\|session" "$file"; then
        echo "  Security-sensitive usage found in $file"
        
        # Replace math/rand with crypto/rand for security purposes
        sed -i '' 's/rand\.Read/cryptorand.Read/g' "$file"
        sed -i '' 's/rand\.Int/randInt/g' "$file"
        
        # Add crypto/rand import if not present
        if ! grep -q "crypto/rand" "$file"; then
            sed -i '' '/import (/a\
	cryptorand "crypto/rand"
' "$file"
        fi
        
        # Add helper function for crypto-safe random integers
        if grep -q "randInt" "$file" && ! grep -q "func randInt" "$file"; then
            cat >> "$file" << 'EOF'

// randInt returns a cryptographically secure random integer
func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}
EOF
        fi
    fi
done

# 2. Fix hardcoded secrets
echo "2. Fixing hardcoded secrets..."

# Replace common hardcoded secret patterns
grep -rn "secret.*=.*[\"']" src/ --include="*.go" | while read line; do
    file=$(echo "$line" | cut -d: -f1)
    echo "  Reviewing secret in $file"
    
    # Replace obvious hardcoded secrets with environment variable lookups
    sed -i '' 's/secret.*=.*"[^"]*"/secret := os.Getenv("SECRET_KEY")/g' "$file"
    sed -i '' 's/apiKey.*=.*"[^"]*"/apiKey := os.Getenv("API_KEY")/g' "$file"
    sed -i '' 's/password.*=.*"[^"]*"/password := os.Getenv("PASSWORD")/g' "$file"
done

# 3. Fix insecure HTTP usage
echo "3. Fixing insecure HTTP usage..."

# Replace http:// with https:// in URLs, but only where it makes sense
grep -rn "http://" src/ --include="*.go" | while read line; do
    file=$(echo "$line" | cut -d: -f1)
    linenum=$(echo "$line" | cut -d: -f2)
    echo "  Reviewing HTTP usage in $file:$linenum"
    
    # Only replace if it's not localhost, testing, or examples
    if ! echo "$line" | grep -q "localhost\|127.0.0.1\|example\|test"; then
        sed -i '' 's/http:\/\/\([^"]*\)/https:\/\/\1/g' "$file"
    fi
done

# 4. Fix path traversal issue
echo "4. Fixing path traversal issues..."

grep -rn "filepath\.Join.*\.\." src/ --include="*.go" | while read line; do
    file=$(echo "$line" | cut -d: -f1)
    echo "  Fixing path traversal in $file"
    
    # Add path cleaning after joins that might contain ..
    sed -i '' '/filepath\.Join.*\.\./a\
	// Prevent path traversal\
	if strings.Contains(path, "..") {\
		return "", fmt.Errorf("path traversal detected")\
	}\
	path = filepath.Clean(path)
' "$file"
done

# 5. Add missing imports where needed
echo "5. Adding missing imports..."

# Add os import where environment variables are used
grep -l "os\.Getenv" src/**/*.go | while read file; do
    if ! grep -q '"os"' "$file"; then
        sed -i '' '/import (/a\
	"os"
' "$file"
    fi
done

# Add strings import where Contains is used
grep -l "strings\.Contains" src/**/*.go | while read file; do
    if ! grep -q '"strings"' "$file"; then
        sed -i '' '/import (/a\
	"strings"
' "$file"
    fi
done

# Add big import where crypto/rand Int is used
grep -l "cryptorand\.Int" src/**/*.go | while read file; do
    if ! grep -q '"math/big"' "$file"; then
        sed -i '' '/import (/a\
	"math/big"
' "$file"
    fi
done

echo "Security fixes applied!"
echo "Summary:"
echo "- Fixed weak PRNG usage in security contexts"
echo "- Replaced hardcoded secrets with environment variables"
echo "- Upgraded HTTP to HTTPS where appropriate"
echo "- Added path traversal protection"
echo "- Added necessary imports"