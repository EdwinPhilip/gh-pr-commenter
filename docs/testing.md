# Testing Guide

This document provides instructions for running tests and ensuring the quality of the GitHub PR Commenter (ghpc) project.

## Running Tests

### Unit Tests

Run all unit tests:

```bash
go test ./...
```

### Specific Test

Run a specific test by specifying the test name:

```bash
go test -timeout 30s -run ^TestFunctionName$ ./...
```

## Test Structure

- **cmd/**: Contains tests for command-line interface code.
- **config/**: Contains tests for configuration management.
- **internal/**: Contains tests for internal logic, including client and comment handling.
- **pkg/**: Contains tests for reusable packages, such as comment handling.

## Writing Tests

### Example Test

Here's an example of a test case for the configuration package:

```go
package config_test

import (
    "os"
    "testing"

    "gh-pr-commenter/config"
    "github.com/spf13/viper"
    "github.com/stretchr/testify/assert"
)

func TestInit_WithDefaults(t *testing.T) {
    os.Clearenv()
    viper.Reset()

    viper.AutomaticEnv()

    config.Init("test-cmd")

    cnf := config.GetConfig()
    assert.Equal(t, config.DefaultProjectName, cnf.ProjectName)
    assert.Equal(t, config.DefaultWorkspace, cnf.Workspace)
    assert.Equal(t, config.DefaultTemplateFile, cnf.TemplateFilename)
    assert.Equal(t, config.DefaultTmpGhpcDir, cnf.TmpGhpcDir)
}
```

### Mocking External Dependencies

Use the `github.com/stretchr/testify/mock` package to mock external dependencies such as GitHub API calls.

### Running Tests with Coverage

Generate a coverage report by running:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Continuous Integration

### GitHub Actions

This project includes GitHub Actions workflows for continuous integration:

- **Build Workflow**: Automatically builds the project for multiple architectures on each push event.
- **Test Workflow**: Runs the test suite on each push and pull request.
- **Release Workflow**: Creates a new release and tags it with version numbers specified during workflow dispatch.

## Debugging Tests

Use the `zap` logging package to add detailed logs for debugging tests.

```go
import "go.uber.org/zap"

logger, _ := zap.NewDevelopment()
defer logger.Sync()
logger.Info("Test started", zap.String("test", "TestInit_WithDefaults"))
```

## Additional Resources

- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Go Testing Documentation](https://golang.org/pkg/testing/)