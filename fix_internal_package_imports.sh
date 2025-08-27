#!/bin/bash

# Fix internal package imports that cause circular dependencies
echo "Fixing internal package imports..."

# Function to fix imports within a package
fix_package_imports() {
    local package_path=$1
    local package_name=$(basename "$package_path")
    
    echo "Fixing imports in $package_path..."
    
    find "$package_path" -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/perplext/LLMrecon/src/$package_name" "$file"; then
            echo "  Processing $file..."
            
            # Calculate relative path from file to package root
            local file_dir=$(dirname "$file")
            local rel_path=$(realpath --relative-to="$file_dir" "$package_path" 2>/dev/null || echo ".")
            
            # Fix imports within the same package
            sed -i.bak \
                -e "s|\"github.com/perplext/LLMrecon/src/$package_name/|\"./$rel_path/|g" \
                -e "s|\"github.com/perplext/LLMrecon/src/$package_name\"|\"./$rel_path\"|g" \
                "$file"
            
            rm -f "$file.bak"
        fi
    done
}

# Fix specific packages that have internal import issues
packages=(
    "src/reporting"
    "src/bundle"
    "src/provider"
    "src/repository"
    "src/update"
    "src/template/security"
    "src/template/format"
    "src/analytics"
    "src/api"
    "src/auth"
    "src/security"
    "src/notification"
    "src/vulnerability"
    "src/version"
    "src/distribution"
    "src/enterprise"
    "src/performance"
    "src/profiling"
    "src/queue"
    "src/utils"
    "src/customization"
    "src/compliance"
    "src/audit"
    "src/monitoring"
    "src/attacks"
    "src/automated"
)

for package in "${packages[@]}"; do
    if [ -d "$package" ]; then
        # Extract package name from path
        package_name=$(echo "$package" | sed 's|src/||' | sed 's|/.*||')
        
        # Fix imports within this package
        find "$package" -name "*.go" -type f | while read -r file; do
            if grep -q "github.com/perplext/LLMrecon/src/$package_name" "$file"; then
                echo "Processing $file..."
                
                # For subpackages, calculate the correct relative path
                file_dir=$(dirname "$file")
                package_root="src/$package_name"
                
                # Calculate how many levels deep we are
                levels=$(echo "$file_dir" | sed "s|$package_root||" | sed 's|[^/]||g' | wc -c)
                levels=$((levels - 1))
                
                if [ $levels -eq 0 ]; then
                    # We're in the package root
                    relative_prefix="./"
                else
                    # We're in a subdirectory, need to go up
                    relative_prefix=""
                    for i in $(seq 1 $levels); do
                        relative_prefix="../$relative_prefix"
                    done
                fi
                
                # Replace imports
                sed -i.bak \
                    -e "s|\"github.com/perplext/LLMrecon/src/$package_name/|\"${relative_prefix}|g" \
                    -e "s|\"github.com/perplext/LLMrecon/src/$package_name\"|\"${relative_prefix%/}\"|g" \
                    "$file"
                
                rm -f "$file.bak"
            fi
        done
    fi
done

echo "Internal package import fixes complete!"

# Show remaining count
echo "Remaining files with internal import issues:"
find src -name "*.go" -type f -path "*/cmd" -prune -o -name "*.go" -exec grep -l "github.com/perplext/LLMrecon/src" {} \; | wc -l