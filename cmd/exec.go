package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/machinebox/graphql"
	"gh-pr-commenter/internal"
)

const maxCommentLength = 55000

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
	//output := out.String()
	output := "Creating a paragraph with more than 65,000 characters without any line breaks is impractical and not readable. Here is a substantial paragraph about Go language instead: Go, often referred to as Golang, is a statically typed, compiled programming language designed by Google. It is known for its simplicity, efficiency, and strong support for concurrent programming. The language was created by Robert Griesemer, Rob Pike, and Ken Thompson and was first released in 2009. Go's syntax is clean and concise, making it easy for developers to learn and use. One of the standout features of Go is its powerful concurrency model, which is based on goroutines and channels. Goroutines are lightweight threads that are managed by the Go runtime, allowing developers to run thousands of them concurrently without significant performance overhead. Channels provide a way for goroutines to communicate with each other, making it easier to build concurrent and parallel applications. Go also has a garbage collector that helps manage memory automatically, reducing the likelihood of memory leaks and other related issues. The language's standard library is extensive, offering a wide range of built-in functions and packages that simplify tasks such as I/O operations, string manipulation, and networking. Additionally, Go's tooling, including the gofmt tool for automatic code formatting, contributes to the language's overall productivity and maintainability. One of the reasons for Go's popularity is its performance. As a compiled language, Go translates code directly into machine code, resulting in fast execution times. This makes it suitable for high-performance applications, such as web servers, networked services, and large-scale distributed systems. The language's simplicity and efficiency have made it a favorite among developers working in various domains, including cloud computing, microservices, and DevOps. Companies like Google, Dropbox, and Docker have adopted Go for its ability to handle large-scale applications with ease. Despite its many strengths, Go has some limitations. Its error handling mechanism, which relies on explicit error checks, can lead to verbose code. Additionally, Go does not support generics, which means developers often need to write repetitive code for different data types. However, the Go community is actively working on addressing these issues, and proposals for adding generics and improving error handling are in progress. Overall, Go's combination of simplicity, performance, and powerful concurrency features make it an excellent choice for modern software development. Its growing ecosystem and active community ensure that it will continue to evolve and remain relevant in the years to come."

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
	parts := splitMessage(commentContent, title)

	// Create new markdown files with the combined content and post each as a comment
	for i, part := range parts {
		partWithID := fmt.Sprintf("%s <!-- Part #%d -->", part, i+1)
		newFilename := fmt.Sprintf(".comment-%s-%d-%s-part-%d.md", repo, prNumber, cmdName, i+1)
		err := os.WriteFile(newFilename, []byte(partWithID), 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}

		// Use the existing logic to post the comment
		internal.UpsertComment(ctx, client, graphqlClient, owner, repo, prNumber, newFilename, title, fmt.Sprintf("Part #%d", i+1))
	}
}

// splitMessage splits the message into parts each with a maximum length of maxCommentLength
func splitMessage(message, title string) []string {
	var parts []string
	for i := 0; i < len(message); i += maxCommentLength {
		end := i + maxCommentLength
		if end > len(message) {
			end = len(message)
		}
		part := message[i:end]
		part = fmt.Sprintf("%s\n%s\n", title, part)
		parts = append(parts, part)
	}
	return parts
}
