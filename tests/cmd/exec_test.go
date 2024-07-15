package cmd_test

import (
	"context"
	"os"
	"testing"

	"gh-pr-commenter/cmd"
	"gh-pr-commenter/config"
	"github.com/google/go-github/v41/github"
	"github.com/jarcoal/httpmock"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

func TestExecuteAndComment(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ctx := context.Background()
	client := github.NewClient(nil)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	// Mock GitHub API responses
	httpmock.RegisterResponder("POST", "https://api.github.com/repos/test-owner/test-repo/statuses/test-commit",
		httpmock.NewStringResponder(201, `{}`))
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
	filename := cnf.TmpGhpcDir + "/.output-test-cmd.md"
	content := "This is a test command output."
	err = os.WriteFile(filename, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	err = cmd.ExecuteAndComment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, "echo Hello")
	assert.NoError(t, err)
}
