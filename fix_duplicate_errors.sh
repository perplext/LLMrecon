#!/bin/bash

echo "Fixing duplicate error handling in database.go..."

# Fix duplicate error checks in database.go
sed -i '' '/^if err != nil {$/N;/^if err != nil {\ntreturn err$/d' src/repository/database.go

# Remove standalone orphaned error returns  
sed -i '' '/^treturn err$/d' src/repository/database.go
sed -i '' '/^return err$/d' src/repository/database.go

# Fix specific patterns where error handling is duplicated
sed -i '' '/^}[[:space:]]*if err != nil {$/,/^}[[:space:]]*$/{/^return err$/d;}' src/repository/database.go

echo "Fixing missing imports in other files..."

# Fix missing imports in bundle/errors
sed -i '' '4a\
import "os"
' src/bundle/errors/categories.go

echo "Fixing missing time import in database.go..."
sed -i '' '3a\
import "time"
' src/repository/database.go

echo "Done!"