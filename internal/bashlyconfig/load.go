package bashlyconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadYAMLFile(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config root must be a YAML mapping")
	}

	return m, nil
}

// LoadComposedConfig loads a YAML file, then applies Bashly-style compose semantics.
// ERB preprocessing is intentionally deferred in the Go clone.
func LoadComposedConfig(path string, keyword string, workdir string) (map[string]any, error) {
	wd, err := filepath.Abs(workdir)
	if err != nil {
		return nil, err
	}

	configPath := path
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(wd, configPath)
	}

	abspath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	v, err := loadAnyYAMLFile(abspath)
	if err != nil {
		return nil, err
	}

	composed, err := composeAny(v, keyword, wd)
	if err != nil {
		return nil, err
	}

	m, ok := composed.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config root must be a YAML mapping")
	}

	return m, nil
}

func loadAnyYAMLFile(path string) (any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read yaml file %s: %w", path, err)
	}

	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("cannot parse yaml file %s: %w", path, err)
	}
	return v, nil
}

func composeAny(v any, keyword string, workdir string) (any, error) {
	switch t := v.(type) {
	case map[string]any:
		return composeMap(t, keyword, workdir)
	case []any:
		out := make([]any, 0, len(t))
		for _, x := range t {
			cx, err := composeAny(x, keyword, workdir)
			if err != nil {
				return nil, err
			}
			out = append(out, cx)
		}
		return out, nil
	default:
		return v, nil
	}
}

func composeMap(m map[string]any, keyword string, workdir string) (any, error) {
	result := map[string]any{}
	for k, v := range m {
		if k == keyword {
			importPath, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("%s must be a string path", keyword)
			}
			resolved := importPath
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(workdir, resolved)
			}
			sub, err := loadAnyYAMLFile(resolved)
			if err != nil {
				// Keep Ruby-like message shape.
				return nil, fmt.Errorf("cannot find import file %s", importPath)
			}
			subComposed, err := composeAny(sub, keyword, workdir)
			if err != nil {
				return nil, err
			}

			subArr, ok := subComposed.([]any)
			if ok {
				return subArr, nil
			}
			subMap, ok := subComposed.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("cannot find a valid YAML in %s", importPath)
			}
			for sk, sv := range subMap {
				result[sk] = sv
			}
			continue
		}

		cv, err := composeAny(v, keyword, workdir)
		if err != nil {
			return nil, err
		}
		result[k] = cv
	}
	return result, nil
}
