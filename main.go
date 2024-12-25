package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gookit/color"
)

// Basic data structures
type ShellData struct {
	Histories    map[string][]CommandEntry
	CommonCmds   map[string]int
	TimePatterns map[string]int
	Insights     DetailedInsights
	ShellConfigs map[string]ShellConfig
}

type CommandEntry struct {
	Command    string
	Timestamp  time.Time
	Count      int
	Categories []string
}

type DetailedInsights struct {
	TechnicalProfile TechProfile
	WorkPatterns     WorkPatterns
	ToolUsage        ToolUsage
}

type TechProfile struct {
	PrimaryRole     string
	SecondarySkills []string
	TechStack       []string
	Proficiency     map[string]float64
}

type WorkPatterns struct {
	PeakHours       []int
	CommonWorkflows []string
	Productivity    map[string]float64
}

type ToolUsage struct {
	Editors    map[string]int
	Languages  map[string]int
	BuildTools map[string]int
}

type Logger struct {
	Info  *log.Logger
	Error *log.Logger
}

type ShellConfig struct {
	ConfigFiles map[string]ConfigInfo
	Plugins     []PluginInfo
	Aliases     map[string]string
	Environment map[string]string
}

type ConfigInfo struct {
	Path     string
	Modified time.Time
	Content  string
}

type PluginInfo struct {
	Name        string
	Source      string
	LastUpdated time.Time
}

// Model implementation
type Model struct {
	viewport    viewport.Model
	progress    progress.Model
	loading     bool
	err         error
	shellData   ShellData
	currentView string
	tabs        []string
	activeTab   int
	logger      Logger
}

func initShellData() ShellData {
	return ShellData{
		Histories:    make(map[string][]CommandEntry),
		CommonCmds:   make(map[string]int),
		TimePatterns: make(map[string]int),
		Insights: DetailedInsights{
			TechnicalProfile: TechProfile{
				Proficiency: make(map[string]float64),
			},
			WorkPatterns: WorkPatterns{
				Productivity: make(map[string]float64),
			},
			ToolUsage: ToolUsage{
				Editors:    make(map[string]int),
				Languages:  make(map[string]int),
				BuildTools: make(map[string]int),
			},
		},
		ShellConfigs: make(map[string]ShellConfig),
	}
}

func initialModel() Model {
	// Create log file
	logFile, err := os.OpenFile("shell_analyzer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	logger := Logger{
		Info:  log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		Error: log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	tabs := []string{"Overview", "Tech Profile", "Work Patterns", "Tool Usage"}

	return Model{
		viewport:    viewport.New(100, 30),
		progress:    progress.New(progress.WithDefaultGradient()),
		loading:     true,
		currentView: "main",
		tabs:        tabs,
		activeTab:   0,
		shellData:   initShellData(),
		logger:      logger,
	}
}

// Implement tea.Model interface
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		analyzeShells,
		tea.EnterAltScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, nil
		}
	case ShellData:
		m.loading = false
		m.shellData = msg
		m.logger.Info.Printf("Shell analysis completed. Found %d shell histories", len(msg.Histories))
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	// Minimalist header with updated name
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Render(`
🚀 K8AU SHELL ANALYSER				 
Shell Analytics & Configuration Tool
`)

	if m.loading {
		return header + "\n" + renderLoading()
	}

	var content string
	switch m.tabs[m.activeTab] {
	case "Overview":
		content = renderOverview(m.shellData)
	case "Tech Profile":
		content = renderTechProfile(m.shellData.Insights.TechnicalProfile)
	case "Work Patterns":
		content = renderWorkPatterns(m.shellData.Insights.WorkPatterns)
	case "Tool Usage":
		content = renderToolUsage(m.shellData.Insights.ToolUsage)
	}

	// Add footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\n\nPress 'q' to quit • Use 'tab' to switch tabs • By Ksauraj")

	return fmt.Sprintf("%s\n%s\n%s%s",
		header,
		renderTabs(m.tabs, m.activeTab),
		content,
		footer)
}

// Render functions
func renderLoading() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render("Analyzing your shell history... 🔍")
}

func renderTabs(tabs []string, active int) string {
	var tabsDisplay strings.Builder

	for i, tab := range tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2)

		if i == active {
			style = style.
				Bold(true).
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15"))
		}

		tabsDisplay.WriteString(style.Render(tab))
	}

	return tabsDisplay.String()
}

