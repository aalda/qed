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
// #include "extended.h"
import "C"

// FilterPolicy is a factory type for bloom filters. When a filter
// policy is set, every newly created SST file will contain a bloom
// filter which is used to determine if the file may contain the key
// we are looking for, and so, reduce the number of reads. The filter
// is essentially a bit array.
type FilterPolicy interface {
	// keys contains a list of keys (potentially with duplicates)
	// that are ordered according to the user supplied comparator.
	CreateFilter(keys [][]byte) []byte

	// "filter" contains the data appended by a preceding call to
	// CreateFilter(). This method must return true if the key
	// was in the list of keys passed to CreateFilter().
	// This method may return true or false if the key was not on the
	// list, but it should aim to return false with a high probability.
	KeyMayMatch(key []byte, filter []byte) bool

	// Return the name of this policy.
	Name() string
}

// NewNativeFilterPolicy creates a FilterPolicy object.
func NewNativeFilterPolicy(c *C.rocksdb_filterpolicy_t) FilterPolicy {
	return nativeFilterPolicy{c}
}

type nativeFilterPolicy struct {
	c *C.rocksdb_filterpolicy_t
}

func (fp nativeFilterPolicy) CreateFilter(keys [][]byte) []byte {
	return nil
}

func (fp nativeFilterPolicy) KeyMayMatch(key []byte, filter []byte) bool {
	return false

}

func (fp nativeFilterPolicy) Name() string {
	return ""
}

// NewBloomFilterPolicy returns a new filter policy that uses a block-based
// bloom filter with approximately the specified number of bits per key.
// A good value for bits_per_key is 10, which yields a filter with ~1% false
// positive rate.
//
// Note: if you are using a custom comparator that ignores some parts
// of the keys being compared, you must not use NewBloomFilterPolicy()
// and must provide your own FilterPolicy that also ignores the
// corresponding parts of the keys.  For example, if the comparator
// ignores trailing spaces, it would be incorrect to use a
// FilterPolicy (like NewBloomFilterPolicy) that does not ignore
// trailing spaces in keys.
func NewBloomFilterPolicy(bitsPerKey int) FilterPolicy {
	return NewNativeFilterPolicy(
		C.rocksdb_filterpolicy_create_bloom(C.int(bitsPerKey)),
	)
}

// NewFullBloomFilterPolicy returns a new filter policy that uses a full bloom filter
// for the entire SST file, with approximately the specified number of bits per key.
// A good value for bits_per_key is 10, which yields a filter with ~1% false positive rate.
//
// Note: if you are using a custom comparator that ignores some parts
// of the keys being compared, you must not use NewBloomFilterPolicy()
// and must provide your own FilterPolicy that also ignores the
// corresponding parts of the keys.  For example, if the comparator
// ignores trailing spaces, it would be incorrect to use a
// FilterPolicy (like NewBloomFilterPolicy) that does not ignore
// trailing spaces in keys.
func NewFullBloomFilterPolicy(bitsPerKey int) FilterPolicy {
	return NewNativeFilterPolicy(
		C.rocksdb_filterpolicy_create_bloom_full(C.int(bitsPerKey)),
	)
}

// Hold references to filter policies.
var filterPolicies = NewCOWList()

type filterPolicyWrapper struct {
	name         *C.char
	filterPolicy FilterPolicy
}

func registerFilterPolicy(fp FilterPolicy) int {
	return filterPolicies.Append(filterPolicyWrapper{C.CString(fp.Name()), fp})
}

//export rocksdb_filterpolicy_create_filter
func rocksdb_filterpolicy_create_filter(idx int, cKeys **C.char, cKeysLen *C.size_t, cNumKeys C.int, cDstLen *C.size_t) *C.char {
	rawKeys := charSlice(cKeys, cNumKeys)
	keysLen := sizeSlice(cKeysLen, cNumKeys)
	keys := make([][]byte, int(cNumKeys))
	for i, len := range keysLen {
		keys[i] = charToBytes(rawKeys[i], len)
	}

	dst := filterPolicies.Get(idx).(filterPolicyWrapper).filterPolicy.CreateFilter(keys)
	*cDstLen = C.size_t(len(dst))
	return cByteSlice(dst)
}

//export rocksdb_filterpolicy_key_may_match
func rocksdb_filterpolicy_key_may_match(idx int, cKey *C.char, cKeyLen C.size_t, cFilter *C.char, cFilterLen C.size_t) C.uchar {
	key := charToBytes(cKey, cKeyLen)
	filter := charToBytes(cFilter, cFilterLen)
	return boolToUchar(filterPolicies.Get(idx).(filterPolicyWrapper).filterPolicy.KeyMayMatch(key, filter))
}

//export rocksdb_filterpolicy_name
func rocksdb_filterpolicy_name(idx int) *C.char {
	return filterPolicies.Get(idx).(filterPolicyWrapper).name
}
