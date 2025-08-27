#!/bin/bash

echo "Comprehensive Go syntax fix script..."
echo "This will fix common patterns across all Go files"

# Function to fix common syntax issues
fix_file() {
    local file="$1"
    if [ ! -f "$file" ]; then
        return
    fi
    
    # Create a temporary file
    local tmpfile="${file}.tmp"
    
    # Process the file line by line
    local in_struct=0
    local in_func=0
    local brace_count=0
    local prev_line=""
    local skip_next=0
    
    while IFS= read -r line || [ -n "$line" ]; do
        if [ "$skip_next" -eq 1 ]; then
            skip_next=0
            continue
        fi
        
        # Check for struct/interface/type definitions missing closing braces
        if [[ "$line" =~ ^[[:space:]]*type[[:space:]] ]]; then
            in_struct=1
            brace_count=0
        fi
        
        # Count braces
        if [ "$in_struct" -eq 1 ]; then
            brace_count=$((brace_count + $(echo "$line" | tr -cd '{' | wc -c)))
            brace_count=$((brace_count - $(echo "$line" | tr -cd '}' | wc -c)))
            
            if [ "$brace_count" -eq 0 ] && [[ "$line" =~ \} ]]; then
                in_struct=0
            fi
        fi
        
        # Fix patterns like "}func" or "}type" 
        if [[ "$line" =~ ^}(func|type|var|const) ]]; then
            echo "}"
            echo ""
            echo "${line:1}" # Remove the leading }
            continue
        fi
        
        # Fix pattern where closing brace is on wrong line
        if [[ "$prev_line" =~ ^[[:space:]]*$ ]] && [[ "$line" =~ ^}[[:space:]]*$ ]]; then
            # Check next line
            IFS= read -r next_line || next_line=""
            if [[ "$next_line" =~ ^(func|type|var|const|//) ]]; then
                echo "$line"
                echo ""
                echo "$next_line"
                skip_next=1
                prev_line="$next_line"
                continue
            else
                echo "$line"
                if [ -n "$next_line" ]; then
                    echo "$next_line"
                    skip_next=1
                fi
            fi
        else
            echo "$line"
        fi
        
        prev_line="$line"
    done < "$file" > "$tmpfile"
    
    # Move temp file back
    mv "$tmpfile" "$file"
}

# Find all Go files with syntax errors and fix them
echo "Finding and fixing Go files with syntax errors..."

# Priority files that are commonly broken
priority_files=(
    "src/version/analyzer_helpers.go"
    "src/version/constants.go" 
    "src/version/dependency.go"
    "src/template/management/types/interfaces.go"
    "src/template/management/types/types.go"
    "src/template/management/validation/input_validator.go"
    "src/template/management/parser/parser.go"
    "src/template/management/registry/adapter.go"
    "src/template/management/registry/registry.go"
)

for file in "${priority_files[@]}"; do
    if [ -f "$file" ]; then
        echo "Fixing $file..."
        fix_file "$file"
    fi
done

# Run go vet to see remaining issues
echo ""
echo "Checking remaining syntax errors..."
go vet ./src/version/*.go 2>&1 | head -20

echo ""
echo "Script completed. Run 'go vet ./...' to see all remaining issues."