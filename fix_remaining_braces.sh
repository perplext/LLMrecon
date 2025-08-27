#!/bin/bash

# Fix remaining missing closing braces in bundle_compliance.go

file="src/cmd/bundle_compliance.go"

# Create a temporary file to work with
cp "$file" "${file}.tmp"

# Use sed to add missing closing braces before function declarations
# Pattern: return statement followed by blank line and function declaration
sed -i '' -E 's/(return [^}]+)/\1/g; /^[[:space:]]*return [^}]*$/N; s/(return [^}]*)\n$/\1/g' "${file}.tmp"

# More specific fixes - add closing braces where functions are missing them
# Look for specific patterns where functions end incorrectly
sed -i '' -E '
# Fix pattern: } followed by blank line and func
/^[[:space:]]*}[[:space:]]*$/N; s/^([[:space:]]*}[[:space:]]*)([[:space:]]*func [[:alpha:]_][[:alnum:]_]*)/\1}\n\n\2/g

# Fix pattern: return statement followed by func
/^[[:space:]]*return [^;]*;?[[:space:]]*$/N; s/^([[:space:]]*return [^;]*;?[[:space:]]*)([[:space:]]*func [[:alpha:]_][[:alnum:]_]*)/\1}\n\n\2/g

# Fix specific missing braces by function names
/^func saveMarkdownDocument/i\
}

/^func saveJSONDocument/i\
}

/^func saveHTMLDocument/i\
}

/^func generateOWASPTemplates/i\
}

/^func generateISO42001Templates/i\
}

/^func generateNISTTemplates/i\
}

/^func generateEUAITemplates/i\
}
' "${file}.tmp"

# Move the temporary file back
mv "${file}.tmp" "$file"

echo "Applied brace fixes to $file"