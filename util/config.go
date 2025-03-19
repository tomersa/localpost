package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigFilePath is the path to the .localpost-config file.
const ConfigFilePath = "./.localpost-config"
const DefaultRequestsDir = "requests"

type Env struct {
	Name string            `yaml:"-"`       // Current environment name (not in YAML for Envs values)
	Vars map[string]string `yaml:",inline"` // Environment variables (inline for Envs)
}

type configFile struct {
	Env  string         `yaml:"env"`
	Envs map[string]Env `yaml:"envs"`
}

type Body struct {
	Json           map[string]interface{} `yaml:"json,omitempty"`
	FormUrlEncoded map[string]string      `yaml:"form-urlencoded,omitempty"`
	Form           struct {
		Fields map[string]string `yaml:"fields,omitempty"`
		Files  map[string]string `yaml:"files,omitempty"`
	} `yaml:"form-data,omitempty"`
	Text string `yaml:"text,omitempty"`
}

type Request struct {
	Method  string            // Not in YAML, from filename
	URL     string            `yaml:"url"` // Required
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    Body              `yaml:"body,omitempty"` // Renamed from Data
	SetEnv  map[string]struct {
		Header string `yaml:"header,omitempty"`
		Body   string `yaml:"body,omitempty"`
	} `yaml:"set-env-var,omitempty"`
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
// Creates the file with a default config if it doesn’t exist.
func LoadEnv() (Env, error) {
	data, err := GetConfig()
	if err != nil {
		return Env{}, err
	}

	if data == "" {
		// File doesn’t exist—create default config
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

// saveConfigFile marshals and writes a configFile struct to .localpost-config
func saveConfigFile(config configFile) error {
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	if err := os.WriteFile(ConfigFilePath, data, 0644); err != nil {
		return fmt.Errorf("error writing to %s: %v", ConfigFilePath, err)
	}
	return nil
}

func ExecuteRequest(req Request) (reqHeaders map[string]string, reqBody string, status string, respHeaders map[string][]string, respBody string, duration time.Duration, err error) {
	env, err := LoadEnv()
	if err != nil {
		return nil, "", "", nil, "", 0, fmt.Errorf("error loading env: %v", err)
	}

	finalURL := ReplaceEnvVars(req.URL, env.Vars)
	finalURL, err = ReplaceParams(finalURL, env.Vars)
	if err != nil {
		return nil, "", "", nil, "", 0, err
	}

	if !strings.HasPrefix(finalURL, "http://") && !strings.HasPrefix(finalURL, "https://") {
		return nil, "", "", nil, "", 0, fmt.Errorf("invalid URL: %s (must resolve to http:// or https://)", finalURL)
	}

	for key, value := range req.Headers {
		req.Headers[key] = ReplaceEnvVars(value, env.Vars)
	}

	var body io.Reader
	contentType := req.Headers["Content-Type"]
	switch contentType {
	case "application/json", "":
		if len(req.Body.Json) > 0 {
			bodyBytes, err := json.Marshal(req.Body.Json)
			if err != nil {
				return nil, "", "", nil, "", 0, fmt.Errorf("error marshaling JSON body: %v", err)
			}
			bodyStr := ReplaceEnvVars(string(bodyBytes), env.Vars)
			body = strings.NewReader(bodyStr)
			if contentType == "" {
				contentType = "application/json"
			}
		}
	case "application/x-www-form-urlencoded":
		if len(req.Body.FormUrlEncoded) > 0 {
			data := make(url.Values)
			for k, v := range req.Body.FormUrlEncoded {
				data.Set(k, v)
			}
			bodyStr := ReplaceEnvVars(data.Encode(), env.Vars)
			body = strings.NewReader(bodyStr)
		}
	case "multipart/form-data":
		if len(req.Body.Form.Fields) > 0 || len(req.Body.Form.Files) > 0 {
			bodyBuffer := &bytes.Buffer{}
			writer := multipart.NewWriter(bodyBuffer)

			for k, v := range req.Body.Form.Fields {
				writer.WriteField(k, v)
			}
			for k, filePath := range req.Body.Form.Files {
				file, err := os.Open(filePath)
				if err != nil {
					return nil, "", "", nil, "", 0, fmt.Errorf("error opening file %s: %v", filePath, err)
				}
				defer file.Close()
				part, err := writer.CreateFormFile(k, filepath.Base(filePath))
				if err != nil {
					return nil, "", "", nil, "", 0, fmt.Errorf("error creating form file %s: %v", k, err)
				}
				_, err = io.Copy(part, file)
				if err != nil {
					return nil, "", "", nil, "", 0, fmt.Errorf("error writing file %s to form: %v", k, err)
				}
			}
			err := writer.Close()
			if err != nil {
				return nil, "", "", nil, "", 0, fmt.Errorf("error closing form writer: %v", err)
			}
			body = bodyBuffer
			contentType = writer.FormDataContentType()
		}
	case "text/plain":
		if req.Body.Text != "" {
			bodyStr := ReplaceEnvVars(req.Body.Text, env.Vars)
			body = strings.NewReader(bodyStr)
		}
	default:
		if len(req.Body.Json) > 0 || len(req.Body.FormUrlEncoded) > 0 || len(req.Body.Form.Fields) > 0 || len(req.Body.Form.Files) > 0 || req.Body.Text != "" {
			return nil, "", "", nil, "", 0, fmt.Errorf("unsupported or missing Content-Type for body: %s", contentType)
		}
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(req.Method, finalURL, body)
	if err != nil {
		return nil, "", "", nil, "", 0, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", "", nil, "", 0, fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()
	duration = time.Since(start)

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", nil, "", 0, fmt.Errorf("error reading response: %v", err)
	}
	respBody = string(respBodyBytes)

	// Update env vars from response
	for envKey, rule := range req.SetEnv {
		var value string
		if rule.Header != "" {
			value = resp.Header.Get(rule.Header)
		} else if rule.Body != "" {
			var respData map[string]interface{}
			if len(respBody) > 0 {
				json.Unmarshal([]byte(respBody), &respData)
				if v, ok := respData[rule.Body].(string); ok {
					value = v
				}
			}
		}
		if value != "" {
			updatedEnv, err := SetEnvVar(envKey, value)
			if err != nil {
				return nil, "", "", nil, "", 0, fmt.Errorf("error setting env var %s: %v", envKey, err)
			}
			env = updatedEnv
		}
	}

	return req.Headers, reqBody, resp.Status, resp.Header, respBody, duration, nil
}

// ParseRequest parses a request YAML file into a Request struct.
func ParseRequest(filePath string) (Request, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Request{}, fmt.Errorf("error reading request file: %v", err)
	}

	var req Request
	if err := yaml.Unmarshal(data, &req); err != nil {
		return Request{}, fmt.Errorf("error parsing request file: %v", err)
	}

	if req.URL == "" {
		return Request{}, fmt.Errorf("URL is required in request file")
	}

	return req, nil
}

// ReplaceEnvVars replaces {{VAR}} placeholders using the provided env vars.
func ReplaceEnvVars(s string, envVars map[string]string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.Trim(match, "{}")
		if value, ok := envVars[key]; ok {
			return value
		}
		return match
	})
}

// ReplaceParams replaces {param} placeholders using the provided env vars.
func ReplaceParams(s string, envVars map[string]string) (string, error) {
	matches := paramRegex.FindAllStringSubmatch(s, -1)
	result := s
	for _, match := range matches {
		param := match[1]
		placeholder := fmt.Sprintf("{%s}", param)
		if value, ok := envVars[strings.ToUpper(param)]; ok {
			result = strings.ReplaceAll(result, placeholder, value)
		} else {
			return result, fmt.Errorf("no value provided for parameter '%s' in URL: %s", param, s)
		}
	}
	return result, nil
}

var envVarRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)
var paramRegex = regexp.MustCompile(`\{([^}]+)\}`)
