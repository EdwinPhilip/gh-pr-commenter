package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gh-pr-commenter/internal"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxCommentLength = 55000

// ExecuteAndComment runs the provided command, captures its output, and posts it as a comment on the PR
func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber int, command string) {
	// Split the command into command and arguments
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		log.Printf("Empty command")
		return
	}
	cmdName := cmdArgs[0]
	cmdArgs = cmdArgs[1:]

	// Execute the provided command and capture its output
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	// Always capture command output, even if there's an error
	output := out.String()

	// Handle errors and log them
	if err != nil {
		log.Printf("Error running command: %v\n", err)
		output += fmt.Sprintf("\nError running command: %v\n", err)
	}

	// Split output if it exceeds maxCommentLength
	parts := splitMessage(output)

	// Read the template file
	templateContent, err := os.ReadFile("template.md")
	if err != nil {
		log.Printf("Error reading template file: %v\n", err)
		return
	}

	// Create new markdown files with the combined content and post each as a comment
	for i, part := range parts {
		partWithID := strings.Replace(string(templateContent), "---OUTPUT---", part, 1)
		partWithID = fmt.Sprintf("## %s output\n%s <!-- Part #%d -->", cmdName, partWithID, i+1)
		newFilename := fmt.Sprintf(".comment-%s-%d-%s-part-%d.md", repo, prNumber, cmdName, i+1)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		// Use the existing logic to post the comment
		internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename, fmt.Sprintf("## %s output", cmdName), fmt.Sprintf("Part #%d", i+1))
	}
}

// splitMessage splits the message into parts each with a maximum length of maxCommentLength
func splitMessage(message string) []string {
	var parts []string
	for i := 0; i < len(message); i += maxCommentLength {
		end := i + maxCommentLength
		if end > len(message) {
			end = len(message)
		}
		part := message[i:end]
		parts = append(parts, part)
	}
	return parts
}
