package encoding

import (
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding/zstd"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

func marshalInt64DeltaXorZSTD(dst []byte, a []int64, precisionBits uint8) (result []byte, mt MarshalType, firstValue int64) {
	if len(a) == 0 {
		logger.Panicf("BUG: a must contain at least one item")
	}
	if isConst(a) {
		firstValue = a[0]
		return dst, MarshalTypeConst, firstValue
	}
	if isDeltaConst(a) {
		firstValue = a[0]
		dst = MarshalVarInt64(dst, a[1]-a[0])
		return dst, MarshalTypeDeltaConst, firstValue
	}

	bb := bbPool.Get()
	// Non-gauge values, i.e. counters are better compressed with delta2 encoding.
	mt = MarshalTypeZSTDDeltaXor
	bb.B, firstValue = marshalInt64DeltaXor(bb.B[:0], a, precisionBits)

	// Try compressing the result.
	dstOrig := dst
	if len(bb.B) >= minCompressibleBlockSize {
		compressLevel := getCompressLevel(len(a))
		dst = CompressZSTDLevel(dst, bb.B, compressLevel)
	}
	if len(bb.B) < minCompressibleBlockSize || float64(len(dst)-len(dstOrig)) > 0.9*float64(len(bb.B)) {
		// Ineffective compression. Store plain data.
		mt = MarshalTypeDeltaXor
		dst = append(dstOrig, bb.B...)
	}
	bbPool.Put(bb)

	//compressLevel := getCompressLevel(len(a))
	//dst = CompressZSTDLevel(dst, bb.B, compressLevel)
	return dst, mt, firstValue
}

func unmarshalInt64DeltaXorZSTD(dst []int64, src []byte, firstValue int64, itemsCount int) ([]int64, error) {
	valuesDecompressed, err := zstd.Decompress(nil, src)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress ZSTD delta from %d bytes; src=%X: %w", len(src), src, err)
	}
	dst, err = unmarshalInt64DeltaXor(nil, valuesDecompressed, firstValue, itemsCount)
	return dst, err
}
