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

	"github.com/spf13/cobra"

	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/storage/rocks"
	"github.com/bbva/qed/util"
)

var serverVerify *cobra.Command = &cobra.Command{
	Use:   "verify",
	Short: "Verify QED Log integrity",
	Long:  "Verify QED Log database has not been corrupted",
	RunE:  runServerVerify,
}

var path string

func init() {
	serverVerify.Flags().StringVar(&path, "db-path", "", "Path to the database")
	serverVerify.MarkFlagRequired("db-path")
	serverCmd.AddCommand(serverVerify)
}

func runServerVerify(cmd *cobra.Command, args []string) error {

	dbOpts := rocks.DefaultOptions()
	dbOpts.Path = path
	dbOpts.ReadOnly = true
	db, err := rocks.NewRocksDBStoreWithOpts(dbOpts)
	if err != nil {
		return err
	}
	defer db.Close()

	last, err := db.GetLast(storage.HistoryTable)
	if last == nil {
		return fmt.Errorf("%v: empty database", err)
	}

	if err != nil {
		return err
	}

	lastIndex := util.BytesAsUint64(last.Key[:8])

	fmt.Printf("Verifying integrity from indexes %d and %d\n", 0, lastIndex)

	reader := db.GetAll(storage.HistoryTable)
	defer reader.Close()

	count := uint64(0)
	for {
		entries := make([]*storage.KVPair, 1000)
		n, err := reader.Read(entries)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			index := util.BytesAsUint64(entries[i].Key[:8])
			height := util.BytesAsUint16(entries[i].Key[8:])
			//fmt.Printf("Index [%d] - Count [%d]\n", index, count)
			if height == uint16(0) {
				for j := count; count < index; j++ {
					fmt.Printf("Gap found: missing index %d\n", count)
					count++
				}
				count++
			}

		}

	}
	if count-1 != lastIndex {
		fmt.Printf("Last index [%d] not reached: %d\n", lastIndex, count-1)
	}

	return nil

}
