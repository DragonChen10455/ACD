package encoding

import (
	"fmt"
)

type SegmentHeader2 struct {
	// firstValue is the first value marshaled from values
	firstValue int64

	// count is the number of this segment
	count uint64

	// marshalType is the type used for encoding the timestamps block
	marshalType MarshalType
}

func (th *SegmentHeader2) GetSegmentHeaderLength() int {
	return 17
}

func NewSegmentHeader2(firstValue int64, count uint64, marshalType MarshalType) *SegmentHeader2 {
	return &SegmentHeader2{
		firstValue:  firstValue,
		count:       count,
		marshalType: marshalType,
	}
}

func (th *SegmentHeader2) GetFirstValue() int64 {
	return th.firstValue
}

func (th *SegmentHeader2) GetCount() uint64 {
	return th.count
}

func (th *SegmentHeader2) GetMarshalType() MarshalType {
	return th.marshalType
}

func (th *SegmentHeader2) Reset() {
	th.firstValue = 0
	th.count = 0
	th.marshalType = 0
}

// Marshal appends marshaled th to dst and returns the result.
func (th *SegmentHeader2) Marshal(dst []byte) []byte {
	dst = MarshalInt64(dst, th.firstValue)
	dst = MarshalUint64(dst, th.count)
	dst = append(dst, byte(th.marshalType))
	return dst
}

// Unmarshal unmarshals th from src and returns the tail left after the unmarshaling.
func (th *SegmentHeader2) Unmarshal(src []byte) ([]byte, error) {
	th.Reset()

	if len(src) < th.GetSegmentHeaderLength() {
		return src, fmt.Errorf("cannot unmarshal segmentHeader from %d bytes; need at least 17 bytes", len(src))
	}

	th.firstValue = UnmarshalInt64(src)
	th.count = UnmarshalUint64(src[8:])
	th.marshalType = MarshalType(src[16])

	return src[9:], nil
}
