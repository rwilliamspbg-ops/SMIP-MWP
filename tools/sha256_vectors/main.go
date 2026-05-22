package main

import (
    "crypto/sha256"
    "encoding/binary"
    "encoding/hex"
    "fmt"
)

func makeID(fill byte) [32]byte {
    var id [32]byte
    for i := 0; i < 32; i++ {
        id[i] = fill
    }
    return id
}

func be32FromHash(src, dst [32]byte, flow uint32) uint32 {
    h := sha256.New()
    h.Write(src[:])
    h.Write(dst[:])
    var b [4]byte
    binary.BigEndian.PutUint32(b[:], flow)
    h.Write(b[:])
    sum := h.Sum(nil)
    return binary.BigEndian.Uint32(sum[:4])
}

func hexID(id [32]byte) string {
    return hex.EncodeToString(id[:])
}

func main() {
    // sample fills and flows
    fills := []byte{0x00, 0x01, 0x10, 0x7f, 0xff}
    flows := []uint32{0, 1, 42, 12345}
    fmt.Printf("src_hex,dst_hex,flow,be32,be32_mod_1,be32_mod_2,be32_mod_3,be32_mod_4,be32_mod_5,be32_mod_8\n")
    for _, sf := range fills {
        for _, df := range fills {
            src := makeID(sf)
            dst := makeID(df)
            for _, flow := range flows {
                be := be32FromHash(src, dst, flow)
                fmt.Printf("%s,%s,%d,%d,%d,%d,%d,%d,%d,%d\n", hexID(src), hexID(dst), flow, be,
                    be%1, be%2, be%3, be%4, be%5, be%8)
            }
        }
    }
}
