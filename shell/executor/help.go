package executor

import (
	"fmt"
	"strings"

	"github.com/bbva/qed/shell/util"
)

// HelpExecutor displays the help for qed-shell
func HelpExecutor(s string) {
	s = strings.TrimPrefix(s, "help")
	command, action := util.GetCommandAndAction(s)
	if command != "" && action != "" {
		printHelpForAction(command, action)
		return
	}
	displayUsage()
}

// displayUsage displays available commands for qed-shell
func displayUsage() {
	fmt.Println("")
	displayCommandHelp("clear", "Clear the screen")
	displayCommandHelp("version", "Display qed-shell version")
	displayCommandHelp("quit | exit", "Exit qed-shell")
	displayCommandHelp("help", "Display this help message")
	fmt.Println("")
}

func displayCommandHelp(command, description string) {
	fmt.Printf(" %s: %s\n", command, description)
}

func printHelpForAction(serviceID, actionID string) {
}
