package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/machinebox/graphql"
	"gh-pr-commenter/cmd"
	"gh-pr-commenter/internal"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "exec" {
		log.Fatalf("usage: %s exec <command>", os.Args[0])
	}

	owner := os.Getenv("BASE_REPO_OWNER")
	repo := os.Getenv("BASE_REPO_NAME")
	prNumberStr := os.Getenv("PULL_NUM")
	if owner == "" || repo == "" || prNumberStr == "" {
		log.Fatalf("Environment variables BASE_REPO_OWNER, BASE_REPO_NAME, and PULL_NUM must be set")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		log.Fatalf("Invalid PR number: %s", prNumberStr)
	}
	command := os.Args[2]

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	cmd.ExecuteAndComment(ctx, client, graphqlClient, owner, repo, prNumber, command)
}