func renderOverview(data ShellData) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(1)

	var content strings.Builder
	content.WriteString(color.Green.Sprintf("📊 Shell Usage Overview\n\n"))

	for shell, history := range data.Histories {
		content.WriteString(fmt.Sprintf("Shell: %s\n", color.Cyan.Sprint(shell)))
		content.WriteString(fmt.Sprintf("Commands: %d\n", len(history)))

		// Add shell configuration information
		if config, exists := data.ShellConfigs[shell]; exists {
			content.WriteString("\nConfiguration:\n")
			content.WriteString(fmt.Sprintf("• Aliases: %d\n", len(config.Aliases)))
			content.WriteString(fmt.Sprintf("• Plugins: %d\n", len(config.Plugins)))
			content.WriteString(fmt.Sprintf("• Environment Variables: %d\n", len(config.Environment)))

			// List plugins if any
			if len(config.Plugins) > 0 {
				content.WriteString("\nInstalled Plugins:\n")
				for _, plugin := range config.Plugins {
					content.WriteString(fmt.Sprintf("• %s (from %s)\n",
						color.Yellow.Sprint(plugin.Name),
						plugin.Source))
				}
			}

			// List some aliases if any
			if len(config.Aliases) > 0 {
				content.WriteString("\nSome Aliases:\n")
				count := 0
				for alias, command := range config.Aliases {
					if count >= 5 { // Show only first 5 aliases
						break
					}
					content.WriteString(fmt.Sprintf("• %s → %s\n",
						color.Yellow.Sprint(alias),
						command))
					count++
				}
			}
		}
		content.WriteString("\n")
	}

	return style.Render(content.String())
}

func renderTechProfile(profile TechProfile) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(1)

	var content strings.Builder
	content.WriteString(color.Green.Sprintf("💻 Technical Profile\n\n"))

	// Primary Role
	if profile.PrimaryRole != "" {
		content.WriteString(fmt.Sprintf("🎯 Primary Role: %s\n\n",
			color.Cyan.Sprint(profile.PrimaryRole)))
	} else {
		content.WriteString("🎯 Primary Role: Not enough data\n\n")
	}

	// Tech Stack
	content.WriteString("💻 Tech Stack:\n")
	if len(profile.TechStack) > 0 {
		for _, tech := range profile.TechStack {
			content.WriteString(fmt.Sprintf("• %s\n", tech))
		}
	} else {
		content.WriteString("No tech stack data available\n")
	}
	content.WriteString("\n")

	// Secondary Skills
	content.WriteString("🛠️  Secondary Skills:\n")
	if len(profile.SecondarySkills) > 0 {
		for _, skill := range profile.SecondarySkills {
			content.WriteString(fmt.Sprintf("• %s\n", skill))
		}
	} else {
		content.WriteString("No secondary skills data available\n")
	}
	content.WriteString("\n")

	// Proficiency Levels
	content.WriteString("📊 Proficiency Levels:\n")
	if len(profile.Proficiency) > 0 {
		// Sort proficiencies for consistent display
		var items []struct {
			Name  string
			Level float64
		}
		for tech, level := range profile.Proficiency {
			items = append(items, struct {
				Name  string
				Level float64
			}{tech, level})
		}
		// Sort by proficiency level in descending order
		sort.Slice(items, func(i, j int) bool {
			return items[i].Level > items[j].Level
		})

		for _, item := range items {
			bars := int(item.Level * 20)
			if bars < 0 {
				bars = 0
			}
			barStr := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
			content.WriteString(fmt.Sprintf("%-15s %s %.1f%%\n",
				item.Name, barStr, item.Level*100))
		}
	} else {
		content.WriteString("No proficiency data available\n")
	}

	return style.Render(content.String())
}

func renderWorkPatterns(patterns WorkPatterns) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(1)

	var content strings.Builder
	content.WriteString(color.Yellow.Sprintf("⏰ Work Patterns\n\n"))

	// Daily Activity
	content.WriteString("📅 Daily Activity:\n")
	for _, hour := range patterns.PeakHours {
		content.WriteString(fmt.Sprintf("Peak hour: %02d:00\n", hour))
	}
	content.WriteString("\n")

	// Productivity Metrics
	content.WriteString("📈 Productivity Metrics:\n")
	for metric, value := range patterns.Productivity {
		bars := int(value * 20)
		barStr := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
		content.WriteString(fmt.Sprintf("%-20s %s %.1f%%\n", metric, barStr, value*100))
	}
	content.WriteString("\n")

	// Common Workflows
	content.WriteString("🔄 Common Workflows:\n")
	for _, workflow := range patterns.CommonWorkflows {
		content.WriteString(fmt.Sprintf("• %s\n", workflow))
	}

	return style.Render(content.String())
}

