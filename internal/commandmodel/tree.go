package commandmodel

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/dimitar-trifonov/go-bashly/internal/settings"
)

type Flag struct {
	Long     string   `json:"long,omitempty"`
	Short    string   `json:"short,omitempty"`
	Required bool     `json:"required"`
	Allowed  []string `json:"allowed,omitempty"`
	Private  bool     `json:"private"`
}

type Arg struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type EnvVar struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func parseFlags(v any) []Flag {
	list, ok := v.([]any)
	if !ok {
		return nil
	}

	out := make([]Flag, 0, len(list))
	for _, raw := range list {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		lng, _ := asString(m["long"])
		shrt, _ := asString(m["short"])
		req, _ := asBool(m["required"])
		priv, _ := asBool(m["private"])
		var allowed []string
		if rawAllowed, ok := m["allowed"]; ok {
			if arr, ok := rawAllowed.([]any); ok {
				for _, a := range arr {
					if s, ok := a.(string); ok {
						allowed = append(allowed, s)
					}
				}
			}
		}
		out = append(out, Flag{Long: lng, Short: shrt, Required: req, Allowed: allowed, Private: priv})
	}
	return out
}

func parseArgs(v any) []Arg {
	list, ok := v.([]any)
	if !ok {
		return nil
	}

	out := make([]Arg, 0, len(list))
	for _, raw := range list {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := asString(m["name"])
		if name == "" {
			continue
		}
		req, _ := asBool(m["required"])
		out = append(out, Arg{Name: name, Required: req})
	}
	return out
}

func parseEnvVars(v any) []EnvVar {
	list, ok := v.([]any)
	if !ok {
		return nil
	}

	out := make([]EnvVar, 0, len(list))
	for _, raw := range list {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		name, _ := asString(m["name"])
		if name == "" {
			continue
		}
		priv, _ := asBool(m["private"])
		out = append(out, EnvVar{Name: name, Private: priv})
	}
	return out
}

type Command struct {
	Name        string     `json:"name"`
	Parents     []string   `json:"parents,omitempty"`
	FullName    string     `json:"full_name"`
	ActionName  string     `json:"action_name"`
	Private     bool       `json:"private"`
	Expose      string     `json:"expose,omitempty"`
	Alias       []string   `json:"alias,omitempty"`
	Filename    string     `json:"filename,omitempty"`
	Description string     `json:"description,omitempty"`
	Args        []Arg      `json:"args,omitempty"`
	Flags       []Flag     `json:"flags,omitempty"`
	EnvVars     []EnvVar   `json:"environment_variables,omitempty"`
	Commands    []*Command `json:"commands,omitempty"`
}

type TreePrintOptions struct {
	ShowDetails   bool
	RevealPrivate bool
}

// DeepCommands returns all commands in the tree, depth-first.
// If includeSelf is true, includes the root node as the first element.
func DeepCommands(root *Command, includeSelf bool) []*Command {
	out := make([]*Command, 0)
	if includeSelf {
		out = append(out, root)
	}
	for _, c := range root.Commands {
		out = append(out, deepCommandsFrom(c)...)
	}
	return out
}

func deepCommandsFrom(c *Command) []*Command {
	out := []*Command{c}
	for _, child := range c.Commands {
		out = append(out, deepCommandsFrom(child)...)
	}
	return out
}

// PrintTree prints a human-friendly command tree representation.
// Intended for Option A "inspect" output.
func PrintTree(w io.Writer, root *Command, opts TreePrintOptions) {
	printTreeNode(w, root, "", true, opts)
}

func printTreeNode(w io.Writer, c *Command, prefix string, isLast bool, opts TreePrintOptions) {
	if c.Private && !opts.RevealPrivate {
		return
	}

	connector := "├─"
	nextPrefix := prefix + "│ "
	if isLast {
		connector = "└─"
		nextPrefix = prefix + "  "
	}

	if prefix == "" {
		// Root
		line := c.FullName
		if opts.ShowDetails {
			line = formatDetails(c, opts)
		}
		fmt.Fprintf(w, "%s\n", line)
	} else {
		line := c.Name
		if opts.ShowDetails {
			line = formatDetails(c, opts)
		}
		fmt.Fprintf(w, "%s%s %s\n", prefix, connector, line)
	}

	for i, child := range c.Commands {
		printTreeNode(w, child, nextPrefix, i == len(c.Commands)-1, opts)
	}
}

func formatDetails(c *Command, opts TreePrintOptions) string {
	parts := []string{c.Name}
	if c.Filename != "" {
		parts = append(parts, "["+c.Filename+"]")
	}
	if c.Private {
		parts = append(parts, "(private)")
	}
	if len(c.Alias) > 1 {
		parts = append(parts, "alias="+strings.Join(c.Alias[1:], ","))
	}

	flagsCount := len(c.VisibleFlags(opts.RevealPrivate))
	if flagsCount > 0 {
		parts = append(parts, fmt.Sprintf("flags=%d", flagsCount))
	}
	envCount := len(c.VisibleEnvVars(opts.RevealPrivate))
	if envCount > 0 {
		parts = append(parts, fmt.Sprintf("env=%d", envCount))
	}
	return strings.Join(parts, " ")
}

