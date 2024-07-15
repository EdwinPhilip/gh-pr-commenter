package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultProjectName  = "atlantis"
	DefaultWorkspace    = "default"
	DefaultTemplateFile = "template.md"
	DefaultTmpGhpcDir   = "/tmp/ghpc"
)

type Config struct {
	HeadCommit        string
	ProjectName       string
	GHStatusContext   string
	Workspace         string
	BaseRepoOwner     string
	BaseRepoName      string
	PullNum           string
	TemplateFilename  string
	GithubToken       string
	ProjectRunDetails string
	ProjectIdentifier string
	TmpGhpcDir        string
}

var config *Config

func Init(cmdName string) {
	viper.AutomaticEnv()

	viper.SetDefault("PROJECT_NAME", DefaultProjectName)
	viper.SetDefault("WORKSPACE", DefaultWorkspace)
	viper.SetDefault("TEMPLATE_FILENAME", DefaultTemplateFile)
	viper.SetDefault("TMP_GHPC_DIR", DefaultTmpGhpcDir)

	config = &Config{
		HeadCommit:       viper.GetString("HEAD_COMMIT"),
		ProjectName:      viper.GetString("PROJECT_NAME"),
		GHStatusContext:  viper.GetString("GH_STATUS_CONTEXT"),
		Workspace:        viper.GetString("WORKSPACE"),
		BaseRepoOwner:    viper.GetString("BASE_REPO_OWNER"),
		BaseRepoName:     viper.GetString("BASE_REPO_NAME"),
		PullNum:          viper.GetString("PULL_NUM"),
		TemplateFilename: viper.GetString("TEMPLATE_FILENAME"),
		GithubToken:      viper.GetString("GITHUB_TOKEN"),
		TmpGhpcDir:       viper.GetString("TMP_GHPC_DIR"),
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

	validateConfig()
}

func validateConfig() {
	requiredKeys := []string{
		"HEAD_COMMIT", "BASE_REPO_OWNER", "BASE_REPO_NAME", "PULL_NUM", "GITHUB_TOKEN",
	}

	missingKeys := []string{}
	for _, key := range requiredKeys {
		if !viper.IsSet(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		log.Fatalf("Missing required environment variables: %s", strings.Join(missingKeys, ", "))
	}
}

func GetConfig() *Config {
	return config
}
