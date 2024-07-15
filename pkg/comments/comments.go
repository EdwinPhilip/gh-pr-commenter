package comments

import (
	"context"
	"fmt"
	"strings"
	"os"
	"log"

	"gh-pr-commenter/internal"
	"gh-pr-commenter/config"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
)

const maxCommentLength = 55000

// Comment posts comments on the specified PR
func Comment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) error {
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		return fmt.Errorf("empty command")
	}
	cmdName := cmdArgs[0]
	config.Init(cmdName)
	cnf, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	outputFilename := fmt.Sprintf("%s/.output-%s.md", cnf.TmpGhpcDir ,cmdName)
	output, err := os.ReadFile(outputFilename)
	if err != nil {
		return fmt.Errorf("error reading output file: %w", err)
	}
	log.Printf("Output file read successfully")
	log.Printf("Output: %s", output)

	parts := splitMessage(string(output))

	err = createDefaultTemplate(cnf.TemplateFilename, command)
	if err != nil {
		return fmt.Errorf("error creating default template: %w", err)
	}

	templateContent, err := os.ReadFile(cnf.TemplateFilename)
	if err != nil {
		return fmt.Errorf("error reading template file: %w", err)
	}

	for i, part := range parts {
		partWithID := strings.Replace(string(templateContent), "---OUTPUT---", part, 1)
		partWithID = fmt.Sprintf("## %s output\n%s <!-- Part #%d -->", cmdName, partWithID, i+1)
		newFilename := fmt.Sprintf(".comment-%s-%s-%s-part-%d.md", repo, prNumber, cmdName, i+1)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}

		err = internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename, fmt.Sprintf("## %s output", cmdName), fmt.Sprintf("Part #%d", i+1))
		if err != nil {
			return fmt.Errorf("error upserting comment: %w", err)
		}
	}
	return nil
}

func splitMessage(message string) []string {
	var parts []string
	start := 0
	for start < len(message) {
		end := start + maxCommentLength
		if end >= len(message) {
			end = len(message)
		} else {
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

func createDefaultTemplate(filename string, command string) error {
    content := `
<details><summary>Show Output</summary>

`+"```"+`diff
---OUTPUT---
`+"```"+`
</details>
`
	if strings.Contains(command, "trivy") || strings.Contains(command, "tflint") {
		content = `---OUTPUT---`
	}
    return os.WriteFile(filename, []byte(content), 0644)
}
