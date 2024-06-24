package internal

import (
    "context"
    "os"

    "github.com/google/go-github/v41/github"
    "golang.org/x/oauth2"
)

// NewGitHubClient creates and returns a new GitHub client
func NewGitHubClient(ctx context.Context) *github.Client {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        panic("GITHUB_TOKEN environment variable is not set")
    }

    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)
    return github.NewClient(tc)
}
