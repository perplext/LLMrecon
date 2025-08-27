#!/bin/bash

file="src/template/management/execution/optimized_engine.go"

# Replace pattern: return statement followed by empty line and comment, with return statement, closing brace, empty line, and comment
sed -i '' 's/\treturn[^}]*$/&\
}/g' "$file"

# Replace specific patterns that are common
sed -i '' '
# Fix functions that end with return statement followed by comment
/\treturn.*$/{
  N
  /\n$/{
    N
    /\n\/\//s/$/\
}/
  }
}' "$file"

echo "Fixed closing braces in $file"