func renderToolUsage(usage ToolUsage) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(1)

	var content strings.Builder
	content.WriteString(color.Magenta.Sprintf("🔧 Tool Usage Statistics\n\n"))

	// Calculate total usage
	total := 0
	for _, count := range usage.Editors {
		total += count
	}

	// Editors Section
	content.WriteString("📝 Editors:\n")
	if total > 0 {
		for editor, count := range usage.Editors {
			percentage := float64(count) / float64(total) * 100
			bars := int(percentage / 5)
			if bars < 0 {
				bars = 0
			}
			barStr := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
			content.WriteString(fmt.Sprintf("%-15s: %s (%d uses, %.1f%%)\n", editor, barStr, count, percentage))
		}
	} else {
		content.WriteString("No editor usage data available\n")
	}
	content.WriteString("\n")

	// Languages Section
	content.WriteString("💻 Programming Languages:\n")
	if total > 0 {
		for lang, count := range usage.Languages {
			bars := int(float64(count) / float64(total) * 20)
			if bars < 0 {
				bars = 0
			}
			barStr := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
			content.WriteString(fmt.Sprintf("%-15s: %s (%d uses)\n", lang, barStr, count))
		}
	} else {
		content.WriteString("No language usage data available\n")
	}
	content.WriteString("\n")

	// Build Tools Section
	content.WriteString("🛠️  Build Tools:\n")
	if total > 0 {
		for tool, count := range usage.BuildTools {
			bars := int(float64(count) / float64(total) * 20)
			if bars < 0 {
				bars = 0
			}
			barStr := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
			content.WriteString(fmt.Sprintf("%-15s: %s (%d uses)\n", tool, barStr, count))
		}
	} else {
		content.WriteString("No build tool usage data available\n")
	}

	return style.Render(content.String())
}

// Shell analysis function
func analyzeShells() tea.Msg {
	data := initShellData()

	// Read shell histories
	shellPaths := map[string]string{
		"bash": "~/.bash_history",
		"zsh":  "~/.zsh_history",
		"fish": "~/.local/share/fish/fish_history",
	}

	for shell, path := range shellPaths {
		expandedPath := expandPath(path)
		if history, err := readHistory(expandedPath); err == nil {
			data.Histories[shell] = history
			analyzeCommands(history, &data)
			data.ShellConfigs[shell] = analyzeShellConfigs(shell)
		}
	}

	return data
}

func readHistory(path string) ([]CommandEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []CommandEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if cmd := cleanHistoryLine(line); cmd != "" {
			entries = append(entries, CommandEntry{
				Command:    cmd,
				Timestamp:  time.Now(), // For simplicity
				Categories: categorizeCommand(cmd),
			})
		}
	}

	return entries, scanner.Err()
}

