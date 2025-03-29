package util

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const LocalpostDir = "lpost"
const RequestsDir = LocalpostDir + "/requests"
const SchemasDir = LocalpostDir + "/schemas"
const ConfigFile = "config.yaml"
const ConfigFilePath = LocalpostDir + "/" + ConfigFile

// configFile is an internal struct for parsing config.yaml.
type configFile struct {
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
// Creates the file with a default config if it doesn’t exist.
func ReadConfig() (*configFile, error) {
	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if file doesn't exist
			defaultConfig := &configFile{
				Env: "dev",
				Envs: map[string]Env{
					"dev": {Vars: make(map[string]string), Cookies: make(map[string]string)},
				},
			}
			if err := writeConfig(defaultConfig); err != nil {
				return nil, fmt.Errorf("error creating default %s: %v", ConfigFilePath, err)
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("error reading %s: %v", ConfigFilePath, err)
	}

	var config configFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", ConfigFilePath, err)
	}

	// Apply defaults if fields are missing
	if config.Env == "" {
		config.Env = "dev"
	}
	if config.Envs == nil {
		config.Envs = map[string]Env{
			"dev": {Vars: make(map[string]string), Cookies: make(map[string]string)},
		}
	}
	if _, ok := config.Envs[config.Env]; !ok {
		config.Envs[config.Env] = Env{
			Vars:    make(map[string]string),
			Cookies: make(map[string]string),
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
			Vars:    make(map[string]string),
			Cookies: make(map[string]string),
		}
	}

	return writeConfig(config)
}

// SetCookie updates a cookie in config.yaml for the current environment.
func SetCookie(name, value string) error {
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}

	currentEnv := config.Envs[config.Env]
	if currentEnv.Cookies == nil {
		currentEnv.Cookies = make(map[string]string)
	}
	currentEnv.Cookies[name] = value
	config.Envs[config.Env] = currentEnv

	return writeConfig(config)
}

// writeConfig writes the config struct to config.yaml.
func writeConfig(config *configFile) error {
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
	// First, try parsing as a URL to handle query params
	u, err := url.Parse(input)
	if err == nil && u.RawQuery != "" {
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
		return u.String(), nil
	}

	// If not a URL or no query params, treat as a plain string
	return replaceString(input, vars), nil
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
