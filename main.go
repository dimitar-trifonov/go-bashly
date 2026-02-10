package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dimitar-trifonov/go-bashly/internal/bashlyconfig"
	"github.com/dimitar-trifonov/go-bashly/internal/commandmodel"
	"github.com/dimitar-trifonov/go-bashly/internal/generate"
	"github.com/dimitar-trifonov/go-bashly/internal/settings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "version":
		printVersion()
		os.Exit(0)
	case "inspect":
		runInspect(os.Args[2:])
	case "generate":
		runGenerate(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("go-bashly version 0.1.0")
	fmt.Println("A Go clone of bashly CLI generator")
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "go-bashly - Go clone of bashly")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  go-bashly version")
	fmt.Fprintln(os.Stderr, "  go-bashly inspect [--config <path>] [--workdir <dir>] [--format tree|json]")
	fmt.Fprintln(os.Stderr, "  go-bashly generate [--config <path>] [--workdir <dir>] [--force] [--dry-run]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --config <path>  Path to bashly.yml (default: src/bashly.yml)")
	fmt.Fprintln(os.Stderr, "  --workdir <dir>  Working directory (default: .)")
	fmt.Fprintln(os.Stderr, "  --format <fmt>   Output format for inspect: tree or json (default: tree)")
	fmt.Fprintln(os.Stderr, "  --force         Overwrite existing files")
	fmt.Fprintln(os.Stderr, "  --dry-run       Show what would be generated without writing files")
}

func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "", "Path to bashly.yml")
	workdir := fs.String("workdir", "", "Working directory used to locate settings.yml (defaults to current directory)")
	format := fs.String("format", "tree", "Output format: tree or json")
	_ = fs.Parse(args)

	wd := *workdir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	wd, err := filepath.Abs(wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	st, err := settings.Load(wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	config := *configPath
	if config == "" {
		config = st.ConfigPath
	}

	cfg, err := bashlyconfig.LoadComposedConfig(config, "import", wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	root, err := commandmodel.BuildFromConfigMap(cfg, st)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := writeInspectOutput(os.Stdout, *format, root, st); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func writeInspectOutput(w io.Writer, format string, root *commandmodel.Command, st settings.Settings) error {
	switch format {
	case "tree", "":
		commandmodel.PrintTree(w, root, commandmodel.TreePrintOptions{
			ShowDetails:   true,
			RevealPrivate: st.RevealPrivate(),
		})
		return nil
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(root)
	default:
		return fmt.Errorf("unknown --format: %s (expected tree or json)", format)
	}
}

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "", "Path to bashly.yml")
	workdir := fs.String("workdir", "", "Working directory used to locate settings.yml (defaults to current directory)")
	force := fs.Bool("force", false, "Overwrite existing partial files")
	dryRun := fs.Bool("dry-run", false, "Print planned changes without writing files")
	_ = fs.Parse(args)

	wd := *workdir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	wd, err := filepath.Abs(wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	st, err := settings.Load(wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	config := *configPath
	if config == "" {
		config = st.ConfigPath
	}

	cfg, err := bashlyconfig.LoadComposedConfig(config, "import", wd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	root, err := commandmodel.BuildFromConfigMap(cfg, st)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	res, err := generate.EnsureCommandPartials(root, st, generate.Options{
		Workdir: wd,
		Force:   *force,
		DryRun:  *dryRun,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	master, err := generate.EnsureMasterScript(root, st, generate.Options{
		Workdir: wd,
		Force:   *force,
		DryRun:  *dryRun,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if *dryRun {
		for _, p := range res.Created {
			fmt.Fprintln(os.Stdout, p)
		}
		if master.Written {
			fmt.Fprintln(os.Stdout, master.Path)
		}
		return
	}

	for _, p := range res.Created {
		fmt.Fprintln(os.Stdout, "created:", p)
	}
	if master.Written {
		fmt.Fprintln(os.Stdout, "created:", master.Path)
	}
}
