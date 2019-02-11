/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package util

import "encoding/binary"

func Uint64AsBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func Uint64AsPaddedBytes(i uint64, n int) []byte {
	return AddPaddingToBytes(Uint64AsBytes(i), n)
}

func Uint16AsBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

func Uint16AsPaddedBytes(i uint16, n int) []byte {
	return AddPaddingToBytes(Uint16AsBytes(i), n)
}

func AddPaddingToBytes(b []byte, n int) []byte {
	if len(b)/8 >= n {
		return b
	}
	tmp := make([]byte, n, n)
	copy(tmp[n-len(b):], b)
	return tmp
}

func BytesAsUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func BytesAsUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}
