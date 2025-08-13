# User Customization Preservation System

## Overview

The User Customization Preservation System is designed to identify, preserve, and reapply user customizations during the template and module update process. This system ensures that any modifications made by users to templates or modules are not lost when updates are applied.

## Features

- **Customization Detection**: Automatically identifies user customizations in templates and modules
- **Customization Markers**: Supports special markers to indicate user customizations
- **Preservation Policies**: Different policies for handling customizations during updates
- **Conflict Resolution**: Merges user customizations with updated content
- **Logging**: Detailed logging of all customization preservation actions

## Customization Markers

The system recognizes the following customization markers in files:

1. **User Customization**:
   ```
   # USER CUSTOMIZATION BEGIN
   # Your customized content here
   # USER CUSTOMIZATION END
   ```

2. **Custom Code**:
   ```
   # CUSTOM CODE BEGIN
   # Your custom code here
   # CUSTOM CODE END
   ```

3. **Do Not Modify**:
   ```
   # DO NOT MODIFY BEGIN
   # Content that should never be modified during updates
   # DO NOT MODIFY END
   ```

## Preservation Policies

The system supports different policies for handling customizations:

1. **Always Preserve**: Always keeps the user's customized version, regardless of updates
2. **Preserve with Conflict Resolution**: Attempts to merge the user's customizations with the updated content
3. **Ask User**: Prompts the user to decide what to do with the customization (currently defaults to preserving)
4. **Discard**: Discards the customization and uses the updated version

The policy is determined based on the type of customization marker used:
- `DO NOT MODIFY` markers use the **Always Preserve** policy
- `USER CUSTOMIZATION` and `CUSTOM CODE` markers use the **Preserve with Conflict Resolution** policy
- Unmarked customizations use the **Ask User** policy

## How It Works

1. **Detection Phase**:
   - The system compares the installed templates and modules with their original versions
   - Files with differences are identified as customized
   - Customization markers within files are detected

2. **Preservation Phase**:
   - Before applying updates, customized files are backed up
   - A registry of customizations is maintained with metadata about each customization

3. **Update Phase**:
   - Updates are applied normally to templates and modules

4. **Reapplication Phase**:
   - After updates are applied, the system examines each customization in the registry
   - Based on the preservation policy, customizations are reapplied to the updated files
   - Conflicts are resolved by merging the customized content with the updated content

## Using the System

The User Customization Preservation System is automatically integrated with the update process. When updates are applied using the `UpdateWithCustomizationPreservation` method, the system will:

1. Detect existing customizations
2. Preserve them before the update
3. Apply the update
4. Reapply the customizations to the updated files

### Making Customizations

To ensure your customizations are properly preserved during updates:

1. Use customization markers to indicate your modifications
2. Choose the appropriate marker type based on the nature of your customization:
   - `USER CUSTOMIZATION` for general customizations
   - `CUSTOM CODE` for custom code blocks
   - `DO NOT MODIFY` for critical customizations that should never be changed

### Example

Original template file:
```yaml
id: example-template
info:
  name: Example Template
  version: 1.0.0
  description: An example template
```

Customized template file:
```yaml
id: example-template
info:
  name: Example Template
  version: 1.0.0
  description: An example template
# USER CUSTOMIZATION BEGIN
# This is my custom description
# USER CUSTOMIZATION END
```

After update to version 1.1.0:
```yaml
id: example-template
info:
  name: Example Template
  version: 1.1.0
  description: An updated example template
# USER CUSTOMIZATION BEGIN
# This is my custom description
# USER CUSTOMIZATION END
```

## Customization Registry

The system maintains a registry of all detected customizations in the `data/customization-registry.json` file. This registry contains information about each customization, including:

- Unique identifier
- Type (template or module)
- Path to the customized file
- Component ID
- Base version
- Customization date
- Original and customized file hashes
- Customization markers
- Preservation policy

## Logging

The system logs all customization preservation actions to a log file in the temporary directory during updates. This log includes information about:

- Detected customizations
- Preserved customizations
- Reapplied customizations
- Any errors or warnings encountered during the process

## Technical Details

The User Customization Preservation System consists of three main components:

1. **Registry**: Maintains a record of all customizations
2. **Detector**: Identifies customizations by comparing files and detecting markers
3. **Preserver**: Preserves and reapplies customizations during updates

These components work together to ensure that user customizations are not lost during the update process.
