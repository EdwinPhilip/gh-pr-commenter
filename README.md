Sure, here's a basic README.md template for your project, considering the modifications and GitHub Actions setup:

```markdown
# GitHub PR Commenter

GitHub PR Commenter is a tool that automates the process of executing commands and posting their output as comments on GitHub Pull Requests (PRs).

## Overview

This tool leverages GitHub's API and GraphQL to interact with PRs and comments. It allows you to run commands, capture their output, and post the output as comments on GitHub PRs.

## Features

- **Comment Automation:** Automatically posts comments on pull requests with command outputs.
- **GitHub Actions Integration:** Uses GitHub Actions for building, testing, and releasing.
- **Multi-Architecture Builds:** Supports builds for multiple architectures including Linux and Darwin (AMD64 and ARM64).
- **Output Handling:** Handles command outputs exceeding 55000 characters by splitting them into multiple comments.

## Usage

### Prerequisites

- Go 1.21.6 or higher installed on your machine.
- GitHub token set as `GITHUB_TOKEN` environment variable.

### Installation

Clone the repository:

```bash
git clone https://github.com/EdwinPhilip/gh-pr-commenter.git
cd gh-pr-commenter
```

Build the project:

```bash
go build -o ghpc .
```

### Configuration

1. Set environment variables:
   - `BASE_REPO_OWNER`: Owner of the base repository.
   - `BASE_REPO_NAME`: Name of the base repository.
   - `PULL_NUM`: PR number where the comments will be posted.
   - `GITHUB_TOKEN`: GitHub token

2. Customize `template.md` file for comment formatting.

### Running

Execute a command and post its output as a PR comment:

```bash
./ghpc exec "tflint"
```

### GitHub Actions

This project includes GitHub Actions for continuous integration and release automation:

- **Build Workflow**: Automatically builds the project for multiple architectures on each push event.
- **Release Workflow**: Creates a new release and tags it with version numbers specified during workflow dispatch.

### Versioning

The project follows Semantic Versioning (SemVer). Tags pushed to the main branch with a version format (`v*.*.*`) will trigger a version bump and release.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
