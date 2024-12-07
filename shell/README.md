# Shell

## Features

- **POSIX Compliance**
- **Builtin Commands**:
  - `cd`: Change the working directory, with support for `~` (HOME directory).
  - `pwd`: Print the current working directory.
  - `echo`: Print arguments to standard output, with support for escape sequences.
  - `type`: Display the type of a command (builtin or external).
  - `exit`: Exit the shell with a specified status code.
- **PATH Support**: Executes external programs located in directories specified by the `PATH` environment variable.
- **HOME Directory Shortcut**: Recognizes `~` as the `HOME` environment variable in commands like `cd`.
- **Command Execution**: Handles both built-in commands and external program execution.
- **Error Handling**: Provides user-friendly error messages for invalid commands or syntax errors.

## Credits

- Codecrafters
