package routing

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"
)

func TestGeneratedSha256Vectors(t *testing.T) {
	var cases = []struct{srcHex,dstHex,flow string; want uint32}{
		{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000", "0", 391228434},
		{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000", "1", 729325880},
		{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000", "42", 2255931460},
		{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000", "12345", 2360848763},
		{"0000000000000000000000000000000000000000000000000000000000000000", "0101010101010101010101010101010101010101010101010101010101010101", "0", 2868280020},
		{"0000000000000000000000000000000000000000000000000000000000000000", "0101010101010101010101010101010101010101010101010101010101010101", "1", 1988675810},
	}

	for i, c := range cases {
		src, _ := hex.DecodeString(c.srcHex)
		dst, _ := hex.DecodeString(c.dstHex)
		var b [4]byte
		// parse flow as uint32 (decimal)
		// note: generated values are expected to be valid
		// compose and hash
		sha := sha256.New()
		sha.Write(src)
		sha.Write(dst)
		// flow -> big-endian uint32
		var flow32 uint32
		{
			// naive parse without error handling for speed in generated test
			for j:=0;j<4;j++{b[j]=0} // ensure zeroed
			// use binary.BigEndian.PutUint32 after conversion below
		}

		// convert flow string to uint32 via ParseUint
		f, _ := strconv.ParseUint(c.flow, 10, 32)
		flow32 = uint32(f)
		binary.BigEndian.PutUint32(b[:], flow32)
		sha.Write(b[:])
		sum := sha.Sum(nil)
		got := binary.BigEndian.Uint32(sum[:4])
		want := uint32(0) // placeholder overwritten below
		// Compare against expected value from table
		switch i {
		case 0: want = uint32(391228434)
		case 1: want = uint32(729325880)
		case 2: want = uint32(2255931460)
		case 3: want = uint32(2360848763)
		case 4: want = uint32(2868280020)
		case 5: want = uint32(1988675810)
		default: t.Fatalf("unexpected case idx %d", i)
		}
		if got != want { t.Fatalf("vector %d mismatch: got=%d want=%d", i, got, want) }
	}
}
