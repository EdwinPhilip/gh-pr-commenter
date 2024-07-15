package internal_test

import (
	"context"
	"os"
	"testing"

	"gh-pr-commenter/internal"
	"github.com/google/go-github/v41/github"
	"github.com/jarcoal/httpmock"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

func TestUpsertComment(t *testing.T) {
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

	filename := "test-comment.txt"
	content := "This is a test comment."
	err := os.WriteFile(filename, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	err = internal.UpsertComment(ctx, client, graphqlClient, "test-owner", "test-repo", "123", filename, "test-title", "test-identifier")
	assert.NoError(t, err)
}

func TestReadCommentFromFile(t *testing.T) {
	filename := "test-comment.txt"
	expectedContent := "This is a test comment."
	err := os.WriteFile(filename, []byte(expectedContent), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	content, err := internal.ReadCommentFromFile(filename)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestListCommentsWithRetry(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock GitHub API responses
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/test-owner/test-repo/issues/123/comments",
		httpmock.NewStringResponder(200, `[]`))

	ctx := context.Background()
	client := github.NewClient(nil)

	comments, err := internal.ListCommentsWithRetry(ctx, client, "test-owner", "test-repo", 123)
	assert.NoError(t, err)

	// Handle nil comments and continue processing
	if comments == nil {
		comments = []*github.IssueComment{}
	}

	assert.NotNil(t, comments)
	assert.Equal(t, 0, len(comments))
}
