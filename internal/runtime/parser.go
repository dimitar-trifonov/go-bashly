package runtime

import (
	"fmt"
	"strings"

	"github.com/dimitar-trifonov/go-bashly/internal/commandmodel"
	"github.com/dimitar-trifonov/go-bashly/internal/settings"
)

// ParsedArgs represents the result of parsing command line arguments.
type ParsedArgs struct {
	Command    *commandmodel.Command
	Flags      map[string]string // long/short flag -> value
	Positional []string          // positional arguments
	Remaining  []string          // arguments after command resolution
	HelpAsked  bool              // true if --help or -h was present
}

// ParseArgs parses argv according to bashly semantics.
// It recognizes --help/-h globally, resolves command path, parses flags and positional args.
func ParseArgs(argv []string, root *commandmodel.Command, st settings.Settings) (*ParsedArgs, error) {
	p := &ParsedArgs{
		Flags:      make(map[string]string),
		Positional: []string{},
		Remaining:  []string{},
	}

	// 1) Global --help detection (before any command-specific parsing)
	if contains(argv, "--help") || contains(argv, "-h") {
		p.HelpAsked = true
		p.Command = root
		return p, nil
	}

	// 2) Resolve command path (first matching command/alias)
	cmd, remaining := resolveCommandPath(root, argv)
	if cmd == nil {
		return nil, fmt.Errorf("unknown command")
	}
	p.Command = cmd
	p.Remaining = remaining

	// 3) Parse flags and collect positional args from remaining args
	parseFlagsAndArgs(p, remaining)

	return p, nil
}

// resolveCommandPath walks the command tree using argv and returns the matched command and leftover args.
func resolveCommandPath(root *commandmodel.Command, argv []string) (*commandmodel.Command, []string) {
	current := root
	remaining := argv

	for len(remaining) > 0 {
		next := findChild(current, remaining[0])
		if next == nil {
			break
		}
		current = next
		remaining = remaining[1:]
	}

	return current, remaining
}

// findChild finds a direct child command matching name or alias.
func findChild(parent *commandmodel.Command, name string) *commandmodel.Command {
	for _, child := range parent.Commands {
		// Exact name match
		if child.Name == name {
			return child
		}
		// Alias match (including wildcards like c*)
		for _, alias := range child.Alias {
			if strings.HasPrefix(alias, "*") {
				prefix := strings.TrimSuffix(alias, "*")
				if strings.HasPrefix(name, prefix) {
					return child
				}
			} else if alias == name {
				return child
			}
		}
	}
	return nil
}

// parseFlagsAndArgs parses flags and positional arguments from remaining args.
func parseFlagsAndArgs(p *ParsedArgs, args []string) {
	i := 0
	for i < len(args) {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			// Long flag: --flag or --flag=value
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				p.Flags[parts[0]] = parts[1]
			} else {
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					p.Flags[arg] = args[i+1]
					i++
				} else {
					p.Flags[arg] = "true"
				}
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short flags: -f value or -abc (compact)
			if len(arg) == 2 {
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					p.Flags[arg] = args[i+1]
					i++
				} else {
					p.Flags[arg] = "true"
				}
			} else {
				// Compact: -abc => -a -b -c
				for _, ch := range arg[1:] {
					p.Flags["-"+string(ch)] = "true"
				}
			}
		} else {
			p.Positional = append(p.Positional, arg)
		}
		i++
	}
}

// ValidateArgs checks required args/flags and allowed values.
func ValidateArgs(p *ParsedArgs) error {
	// Required arguments
	for _, arg := range p.Command.Args {
		if arg.Required && !contains(p.Positional, arg.Name) {
			return fmt.Errorf("missing required argument: %s", arg.Name)
		}
	}

	// Required flags
	for _, flag := range p.Command.Flags {
		if flag.Required {
			value := p.Flags[flag.Long]
			if value == "" {
				value = p.Flags[flag.Short]
			}
			if value == "" {
				name := flag.Long
				if name == "" {
					name = flag.Short
				}
				return fmt.Errorf("missing required flag: %s", name)
			}
		}
	}

	// Allowed values
	for _, flag := range p.Command.Flags {
		value := p.Flags[flag.Long]
		if value == "" {
			value = p.Flags[flag.Short]
		}
		if value != "" && len(flag.Allowed) > 0 && !contains(flag.Allowed, value) {
			name := flag.Long
			if name == "" {
				name = flag.Short
			}
			return fmt.Errorf("invalid value for %s: %s", name, value)
		}
	}

	return nil
}

// contains is a small helper for string slice membership.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
