package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
	"gh-pr-commenter/internal"
)

// ExecuteAndComment runs the provided command, captures its output, and posts it as a comment on the PR
func ExecuteAndComment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber int, command []string) {
	// Execute the provided command and capture its output
	cmd := exec.Command(command[0], command[1:]...)
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
	title := fmt.Sprintf("## %s Output\n", command[0])
	commentContent = title + commentContent

	// Create a new markdown file with the combined content
	commandName := strings.Split(command[0], "/")
	newFilename := fmt.Sprintf("%s-%d-%s.md", repo, prNumber, commandName[len(commandName)-1])
	err = os.WriteFile(newFilename, []byte(commentContent), 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	// Read the new markdown file content
	commentMessage, err := internal.ReadCommentFromFile(newFilename)
	if err != nil {
		log.Fatalf("Error reading comment from file: %v", err)
	}

	// Use the existing logic to post the comment
	internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, commentMessage)
}
