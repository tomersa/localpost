package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const ConfigFilePath = "./.localpost-config"
const RequestsDir = "requests"

// configFile is an internal struct for parsing .localpost-config YAML.
type configFile struct {
	Env  string         `yaml:"env"`
	Envs map[string]Env `yaml:"envs"`
}

// GetConfig reads and returns the raw YAML content of .localpost-config.
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

// LoadEnv loads the .localpost-config configuration and returns the current environment.
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

// SetEnvName updates the current environment name in .localpost-config and returns the updated Env.
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

// SetEnvVar sets an environment variable for the current environment in .localpost-config
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

// saveConfigFile writes the config to .localpost-config.
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
