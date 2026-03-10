// Package allowlist implements command validation for the proxy server
package allowlist

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Allowlist defines which commands and arguments are permitted
type Allowlist struct {
	commands map[string]CommandSpec
}

// CommandSpec defines the specification for an allowed command
type CommandSpec struct {
	Args         []ArgSpec           // Allowed arguments and patterns
	EnvVars      []string            // Allowed environment variables
	WorkingDirs  []string            // Allowed working directories
	MaxOutput    int64               // Maximum output size in bytes
	Timeout      int                 // Maximum execution time in seconds
	NeedsScope   string              // Required scope (e.g., "git:write")
}

// ArgSpec defines an allowed argument pattern
type ArgSpec struct {
	Pattern string // Exact match or glob pattern
	IsGlob  bool   // Whether Pattern is a glob
}

// New creates a new allowlist with default allowed commands
func New() *Allowlist {
	a := &Allowlist{
		commands: make(map[string]CommandSpec),
	}
	a.loadDefaults()
	return a
}

// loadDefaults loads the default set of allowed commands
func (a *Allowlist) loadDefaults() {
	// Git commands
	a.commands["git"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "status"},
			{Pattern: "log"},
			{Pattern: "diff"},
			{Pattern: "branch"},
			{Pattern: "checkout"},
			{Pattern: "fetch"},
			{Pattern: "pull"},
			{Pattern: "push"},
			{Pattern: "clone"},
			{Pattern: "commit"},
			{Pattern: "add"},
			{Pattern: "reset"},
			{Pattern: "merge"},
			{Pattern: "rebase"},
			{Pattern: "remote"},
			{Pattern: "config"},
		},
		EnvVars: []string{
			"GIT_*",
			"HOME",
			"USER",
		},
		Timeout:    300,
		NeedsScope: "git",
	}

	// gt/gastown commands
	a.commands["gt"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "status"},
			{Pattern: "prime"},
			{Pattern: "done"},
			{Pattern: "mail"},
			{Pattern: "tap"},
			{Pattern: "costs"},
			{Pattern: "mesh"},
		},
		EnvVars: []string{
			"GT_*",
			"HOME",
			"USER",
		},
		Timeout: 60,
	}

	// bd (beads) commands
	a.commands["bd"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "status"},
			{Pattern: "list"},
			{Pattern: "show"},
			{Pattern: "add"},
			{Pattern: "done"},
			{Pattern: "close"},
			{Pattern: "mol"},
			{Pattern: "state"},
			{Pattern: "init"},
			{Pattern: "sync"},
			{Pattern: "config"},
		},
		EnvVars: []string{
			"BD_*",
			"HOME",
		},
		Timeout: 60,
	}

	// Python commands (for testing, linting)
	a.commands["python3"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "-m"},
			{Pattern: "pytest"},
			{Pattern: "pip"},
			{Pattern: "mypy"},
			{Pattern: "ruff"},
			{Pattern: "*"}, // Allow script paths
		},
		EnvVars: []string{
			"PYTHON*",
			"PATH",
			"HOME",
			"VIRTUAL_ENV",
		},
		Timeout: 300,
	}

	// Node/npm commands
	a.commands["npm"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "test"},
			{Pattern: "run"},
			{Pattern: "build"},
			{Pattern: "lint"},
			{Pattern: "install"},
		},
		EnvVars: []string{
			"NODE*",
			"NPM*",
			"PATH",
			"HOME",
		},
		Timeout: 300,
	}

	a.commands["npx"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "tsc"},
			{Pattern: "eslint"},
			{Pattern: "next"},
		},
		EnvVars: []string{
			"NODE*",
			"PATH",
			"HOME",
		},
		Timeout: 300,
	}

	// Basic shell commands
	a.commands["ls"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 10,
	}

	a.commands["cat"] = CommandSpec{
		Args:        []ArgSpec{{Pattern: "*"}},
		WorkingDirs: []string{"/workspace", "/tmp", "/etc"},
		MaxOutput:   10 * 1024 * 1024, // 10MB
		Timeout:     10,
	}

	a.commands["pwd"] = CommandSpec{
		Timeout: 5,
	}

	a.commands["echo"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 5,
	}

	a.commands["mkdir"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 10,
	}

	a.commands["cp"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 30,
	}

	a.commands["mv"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 30,
	}

	a.commands["rm"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "-r"},
			{Pattern: "-rf"},
			{Pattern: "-f"},
			{Pattern: "-v"},
			{Pattern: "--recursive"},
			{Pattern: "--force"},
			{Pattern: "--verbose"},
			{Pattern: "*"}, // File paths (validated by SafePath)
		},
		WorkingDirs: []string{"/workspace", "/tmp"},
		MaxOutput:   1024,
		Timeout:     10,
	}

	a.commands["find"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "-name"},
			{Pattern: "-type"},
			{Pattern: "-maxdepth"},
			{Pattern: "-mtime"},
			{Pattern: "-exec"},
			{Pattern: "-print"},
			{Pattern: "-print0"},
			{Pattern: "-o"},
			{Pattern: "-a"},
			{Pattern: "!"},
			{Pattern: "("},
			{Pattern: ")"},
			{Pattern: "*"}, // Paths and values
		},
		WorkingDirs: []string{"/workspace", "/tmp"},
		MaxOutput:   10 * 1024 * 1024, // 10MB
		Timeout:     60,
	}

	a.commands["grep"] = CommandSpec{
		Args:    []ArgSpec{{Pattern: "*"}},
		Timeout: 30,
	}

	a.commands["curl"] = CommandSpec{
		Args: []ArgSpec{
			{Pattern: "-L"},
			{Pattern: "-s"},
			{Pattern: "-S"},
			{Pattern: "-o"},
			{Pattern: "-O"},
			{Pattern: "--output"},
			{Pattern: "--max-time"},
			{Pattern: "--connect-timeout"},
			{Pattern: "-H"},
			{Pattern: "--header"},
			{Pattern: "-X"},
			{Pattern: "--request"},
			{Pattern: "http*"}, // URLs
		},
		WorkingDirs: []string{"/workspace", "/tmp"},
		MaxOutput:   100 * 1024 * 1024, // 100MB for downloads
		Timeout:     300,
	}
}

