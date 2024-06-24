package internal

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxCommentLength = 62000
const identifierPrefix = "<!-- Comment"
const maxRetries = 3

// ReadCommentFromFile reads the comment message from a file
func ReadCommentFromFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	return string(content), nil
}

// splitMessage splits the message into parts each with a maximum length of maxCommentLength
func splitMessage(message string) []string {
	var parts []string
	for i := 0; i < len(message); i += maxCommentLength {
		end := i + maxCommentLength
		if end > len(message) {
			end = len(message)
		}
		part := fmt.Sprintf("%s <!-- %s #%d -->\n%s", identifierPrefix, identifierPrefix, len(parts)+1, message[i:end])
		parts = append(parts, part)
	}
	return parts
}

// UpsertComment handles creating or updating comments on the specified PR
func UpsertComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber int, message string) {
	comments, err := listCommentsWithRetry(ctx, client, owner, repo, prNumber)
	if err != nil {
		fmt.Printf("Error listing comments: %v\n", err)
		return
	}

	// Determine the title
	lines := strings.Split(message, "\n")
	title := ""
	if len(lines) > 0 && strings.HasPrefix(lines[0], "##") {
		title = strings.TrimSpace(lines[0][2:])
	} else {
		title = fmt.Sprintf("Output of command for PR %d", prNumber)
		message = fmt.Sprintf("## %s\n%s", title, message)
	}

	parts := splitMessage(message)
	existingComments := filterCommentsByTitle(comments, identifierPrefix)

	// Always hide previous comments
	minimizeComments(ctx, graphqlClient, existingComments)

	// Always create new parts with unique content to avoid collapsing
	timestamp := time.Now().Format(time.RFC3339)
	for i, part := range parts {
		uniquePart := fmt.Sprintf("%s\n<!-- Unique ID: %d %s -->", part, i+1, timestamp)
		comment := &github.IssueComment{Body: &uniquePart}
		err := createCommentWithRetry(ctx, client, owner, repo, prNumber, comment)
		if err != nil {
			fmt.Printf("Error creating comment: %v\n", err)
			return
		}
	}

	fmt.Println("Comment upserted successfully.")
}

// listCommentsWithRetry lists comments with retry logic
func listCommentsWithRetry(ctx context.Context, client *github.Client, owner, repo string, prNumber int) ([]*github.IssueComment, error) {
	var comments []*github.IssueComment
	var err error
	for i := 0; i < maxRetries; i++ {
		comments, _, err = client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
		if err == nil {
			return comments, nil
		}
		fmt.Printf("Error listing comments (attempt %d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(time.Second * time.Duration(1<<i)) // Exponential backoff
	}
	return nil, fmt.Errorf("error listing comments after %d retries: %w", maxRetries, err)
}

// createCommentWithRetry creates a comment with retry logic
func createCommentWithRetry(ctx context.Context, client *github.Client, owner, repo string, prNumber int, comment *github.IssueComment) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
		if err == nil {
			return nil
		}
		fmt.Printf("Error creating comment (attempt %d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(time.Second * time.Duration(1<<i)) // Exponential backoff
	}
	return fmt.Errorf("error creating comment after %d retries: %w", maxRetries, err)
}

// filterCommentsByTitle filters comments to find those that match the given title
func filterCommentsByTitle(comments []*github.IssueComment, title string) []*github.IssueComment {
	var filtered []*github.IssueComment
	for _, comment := range comments {
		if strings.HasPrefix(comment.GetBody(), title) && strings.Contains(comment.GetBody(), identifierPrefix) {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// minimizeComments hides the given comments using the minimizeComment GraphQL mutation
func minimizeComments(ctx context.Context, graphqlClient *graphql.Client, comments []*github.IssueComment) {
	for _, comment := range comments {
		err := minimizeCommentWithRetry(ctx, graphqlClient, comment.GetNodeID())
		if err != nil {
			fmt.Printf("Error minimizing comment: %v\n", err)
			return
		}
	}
	fmt.Println("Comments minimized successfully.")
}

// minimizeCommentWithRetry sends the minimizeComment GraphQL mutation with retry logic
func minimizeCommentWithRetry(ctx context.Context, graphqlClient *graphql.Client, commentNodeID string) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = minimizeComment(ctx, graphqlClient, commentNodeID)
		if err == nil {
			return nil
		}
		fmt.Printf("Error minimizing comment (attempt %d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(time.Second * time.Duration(1<<i)) // Exponential backoff
	}
	return fmt.Errorf("error minimizing comment after %d retries: %w", maxRetries, err)
}

// minimizeComment sends the minimizeComment GraphQL mutation
func minimizeComment(ctx context.Context, graphqlClient *graphql.Client, commentNodeID string) error {
	req := graphql.NewRequest(`
		mutation($id: ID!) {
			minimizeComment(input: {subjectId: $id, classifier: OUTDATED}) {
				minimizedComment {
					isMinimized
					minimizedReason
				}
			}
		}
	`)
	req.Var("id", commentNodeID)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))

	var respData struct {
		MinimizeComment struct {
			MinimizedComment struct {
				IsMinimized    bool
				MinimizedReason string
			}
		}
	}

	if err := graphqlClient.Run(ctx, req, &respData); err != nil {
		return err
	}

	if !respData.MinimizeComment.MinimizedComment.IsMinimized {
		return fmt.Errorf("failed to minimize comment: %s", respData.MinimizeComment.MinimizedComment.MinimizedReason)
	}

	return nil
}
