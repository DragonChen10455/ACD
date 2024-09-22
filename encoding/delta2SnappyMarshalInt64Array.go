package encoding

import (
	"fmt"
	"github.com/golang/snappy"
)

func marshalInt64Delta2Snappy(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	bb := bbPool.Get()
	bb.B, _ = marshalInt64NearestDelta2(bb.B[:0], a, 64)
	dst = snappy.Encode(nil, bb.B)
	bbPool.Put(bb)
	mt = MarshalTypeDelta2Snappy
	return dst, mt, a[0]
}

func unmarshalInt64Delta2Snappy(dst []int64, src []byte, firstValue int64, itemsCount int) ([]int64, error) {
	valuesDecompressed, err := snappy.Decode(nil, src)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress Snappy delta from %d bytes; src=%X: %w", len(src), src, err)
	}
	dst, err = unmarshalInt64NearestDelta2(nil, valuesDecompressed, firstValue, itemsCount)
	return dst, err
}
