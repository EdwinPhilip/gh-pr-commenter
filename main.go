package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/machinebox/graphql"
	"gh-pr-commenter/cmd"
	"gh-pr-commenter/internal"
	"gh-pr-commenter/config"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s exec|comment <command>", os.Args[0])
	}
	runCommand := os.Args[1]
	command := os.Args[2]
	cmdArgs := strings.Fields(command)
	cmdName := cmdArgs[0]
	if len(cmdArgs) == 0 {
		log.Printf("Empty command")
		return
	}
	config.Init(cmdName)
	cnf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error getting config: %v", err)
	}

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	switch runCommand {
	case "exec":
		err = cmd.ExecuteAndComment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, command)
	case "comment":
		err = cmd.Comment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, command)
	default:
		log.Fatalf("unknown command: %s", runCommand)
	}
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
