package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
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

var (
	config *Config
	logger *zap.Logger
)

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

	// Log the environment variables right before validation
	fmt.Println("HEAD_COMMIT:", config.HeadCommit)
	fmt.Println("BASE_REPO_OWNER:", config.BaseRepoOwner)
	fmt.Println("BASE_REPO_NAME:", config.BaseRepoName)
	fmt.Println("PULL_NUM:", config.PullNum)
	fmt.Println("GITHUB_TOKEN:", config.GithubToken)

	ValidateConfig()
	initLogger()
}

func ValidateConfig() {
	requiredKeys := []string{
		"HEAD_COMMIT", "BASE_REPO_OWNER", "BASE_REPO_NAME", "PULL_NUM", "GITHUB_TOKEN",
	}

	missingKeys := []string{}
	for _, key := range requiredKeys {
		if viper.GetString(key) == "" {
			missingKeys = append(missingKeys, key)
		} else {
			log.Printf("Environment variable %s is set to: %s", key, viper.GetString(key))
		}
	}

	if len(missingKeys) > 0 {
		log.Fatalf("Missing required environment variables: %s", strings.Join(missingKeys, ", "))
	}
}

func initLogger() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}
}

func GetConfig() *Config {
	return config
}

func GetLogger() *zap.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