// IsAllowed checks if a command with arguments is allowed
func (a *Allowlist) IsAllowed(command string, args []string) bool {
	spec, exists := a.commands[command]
	if !exists {
		return false
	}

	// Check each argument against allowed patterns
	for _, arg := range args {
		if !a.isArgAllowed(arg, spec.Args) {
			return false
		}
	}

	return true
}

// isArgAllowed checks if an argument matches any allowed pattern
func (a *Allowlist) isArgAllowed(arg string, allowed []ArgSpec) bool {
	for _, spec := range allowed {
		if spec.IsGlob {
			// Use filepath.Match for glob patterns
			if matched, _ := filepath.Match(spec.Pattern, arg); matched {
				return true
			}
		} else {
			// Exact match or prefix match for flags
			if spec.Pattern == "*" {
				return true
			}
			if arg == spec.Pattern {
				return true
			}
			// Allow flags that start with allowed patterns
			if strings.HasPrefix(arg, "-") && strings.HasPrefix(spec.Pattern, "-") {
				if strings.HasPrefix(arg, spec.Pattern) {
					return true
				}
			}
		}
	}
	return false
}

// AddCommand adds a new command to the allowlist
func (a *Allowlist) AddCommand(name string, spec CommandSpec) error {
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	a.commands[name] = spec
	return nil
}

// RemoveCommand removes a command from the allowlist
func (a *Allowlist) RemoveCommand(name string) {
	delete(a.commands, name)
}

// ListCommands returns a list of allowed command names
func (a *Allowlist) ListCommands() []string {
	commands := make([]string, 0, len(a.commands))
	for name := range a.commands {
		commands = append(commands, name)
	}
	return commands
}

// GetSpec returns the spec for a command
func (a *Allowlist) GetSpec(command string) (CommandSpec, bool) {
	spec, exists := a.commands[command]
	return spec, exists
}
