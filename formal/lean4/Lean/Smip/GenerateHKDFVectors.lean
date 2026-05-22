import Smip.CryptoHandshakeSpec
open Smip
open System

/-- Convert a nibble (0..15) to a hex character. -/
def nibbleToChar : Nat -> Char
| 0 => '0'
| 1 => '1'
| 2 => '2'
| 3 => '3'
| 4 => '4'
| 5 => '5'
| 6 => '6'
| 7 => '7'
| 8 => '8'
| 9 => '9'
| 10 => 'a'
| 11 => 'b'
| 12 => 'c'
| 13 => 'd'
| 14 => 'e'
| 15 => 'f'
| _ => '?'

def byteToHex (n : Nat) : String :=
  let hi := n / 16
  let lo := n % 16
  String.ofList [nibbleToChar hi, nibbleToChar lo]

def bytesToHex (bs : List Nat) : String :=
  bs.foldl (fun acc b => acc ++ byteToHex b) ""

def sampleCombined0 : ByteSeq := List.replicate 32 0
def sampleCombined1 : ByteSeq := List.replicate 32 1
def sampleSession0 : ByteSeq := [115,101,115,115,105,111,110,45,48] -- "session-0"
def sampleSession1 : ByteSeq := [115,101,115,115,105,111,110,45,49] -- "session-1"

def lineOf (combined sessionInfo : ByteSeq) : String :=
  let mat := deriveSessionMaterial combined sessionInfo
  let combinedHex := bytesToHex combined
  let sessionHex := bytesToHex sessionInfo
  let keyHex := bytesToHex mat.key
  let nonceHex := bytesToHex mat.nonceBase
  let maskStr := toString mat.seqMask
  combinedHex ++ "," ++ sessionHex ++ "," ++ keyHex ++ "," ++ nonceHex ++ "," ++ maskStr

def csvContent : String :=
  let header := "combined_hex,sessionInfo_hex,key_hex,nonce_base_hex,seqMask\n"
  header ++ (lineOf sampleCombined0 sampleSession0) ++ "\n" ++ (lineOf sampleCombined1 sampleSession1) ++ "\n"

def generate : IO Unit := do
  let fname := "SmipHKDFVectors_from_lean.csv"
  IO.FS.writeFile fname csvContent
  IO.println s!"wrote {fname}"
