# Changelog Association Mechanism

## Overview

This document defines the changelog association mechanism for the LLMreconing Tool. It outlines how changes are tracked, associated with versions, and presented to users across the core binary, templates, and provider modules.

## Changelog Structure

### Core Binary Changelog

The core binary will maintain a structured changelog in the repository root:

**File Location**: `CHANGELOG.md`

**Format**:
```markdown
# Changelog

All notable changes to the LLMreconing Tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Feature in development

## [1.2.3] - 2025-05-16

### Added
- New command for offline bundle verification
- Support for custom template categories

### Changed
- Improved error messages for template validation
- Enhanced performance for large template sets

### Fixed
- Issue with template loading on Windows
- Bug in version comparison logic

### Security
- Updated dependencies to address CVE-2025-12345

## [1.2.2] - 2025-05-01

...
```

### Template Changelog

Templates will have changelog information at two levels:

1. **Collection Level**: Overall changes to the template collection
2. **Individual Template Level**: Changes to specific templates

**Collection Changelog Location**: `templates/CHANGELOG.md`

**Format**:
```markdown
# Template Changelog

## [1.1.0] - 2025-05-15

### Added
- 5 new prompt injection templates
- 3 new data leakage templates

### Updated
- Improved effectiveness of jailbreak templates
- Enhanced detection patterns in information disclosure templates

### Removed
- Deprecated legacy template format

## [1.0.0] - 2025-04-01

...
```

**Individual Template Changes**: Each template file will include a change history section:

```yaml
id: prompt-injection-basic
name: Basic Prompt Injection Test
version: 1.2.0
category: prompt_injection
author: Security Team
created: 2025-03-15
last_updated: 2025-05-10

# Change history
changes:
  - version: 1.2.0
    date: 2025-05-10
    description: Improved detection patterns and success criteria
  - version: 1.1.0
    date: 2025-04-20
    description: Added support for multi-turn conversations
  - version: 1.0.0
    date: 2025-03-15
    description: Initial version

# Template content
...
```

### Module Changelog

Each provider module will maintain its own changelog:

**Location**: `modules/providers/{module_name}/CHANGELOG.md`

**Format**:
```markdown
# OpenAI Provider Module Changelog

## [1.2.0] - 2025-05-10

### Added
- Support for GPT-4o model
- Streaming response capability

### Changed
- Improved error handling for rate limits
- Enhanced token counting accuracy

### Fixed
- Issue with long prompt handling

## [1.1.0] - 2025-04-15

...
```

## Changelog Generation

### Automated Generation

The system will support automated changelog generation:

1. **Conventional Commits**: Encourage the use of conventional commit messages:
   ```
   feat: add new template for prompt injection
   fix: resolve issue with template loading
   docs: update README with new features
   ```

2. **Generation Tools**: Use tools like `git-cliff` or `conventional-changelog` to automatically generate changelog entries from commits.

3. **CI/CD Integration**: Automate changelog updates as part of the release process.

### Manual Curation

While automation helps, manual curation ensures quality:

1. **Pre-Release Review**: Review and edit the generated changelog before release.

2. **Grouping Related Changes**: Combine related changes for clarity.

3. **User-Focused Descriptions**: Ensure descriptions are meaningful to users, not just developers.

## Changelog Association with Versions

### Version Tagging

Each release will be tagged in the version control system:

1. **Core Binary**: Tags like `v1.2.3`
2. **Templates**: Tags like `templates-v1.1.0`
3. **Modules**: Tags like `module-openai-v1.2.0`

### Release Notes

Release notes will be generated for each tagged release:

1. **GitHub/GitLab Releases**: Create releases with changelog content.
2. **Documentation Site**: Update release notes section.
3. **In-Tool Access**: Make release notes accessible via CLI.

### Changelog References

Version information will include references to changelogs:

```json
{
  "core": {
    "version": "1.2.3",
    "changelog_url": "https://github.com/org/LLMrecon/blob/main/CHANGELOG.md#123---2025-05-16"
  },
  "templates": {
    "version": "1.1.0",
    "changelog_url": "https://github.com/org/LLMrecon-templates/blob/main/CHANGELOG.md#110---2025-05-15"
  }
}
```

## User Interface for Changelogs

### CLI Commands

Users can access changelog information via CLI:

1. **Version Command**:
   ```
   $ LLMrecon version
   LLMreconing Tool v1.2.3
   Templates: v1.1.0
   Modules:
   - OpenAI Provider v1.2.0
   - Anthropic Provider v1.1.0
   ```

