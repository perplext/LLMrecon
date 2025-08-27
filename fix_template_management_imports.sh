#!/bin/bash

# Fix imports in template/management files that incorrectly reference security/access/interfaces
# These should use template/management/interfaces instead

echo "Fixing template management imports..."

# List of files to fix
files=(
    "src/template/management/main.go"
    "src/template/management/manager.go"
    "src/template/management/schema.go"
    "src/template/management/execution/parallel_pipeline.go"
    "src/template/management/execution/engine.go"
    "src/template/management/execution/async_executor.go"
    "src/template/management/execution/detection.go"
    "src/template/management/execution/optimized_engine.go"
    "src/template/management/loader/resource_efficient_loader.go"
    "src/template/management/loader/loader.go"
    "src/template/management/loader/optimized_loader.go"
    "src/template/management/management.go"
    "src/template/management/streaming/stream_processor.go"
    "src/template/management/optimized_manager.go"
    "src/template/management/reporting/reporter.go"
    "src/template/management/benchmark/benchmark.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "Fixing $file..."
        sed -i '' 's|github.com/perplext/LLMrecon/src/security/access/interfaces|github.com/perplext/LLMrecon/src/template/management/interfaces|g' "$file"
    fi
done

echo "Done fixing template management imports."