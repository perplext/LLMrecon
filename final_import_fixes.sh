#!/bin/bash

# Fix imports in specific files that still have issues

# Fix audit package files
echo "Fixing src/security/access/audit/audit_event.go"
sed -i '' 's/^package audit$/package audit\n\nimport "time"/' src/security/access/audit/audit_event.go

echo "Fixing src/security/access/audit/audit_logger.go"
sed -i '' 's/^package audit$/package audit\n\nimport (\n\t"os"\n\t"time"\n)/' src/security/access/audit/audit_logger.go

echo "Fixing src/security/access/audit/audit_manager.go"
sed -i '' 's/^package audit$/package audit\n\nimport "time"/' src/security/access/audit/audit_manager.go

# Fix provider/core/logger.go - add os import
echo "Fixing src/provider/core/logger.go"
sed -i '' 's/^import ($/import (\n\t"os"/' src/provider/core/logger.go

# Fix bundle/errors files
echo "Fixing src/bundle/errors/audit.go"
sed -i '' 's/^import ($/import (\n\t"io"/' src/bundle/errors/audit.go

echo "Fixing src/bundle/errors/categories.go"
sed -i '' 's/^package errors$/package errors\n\nimport "io"/' src/bundle/errors/categories.go

echo "Fixing src/bundle/errors/errors.go"
sed -i '' 's/^package errors$/package errors\n\nimport "time"/' src/bundle/errors/errors.go

echo "Fixing src/bundle/errors/recovery.go"
sed -i '' 's/^package errors$/package errors\n\nimport "io"/' src/bundle/errors/recovery.go

# Fix cert_chain.go and cert_chain_utils.go 
echo "Fixing src/security/communication/cert_chain.go"
sed -i '' 's/^import ($/import (\n\t"time"/' src/security/communication/cert_chain.go

echo "Fixing src/security/communication/cert_chain_utils.go"
# Merge the two import blocks
awk '
BEGIN { in_first_import=0; in_second_import=0; printed_import=0 }
/^import \(/ && !printed_import { 
    print "import ("
    print "\t\"os\""
    print "\t\"path/filepath\""
    print "\t\"time\""
    printed_import=1
    in_first_import=1
    next
}
/^\)/ && in_first_import { in_first_import=0; next }
in_first_import { next }
/^import \(/ && printed_import { in_second_import=1; next }
!in_second_import { print }
' src/security/communication/cert_chain_utils.go > src/security/communication/cert_chain_utils.go.tmp
mv src/security/communication/cert_chain_utils.go.tmp src/security/communication/cert_chain_utils.go

# Fix the rest of the files needing time/io imports
echo "Fixing remaining files..."

for file in src/security/api/anomaly_detection.go src/security/api/secure_logging.go src/security/api/rate_limiter.go; do
  if grep -q "^import (" "$file" && ! grep -q '"time"' "$file"; then
    sed -i '' 's/^import ($/import (\n\t"time"/' "$file"
  fi
done

for file in src/security/api/secure_logging.go; do
  if grep -q "^import (" "$file" && ! grep -q '"io"' "$file"; then
    sed -i '' 's/^import ($/import (\n\t"io"/' "$file"
  fi
done

for file in src/security/prompt/types.go src/security/prompt/protection.go src/security/prompt/pattern_library.go src/security/prompt/advanced_jailbreak_detector.go src/security/prompt/enhanced_pattern_library.go; do
  if [ -f "$file" ] && grep -q "^import (" "$file" && ! grep -q '"time"' "$file"; then
    sed -i '' 's/^import ($/import (\n\t"time"/' "$file"
  fi
done

echo "Import fixes completed"