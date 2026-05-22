package main

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"smip-mwp/internal/crypto"
)

func main() {
	root := filepath.Join("formal", "lean4")
	pattern := "SmipSha256Vectors*.csv"
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob: %v\n", err)
		os.Exit(2)
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no vector CSVs found under %s\n", root)
		os.Exit(0)
	}

	// We'll collect rows and generate a Go test file
	type row struct{ src, dst, flow, be32 string }
	var rows []row

	for _, path := range matches {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open %s: %v\n", path, err)
			os.Exit(2)
		}
		r := csv.NewReader(f)
		// header
		if _, err := r.Read(); err != nil {
			_ = f.Close()
			fmt.Fprintf(os.Stderr, "read header %s: %v\n", path, err)
			os.Exit(2)
		}
		for {
			rec, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = f.Close()
				fmt.Fprintf(os.Stderr, "read csv %s: %v\n", path, err)
				os.Exit(2)
			}
			if len(rec) < 4 {
				continue
			}
			rows = append(rows, row{rec[0], rec[1], rec[2], rec[3]})
		}
		_ = f.Close()
	}

	outDir := filepath.Join("internal", "routing")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(2)
	}
	outPath := filepath.Join(outDir, "generated_sha256_vectors_test.go")
	w, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(2)
	}
	defer w.Close()

	w.WriteString("package routing\n\n")
	w.WriteString("import (\n")
	w.WriteString("\t\"crypto/sha256\"\n")
	w.WriteString("\t\"encoding/binary\"\n")
	w.WriteString("\t\"encoding/hex\"\n")
	w.WriteString("\t\"strconv\"\n")
	w.WriteString("\t\"testing\"\n")
	w.WriteString(")\n\n")
	w.WriteString("func TestGeneratedSha256Vectors(t *testing.T) {\n")
	w.WriteString("\tvar cases = []struct{srcHex,dstHex,flow string; want uint32}{\n")
	for _, r := range rows {
		// sanitize inputs
		src := strings.ToLower(strings.TrimSpace(r.src))
		dst := strings.ToLower(strings.TrimSpace(r.dst))
		flow := strings.TrimSpace(r.flow)
		want := strings.TrimSpace(r.be32)
		fmt.Fprintf(w, "\t\t{\"%s\", \"%s\", \"%s\", %s},\n", src, dst, flow, want)
	}
	w.WriteString("\t}\n\n")
	w.WriteString("\tfor i, c := range cases {\n")
	w.WriteString("\t\tsrc, _ := hex.DecodeString(c.srcHex)\n")
	w.WriteString("\t\tdst, _ := hex.DecodeString(c.dstHex)\n")
	w.WriteString("\t\tvar b [4]byte\n")
	w.WriteString("\t\t// parse flow as uint32 (decimal)\n")
	w.WriteString("\t\t// note: generated values are expected to be valid\n")
	w.WriteString("\t\t// compose and hash\n")
	w.WriteString("\t\tsha := sha256.New()\n")
	w.WriteString("\t\tsha.Write(src)\n")
	w.WriteString("\t\tsha.Write(dst)\n")
	w.WriteString("\t\t// flow -> big-endian uint32\n")
	w.WriteString("\t\tvar flow32 uint32\n")
	w.WriteString("\t\t{\n")
	w.WriteString("\t\t\t// naive parse without error handling for speed in generated test\n")
	w.WriteString("\t\t\tfor j:=0;j<4;j++{b[j]=0} // ensure zeroed\n")
	w.WriteString("\t\t\t// use binary.BigEndian.PutUint32 after conversion below\n")
	w.WriteString("\t\t}\n\n")
	w.WriteString("\t\t// convert flow string to uint32 via ParseUint\n")
	w.WriteString("\t\tf, _ := strconv.ParseUint(c.flow, 10, 32)\n")
	w.WriteString("\t\tflow32 = uint32(f)\n")
	w.WriteString("\t\tbinary.BigEndian.PutUint32(b[:], flow32)\n")
	w.WriteString("\t\tsha.Write(b[:])\n")
	w.WriteString("\t\tsum := sha.Sum(nil)\n")
	w.WriteString("\t\tgot := binary.BigEndian.Uint32(sum[:4])\n")
	w.WriteString("\t\twant := uint32(0) // placeholder overwritten below\n")
	w.WriteString("\t\t// Compare against expected value from table\n")
	w.WriteString("\t\tswitch i {\n")
	for i := range rows {
		fmt.Fprintf(w, "\t\tcase %d: want = uint32(%s)\n", i, rows[i].be32)
	}
	w.WriteString("\t\tdefault: t.Fatalf(\"unexpected case idx %d\", i)\n")
	w.WriteString("\t\t}\n")
	w.WriteString("\t\tif got != want { t.Fatalf(\"vector %d mismatch: got=%d want=%d\", i, got, want) }\n")
	w.WriteString("\t}\n")
	w.WriteString("}\n")

	// Ensure file is formatted enough; users can `gofmt` later.
	fmt.Printf("wrote %s\n", outPath)

	// Always derive HKDF vectors using the Go implementation so generated
	// test vectors reflect the production HKDF behavior used by the code.
	hkdfOutPath := filepath.Join("internal", "crypto", "generated_hkdf_vectors.csv")
	hf, err := os.Create(hkdfOutPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create hkdf csv: %v\n", err)
		os.Exit(2)
	}
	cw := csv.NewWriter(hf)
	_ = cw.Write([]string{"combined_hex", "sessionInfo_hex", "key_hex", "nonce_base_hex", "seqMask"})

	samples := []struct{ combined, session string }{
		{strings.Repeat("00", 32), "session-0"},
		{strings.Repeat("01", 32), "session-1"},
	}
	for _, s := range samples {
		combined, _ := hex.DecodeString(s.combined)
		key, nonceBase, seqMask, err := crypto.ExportDerivedSessionMaterial(combined, []byte(s.session))
		if err != nil {
			fmt.Fprintf(os.Stderr, "derive hkdf: %v\n", err)
			os.Exit(2)
		}
		_ = cw.Write([]string{strings.ToLower(s.combined), fmt.Sprintf("%x", []byte(s.session)), fmt.Sprintf("%x", key), fmt.Sprintf("%x", nonceBase), fmt.Sprintf("%d", seqMask)})
	}
	cw.Flush()
	_ = hf.Close()
	fmt.Printf("wrote %s\n", hkdfOutPath)

	// Generate AEAD vectors by encrypting a small plaintext under derived sessions.
	aeadOut := filepath.Join("internal", "crypto", "generated_aead_vectors.csv")
	af, err := os.Create(aeadOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create aead csv: %v\n", err)
		os.Exit(2)
	}
	aw := csv.NewWriter(af)
	_ = aw.Write([]string{"combined_hex", "sessionInfo_hex", "seq", "plaintext_hex", "ciphertext_hex"})

	// read the hkdf csv we will base AEAD vectors on
	hf2, err := os.Open(hkdfOutPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open hkdf csv: %v\n", err)
		os.Exit(2)
	}
	defer hf2.Close()
	hr := csv.NewReader(hf2)
	// skip header
	if _, err := hr.Read(); err != nil {
		fmt.Fprintf(os.Stderr, "read hkdf header: %v\n", err)
		os.Exit(2)
	}
	plaintext := []byte("hello-from-generator")
	for {
		rec, err := hr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "read hkdf csv: %v\n", err)
			os.Exit(2)
		}
		combinedHex := rec[0]
		sInfoHex := rec[1]
		combined, _ := hex.DecodeString(combinedHex)
		sInfo, _ := hex.DecodeString(sInfoHex)
		// create a session and encrypt with seq=0
		sess, err := crypto.NewHybridSession(combined, sInfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "NewHybridSession: %v\n", err)
			os.Exit(2)
		}
		// ensure dst has capacity
		dst := make([]byte, len(plaintext)+crypto.TagSize)
		copy(dst, plaintext)
		// use EncryptInPlace semantics
		if err := sess.EncryptInPlace(dst[:len(plaintext)], 0); err != nil {
			fmt.Fprintf(os.Stderr, "EncryptInPlace: %v\n", err)
			os.Exit(2)
		}
		cipherHex := fmt.Sprintf("%x", dst[:len(plaintext)+crypto.TagSize])
		_ = aw.Write([]string{combinedHex, sInfoHex, "0", fmt.Sprintf("%x", plaintext), cipherHex})
	}
	aw.Flush()
	_ = af.Close()
	fmt.Printf("wrote %s\n", aeadOut)

	// Emit hkdf test that validates the CSV content matches ExportDerivedSessionMaterial
	hkdfTestOut := filepath.Join("internal", "crypto", "generated_hkdf_vectors_test.go")
	g, err := os.Create(hkdfTestOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create hkdf test: %v\n", err)
		os.Exit(2)
	}
	defer g.Close()
	g.WriteString("package crypto\n\n")
	g.WriteString("import (\n\t\"encoding/csv\"\n\t\"encoding/hex\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"testing\"\n)\n\n")
	g.WriteString("func TestGeneratedHKDFVectors(t *testing.T) {\n")
	g.WriteString("\tf, err := os.Open(filepath.Join(\"..\", \"..\", \"internal\", \"crypto\", \"generated_hkdf_vectors.csv\"))\n")
	g.WriteString("\tif err != nil { t.Skipf(\"hkdf vectors not found: %v\", err); return }\n")
	g.WriteString("\tdefer f.Close()\n")
	g.WriteString("\tr := csv.NewReader(f)\n")
	g.WriteString("\t// header\n\tif _, err := r.Read(); err != nil { t.Fatalf(\"read header: %v\", err) }\n")
	g.WriteString("\tfor i:=0;;i++{\n\t\trec, err := r.Read(); if err != nil { break }\n")
	g.WriteString("\t\tcombined, _ := hex.DecodeString(rec[0])\n")
	g.WriteString("\t\tsInfo, _ := hex.DecodeString(rec[1])\n")
	g.WriteString("\t\twantKey, _ := hex.DecodeString(rec[2])\n")
	g.WriteString("\t\twantNonce, _ := hex.DecodeString(rec[3])\n")
	g.WriteString("\t\twantMaskStr := rec[4]\n")
	g.WriteString("\t\tkey, nonce, seqMask, err := ExportDerivedSessionMaterial(combined, sInfo)\n")
	g.WriteString("\t\tif err != nil { t.Fatalf(\"derive failed: %v\", err) }\n")
	g.WriteString("\t\tif !equalBytes(key, wantKey) { t.Fatalf(\"key mismatch\") }\n")
	g.WriteString("\t\tif !equalBytes(nonce, wantNonce) { t.Fatalf(\"nonce mismatch\") }\n")
	g.WriteString("\t\tvar wantMask uint64\n")
	g.WriteString("\t\t_, _ = fmt.Sscan(wantMaskStr, &wantMask)\n")
	g.WriteString("\t\tif seqMask != wantMask { t.Fatalf(\"seqMask mismatch: got=%d want=%d\", seqMask, wantMask) }\n")
	g.WriteString("\t}\n")
	g.WriteString("}\n\n")
	g.WriteString("func equalBytes(a,b []byte) bool{ if len(a)!=len(b){return false}; for i:=range a{ if a[i]!=b[i]{return false} }; return true }\n")
	fmt.Printf("wrote %s\n", hkdfTestOut)

	// Emit AEAD test that validates decryption of generated ciphertexts
	aeadTestOut := filepath.Join("internal", "crypto", "generated_aead_vectors_test.go")
	h, err := os.Create(aeadTestOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create aead test: %v\n", err)
		os.Exit(2)
	}
	defer h.Close()
	h.WriteString("package crypto\n\n")
	h.WriteString("import (\n\t\"encoding/csv\"\n\t\"encoding/hex\"\n\t\"fmt\"\n\t\"os\"\n\t\"path/filepath\"\n\t\"testing\"\n)\n\n")
	h.WriteString("func TestGeneratedAEADVectors(t *testing.T) {\n")
	h.WriteString("\tf, err := os.Open(filepath.Join(\"..\", \"..\", \"internal\", \"crypto\", \"generated_aead_vectors.csv\"))\n")
	h.WriteString("\tif err != nil { t.Skipf(\"aead vectors not found: %v\", err); return }\n")
	h.WriteString("\tdefer f.Close()\n")
	h.WriteString("\tr := csv.NewReader(f)\n")
	h.WriteString("\t// header\n\tif _, err := r.Read(); err != nil { t.Fatalf(\"read header: %v\", err) }\n")
	h.WriteString("\tfor {\n\t\trec, err := r.Read(); if err != nil { break }\n")
	h.WriteString("\t\tcombined, _ := hex.DecodeString(rec[0])\n")
	h.WriteString("\t\tsInfo, _ := hex.DecodeString(rec[1])\n")
	h.WriteString("\t\tseq := rec[2]\n")
	h.WriteString("\t\tpt, _ := hex.DecodeString(rec[3])\n")
	h.WriteString("\t\tcipher, _ := hex.DecodeString(rec[4])\n")
	h.WriteString("\t\tsess, err := NewHybridSession(combined, sInfo)\n")
	h.WriteString("\t\tif err != nil { t.Fatalf(\"NewHybridSession: %v\", err) }\n")
	h.WriteString("\t\t// decrypt in place: allocate buffer and copy cipher text\n")
	h.WriteString("\t\tbuf := make([]byte, len(cipher))\n")
	h.WriteString("\t\tcopy(buf, cipher)\n")
	h.WriteString("\t\t// parse seq decimal and use it\n")
	h.WriteString("\t\tvar seqv uint64\n")
	h.WriteString("\t\t_, _ = fmt.Sscan(seq, &seqv)\n")
	h.WriteString("\t\t_, err = sess.DecryptInPlace(buf, seqv)\n")
	h.WriteString("\t\tif err != nil { t.Fatalf(\"decrypt failed: %v\", err) }\n")
	h.WriteString("\t\t// compare plaintext prefix\n")
	h.WriteString("\t\tif len(buf) < len(pt) || string(buf[:len(pt)]) != string(pt) { t.Fatalf(\"plaintext mismatch\") }\n")
	h.WriteString("\t}\n")
	h.WriteString("}\n")
	fmt.Printf("wrote %s\n", aeadTestOut)
}
