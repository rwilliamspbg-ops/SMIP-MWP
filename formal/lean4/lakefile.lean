import Lake
open Lake DSL

package "smip_mwp" {
  -- package configuration
}

lean_lib Smip where
  srcDir := "Lean"

-- Provide an executable entrypoint to generate HKDF vectors from the model.
-- Run with `lake run GenerateHKDFVectors`.
lean_exe GenerateHKDFVectors where
  root := `Smip.GenerateHKDFVectors
