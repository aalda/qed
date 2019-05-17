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

// #include "rocksdb/c.h"
import "C"

// A Comparator object provides a total order across slices that are
// used as keys in an sstable or a database.
type Comparator interface {
	// Three-way comparison. Returns value:
	//   < 0 iff "a" < "b",
	//   == 0 iff "a" == "b",
	//   > 0 iff "a" > "b"
	Compare(a, b []byte) int

	// The name of the comparator.
	Name() string
}

// NewNativeComparator creates a Comparator object.
func NewNativeComparator(c *C.rocksdb_comparator_t) Comparator {
	return nativeComparator{c}
}

type nativeComparator struct {
	c *C.rocksdb_comparator_t
}

func (c nativeComparator) Compare(a, b []byte) int {
	return 0
}

func (c nativeComparator) Name() string {
	return ""
}

// Hold references to comparators.
var comparators = NewCOWList()

type comparatorWrapper struct {
	name       *C.char
	comparator Comparator
}

func registerComparator(cmp Comparator) int {
	return comparators.Append(comparatorWrapper{C.CString(cmp.Name()), cmp})
}

//export rocksdb_comparator_compare
func rocksdb_comparator_compare(idx int, cKeyA *C.char, cKeyALen C.size_t, cKeyB *C.char, cKeyBLen C.size_t) C.int {
	keyA := charToBytes(cKeyA, cKeyALen)
	keyB := charToBytes(cKeyB, cKeyBLen)
	return C.int(comparators.Get(idx).(comparatorWrapper).comparator.Compare(keyA, keyB))
}

//export rocksdb_comparator_name
func rocksdb_comparator_name(idx int) *C.char {
	return comparators.Get(idx).(comparatorWrapper).name
}
