package executor

import (
	"context"
	"fmt"

	"github.com/bbva/qed/client"
)

func AddExecutor(ctx context.Context, s string) {
	client := ctx.Value("client").(*client.HTTPClient)
	snapshot, err := client.Add("test1")
	if err != nil {
		fmt.Printf(" [ERROR] %s", err)
	}
	fmt.Printf("\nReceived snapshot with values:\n\n")
	fmt.Printf(" EventDigest: %x\n", snapshot.EventDigest)
	fmt.Printf(" HyperDigest: %x\n", snapshot.HyperDigest)
	fmt.Printf(" HistoryDigest: %x\n", snapshot.HistoryDigest)
	fmt.Printf(" Version: %d\n\n", snapshot.Version)
}
