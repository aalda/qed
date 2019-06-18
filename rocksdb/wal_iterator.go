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

// #include <stdlib.h>
// #include "rocksdb/c.h"
import "C"
import (
	"errors"
	"unsafe"
)

type WALIterator struct {
	it *C.rocksdb_wal_iterator_t
}

// NewNativeIterator creates a WALIterator object.
func NewNativeWALIterator(c unsafe.Pointer) *WALIterator {
	return &WALIterator{it: (*C.rocksdb_wal_iterator_t)(c)}
}

// Valid returns false only when an Iterator has iterated past either the
// first or the last key in the database. An iterator is either positioned
// at a key/value pair, or not valid.
func (iter *WALIterator) Valid() bool {
	return C.rocksdb_wal_iter_valid(iter.it) != 0
}

// Next moves the iterator to the next sequential key in the database.
// After this call, Valid() is true if the iterator was not positioned
// at the last entry in the source.
// REQUIRES: Valid()
func (iter *WALIterator) Next() {
	C.rocksdb_wal_iter_next(iter.it)
}

func (iter *WALIterator) Status() error {
	var cErr *C.char
	C.rocksdb_wal_iter_status(iter.it, &cErr)
	if cErr != nil {
		errors.New(C.GoString(cErr))
	}
	return nil
}

// Close closes the iterator.
func (iter *WALIterator) Close() {
	C.rocksdb_wal_iter_destroy(iter.it)
	iter.it = nil
}
