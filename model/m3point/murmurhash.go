package m3point

const (
	low  = 0x00000000ffffffff
	high = 0xffffffff00000000
	c1   = 0xcc9e2d51
	c2   = 0x1b873593
	r1a  = 15
	r1b  = 17
	r2a  = 13
	r2b  = 19
	m    = 4
	n    = 0xe6546b64
)

type Int32Id int32

type Int64Id int64

func (id Int32Id) MurmurHash() uint32 {
	return MurmurHashUint32([]uint32{uint32(id)})
}

func (id Int64Id) MurmurHash() uint32 {
	c64 := uint64(id)
	return MurmurHashUint32([]uint32{uint32(c64 & low), uint32(c64 & high >> 32)})
}

func murmurHashToInt(m uint32, size int) int {
	res := int(m) % size
	if res < 0 {
		return -res
	}
	return res
}

func MurmurHashUint32(data []uint32) uint32 {
	// Using Murmur 3 implementation
	// Found after research from https://softwareengineering.stackexchange.com/questions/49550/which-hashing-algorithm-is-best-for-uniqueness-and-speed/145633#145633?newreg=fcc6e22e2d1647e29d38f8d710248230
	h1 := uint32(0)
	for _, k1 := range data {
		k1 *= c1
		k1 = (k1 << r1a) | (k1 >> r1b)
		k1 *= c2
		h1 ^= k1
		h1 = (h1 << r2a) | (h1 >> r2b)
		h1 = h1*m + h1 + n
	}
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}

