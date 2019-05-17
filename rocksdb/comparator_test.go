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

package rocksdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComparator(t *testing.T) {

	db := newTestDB(t, "TestComparator", func(opts *Options) {
		opts.SetComparator(&bytesReverseComparator{})
	})
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, key := range givenKeys {
		require.NoError(t, db.Put(wo, key, []byte("value")))
	}

	// create a iterator to collect the keys
	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	// we seek to the last key and iterate in reverse order
	// to match given keys
	var actualKeys [][]byte
	for iter.SeekToLast(); iter.Valid(); iter.Prev() {
		key := make([]byte, 4)
		copy(key, iter.Key().Data())
		actualKeys = append(actualKeys, key)
	}
	require.NoError(t, iter.Err())

	// ensure that the order is correct
	require.Equal(t, actualKeys, givenKeys)

}

type bytesReverseComparator struct{}

func (cmp *bytesReverseComparator) Name() string {
	return "rocksdb.bytes-reverse"
}

func (cmp *bytesReverseComparator) Compare(a, b []byte) int {
	return bytes.Compare(a, b) * -1
}
