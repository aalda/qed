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

package hyper

import (
	"sync"

	"github.com/bbva/qed/log"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/util"
)

const (
	//CacheSize int = (1118481) * ((31 * 33) + 34) // (2^0+2^4 + 2^8 + 2^12 + 2^16 + 2^20) batches * batchSize (31 nodes * 33 bytes + 34 bytes from key)
	CacheSize int = (2000000) * ((31 * 33) + 34)
)

type HyperTree struct {
	store   storage.Store
	cache   cache.ModifiableCache
	hasherF func() hashing.Hasher

	hasher           hashing.Hasher
	cacheHeightLimit uint16
	defaultHashes    []hashing.Digest
	batchLoader      batchLoader

	sync.RWMutex
}

func NewHyperTree(hasherF func() hashing.Hasher, store storage.Store, cache cache.ModifiableCache) *HyperTree {

	hasher := hasherF()
	numBits := hasher.Len()
	cacheHeightLimit := numBits - min(24, (numBits/8)*4)

	tree := &HyperTree{
		store:            store,
		cache:            cache,
		hasherF:          hasherF,
		hasher:           hasher,
		cacheHeightLimit: cacheHeightLimit,
		defaultHashes:    make([]hashing.Digest, numBits),
		batchLoader:      NewDefaultBatchLoader(store, cache, cacheHeightLimit),
	}

	tree.defaultHashes[0] = tree.hasher.Do([]byte{0x0}, []byte{0x0})
	for i := uint16(1); i < hasher.Len(); i++ {
		tree.defaultHashes[i] = tree.hasher.Do(tree.defaultHashes[i-1], tree.defaultHashes[i-1])
	}

	// warm-up cache
	tree.RebuildCache()

	return tree
}

func (t *HyperTree) Add(eventDigest hashing.Digest, version uint64) (hashing.Digest, []*storage.Mutation, error) {
	t.Lock()
	defer t.Unlock()

	//log.Debugf("Adding new event digest %x with version %d", eventDigest, version)

	versionAsBytes := util.Uint64AsBytes(version)

	// build a stack of operations and then interpret it to generate the root hash
	ops := pruneToInsert(eventDigest, versionAsBytes, t.cacheHeightLimit, t.batchLoader)
	ctx := &pruningContext{
		Hasher:        t.hasher,
		Cache:         t.cache,
		DefaultHashes: t.defaultHashes,
		Mutations:     make([]*storage.Mutation, 0),
	}

	rh := ops.Pop().Interpret(ops, ctx)

	return rh, ctx.Mutations, nil
}

func (t *HyperTree) QueryMembership(eventDigest hashing.Digest) (proof *QueryProof, err error) {
	t.Lock()
	defer t.Unlock()

	//log.Debugf("Proving membership for index %d", eventDigest)

	// build a stack of operations and then interpret it to generate the audit path
	ops := pruneToFind(eventDigest, t.batchLoader)
	ctx := &pruningContext{
		Hasher:        t.hasher,
		Cache:         t.cache,
		DefaultHashes: t.defaultHashes,
		AuditPath:     make(AuditPath, 0),
	}

	ops.Pop().Interpret(ops, ctx)

	// ctx.Value is nil if the digest does not exist
	return NewQueryProof(eventDigest, ctx.Value, ctx.AuditPath, t.hasherF()), nil
}

func (t *HyperTree) RebuildCache() {
	t.Lock()
	defer t.Unlock()

	// warm up cache
	log.Info("Warming up hyper cache...")

	// get all nodes at cache limit height
	start := make([]byte, 2+t.hasher.Len()/8)
	end := make([]byte, 2+t.hasher.Len()/8)
	start[1] = byte(t.cacheHeightLimit)
	end[1] = byte(t.cacheHeightLimit + 1)
	nodes, err := t.store.GetRange(storage.HyperTable, start, end)
	if err != nil {
		log.Fatalf("Oops, something went wrong: %v", err)
	}

	// insert every node into cache
	for _, node := range nodes {
		ops := pruneToRebuild(node.Key[2:], node.Value, t.cacheHeightLimit, t.batchLoader)
		ctx := &pruningContext{
			Hasher:        t.hasher,
			Cache:         t.cache,
			DefaultHashes: t.defaultHashes,
		}
		ops.Pop().Interpret(ops, ctx)
	}
}

func (t *HyperTree) Close() {
	t.Lock()
	defer t.Unlock()

	t.cache = nil
	t.hasher = nil
	t.defaultHashes = nil
	t.store = nil
	t.batchLoader = nil
}

func min(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}
