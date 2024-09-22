package encoding

type ContainerHeader struct {
	// segmentCount is the number of the segment inside block
	segmentCount uint64

	// segmentOffsetArray is the array of segment offset where the size of array is segmentCount
	segmentOffsetArray []uint64
}

func NewContainerHeader(segmentCount uint64) *ContainerHeader {
	return &ContainerHeader{
		segmentCount: segmentCount,
	}
}

func (th *ContainerHeader) GetSegmentCount() uint64 {
	return th.segmentCount
}

func (th *ContainerHeader) GetSegmentOffsetArray() []uint64 {
	return th.segmentOffsetArray
}

func (th *ContainerHeader) AddSegmentOffset(startOffset uint64) {
	th.segmentOffsetArray = append(th.segmentOffsetArray, startOffset)
}

// Marshal appends marshaled th to dst and returns the result.
func (th *ContainerHeader) Marshal(dst []byte) []byte {
	dst = MarshalUint64(dst, th.segmentCount)
	dst = MarshalVarUint64s(dst, th.segmentOffsetArray)
	return dst
}

// Unmarshal unmarshals th from src and returns the tail left after the unmarshaling.
func (th *ContainerHeader) Unmarshal(src []byte) error {

	th.segmentCount = UnmarshalUint64(src)
	dstByte := src[8:]
	for len(dstByte) > 0 {
		remainTail, dstValue, err := UnmarshalVarUint64(dstByte)
		if err != nil {
			return err
		}
		th.segmentOffsetArray = append(th.segmentOffsetArray, dstValue)
		dstByte = remainTail
	}
	return nil
}
