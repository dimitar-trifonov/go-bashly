package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Env                    string
	SourceDir              string
	ConfigPath             string
	TargetDir              string
	CommandsDir            string // empty means nil (~)
	LibDir                 string
	ExtraLibDirs           []string
	PartialsExtension      string
	TabIndent              bool
	Formatter              string
	EnableHeaderComment    string
	EnableBash3Bouncer     string
	EnableInspectArgs      string
	EnableViewMarkers      string
	EnableDepsArray        string
	EnableEnvVarNamesArray string
	EnableSourcing         string
	PrivateRevealKey       string
}

func Default() Settings {
	return Settings{
		Env:                    "development",
		SourceDir:              "src",
		ConfigPath:             "%{source_dir}/bashly.yml",
		TargetDir:              ".",
		CommandsDir:            "",
		LibDir:                 "lib",
		ExtraLibDirs:           []string{},
		PartialsExtension:      "sh",
		TabIndent:              false,
		Formatter:              "internal",
		EnableHeaderComment:    "always",
		EnableBash3Bouncer:     "always",
		EnableInspectArgs:      "development",
		EnableViewMarkers:      "development",
		EnableDepsArray:        "always",
		EnableEnvVarNamesArray: "always",
		EnableSourcing:         "development",
		PrivateRevealKey:       "",
	}
}

// Load resolves effective settings for a given workdir.
// This is a minimal subset aligned with bashly_settings_resolution.elst.cue.
func Load(workdir string) (Settings, error) {
	wd, err := filepath.Abs(workdir)
	if err != nil {
		return Settings{}, err
	}

	st := Default()

	// 1) Load optional user settings file.

	path := selectUserSettingsPath(wd)
	var user map[string]any
	if path != "" {
		m, err := loadYAMLMap(path)
		if err != nil {
			return Settings{}, err
		}
		user = m
		applyMap(&st, m)
	}

	// 2) Resolve env (config first, then env var override).
	applyEnv(&st)

	// 3) Apply per-env overrides from config (env var precedence remains in effect).
	if user != nil {
		applyPerEnvOverrides(&st, user)
		// Env vars are final authority.
		applyEnv(&st)
	}

	// 4) Interpolate config_path.
	st.ConfigPath = strings.ReplaceAll(st.ConfigPath, "%{source_dir}", st.SourceDir)
	return st, nil
}

func (s Settings) RevealPrivate() bool {
	if strings.TrimSpace(s.PrivateRevealKey) == "" {
		return false
	}
	_, ok := os.LookupEnv(s.PrivateRevealKey)
	return ok
}

func selectUserSettingsPath(wd string) string {
	if p, ok := os.LookupEnv("BASHLY_SETTINGS_PATH"); ok && strings.TrimSpace(p) != "" {
		return p
	}
	p1 := filepath.Join(wd, "bashly-settings.yml")
	if existsFile(p1) {
		return p1
	}
	p2 := filepath.Join(wd, "settings.yml")
	if existsFile(p2) {
		return p2
	}
	return ""
}

func existsFile(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func loadYAMLMap(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("parse settings yaml: %w", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("settings root must be a YAML mapping")
	}
	return m, nil
}

func applyMap(s *Settings, m map[string]any) {
	if v, ok := m["env"].(string); ok && v != "" {
		s.Env = v
	}
	if v, ok := m["source_dir"].(string); ok {
		s.SourceDir = v
	}
	if v, ok := m["config_path"].(string); ok {
		s.ConfigPath = v
	}
	if v, ok := m["target_dir"].(string); ok {
		s.TargetDir = v
	}
	if v, ok := m["commands_dir"]; ok {
		// YAML ~ becomes nil; treat nil as empty string
		if v == nil {
			s.CommandsDir = ""
		} else if sv, ok := v.(string); ok {
			s.CommandsDir = sv
		}
	}
	if v, ok := m["lib_dir"].(string); ok && v != "" {
		s.LibDir = v
	}
	if v, ok := m["extra_lib_dirs"]; ok {
		if v == nil {
			s.ExtraLibDirs = []string{}
		} else if arr, ok := v.([]any); ok {
			extra := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					extra = append(extra, str)
				}
			}
			s.ExtraLibDirs = extra
		}
	}
	if v, ok := m["partials_extension"].(string); ok && v != "" {
		s.PartialsExtension = v
	}
	if v, ok := m["tab_indent"]; ok {
		if v == nil {
			s.TabIndent = false
		} else if bv, ok := v.(bool); ok {
			s.TabIndent = bv
		}
	}
	if v, ok := m["formatter"].(string); ok && v != "" {
		s.Formatter = v
	}
	if v, ok := m["enable_header_comment"].(string); ok && v != "" {
		s.EnableHeaderComment = v
	}
	if v, ok := m["enable_bash3_bouncer"].(string); ok && v != "" {
		s.EnableBash3Bouncer = v
	}
	if v, ok := m["enable_inspect_args"].(string); ok && v != "" {
		s.EnableInspectArgs = v
	}
	if v, ok := m["enable_view_markers"].(string); ok && v != "" {
		s.EnableViewMarkers = v
	}
	if v, ok := m["enable_deps_array"].(string); ok && v != "" {
		s.EnableDepsArray = v
	}
	if v, ok := m["enable_env_var_names_array"].(string); ok && v != "" {
		s.EnableEnvVarNamesArray = v
	}
	if v, ok := m["enable_sourcing"].(string); ok && v != "" {
		s.EnableSourcing = v
	}
	if v, ok := m["private_reveal_key"]; ok {
		if v == nil {
			s.PrivateRevealKey = ""
		} else if sv, ok := v.(string); ok {
			s.PrivateRevealKey = sv
		}
	}
}

