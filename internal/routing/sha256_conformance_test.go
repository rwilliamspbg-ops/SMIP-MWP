package routing

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestPredictiveBe32Conformance(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "formal", "lean4", "SmipSha256Vectors_sample.csv"))
	if err != nil {
		t.Skipf("vectors not found: %v", err)
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	// read header
	if _, err := r.Read(); err != nil {
		t.Fatalf("read header: %v", err)
	}
	for i := 0; ; i++ {
		rec, err := r.Read()
		if err != nil {
			break
		}
		if len(rec) < 4 {
			t.Fatalf("unexpected record length: %v", rec)
		}
		srcHex := rec[0]
		dstHex := rec[1]
		flowStr := rec[2]
		be32Str := rec[3]

		src, err := hex.DecodeString(srcHex)
		if err != nil {
			t.Fatalf("decode src hex: %v", err)
		}
		dst, err := hex.DecodeString(dstHex)
		if err != nil {
			t.Fatalf("decode dst hex: %v", err)
		}
		flow64, err := strconv.ParseUint(flowStr, 10, 32)
		if err != nil {
			t.Fatalf("parse flow: %v", err)
		}
		wantBe32, err := strconv.ParseUint(be32Str, 10, 32)
		if err != nil {
			t.Fatalf("parse be32: %v", err)
		}

		// compute Go-style sha256(src||dst||flow_be32)
		h := sha256.New()
		h.Write(src)
		h.Write(dst)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(flow64))
		h.Write(b[:])
		sum := h.Sum(nil)
		got := binary.BigEndian.Uint32(sum[:4])

		if uint64(got) != wantBe32 {
			t.Fatalf("vector %d mismatch: got=%d want=%d", i, got, wantBe32)
		}
	}
}
