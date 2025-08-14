#!/bin/bash

# Files with duplicate imports
files=(
  "src/bundle/errors/audit.go"
  "src/security/prompt/enhanced_pattern_library.go"
  "src/security/prompt/advanced_jailbreak_detector.go"
  "src/security/prompt/pattern_library.go"
  "src/security/prompt/protection.go"
  "src/security/prompt/types.go"
  "src/security/api/rate_limiter.go"
  "src/security/api/secure_logging.go"
  "src/security/api/anomaly_detection.go"
  "src/provider/core/logger.go"
  "src/security/communication/cert_chain_utils.go"
  "src/security/communication/cert_chain.go"
)

for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    echo "Cleaning $file"
    # Create a temporary file with proper imports
    awk '
      /^package / { print; seen_package=1; next }
      /^import "/ && seen_package && !seen_first_import { 
        # Skip standalone imports after package
        next 
      }
      /^import \(/ { in_import=1; print; next }
      /^\)/ && in_import { in_import=0; print; next }
      { print }
    ' "$file" > "$file.tmp"
    mv "$file.tmp" "$file"
  fi
done

echo "Import cleanup completed"