2. **Changelog Command**:
   ```
   $ LLMrecon changelog
   # Shows changelog for current version
   ```

3. **Changelog with Version**:
   ```
   $ LLMrecon changelog --version=1.2.3
   # Shows changelog for specific version
   ```

4. **Component-Specific Changelog**:
   ```
   $ LLMrecon changelog --component=templates
   # Shows template changelog
   ```

5. **Diff Between Versions**:
   ```
   $ LLMrecon changelog --from=1.2.0 --to=1.2.3
   # Shows changes between versions
   ```

### API Endpoints

Changelog information will be available via API:

1. **Version Endpoint**:
   ```
   GET /api/v1/version
   ```
   Response includes changelog URLs.

2. **Changelog Endpoint**:
   ```
   GET /api/v1/changelog?component=core&version=1.2.3
   ```
   Returns changelog content for specified component and version.

### Update Notifications

When updates are available, changelog highlights will be shown:

```
Updates available:

Core: v1.2.3 → v1.3.0
Highlights:
- New feature: Automated scanning
- Improved template management
- 5 bug fixes

Run 'LLMrecon update apply' to update.
```

## Changelog Categories

Changes will be categorized consistently:

1. **Added**: New features or capabilities
2. **Changed**: Changes to existing functionality
3. **Deprecated**: Features that will be removed in future versions
4. **Removed**: Features that have been removed
5. **Fixed**: Bug fixes
6. **Security**: Security-related changes or fixes

For templates, additional categories may include:

1. **New Templates**: Newly added templates
2. **Updated Templates**: Existing templates that have been improved
3. **Effectiveness Improvements**: Changes that improve detection capabilities

## Changelog Storage and Retrieval

### Local Storage

Changelog information will be stored locally:

1. **Core Changelog**: Stored with the tool installation
2. **Template Changelog**: Stored in the templates directory
3. **Module Changelog**: Stored with each module

### Remote Retrieval

Changelogs can be retrieved from remote sources:

1. **GitHub/GitLab API**: Fetch changelog content from repositories
2. **Documentation Site**: Retrieve formatted changelog content
3. **Update Server**: Include changelog highlights in update metadata

### Caching

To improve performance:

1. **Local Cache**: Cache remote changelog content
2. **Update with Versions**: Refresh cache when checking for updates
3. **Offline Access**: Ensure changelogs are available offline

## Implementation Details

### Changelog Parser

Implement a parser to extract structured information from changelog files:

```go
type ChangelogEntry struct {
    Version     string
    ReleaseDate string
    Changes     map[string][]string // Category -> Changes
}

func ParseChangelog(content string) ([]ChangelogEntry, error) {
    // Parse markdown changelog into structured data
}
```

### Changelog Generator

Implement tools to generate changelog content:

```go
func GenerateChangelog(fromTag, toTag string) (string, error) {
    // Generate changelog from git commits between tags
}
```

### Changelog Storage

Define storage format for changelog data:

```go
type ChangelogData struct {
    Component   string
    Version     string
    ReleaseDate string
    Entries     map[string][]string // Category -> Changes
    URL         string              // Remote URL to full changelog
}
```

## Integration with Version Management

### Update Process

During updates:

1. **Pre-Update Information**: Show changelog highlights before applying updates
2. **Post-Update Summary**: Display applied changes after successful update
3. **Selective Updates**: Allow users to review changes before selecting components to update

### Version History

When displaying version history:

1. **Include Change Summaries**: Show key changes for each version
2. **Link to Full Changelogs**: Provide access to complete changelog content
3. **Filter by Category**: Allow filtering history by change category

## Best Practices for Changelog Maintenance

### For Developers

1. **Write User-Focused Entries**: Describe changes from the user's perspective
2. **Be Specific**: Provide enough detail to understand the change
3. **Group Related Changes**: Combine related changes under a single entry
4. **Reference Issues/PRs**: Link to relevant issues or pull requests
5. **Maintain Consistency**: Use consistent style and format

### For Release Managers

1. **Review Before Release**: Ensure changelog is complete and accurate
2. **Highlight Important Changes**: Emphasize significant changes
3. **Check for Breaking Changes**: Ensure breaking changes are clearly marked
4. **Verify Links**: Ensure all references and links work
5. **Update All Components**: Ensure changelogs are updated for all components

## Conclusion

This changelog association mechanism provides a comprehensive approach to tracking and communicating changes across the LLMreconing Tool's components. By implementing this system, we ensure that users have clear visibility into what has changed between versions, helping them understand the impact of updates and make informed decisions about when to update.
