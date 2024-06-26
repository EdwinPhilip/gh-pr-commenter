package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"gh-pr-commenter/cmd"
	"gh-pr-commenter/internal"

	"github.com/machinebox/graphql"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <command>", os.Args[1])
	}

	owner := os.Getenv("BASE_REPO_OWNER")
	repo := os.Getenv("BASE_REPO_NAME")
	prNumberStr := os.Getenv("PULL_NUM")
	if owner == "" || repo == "" || prNumberStr == "" {
		log.Fatalf("Environment variables OWNER, REPO, and PR_NUMBER must be set")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		log.Fatalf("Invalid PR number: %s", prNumberStr)
	}

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	command := os.Args[2:]
    fmt.Println("Command: ", command)
	cmd.ExecuteAndComment(ctx, client, graphqlClient, owner, repo, prNumber, command)
}
