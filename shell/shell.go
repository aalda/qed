package shell

import (
	"context"
	"fmt"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/protocol"
	"github.com/c-bata/go-prompt"
	"github.com/c-bata/go-prompt/completer"
)

var (
	version  string
	revision string
)

type Config struct {
	QEDURL            string
	APIKey            string
	SnapshotsStoreURL string
}

type Shell struct {
	conf *Config

	client         *client.HTTPClient
	snapshotsStore map[uint64]*protocol.SignedSnapshot
}

func NewShell(conf *Config) (*Shell, error) {

	clientCfg := client.DefaultConfig()
	clientCfg.APIKey = "blah"
	clientCfg.Endpoints = []string{"http://localhost:8800"}
	clientCfg.Insecure = true

	client := client.NewHTTPClient(*clientCfg)

	return &Shell{
		conf:           conf,
		client:         client,
		snapshotsStore: make(map[uint64]*protocol.SignedSnapshot),
	}, nil
}

func (s *Shell) Run() error {

	ctx := context.WithValue(context.Background(), "client", s.client)

	displayBanner()
	prompt.New(
		Executor(ctx, prompt.NewStdoutWriter()),
		Completer,
		prompt.OptionTitle("qed-shell: the interactive QED client"),
		prompt.OptionPrefix("qed> "),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	).Run()

	return nil
}

func displayBanner() {
	fmt.Printf("\nqed-shell %s (rev-%s): the interactive QED client\n", version, revision)
	fmt.Printf("\nPlease use `exit`, `quit` or `Ctrl-D` to exit this program.\n")
	fmt.Printf("Type `help` for help.\n\n")
}
