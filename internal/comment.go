package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxRetries = 3
const minimizedMarker = "<!-- MINIMIZED -->"

// ReadCommentFromFile reads the comment message from a file
func ReadCommentFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	return string(content), nil
}

// UpsertComment handles creating or updating comments on the specified PR
func UpsertComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, commentContentFile, title, identifier string) error {
	message, err := ReadCommentFromFile(commentContentFile)
	if err != nil {
		return fmt.Errorf("error reading comment file: %v", err)
	}
	pullNum, err := strconv.Atoi(prNumber)
	if err != nil {
		return fmt.Errorf("error converting PR number: %v", err)
	}
	comments, err := ListCommentsWithRetry(ctx, client, owner, repo, pullNum)
	if err != nil {
		return fmt.Errorf("error listing comments: %v", err)
	}

	existingComments := filterCommentsByTitleAndIdentifier(comments, title, identifier)

	// Always hide previous comments
	err = minimizeComments(ctx, graphqlClient, existingComments)
	if err != nil {
		return fmt.Errorf("error minimizing comments: %v", err)
	}

	// Always create new parts with unique content to avoid collapsing
	timestamp := time.Now().Format(time.RFC3339)
	uniquePart := fmt.Sprintf("%s\n<!-- Unique ID: %s -->", message, timestamp)
	comment := &github.IssueComment{Body: &uniquePart}
	err = createCommentWithRetry(ctx, client, owner, repo, pullNum, comment)
	if err != nil {
		return fmt.Errorf("error creating comment: %v", err)
	}

	fmt.Println("Comment upserted successfully.")
	return nil
}

// ListCommentsWithRetry lists comments with retry logic and pagination
func ListCommentsWithRetry(ctx context.Context, client *github.Client, owner, repo string, pullNum int) ([]*github.IssueComment, error) {
	var allComments []*github.IssueComment
	var err error

	for i := 0; i < maxRetries; i++ {
		opts := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
		for {
			comments, resp, err := client.Issues.ListComments(ctx, owner, repo, pullNum, opts)
			if err != nil {
				fmt.Printf("Error listing comments (attempt %d/%d): %v\n", i+1, maxRetries, err)
				time.Sleep(time.Second * time.Duration(1<<i)) // Exponential backoff
				break
			}
			allComments = append(allComments, comments...)
			if resp.NextPage == 0 {
				return allComments, nil
			}
			opts.Page = resp.NextPage
		}
	}

	return nil, fmt.Errorf("error listing comments after %d retries: %w", maxRetries, err)
}

// createCommentWithRetry creates a comment with retry logic
func createCommentWithRetry(ctx context.Context, client *github.Client, owner, repo string, pullNum int, comment *github.IssueComment) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		_, _, err = client.Issues.CreateComment(ctx, owner, repo, pullNum, comment)
		if err == nil {
			return nil
		}
		fmt.Printf("Error creating comment (attempt %d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(time.Second * time.Duration(1<<i)) // Exponential backoff
	}
	return fmt.Errorf("error creating comment after %d retries: %w", maxRetries, err)
}

// filterCommentsByTitleAndIdentifier filters comments to find those that match the given title and identifier
func filterCommentsByTitleAndIdentifier(comments []*github.IssueComment, title, identifier string) []*github.IssueComment {
	var filtered []*github.IssueComment
	for _, comment := range comments {
		if strings.Contains(comment.GetBody(), title) && strings.Contains(comment.GetBody(), identifier) {
			filtered = append(filtered, comment)
		}
	}
	return filtered
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

// minimizeComments hides the given comments using the minimizeComment GraphQL mutation
func minimizeComments(ctx context.Context, graphqlClient *graphql.Client, comments []*github.IssueComment) error {
	for _, comment := range comments {
		// Skip comments that already have the minimized marker
		if strings.Contains(comment.GetBody(), minimizedMarker) {
			continue
		}
		err := minimizeCommentWithRetry(ctx, graphqlClient, comment.GetNodeID())
		if err != nil {
			return fmt.Errorf("error minimizing comment: %v", err)
		}
		// Add the minimized marker to the comment body
		updatedBody := comment.GetBody() + "\n" + minimizedMarker
		err = updateCommentBody(ctx, graphqlClient, comment.GetNodeID(), updatedBody)
		if err != nil {
			return fmt.Errorf("error updating comment body: %v", err)
		}
	}
	fmt.Println("Comments minimized successfully.")
	return nil
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
	fmt.Printf("Comment minimized: %s\n", commentNodeID)
	return nil
}

// updateCommentBody updates the body of a minimized comment to include the minimized marker
func updateCommentBody(ctx context.Context, graphqlClient *graphql.Client, commentNodeID, updatedBody string) error {
	req := graphql.NewRequest(`
		mutation($id: ID!, $body: String!) {
			updateIssueComment(input: {id: $id, body: $body}) {
				issueComment {
					body
				}
			}
		}
	`)
	req.Var("id", commentNodeID)
	req.Var("body", updatedBody)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))

	var respData struct {
		UpdateIssueComment struct {
			IssueComment struct {
				Body string
			}
		}
	}

	if err := graphqlClient.Run(ctx, req, &respData); err != nil {
		return err
	}

	fmt.Printf("Comment updated with minimized marker: %s\n", commentNodeID)
	return nil
}
