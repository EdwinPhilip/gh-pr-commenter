package cmd

import (
	"context"

	"gh-pr-commenter/pkg/comments"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

func Comment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) {
	comments.Comment(ctx, client, graphqlClient, owner, repo, prNumber, command)
}
