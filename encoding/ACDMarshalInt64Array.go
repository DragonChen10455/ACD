package encoding

import (
	"fmt"
	"github.com/steakknife/hamming"
	"math"
)

func marshalInt64ACD(dst []byte, a []int64, precisionBits uint8) (result []byte, mt MarshalType, firstValue int64) {
	segments := splitSegmentBits(a)
	// segment count
	segmentCount := len(segments)
	fmt.Printf("SegmentSA-by-bit2(19) SegmentCount:%v\n", segmentCount)
	extraHeader := 0
	// container
	// ———————————————————————————————————————————————————
	// containerHeader + segment0 + segment1 + ... + segmentN
	// ———————————————————————————————————————————————————

	// containerHeader
	// —————————————————————————————————————————————————————————————————————
	// segmentCount + segmentOffset0 + segmentOffset1 + ... + segmentOffsetN
	// —————————————————————————————————————————————————————————————————————
	containerH := NewContainerHeader(uint64(segmentCount))

	var segmentsByteArray []byte
	startOffset := 0
	//fmt.Printf("expected:\n")
	for i, v := range segments {
		// compress value to get resultValue
		resultValueByteArray, mt, firstValue := MarshalInt64sSelfAdaptive(nil, v, precisionBits)
		// segmentHeader
		// ———————————————————————————————————————————
		// firstValue + marshalType
		// ———————————————————————————————————————————
		segH := NewSegmentHeader2(firstValue, uint64(len(segments[i])), mt)
		// compress header to get segmentHeader
		segmentHeaderByteArray := segH.Marshal(nil)
		// 1.store segmentHeader into segmentsByteArray
		segmentsByteArray = append(segmentsByteArray, segmentHeaderByteArray...)
		// 2.store resultValue into segmentsByteArray
		segmentsByteArray = append(segmentsByteArray, resultValueByteArray...)
		// store offset into containerHeader's segmentOffsetArray
		containerH.AddSegmentOffset(uint64(startOffset))
		// [,)
		endOffset := startOffset + len(segmentHeaderByteArray) + len(resultValueByteArray)
		startOffset = endOffset
		extraHeader += len(segmentHeaderByteArray)
		//fmt.Printf("segment(%d) firstValue: %v, count : %v, marshalType: %v\n", i,
		//	segH.GetFirstValue(), segH.GetCount(), segH.GetMarshalType())
	}

	// store containerHeader to dst
	containerHeaderByteArray := containerH.Marshal(nil)
	dst = append(dst, byte(len(containerHeaderByteArray)))
	dst = append(dst, containerHeaderByteArray...)
	// store segment0 + segment1 + ... + segmentN to dst
	dst = append(dst, segmentsByteArray...)
	//fmt.Printf("containerH segmentCount:%v, segmentOffsetArray:%v\n",
	//	containerH.GetSegmentCount(), containerH.GetSegmentOffsetArray())
	mt = MarshalTypeACD
	//fmt.Printf("SegmentSA-by-bit2(19) Extra Space:%v\n", 1+extraHeader+len(containerHeaderByteArray))
	return dst, mt, firstValue
}

