/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
)

func newClientCommand(ctx *cmdContext) *cobra.Command {
	clientCtx := &clientContext{}
	clientCtx.config = client.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client mode for qed",
		Long:  `Client process for emitting events to a qed server`,
	}

	f := cmd.PersistentFlags()
	f.StringSliceVarP(&clientCtx.config.Endpoints, "endpoints", "e", []string{"127.0.0.1:8800"}, "Endpoint for REST requests on (host:port)")
	f.BoolVar(&clientCtx.config.Insecure, "insecure", false, "Allow self signed certificates")
	f.DurationVar(&clientCtx.config.Timeout, "timeout-seconds", 10*time.Second, "Seconds to cut the connection")
	f.DurationVar(&clientCtx.config.DialTimeout, "dial-timeout-seconds", 5*time.Second, "Seconds to cut the dialing")
	f.DurationVar(&clientCtx.config.HandshakeTimeout, "handshake-timeout-seconds", 5*time.Second, "Seconds to cut the handshaking")

	// Lookups
	v.BindPFlag("client.endpoints", f.Lookup("endpoints"))
	v.BindPFlag("client.insecure", f.Lookup("insecure"))
	v.BindPFlag("client.timeout.connection", f.Lookup("timeout-seconds"))
	v.BindPFlag("client.timeout.dial", f.Lookup("dial-timeout-seconds"))
	v.BindPFlag("client.timeout.handshake", f.Lookup("handshake-timeout-seconds"))

	clientPreRun := func(cmd *cobra.Command, args []string) {

		log.SetLogger("QEDClient", ctx.logLevel)

		clientCtx.config.APIKey = ctx.apiKey
		clientCtx.config.Endpoints = v.GetStringSlice("client.endpoints")
		clientCtx.config.Insecure = v.GetBool("client.insecure")
		clientCtx.config.Timeout = v.GetDuration("client.timeout.connection")
		clientCtx.config.DialTimeout = v.GetDuration("client.timeout.dial")
		clientCtx.config.HandshakeTimeout = v.GetDuration("client.timeout.handshake")
		clientCtx.config.ReadPreference = client.Any
		clientCtx.config.EnableTopologyDiscovery = false
		clientCtx.config.EnableHealthChecks = false
		clientCtx.config.MaxRetries = 0

		client, err := client.NewHTTPClientFromConfig(clientCtx.config)
		if err != nil {
			panic(fmt.Sprintf("Unable to start http client: %v", err))
		}
		clientCtx.client = client
	}

	cmd.AddCommand(
		newAddCommand(clientCtx, clientPreRun),
		newMembershipCommand(clientCtx, clientPreRun),
		newIncrementalCommand(clientCtx, clientPreRun),
	)

	return cmd
}