func (c *Command) VisibleFlags(revealPrivate bool) []Flag {
	if revealPrivate {
		return c.Flags
	}
	out := make([]Flag, 0, len(c.Flags))
	for _, f := range c.Flags {
		if f.Private {
			continue
		}
		out = append(out, f)
	}
	return out
}

func (c *Command) VisibleEnvVars(revealPrivate bool) []EnvVar {
	if revealPrivate {
		return c.EnvVars
	}
	out := make([]EnvVar, 0, len(c.EnvVars))
	for _, ev := range c.EnvVars {
		if ev.Private {
			continue
		}
		out = append(out, ev)
	}
	return out
}

// BuildFromConfigMap builds a command tree similar to Ruby Script::Command.
// This is intentionally minimal for Option A: "inspect".
func BuildFromConfigMap(cfg map[string]any, st settings.Settings) (*Command, error) {
	name, _ := asString(cfg["name"])
	if name == "" {
		name = "root"
	}

	root := &Command{
		Name:       name,
		Parents:    nil,
		FullName:   name,
		ActionName: "root",
		Private:    false,
	}

	// Root command partial is always root_command.<ext> in Ruby when commands_dir is nil (~).
	ext := st.PartialsExtension
	if ext == "" {
		ext = "sh"
	}
	if st.CommandsDir != "" {
		root.Filename = filepath.Join(st.CommandsDir, "root."+ext)
	} else {
		root.Filename = "root_command." + ext
	}

	root.Description, _ = asString(cfg["description"])
	root.Args = parseArgs(cfg["args"])
	root.Flags = parseFlags(cfg["flags"])
	root.EnvVars = parseEnvVars(cfg["environment_variables"])

	cmds, ok := cfg["commands"]
	if ok {
		list, ok := cmds.([]any)
		if !ok {
			return nil, fmt.Errorf("config.commands must be a list")
		}
		children, err := buildChildren(list, root, st)
		if err != nil {
			return nil, err
		}
		root.Commands = children
	}

	return root, nil
}

func buildChildren(list []any, parent *Command, st settings.Settings) ([]*Command, error) {
	out := make([]*Command, 0, len(list))
	for i, raw := range list {
		opts, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("commands[%d] must be a mapping", i)
		}

		name, _ := asString(opts["name"])
		if name == "" {
			return nil, fmt.Errorf("commands[%d].name is required", i)
		}

		parents := append([]string{}, parent.Parents...)
		parents = append(parents, parent.Name)

		privateVal, _ := asBool(opts["private"])
		expose, _ := asString(opts["expose"])
		desc, _ := asString(opts["description"])

		cmd := &Command{
			Name:        name,
			Parents:     parents,
			FullName:    strings.Join(append(append([]string{}, parents...), name), " "),
			ActionName:  computeActionName(parents, name),
			Private:     privateVal,
			Expose:      expose,
			Alias:       normalizeAlias(opts["alias"], name),
			Filename:    resolveFilename(opts, parents, name, st),
			Description: desc,
		}
		cmd.Args = parseArgs(opts["args"])
		cmd.Flags = parseFlags(opts["flags"])
		cmd.EnvVars = parseEnvVars(opts["environment_variables"])

		if sub, ok := opts["commands"]; ok {
			subList, ok := sub.([]any)
			if !ok {
				return nil, fmt.Errorf("%s.commands must be a list", cmd.FullName)
			}
			children, err := buildChildren(subList, cmd, st)
			if err != nil {
				return nil, err
			}
			cmd.Commands = children
		}

		out = append(out, cmd)
	}
	return out, nil
}

func computeActionName(parents []string, name string) string {
	// Ruby special-cases root; for children: parents[1..] + [name]
	if len(parents) == 0 {
		return "root"
	}
	if len(parents) == 1 {
		return name
	}
	parts := append([]string{}, parents[1:]...)
	parts = append(parts, name)
	return strings.Join(parts, " ")
}

func normalizeAlias(v any, name string) []string {
	alts := []string{}
	switch t := v.(type) {
	case nil:
		// none
	case string:
		if t != "" {
			alts = append(alts, t)
		}
	case []any:
		for _, a := range t {
			if s, ok := a.(string); ok && s != "" {
				alts = append(alts, s)
			}
		}
	}

	out := make([]string, 0, 1+len(alts))
	out = append(out, name)
	out = append(out, alts...)
	return out
}

func resolveFilename(opts map[string]any, parents []string, name string, st settings.Settings) string {
	// Explicit filename wins.
	if s, ok := asString(opts["filename"]); ok && s != "" {
		return s
	}

	action := computeActionName(parents, name)
	ext := st.PartialsExtension
	if ext == "" {
		ext = "sh"
	}

	if st.CommandsDir != "" {
		p := filepath.FromSlash(strings.ReplaceAll(action, " ", "/")) + "." + ext
		return filepath.Join(st.CommandsDir, p)
	}

	// When commands_dir is nil (~), Ruby uses a flat name under source_dir.
	return underscore(strings.ReplaceAll(action, " ", "_")) + "_command." + ext
}

func underscore(s string) string {
	// Very small helper (good enough for inspect output).
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ToLower(s)
	return s
}

func asString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

func asBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}
