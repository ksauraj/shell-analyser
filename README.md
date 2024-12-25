# K8AU Shell Analyzer (alpha)

A powerful terminal-based tool that analyzes your shell usage patterns, technical profile, and configuration details across multiple shells (bash, zsh, and fish).

![Shell Analyzer Screenshot](screenshot.png)

## Features

- 📊 **Shell Usage Analysis**: Tracks command history and usage patterns across multiple shells
- 💻 **Technical Profile**: Identifies your primary role and tech stack based on command usage
- ⏰ **Work Pattern Analysis**: Discovers peak productivity hours and common workflows
- 🔧 **Tool Usage Statistics**: Monitors your usage of editors, programming languages, and build tools
- ⚙️ **Configuration Analysis**: Reviews shell configs, aliases, and plugins

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/k8au-shell-analyzer.git

# Navigate to the project directory
cd k8au-shell-analyzer

# Build the project
go build -o shell-analyzer
```

## Usage

Simply run the compiled binary:

```bash
./shell-analyzer
```

### Navigation
- Use `tab` to switch between different views
- Press `q` to quit the application
- Use mouse or keyboard to navigate content

## Views

1. **Overview**: General shell usage statistics and configuration details
2. **Tech Profile**: Analysis of your technical skills and proficiency
3. **Work Patterns**: Insights into your working hours and productivity
4. **Tool Usage**: Detailed breakdown of your development tools usage

## Requirements

- Go 1.16 or higher
- One or more of the following shells installed:
  - Bash
  - Zsh
  - Fish

## Dependencies

- github.com/charmbracelet/bubbles
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/gookit/color

## Screenshots
![image](https://github.com/user-attachments/assets/a107ce4b-9ff6-4cf2-ac5e-a227cba4d30b)
![image](https://github.com/user-attachments/assets/d6b5d259-33a5-43b6-a3c8-22a72a5485dd)
![image](https://github.com/user-attachments/assets/e7159320-a143-4195-9b20-318cd073a201)




## License

MIT License

## Author

Ksauraj

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
