package bloom

import "fmt"

type Bloom struct {
	Bits []byte
	Hash int
}

func NewBloom(numKeys int) (*Bloom, error) {
	if numKeys <= 0 {
		return nil, fmt.Errorf("improrer size for the bloom")
	}

	return &Bloom{}, nil
}
