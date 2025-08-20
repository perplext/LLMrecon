#!/bin/bash

echo "Fixing missing closing braces in Go files..."

# Fix src/version/analyzer.go
echo "Fixing src/version/analyzer.go..."
sed -i.bak '188s/$/}/' src/version/analyzer.go

# Fix src/template/management/types/interfaces.go
echo "Fixing src/template/management/types/interfaces.go..."
# Need to check this file for struct/interface definitions

# Fix src/template/management/validation/input_validator.go
echo "Fixing src/template/management/validation/input_validator.go..."
# Need to check for missing braces in struct

echo "Running go vet to check progress..."
go vet ./src/version/... 2>&1 | head -10