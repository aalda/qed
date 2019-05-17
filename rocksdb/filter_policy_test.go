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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterPolicy(t *testing.T) {

	var (
		givenKeys          = [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
		givenFilter        = []byte("key")
		createFilterCalled = false
		keyMayMatchCalled  = false
	)

	policy := &mockFilterPolicy{
		createFilter: func(keys [][]byte) []byte {
			createFilterCalled = true
			require.Equal(t, keys, givenKeys)
			return givenFilter
		},
		keyMayMatch: func(key, filter []byte) bool {
			keyMayMatchCalled = true
			require.Equal(t, key, givenKeys[0])
			require.Equal(t, filter, givenFilter)
			return true
		},
	}

	db := newTestDB(t, "TestFilterPolicy", func(opts *Options) {
		blockOpts := NewDefaultBlockBasedTableOptions()
		blockOpts.SetFilterPolicy(policy)
		opts.SetBlockBasedTableFactory(blockOpts)
	})
	defer db.Close()

	// insert given keys
	wo := NewDefaultWriteOptions()
	for _, key := range givenKeys {
		require.NoError(t, db.Put(wo, key, []byte("value")))
	}

	// flush to trigger the filter creation
	require.NoError(t, db.Flush(NewDefaultFlushOptions()))
	require.True(t, createFilterCalled)

	// test key should match call
	ro := NewDefaultReadOptions()
	val1, err := db.Get(ro, givenKeys[0])
	defer val1.Free()
	require.NoError(t, err)
	require.True(t, keyMayMatchCalled)

}

type mockFilterPolicy struct {
	createFilter func(keys [][]byte) []byte
	keyMayMatch  func(key, filter []byte) bool
}

func (m *mockFilterPolicy) Name() string {
	return "rocksdb.test"
}

func (m *mockFilterPolicy) CreateFilter(keys [][]byte) []byte {
	return m.createFilter(keys)
}

func (m *mockFilterPolicy) KeyMayMatch(key, filter []byte) bool {
	return m.keyMayMatch(key, filter)
}
