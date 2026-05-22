package crypto

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedAEADVectors(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "internal", "crypto", "generated_aead_vectors.csv"))
	if err != nil { t.Skipf("aead vectors not found: %v", err); return }
	defer f.Close()
	r := csv.NewReader(f)
	// header
	if _, err := r.Read(); err != nil { t.Fatalf("read header: %v", err) }
	for {
		rec, err := r.Read(); if err != nil { break }
		combined, _ := hex.DecodeString(rec[0])
		sInfo, _ := hex.DecodeString(rec[1])
		seq := rec[2]
		pt, _ := hex.DecodeString(rec[3])
		cipher, _ := hex.DecodeString(rec[4])
		sess, err := NewHybridSession(combined, sInfo)
		if err != nil { t.Fatalf("NewHybridSession: %v", err) }
		// decrypt in place: allocate buffer and copy cipher text
		buf := make([]byte, len(cipher))
		copy(buf, cipher)
		// parse seq decimal and use it
		var seqv uint64
		_, _ = fmt.Sscan(seq, &seqv)
		_, err = sess.DecryptInPlace(buf, seqv)
		if err != nil { t.Fatalf("decrypt failed: %v", err) }
		// compare plaintext prefix
		if len(buf) < len(pt) || string(buf[:len(pt)]) != string(pt) { t.Fatalf("plaintext mismatch") }
	}
}
