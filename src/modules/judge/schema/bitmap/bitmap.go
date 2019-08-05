package bitmap

type BitMap struct {
	size int
	bits []bool
}

func NewBitMap(size int, init ...bool) *BitMap {
	bits := make([]bool, size)
	if len(init) > 0 && init[0] {
		for i := 0; i < size; i++ {
			bits[i] = true
		}
	}

	return &BitMap{
		size: size,
		bits: bits,
	}
}

func (b *BitMap) Set(pos ...int) {
	if len(pos) == 0 {
		return
	}
	for i := range pos {
		if pos[i] < 0 {
			return
		}
		if pos[i] >= b.size {
			return
		}
		b.bits[pos[i]] = true
	}
}

// [start, end), 左包含 右不包含
func (b *BitMap) SetRange(start, end int) {
	if end < start {
		return
	}
	if start < 0 {
		return
	}

	for i := start; i < b.size && i < end; i++ {
		b.bits[i] = true
	}
}

func (b *BitMap) IsSet(pos int) bool {
	if pos < 0 {
		return false
	}
	if pos >= b.size {
		return false
	}
	return b.bits[pos]
}
