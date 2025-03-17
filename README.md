# Localpost

Localpost is a command-line tool for creating and executing HTTP requests, with support for environment variables and dynamic response handling.

## Download Binaries
Grab the latest release from [GitHub Releases](https://github.com/yourusername/localpost/releases):

## Usage

### Commands
- **`add-request`**: Create a new request YAML file interactively.
  ```bash
  ./localpost add-request
  ```
    - Prompts for nickname, URL, method, body type (if applicable), and Accept header.

- **`request <METHOD_name>`**: Execute a request from a YAML file.
  ```bash
  ./localpost request POST_login
  ```

- **`set-env <env>`**: Set the current environment in `.localpost`.
  ```bash
  ./localpost set-env prod
  ```

- **`set-env-var <key> <value>`**: Set an environment variable for the current environment.
  ```bash
  ./localpost set-env-var BASE_URL https://api.example.com
  ```

- **`show-env`**: Display the current environment and variables.
  ```bash
  ./localpost show-env
  # Optional: --all for full .localpost-config
  ```

- **Global Flag**: Override environment temporarily:
  ```bash
  ./localpost -e prod request GET_users
  ```

### Request YAML Format
Request files are stored in the `requests/` directory. Example (`requests/POST_login.yaml`):
```yaml
url: "{BASE_URL}/login"
headers:
  Accept: application/json
  Content-Type: application/json
body:
  json:
    username: user
    password: pass
set-env-var:
  TOKEN:
    body: jwt-token
```

- **Body Types**:
    - `json`: JSON object (e.g., `{"key": "value"}`).
    - `form-urlencoded`: Key-value pairs (e.g., `key=value`).
    - `form-data`: Fields and files (e.g., `fields: {field: value}`, `files: {file: path}`).
    - `text`: Plain text (e.g., `example text`).

See more examples in the [Request Templates](#request-templates) section.

## Configuration
- **File**: `.localpost` in the current directory.
- **Format**:
  ```yaml
  env: dev
  envs:
    dev:
      BASE_URL: https://api.example.com
    prod:
      BASE_URL: https://api.prod.com
  ```

## Request Templates
Check out detailed examples at [github.com/yourusername/localpost#request-templates](https://github.com/yourusername/localpost#request-templates) for common use cases like JSON POSTs, file uploads, and dynamic env vars.

## Building from Source
1. Clone the repo:
   ```bash
   git clone https://github.com/yourusername/localpost.git
   cd localpost
   ```
2. Build:
   ```bash
   go build -o localpost
   ```
3. Run:
   ```bash
   ./localpost add-request
   ```

## Contributing
Feel free to submit issues or PRs at [github.com/yourusername/localpost](https://github.com/yourusername/localpost).

## License
[MIT License](LICENSE)
