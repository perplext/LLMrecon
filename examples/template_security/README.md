# Template Security Framework

This example demonstrates the comprehensive template validation and sandboxing framework for securing template execution in the LLMrecon project. The framework provides robust protection against template injection attacks and ensures safe execution of templates through validation, sandboxing, and a comprehensive approval workflow.

## Features

The security framework includes the following components:

1. **Template Validator**: Validates templates against security rules, checking for:
   - Syntax and semantic correctness
   - Potentially dangerous functions and patterns
   - Input validation issues
   - Performance concerns
   - Complexity metrics

2. **Template Sandbox**: Executes templates in a controlled environment with:
   - Resource limits (CPU, memory, execution time)
   - Restricted access to file system and network
   - Controlled execution modes (strict, permissive, audit)
   - Optional container-based isolation

3. **Template Scorer**: Evaluates template risk based on:
   - Security issues detected
   - Resource usage patterns
   - Complexity metrics
   - Assigns risk categories (low, medium, high, critical)

4. **Approval Workflow**: Manages template versioning and approval with:
   - Version tracking and storage
   - Status management (draft, pending review, approved, rejected)
   - Approval roles and permissions
   - Audit trail with comments

5. **Metrics and Monitoring**: Collects and displays metrics for:
   - Validation statistics and issues
   - Execution performance and resource usage
   - Workflow status and approvals
   - Security alerts and notifications

6. **Dashboard**: Provides a web-based dashboard for:
   - Real-time monitoring of security metrics
   - Visualization of template risks and issues
   - Tracking of workflow status
   - Management of security alerts

## Usage

### Building the Example

```bash
go build -o template-security main.go
```

### Running the Example

```bash
# Validate and execute a safe template
./template-security --template=sample_templates/safe_template.tmpl

# Validate and execute a risky template with strict mode
./template-security --template=sample_templates/risky_template.tmpl --mode=strict

# Use container-based sandbox for additional isolation
./template-security --template=sample_templates/risky_template.tmpl --container

# Enable approval workflow
./template-security --template=sample_templates/safe_template.tmpl --workflow --user=admin

# Only validate without execution
./template-security --template=sample_templates/risky_template.tmpl --validate --execute=false

# Run in audit mode (logs issues but allows execution)
./template-security --template=sample_templates/risky_template.tmpl --mode=audit

# Process all templates in a directory
./template-security --batch --template-dir=sample_templates --verbose

# Start the dashboard server
./template-security --dashboard --port=8080

# Run the complete demo script
./run_examples.sh
```

### Command Line Options

- `--template`: Path to the template file
- `--mode`: Execution mode (strict, permissive, audit) (default: strict)
- `--container`: Use container-based sandbox (default: false)
- `--workflow`: Enable approval workflow (default: false)
- `--storage`: Storage directory for workflow data (default: ./template_storage)
- `--logs`: Directory for logs (default: ./logs)
- `--validate`: Validate the template (default: true)
- `--execute`: Execute the template (default: true)
- `--user`: User for workflow operations (default: admin)
- `--dashboard`: Start the dashboard server (default: false)
- `--port`: Dashboard server port (default: 8080)
- `--batch`: Run batch processing of all templates in a directory (default: false)
- `--template-dir`: Directory containing templates for batch processing (default: ./sample_templates)
- `--verbose`: Enable verbose output (default: false)

## Sample Templates

The example includes two sample templates:

1. `safe_template.tmpl`: A safe template that only performs simple text formatting without any dangerous operations.
2. `risky_template.tmpl`: A template with potentially dangerous operations that should be detected by the security framework.

## Implementation Details

The security framework is implemented in the following files:

- `src/template/security/sandbox/types.go`: Defines types for execution modes and resource limits
- `src/template/security/sandbox/sandbox.go`: Implements the default sandbox for template execution
- `src/template/security/sandbox/validator.go`: Implements the template validator
- `src/template/security/sandbox/scorer.go`: Implements the template risk scorer
- `src/template/security/sandbox/container.go`: Implements the container-based sandbox
- `src/template/security/sandbox/workflow.go`: Implements the template approval workflow
- `src/template/security/sandbox/metrics.go`: Implements metrics collection and alerting
- `src/template/security/sandbox/integration.go`: Integrates all components into a unified framework
- `examples/template_security/dashboard.go`: Implements the web-based dashboard
- `examples/template_security/main.go`: Provides a command-line interface to the framework

## Security Considerations

- The framework is designed to prevent template injection attacks and ensure safe execution
- It uses a combination of static analysis and runtime sandboxing
- The container-based sandbox provides additional isolation for high-risk templates
- The approval workflow ensures that templates are reviewed before being used in production
- Resource limits prevent denial-of-service attacks through template execution

## Dashboard Features

The web-based dashboard provides a comprehensive view of the template security framework:

1. **Validation Metrics**:
   - Number of templates validated
   - Validation errors and issues
   - Average validation time
   - Distribution of templates by risk category

2. **Execution Metrics**:
   - Number of templates executed
   - Execution success and failure rates
   - Resource usage statistics (CPU, memory)
   - Average execution time

3. **Workflow Metrics**:
   - Template versions created
   - Approval status distribution
   - Pending reviews and approvals
   - Version history

4. **Security Alerts**:
   - Real-time alerts for security issues
   - Critical and high-risk template notifications
   - Resource usage warnings
   - Execution failures

## Batch Processing

The batch processing feature allows you to validate and execute multiple templates at once:

1. **Directory Processing**:
   - Process all templates in a directory
   - Parallel execution for improved performance
   - Comprehensive reporting of results

2. **Filtering Options**:
   - Process only templates with specific extensions (.tmpl, .template)
   - Skip templates with critical security issues
   - Apply consistent validation and execution settings

## Extending the Framework

The framework can be extended in the following ways:

1. Adding custom security checks to the validator
2. Implementing additional sandbox environments
3. Enhancing the risk scoring algorithm
4. Adding integration with external security tools
5. Implementing additional workflow features (e.g., multi-level approval)
6. Extending the dashboard with additional visualizations
7. Adding custom alerting mechanisms (e.g., email, Slack)
8. Implementing template remediation suggestions
