package botlib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vctt94/pong-bisonrelay/ponggame"
)

var (
	defaultHomeDir     = AppDataDir("pongbot", false)
	defaultBRClientDir = AppDataDir("brclient", false)
)

type BotConfig struct {
	DataDir        string
	RPCURL         string
	GRPCHost       string
	GRPCPort       string
	ServerCertPath string
	ClientCertPath string
	ClientKeyPath  string
	RPCUser        string
	RPCPass        string
	Debug          string
}

func writeConfigFile(cfg *BotConfig, configPath string) error {
	configData := fmt.Sprintf(
		`datadir=%s
rpcurl=%s
grpchost=%s
grpcport=%s
servercertpath=%s
clientcertpath=%s
clientkeypath=%s
rpcuser=%s
rpcpass=%s
debug=%s
`,
		cfg.DataDir,
		cfg.RPCURL,
		cfg.GRPCHost,
		cfg.GRPCPort,
		cfg.ServerCertPath,
		cfg.ClientCertPath,
		cfg.ClientKeyPath,
		cfg.RPCUser,
		cfg.RPCPass,
		cfg.Debug,
	)

	return os.WriteFile(configPath, []byte(configData), 0644)
}

func parseConfigFile(configPath string) (*BotConfig, error) {
	const funcName = "parseConfigFile"

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open config file: %w", funcName, err)
	}
	defer file.Close()

	cfg := &BotConfig{}
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
		case "rpcurl":
			cfg.RPCURL = value
		case "grpchost":
			cfg.GRPCHost = value
		case "grpcport":
			cfg.GRPCPort = value
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

func LoadBotConfig() (*BotConfig, error) {
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
		rpcUser, err := ponggame.GenerateRandomString(8)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate rpcuser: %w", funcName, err)
		}

		rpcPass, err := ponggame.GenerateRandomString(8)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to generate rpcpass: %w", funcName, err)
		}

		defaultConfig := BotConfig{
			DataDir: configDir,
			Debug:   "debug",

			GRPCHost:       "localhost",
			GRPCPort:       "50051",
			RPCURL:         "wss://127.0.0.1:7676/ws",
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
	cfg.DataDir = CleanAndExpandPath(cfg.DataDir)
	cfg.ServerCertPath = CleanAndExpandPath(cfg.ServerCertPath)
	cfg.ClientCertPath = CleanAndExpandPath(cfg.ClientCertPath)
	cfg.ClientKeyPath = CleanAndExpandPath(cfg.ClientKeyPath)

	return cfg, nil
}
