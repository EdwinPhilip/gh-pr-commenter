package cmd

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

func Comment(ctx context.Context, client *github.Client, graphqlClient *graphql.Client, owner, repo string, prNumber string, command string) {
	// Split the command into command and arguments
	cmdArgs := strings.Fields(command)
	if len(cmdArgs) == 0 {
		log.Printf("Empty command")
		return
	}
	cmdName := cmdArgs[0]
	config.Init(cmdName)
	cnf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error getting config: %v", err)
	}
	outputFilename := fmt.Sprintf("%s/.output-%s.md", cnf.TmpGhpcDir ,cmdName)
	output, err := os.ReadFile(outputFilename)
	if err != nil {
		log.Fatalf("Error reading output file: %v", err)
	} else {
		log.Printf("Output file read successfully")
		log.Printf("Output: %s", output)
	}

	// Split output if it exceeds maxCommentLength
	parts := splitMessage(string(output))

	err = createDefaultTemplate(cnf.TemplateFilename, command)
	if err != nil {
		log.Fatalf("Error creating default template: %v", err)
	}

	// Read the template file
	templateContent, err := os.ReadFile(cnf.TemplateFilename)
	if err != nil {
		log.Printf("Error reading template file: %v\n", err)
		return
	}

	for i, part := range parts {
		partWithID := strings.Replace(string(templateContent), "---OUTPUT---", part, 1)
		partWithID = fmt.Sprintf("## %s output\n%s <!-- Part #%d -->", cmdName, partWithID, i+1)
		newFilename := fmt.Sprintf(".comment-%s-%s-%s-part-%d.md", repo, prNumber, cmdName, i+1)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		// Use the existing logic to post the comment
		internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename, fmt.Sprintf("## %s output", cmdName), fmt.Sprintf("Part #%d", i+1))
	}

}
