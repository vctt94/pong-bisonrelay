package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

var defaultHomeDir = AppDataDir("pongbot", false)

type config struct {
	DataDir        string `yaml:"DataDir"`
	URL            string `yaml:"URL"`
	ServerCertPath string `yaml:"ServerCertPath"`
	ClientCertPath string `yaml:"ClientCertPath"`
	ClientKeyPath  string `yaml:"ClientKeyPath"`
}

func loadConfig() (*config, error) {
	const funcName = "loadConfig"

	configPath := filepath.Join(defaultHomeDir, "pongbot.conf")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		str := "%s: %w"
		err := fmt.Errorf(str, funcName, err)
		return nil, err
	}
	cfg := config{
		DataDir: defaultHomeDir,
	}
	if err = yaml.UnmarshalStrict(configFile, &cfg); err != nil {
		str := "%s: failed to parse config file: %w"
		err := fmt.Errorf(str, funcName, err)
		return nil, err
	}

	cfg.DataDir = cleanAndExpandPath(cfg.DataDir)
	cfg.ServerCertPath = cleanAndExpandPath(cfg.ServerCertPath)
	cfg.ClientCertPath = cleanAndExpandPath(cfg.ClientCertPath)
	cfg.ClientKeyPath = cleanAndExpandPath(cfg.ClientKeyPath)

	if cfg.DataDir == "" {
		cfg.DataDir = defaultHomeDir
	}

	return &cfg, nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Nothing to do when no path is given.
	if path == "" {
		return path
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows cmd.exe-style
	// %VARIABLE%, but the variables can still be expanded via POSIX-style
	// $VARIABLE.
	path = os.ExpandEnv(path)

	if !strings.HasPrefix(path, "~") {
		return filepath.Clean(path)
	}

	// Expand initial ~ to the current user's home directory, or ~otheruser
	// to otheruser's home directory.  On Windows, both forward and backward
	// slashes can be used.
	path = path[1:]

	var pathSeparators string
	if runtime.GOOS == "windows" {
		pathSeparators = string(os.PathSeparator) + "/"
	} else {
		pathSeparators = string(os.PathSeparator)
	}

	userName := ""
	if i := strings.IndexAny(path, pathSeparators); i != -1 {
		userName = path[:i]
		path = path[i:]
	}

	homeDir := ""
	var u *user.User
	var err error
	if userName == "" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(userName)
	}
	if err == nil {
		homeDir = u.HomeDir
	}
	// Fallback to CWD if user lookup fails or user has no home directory.
	if homeDir == "" {
		homeDir = "."
	}

	return filepath.Join(homeDir, path)
}
