package botlib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultClientHomeDir = AppDataDir("pongclient", false)
	defaultBRDir         = AppDataDir("brclient", false)
)

type ClientConfig struct {
	ServerAddr     string
	RPCURL         string
	ServerCertPath string
	ClientCertPath string
	ClientKeyPath  string
	GRPCServerCert string
	RPCUser        string
	RPCPass        string
}

func writeClientConfigFile(cfg *ClientConfig, configPath string) error {
	configData := fmt.Sprintf(
		`serveraddr=%s
rpcurl=%s
servercertpath=%s
clientcertpath=%s
clientkeypath=%s
grpcservercert=%s
rpcuser=%s
rpcpass=%s
`,
		cfg.ServerAddr,
		cfg.RPCURL,
		cfg.ServerCertPath,
		cfg.ClientCertPath,
		cfg.ClientKeyPath,
		cfg.GRPCServerCert,
		cfg.RPCUser,
		cfg.RPCPass,
	)

	return os.WriteFile(configPath, []byte(configData), 0644)
}

func parseClientConfigFile(configPath string) (*ClientConfig, error) {
	const funcName = "parseClientConfigFile"

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open config file: %w", funcName, err)
	}
	defer file.Close()

	cfg := &ClientConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s: invalid line in config file: %s", funcName, line)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "serveraddr":
			cfg.ServerAddr = value
		case "rpcurl":
			cfg.RPCURL = value
		case "servercertpath":
			cfg.ServerCertPath = value
		case "clientcertpath":
			cfg.ClientCertPath = value
		case "clientkeypath":
			cfg.ClientKeyPath = value
		case "grpcservercert":
			cfg.GRPCServerCert = value
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

func LoadClientConfig() (*ClientConfig, error) {
	const funcName = "loadClientConfig"

	configDir := defaultClientHomeDir
	configPath := filepath.Join(configDir, "pongclient.conf")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("%s: failed to create config directory: %w", funcName, err)
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &ClientConfig{
			ServerAddr:     "localhost:50051",
			RPCURL:         "wss://127.0.0.1:7676/ws",
			ServerCertPath: filepath.Join(defaultBRDir, "rpc.cert"),
			ClientCertPath: filepath.Join(defaultBRDir, "rpc-client.cert"),
			ClientKeyPath:  filepath.Join(defaultBRDir, "rpc-client.key"),
			GRPCServerCert: filepath.Join(configDir, "grpc-server.cert"),
			RPCUser:        "defaultuser",
			RPCPass:        "defaultpass",
		}

		if err := writeClientConfigFile(defaultConfig, configPath); err != nil {
			return nil, fmt.Errorf("%s: failed to write default config file: %w", funcName, err)
		}
	}

	cfg, err := parseClientConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse config file: %w", funcName, err)
	}

	cfg.ServerCertPath = CleanAndExpandPath(cfg.ServerCertPath)
	cfg.ClientCertPath = CleanAndExpandPath(cfg.ClientCertPath)
	cfg.ClientKeyPath = CleanAndExpandPath(cfg.ClientKeyPath)
	cfg.GRPCServerCert = CleanAndExpandPath(cfg.GRPCServerCert)

	return cfg, nil
}
