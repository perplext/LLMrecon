#!/bin/bash

# Script to verify OWASP LLM compliance for templates
# Usage: ./verify-compliance.sh [--templates-dir=<dir>] [--output=<file>] [--format=json|yaml]

# Default values
TEMPLATES_DIR="./examples/templates"
OUTPUT_FORMAT="json"
OUTPUT_FILE=""
VERBOSE=false

# Parse command-line arguments
for arg in "$@"; do
  case $arg in
    --templates-dir=*)
      TEMPLATES_DIR="${arg#*=}"
      shift
      ;;
    --output=*)
      OUTPUT_FILE="${arg#*=}"
      shift
      ;;
    --format=*)
      OUTPUT_FORMAT="${arg#*=}"
      shift
      ;;
    --verbose)
      VERBOSE=true
      shift
      ;;
    *)
      # Unknown option
      echo "Unknown option: $arg"
      echo "Usage: ./verify-compliance.sh [--templates-dir=<dir>] [--output=<file>] [--format=json|yaml] [--verbose]"
      exit 1
      ;;
  esac
done

# Check if templates directory exists
if [ ! -d "$TEMPLATES_DIR" ]; then
  echo "Error: Templates directory '$TEMPLATES_DIR' does not exist."
  exit 1
fi

# Build the command
CMD="./compliance-report"
CMD="$CMD --framework=owasp-llm"
CMD="$CMD --templates=$TEMPLATES_DIR"
CMD="$CMD --format=$OUTPUT_FORMAT"

if [ -n "$OUTPUT_FILE" ]; then
  CMD="$CMD --output=$OUTPUT_FILE"
fi

if [ "$VERBOSE" = true ]; then
  CMD="$CMD --verbose"
fi

# Check if the compliance-report binary exists
if [ ! -f "./compliance-report" ]; then
  echo "Building compliance-report tool..."
  go build -o compliance-report ./cmd/compliance-report
fi

# Run the command
echo "Verifying OWASP LLM compliance for templates in $TEMPLATES_DIR..."
$CMD

# Check if the command succeeded
if [ $? -eq 0 ]; then
  if [ -n "$OUTPUT_FILE" ]; then
    echo "Compliance report generated successfully: $OUTPUT_FILE"
  else
    echo "Compliance report generated successfully."
  fi
else
  echo "Error generating compliance report."
  exit 1
fi

exit 0
