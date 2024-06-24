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
	if len(os.Args) < 5 || (os.Args[1] == "exec" && len(os.Args) < 6) {
		log.Fatalf("usage: %s <owner> <repo> <pr_number> <filename> or %s exec <owner> <repo> <pr_number> <command>", os.Args[0], os.Args[0])
	}

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	if os.Args[1] == "exec" {
		owner := os.Args[2]
		repo := os.Args[3]
		prNumber, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatalf("Invalid PR number: %s", os.Args[4])
		}
		command := os.Args[5:]
		cmd.ExecuteAndComment(ctx, client, graphqlClient, owner, repo, prNumber, command)
	} else {
		owner := os.Args[1]
		repo := os.Args[2]
		prNumber, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("Invalid PR number: %s", os.Args[3])
		}
		filename := os.Args[4]

		handler := cmd.NewCommandHandler(client, graphqlClient, owner, repo, prNumber)
		handler.HandleCommand(ctx, filename)
	}
}
