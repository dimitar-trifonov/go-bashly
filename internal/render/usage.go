package render

import (
	"fmt"
	"strings"

	"github.com/dimitar-trifonov/go-bashly/internal/commandmodel"
)

// PrintUsage renders plain-text help for a specific command.
// Matches bashly_usage_render.elst.cue logic: name, description, usage line, args, flags, subcommands.
func PrintUsage(cmd *commandmodel.Command) string {
	var b strings.Builder

	// Command header: name - description
	desc := cmd.Description
	if desc == "" {
		desc = ""
	}
	b.WriteString(fmt.Sprintf("%s - %s\n", cmd.Name, desc))

	// Usage line: Usage: full_name [args...]
	usageLine := "Usage: " + cmd.FullName
	if len(cmd.Args) > 0 {
		argNames := make([]string, 0, len(cmd.Args))
		for _, arg := range cmd.Args {
			argNames = append(argNames, arg.Name)
		}
		usageLine += " " + strings.Join(argNames, " ")
	}
	b.WriteString(usageLine + "\n")

	// Arguments section
	if len(cmd.Args) > 0 {
		b.WriteString("\nArguments:\n")
		for _, arg := range cmd.Args {
			line := "  " + arg.Name
			if arg.Required {
				line += " (required)"
			}
			b.WriteString("\n" + line)
		}
	}

	// Flags section
	if len(cmd.Flags) > 0 {
		b.WriteString("\nFlags:\n")
		for _, flag := range cmd.Flags {
			line := "  "
			if flag.Long != "" {
				line += flag.Long
			}
			if flag.Short != "" {
				if flag.Long != "" {
					line += ", "
				}
				line += flag.Short
			}
			if flag.Required {
				line += " (required)"
			}
			if len(flag.Allowed) > 0 {
				line += " (allowed: " + strings.Join(flag.Allowed, ", ") + ")"
			}
			b.WriteString("\n" + line)
		}
	}

	// Subcommands section
	if len(cmd.Commands) > 0 {
		b.WriteString("\nCommands:\n")
		for _, sub := range cmd.Commands {
			line := "  " + sub.Name
			if len(sub.Alias) > 1 {
				line += " (" + strings.Join(sub.Alias[1:], ", ") + ")"
			}
			b.WriteString("\n" + line)
		}
	}

	return b.String()
}

// PrintGlobalUsage renders top-level help for the root command.
// Matches bashly_usage_render.elst.cue logic: name, description, usage line, commands, global flags.
func PrintGlobalUsage(root *commandmodel.Command) string {
	var b strings.Builder

	// Global header: name - description
	desc := root.Description
	if desc == "" {
		desc = ""
	}
	b.WriteString(fmt.Sprintf("%s - %s\n", root.Name, desc))

	// Global usage line
	b.WriteString("\nUsage: " + root.Name + " <command> [options]\n")

	// Commands section
	if len(root.Commands) > 0 {
		b.WriteString("\nCommands:\n")
		for _, sub := range root.Commands {
			line := "  " + sub.Name
			if len(sub.Alias) > 1 {
				line += " (" + strings.Join(sub.Alias[1:], ", ") + ")"
			}
			b.WriteString("\n" + line)
		}
	}

	// Global flags section
	if len(root.Flags) > 0 {
		b.WriteString("\nGlobal Flags:\n")
		for _, flag := range root.Flags {
			line := "  "
			if flag.Long != "" {
				line += flag.Long
			}
			if flag.Short != "" {
				if flag.Long != "" {
					line += ", "
				}
				line += flag.Short
			}
			if flag.Required {
				line += " (required)"
			}
			if len(flag.Allowed) > 0 {
				line += " (allowed: " + strings.Join(flag.Allowed, ", ") + ")"
			}
			b.WriteString("\n" + line)
		}
	}

	return b.String()
}
