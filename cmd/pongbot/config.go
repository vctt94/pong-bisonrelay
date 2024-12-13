package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	defaultHomeDir     = AppDataDir("pongbot", false)
	defaultBRClientDir = AppDataDir("brclient", false)
)

type config struct {
	DataDir        string
	URL            string
	ServerCertPath string
	ClientCertPath string
	ClientKeyPath  string
	RPCUser        string
	RPCPass        string
	Debug          string
}

func writeConfigFile(cfg *config, configPath string) error {
	configData := fmt.Sprintf(
		`datadir=%s
url=%s
servercertpath=%s
clientcertpath=%s
clientkeypath=%s
rpcuser=%s
rpcpass=%s
debug=%s
`,
		cfg.DataDir,
		cfg.URL,
		cfg.ServerCertPath,
		cfg.ClientCertPath,
		cfg.ClientKeyPath,
		cfg.RPCUser,
		cfg.RPCPass,
		cfg.Debug,
	)

	return os.WriteFile(configPath, []byte(configData), 0644)
}

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func parseConfigFile(configPath string) (*config, error) {
	const funcName = "parseConfigFile"

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open config file: %w", funcName, err)
	}
	defer file.Close()

	cfg := &config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s: invalid line in config file: %s", funcName, line)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0])) // Convert key to lowercase
		value := strings.TrimSpace(parts[1])

		switch key {
		case "datadir":
			cfg.DataDir = value
		case "debug":
			cfg.Debug = value
		case "url":
			cfg.URL = value
		case "servercertpath":
			cfg.ServerCertPath = value
		case "clientcertpath":
			cfg.ClientCertPath = value
		case "clientkeypath":
			cfg.ClientKeyPath = value
		case "rpcuser":
			cfg.RPCUser = value
		case "rpcpass":
			cfg.RPCPass = value
		default:
			return nil, fmt.Errorf("%s: unknown config key: %s", funcName, key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: error reading config file: %w", funcName, err)
	}

	return cfg, nil
}

func loadConfig() (*config, error) {
	const funcName = "loadConfig"

	// Path to the configuration directory and file
	configDir := defaultHomeDir
	configPath := filepath.Join(configDir, "pongbot.conf")

	// Ensure the configuration directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("%s: failed to create config directory: %w", funcName, err)
		}
	}

	// If the config file does not exist, create it with default values
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		rpcUser, err := GenerateRandomString(8)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate rpcuser: %w", funcName, err)
		}

		rpcPass, err := GenerateRandomString(8)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate rpcpass: %w", funcName, err)
		}

		defaultConfig := config{
			DataDir: configDir,
			Debug:   "debug",

			URL:            "wss://127.0.0.1:7676/ws",
			ServerCertPath: filepath.Join(defaultBRClientDir, "rpc.cert"),
			ClientCertPath: filepath.Join(defaultBRClientDir, "rpc-client.cert"),
			ClientKeyPath:  filepath.Join(defaultBRClientDir, "rpc-client.key"),
			RPCUser:        rpcUser,
			RPCPass:        rpcPass,
		}

		// Write the default configuration to the file
		err = writeConfigFile(&defaultConfig, configPath)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to write default config: %w", funcName, err)
		}
	}

	// Parse the configuration file
	cfg, err := parseConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse config file: %w", funcName, err)
	}

	// Clean and expand paths
	cfg.DataDir = cleanAndExpandPath(cfg.DataDir)
	cfg.ServerCertPath = cleanAndExpandPath(cfg.ServerCertPath)
	cfg.ClientCertPath = cleanAndExpandPath(cfg.ClientCertPath)
	cfg.ClientKeyPath = cleanAndExpandPath(cfg.ClientKeyPath)

	return cfg, nil
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
