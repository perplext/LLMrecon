#!/bin/bash

echo "Fixing syntax errors in Go files..."

# Fix src/version/analyzer.go - remove extra braces at end
echo "Fixing src/version/analyzer.go..."
if [ -f "src/version/analyzer.go" ]; then
    # Remove lines 243-246 which have extra closing braces
    head -n 242 src/version/analyzer.go > src/version/analyzer.go.tmp
    echo "	" >> src/version/analyzer.go.tmp
    echo "	return files, nil" >> src/version/analyzer.go.tmp
    echo "}" >> src/version/analyzer.go.tmp
    mv src/version/analyzer.go.tmp src/version/analyzer.go
fi

# Fix src/version/analyzer_helpers.go - malformed function
echo "Fixing src/version/analyzer_helpers.go..."
if [ -f "src/version/analyzer_helpers.go" ]; then
    sed -i.bak 's/^}$//; s/^}func DiffFiles/func DiffFiles/' src/version/analyzer_helpers.go
    # Remove line 38 which has just a closing brace
    sed -i.bak '38d' src/version/analyzer_helpers.go
fi

# Fix src/version/constants.go - malformed function
echo "Fixing src/version/constants.go..."
if [ -f "src/version/constants.go" ]; then
    sed -i.bak 's/^}$//; s/^}func IsPluginVersionCompatible/func IsPluginVersionCompatible/' src/version/constants.go
    # Remove line 24 which has just a closing brace
    sed -i.bak '24d' src/version/constants.go
fi

# Fix src/version/dependency.go - multiple malformed functions
echo "Fixing src/version/dependency.go..."
if [ -f "src/version/dependency.go" ]; then
    # Fix WithMaxVersion function
    sed -i.bak '60,62d' src/version/dependency.go
    sed -i.bak '59a\
}' src/version/dependency.go
    
    # Fix WithPath function
    sed -i.bak '75,77d' src/version/dependency.go
    sed -i.bak '74a\
}' src/version/dependency.go
    
    # Fix IsCompatible function
    sed -i.bak '80,82d' src/version/dependency.go
    sed -i.bak '79a\
}' src/version/dependency.go
    
    # Fix NewDependencyGraph function
    sed -i.bak '108,111d' src/version/dependency.go
    sed -i.bak '107a\
}' src/version/dependency.go
    
    # Add closing brace for NewDependencyGraph
    sed -i.bak '116a\
}' src/version/dependency.go
fi

# Fix template files
echo "Fixing template management files..."

# Fix interfaces.go
if [ -f "src/template/management/types/interfaces.go" ]; then
    sed -i.bak 's/^}$//; s/^}type TemplateLoader/}\
\
type TemplateLoader/' src/template/management/types/interfaces.go
    # Add closing brace after LoadTemplates
    sed -i.bak '24a\
}' src/template/management/types/interfaces.go
fi

# Fix types.go
if [ -f "src/template/management/types/types.go" ]; then
    # Remove line 69 and add proper closing
    sed -i.bak '69d' src/template/management/types/types.go
    echo "}" >> src/template/management/types/types.go
fi

# Fix input_validator.go
if [ -f "src/template/management/validation/input_validator.go" ]; then
    sed -i.bak 's/^}$//; s/^}func NewInputValidator/}\
\
func NewInputValidator/' src/template/management/validation/input_validator.go
fi

# Fix parser.go
if [ -f "src/template/management/parser/parser.go" ]; then
    sed -i.bak 's/^}$//; s/^}func NewTemplateParser/}\
\
func NewTemplateParser/' src/template/management/parser/parser.go
fi

# Fix registry files
if [ -f "src/template/management/registry/adapter.go" ]; then
    sed -i.bak 's/^}func (a \*RegistryAdapter) CreateTemplate/}\
\
func (a *RegistryAdapter) CreateTemplate/' src/template/management/registry/adapter.go
    sed -i.bak 's/^}func (a \*RegistryAdapter) DeleteTemplate/}\
\
func (a *RegistryAdapter) DeleteTemplate/' src/template/management/registry/adapter.go
    # Remove extra closing brace at line 80
    sed -i.bak '80d' src/template/management/registry/adapter.go
fi

if [ -f "src/template/management/registry/registry.go" ]; then
    sed -i.bak 's/^}type TemplateMetadata/}\
\
type TemplateMetadata/' src/template/management/registry/registry.go
    sed -i.bak 's/^}func NewTemplateRegistry/}\
\
func NewTemplateRegistry/' src/template/management/registry/registry.go
    # Add closing brace after return statement
    sed -i.bak '41a\
}' src/template/management/registry/registry.go
fi

# Clean up backup files
find src -name "*.bak" -delete

echo "Running go vet to check progress..."
go vet ./src/version/... 2>&1 | head -20

echo "Script completed."