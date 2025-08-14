#!/bin/bash

# Files with malformed imports
files=(
  "src/api/scan/service.go"
  "src/provider/middleware/circuit_breaker.go"
  "src/provider/config/config.go"
  "src/security/communication/cert_chain_validation.go"
  "src/security/prompt/advanced_template_monitor.go"
  "src/bundle/errors/enhanced_handler.go"
  "src/security/access/audit/trail/audit_trail.go"
  "src/audit/audit.go"
)

for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    echo "Fixing $file"
    # Replace tab-separated imports on same line with newline-separated
    sed -i '' 's/\t"time"\t/\
\t"time"\
\t/g' "$file"
  fi
done

echo "Fixed malformed imports"