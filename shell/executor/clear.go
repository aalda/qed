package executor

import prompt "github.com/c-bata/go-prompt"

func ClearExecutor(console prompt.ConsoleWriter) {
	console.EraseScreen()
	console.CursorGoTo(0, 0)
	console.Flush()
}
