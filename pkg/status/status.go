package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v41/github"
)

// PostCommitStatus posts the GitHub commit status
func PostCommitStatus(ctx context.Context, client *github.Client, owner, repo, sha, state, context string) error {
	commitState := strings.ToUpper(string(state[0])) + state[1:]
	if state == "failure" {
		commitState = "Failed"
	}
	if state == "pending" {
		commitState = "In Progress"
	}
	status := &github.RepoStatus{
		State:       &state,
		Description: &commitState,
		Context:     &context,
	}
	_, _, err := client.Repositories.CreateStatus(ctx, owner, repo, sha, status)
	if err != nil {
		return fmt.Errorf("error creating commit status: %w", err)
	}
	fmt.Printf("Commit status posted: %s\n", commitState)
	return nil
}
