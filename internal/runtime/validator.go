package runtime

import (
	"github.com/dimitar-trifonov/go-bashly/internal/commandmodel"
)

// ValidateResult holds the outcome of validation.
type ValidateResult struct {
	Valid    bool
	ErrorMsg string
	ExitCode int
}

// ValidateParsed checks required args/flags and allowed values.
// Matches bashly_validation_ux.elst.cue logic: required args, required flags, allowed values.
func ValidateParsed(cmd *commandmodel.Command, parsed *ParsedArgs) ValidateResult {
	// Check required arguments
	for _, arg := range cmd.Args {
		if arg.Required && !contains(parsed.Positional, arg.Name) {
			return ValidateResult{
				Valid:    false,
				ErrorMsg: "missing required argument: " + arg.Name,
				ExitCode: 2,
			}
		}
	}

	// Check required flags
	for _, flag := range cmd.Flags {
		if flag.Required {
			value := parsed.Flags[flag.Long]
			if value == "" {
				value = parsed.Flags[flag.Short]
			}
			if value == "" {
				name := flag.Long
				if name == "" {
					name = flag.Short
				}
				return ValidateResult{
					Valid:    false,
					ErrorMsg: "missing required flag: " + name,
					ExitCode: 2,
				}
			}
		}
	}

	// Check allowed values
	for _, flag := range cmd.Flags {
		value := parsed.Flags[flag.Long]
		if value == "" {
			value = parsed.Flags[flag.Short]
		}
		if value != "" && len(flag.Allowed) > 0 && !contains(flag.Allowed, value) {
			name := flag.Long
			if name == "" {
				name = flag.Short
			}
			return ValidateResult{
				Valid:    false,
				ErrorMsg: "invalid value for " + name + ": " + value,
				ExitCode: 2,
			}
		}
	}

	return ValidateResult{Valid: true, ErrorMsg: "", ExitCode: 0}
}
