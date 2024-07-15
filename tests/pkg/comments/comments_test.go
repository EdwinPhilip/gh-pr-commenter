package comments_test

import (
	"context"
	"os"
	"testing"

	"gh-pr-commenter/config"
	"gh-pr-commenter/pkg/comments"
	"github.com/google/go-github/v41/github"
	"github.com/jarcoal/httpmock"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

func TestComment(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ctx := context.Background()
	client := github.NewClient(nil)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	// Mock GitHub API responses
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/test-owner/test-repo/issues/123/comments",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder("POST", "https://api.github.com/repos/test-owner/test-repo/issues/123/comments",
		httpmock.NewStringResponder(201, `{}`))

	// Set up mock environment variables
	os.Setenv("HEAD_COMMIT", "test-commit")
	os.Setenv("PROJECT_NAME", "test-project")
	os.Setenv("GH_STATUS_CONTEXT", "test-context")
	os.Setenv("WORKSPACE", "test-workspace")
	os.Setenv("BASE_REPO_OWNER", "test-owner")
	os.Setenv("BASE_REPO_NAME", "test-repo")
	os.Setenv("PULL_NUM", "123")
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("TEMPLATE_FILENAME", "test-template.md")
	os.Setenv("TMP_GHPC_DIR", "/tmp/test-ghpc")

	config.Init("test-cmd")
	cnf := config.GetConfig()

	// Ensure the temporary directory exists
	err := os.MkdirAll(cnf.TmpGhpcDir, 0755)
	assert.NoError(t, err)

	// Create a temporary command output file
	filename := cnf.TmpGhpcDir + "/.output-echo.md"
	content := "This is a test command output."
	err = os.WriteFile(filename, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	err = comments.Comment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, "echo Hello")
	assert.NoError(t, err)
}

func TestSplitMessage(t *testing.T) {
	message := "This is a test message that is intentionally long to test the splitMessage function."
	parts := comments.SplitMessage(message)
	assert.True(t, len(parts) > 0)
	assert.Contains(t, parts[0], "This is a test message")
}

func TestCreateDefaultTemplate(t *testing.T) {
	filename := "test-template.md"
	command := "test-command"
	err := comments.CreateDefaultTemplate(filename, command)
	assert.NoError(t, err)
	defer os.Remove(filename)

	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "---OUTPUT---")
}
