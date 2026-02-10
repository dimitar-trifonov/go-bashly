package generate

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// FormatResult holds the outcome of script formatting.
type FormatResult struct {
	Formatted string
	Error     string
}

// FormatScript applies internal or external formatter to script content.
// Matches bashly_formatting_pipeline.elst.cue logic: tab indentation, internal formatter, external formatter.
func FormatScript(content string, formatter string, tabIndent bool) FormatResult {
	// Apply tab indentation first
	if tabIndent {
		content = strings.ReplaceAll(content, "  ", "\t")
	}

	// Choose formatter
	switch formatter {
	case "internal":
		return FormatResult{Formatted: removeExcessNewlines(content), Error: ""}
	case "none":
		return FormatResult{Formatted: content, Error: ""}
	default:
		// External formatter command
		cmd := exec.Command(formatter)
		cmd.Stdin = strings.NewReader(content)
		var out bytes.Buffer
		cmd.Stdout = &out
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return FormatResult{
				Formatted: "",
				Error:     fmt.Sprintf("formatter failed: %v (stderr: %s)", err, stderr.String()),
			}
		}
		return FormatResult{Formatted: out.String(), Error: ""}
	}
}

// removeExcessNewlines removes consecutive blank lines (internal formatter).
// Matches bashly_formatting_pipeline.elst.cue logic: collapse multiple blank lines.
func removeExcessNewlines(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	prevBlank := false

	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue // skip consecutive blank lines
		}
		result = append(result, line)
		prevBlank = isBlank
	}

	return strings.Join(result, "\n")
}
