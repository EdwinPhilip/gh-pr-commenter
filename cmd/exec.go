package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gh-pr-commenter/config"
	"gh-pr-commenter/pkg/status"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxCommentLength = 55000

func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) {
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		log.Printf("Empty command")
		return
	}
	cmdName := cmdArgs[0]
	cmdArgs = cmdArgs[1:]
	outputExitCode := 1
	config.Init(cmdName)
	cnf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error getting config: %v", err)
	}
	status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "pending", cnf.GHStatusContext)
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()

	output := out.String()

	if err != nil {
		log.Printf("Error running command: %v\n", err)
		output += fmt.Sprintf("\nError running command: %v\n", err)
	}
	if (output == "" || strings.Contains(output, "passed")) && err == nil {
		outputExitCode = 0
		if cmdName == "tflint" {
			output = fmt.Sprintf("%s passed.\n\nNo output was generated.", cmdName)
		}
	}
	output = fmt.Sprintf("\n%s\n%s\n\n---\n", cnf.ProjectRunDetails, output)
	newFilename := fmt.Sprintf("%s/.output-%s.md", cnf.TmpGhpcDir ,cmdName)
	
	fileExists := false
	if _, err := os.Stat(newFilename); err == nil {
		fileExists = true
	}
	
	var file *os.File
	if fileExists {
		file, err = os.OpenFile(newFilename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
	} else {
		file, err = os.Create(newFilename)
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}
	}
	defer file.Close()

	if _, err := file.WriteString(output); err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	time.Sleep(5 * time.Second)
	if outputExitCode == 0 {
		status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "success", cnf.GHStatusContext)
		return
	}
	status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "failure", cnf.GHStatusContext)
}
