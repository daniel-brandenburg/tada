# Tada - A Terminal-Based Todo Application

Tada is a simple yet powerful todo application with both CLI and TUI interfaces. It helps you organize and manage your tasks efficiently from the terminal.

![Tada Logo](https://via.placeholder.com/150x150.png?text=Tada)

## Features

- **Flexible Task Management**: Create, edit, complete, and manage tasks with ease
- **Topic Organization**: Group tasks by topics/projects
- **Priority Levels**: Assign priorities to your tasks
- **Status Tracking**: Track task status (todo, in-progress, done, paused, cancelled)
- **Tag Support**: Add tags to categorize and filter tasks
- **Markdown Storage**: Tasks are stored as Markdown files with YAML frontmatter
- **Two Interfaces**:
  - Command-line interface (CLI) for quick operations
  - Terminal user interface (TUI) for interactive management
- **Configurable**: Use `tada config` to view and set preferences
- **Accessible**: Designed with accessibility and usability in mind

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/daniel-brandenburg/tada.git
cd tada

# Build the application
go build -o tada

# Move to a directory in your PATH (optional)
sudo mv tada /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/daniel-brandenburg/tada@latest
```

## Usage

### CLI Commands

#### Add a Task

```bash
# Basic task
tada add "Buy groceries"

# Task with topic
tada add "work/Finish report"

# Task with description and priority
tada add "Fix bug in login" -d "The login button doesn't work on mobile" -p 1

# Task with tags
tada add "Clean house" -t "home,weekend"
```

#### List Tasks

```bash
# List all tasks
tada list

# Filter by status
tada list -s todo
tada list -s in-progress

# Sort tasks
tada list --sort priority
tada list --sort created
```

#### Complete a Task

```bash
# Complete a task
tada complete "Buy groceries"

# Complete a task in a topic
tada complete "work/Finish report"
```

#### Launch TUI

```bash
tada tui
```

### TUI Controls

- **Navigation**:
  - `j/k` or arrow keys: Move up/down
  - `Space/Enter`: Expand topic or edit task
  - `Tab/Shift+Tab`: Navigate between fields in edit mode
  
- **Actions**:
  - `a`: Add a new task
  - `r`: Refresh task list
  - `q`: Quit

## Configuration

Tada supports configuration via the `tada config` command. Config files are stored in your home directory and/or project `.tada` folder. You can view and set preferences such as default sort order, theme, and more:

```bash
tada config show
tada config set theme dark
tada config set defaultSort created
```

## Error Handling

Tada provides user-friendly error messages for common issues, including:
- Missing `.tada` directory (with onboarding guidance)
- File permission errors
- Corrupted or missing task files
- Invalid commands or arguments

## Output Formats

You can list tasks in different formats:

```bash
tada list --format table   # Default
tada list --format json
tada list --format yaml
```

## Accessibility & Manual Testing

Tada is tested for accessibility and usability. Manual testing tasks are tracked in `.tada/tasks/ManualTesting/` and cover:
- TUI color, alignment, and navigation
- Error handling and data integrity
- Accessibility (contrast, screen readers)
- Platform compatibility
- Bulk operations, config, onboarding, and more

## Testing

To run all tests and check coverage:

```bash
go test -v ./...
go test -cover ./...
```

High test coverage is maintained for all critical paths, error handling, and output formats. See `.tada/tasks/ManualTesting/` for manual test scenarios.

## Task Storage

Tasks are stored as Markdown files with YAML frontmatter in the `.tada` directory:

```
.tada/
├── archive/      # Completed tasks
└── tasks/        # Active tasks
    ├── work/     # Tasks in the "work" topic
    └── home/     # Tasks in the "home" topic
```

Each task is saved as a Markdown file with a timestamp and slug of the task title:

```markdown
---
title: Buy groceries
priority: 2
status: todo
tags:
  - shopping
  - errands
created_at: 2025-06-18T14:30:00Z
---

# Buy groceries

Remember to get milk, eggs, and bread.
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - TUI styling
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML parsing

