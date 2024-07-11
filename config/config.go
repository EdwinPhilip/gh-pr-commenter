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
	config = &Config{}
	config.HeadCommit = os.Getenv("HEAD_COMMIT")
	config.ProjectName = os.Getenv("PROJECT_NAME")
	config.GHStatusContext = os.Getenv("GH_STATUS_CONTEXT")
	config.Workspace = os.Getenv("WORKSPACE")
	config.BaseRepoOwner = os.Getenv("BASE_REPO_OWNER")
	config.BaseRepoName = os.Getenv("BASE_REPO_NAME")
	config.PullNum = os.Getenv("PULL_NUM")
	config.TemplateFilename = os.Getenv("TEMPLATE_FILENAME")
	config.GithubToken = os.Getenv("GITHUB_TOKEN")
	config.TmpGhpcDir = os.Getenv("TMP_GHPC_DIR")
	if config.TmpGhpcDir == "" {
		config.TmpGhpcDir = "/tmp/ghpc"
	}
	if config.ProjectName != "" && config.Workspace != "" {
		config.ProjectRunDetails = fmt.Sprintf("<h3>Project: <code>%s</code> Workspace: <code>%s</code><h3>\n", config.ProjectName, config.Workspace)
		config.ProjectIdentifier = fmt.Sprintf("%s-%s", config.ProjectName, config.Workspace)
	} else {
		config.ProjectName = "atlantis"
		config.Workspace = "default"
	}
	if config.GHStatusContext != "" && config.ProjectName != "" {
		config.GHStatusContext = config.GHStatusContext + "/" + cmdName + ": " + config.ProjectName
	} else {
		config.GHStatusContext = "ghpc" + "/" + cmdName
	}
	if config.TemplateFilename == "" {
		config.TemplateFilename = "template.md"
	}
}

func GetConfig() (*Config, error) {
	// Check if any of the config fields are empty, return an error if any field is empty
	if config.HeadCommit == "" {
		return nil, errors.New("required environment variable HEAD_COMMIT not set")
	}
	if config.ProjectName == "" {
		return nil, errors.New("required environment variable PROJECT_NAME not set")
	}
	if config.GHStatusContext == "" {
		return nil, errors.New("required environment variable GH_STATUS_CONTEXT not set")
	}
	if config.Workspace == "" {
		return nil, errors.New("required environment variable WORKSPACE not set")
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
	if config.TemplateFilename == "" {
		return nil, errors.New("required environment variable TEMPLATE_FILENAME not set")
	}
	if config.GithubToken == "" {
		return nil, errors.New("required environment variable GITHUB_TOKEN not set")
	}
	return config, nil
}
