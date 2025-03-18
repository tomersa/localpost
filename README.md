# Localpost

Localpost is a command-line tool for creating and executing HTTP requests, with support for environment variables and dynamic response handling.

## Download Binaries
Grab the latest release from [GitHub Releases](https://github.com/yourusername/localpost/releases):

### Shell Completion

- Enable autocompletion by adding the following to your shell config file:
  ```zsh
  # Zhs: Add to ~/.zshrc
  source <(localpost completion --shell zsh)

  # Bash: Add to ~/.bashrc
  source <(localpost completion --shell bash)
  
  # Fish: Add to ~/.config/fish/config.fish
  source (localpost completion --shell fish | psub)

## Usage

### Commands

| Command                  | Description                                              | Example Usage                         |
|--------------------------|----------------------------------------------------------|---------------------------------------|
| `add-request`            | Create a new request YAML file interactively with prompts for nickname, URL, method, body type, and Accept header. | `$: localpost add-request`            |
| `request <METHOD_name>`  | Execute a request from a YAML file in the `requests/` directory. | `$: localpost request POST_login`     |
| `set-env <env>`          | Set the current environment in `.localpost`.             | `$: localpost set-env prod`           |
| `set-env-var <key> <value>` | Set an environment variable for the current environment in `.localpost`. | `$: localpost set-env-var BASE_URL https://api.example.com` |
| `show-env`               | Display the current environment and variables from `.localpost`. Use `--all` for the full config. | `$: localpost show-env` or `$: localpost show-env --all` |
| `completion`             | Output completion script for your shell (bash, zsh, fish) to stdout. Use `--shell` to specify shell if needed. | `$: source <(localpost completion --shell zsh)` |
- **Global Flag**: Override the environment temporarily with `-e` or `--env`:

- **Body Types**:
    - `json`: JSON object (e.g., `{"key": "value"}`).
    - `form-urlencoded`: Key-value pairs (e.g., `key=value`).
    - `form-data`: Fields and files (e.g., `fields: {field: value}`, `files: {file: path}`).
    - `text`: Plain text (e.g., `example text`).

See more examples in the [Request Templates](#request-templates) section.

## Environment

####  You can store variables for specific env.
You have 2 options for setting env variables. From CLI or with request YAML.
- **CLI**: `$: localpost set-env-var BASE_URL https://api.example.com`.
- **YAML**:
```yaml
set-env-var:
  TOKEN: # Var name
    body: jwt-token # from "jwt-token" param in json body
  Cookie: 
    header: Cookie # from "Cookie" header
```
Then you can use them inside your requests.
```yaml
url: "{BASE_URL}/login"
header: "Cookie {Cookie}"
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
