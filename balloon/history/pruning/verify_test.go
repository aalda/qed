package pruning

import (
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/rand"
	"github.com/stretchr/testify/assert"
)

func TestPruneToVerify(t *testing.T) {

	testCases := []struct {
		index, version uint64
		eventDigest    hashing.Digest
		expectedOp     Operation
	}{
		{
			index:       0,
			version:     0,
			eventDigest: hashing.Digest{0x0},
			expectedOp:  leaf(pos(0, 0), 0),
		},
		{
			index:       0,
			version:     1,
			eventDigest: hashing.Digest{0x0},
			expectedOp: inner(pos(0, 1),
				leaf(pos(0, 0), 0),
				getCache(pos(1, 0))),
		},
		{
			index:       1,
			version:     1,
			eventDigest: hashing.Digest{0x1},
			expectedOp: inner(pos(0, 1),
				getCache(pos(0, 0)),
				leaf(pos(1, 0), 1)),
		},
		{
			index:       1,
			version:     2,
			eventDigest: hashing.Digest{0x1},
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					getCache(pos(0, 0)),
					leaf(pos(1, 0), 1)),
				getCache(pos(2, 1)),
			),
		},
		{
			index:       6,
			version:     6,
			eventDigest: hashing.Digest{0x6},
			expectedOp: inner(pos(0, 3),
				getCache(pos(0, 2)),
				inner(pos(4, 2),
					getCache(pos(4, 1)),
					partial(pos(6, 1),
						leaf(pos(6, 0), 6)))),
		},
		{
			index:       1,
			version:     7,
			eventDigest: hashing.Digest{0x1},
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						getCache(pos(0, 0)),
						leaf(pos(1, 0), 1)),
					getCache(pos(2, 1))),
				getCache(pos(4, 2))),
		},
	}

	for _, c := range testCases {
		prunedOp := PruneToVerify(c.index, c.version, c.eventDigest)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case with index %d and version %d", c.index, c.version)
	}

}

func TestPruneToVerifyIncrementalEnd(t *testing.T) {

	testCases := []struct {
		index, version uint64
		expectedOp     Operation
	}{
		{
			index:      0,
			version:    0,
			expectedOp: getCache(pos(0, 0)),
		},
		{
			index:   0,
			version: 1,
			expectedOp: inner(pos(0, 1),
				getCache(pos(0, 0)),
				getCache(pos(1, 0)),
			),
		},
		{
			index:   0,
			version: 2,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					getCache(pos(0, 0)),
					getCache(pos(1, 0)),
				),
				partial(pos(2, 1),
					getCache(pos(2, 0))),
			),
		},
		{
			index:   0,
			version: 3,
			expectedOp: inner(pos(0, 2),
				inner(pos(0, 1),
					getCache(pos(0, 0)),
					getCache(pos(1, 0)),
				),
				inner(pos(2, 1),
					getCache(pos(2, 0)),
					getCache(pos(3, 0)),
				),
			),
		},
		{
			index:   0,
			version: 4,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						getCache(pos(0, 0)),
						getCache(pos(1, 0))),
					getCache(pos(2, 1)),
				),
				partial(pos(4, 2),
					partial(pos(4, 1),
						getCache(pos(4, 0)))),
			),
		},
		{
			index:   0,
			version: 5,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						getCache(pos(0, 0)),
						getCache(pos(1, 0))),
					getCache(pos(2, 1)),
				),
				partial(pos(4, 2),
					inner(pos(4, 1),
						getCache(pos(4, 0)),
						getCache(pos(5, 0)),
					),
				),
			),
		},
		{
			index:   0,
			version: 6,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						getCache(pos(0, 0)),
						getCache(pos(1, 0))),
					getCache(pos(2, 1)),
				),
				inner(pos(4, 2),
					getCache(pos(4, 1)),
					partial(pos(6, 1),
						getCache(pos(6, 0)))),
			),
		},
		{
			index:   0,
			version: 7,
			expectedOp: inner(pos(0, 3),
				inner(pos(0, 2),
					inner(pos(0, 1),
						getCache(pos(0, 0)),
						getCache(pos(1, 0))),
					getCache(pos(2, 1)),
				),
				inner(pos(4, 2),
					getCache(pos(4, 1)),
					inner(pos(6, 1),
						getCache(pos(6, 0)),
						getCache(pos(7, 0)),
					),
				),
			),
		},
	}

	for _, c := range testCases {
		prunedOp := PruneToVerifyIncrementalEnd(c.index, c.version)
		assert.Equalf(t, c.expectedOp, prunedOp, "The pruned operation should match for test case with index %d and version %d", c.index, c.version)
	}

}

func BenchmarkPruneToVerify(b *testing.B) {

	log.SetLogger("BenchmarkPruneToVerify", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToVerify(0, i, rand.Bytes(32))
		assert.NotNil(b, pruned)
	}

}

func BenchmarkPruneToVerifyConsistent(b *testing.B) {

	log.SetLogger("BenchmarkPruneToVerify", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToVerify(i, i, rand.Bytes(32))
		assert.NotNil(b, pruned)
	}

}

func BenchmarkPruneToVerifyIncrementalEnd(b *testing.B) {

	log.SetLogger("BenchmarkPruneToVerifyIncrementalEnd", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToVerifyIncrementalEnd(0, i)
		assert.NotNil(b, pruned)
	}

}

func BenchmarkPruneToVerifyIncrementalStart(b *testing.B) {

	log.SetLogger("BenchmarkPruneToVerifyIncrementalStart", log.SILENT)

	b.ResetTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		pruned := PruneToVerifyIncrementalStart(i)
		assert.NotNil(b, pruned)
	}

}
