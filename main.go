package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/machinebox/graphql"
	"gh-pr-commenter/cmd"
	"gh-pr-commenter/config"
	"gh-pr-commenter/internal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	// This variable will be set during the build process
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "ghpc",
	Short: "GitHub PR Commenter is a tool to automate command execution and commenting on GitHub PRs",
	Long: `GitHub PR Commenter (ghpc) automates the process of executing commands and posting their output
as comments on GitHub Pull Requests (PRs). Inspired by Atlantis, ghpc streamlines the integration of
command execution and result reporting within the PR workflow.`,
}

var execCmd = &cobra.Command{
	Use:   "exec [command]",
	Short: "Execute a command and capture its output",
	Long:  `Executes the specified command and captures its output in a file.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeCommand("exec", args)
	},
}

var commentCmd = &cobra.Command{
	Use:   "comment [command]",
	Short: "Post the captured output as a PR comment",
	Long:  `Reads the captured output file and posts its content as a comment on the specified pull request.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeCommand("comment", args)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ghpc",
	Long:  `All software has versions. This is ghpc's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GitHub PR Commenter (ghpc) " + version)
	},
}

func main() {
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(commentCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		logger := config.GetLogger()
		if logger != nil {
			logger.Fatal("Failed to execute root command", zap.Error(err))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to execute root command: %v\n", err)
			os.Exit(1)
		}
	}
}

func executeCommand(runCommand string, args []string) {
	command := strings.Join(args, " ")
	cmdName := args[0]

	if len(cmdName) == 0 {
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
