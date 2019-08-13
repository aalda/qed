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

package main

import (
	"fmt"
	"os"

	"github.com/bbva/qed/client"
	"github.com/spf13/cobra"
)

var checkerCmd = &cobra.Command{
	Use:   "consistency-checker",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		qed, err := client.NewHTTPClient(
			client.SetURLs(conf.endpoints[0], conf.endpoints[1:]...),
			client.SetSnapshotStoreURL(conf.snapshotStoreURL),
			client.SetReadPreference(client.Any),
		)
		if err != nil {
			fmt.Printf("Unable to create QED client: %v\n", err)
			return err
		}
		defer qed.Close()

		for end := conf.rangeStart; end < conf.rangeEnd; {
			_, err := qed.Incremental(conf.rangeStart, end)
			if err != nil {
				fmt.Printf("Unable to get proof: %v\n", err)
				return err
			}
			end = end + conf.interval
		}

		return nil
	},
}

type config struct {
	endpoints        []string
	snapshotStoreURL string
	rangeStart       uint64
	rangeEnd         uint64
	interval         uint64
}

func defaultConfig() *config {
	return &config{
		rangeStart: 0,
		rangeEnd:   0,
		interval:   0,
	}
}

var conf *config

func init() {
	conf = defaultConfig()
	checkerCmd.Flags().StringArrayVar(&conf.endpoints, "endpoints", nil, "QED endpoints")
	checkerCmd.Flags().StringVar(&conf.snapshotStoreURL, "snapshot-store-url", "", "Snapshot Store URL")
	checkerCmd.Flags().Uint64Var(&conf.rangeStart, "range-start", 0, "Range start")
	checkerCmd.Flags().Uint64Var(&conf.rangeEnd, "range-end", 0, "Range end")
	checkerCmd.Flags().Uint64Var(&conf.interval, "range-interval", 0, "Interval")
	checkerCmd.MarkFlagRequired("endpoints")
	//checkerCmd.MarkFlagRequired("snapshot-store-url")
	checkerCmd.MarkFlagRequired("range-start")
	checkerCmd.MarkFlagRequired("range-end")
	checkerCmd.MarkFlagRequired("interval")
}

func main() {
	if err := checkerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
