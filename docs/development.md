# Development Guide

This document provides instructions for setting up the development environment and contributing to the GitHub PR Commenter (ghpc) project.

## Prerequisites

- Go 1.21.6 or higher installed on your machine.
- GitHub token set as `GITHUB_TOKEN` environment variable.

## Setting Up the Development Environment

1. **Clone the Repository**

    ```bash
    git clone https://github.com/EdwinPhilip/gh-pr-commenter.git
    cd gh-pr-commenter
    ```

2. **Install Dependencies**

    The project uses Go modules. Install the dependencies by running:

    ```bash
    go mod tidy
    ```

3. **Build the Project**

    ```bash
    go build -o ghpc .
    ```

4. **Run the Application**

    Set the required environment variables and run the application:

    ```bash
    export HEAD_COMMIT="test-commit"
    export PROJECT_NAME="test-project"
    export GH_STATUS_CONTEXT="test-context"
    export WORKSPACE="test-workspace"
    export BASE_REPO_OWNER="test-owner"
    export BASE_REPO_NAME="test-repo"
    export PULL_NUM="123"
    export GITHUB_TOKEN="test-token"

    ./ghpc exec "tflint"
    ```

## Code Structure

- **cmd/**: Contains command-line interface code.
- **config/**: Configuration management code.
- **internal/**: Internal logic, including client and comment handling.
- **pkg/**: Contains reusable packages, such as comment handling.
- **tests/**: Contains test cases for different packages.

## Contributing

1. **Fork the Repository**

    Fork the repository on GitHub and clone your fork locally.

2. **Create a Branch**

    Create a feature branch for your changes:

    ```bash
    git checkout -b my-feature-branch
    ```

3. **Make Changes**

    Make your changes in the feature branch.

4. **Run Tests**

    Ensure all tests pass before pushing your changes:

    ```bash
    go test ./...
    ```

5. **Commit and Push**

    Commit your changes and push to your fork:

    ```bash
    git add .
    git commit -m "Description of your changes"
    git push origin my-feature-branch
    ```

6. **Create a Pull Request**

    Create a pull request from your feature branch to the main branch of the original repository.

## Development Tools

- **Visual Studio Code**: Recommended code editor with Go support.
- **GoLand**: JetBrains IDE for Go development.
- **GitHub CLI**: For managing GitHub repositories from the command line.

## Debugging

Use the `zap` logging package for detailed logging. Logs are configured in the `config` package and used throughout the application.

```go
import "go.uber.org/zap"

logger, _ := zap.NewDevelopment()
defer logger.Sync()
logger.Info("Configuration initialized", zap.String("key", "value"))
```

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [GitHub API Documentation](https://docs.github.com/en/rest)
- [GraphQL Documentation](https://graphql.org/learn/)