func unmarshalInt64ACD(dstValues []int64, src []byte, _ int64, _ int) ([]int64, error) {
	// decompress to get containerHeader byte length
	containerHeaderByteArrayLength := uint64(src[0])
	// decompress containerHeader from 2nd byte with length of containerHeaderByteArrayLength
	containerHeaderEndOffset := containerHeaderByteArrayLength + 1
	var containerH ContainerHeader
	err := containerH.Unmarshal(src[1:containerHeaderEndOffset])
	if err != nil {
		return nil, err
	}
	//fmt.Printf("containerH segmentCount:%v, segmentOffsetArray:%v\n",
	//	containerH.GetSegmentCount(), containerH.GetSegmentOffsetArray())
	// traverse segmentOffsetArray to decompress all the segments
	segmentOffsetArray := containerH.GetSegmentOffsetArray()
	for i := 0; i < int(containerH.GetSegmentCount()); i++ {
		startOffset := segmentOffsetArray[i]
		var segH SegmentHeader2
		segmentHeaderEndOffset := containerHeaderEndOffset + startOffset + uint64(segH.GetSegmentHeaderLength())
		segmentHeaderByteArray := src[containerHeaderEndOffset+startOffset : segmentHeaderEndOffset]
		_, err = segH.Unmarshal(segmentHeaderByteArray)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("firstValue: %v, count : %v, marshalType: %v\n",
		//	segH.GetFirstValue(), segH.GetCount(), segH.GetMarshalType())

		var endOffset uint64
		var resultValueByteArray []byte
		if i < int(containerH.GetSegmentCount())-1 {
			endOffset = containerHeaderEndOffset + segmentOffsetArray[i+1]
			resultValueByteArray = src[segmentHeaderEndOffset:endOffset]
			//fmt.Printf("resultValue.byteSize: %v resultValue[%v:%v]\n", endOffset-resultIndexEndOffset,
			//	resultIndexEndOffset-containerHeaderByteArrayLength-1, endOffset-containerHeaderByteArrayLength-1)
		} else {
			resultValueByteArray = src[segmentHeaderEndOffset:]
		}
		resultValue, errUnmarshalValues := UnmarshalInt64sSelfAdaptive(
			nil, resultValueByteArray, segH.GetMarshalType(), segH.GetFirstValue(), int(segH.GetCount()))
		if errUnmarshalValues != nil {
			return nil, errUnmarshalValues
		}
		dstValues = append(dstValues, resultValue...)
	}
	return dstValues, nil
}

func calNearestValueAttribute(lastValue int64, currentValue int64) (int, int, int, int, int) {
	xor := lastValue ^ currentValue
	hammingDistance := hamming.CountBitsInt64(xor)
	// count the number of zeros before the highest 1
	originNumberA := calBitNumber(lastValue)
	originNumberB := calBitNumber(currentValue)
	originMaxNumber := int(math.Max(float64(originNumberA), float64(originNumberB)))
	xorNumber := calBitNumber(xor)
	leadingZeros := int(math.Abs(float64(originMaxNumber - xorNumber)))
	// Count the number of zeros after the highest 1
	trailingZeros := xorNumber - hammingDistance
	return leadingZeros, trailingZeros, originMaxNumber, xorNumber, hammingDistance
}

// (frontlap，backendlap，delta，deltaOfDelta)
func calNearestValueAttributeParam(lastValue int64, currentValue int64, nextValue int64) (int, int, int64, int64) {
	delta := currentValue - lastValue
	delta2 := nextValue - currentValue
	deltaOfDelta := delta2 - delta
	xorValue := lastValue ^ currentValue
	//fmt.Printf("xor:%b\n", xorValue)
	originNumberOfLastValue := calBitNumber(lastValue)
	originNumberOfCurrentValue := calBitNumber(currentValue)
	originMaxNumber := int(math.Max(float64(originNumberOfLastValue), float64(originNumberOfCurrentValue)))
	xorNumber := calBitNumber(xorValue)
	leadingZeros := int(math.Abs(float64(originMaxNumber - xorNumber)))
	trailingZeros := calTrailingZeros(xorValue)
	return leadingZeros, trailingZeros, delta, deltaOfDelta
}

func calBitNumber(num int64) int {
	count := 0
	for num > 0 {
		if num != 0 {
			count++
		}
		num >>= 1
	}
	return count
}

func calTrailingZeros(value int64) int {
	count := 0
	for value&1 == 0 && value != 0 {
		count++
		value >>= 1
	}
	return count
}

func splitSegmentBits(array []int64) [][]int64 {
	var segmentArray [][]int64
	if len(array) < 3 {
		segmentArray = append(segmentArray, array)
		return segmentArray
	}
	lastValue := array[0]
	group0 := []int64{lastValue}
	g := 0
	segmentArray = append(segmentArray, group0)
	for i := 1; i < len(array)-1; i++ {
		currentValue := array[i]
		nextValue := array[i+1]
		leadingZeros1, trailingZeros1, originMaxNumber1, xorNumber1, hammingDistance1 := calNearestValueAttribute(lastValue, currentValue)
		leadingZeros2, trailingZeros2, originMaxNumber2, xorNumber2, hammingDistance2 := calNearestValueAttribute(lastValue, nextValue)
		if g < 90 && needSplit(leadingZeros1, trailingZeros1, originMaxNumber1, xorNumber1, hammingDistance1,
			leadingZeros2, trailingZeros2, originMaxNumber2, xorNumber2, hammingDistance2) { //新分段
			g++
			group := []int64{currentValue}
			segmentArray = append(segmentArray, group)
		} else {
			segmentArray[g] = append(segmentArray[g], currentValue)
		}
		lastValue = currentValue
	}
	segmentArray[g] = append(segmentArray[g], array[len(array)-1])
	return segmentArray
}

func needSplit(leadingZeros1 int, trailingZeros1 int, originMaxNumber1 int, xorNumber1 int, hammingDistance1 int,
	leadingZeros2 int, trailingZeros2 int, originMaxNumber2 int, xorNumber2 int, hammingDistance2 int) bool {
	return leadingZeros1 < originMaxNumber1/2 && trailingZeros1 < xorNumber1/2 && hammingDistance1 > 25 &&
		leadingZeros2 < originMaxNumber2/2 && trailingZeros2 < xorNumber2/2 && hammingDistance2 > 25
}
