package encoding

import (
	"ACD/chimp"
	"ACD/decimal"
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

func marshalInt64Chimp(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	if len(a) == 0 {
		logger.Panicf("BUG: a must contain at least one item")
	}
	dst = chimp.Compress(dst, a)
	return dst, 0, a[0]
}

func unmarshalInt64Chimp(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	dst = decimal.ExtendInt64sCapacity(dst, itemsCount)
	dst, err := chimp.Decompress(dst, src)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal value: %w", err)
	}
	return dst, err
}