func cleanHistoryLine(line string) string {
	parts := strings.Fields(line)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func categorizeCommand(cmd string) []string {
	categories := []string{}
	patterns := map[string][]string{
		"development": {"git", "docker", "npm", "go", "python"},
		"system":      {"sudo", "systemctl", "ps", "top"},
		"file":        {"ls", "cd", "cp", "mv", "rm"},
	}

	for category, patterns := range patterns {
		for _, pattern := range patterns {
			if strings.HasPrefix(cmd, pattern) {
				categories = append(categories, category)
				break
			}
		}
	}

	return categories
}

func analyzeCommands(entries []CommandEntry, data *ShellData) {
	// Initialize maps for analysis
	langUsage := make(map[string]int)
	toolUsage := make(map[string]int)
	timeOfDay := make(map[int]int)
	commandPatterns := make(map[string]int)

	// Get installed languages
	installedLangs := getInstalledLanguages()

	// Analyze each command
	for _, entry := range entries {
		cmd := entry.Command
		hour := entry.Timestamp.Hour()
		timeOfDay[hour]++

		// Language usage analysis
		for lang := range installedLangs {
			if strings.Contains(cmd, lang) ||
				strings.Contains(cmd, getPackageManager(lang)) {
				langUsage[lang]++
			}
		}

		// Development tool analysis
		tools := []string{"git", "docker", "kubectl", "terraform", "ansible", "make"}
		for _, tool := range tools {
			if strings.HasPrefix(cmd, tool) && checkToolInstalled(tool) {
				toolUsage[tool]++
			}
		}

		// Analyze command patterns
		analyzeCommandPattern(cmd, commandPatterns)
	}

	// Update TechnicalProfile
	techProfile := &data.Insights.TechnicalProfile

	// Calculate primary role based on most used language/tool
	if primaryLang, ok := getMostUsed(langUsage); ok {
		techProfile.PrimaryRole = fmt.Sprintf("%s Developer", strings.Title(primaryLang))
	}

	// Calculate tech stack
	techProfile.TechStack = make([]string, 0)
	for lang := range installedLangs {
		if langUsage[lang] > 0 {
			techProfile.TechStack = append(techProfile.TechStack, lang)
		}
	}

	// Calculate proficiency
	totalCommands := len(entries)
	if totalCommands > 0 {
		for lang, count := range langUsage {
			techProfile.Proficiency[lang] = float64(count) / float64(totalCommands)
		}
		for tool, count := range toolUsage {
			techProfile.Proficiency[tool] = float64(count) / float64(totalCommands)
		}
	}

	// Update WorkPatterns
	patterns := &data.Insights.WorkPatterns
	patterns.PeakHours = getPeakHours(timeOfDay)

	// Calculate productivity metrics based on command complexity and variety
	patterns.Productivity = calculateProductivityMetrics(entries, commandPatterns)
}

func getPackageManager(lang string) string {
	managers := map[string]string{
		"python": "pip",
		"node":   "npm",
		"go":     "go get",
		"rust":   "cargo",
		"ruby":   "gem",
		"php":    "composer",
	}
	return managers[lang]
}

func analyzeCommandPattern(cmd string, patterns map[string]int) {
	// Define common command patterns
	patternMap := map[string]*regexp.Regexp{
		"git_workflow": regexp.MustCompile(`git (commit|push|pull|merge)`),
		"build":        regexp.MustCompile(`(make|build|compile)`),
		"deploy":       regexp.MustCompile(`(deploy|kubectl|docker)`),
		"test":         regexp.MustCompile(`test|spec|pytest`),
	}

	for pattern, regex := range patternMap {
		if regex.MatchString(cmd) {
			patterns[pattern]++
		}
	}
}

func getMostUsed(usage map[string]int) (string, bool) {
	var maxKey string
	var maxVal int
	for k, v := range usage {
		if v > maxVal {
			maxKey = k
			maxVal = v
		}
	}
	return maxKey, maxVal > 0
}

func getPeakHours(timeOfDay map[int]int) []int {
	type hourCount struct {
		hour  int
		count int
	}

	var hours []hourCount
	for h, c := range timeOfDay {
		hours = append(hours, hourCount{h, c})
	}

	sort.Slice(hours, func(i, j int) bool {
		return hours[i].count > hours[j].count
	})

	// Return top 3 peak hours
	var peaks []int
	for i := 0; i < len(hours) && i < 3; i++ {
		peaks = append(peaks, hours[i].hour)
	}
	return peaks
}

func calculateProductivityMetrics(entries []CommandEntry, patterns map[string]int) map[string]float64 {
	metrics := make(map[string]float64)
	totalCommands := len(entries)

	if totalCommands == 0 {
		return metrics
	}

	// Command variety score
	uniqueCommands := make(map[string]bool)
	for _, entry := range entries {
		uniqueCommands[entry.Command] = true
	}
	metrics["Command Variety"] = float64(len(uniqueCommands)) / float64(totalCommands)

	// Workflow complexity score
	workflowScore := float64(patterns["git_workflow"]+patterns["build"]+
		patterns["deploy"]+patterns["test"]) / float64(totalCommands)
	metrics["Workflow Complexity"] = workflowScore

	return metrics
}

func checkToolInstalled(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func getInstalledLanguages() map[string]string {
	languages := map[string]string{
		// Programming Languages
		"python":  "python --version",
		"python3": "python3 --version",
		"node":    "node --version",
		"go":      "go version",
		"java":    "java -version",
		"ruby":    "ruby --version",
		"php":     "php --version",
		"rust":    "rustc --version",
		"perl":    "perl --version",
		"scala":   "scala -version",
		"kotlin":  "kotlin -version",
		"swift":   "swift --version",
		"r":       "R --version",
		"julia":   "julia --version",
		"haskell": "ghc --version",
		"elixir":  "elixir --version",
		"erlang":  "erl -version",
		"clang":   "clang --version",
		"gcc":     "gcc --version",
		"dotnet":  "dotnet --version",
		"lua":     "lua -v",
		"ocaml":   "ocaml -version",
		"dart":    "dart --version",
		"zig":     "zig version",
		"nim":     "nim --version",

		// Build Tools & Package Managers
		"maven":    "mvn --version",
		"gradle":   "gradle --version",
		"npm":      "npm --version",
		"yarn":     "yarn --version",
		"pnpm":     "pnpm --version",
		"pip":      "pip --version",
		"cargo":    "cargo --version",
		"composer": "composer --version",
		"bundler":  "bundle --version",

		// DevOps & Cloud Tools
		"docker":    "docker --version",
		"kubectl":   "kubectl version --client",
		"terraform": "terraform version",
		"ansible":   "ansible --version",
		"vagrant":   "vagrant --version",
		"helm":      "helm version",
		"aws":       "aws --version",
		"gcloud":    "gcloud --version",
		"azure":     "az --version",

		// Version Control
		"git":       "git --version",
		"svn":       "svn --version",
		"mercurial": "hg --version",

		// Databases
		"mysql":   "mysql --version",
		"psql":    "psql --version",
		"mongodb": "mongod --version",
		"redis":   "redis-cli --version",

		// Web Servers & Tools
		"nginx":   "nginx -v",
		"apache2": "apache2 -v",
		"curl":    "curl --version",
		"wget":    "wget --version",

		// Text Editors & IDEs
		"vim":   "vim --version",
		"nvim":  "nvim --version",
		"emacs": "emacs --version",
		"code":  "code --version",

		// Shell & Terminal Tools
		"zsh":  "zsh --version",
		"bash": "bash --version",
		"fish": "fish --version",
		"tmux": "tmux -V",
	}

	installed := make(map[string]string)
	for lang, cmd := range languages {
		if out, err := exec.Command("sh", "-c", cmd).Output(); err == nil {
			installed[lang] = string(out)
		}
	}

	// Sort and keep only top 10 most used
	type usageEntry struct {
		name  string
		count int
	}
	var usageList []usageEntry
	for name := range installed {
		count := 0
		// Count occurrences in command history (you'll need to pass this data somehow)
		// For now, we'll just store all installed ones
		usageList = append(usageList, usageEntry{name, count})
	}

	// Sort by usage count
	sort.Slice(usageList, func(i, j int) bool {
		return usageList[i].count > usageList[j].count
	})

	// Keep only top 10
	result := make(map[string]string)
	for i := 0; i < len(usageList) && i < 10; i++ {
		name := usageList[i].name
		result[name] = installed[name]
	}

	return result
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func analyzeShellConfigs(shell string) ShellConfig {
	configPaths := map[string][]string{
		"bash": {
			"~/.bashrc",
			"~/.bash_profile",
			"~/.bash_aliases",
		},
		"zsh": {
			"~/.zshrc",
			"~/.zsh_plugins",
			"~/.zprofile",
		},
		"fish": {
			"~/.config/fish/config.fish",
			"~/.config/fish/functions",
			"~/.config/fish/conf.d",
		},
	}

	config := ShellConfig{
		ConfigFiles: make(map[string]ConfigInfo),
		Aliases:     make(map[string]string),
		Environment: make(map[string]string),
		Plugins:     make([]PluginInfo, 0),
	}

	// Read and analyze config files
	for _, paths := range configPaths[shell] {
		expandedPath := expandPath(paths)
		if info, err := os.Stat(expandedPath); err == nil {
			content, _ := os.ReadFile(expandedPath)
			config.ConfigFiles[paths] = ConfigInfo{
				Path:     expandedPath,
				Modified: info.ModTime(),
				Content:  string(content),
			}

			// Parse the config file
			parseShellConfig(string(content), &config)
		}
	}

	// Detect plugins based on shell type
	detectPlugins(shell, &config)

	return config
}

func parseShellConfig(content string, config *ShellConfig) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse aliases
		if strings.HasPrefix(line, "alias ") {
			parts := strings.SplitN(strings.TrimPrefix(line, "alias "), "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
				config.Aliases[name] = value
			}
		}

		// Parse environment variables
		if strings.HasPrefix(line, "export ") {
			parts := strings.SplitN(strings.TrimPrefix(line, "export "), "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
				config.Environment[name] = value
			}
		}
	}
}

func detectPlugins(shell string, config *ShellConfig) {
	switch shell {
	case "zsh":
		detectZshPlugins(config)
	case "fish":
		detectFishPlugins(config)
	case "bash":
		detectBashPlugins(config)
	}
}

func detectZshPlugins(config *ShellConfig) {
	// Check for common plugin managers
	pluginManagers := []string{
		"~/.oh-my-zsh",
		"~/.antigen",
		"~/.zinit",
		"~/.zplug",
	}

	for _, manager := range pluginManagers {
		path := expandPath(manager)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			config.Plugins = append(config.Plugins, PluginInfo{
				Name:        filepath.Base(manager),
				Source:      path,
				LastUpdated: info.ModTime(),
			})
		}
	}
}

