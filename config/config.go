package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	HeadCommit       string
	ProjectName      string
	GHStatusContext  string
	Workspace        string
	BaseRepoOwner    string
	BaseRepoName     string
	PullNum          string
	TemplateFilename string
	GithubToken      string
	ProjectRunDetails string
	ProjectIdentifier string
	TmpGhpcDir       string
}

var config *Config

func Init(cmdName string) {
	config = &Config{
		HeadCommit:       getEnv("HEAD_COMMIT", ""),
		ProjectName:      getEnv("PROJECT_NAME", "atlantis"),
		GHStatusContext:  getEnv("GH_STATUS_CONTEXT", ""),
		Workspace:        getEnv("WORKSPACE", "default"),
		BaseRepoOwner:    getEnv("BASE_REPO_OWNER", ""),
		BaseRepoName:     getEnv("BASE_REPO_NAME", ""),
		PullNum:          getEnv("PULL_NUM", ""),
		TemplateFilename: getEnv("TEMPLATE_FILENAME", "template.md"),
		GithubToken:      getEnv("GITHUB_TOKEN", ""),
		TmpGhpcDir:       getEnv("TMP_GHPC_DIR", "/tmp/ghpc"),
	}

	if config.ProjectName != "" && config.Workspace != "" {
		config.ProjectRunDetails = fmt.Sprintf("<h3>Project: <code>%s</code> Workspace: <code>%s</code></h3>\n", config.ProjectName, config.Workspace)
		config.ProjectIdentifier = fmt.Sprintf("%s-%s", config.ProjectName, config.Workspace)
	}
	if config.GHStatusContext != "" && config.ProjectName != "" {
		config.GHStatusContext = config.GHStatusContext + "/" + cmdName + ": " + config.ProjectName
	} else {
		config.GHStatusContext = "ghpc" + "/" + cmdName
	}
}

func GetConfig() (*Config, error) {
	if config.HeadCommit == "" {
		return nil, errors.New("required environment variable HEAD_COMMIT not set")
	}
	if config.BaseRepoOwner == "" {
		return nil, errors.New("required environment variable BASE_REPO_OWNER not set")
	}
	if config.BaseRepoName == "" {
		return nil, errors.New("required environment variable BASE_REPO_NAME not set")
	}
	if config.PullNum == "" {
		return nil, errors.New("required environment variable PULL_NUM not set")
	}
	if config.GithubToken == "" {
		return nil, errors.New("required environment variable GITHUB_TOKEN not set")
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
