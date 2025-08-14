#!/bin/bash

# Files that need time import
time_files=(
  "src/api/scan/service.go"
  "src/api/scan/types.go"
  "src/audit/audit.go"
  "src/bundle/errors/audit.go"
  "src/bundle/errors/enhanced_handler.go"
  "src/provider/config/config.go"
  "src/provider/middleware/circuit_breaker.go"
  "src/provider/middleware/logging.go"
  "src/provider/middleware/request_queue.go"
  "src/provider/middleware/usage_tracker.go"
  "src/security/access/audit/trail/audit_trail.go"
  "src/security/access/audit/trail/loggers.go"
  "src/security/communication/cert_chain_validation.go"
  "src/security/communication/tls.go"
  "src/security/prompt/advanced_template_monitor.go"
)

# Add time import to files that need it
for file in "${time_files[@]}"; do
  if [ -f "$file" ]; then
    echo "Adding time to $file"
    # Check if file already has imports
    if grep -q "^import (" "$file"; then
      # Add time to existing imports if not already there
      if ! grep -q '"time"' "$file"; then
        sed -i '' '/^import (/a\
\	"time"' "$file"
      fi
    else
      # Add new import block after package
      sed -i '' '/^package /a\
\
import "time"' "$file"
    fi
  fi
done

# Files that need os import
os_files=(
  "src/audit/vault.go"
  "src/bundle/errors/handler.go"
  "src/provider/middleware/retry.go"
  "src/security/access/common/cache.go"
  "src/templates/persistence.go"
)

for file in "${os_files[@]}"; do
  if [ -f "$file" ]; then
    echo "Adding os to $file"
    if grep -q "^import (" "$file"; then
      if ! grep -q '"os"' "$file"; then
        sed -i '' '/^import (/a\
\	"os"' "$file"
      fi
    else
      sed -i '' '/^package /a\
\
import "os"' "$file"
    fi
  fi
done

# Files that need filepath import
filepath_files=(
  "src/audit/vault.go"
  "src/bundle/errors/handler.go"
  "src/provider/middleware/retry.go"
  "src/security/access/common/cache.go"
  "src/templates/persistence.go"
)

for file in "${filepath_files[@]}"; do
  if [ -f "$file" ]; then
    echo "Adding filepath to $file"
    if grep -q "^import (" "$file"; then
      if ! grep -q '"path/filepath"' "$file"; then
        sed -i '' '/^import (/a\
\	"path/filepath"' "$file"
      fi
    else
      sed -i '' '/^package /a\
\
import "path/filepath"' "$file"
    fi
  fi
done

# Files that need io import
io_files=(
  "src/api/scan/api.go"
  "src/bundle/errors/aggregator.go"
)

for file in "${io_files[@]}"; do
  if [ -f "$file" ]; then
    echo "Adding io to $file"
    if grep -q "^import (" "$file"; then
      if ! grep -q '"io"' "$file"; then
        sed -i '' '/^import (/a\
\	"io"' "$file"
      fi
    else
      sed -i '' '/^package /a\
\
import "io"' "$file"
    fi
  fi
done

echo "Import fixes completed"