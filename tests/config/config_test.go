package config_test

import (
	"os"
	"fmt"
	"testing"

	"gh-pr-commenter/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInit_WithDefaults(t *testing.T) {
	os.Clearenv()
	viper.Reset()

	envVars := map[string]string{
		"HEAD_COMMIT":     "test-commit",
		"BASE_REPO_OWNER": "test-owner",
		"BASE_REPO_NAME":  "test-repo",
		"PULL_NUM":        "123",
		"GITHUB_TOKEN":    "test-token",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		t.Logf("Set %s=%s", key, value)
		fmt.Println(os.Getenv(key))
	}

	viper.AutomaticEnv()

	config.Init("test-cmd")

	cnf := config.GetConfig()
	assert.Equal(t, config.DefaultProjectName, cnf.ProjectName)
	assert.Equal(t, config.DefaultWorkspace, cnf.Workspace)
	assert.Equal(t, config.DefaultTemplateFile, cnf.TemplateFilename)
	assert.Equal(t, config.DefaultTmpGhpcDir, cnf.TmpGhpcDir)
}

func TestInit_WithEnvVariables(t *testing.T) {
	os.Clearenv()
	viper.Reset()

	envVars := map[string]string{
		"HEAD_COMMIT":      "test-commit",
		"PROJECT_NAME":     "test-project",
		"GH_STATUS_CONTEXT": "test-context",
		"WORKSPACE":        "test-workspace",
		"BASE_REPO_OWNER":  "test-owner",
		"BASE_REPO_NAME":   "test-repo",
		"PULL_NUM":         "123",
		"GITHUB_TOKEN":     "test-token",
		"TEMPLATE_FILENAME": "test-template.md",
		"TMP_GHPC_DIR":     "/tmp/test-ghpc",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		t.Logf("Set %s=%s", key, value)
	}

	viper.AutomaticEnv() // Call before config.Init

	config.Init("test-cmd")

	cnf := config.GetConfig()
	assert.Equal(t, "test-commit", cnf.HeadCommit)
	assert.Equal(t, "test-project", cnf.ProjectName)
	assert.Equal(t, "test-context/test-cmd: test-project", cnf.GHStatusContext)
	assert.Equal(t, "test-workspace", cnf.Workspace)
	assert.Equal(t, "test-owner", cnf.BaseRepoOwner)
	assert.Equal(t, "test-repo", cnf.BaseRepoName)
	assert.Equal(t, "123", cnf.PullNum)
	assert.Equal(t, "test-template.md", cnf.TemplateFilename)
	assert.Equal(t, "test-token", cnf.GithubToken)
	assert.Equal(t, "/tmp/test-ghpc", cnf.TmpGhpcDir)
}

func TestValidateConfig_MissingKeys(t *testing.T) {
	os.Clearenv()
	viper.Reset()

	envVars := map[string]string{
		"HEAD_COMMIT":     "test-commit",
		"BASE_REPO_OWNER": "test-owner",
		"BASE_REPO_NAME":  "test-repo",
		"PULL_NUM":        "123",
		"GITHUB_TOKEN":    "test-token",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		t.Logf("Set %s=%s", key, value)
		fmt.Println(os.Getenv(key))
	}

	viper.AutomaticEnv() // Call before config.Init

	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r, "Missing required environment variables")
		}
	}()

	config.Init("test-cmd")
	config.ValidateConfig()
}
