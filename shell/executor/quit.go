package executor

import (
	"fmt"
	"os"
)

// QuitExecutor Leaves the shell client
func QuitExecutor() {
	fmt.Printf("\nBye!\n\n")
	os.Exit(0)
}