func detectFishPlugins(config *ShellConfig) {
	fishPluginPath := expandPath("~/.config/fish/conf.d")
	if files, err := os.ReadDir(fishPluginPath); err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".fish") {
				info, _ := file.Info()
				config.Plugins = append(config.Plugins, PluginInfo{
					Name:        strings.TrimSuffix(file.Name(), ".fish"),
					Source:      filepath.Join(fishPluginPath, file.Name()),
					LastUpdated: info.ModTime(),
				})
			}
		}
	}
}

func detectBashPlugins(config *ShellConfig) {
	// Check for common bash plugin managers and extensions
	bashPluginPaths := []string{
		"~/.bash_it",
		"~/.local/share/bash-completion",
	}

	for _, path := range bashPluginPaths {
		expandedPath := expandPath(path)
		if info, err := os.Stat(expandedPath); err == nil && info.IsDir() {
			config.Plugins = append(config.Plugins, PluginInfo{
				Name:        filepath.Base(path),
				Source:      expandedPath,
				LastUpdated: info.ModTime(),
			})
		}
	}
}

func analyzeCommandComplexity(data *ShellData) float64 {
	var totalCommands, complexCommands float64

	for _, history := range data.Histories {
		for _, entry := range history {
			totalCommands++

			// Count pipes and redirections
			if strings.Contains(entry.Command, "|") ||
				strings.Contains(entry.Command, ">") ||
				strings.Contains(entry.Command, "<") {
				complexCommands++
			}

			// Count commands with multiple arguments
			if len(strings.Fields(entry.Command)) > 2 {
				complexCommands += 0.5
			}
		}
	}

	if totalCommands == 0 {
		return 0
	}

	return complexCommands / totalCommands
}

