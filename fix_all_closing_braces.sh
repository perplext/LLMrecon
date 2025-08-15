#!/bin/bash

echo "Fixing all extra closing braces systematically..."

# Fix repository/interfaces/repository.go
sed -i '' '69d' src/repository/interfaces/repository.go

# Fix template/management/interfaces files
echo "}" >> src/template/management/interfaces/interfaces.go

# Remove extra braces from specific lines in various files
sed -i '' '30s/^}//' src/template/management/interfaces/loader.go
sed -i '' '51s/^}//' src/template/management/interfaces/loader.go
sed -i '' '12s/^}//' src/template/management/interfaces/validator.go
sed -i '' '39s/^}//' src/template/management/interfaces/validator.go
sed -i '' '45s/^}//' src/template/management/interfaces/validator.go

# Fix bundle/errors files
sed -i '' '26s/^}//' src/bundle/errors/audit.go
sed -i '' '24s/^}//' src/bundle/errors/categories.go
sed -i '' '22s/^}//' src/bundle/errors/enhanced_handler.go
sed -i '' '76s/^}//' src/bundle/errors/enhanced_handler.go
sed -i '' '87s/^}//' src/bundle/errors/enhanced_handler.go
sed -i '' '105s/^}//' src/bundle/errors/enhanced_handler.go
sed -i '' '115s/^}//' src/bundle/errors/enhanced_handler.go
sed -i '' '106s/^}//' src/bundle/errors/errors.go
sed -i '' '30s/^}//' src/bundle/errors/recovery.go
sed -i '' '137s/^}//' src/bundle/errors/recovery.go

# Fix security/access/audit/trail files
sed -i '' '18s/^}//' src/security/access/audit/trail/audit_trail.go
sed -i '' '32s/^}//' src/security/access/audit/trail/audit_trail.go

echo "Testing compilation..."
go build -o /tmp/test ./src/main.go 2>&1 | grep -c "syntax error" || echo "0"
