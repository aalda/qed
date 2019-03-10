package shell

import (
	"context"
	"strings"

	"github.com/bbva/qed/shell/executor"
	"github.com/bbva/qed/shell/util"
	prompt "github.com/c-bata/go-prompt"
)

func Executor(ctx context.Context, w prompt.ConsoleWriter) func(s string) {
	return func(s string) {
		s = strings.TrimSpace(s)
		cmd, _ := util.GetCommandAndAction(s)
		if s == "" {
			return
		}
		switch cmd {
		case "add":
			executor.AddExecutor(ctx, s)
		case "clear":
			executor.ClearExecutor(w)
		case "quit", "exit":
			executor.QuitExecutor()
		case "help":
			executor.HelpExecutor(s)
		case "version":
			executor.VersionExecutor()
		}
		return
	}
}
