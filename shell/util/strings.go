package util

import "strings"

// GetCommandAndAction returns the command and action for a command
func GetCommandAndAction(s string) (command, action string) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, " ")
	var commandSet, actionSet bool
	for i := 0; i < len(parts); i++ {
		trimmedPart := strings.TrimSpace(parts[i])
		if trimmedPart != "" {
			if !commandSet {
				command = trimmedPart
				commandSet = true
			} else if !actionSet {
				action = trimmedPart
				actionSet = true
				return
			}
		}
	}
	return
}
