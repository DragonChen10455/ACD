package encoding

import (
	"ACD/decimal"
	"ACD/gorillaz"
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

func marshalInt64Gorilla(dst []byte, a []int64, _ uint8) (result []byte, mt MarshalType, firstValue int64) {
	if len(a) == 0 {
		logger.Panicf("BUG: a must contain at least one item")
	}
	dst = gorillaz.Compress(dst, a)
	mt = MarshalTypeGorillaZ
	return dst, mt, a[0]
}

func unmarshalInt64Gorilla(dst []int64, src []byte, _ MarshalType, _ int64, itemsCount int) ([]int64, error) {
	dst = decimal.ExtendInt64sCapacity(dst, itemsCount)
	dst, err := gorillaz.Decompress(dst, src)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal value: %w", err)
	}
	return dst, err
}
