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

func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) error {
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		return fmt.Errorf("empty command")
	}
	cmdName := cmdArgs[0]
	cmdArgs = cmdArgs[1:]
	outputExitCode := 1
	config.Init(cmdName)
	cnf := config.GetConfig()
	err := status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "pending", cnf.GHStatusContext)
	if err != nil {
		return fmt.Errorf("error posting commit status: %w", err)
	}
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
			return fmt.Errorf("error opening file: %w", err)
		}
	} else {
		file, err = os.Create(newFilename)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
	}
	defer file.Close()

	if _, err := file.WriteString(output); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	time.Sleep(5 * time.Second)
	if outputExitCode == 0 {
		err = status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "success", cnf.GHStatusContext)
		if err != nil {
			return fmt.Errorf("error posting success status: %w", err)
		}
		return nil
	}
	err = status.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "failure", cnf.GHStatusContext)
	if err != nil {
		return fmt.Errorf("error posting failure status: %w", err)
	}
	return nil
}
