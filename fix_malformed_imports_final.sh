#!/bin/bash

echo "Fixing malformed imports in Go files..."

# Fix retry.go
echo "Fixing src/provider/middleware/retry.go..."
cat > /tmp/retry_imports.txt << 'EOF'
// Package middleware provides middleware components for the Multi-Provider LLM Integration Framework.
package middleware

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/llmrecon/llmrecon/src/provider/core"
)
EOF

# Get the rest of the file after imports
sed -n '/^func/,$p' src/provider/middleware/retry.go > /tmp/retry_rest.txt

# Combine them
cat /tmp/retry_imports.txt > src/provider/middleware/retry.go
echo "" >> src/provider/middleware/retry.go
cat /tmp/retry_rest.txt >> src/provider/middleware/retry.go

# Fix enhanced_handler.go
echo "Fixing src/bundle/errors/enhanced_handler.go..."
sed -i '' '5s/.*import "time"/import (\n\t"time"\n)/' src/bundle/errors/enhanced_handler.go 2>/dev/null || true

# Check if it needs more fixing
if grep -q 'import "time"$' src/bundle/errors/enhanced_handler.go 2>/dev/null; then
    # Read the file and fix it
    head -4 src/bundle/errors/enhanced_handler.go > /tmp/enhanced_head.txt
    echo 'import (' >> /tmp/enhanced_head.txt
    echo '	"context"' >> /tmp/enhanced_head.txt
    echo '	"fmt"' >> /tmp/enhanced_head.txt
    echo '	"io"' >> /tmp/enhanced_head.txt
    echo '	"sync"' >> /tmp/enhanced_head.txt
    echo '	"time"' >> /tmp/enhanced_head.txt
    echo ')' >> /tmp/enhanced_head.txt
    echo '' >> /tmp/enhanced_head.txt
    sed -n '/^type/,$p' src/bundle/errors/enhanced_handler.go >> /tmp/enhanced_head.txt
    mv /tmp/enhanced_head.txt src/bundle/errors/enhanced_handler.go
fi

# Fix local.go
echo "Fixing src/repository/local.go..."
if grep -q 'import "context"' src/repository/local.go 2>/dev/null; then
    sed -i '' '3a\
import (\
	"context"\
	"fmt"\
	"io"\
	"os"\
	"path/filepath"\
	"strings"\
	"time"\
)
' src/repository/local.go
    sed -i '' '/^import "context"/d' src/repository/local.go
fi

# Fix bundle_categorize.go
echo "Fixing src/cmd/bundle_categorize.go..."
if grep -q 'import "encoding/json"' src/cmd/bundle_categorize.go 2>/dev/null; then
    head -3 src/cmd/bundle_categorize.go > /tmp/bundle_cat_head.txt
    echo '' >> /tmp/bundle_cat_head.txt
    echo 'import (' >> /tmp/bundle_cat_head.txt
    echo '	"encoding/json"' >> /tmp/bundle_cat_head.txt
    echo '	"fmt"' >> /tmp/bundle_cat_head.txt
    echo '	"strings"' >> /tmp/bundle_cat_head.txt
    echo '' >> /tmp/bundle_cat_head.txt
    echo '	"github.com/spf13/cobra"' >> /tmp/bundle_cat_head.txt
    echo ')' >> /tmp/bundle_cat_head.txt
    echo '' >> /tmp/bundle_cat_head.txt
    sed -n '/^func/,$p' src/cmd/bundle_categorize.go >> /tmp/bundle_cat_head.txt
    mv /tmp/bundle_cat_head.txt src/cmd/bundle_categorize.go
fi

echo "Done fixing malformed imports!"