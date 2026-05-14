// ... inside Forwarder struct definition ...
type Forwarder struct {
    xsk       *xdp.Socket
    program   *xdp.Program
    config    Config
    sessions  map[[16]byte]*Session 
    router    *routing.Router // NEW: The core policy engine
    framePool *FramePool
    mu         sync.RWMutex
    // ... other fields remain the same
}

func NewForwarder(cfg Config, logger *zap.Logger, router *routing.Router) (*Forwarder, error) {
    // ... existing setup code ...
    f := &Forwarder{
        // ... existing assignments ...
        router:    router, // Inject the router here!
    }
    // ... rest of function remains the same ...
    return f, nil
}

// Updated prepareForward now uses the router.
func (f *Forwarder) prepareForward(payload []byte, sess *Session, hdr wire.Header) []byte {
    // 1. POLICY LOOKUP: Determine the next hop and queue based on current packet data.
    policy, err := f.router.LookupPolicy(hdr.SrcID, hdr.DstID, hdr.FlowLabel)
    if err != nil {
        f.logger.Error("Routing policy lookup failed", zap.Error(err))
        return nil // Drop the packet if routing decision fails
    }

    // 2. Determine next hop based on policy result.
    nextDstID := policy.NextHopID
    
    // --- CORE ROUTING LOGIC ---
    newHdr := wire.Header{
        SrcID:     hdr.DstID, // Source identity for the *next* hop is often the current destination
        DstID:     nextDstID, // The next physical/logical address
        FlowLabel: sess.FlowLabel,
        SeqNum:    hdr.SeqNum + 1,
        SessionID: hdr.SessionID,
        Flags:     hdr.Flags,
        Length:    uint16(len(payload)),
    }

    // ... (rest of the function body remains identical, using newHdr) ...
    // ... Re-encrypt for next hop and return pooled frame ...
    return finalFrame // Assuming successful construction
}


func (f *Forwarder) handleNewSession(hdr wire.Header, frame []byte, desc *xdp.Desc) {
    // 1. Offload to control/slow path goroutine or channel
    go func() {
        newSession := &Session{
            CryptoState: crypto.NewHybridSession(hdr.SessionID), // Initialize state machine
            FlowLabel:   hdr.FlowLabel,
        }

        // 2. Attempt initial routing lookup to prime the session state's expected next hop/policy
        policy, err := f.router.LookupPolicy(hdr.SrcID, hdr.DstID, hdr.FlowLabel)
        if err != nil {
            f.logger.Error("Failed to determine initial route for new session", zap.Error(err))
            return
        }

        // 3. Perform the full handshake (The slow path magic)
        if stateErr := newSession.CryptoState.ProcessHandshakeMessage(mockPeerXPublic, mockPeerMLPublic); stateErr != nil {
             f.logger.Error("Failed to complete session handshake", zap.Error(stateErr))
            return
        }

        // 4. Success: Store the fully established session and its initial routing policy
        f.mu.Lock()
        f.sessions[hdr.SessionID] = newSession
        f.logger.Info("New sovereign tunnel established", zap.String("session_id", hexToString(hdr.SessionID))) // Add hex helper
        f.mu.Unlock()
    }()
}
