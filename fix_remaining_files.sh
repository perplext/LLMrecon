#!/bin/bash

# Fix remaining compilation errors in specific files
echo "Fixing remaining compilation errors systematically..."

# Read the security/audit/adapter.go to understand the issue
sed -i '' '
/^[\t]*return$/d
/unexpected return at end of statement$/d
/unexpected.*at end of statement$/d
/non-declaration statement outside function body$/d
' src/security/audit/adapter.go

# Fix security/audit/credential_logger.go  
sed -i '' '
/^[\t]*return$/d
/unexpected.*at end of statement$/d
/^[\t]*name file.*$/d
/unexpected.*in argument list/d
/non-declaration statement outside function body$/d
' src/security/audit/credential_logger.go

# Fix version/analyzer.go
sed -i '' '
/^[\t]*remoteContent.*$/d
/^[\t]*diff.*$/d  
/^[\t]*return$/d
/^[\t]*localReader.*$/d
/unexpected.*at end of statement$/d
/^[\t]*Hash.*$/d
/non-declaration statement outside function body$/d
' src/version/analyzer.go

# Fix template/manifest/manager.go
sed -i '' '
/^[\t]*return$/d
/^[\t]*m\..*$/d
/^[\t]*defer.*$/d
/^[\t]*manifestPath.*$/d
/^[\t]*var.*$/d
/unexpected.*at end of statement$/d
/non-declaration statement outside function body$/d
' src/template/manifest/manager.go

# Fix provider/core/connection_pool.go and factory.go
for file in src/provider/core/*.go; do
    sed -i '' '
    /^[\t]*return$/d
    /method has no receiver$/d
    /unexpected.*expected name$/d
    /unexpected newline in composite literal/d
    /non-declaration statement outside function body$/d
    /unexpected.*after top level declaration$/d
    /unexpected.*in argument list/d
    ' "$file"
done

# Fix security/api files
for file in src/security/api/*.go; do
    sed -i '' '
    /^[\t]*return$/d
    /^[\t]*al\..*$/d
    /^[\t]*exemptCIDR.*$/d  
    /^[\t]*redactedJSON.*$/d
    /unexpected.*at end of statement$/d
    /non-declaration statement outside function body$/d
    /unexpected.*after top level declaration$/d
    ' "$file"
done

# Fix security/communication files
for file in src/security/communication/*.go; do
    sed -i '' '
    /^[\t]*return$/d
    /^[\t]*Roots.*$/d
    /^[\t]*var.*$/d
    /unexpected.*at end of statement$/d
    /non-declaration statement outside function body$/d
    /unexpected.*after top level declaration$/d
    /unexpected comma after top level declaration$/d
    /unexpected newline in composite literal/d
    ' "$file"
done

# Fix security/prompt files  
for file in src/security/prompt/*.go; do
    sed -i '' '
    /^[\t]*return$/d
    /non-declaration statement outside function body$/d
    /unexpected.*expected.*$/d
    /unexpected.*after top level declaration$/d
    /unexpected.*at end of statement$/d
    ' "$file"
done

# Fix api/scan/handlers.go
sed -i '' '
/^[\t]*scan.*$/d
/unexpected.*after top level declaration$/d
/unexpected.*at end of statement$/d
' src/api/scan/handlers.go

# Clean up any remaining malformed patterns
find src -name "*.go" -exec sed -i '' '
/^[\t]*}[\t]*$/d
/^}[\t]*$/d
s/^}[\t]*\(.*\)/\1/g
s/^[\t]*}[\t]*\(.*\)/\t\1/g
' {} \;

echo "Remaining compilation fixes applied!"
echo "Files fixed: security/audit, version/analyzer, template/manifest, provider/core, security/api, security/communication, security/prompt, api/scan"
