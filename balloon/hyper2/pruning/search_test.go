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

package pruning

import (
	"testing"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneToFind(t *testing.T) {

	testCases := []struct {
		index         []byte
		cachedBatches map[string][]byte
		storedBatches map[string][]byte
		expectedOps   []op
	}{
		{
			// search for index=0 on an empty tree
			index:         []byte{0},
			cachedBatches: map[string][]byte{},
			storedBatches: map[string][]byte{},
			expectedOps: []op{
				{NoOpCode, pos(0, 8)}, // empty audit path
			},
		},
		{
			// search for index=0 on a tree with one leaf (index=0, value=0)
			index: []byte{0},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{CollectHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{CollectHashCode, pos(64, 6)},
				{GetDefaultHashCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{CollectHashCode, pos(32, 5)},
				{GetDefaultHashCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{CollectHashCode, pos(16, 4)},
				{GetDefaultHashCode, pos(16, 4)},
				{GetProvidedHashCode, pos(0, 4)}, // we stop traversing at the shortcut (index=0)
			},
		},
		{
			// search for index=1 on tree with 1 leaf (index=0, value=0)
			// we traverse until the previous shortcut position even if the leaf does not exist
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{CollectHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{CollectHashCode, pos(64, 6)},
				{GetDefaultHashCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{CollectHashCode, pos(32, 5)},
				{GetDefaultHashCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{CollectHashCode, pos(16, 4)},
				{GetDefaultHashCode, pos(16, 4)},
				{GetProvidedHashCode, pos(0, 4)}, // stop at the position of the shorcut (index=0)
			},
		},
		{
			// search for index=1 on tree with 2 leaves ([index=0, value=0], [index=1, value=1])
			// we traverse until the end of the tree
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x01, 0x00, // iBatch 0 -> hash=0x01
					0x01, 0x00, // iBatch 1 -> hash=0x01
					0x01, 0x00, // iBatch 3 -> hash=0x01
					0x01, 0x00, // iBatch 7 -> hash=0x01
					0x01, 0x00, // iBatch 15 -> hash=0x01
				},
			},
			storedBatches: map[string][]byte{
				pos(1, 0).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x01, 0x01, // iBatch 0 -> hash=0x01 (shortcut index=0)
					0x01, 0x02, // iBatch 1 -> key=0x01
					0x01, 0x02, // iBatch 2 -> value=0x01
				},
				pos(0, 0).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
				pos(0, 4).StringId(): []byte{
					0xd1, 0x01, 0x80, 0x00, // bitmap: 11010001 00000001 10000000 00000000
					0x01, 0x00, // iBatch 0 -> hash=0x01
					0x01, 0x00, // iBatch 1 -> hash=0x01
					0x01, 0x00, // iBatch 3 -> hash=0x01
					0x01, 0x00, // iBatch 7 -> hash=0x01
					0x00, 0x00, // iBatch 15 -> hash=0x00
					0x01, 0x00, // iBatch 16 -> hash=0x01
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{CollectHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{CollectHashCode, pos(64, 6)},
				{GetDefaultHashCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{CollectHashCode, pos(32, 5)},
				{GetDefaultHashCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{CollectHashCode, pos(16, 4)},
				{GetDefaultHashCode, pos(16, 4)},
				{InnerHashCode, pos(0, 4)},
				{CollectHashCode, pos(8, 3)},
				{GetDefaultHashCode, pos(8, 3)},
				{InnerHashCode, pos(0, 3)},
				{CollectHashCode, pos(4, 2)},
				{GetDefaultHashCode, pos(4, 2)},
				{InnerHashCode, pos(0, 2)},
				{CollectHashCode, pos(2, 1)},
				{GetDefaultHashCode, pos(2, 1)},
				{InnerHashCode, pos(0, 1)},
				{GetProvidedHashCode, pos(1, 0)}, // shortcut found but not collected
				{CollectHashCode, pos(0, 0)},
				{GetProvidedHashCode, pos(0, 0)}, // we take the hash of the index=0 position from the batch
			},
		},
		{
			// search for index=8 on tree with 1 leaf (index: 0, value: 0)
			// we traverse until the previous shortcut position even if the leaf does not exist
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{CollectHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{CollectHashCode, pos(64, 6)},
				{GetDefaultHashCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{CollectHashCode, pos(32, 5)},
				{GetDefaultHashCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{CollectHashCode, pos(16, 4)},
				{GetDefaultHashCode, pos(16, 4)},
				{GetProvidedHashCode, pos(0, 4)}, // stop at the position of the shorcut (index=0)
			},
		},
		{
			// search for index=12 on tree with 3 leaves ([index:0, value:0], [index:8, value:8], [index:12, value:12])
			index: []byte{12},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x04, 0x00, // iBatch 0 -> hash=0x04
					0x04, 0x00, // iBatch 1 -> hash=0x04
					0x04, 0x00, // iBatch 3 -> hash=0x04
					0x04, 0x00, // iBatch 7 -> hash=0x04
					0x04, 0x00, // iBatch 15 -> hash=0x04
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xfe, 0x1e, 0x00, 0x00, // bitmap: 11111110 00011110 00000000 00000000
					0x04, 0x00, // iBatch 0 -> hash=0x08
					0x00, 0x01, // iBatch 1 -> hash=0x00 (shortcut index=0)
					0x04, 0x00, // iBatch 2 -> hash=0x04
					0x00, 0x02, // iBatch 3 -> key=0x00
					0x00, 0x02, // iBatch 4 -> value=0x00
					0x08, 0x01, // iBatch 5 -> hash=0x08 (shortcut index=8)
					0x0c, 0x01, // iBatch 6 -> hash=0x0c (shortcut index=12)
					0x08, 0x02, // iBatch 11 -> key=0x08
					0x08, 0x02, // iBatch 12 -> value=0x08
					0x0c, 0x02, // iBatch 13 -> key=0x0c
					0x0c, 0x02, // iBatch 14 -> value=0x0c
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{CollectHashCode, pos(128, 7)},
				{GetDefaultHashCode, pos(128, 7)},
				{InnerHashCode, pos(0, 7)},
				{CollectHashCode, pos(64, 6)},
				{GetDefaultHashCode, pos(64, 6)},
				{InnerHashCode, pos(0, 6)},
				{CollectHashCode, pos(32, 5)},
				{GetDefaultHashCode, pos(32, 5)},
				{InnerHashCode, pos(0, 5)},
				{CollectHashCode, pos(16, 4)},
				{GetDefaultHashCode, pos(16, 4)},
				{InnerHashCode, pos(0, 4)},
				{InnerHashCode, pos(8, 3)},
				{GetProvidedHashCode, pos(12, 2)}, // found shortcut index=12
				{CollectHashCode, pos(8, 2)},
				{GetProvidedHashCode, pos(8, 2)}, // shortcut index=8
				{CollectHashCode, pos(0, 3)},
				{GetProvidedHashCode, pos(0, 3)}, // shortcut index=0
			},
		},
		{
			// search for index=128 on tree with one leaf ([index:0, value:0]
			index: []byte{128},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedOps: []op{
				{InnerHashCode, pos(0, 8)},
				{NoOpCode, pos(128, 7)}, // not found
				{CollectHashCode, pos(0, 7)},
				{GetProvidedHashCode, pos(0, 7)}, // we discard the previous path updated by the previous insertion
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4

	for i, c := range testCases {
		loader := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)
		prunedOps := PruneToFind(c.index, loader).List()
		require.Truef(t, len(c.expectedOps) == len(prunedOps), "The size of the pruned ops should match the expected for test case %d", i)
		for j := 0; j < len(prunedOps); j++ {
			assert.Equalf(t, c.expectedOps[j].Code, prunedOps[j].Code, "The pruned operation's code should match for test case %d", i)
			assert.Equalf(t, c.expectedOps[j].Pos, prunedOps[j].Pos, "The pruned operation's position should match for test case %d", i)
		}
	}
}

func TestSearchInterpretation(t *testing.T) {

	testCases := []struct {
		index             []byte
		cachedBatches     map[string][]byte
		storedBatches     map[string][]byte
		expectedAuditPath navigation.AuditPath
	}{
		{
			// search for index=0 on empty tree
			index:             []byte{0},
			cachedBatches:     map[string][]byte{},
			storedBatches:     map[string][]byte{},
			expectedAuditPath: navigation.AuditPath{},
		},
		{
			// search for index=0 on tree with only one leaf
			index: []byte{0},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
			},
		},
		{
			// search for index=1 on tree with 1 leaf (index=0, value=0)
			// we traverse until the previous shortcut position even if the leaf does not exist
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
			},
		},
		{
			// search for index=1 on tree with 2 leaves ([index=0, value=0], [index=1, value=1])
			// we traverse until the end of the tree
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x01, 0x00, // iBatch 0 -> hash=0x01
					0x01, 0x00, // iBatch 1 -> hash=0x01
					0x01, 0x00, // iBatch 3 -> hash=0x01
					0x01, 0x00, // iBatch 7 -> hash=0x01
					0x01, 0x00, // iBatch 15 -> hash=0x01
				},
			},
			storedBatches: map[string][]byte{
				pos(1, 0).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x01, 0x01, // iBatch 0 -> hash=0x01 (shortcut index=0)
					0x01, 0x02, // iBatch 1 -> key=0x01
					0x01, 0x02, // iBatch 2 -> value=0x01
				},
				pos(0, 0).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
				pos(0, 4).StringId(): []byte{
					0xd1, 0x01, 0x80, 0x00, // bitmap: 11010001 00000001 10000000 00000000
					0x01, 0x00, // iBatch 0 -> hash=0x01
					0x01, 0x00, // iBatch 1 -> hash=0x01
					0x01, 0x00, // iBatch 3 -> hash=0x01
					0x01, 0x00, // iBatch 7 -> hash=0x01
					0x00, 0x00, // iBatch 15 -> hash=0x00
					0x01, 0x00, // iBatch 16 -> hash=0x01
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
				pos(8, 3).StringId():   []byte{0x0},
				pos(4, 2).StringId():   []byte{0x0},
				pos(2, 1).StringId():   []byte{0x0},
				pos(0, 0).StringId():   []byte{0x0},
			},
		},
		{
			// search for index=8 on tree with 1 leaf (index: 0, value: 0)
			// we traverse until the previous shortcut position even if the leaf does not exist
			index: []byte{1},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
			},
		},
		{
			// search for index=12 on tree with 2 leaves ([index:0, value:0], [index:8, value:8])
			index: []byte{12},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x04, 0x00, // iBatch 0 -> hash=0x04
					0x04, 0x00, // iBatch 1 -> hash=0x04
					0x04, 0x00, // iBatch 3 -> hash=0x04
					0x04, 0x00, // iBatch 7 -> hash=0x04
					0x04, 0x00, // iBatch 15 -> hash=0x04
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xfe, 0x1e, 0x00, 0x00, // bitmap: 11111110 00011110 00000000 00000000
					0x04, 0x00, // iBatch 0 -> hash=0x08
					0x00, 0x01, // iBatch 1 -> hash=0x00 (shortcut index=0)
					0x04, 0x00, // iBatch 2 -> hash=0x04
					0x00, 0x02, // iBatch 3 -> key=0x00
					0x00, 0x02, // iBatch 4 -> value=0x00
					0x08, 0x01, // iBatch 5 -> hash=0x08 (shortcut index=8)
					0x0c, 0x01, // iBatch 6 -> hash=0x0c (shortcut index=12)
					0x08, 0x02, // iBatch 11 -> key=0x08
					0x08, 0x02, // iBatch 12 -> value=0x08
					0x0c, 0x02, // iBatch 13 -> key=0x0c
					0x0c, 0x02, // iBatch 14 -> value=0x0c
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(128, 7).StringId(): []byte{0x0},
				pos(64, 6).StringId():  []byte{0x0},
				pos(32, 5).StringId():  []byte{0x0},
				pos(16, 4).StringId():  []byte{0x0},
				pos(8, 2).StringId():   []byte{0x8},
				pos(0, 3).StringId():   []byte{0x0},
			},
		},
		{
			// search for index=128 on tree with one leaf ([index:0, value:0]
			index: []byte{128},
			cachedBatches: map[string][]byte{
				pos(0, 8).StringId(): []byte{
					0xd1, 0x01, 0x00, 0x00, // bitmap: 11010001 00000001 00000000 00000000
					0x00, 0x00, // iBatch 0 -> hash=0x00
					0x00, 0x00, // iBatch 1 -> hash=0x00
					0x00, 0x00, // iBatch 3 -> hash=0x00
					0x00, 0x00, // iBatch 7 -> hash=0x00
					0x00, 0x00, // iBatch 15 -> hash=0x00
				},
			},
			storedBatches: map[string][]byte{
				pos(0, 4).StringId(): []byte{
					0xe0, 0x00, 0x00, 0x00, // bitmap: 11100000 00000000 00000000 00000000
					0x00, 0x01, // iBatch 0 -> hash=0x00 (shortcut index=0)
					0x00, 0x02, // iBatch 1 -> key=0x00
					0x00, 0x02, // iBatch 2 -> value=0x00
				},
			},
			expectedAuditPath: navigation.AuditPath{
				pos(0, 7).StringId(): []byte{0x0},
			},
		},
	}

	batchLevels := uint16(1)
	cacheHeightLimit := batchLevels * 4
	defaultHashes := []hashing.Digest{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}}

	for i, c := range testCases {
		cache := cache.NewFakeCache([]byte{0x0})
		batches := NewFakeBatchLoader(c.cachedBatches, c.storedBatches, cacheHeightLimit)

		ops := PruneToFind(c.index, batches)
		ctx := &Context{
			Hasher:        hashing.NewFakeXorHasher(),
			Cache:         cache,
			DefaultHashes: defaultHashes,
			AuditPath:     navigation.NewAuditPath(),
		}

		ops.Pop().Interpret(ops, ctx)
		assert.Equalf(t, c.expectedAuditPath, ctx.AuditPath, "Audit path error in test case %d", i)

	}
}
