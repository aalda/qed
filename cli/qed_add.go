package cli

import (
	"github.com/spf13/cobra"
)

func newAddCommand(ctx *Context) *cobra.Command {

	var key, value string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an event",
		Long:  `Add an event to the authenticated data structure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx.Logger().Printf("Adding key [ %s ] with value [ %s ]\n", key, value)

			snapshot := ctx.client.Add(key)
			ctx.logger.Println(snapshot)
			// ctx.Logger().Printf("Reponse Status: %s\n", resp.Status)
			// ctx.Logger().Printf("Reponse Headers: %v\n", resp.Header)
			// body, _ := ioutil.ReadAll(resp.Body)
			// ctx.Logger().Printf("Reponse Body: %v\n", string(body))
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to add")
	cmd.Flags().StringVar(&value, "value", "", "Value to add")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")

	return cmd
}
