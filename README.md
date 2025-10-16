# Cmd-Vault

A retro-style, keyboard-driven TUI for saving, browsing, and executing your favorite shell commands. Never forget that complex `ffmpeg` or `git` command again!

Note: Add a screenshot 

## Features

*   **Interactive TUI**: A fast, keyboard-driven Terminal User Interface for managing your command library.
*   **CRUD Operations**: Easily **A**dd, **E**dit, and **D**elete commands.
*   **Command Execution**: Run saved commands directly from the TUI and view their output in a dedicated panel.
*   **Built-in File Browser**: Navigate your filesystem to run commands in specific directories.
*   **Mini-Terminal**: Run one-off, temporary commands in any directory using the file browser.
*   **Paste Functionality**: Paste saved commands into the mini-terminal for quick modifications before running.
*   **Non-Interactive Mode**: Execute saved commands directly from your shell for scripting or quick access (`cmd-vault run <command_name>`).
*   **Usage Tracking**: Automatically counts how many times each command is run.
*   **Responsive Layout**: The TUI layout adapts to your terminal's width, switching between horizontal and vertical views.

## Installation

Make sure you have Go installed (version 1.21+ is recommended).

```sh
# Clone the repository
git clone https://github.com/kanekitakitos/cmd-vault.git
cd cmd-vault

# Build the binary
go build .

# Or install it directly to your $GOPATH/bin
go install .
```

## Usage

### Interactive TUI

To start the main application, simply run:

```sh
cmd-vault
```

By default, it will create and use a database file named `lazycmd.db` in the current directory.

#### Keybindings

The TUI is designed to be used entirely with the keyboard.

| Key(s)      | Action                                       |
|-------------|----------------------------------------------|
| `↑`/`k`, `↓`/`j`| Navigate lists (commands, files, etc.)       |
| `r`         | **R**un selected command (or open mini-terminal) |
| `s`         | Open/close file brow**s**er                  |
| `o`         | Focus/scroll **o**utput panel                |
| `a`         | **A**dd a new command                        |
| `e`         | **E**dit the selected command                |
| `d`         | **D**elete the selected command              |
| `c`         | **C**opy current path (in file browser)      |
| `p`         | **P**aste saved command (in mini-terminal)   |
| `x`         | Show/hide contextual help                    |
| `?`         | Show the main help screen                    |
| `q` / `esc` | Quit the program or cancel an action         |
| `ctrl+c`    | Force quit the application                   |

### Non-Interactive Mode

You can run a saved command directly without entering the TUI. This is useful for scripts or integrating with other tools.

```sh
# Run a command named 'list-files'
cmd-vault run list-files
```

### Configuration

#### Database Path

You can specify a custom path for the SQLite database file using the `--db` flag. This flag works for both the TUI and the `run` subcommand.

```sh
# Start the TUI with a database in your home directory
cmd-vault --db ~/.config/cmd-vault/commands.db

# Run a command using the same custom database
cmd-vault run my-command --db ~/.config/cmd-vault/commands.db
```

## How It Works

Cmd-Vault stores all your commands in a local SQLite database. The TUI is built using the wonderful Bubble Tea framework, which makes it easy to build stateful, responsive terminal applications.

*   **Backend**: Standard Go `database/sql` package with the `mattn/go-sqlite3` driver.
*   **CLI Framework**: Cobra for robust command-line argument parsing.
*   **TUI**: Bubble Tea for the application model and Lip Gloss for styling.

## Development

To build the project from source:

```sh
# Get dependencies
go mod tidy

# Build the binary for your system
go build -o cmd-vault .
```

## To-Do / Future Ideas

- [ ] Add command tagging/categorization.
- [ ] Implement a more powerful search/filter feature for the command list.
- [ ] Add support for environment variable placeholders in commands (e.g., `echo $HOME`).
- [ ] Cross-platform testing and support (currently developed on Windows).
- [ ] Add import/export functionality for the command database (e.g., JSON, CSV).

