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
	outputExitCode := 1 
	headCommit := os.Getenv("HEAD_COMMIT")
	project_name := os.Getenv("PROJECT_NAME")
	if headCommit == "" {
		log.Fatalf("HEAD_COMMIT environment variable not set")
	}
	ghStatusContext := os.Getenv("GH_STATUS_CONTEXT")
	if ghStatusContext == "" {
		log.Printf("GH_STATUS_CONTEXT environment variable not set")
		ghStatusContext = "ghpc" + "/" + cmdName
	} else {
		ghStatusContext = ghStatusContext + "/" + cmdName
	}
	if project_name != "" {
		ghStatusContext = ghStatusContext + ": " + project_name
	}
	internal.PostCommitStatus(ctx, client, owner, repo, headCommit, "pending", ghStatusContext)
	// Execute the provided command and capture its output
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	// Always capture command output, even if there's an error
	output := out.String()
	project_run_details := ""
	project_identifier := ""
	repo_rel_dir := os.Getenv("REPO_REL_DIR")
	workspace := os.Getenv("WORKSPACE")
	if project_name != "" && repo_rel_dir != "" && workspace != "" {
		project_run_details = fmt.Sprintf("project: `%s` dir: `%s` workspace: `%s`\n", project_name, repo_rel_dir, workspace)
		project_identifier = fmt.Sprintf("%s-%s", project_name, workspace)
	}

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

	// Split output if it exceeds maxCommentLength
	parts := splitMessage(output)

	templateFilename := os.Getenv("GHPC_TEMPLATE_FILE")
	if templateFilename == "" {
		templateFilename = "template.md"
		err = createDefaultTemplate(templateFilename, command)
		if err != nil {
			log.Fatalf("Error creating default template: %v", err)
		}
	}

	// Read the template file
	templateContent, err := os.ReadFile(templateFilename)
	if err != nil {
		log.Printf("Error reading template file: %v\n", err)
		return
	}

	// Create new markdown files with the combined content and post each as a comment
	for i, part := range parts {
		partWithID := strings.Replace(string(templateContent), "---OUTPUT---", part, 1)
		partWithID = fmt.Sprintf("## %s output\n\n%s\n%s <!-- Part #%d %s -->", cmdName, project_run_details, partWithID, i+1, project_identifier)
		newFilename := fmt.Sprintf(".comment-%s-%d-%s-part-%d-%s.md", repo, prNumber, cmdName, i+1, project_identifier)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		// Use the existing logic to post the comment
		internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename, fmt.Sprintf("## %s output", cmdName), fmt.Sprintf("Part #%d %s", i+1, project_identifier))
	}
	// sleep for 5 seconds to allow the comment to be posted
	time.Sleep(5 * time.Second)
	if outputExitCode == 0 {
		internal.PostCommitStatus(ctx, client, owner, repo, headCommit, "success", ghStatusContext)
		return
	}
	internal.PostCommitStatus(ctx, client, owner, repo, headCommit, "failure", ghStatusContext)
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