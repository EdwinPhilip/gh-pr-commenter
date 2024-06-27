package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/machinebox/graphql"
	"gh-pr-commenter/cmd"
	"gh-pr-commenter/internal"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "exec" {
		log.Fatalf("usage: %s exec <command>", os.Args[0])
	}

	owner := os.Getenv("BASE_REPO_OWNER")
	repo := os.Getenv("BASE_REPO_NAME")
	prNumberStr := os.Getenv("PULL_NUM")
	if owner == "" || repo == "" || prNumberStr == "" {
		log.Fatalf("Environment variables BASE_REPO_OWNER, BASE_REPO_NAME, and PULL_NUM must be set")
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		log.Fatalf("Invalid PR number: %s", prNumberStr)
	}
	command := os.Args[2]

	// always create template.md if it exists
	templateFilename := "template.md"
	err = createDefaultTemplate(templateFilename, command)
	if err != nil {
		log.Fatalf("Error creating default template: %v", err)
	}

	ctx := context.Background()
	client := internal.NewGitHubClient(ctx)
	graphqlClient := graphql.NewClient("https://api.github.com/graphql")

	cmd.ExecuteAndComment(ctx, client, graphqlClient, owner, repo, prNumber, command)
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
