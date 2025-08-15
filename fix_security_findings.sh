#!/bin/bash

# Script to fix security findings systematically
# This script addresses the most common security issues found by gosec

echo "Starting security findings remediation..."

# Function to fix G404 - Weak random number generator
fix_g404() {
    echo "Fixing G404 - Weak random number generator issues..."
    
    # Replace math/rand with crypto/rand for cryptographic purposes
    find src -name "*.go" -type f -exec grep -l "math/rand" {} \; | while read -r file; do
        if grep -q "crypto" "$file" || grep -q "password" "$file" || grep -q "token" "$file" || grep -q "key" "$file" || grep -q "secret" "$file"; then
            echo "Fixing crypto-related rand usage in $file"
            # Add crypto/rand import if not present
            if ! grep -q "crypto/rand" "$file"; then
                sed -i '' '/import (/a\
	"crypto/rand"
' "$file"
            fi
            
            # Replace specific weak random patterns
            sed -i '' 's/rand\.Intn(/randomInt(/g' "$file"
            sed -i '' 's/rand\.Int63(/randomInt64(/g' "$file"
            sed -i '' 's/rand\.Float64(/randomFloat64(/g' "$file"
        fi
    done
}

# Function to fix G306 - Poor file permissions
fix_g306() {
    echo "Fixing G306 - Poor file permissions..."
    
    # Fix file creation permissions
    find src -name "*.go" -type f -exec grep -l "0644\|0666\|0777" {} \; | while read -r file; do
        echo "Fixing file permissions in $file"
        sed -i '' 's/0644/0600/g' "$file"
        sed -i '' 's/0666/0600/g' "$file"  
        sed -i '' 's/0777/0700/g' "$file"
    done
}

# Function to fix G401 - Weak cryptographic algorithms
fix_g401() {
    echo "Fixing G401 - Weak cryptographic algorithms..."
    
    # Replace MD5 and SHA1 with stronger algorithms
    find src -name "*.go" -type f -exec grep -l "md5\|sha1" {} \; | while read -r file; do
        if grep -q "crypto" "$file"; then
            echo "Fixing weak crypto in $file"
            sed -i '' 's/md5\.New()/sha256.New()/g' "$file"
            sed -i '' 's/sha1\.New()/sha256.New()/g' "$file"
            sed -i '' 's/"crypto\/md5"/"crypto\/sha256"/g' "$file"
            sed -i '' 's/"crypto\/sha1"/"crypto\/sha256"/g' "$file"
        fi
    done
}

# Run fixes
fix_g404
fix_g306  
fix_g401

echo "Security findings remediation completed!"
echo "Note: G104 (unhandled errors) and G304 (file path injection) require manual review and fixing."