package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v41/github"
)

// post github commit status
func PostCommitStatus(ctx context.Context, client *github.Client, owner, repo, sha, state, context string) error {
	commit_state := strings.ToUpper(string(state[0])) + state[1:]
	status := &github.RepoStatus{
		State:       &state,
		Description: &commit_state,
		Context:     &context,
	}
	// capture response
	_, _, err := client.Repositories.CreateStatus(ctx, owner, repo, sha, status)
	fmt.Printf("Commit status posted: %s\n", commit_state)
	// log response
	if err != nil {
		fmt.Printf("Error creating commit status: %v\n", err)
	}
	return err
}
