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

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
	"gh-pr-commenter/internal"
)

const maxCommentLength = 55000
const identifierPrefix = "<!-- Part"

// ExecuteAndComment runs the provided command, captures its output, and posts it as a comment on the PR
func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber int, command string) {
	// Split the command into command and arguments
	cmdArgs := strings.Fields(command)
	cmdName := cmdArgs[0]
	cmdArgs = cmdArgs[1:]

	// Execute the provided command and capture its output
	cmd := exec.Command(cmdName, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
	output := out.String()

	// Read the template file
	templateContent, err := os.ReadFile("template.md")
	if err != nil {
		log.Fatalf("Error reading template file: %v", err)
	}

	// Replace ---OUTPUT--- with the actual command output
	commentContent := strings.Replace(string(templateContent), "---OUTPUT---", output, 1)

	// Prepend the title to the comment content
	title := fmt.Sprintf("### %s output", cmdName)
	commentContent = title + "\n" + commentContent

	// Split message if it exceeds maxCommentLength
	parts := splitMessage(commentContent)

	// Create new markdown files with the combined content and post each as a comment
	for i, part := range parts {
		partWithID := fmt.Sprintf("%s <!-- Part #%d -->", part, i+1)
		newFilename := fmt.Sprintf(".comment-%s-%d-%s-part-%d.md", repo, prNumber, cmdName, i+1)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		// Use the existing logic to post the comment
		internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename)
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