func generateRecommendations(data *ShellData) []string {
	recommendations := []string{}

	// Analyze shell configuration
	for shell, config := range data.ShellConfigs {
		if len(config.Aliases) < 5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider adding more aliases to your %s configuration to improve productivity", shell))
		}

		if len(config.Plugins) < 3 {
			recommendations = append(recommendations,
				fmt.Sprintf("Explore popular %s plugins to enhance your shell experience", shell))
		}
	}

	return recommendations
}

func generateWorkflowTips(data *ShellData) []string {
	tips := []string{}

	// Analyze command patterns
	commonPatterns := analyzeCommandPatterns(data)
	for pattern, count := range commonPatterns {
		if count > 10 {
			tips = append(tips, fmt.Sprintf(
				"You frequently use '%s'. Consider creating an alias for this pattern", pattern))
		}
	}

	return tips
}

func analyzeCommandPatterns(data *ShellData) map[string]int {
	patterns := make(map[string]int)

	for _, history := range data.Histories {
		for _, entry := range history {
			// Look for common command sequences
			parts := strings.Fields(entry.Command)
			if len(parts) > 1 {
				pattern := strings.Join(parts[:2], " ")
				patterns[pattern]++
			}
		}
	}

	return patterns
}

func main() {
	p := tea.NewProgram(initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion())

	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
