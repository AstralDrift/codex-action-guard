package cli

import "strings"

func normalizeFlagArgs(args []string, valueFlags map[string]bool) []string {
	var flags []string
	var positionals []string
	var afterSeparator []string
	afterDoubleDash := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if afterDoubleDash {
			afterSeparator = append(afterSeparator, arg)
			continue
		}
		if arg == "--" {
			afterDoubleDash = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			flags = append(flags, arg)
			name := strings.TrimPrefix(arg, "--")
			if eq := strings.Index(name, "="); eq >= 0 {
				name = name[:eq]
			}
			if valueFlags[name] && !strings.Contains(arg, "=") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}

	out := append(flags, positionals...)
	if afterDoubleDash {
		out = append(out, "--")
		out = append(out, afterSeparator...)
	}
	return out
}
