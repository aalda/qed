package rocks

import (
	"bytes"
	"fmt"
)

type VersionedKeyComparator struct{}

func NewVersionedKeyComparator() *VersionedKeyComparator {
	return &VersionedKeyComparator{}
}

// Compare implements the db.Comparer interface. This is used to compare raw SST
// keys in the iterator and assumes that all keys compared are versioned keys.
func (c VersionedKeyComparator) Compare(a, b []byte) int {

	// We assume every key in these SSTs is a rocksdb "internal" key with an 8b
	// suffix and need to remove those to compare user keys below.
	if len(a) < 8 || len(b) < 8 {
		// Special case: either key is empty, so bytes.Compare should work.
		if len(a) == 0 || len(b) == 0 {
			return bytes.Compare(a, b)
		}
		panic(fmt.Sprintf("invalid keys: compare expects internal keys with 8b suffix: a: %v b: %v", a, b))
	}

	a1 := a[:len(a)-8]
	a2 := a[len(a)-8:]
	b1 := b[:len(b)-8]
	b2 := b[len(b)-8:]

	if c1 := bytes.Compare(a1, b1); c1 != 0 {
		return c1
	}

	if len(a2) == 0 {
		if len(b2) == 0 {
			return 0
		}
		return -1
	} else if len(b2) == 0 {
		return 1
	}

	if c2 := bytes.Compare(b2, a2); c2 != 0 {
		return c2
	}
	// If versioned keys are the same, fallback to comparing raw internal keys
	// in case the internal suffix differentiates them.
	return bytes.Compare(a, b)
}

func (c VersionedKeyComparator) Name() string {
	return "VersionedKeyComparator"
}
