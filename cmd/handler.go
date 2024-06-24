package cmd

import (
	"context"
	"log"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
	"gh-pr-commenter/internal"
)

// RepoInfo holds the repository information
type RepoInfo struct {
	Owner    string
	Repo     string
	PRNumber int
}

// CommandHandler handles CLI commands
type CommandHandler struct {
	Client        *github.Client
	GraphqlClient *graphql.Client
	RepoInfo      RepoInfo
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber int) *CommandHandler {
	return &CommandHandler{
		Client: client,
		GraphqlClient: graphqlClient,
		RepoInfo: RepoInfo{
			Owner:    owner,
			Repo:     repo,
			PRNumber: prNumber,
		},
	}
}

// HandleCommand processes the CLI command
func (h *CommandHandler) HandleCommand(ctx context.Context, filename string) {
	message, err := internal.ReadCommentFromFile(filename)
	if err != nil {
		log.Fatalf("Error reading comment from file: %v", err)
	}

	internal.UpsertComment(ctx, h.Client, h.GraphqlClient, h.RepoInfo.Owner, h.RepoInfo.Repo, h.RepoInfo.PRNumber, message)
}
