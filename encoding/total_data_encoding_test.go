package encoding

import (
	"ACD/decimal"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func checkPrecisionBits(a, b []int64, precisionBits uint8) error {
	if len(a) != len(b) {
		return fmt.Errorf("different-sized arrays: %d vs %d", len(a), len(b))
	}
	for i, av := range a {
		bv := b[i]
		if av < bv {
			av, bv = bv, av
		}
		eps := av - bv
		if eps == 0 {
			continue
		}
		if av < 0 {
			av = -av
		}
		pbe := uint8(1)
		for eps < av {
			av >>= 1
			pbe++
		}
		if pbe < precisionBits {
			return fmt.Errorf("too low precisionBits for\na=%d\nb=%d\ngot %d; expecting %d; compared values: %d vs %d, eps=%d",
				a, b, pbe, precisionBits, a[i], b[i], eps)
		}
	}
	return nil
}

func TestCompressDecompressValues(t *testing.T) {

	//var dataFileNamePrefix = "point_plot_"
	var values1, values2, values3 []int64
	var tmpValues2, tmpValues3 []float64

	var index []int64
	var values1V1 []int64
	var values1V2 []int64
	var values2U1 []float64
	var values2U2 []float64
	var values3W1 []float64
	var values3W2 []float64

	v1 := int64(0)
	v2 := int64(0)
	for i := 0; i < 8*1024; i++ {
		v1 += int64(rand.NormFloat64() * 1e2)
		v2 += int64(rand.NormFloat64() * 1e2)
		index = append(index, int64(i))
		values1V1 = append(values1V1, v1)
		values1V2 = append(values1V2, v2)
		values1 = append(values1, ZOrderEncodeFloat754(float32(v1), float32(v2), 32))
	}

	//drawInt64(values1V1, values1V2, dataFileNamePrefix+"int64_asc")
	//drawInt64(index, values1, dataFileNamePrefix+"int64_asc_z2")
	//writeToFileInt64(values1V1, values1V2, values1, dataFileNamePrefix+"int64_asc.txt")

	u1 := float64(0)
	u2 := float64(0)
	for i := 0; i < 8*1024; i++ {
		u1 += rand.NormFloat64() * 1e2
		u2 += rand.NormFloat64() * 1e2
		values2U1 = append(values2U1, u1)
		values2U2 = append(values2U2, u2)
		tmpValues2 = append(tmpValues2, float64(ZOrderEncodeFloat754(float32(u1), float32(u2), 32)))
	}
	values2, _ = decimal.AppendFloatToInt64(values2, tmpValues2)

	//drawFloat64(values2U1, values2U2, dataFileNamePrefix+"float64_asc")
	//drawInt64(index, values2, dataFileNamePrefix+"float64_asc_z2")
	//writeToFileFloat64(values2U1, values2U2, values2, dataFileNamePrefix+"float64_asc.txt")

	for i := 0; i < 8*1024; i++ {
		w1 := rand.NormFloat64() * 1e2
		w2 := rand.NormFloat64() * 1e2
		values3W1 = append(values3W1, w1)
		values3W2 = append(values3W2, w2)
		tmpValues3 = append(tmpValues3, float64(ZOrderEncodeFloat754(float32(w1), float32(w2), 32)))
	}
	values3, _ = decimal.AppendFloatToInt64(values3, tmpValues3)

	//drawFloat64(values3W1, values3W2, dataFileNamePrefix+"float64_norm")
	//drawInt64(index, values3, dataFileNamePrefix+"float64_norm_z2")
	//writeToFileFloat64(values3W1, values3W2, values3, dataFileNamePrefix+"float64_norm.txt")

	testCompressDecompressValuesOld(t, values1, values1V1, values1V2)
}

func TestTSBSBenchmark(t *testing.T) {
	var values []int64
	var xValuesFloat64 []float64
	var yValuesFloat64 []float64
	fileName := "../data/data"
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("无法打开文件:", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineText := scanner.Text()
		if !strings.Contains(lineText, "latitude") || !strings.Contains(lineText, "longitude") {
			continue
		}
		record := strings.Split(lineText, ",")
		ty := record[0]
		if ty == "readings" {
			x, err := strconv.ParseFloat(record[8][9:], 64)
			y, err := strconv.ParseFloat(record[9][10:], 64)
			if err != nil {
				fmt.Printf("parse data error: %v\n", err)
				continue
			}
			xValuesFloat64 = append(xValuesFloat64, x)
			yValuesFloat64 = append(yValuesFloat64, y)
			values = append(values, IEEEEncodeFloat(float32(x), float32(y), 32))
		}
	}
	fmt.Printf("reading size: %v\n", len(values))
	xValues, _ := decimal.AppendFloatToInt64(nil, xValuesFloat64)
	yValues, _ := decimal.AppendFloatToInt64(nil, yValuesFloat64)

	xValues = xValues[:len(xValues)/4]
	yValues = yValues[:len(yValues)/4]
	values = values[:len(values)/4]

	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	fmt.Printf("source xValues size:%v\n", sourceXValuesSize)
	fmt.Printf("source yValues size:%v\n", sourceYValuesSize)

	testCompressDecompressValuesOld(t, values, xValues, yValues)
	testCompressDecompressValuesDeltaLZ4(t, values, xValues, yValues)
	testCompressDecompressValuesDelta2LZ4(t, values, xValues, yValues)
	testCompressDecompressValuesDeltaSnappy(t, values, xValues, yValues)
	testCompressDecompressValuesDelta2Snappy(t, values, xValues, yValues)
	testCompressDecompressValuesGorilla(t, values, xValues, yValues)
	testCompressDecompressValuesChimp(t, values, xValues, yValues)
	testCompressDecompressValueACD(t, values, xValues, yValues)
}

func testCompressDecompressValuesOld(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Old-xValues
	start1 := time.Now()
	result, mt, firstValue := marshalInt64Array(nil, xValues, 64)
	end1 := time.Now()
	start2 := time.Now()
	values2, err := unmarshalInt64Array(nil, result, mt, firstValue, len(xValues))
	end2 := time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(xValues, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	if len(xValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(xValues); i++ {
		if xValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}
	fmt.Printf("OLD-xValues(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Old-yValues
	start1 = time.Now()
	result, mt, firstValue = marshalInt64Array(nil, yValues, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(yValues))
	end2 = time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(yValues, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	if len(yValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(yValues); i++ {
		if yValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}

	fmt.Printf("OLD-yValues(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// Old-values
	start1 = time.Now()
	result, mt, firstValue = marshalInt64Array(nil, values, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(values))
	end2 = time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(values, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	if len(values) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(values); i++ {
		if values[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}

	fmt.Printf("OLD(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesDeltaLZ4(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// DeltaLZ4-xValues
	start1 := time.Now()
	result, _, _ := marshalInt64DeltaLZ4(nil, xValues, 64)
	end1 := time.Now()
	fmt.Printf("DeltaLZ4-xValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// DeltaLZ4-yValues
	start1 = time.Now()
	result, _, _ = marshalInt64DeltaLZ4(nil, yValues, 64)
	end1 = time.Now()
	fmt.Printf("DeltaLZ4-yValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// DeltaLZ4-values
	start1 = time.Now()
	result, _, _ = marshalInt64DeltaLZ4(nil, values, 64)
	end1 = time.Now()
	fmt.Printf("DeltaLZ4 compressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesDelta2LZ4(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Delta2LZ4-xValues
	start1 := time.Now()
	result, _, _ := marshalInt64Delta2LZ4(nil, xValues, 64)
	end1 := time.Now()
	fmt.Printf("Delta2LZ4-xValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Delta2LZ4-yValues
	start1 = time.Now()
	result, _, _ = marshalInt64Delta2LZ4(nil, yValues, 64)
	end1 = time.Now()
	fmt.Printf("Delta2LZ4-yValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// Delta2LZ4-values
	start1 = time.Now()
	result, _, _ = marshalInt64Delta2LZ4(nil, values, 64)
	end1 = time.Now()
	fmt.Printf("Delta2LZ4 compressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesDeltaSnappy(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// DeltaSnappy-xValues
	start1 := time.Now()
	result, _, _ := marshalInt64DeltaSnappy(nil, xValues, 64)
	end1 := time.Now()
	fmt.Printf("DeltaSnappy-xValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// DeltaSnappy-yValues
	start1 = time.Now()
	result, _, _ = marshalInt64DeltaSnappy(nil, yValues, 64)
	end1 = time.Now()
	fmt.Printf("DeltaSnappy-yValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// DeltaSnappy-values
	start1 = time.Now()
	result, _, _ = marshalInt64DeltaSnappy(nil, values, 64)
	end1 = time.Now()
	fmt.Printf("DeltaSnappy compressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesDelta2Snappy(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Delta2Snappy-xValues
	start1 := time.Now()
	result, _, _ := marshalInt64Delta2Snappy(nil, xValues, 64)
	end1 := time.Now()
	fmt.Printf("Delta2Snappy-xValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Delta2Snappy-yValues
	start1 = time.Now()
	result, _, _ = marshalInt64Delta2Snappy(nil, yValues, 64)
	end1 = time.Now()
	fmt.Printf("Delta2Snappy-yValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// Delta2Snappy-values
	start1 = time.Now()
	result, _, _ = marshalInt64Delta2Snappy(nil, values, 64)
	end1 = time.Now()
	fmt.Printf("Delta2Snappy compressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesGorilla(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Gorilla-xValues
	start1 := time.Now()
	result, mt, firstValue := marshalInt64Gorilla(nil, xValues, 64)
	end1 := time.Now()
	start2 := time.Now()
	values2, err := unmarshalInt64Gorilla(nil, result, mt, firstValue, len(values))
	end2 := time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}

	if len(xValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(xValues); i++ {
		if xValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}
	fmt.Printf("Gorilla-xValues compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Gorilla-yValues
	start1 = time.Now()
	result, mt, firstValue = marshalInt64Gorilla(nil, yValues, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Gorilla(nil, result, mt, firstValue, len(values))
	end2 = time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}

	if len(yValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(yValues); i++ {
		if yValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}

	fmt.Printf("Gorilla-yValues compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// Gorilla-values
	start1 = time.Now()
	result, mt, firstValue = marshalInt64Gorilla(nil, values, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Gorilla(nil, result, mt, firstValue, len(values))
	end2 = time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}

	if len(values) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(values); i++ {
		if values[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}
	fmt.Printf("Gorilla compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValuesChimp(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Chimp-xValues
	start1 := time.Now()
	result, _, _ := marshalInt64Chimp(nil, xValues, 64)
	end1 := time.Now()
	fmt.Printf("Chimp-xValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Chimp-yValues
	start1 = time.Now()
	result, _, _ = marshalInt64Chimp(nil, yValues, 64)
	end1 = time.Now()
	fmt.Printf("Chimp-yValues compressed:%d compressed size:%d ratio:%.8f\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// Chimp-values
	start1 = time.Now()
	result, _, _ = marshalInt64Chimp(nil, values, 64)
	end1 = time.Now()
	fmt.Printf("Chimp compressed:%d compressed size:%d ratio:%.8f\n\n",
		end1.UnixNano()-start1.UnixNano(),
		len(result),
		float64(sourceXYValuesSize)/float64(len(result)))
}

func testCompressDecompressValueACD(t *testing.T, values []int64, xValues []int64, yValues []int64) {
	sourceXValuesSize := 8 * len(xValues)
	sourceYValuesSize := 8 * len(yValues)
	sourceXYValuesSize := sourceXValuesSize + sourceYValuesSize

	// Old-xValues
	start1 := time.Now()
	result, mt, firstValue := MarshalInt64sSelfAdaptive(nil, xValues, 64)
	end1 := time.Now()
	start2 := time.Now()
	values2, err := UnmarshalInt64sSelfAdaptive(nil, result, mt, firstValue, len(xValues))
	end2 := time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(xValues, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	if len(xValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(xValues); i++ {
		if xValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}
	fmt.Printf("OLD-xValues(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(xValues))/float64(len(result)))

	// Old-yValues
	start1 = time.Now()
	result, mt, firstValue = marshalInt64Array(nil, yValues, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(yValues))
	end2 = time.Now()
	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if err := checkPrecisionBits(yValues, values2, 64); err != nil {
		t.Fatalf("too low precision for values: %s", err)
	}

	if len(yValues) != len(values2) {
		t.Fatalf("unmarshal length does not match\n")
	}
	for i := 0; i < len(yValues); i++ {
		if yValues[i] != values2[i] {
			t.Fatalf("unmarshal items does not match, values want: %d, but values2 got %d\n",
				values[i], values2[i])
		}
	}

	fmt.Printf("OLD-yValues(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result),
		float64(8*len(yValues))/float64(len(result)))

	// SegmentSA by bit2(19)
	start1 = time.Now()
	result, mt, firstValue = marshalInt64ACD(nil, values, 64)
	end1 = time.Now()
	start2 = time.Now()
	values2, err = unmarshalInt64Array(nil, result, mt, firstValue, len(values))
	end2 = time.Now()

	if err != nil {
		t.Fatalf("cannot unmarshal values: %s", err)
	}
	if len(values) != len(values2) {
		t.Fatalf("unmarshal length does not match len(values):%v, len(values2)%v\n", len(values), len(values2))
	}
	for i := 0; i < len(values); i++ {
		if values[i] != values2[i] {
			t.Fatalf("unmarshal items does not match,index %v, values want: %d, but values2 got %d\n",
				i, values[i], values2[i])
		}
	}
	fmt.Printf("SegmentSA-by-bit2(%d) compressed:%d decompressed:%d compressed size:%d ratio:%.8f\n\n", mt,
		end1.UnixNano()-start1.UnixNano(),
		end2.UnixNano()-start2.UnixNano(),
		len(result), float64(sourceXYValuesSize)/float64(len(result)))
}
