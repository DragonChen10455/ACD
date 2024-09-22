package encoding

import (
	"fmt"
	"github.com/bkaradzic/go-lz4"
)

func marshalInt64DeltaLZ4(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	var err error
	bb := bbPool.Get()
	bb.B, _ = marshalInt64NearestDelta(bb.B[:0], a, 64)
	dst, err = lz4.Encode(nil, bb.B)
	if err != nil {
		return
	}
	bbPool.Put(bb)
	mt = MarshalTypeDeltaLZ4
	return dst, mt, a[0]
}

func unmarshalInt64DeltaLZ4(dst []int64, src []byte, firstValue int64, itemsCount int) ([]int64, error) {
	valuesDecompressed, err := lz4.Decode(nil, src)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress LZ4 delta from %d bytes; src=%X: %w", len(src), src, err)
	}
	dst, err = unmarshalInt64NearestDelta(nil, valuesDecompressed, firstValue, itemsCount)
	return dst, err
}
