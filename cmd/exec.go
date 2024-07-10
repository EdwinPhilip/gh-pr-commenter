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

	"gh-pr-commenter/internal"
	"gh-pr-commenter/config"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxCommentLength = 55000

// ExecuteAndComment runs the provided command, captures its output, and posts it as a comment on the PR
func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) {
	// Split the command into command and arguments
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
	internal.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "pending", cnf.GHStatusContext)
	// Execute the provided command and capture its output
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()

	// Always capture command output, even if there's an error
	output := out.String()

	// Handle errors and log them
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
	newFilename := fmt.Sprintf(".output-%s.md", cmdName)
	err = os.WriteFile(newFilename, []byte(output), 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	// sleep for 5 seconds to allow the comment to be posted
	time.Sleep(5 * time.Second)
	if outputExitCode == 0 {
		internal.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "success", cnf.GHStatusContext)
		return
	}
	internal.PostCommitStatus(ctx, client, owner, repo, cnf.HeadCommit, "failure", cnf.GHStatusContext)
}

// splitMessage splits the message into parts each with a maximum length of maxCommentLength
func splitMessage(message string) []string {
	var parts []string
	start := 0
	for start < len(message) {
		// Determine the end index, but don't split in the middle of a line
		end := start + maxCommentLength
		if end >= len(message) {
			end = len(message)
		} else {
			// Look for the last newline character within the maxCommentLength
			lastNewline := strings.LastIndex(message[start:end], "\n")
			if lastNewline != -1 {
				end = start + lastNewline + 1
			}
		}
		part := message[start:end]
		parts = append(parts, part)
		start = end
	}
	return parts
}

// createDefaultTemplate creates a default template.md file with initial content
func createDefaultTemplate(filename string, command string) error {
    content := `
<details><summary>Show Output</summary>

`+"```"+`diff
---OUTPUT---
`+"```"+`
</details>
`
	// if command contains trivy or tflint, use empty template as content
	if strings.Contains(command, "trivy") {
		content = `---OUTPUT---`
	}
	if strings.Contains(command, "tflint") {
		content = "```"+`diff
---OUTPUT---
`+"```\n"
	}
    return os.WriteFile(filename, []byte(content), 0644)
}