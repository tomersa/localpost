# Localpost

Localpost is a command-line tool for creating and executing HTTP requests, with support for environment variables and dynamic response handling.

## How it works?
Localpost uses your Git repo to share HTTP requests. Each request is a YAML file in the `requests/` folder, named `METHOD_request_nickname.yaml`, ready to commit and collaborate.

> ℹ️ **For creation of request definition you can use `lpost add-request`**:

> ℹ️ **Shorthand**: Use `lpost` alias that is already set for you.

#### Request definition example:
```yaml
# ./requests/POST_login.yaml
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
Now you can execute this request with `$: lpost -r POST_login`.

> ℹ️ **Collaboration**: Commit your request definitions files to your repo to collaborate with others, or manage them locally without sharing. Key files to include: the `requests/` directory and `.localpost-config`.


##  Installation
- Grab the latest release from [GitHub Releases](https://github.com/moshe5745/localpost/releases):
  - ### macOS Intel (amd64)
    ```bash
    curl -L https://github.com/moshe5745/localpost/releases/download/v1.0.0/localpost-v1.0.0-darwin-amd64 -o localpost
    ```
  - ### macOS Apple Silicon (arm64)
    ```bash
    curl -L https://github.com/moshe5745/localpost/releases/download/v1.0.0/localpost-v1.0.0-darwin-arm64 -o localpost  
    ```
  - ### Make it executable
    ```bash
    chmod +x localpost
    ```
  - ### Move to /usr/local/bin (So it will be globally available in your machine)
    ```bash
    sudo mv localpost /usr/local/bin/localpost
    ```

### Shell Completion installation
- Enable autocompletion by adding the following to your shell config file.
  - Zsh
    ```zsh
    source <(lpost completion --shell zsh)
    # Add to ~/.zshrc
    ```
  - Bash
    ```bash
    source <(lpost completion --shell bash)
    # Add to ~/.bashrc
    ```
  - Fish
    ```bash
    source (lpost completion --shell fish | psub)
    # Add to ~/.config/fish/config.fish
    ```
- Use TAB key for completion
> ⚠️  After adding the completion line to your shell config (e.g., ~/.zshrc), run `source ~/.zshrc` (or equivalent) to apply it immediately, or restart your shell.

## Basic usage example
```bash
lpost add-request
# POST_login added
```
```bash
lpost set-env prod
# Default env is dev
```
```bash
lpost set-env-var BASE_URL https://example.com
```
```bash
lpost -r POST_login
```
```bash
$: 
+-----------+----------+
| STATUS    | TIME     |
+-----------+----------+
| 200 OK    | 20ms     |
+-----------+----------+
| BODY                 |
+----------------------+
| {"TOKEN":"123456"}   |
+----------------------+
```
```bash
lpost -r POST_login -v
# Verbose for debugging 
```
```bash
$:
-----
Status: 200 OK
Time: 18ms
URL: http://localhost:8080/login
-----
Request
  Headers:
    Content-Type: application/json
    ...
Request
  Body:
    
-----
Response
  Headers:
    Content-Type: [application/json]
    ...
Request
  Body:
    {"TOKEN":"123456"}
```

## Environment Variables
- Use `lpost` to store and manage variables for different environments (e.g., `dev`, `prod`).
- Let you set custom values (e.g., API endpoints, credentials) per environment.

> ℹ️ **Note**: These environment variables are unique to `localpost` and stored in `.localpost-config`—they’re separate from your shell’s environment, `.env` files, or other tools.

### Setting Environment Variables
You can set variables in two ways:
- **CLI**:
  ```bash
  lpost set-env-var BASE_URL https://api.example.com
  ```
- **YAML**:
  ```yaml
  ...
  set-env-var:
    TOKEN: # Var name
      body: jwt-token # from "jwt-token" param in JSON body
    Cookie: 
      header: Cookie # from "Cookie" header
  ```
Then you can use them inside your requests definitions:
```yaml
url: "{BASE_URL}/login"
headers:
  Cookie: "{Cookie}"
```

## Configuration file
For storing the collection env vars and current env used.

- **File**: `.localpost-config` in the current directory.
- **Format**:
  ```yaml
  env: dev # Current env used
  envs:    # All envs list
    dev:
      BASE_URL: https://api.example.com
      TOKEN: 123
    prod:
      BASE_URL: https://api.prod.com
      TOKEN: 456
  ```
> ℹ️ Note: `.localpost-config` is created automatically on first use with a default `env: dev` if it doesn’t exist.

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

## Commands full list

| Command                  | Description                                                                                                        | Example Usage                         |
|--------------------------|--------------------------------------------------------------------------------------------------------------------|---------------------------------------|
| `add-request`            | Create a new request YAML file interactively with prompts for nickname, URL, method, body type, and Accept header. | `$: lpost add-request`                |
| `request <METHOD_name>`  | Execute a request from a YAML file in the `requests/` directory. Shorthand: `-r`.                                  | `$: lpost -r POST_login`             |
| `set-env <env>`          | Set the current environment in `.localpost-config`.                                                                | `$: lpost set-env prod`              |
| `set-env-var <key> <value>` | Set an environment variable for the current environment in `.localpost-config`.                                    | `$: lpost set-env-var BASE_URL https://api.example.com` |
| `show-env`               | Display the current environment and variables from `.localpost-config`. Use `--all` for the full config.           | `$: lpost show-env` or `$: lpost show-env --all` |
| `completion`             | Output completion script for your shell (bash, zsh, fish) to stdout. Requires `--shell` flag.                      | `$: source <(lpost completion --shell zsh)` |

- **Global Flag**: Override the environment temporarily with `-e` or `--env`:
  ```bash
  lpost -e prod -r GET_users
  ```

- **Body Types**:
  - `json`: JSON object (e.g., `{"key": "value"}`).
  - `form-urlencoded`: Key-value pairs (e.g., `key=value`).
  - `form-data`: Fields and files (e.g., `fields: {field: value}`, `files: {file: path}`).
  - `text`: Plain text (e.g., `example text`).

## Building from Source
1. Clone the repo:
   ```bash
   git clone https://github.com/moshe5745/localpost.git
   cd localpost
   ```
2. Build:
   ```bash
   go build -o localpost
   ```
3. Run:
   ```bash
   lpost add-request
   ```

## Contributing
Feel free to submit issues or PRs at [github.com/moshe5745/localpost](https://github.com/moshe5745/localpost).

## License
[MIT License](LICENSE)
```