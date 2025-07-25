package util

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is an internal struct for parsing config.yaml.
type Config struct {
	Env  string         `yaml:"env"`
	Envs map[string]Env `yaml:"envs"`
}

// CheckRepoContext verifies if the current directory contains a valid localpost project.
func CheckRepoContext() error {
	if _, err := os.Stat(LocalpostDir); os.IsNotExist(err) {
		return fmt.Errorf("lpost directory not found")
	}
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		return fmt.Errorf("config.yaml not found")
	}
	return nil
}

// ReadConfig reads and returns the parsed config.yaml with defaults applied if missing.
func ReadConfig() (*Config, error) {
	defaultEnv := map[string]Env{
		"dev": {
			Vars:    make(map[string]string),
			Timeout: 10, // Default timeout: 10 seconds
		},
	}
	defaultConfig := &Config{
		Env:  "dev",
		Envs: defaultEnv,
	}

	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := writeConfig(defaultConfig); err != nil {
				return nil, fmt.Errorf("error creating default %s: %v", ConfigFilePath, err)
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("error reading %s: %v", ConfigFilePath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", ConfigFilePath, err)
	}

	// Apply defaults
	if config.Env == "" {
		config.Env = "dev"
	}
	if config.Envs == nil {
		config.Envs = defaultEnv
	}
	if currentEnv, ok := config.Envs[config.Env]; !ok {
		// If env doesn't exist, use default with empty vars
		config.Envs[config.Env] = Env{
			Vars:    make(map[string]string),
			Timeout: defaultEnv["dev"].Timeout,
		}
	} else if currentEnv.Timeout == 0 {
		// If env exists but timeout is unset, apply default
		config.Envs[config.Env] = Env{
			Vars:    currentEnv.Vars,
			Login:   currentEnv.Login,
			Timeout: currentEnv.Timeout,
		}
	}

	return &config, nil
}

// LoadEnv loads the current environment from config.yaml.
func LoadEnv() (Env, error) {
	if err := CheckRepoContext(); err != nil {
		return Env{}, err
	}

	config, err := ReadConfig()
	if err != nil {
		return Env{}, err
	}

	currentEnv := config.Envs[config.Env]
	currentEnv.Name = config.Env
	return currentEnv, nil
}

// SetEnvVar updates an environment variable in config.yaml for the current environment.
func SetEnvVar(key, value string) error {
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}

	currentEnv := config.Envs[config.Env]
	if currentEnv.Vars == nil {
		currentEnv.Vars = make(map[string]string)
	}
	currentEnv.Vars[key] = value
	config.Envs[config.Env] = currentEnv

	return writeConfig(config)
}

// SetEnv sets the active environment in config.yaml.
func SetEnv(envName string) error {
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}

	config.Env = envName
	if _, ok := config.Envs[envName]; !ok {
		config.Envs[envName] = Env{
			Vars: make(map[string]string),
		}
	}

	return writeConfig(config)
}

// writeConfig writes the config struct to config.yaml.
func writeConfig(config *Config) error {
	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	if err := os.WriteFile(ConfigFilePath, out, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}
	return nil
}

// replacePlaceholders replaces placeholders like {VAR} in the input string or URL query params with values from vars.
func replacePlaceholders(input string, vars map[string]string) (string, error) {
	// Replace placeholders in the entire string first
	replaced := replaceString(input, vars)

	// Parse as URL to handle query params
	u, err := url.Parse(replaced)
	if err != nil {
		return replaced, nil // If not a valid URL, return as-is
	}

	if u.RawQuery != "" {
		// Handle URL query params
		q := u.Query()
		for key, values := range q {
			if len(values) > 0 {
				newVal := replaceString(values[0], vars)
				if newVal != values[0] {
					q.Set(key, newVal)
				}
			}
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// replaceString replaces {VAR} placeholders in a string with values from vars.
func replaceString(input string, vars map[string]string) string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		key := strings.Trim(match, "{}")
		if value, ok := vars[key]; ok {
			return value
		}
		return match
	})
}
