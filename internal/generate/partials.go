package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimitar-trifonov/go-bashly/internal/commandmodel"
	"github.com/dimitar-trifonov/go-bashly/internal/settings"
)

type Options struct {
	Workdir string
	Force   bool
	DryRun  bool
}

type Result struct {
	Created []string
	Skipped []string
}

func EnsureCommandPartials(root *commandmodel.Command, st settings.Settings, opts Options) (Result, error) {
	srcDir := filepath.Join(opts.Workdir, st.SourceDir)

	cmds := commandmodel.DeepCommands(root, true)

	res := Result{}
	for _, c := range cmds {
		if c.Filename == "" {
			continue
		}
		path := filepath.Join(srcDir, c.Filename)

		if !opts.Force {
			if _, err := os.Stat(path); err == nil {
				res.Skipped = append(res.Skipped, path)
				continue
			}
		}

		if opts.DryRun {
			res.Created = append(res.Created, path)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return res, fmt.Errorf("create directory: %w", err)
		}

		content := defaultCommandPartialContent(filepath.ToSlash(filepath.Join(st.SourceDir, c.Filename)), c.FullName)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return res, fmt.Errorf("write partial: %w", err)
		}

		res.Created = append(res.Created, path)
	}

	return res, nil
}

func defaultCommandPartialContent(relPath string, fullCommandName string) string {
	// Ruby bashly uses echo statements (not comments) so the generated command function
	// produces helpful output when run.
	b := &strings.Builder{}
	fmt.Fprintf(b, "echo \"# This file is located at '%s'.\"\n", relPath)
	fmt.Fprintf(b, "echo \"# It contains the implementation for the '%s' command.\"\n", fullCommandName)
	fmt.Fprintf(b, "inspect_args\n")
	return b.String()
}