func applyPerEnvOverrides(s *Settings, m map[string]any) {
	env := strings.TrimSpace(s.Env)
	if env == "" {
		return
	}

	// All keys except env are eligible for per-env override.
	if v, ok := m["source_dir_"+env].(string); ok {
		s.SourceDir = v
	}
	if v, ok := m["config_path_"+env].(string); ok {
		s.ConfigPath = v
	}
	if v, ok := m["target_dir_"+env].(string); ok {
		s.TargetDir = v
	}
	if v, ok := m["commands_dir_"+env]; ok {
		if v == nil {
			s.CommandsDir = ""
		} else if sv, ok := v.(string); ok {
			s.CommandsDir = sv
		}
	}
	if v, ok := m["lib_dir_"+env].(string); ok && v != "" {
		s.LibDir = v
	}
	if v, ok := m["extra_lib_dirs_"+env]; ok {
		if v == nil {
			s.ExtraLibDirs = []string{}
		} else if arr, ok := v.([]any); ok {
			extra := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					extra = append(extra, str)
				}
			}
			s.ExtraLibDirs = extra
		}
	}
	if v, ok := m["partials_extension_"+env].(string); ok && v != "" {
		s.PartialsExtension = v
	}
	if v, ok := m["tab_indent_"+env]; ok {
		if v == nil {
			s.TabIndent = false
		} else if bv, ok := v.(bool); ok {
			s.TabIndent = bv
		}
	}
	if v, ok := m["formatter_"+env].(string); ok && v != "" {
		s.Formatter = v
	}
	if v, ok := m["enable_header_comment_"+env].(string); ok && v != "" {
		s.EnableHeaderComment = v
	}
	if v, ok := m["enable_bash3_bouncer_"+env].(string); ok && v != "" {
		s.EnableBash3Bouncer = v
	}
	if v, ok := m["enable_inspect_args_"+env].(string); ok && v != "" {
		s.EnableInspectArgs = v
	}
	if v, ok := m["enable_view_markers_"+env].(string); ok && v != "" {
		s.EnableViewMarkers = v
	}
	if v, ok := m["enable_deps_array_"+env].(string); ok && v != "" {
		s.EnableDepsArray = v
	}
	if v, ok := m["enable_env_var_names_array_"+env].(string); ok && v != "" {
		s.EnableEnvVarNamesArray = v
	}
	if v, ok := m["enable_sourcing_"+env].(string); ok && v != "" {
		s.EnableSourcing = v
	}
	if v, ok := m["private_reveal_key_"+env]; ok {
		if v == nil {
			s.PrivateRevealKey = ""
		} else if sv, ok := v.(string); ok {
			s.PrivateRevealKey = sv
		}
	}
}

func applyEnv(s *Settings) {
	if v, ok := os.LookupEnv("BASHLY_ENV"); ok && v != "" {
		s.Env = v
	}
	if v, ok := os.LookupEnv("BASHLY_SOURCE_DIR"); ok {
		s.SourceDir = v
	}
	if v, ok := os.LookupEnv("BASHLY_CONFIG_PATH"); ok {
		s.ConfigPath = v
	}
	if v, ok := os.LookupEnv("BASHLY_TARGET_DIR"); ok {
		s.TargetDir = v
	}
	if v, ok := os.LookupEnv("BASHLY_COMMANDS_DIR"); ok {
		s.CommandsDir = v
	}
	if v, ok := os.LookupEnv("BASHLY_LIB_DIR"); ok {
		s.LibDir = v
	}
	if v, ok := os.LookupEnv("BASHLY_EXTRA_LIB_DIRS"); ok {
		// Split comma-separated string
		parts := strings.Split(v, ",")
		extra := make([]string, 0, len(parts))
		for _, part := range parts {
			extra = append(extra, strings.TrimSpace(part))
		}
		s.ExtraLibDirs = extra
	}
	if v, ok := os.LookupEnv("BASHLY_PARTIALS_EXTENSION"); ok && v != "" {
		s.PartialsExtension = v
	}
	if v, ok := os.LookupEnv("BASHLY_TAB_INDENT"); ok {
		if parsed, ok := parseEnvBool(v); ok {
			s.TabIndent = parsed
		}
	}
	if v, ok := os.LookupEnv("BASHLY_FORMATTER"); ok && v != "" {
		s.Formatter = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_HEADER_COMMENT"); ok && v != "" {
		s.EnableHeaderComment = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_BASH3_BOUNCER"); ok && v != "" {
		s.EnableBash3Bouncer = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_INSPECT_ARGS"); ok && v != "" {
		s.EnableInspectArgs = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_VIEW_MARKERS"); ok && v != "" {
		s.EnableViewMarkers = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_DEPS_ARRAY"); ok && v != "" {
		s.EnableDepsArray = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_ENV_VAR_NAMES_ARRAY"); ok && v != "" {
		s.EnableEnvVarNamesArray = v
	}
	if v, ok := os.LookupEnv("BASHLY_ENABLE_SOURCING"); ok && v != "" {
		s.EnableSourcing = v
	}
	if v, ok := os.LookupEnv("BASHLY_PRIVATE_REVEAL_KEY"); ok {
		s.PrivateRevealKey = v
	}
}

func parseEnvBool(s string) (bool, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "0", "false", "no":
		return false, true
	case "1", "true", "yes":
		return true, true
	default:
		return false, false
	}
}
