package main

import (
	"context"
	"os"
	"strings"
	"fmt"

	"github.com/machinebox/graphql"
	"gh-pr-commenter/cmd"
	"gh-pr-commenter/config"
	"gh-pr-commenter/internal"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 3 {
		config.GetLogger().Fatal("usage", zap.String("usage", fmt.Sprintf("%s exec|comment <command>", os.Args[0])))
	}
	runCommand := os.Args[1]
	command := os.Args[2]
	cmdArgs := strings.Fields(command)
	cmdName := cmdArgs[0]
	if len(cmdArgs) == 0 {
		config.GetLogger().Warn("Empty command")
		return
	}
	config.Init(cmdName)
	cnf := config.GetConfig()

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	var err error
	switch runCommand {
	case "exec":
		err = cmd.ExecuteAndComment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, command)
	case "comment":
		err = cmd.Comment(ctx, client, graphqlClient, cnf.BaseRepoOwner, cnf.BaseRepoName, cnf.PullNum, command)
	default:
		config.GetLogger().Fatal("unknown command", zap.String("command", runCommand))
	}
	if err != nil {
		config.GetLogger().Fatal("Error executing command", zap.Error(err))
	}
}
