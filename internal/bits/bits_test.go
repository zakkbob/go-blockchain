package bits_test

import (
	"encoding/binary"
	"math/bits"
	"testing"

	internalbits "github.com/zakkbob/go-blockchain/internal/bits"
)

func TestLeadingZerosByte(t *testing.T) {
	for i := range 256 {
		expected := bits.LeadingZeros8(uint8(i))
		got := internalbits.LeadingZerosByte(byte(i))

		if expected != got {
			t.Errorf("got %d, expected %d for %d", got, expected, i)
		}
	}
}

func TestLeadingZerosBytes(t *testing.T) {
	for i := range 256 * 256 * 256 {
		expected := bits.LeadingZeros32(uint32(i))
		bytes := make([]byte, 0)
		bytes = binary.BigEndian.AppendUint32(bytes, uint32(i))
		got := internalbits.LeadingZerosBytes(bytes)

		if expected != got {
			t.Errorf("got %d, expected %d for %d", got, expected, i)
		}
	}
}
