package util

import "time"

// Response holds the results of an HTTP request execution.
type Response struct {
	ReqURL      string              // Final URL after env var substitution
	ReqHeaders  map[string]string   // RequestDefinition headers sent
	ReqBody     string              // RequestDefinition body sent
	StatusCode  int                 // HTTP status (e.g., "200 OK")
	RespHeaders map[string][]string // Response headers received
	RespBody    string              // Response body received
	Duration    time.Duration       // Time taken for the request
}

// Env represents an environment with its variables and cookies.
type Env struct {
	Name    string            `yaml:"-"`                 // Current environment name (not in YAML for Envs values)
	Vars    map[string]string `yaml:",inline"`           // Environment variables (inline for Envs)
	Cookies map[string]string `yaml:"cookies,omitempty"` // Cookies stored as name=value pairs
}

// Body represents the request body content.
type Body struct {
	Json           map[string]interface{} `yaml:"json,omitempty"`
	FormUrlEncoded map[string]string      `yaml:"form-urlencoded,omitempty"`
	Form           struct {
		Fields map[string]string `yaml:"fields,omitempty"`
		Files  map[string]string `yaml:"files,omitempty"`
	} `yaml:"form-data,omitempty"`
	Text string `yaml:"text,omitempty"`
}

// LoginConfig defines the login request and status codes for retry.
type LoginConfig struct {
	Request     string `yaml:"request"`      // Name of the login request (e.g., "POST_login")
	TriggeredBy []int  `yaml:"triggered_by"` // StatusCode codes triggering login (e.g., [400, 401, 403])
}

// RequestDefinition defines an HTTP request from a YAML file.
type RequestDefinition struct {
	Method  string            // Not in YAML, from filename
	URL     string            `yaml:"url"` // Required
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    Body              `yaml:"body,omitempty"`
	SetEnv  map[string]struct {
		Header string `yaml:"header,omitempty"`
		Body   string `yaml:"body,omitempty"`
	} `yaml:"set-env-var,omitempty"`
	PreFlight  string       `yaml:"pre-flight,omitempty"`  // RequestDefinition to run before
	PostFlight string       `yaml:"post-flight,omitempty"` // RequestDefinition to run after
	Login      *LoginConfig `yaml:"login,omitempty"`       // Login request to run on specified status codes
}
