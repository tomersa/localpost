package util

// Response holds the results of an HTTP request execution.
type Response struct {
	ReqURL      string              // Final URL after env var substitution
	ReqHeaders  map[string]string   // RequestDefinition headers sent
	ReqBody     string              // RequestDefinition body sent
	StatusCode  int                 // HTTP status (e.g., "200 OK")
	RespHeaders map[string][]string // Response headers received
	RespBody    string              // Response body received
}

// Ephemeral holds runtime cookies and variables.
type Ephemeral struct {
	Cookies map[string]string `yaml:"cookies,omitempty"`
	Vars    map[string]string `yaml:"vars,omitempty"`
}

// Env represents an environment with its persistent variables and login config.
type Env struct {
	Name    string            `yaml:"-"`       // Current environment name
	Vars    map[string]string `yaml:",inline"` // Persistent environment variables
	Login   *LoginConfig      `yaml:"login,omitempty"`
	Timeout int               `yaml:"timeout,omitempty"`
}

// LoginConfig defines the login request and status codes for retry.
type LoginConfig struct {
	Request     string `yaml:"request"`
	TriggeredBy []int  `yaml:"triggered_by"`
}

// RequestDefinition defines an HTTP request from a YAML file.
type RequestDefinition struct {
	Method  string            // Not in YAML, from filename
	URL     string            `yaml:"url,omitempty"` // Optional
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    Body              `yaml:"body,omitempty"`
	SetEnv  map[string]struct {
		Header string `yaml:"header,omitempty"`
		Body   string `yaml:"body,omitempty"`
	} `yaml:"set-env-var,omitempty"`
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
