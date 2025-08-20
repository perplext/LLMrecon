#!/bin/bash

echo "Comprehensive syntax fix for Go files..."

# Function to add missing closing brace after a return statement before a comment
fix_missing_brace_after_return() {
    local file=$1
    echo "Processing $file..."
    
    # Pattern: return statement followed by blank line and comment starting with //
    # Add closing brace between them
    sed -i.bak -E '
        /^[[:space:]]*return.*[^}]$/ {
            N
            /\n$/ {
                N
                /\n\/\// {
                    s/(\n)(\/\/)/\1}\n\2/
                }
            }
        }
    ' "$file"
}

# Fix src/version package files
echo "=== Fixing src/version package ==="

# src/version/analyzer.go - already partially fixed, check for more
fix_missing_brace_after_return "src/version/analyzer.go"

# Check line 244 issue
sed -i.bak '244s/^/}/' src/version/analyzer.go 2>/dev/null || true

# src/version/analyzer_helpers.go
echo "Fixing src/version/analyzer_helpers.go..."
# Line 40 issue - missing closing brace
sed -i.bak '39a\
}' src/version/analyzer_helpers.go

# src/version/constants.go
echo "Fixing src/version/constants.go..."
# Line 26 issue
sed -i.bak '25a\
}' src/version/constants.go

# src/version/dependency.go
echo "Fixing src/version/dependency.go..."
# Multiple issues - check structure
cat > /tmp/fix_dependency.sed << 'EOF'
# Fix line 62 - missing closing brace
61a\
}
# Fix line 77 - missing closing brace  
76a\
}
# Fix line 82 - missing closing brace
81a\
}
# Fix line 111 - missing closing brace
110a\
}
EOF
sed -i.bak -f /tmp/fix_dependency.sed src/version/dependency.go

# Fix src/template/management/types/interfaces.go
echo "=== Fixing src/template/management/types ==="
echo "Fixing src/template/management/types/interfaces.go..."
# Line 19 - missing closing brace for struct
sed -i.bak '18a\
}' src/template/management/types/interfaces.go

# Fix src/template/management/types/types.go
echo "Fixing src/template/management/types/types.go..."
# EOF issue - missing closing brace
echo "}" >> src/template/management/types/types.go

# Fix src/template/management/validation/input_validator.go
echo "=== Fixing src/template/management/validation ==="
echo "Fixing src/template/management/validation/input_validator.go..."
# Line 27 - missing closing brace for struct
sed -i.bak '26a\
}' src/template/management/validation/input_validator.go

# Fix src/template/management/parser/parser.go
echo "=== Fixing src/template/management/parser ==="
echo "Fixing src/template/management/parser/parser.go..."
# Line 22 - missing closing brace for struct
sed -i.bak '21a\
}' src/template/management/parser/parser.go

# Fix src/template/management/registry files
echo "=== Fixing src/template/management/registry ==="
echo "Fixing src/template/management/registry/adapter.go..."
# Multiple missing braces
sed -i.bak '60a\
}' src/template/management/registry/adapter.go
sed -i.bak '72a\
}' src/template/management/registry/adapter.go

echo "Fixing src/template/management/registry/registry.go..."
sed -i.bak '22a\
}' src/template/management/registry/registry.go
sed -i.bak '37a\
}' src/template/management/registry/registry.go

# Fix src/security/api files
echo "=== Fixing src/security/api ==="
echo "Fixing src/security/api/anomaly_detection.go..."
sed -i.bak '72a\
}' src/security/api/anomaly_detection.go

echo "Fixing src/security/api/ip_allowlist.go..."
sed -i.bak '47a\
}' src/security/api/ip_allowlist.go

# Fix src/security/prompt files
echo "=== Fixing src/security/prompt ==="
echo "Fixing src/security/prompt/advanced_jailbreak_detector.go..."
echo "}" >> src/security/prompt/advanced_jailbreak_detector.go

echo "Fixing src/security/prompt/advanced_template_monitor.go..."
sed -i.bak '30a\
}' src/security/prompt/advanced_template_monitor.go
sed -i.bak '82a\
}' src/security/prompt/advanced_template_monitor.go
sed -i.bak '89a\
}' src/security/prompt/advanced_template_monitor.go

# Fix src/provider/core files
echo "=== Fixing src/provider/core ==="
echo "Fixing src/provider/core/connection_pool.go..."
sed -i.bak '467a\
}' src/provider/core/connection_pool.go

echo "Fixing src/provider/core/factory.go..."
sed -i.bak '16a\
}' src/provider/core/factory.go

# Clean up backup files
find src -name "*.bak" -delete

echo "Running go vet to check progress..."
go vet ./src/version/... 2>&1 | head -20

echo "Script completed. Check output above for remaining errors."