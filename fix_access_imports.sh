#!/bin/bash

# Fix security/access package imports
# Replace full module paths with relative imports

echo "Fixing security/access package imports..."

# Files to fix - all Go files in security/access that have the problematic imports
files=(
    "src/security/access/impl/store_adapters.go"
    "src/security/access/impl/factory.go"
    "src/security/access/impl/additional_adapters.go"
    "src/security/access/impl/security_manager_adapter.go"
    "src/security/access/impl/converters.go"
    "src/security/access/rbac/rbac_manager.go"
    "src/security/access/audit_logger_impl.go"
    "src/security/access/in_memory_stores.go"
    "src/security/access/types/security_types.go"
    "src/security/access/auth.go"
    "src/security/access/factory.go"
    "src/security/access/tests/test_helpers.go"
    "src/security/access/security.go"
    "src/security/access/converters/audit_converter.go"
    "src/security/access/converters/security_converters.go"
    "src/security/access/converters/models_converter.go"
    "src/security/access/converters/converters.go"
    "src/security/access/mfa/boundary.go"
    "src/security/access/mfa/api.go"
    "src/security/access/access_control_system.go"
    "src/security/access/adapters/adapters.go"
    "src/security/access/adapters/security_manager_adapter.go"
    "src/security/access/audit/audit_manager.go"
    "src/security/access/audit/audit_event.go"
    "src/security/access/auth_manager_impl.go"
    "src/security/access/db/stub_stores.go"
    "src/security/access/db/factory.go"
    "src/security/access/db/adapter/audit_store_adapter.go"
    "src/security/access/db/adapter/vulnerability_store_adapter.go"
    "src/security/access/db/adapter/incident_store_adapter.go"
    "src/security/access/db/adapter/user_store_adapter.go"
    "src/security/access/db/adapter/session_store_adapter.go"
    "src/security/access/api/incident_handlers.go"
    "src/security/access/api/server.go"
    "src/security/access/api/mfa_api.go"
    "src/security/access/api/audit_handlers.go"
    "src/security/access/api/auth_handlers.go"
    "src/security/access/api/role_handlers.go"
    "src/security/access/api/middleware.go"
    "src/security/access/api/vulnerability_handlers.go"
    "src/security/access/api/user_handlers.go"
    "src/security/access/integration.go"
    "src/security/access/stub_managers.go"
    "src/security/access/mfa_context.go"
    "src/security/access/legacy_interfaces.go"
    "src/security/access/interfaces/stores.go"
    "src/security/access/interfaces/security.go"
)

for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "Processing $file..."
        
        # Replace full imports with relative imports
        sed -i.bak \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/models"|"../models"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/common"|"../common"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/mfa"|"../mfa"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/rbac"|"../rbac"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/interfaces"|"../interfaces"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/audit"|"../audit"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/adapters"|"../adapters"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/converters"|"../converters"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/db/adapter"|"../db/adapter"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/db"|"../db"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/impl"|"../impl"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access/types"|"../types"|g' \
            -e 's|"github.com/perplext/LLMrecon/src/security/access"|".."|g' \
            "$file"
        
        # Remove backup file
        rm -f "$file.bak"
    fi
done

echo "Security/access import fixes complete!"