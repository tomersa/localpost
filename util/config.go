package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const LocalpostDir = "lpost"
const RequestsDir = LocalpostDir + "/requests"
const ResponsesDir = LocalpostDir + "/responses"
const ConfigFile = "config.yaml"
const ConfigFilePath = LocalpostDir + "/" + ConfigFile

// configFile is an internal struct for parsing config.yaml.
type configFile struct {
	Env  string         `yaml:"env"`
	Envs map[string]Env `yaml:"envs"`
}

// CheckRepoContext verifies if the current directory contains a lpost/ folder.
func CheckRepoContext() error {
	// Check lpost/ directory
	if _, err := os.Stat("lpost"); os.IsNotExist(err) {
		return fmt.Errorf("localpost context not found (lpost directory).\nMake sure you running in the right directory or run 'lpost init' to init localpost.\n")
	}

	// Check config.yaml file
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		return fmt.Errorf("localpost context not found (%s).\nMake sure you running in the right directory or run 'lpost init' to init localpost.\n", ConfigFilePath)
	}

	// Read and validate config.yaml
	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", ConfigFilePath, err)
	}

	var config configFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid %s: %v", ConfigFilePath, err)
	}

	// Basic validation: ensure 'env' field exists
	if config.Env == "" {
		return fmt.Errorf("invalid %s: missing 'env' field", ConfigFilePath)
	}

	return nil
}

// GetConfig reads and returns the raw YAML content of config.yaml.
// Returns an empty string if the file doesn’t exist, or an error if reading fails.
func GetConfig() (string, error) {
	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("error reading %s: %v", ConfigFilePath, err)
	}
	return string(data), nil
}

// LoadEnv loads the config.yaml configuration and returns the current environment.
// Creates the file with a default config if it doesn’t exist, but only in a valid repo context.
func LoadEnv() (Env, error) {
	data, err := GetConfig()
	if err != nil {
		return Env{}, err
	}

	if data == "" {
		defaultConfig := configFile{
			Env: "dev",
			Envs: map[string]Env{
				"dev": {Vars: make(map[string]string)},
			},
		}
		if err := saveConfigFile(defaultConfig); err != nil {
			return Env{}, fmt.Errorf("error creating default %s: %v", ConfigFilePath, err)
		}
		return Env{
			Name: "dev",
			Vars: make(map[string]string),
		}, nil
	}

	var config configFile
	if err := yaml.Unmarshal([]byte(data), &config); err != nil {
		return Env{}, fmt.Errorf("error parsing %s: %v (content: %s)", ConfigFilePath, err, data)
	}

	if config.Env == "" {
		config.Env = "dev"
	}

	if config.Envs == nil {
		config.Envs = map[string]Env{
			"dev": {Vars: make(map[string]string)},
		}
	}

	vars, ok := config.Envs[config.Env]
	if !ok {
		vars = Env{Vars: make(map[string]string)}
	}

	return Env{
		Name: config.Env,
		Vars: vars.Vars,
	}, nil
}

// SetEnvName updates the current environment name in config.yaml and returns the updated Env.
func SetEnvName(envName string) (Env, error) {
	data, err := GetConfig()
	if err != nil {
		return Env{}, err
	}

	config := configFile{
		Env: envName,
	}
	if data != "" {
		if err := yaml.Unmarshal([]byte(data), &config); err != nil {
			return Env{}, fmt.Errorf("error parsing %s: %v", ConfigFilePath, err)
		}
	}

	if config.Envs == nil {
		config.Envs = map[string]Env{
			"dev": {Vars: make(map[string]string)},
		}
	}

	if config.Envs[envName].Vars == nil {
		config.Envs[envName] = Env{Vars: make(map[string]string)}
	}

	if err := saveConfigFile(config); err != nil {
		return Env{}, err
	}

	return Env{
		Name: envName,
		Vars: config.Envs[envName].Vars,
	}, nil
}

// SetEnvVar sets an environment variable for the current environment in config.yaml
// Returns the updated Env with the new variable set.
func SetEnvVar(key, value string) (Env, error) {
	data, err := GetConfig()
	if err != nil {
		return Env{}, err
	}

	config := configFile{
		Env: "dev", // Default if no file
	}
	if data != "" {
		if err := yaml.Unmarshal([]byte(data), &config); err != nil {
			return Env{}, fmt.Errorf("error parsing %s: %v", ConfigFilePath, err)
		}
	}

	if config.Envs == nil {
		config.Envs = map[string]Env{
			"dev": {Vars: make(map[string]string)},
		}
	}

	if config.Envs[config.Env].Vars == nil {
		config.Envs[config.Env] = Env{Vars: make(map[string]string)}
	}
	config.Envs[config.Env].Vars[key] = value

	if err := saveConfigFile(config); err != nil {
		return Env{}, err
	}

	return Env{
		Name: config.Env,
		Vars: config.Envs[config.Env].Vars,
	}, nil
}

// saveConfigFile writes the config file.
func saveConfigFile(config configFile) error {
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(ConfigFilePath, data, 0644); err != nil {
		return fmt.Errorf("error writing %s: %v", ConfigFilePath, err)
	}

	return nil
}
