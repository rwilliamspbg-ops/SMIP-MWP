package crypto
// ... (imports remain the same) ...

// SessionState defines the current phase of key negotiation
type SessionState int

const (
    // UNINITIALIZED: No handshake attempt has been made for this session ID.
    UNINITIALIZED SessionState = iota
    // AWAITING_PEER_PUBKEY: We have sent our public keys and are waiting for the peer's response.
    AWAITING_PEER_PUBKEY 
    // READY_FOR_AUTH: Both sides exchanged keys; now we verify integrity or exchange final auth material.
    READY_FOR_AUTH 
    // ESTABLISHED: Keys have been successfully derived and are ready for data transfer.
    ESTABLISHED
)

// HybridSession holds the derived session state for a sovereign tunnel
type HybridSession struct {
    State      SessionState // Current state machine position
    Aead       cipher.AEAD
    NonceBase  [NonceSize]byte 
    SeqMask    uint64         
    
    // Temporary storage for handshake material (Only valid during negotiation)
    LocalXPrivate []byte
    LocalMLPrivate []byte

    // Peer's public key material received so far
    PeerXPublic MockPublicKey 
    PeerMLPublic MockPublicKey 
}


// NewHybridSession creates a session object, but does NOT perform the handshake.
func NewHybridSession(sessionID [16]byte) *HybridSession {
    return &HybridSession{
        State: UNINITIALIZED,
        // Note: NonceBase and SeqMask will be initialized on first use/handshake success
    }
}

// InitiateHandshake sets the local private keys and moves the state machine to AWAITING_PEER_PUBKEY.
func (s *HybridSession) InitiateHandshake(rand io.Reader) error {
    if s.State != UNINITIALIZED {
        return fmt.Errorf("cannot initiate handshake from current state: %v", s.State)
    }

    // 1. Generate/load local keys for the session
    if _, err := rand.Read(s.LocalXPrivate); err != nil { return err } // Mock Private Key X25519
    if _, err := rand.Read(s.LocalMLPrivate); err != nil { return err }  // Mock Private Key ML-KEM

    s.State = AWAITING_PEER_PUBKEY
    fmt.Println("INFO: Handshake initiated locally.")
    return nil
}


// ProcessHandshakeMessage processes an incoming packet, advancing the state machine if successful.
// This function should be called in a dedicated slow path goroutine.
func (s *HybridSession) ProcessHandshakeMessage(peerXPublic MockPublicKey, peerMLPublic MockPublicKey) error {
    switch s.State {
    case UNINITIALIZED:
        return fmt.Errorf("session not yet initialized; must call InitiateHandshake first")

    case AWAITING_PEER_PUBKEY:
        // This is the expected state upon receiving a public key payload.
        if len(peerXPublic) == 0 || len(peerMLPublic) == 0 {
            return fmt.Errorf("missing peer public keys in handshake message")
        }
        s.PeerXPublic = peerXPublic
        s.PeerMLPublic = peerMLPublic

        // Attempt to derive the shared secret material (the first half of the process)
        sharedSecret, err := s.performHybridKeyExchange()
        if err != nil { return fmt.Errorf("failed during hybrid key exchange: %w", err) }
        
        // Now that we have the material, finalize the session keys (HKDF Expansion)
        // Assume a fixed "sessionInfo" for simplicity here; in reality this includes flow label/etc.
        sessionInfo := []byte(fmt.Sprintf("MWP_SESSION_%x", s.LocalXPrivate)) 
        newSession, err := NewHybridSessionFromMaterial(sharedSecret, sessionInfo) // Use a dedicated constructor helper
        if err != nil { return fmt.Errorf("failed to expand keys: %w", err) }

        // State is now ready for data/authentication exchange (optional state transition here)
        s.Aead = newSession.Aead 
        // For simplicity, we jump to ESTABLISHED after successful key derivation
        s.State = ESTABLISHED 
        return nil

    case ESTABLISHED:
        return fmt.Errorf("session already established; received unexpected handshake packet")
    default:
        return fmt.Errorf("unhandled state transition from %v", s.State)
    }
}


// performHybridKeyExchange runs the core cryptographic exchange logic (Unchanged from before, but now part of a method).
func (s *HybridSession) performHybridKeyExchange() ([]byte, error) {
    // 1. Perform Key Agreement Exchanges
    xSecret, err := mockX25519KeyExchange(s.LocalXPrivate, s.PeerXPublic)
    if err != nil { return nil, fmt.Errorf("x25519 exchange failed: %w", err) }

    mlSecret, err := mockMlkemKeyExchange(s.LocalMLPrivate, s.PeerMLPublic)
    if err != nil { return nil, fmt.Errorf("ML-KEM exchange failed: %w", err) }

    // 2. Concatenate the shared secrets
    combinedSharedSecret := append(xSecret, mlSecret...)
    return combinedSharedSecret, nil
}


// --- Helper Functions (Update/Add to file): ---

// NewHybridSessionFromMaterial constructs a full session from already derived secret material.
func NewHybridSessionFromMaterial(sharedSecret []byte, sessionInfo []byte) (*HybridSession, error) {
    return NewHybridSession(sharedSecret, sessionInfo) // Reuse the existing constructor logic but use this helper to bypass key generation if needed
}

// Note: All previous methods (EncryptInPlace, DecryptInPlace, etc.) remain valid 
// as long as the state reaches ESTABLISHED.

// [ ... Rest of the file (EncryptInPlace, DecryptInPlace, etc.) remains unchanged ...] 
