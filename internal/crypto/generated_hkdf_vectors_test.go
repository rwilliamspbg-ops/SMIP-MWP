package crypto

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedHKDFVectors(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "internal", "crypto", "generated_hkdf_vectors.csv"))
	if err != nil { t.Skipf("hkdf vectors not found: %v", err); return }
	defer f.Close()
	r := csv.NewReader(f)
	// header
	if _, err := r.Read(); err != nil { t.Fatalf("read header: %v", err) }
	for i:=0;;i++{
		rec, err := r.Read(); if err != nil { break }
		combined, _ := hex.DecodeString(rec[0])
		sInfo, _ := hex.DecodeString(rec[1])
		wantKey, _ := hex.DecodeString(rec[2])
		wantNonce, _ := hex.DecodeString(rec[3])
		wantMaskStr := rec[4]
		key, nonce, seqMask, err := ExportDerivedSessionMaterial(combined, sInfo)
		if err != nil { t.Fatalf("derive failed: %v", err) }
		if !equalBytes(key, wantKey) { t.Fatalf("key mismatch") }
		if !equalBytes(nonce, wantNonce) { t.Fatalf("nonce mismatch") }
		var wantMask uint64
		_, _ = fmt.Sscan(wantMaskStr, &wantMask)
		if seqMask != wantMask { t.Fatalf("seqMask mismatch: got=%d want=%d", seqMask, wantMask) }
	}
}

func equalBytes(a,b []byte) bool{ if len(a)!=len(b){return false}; for i:=range a{ if a[i]!=b[i]{return false} }; return true }
