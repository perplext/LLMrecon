#!/bin/bash

# Fix missing time imports
for file in \
  src/security/communication/cert_chain.go \
  src/security/communication/cert_chain_utils.go \
  src/security/api/anomaly_detection.go \
  src/security/api/secure_logging.go \
  src/security/api/rate_limiter.go \
  src/security/prompt/types.go \
  src/security/prompt/protection.go \
  src/security/prompt/pattern_library.go \
  src/security/prompt/advanced_jailbreak_detector.go \
  src/security/prompt/enhanced_pattern_library.go
do
  if [ -f "$file" ]; then
    # Check if time is already imported
    if ! grep -q '"time"' "$file"; then
      # Add time import after package declaration
      sed -i '' '/^package /a\
\
import "time"
' "$file"
    fi
  fi
done

# Fix missing os import
for file in \
  src/security/communication/cert_chain_utils.go \
  src/provider/core/logger.go
do
  if [ -f "$file" ]; then
    if ! grep -q '"os"' "$file"; then
      sed -i '' '/^package /a\
\
import "os"
' "$file"
    fi
  fi
done

# Fix missing filepath import
if [ -f "src/security/communication/cert_chain_utils.go" ]; then
  if ! grep -q '"path/filepath"' "src/security/communication/cert_chain_utils.go"; then
    sed -i '' 's/import "os"/import (\n\t"os"\n\t"path\/filepath"\n)/' "src/security/communication/cert_chain_utils.go"
  fi
fi

# Fix missing io import
for file in \
  src/security/api/secure_logging.go \
  src/bundle/errors/audit.go
do
  if [ -f "$file" ]; then
    if ! grep -q '"io"' "$file"; then
      sed -i '' '/^package /a\
\
import "io"
' "$file"
    fi
  fi
done

echo "Import fixes completed"