# Localpost

Localpost is a command-line tool for creating and executing HTTP requests, with support for environment variables and dynamic response handling.

## How it works?
Localpost uses your Git repo to share HTTP requests. Each request is a YAML file in the `requests/` folder, named `METHOD_request_nickname.yaml`, ready to commit and collaborate.

Example:
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
Now you can execute this request with `$: localpost -r POST_login`.

## Installation
- Grab the latest release from [GitHub Releases](https://github.com/yourusername/localpost/releases):
  ```bash
  # macOS Intel (amd64)
  curl -L https://github.com/yourusername/localpost/releases/download/v1.0.0/localpost-v1.0.0-darwin-amd64.zip -o localpost.zip && unzip localpost.zip && chmod +x localpost-darwin-amd64 && sudo mv localpost-darwin-amd64 /usr/local/bin/localpost

  # macOS Apple Silicon (arm64)
  curl -L https://github.com/yourusername/localpost/releases/download/v1.0.0/localpost-v1.0.0-darwin-arm64.zip -o localpost.zip && unzip localpost.zip && chmod +x localpost-darwin-arm64 && sudo mv localpost-darwin-arm64 /usr/local/bin/localpost
  ```

## Shell Completion
- Enable autocompletion by adding the following to your shell config file:
  ```zsh
  # Zsh: Add to ~/.zshrc
  source <(localpost completion --shell zsh)

  # Bash: Add to ~/.bashrc
  source <(localpost completion --shell bash)
  
  # Fish: Add to ~/.config/fish/config.fish
  source (localpost completion --shell fish | psub)
  ```
- Use it with TAB key


## Usage
```bash
$: localpost add-request
# POST_login added
$: localpost set-env prod
$: localpost set-env-var BASE_URL https://example.com
$: localpost request POST_login # or localpost -r POST_login
# { TOKEN: 123456 }
```

## Environment
#### You can store variables for specific envs.
You have 2 options for setting env variables: from CLI or with request YAML from .
- **CLI**: `$: localpost set-env-var BASE_URL https://api.example.com`
- **YAML**:
  ```yaml
  ...
  set-env-var:
    TOKEN: # Var name
      body: jwt-token # from "jwt-token" param in JSON body
    Cookie: 
      header: Cookie # from "Cookie" header
  ```
Then you can use them inside your requests:
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
| `add-request`            | Create a new request YAML file interactively with prompts for nickname, URL, method, body type, and Accept header. | `$: localpost add-request`            |
| `request <METHOD_name>`  | Execute a request from a YAML file in the `requests/` directory. Shorthand: `-r`.                                  | `$: localpost -r POST_login`          |
| `set-env <env>`          | Set the current environment in `.localpost-config`.                                                                | `$: localpost set-env prod`           |
| `set-env-var <key> <value>` | Set an environment variable for the current environment in `.localpost-config`.                                    | `$: localpost set-env-var BASE_URL https://api.example.com` |
| `show-env`               | Display the current environment and variables from `.localpost-config`. Use `--all` for the full config.           | `$: localpost show-env` or `$: localpost show-env --all` |
| `completion`             | Output completion script for your shell (bash, zsh, fish) to stdout. Use `--shell` to specify shell if needed.     | `$: source <(localpost completion --shell zsh)` |

- **Global Flag**: Override the environment temporarily with `-e` or `--env`:
  ```bash
  $: localpost -e prod -r GET_users
  ```

- **Body Types**:
    - `json`: JSON object (e.g., `{"key": "value"}`).
    - `form-urlencoded`: Key-value pairs (e.g., `key=value`).
    - `form-data`: Fields and files (e.g., `fields: {field: value}`, `files: {file: path}`).
    - `text`: Plain text (e.g., `example text`).

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
   $: localpost add-request
   ```

## Contributing
Feel free to submit issues or PRs at [github.com/yourusername/localpost](https://github.com/yourusername/localpost).

## License
[MIT License](LICENSE)
```