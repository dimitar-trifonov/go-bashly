package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimitar-trifonov/go-bashly/internal/settings"
)

// MergeLibs discovers and merges lib files from lib_dir and extra_lib_dirs.
// Matches bashly_lib_merge.elst.cue logic: discover, filter .sh files, concatenate.
func MergeLibs(sourceDir, libDir string, extraLibDirs []string) (string, error) {
	var libFiles []string

	// Discover lib files in lib_dir
	libPath := filepath.Join(sourceDir, libDir)
	if entries, err := os.ReadDir(libPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				libFiles = append(libFiles, filepath.Join(libPath, entry.Name()))
			}
		}
	}

	// Discover lib files in extra_lib_dirs
	for _, extraDir := range extraLibDirs {
		if entries, err := os.ReadDir(extraDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
					libFiles = append(libFiles, filepath.Join(extraDir, entry.Name()))
				}
			}
		}
	}

	// Concatenate lib content
	var parts []string
	for _, file := range libFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read lib file %s: %w", file, err)
		}
		parts = append(parts, string(content))
	}

	return strings.Join(parts, "\n"), nil
}

// EmitFeatureToggles generates conditional sections based on enable_* settings.
// Matches bashly_lib_merge.elst.cue logic: inspect args, view markers, deps array, env var names, sourcing.
func EmitFeatureToggles(st settings.Settings) string {
	var b strings.Builder

	// enable_inspect_args
	if isEnabled(st.EnableInspectArgs, st.Env) {
		b.WriteString("inspect_args() {\n")
		b.WriteString("  echo \"args: $@\"\n")
		b.WriteString("}\n\n")
	}

	// enable_view_markers
	if isEnabled(st.EnableViewMarkers, st.Env) {
		b.WriteString("# VIEW MARKERS ENABLED\n")
		b.WriteString("echo 'view markers are on'\n\n")
	}

	// enable_deps_array
	if isEnabled(st.EnableDepsArray, st.Env) {
		b.WriteString("declare -a deps=()\n")
		b.WriteString("# Dependencies array populated by script\n\n")
	}

	// enable_env_var_names_array
	if isEnabled(st.EnableEnvVarNamesArray, st.Env) {
		b.WriteString("declare -a env_var_names=()\n")
		b.WriteString("# Environment variable names array populated by script\n\n")
	}

	// enable_sourcing
	if isEnabled(st.EnableSourcing, st.Env) {
		b.WriteString("# Source additional files if needed\n")
		b.WriteString("# for file in \"${SCRIPT_DIR}/lib/*.sh\"; do\n")
		b.WriteString("#   source \"$file\"\n")
		b.WriteString("# done\n\n")
	}

	return b.String()
}
