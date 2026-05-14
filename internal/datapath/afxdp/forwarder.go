// ... inside the Forwarder struct definition ...
type Session struct {
    CryptoState *crypto.HybridSession // Now manages state machine
    NextHop     net.HardwareAddr
    FlowLabel   uint32
}

// The map key needs to be updated to reflect that we store a pointer/reference 
// that can manage the session lifecycle, not just data.

// --- Updated handleNewSession function signature and logic ---
func (f *Forwarder) handleNewSession(hdr wire.Header, frame []byte, desc *xdp.Desc) {
    // 1. Create a new Session object placeholder
    newSession := &Session{
        CryptoState: crypto.NewHybridSession(hdr.SessionID), // Initializes state to UNINITIALIZED
        FlowLabel:   hdr.FlowLabel,
    }

    // 2. ASYNCHRONOUS/SLOW PATH HANDSHAKE INITIATION
    // This entire block MUST run in a background goroutine or dedicated channel worker.
    go func() {
        // A) Attempt to initiate handshake (Assuming the first packet we see is an initiator's key exchange message)
        if err := newSession.CryptoState.InitiateHandshake(rand.Reader); err != nil { 
            f.logger.Warn("Failed to start handshake: ", zap.Error(err))
            return 
        }

        // B) Process the incoming packet as the first key exchange message payload (CRITICAL MOCK POINT)
        // In a real implementation, you would parse the *first few bytes* of 'frame' for public keys.
        if err := newSession.CryptoState.ProcessHandshakeMessage(mockPeerXPublic, mockPeerMLPublic); err != nil { 
            f.logger.Error("Handshake failed during key exchange", zap.Error(err))
            // Handle failure: drop packet or use a fallback/pre-shared key
            return
        }

        // C) Success! The session is now established in the crypto state machine.
        f.mu.Lock()
        f.sessions[hdr.SessionID] = newSession // Persist the fully initialized session
        f.mu.Unlock()
    }()
}


// --- Update processBatch logic ---
func (f *Forwarder) processBatch(numRx int) {
    rxDescs := f.xsk.Receive(numRx)
    for i := range rxDescs {
        desc := &rxDescs[i] // Use the descriptor pointer for reuse later
        frame := f.xsk.GetFrame(desc) 

        // ... header parsing (same as before) ...
        hdr, err := wire.ParseHeader(frame)
        if err != nil { continue }

        f.mu.RLock()
        sess, ok := f.sessions[hdr.SessionID]
        f.mu.RUnlock()

        if !ok || sess.CryptoState.State != crypto.ESTABLISHED {
            // 1. State check: If session doesn't exist OR if it is NOT ESTABLISHED, 
            //    it means this packet might be a handshake message (or we haven't finished setup).
            
            if newSession == nil { // Simple heuristic: if no session found, try initiating one
                 newSession = &Session{
                    CryptoState: crypto.NewHybridSession(hdr.SessionID),
                    FlowLabel:   hdr.FlowLabel,
                }
            }

            // Send the packet/state to the slow path for processing.
            f.handleNewSession(hdr, frame, desc) 
            continue
        }

        // === ZERO-COPY HOT PATH (Only runs if session is ESTABLISHED) ===
        if f.reuseDescriptorForForward(frame, desc, sess, hdr) {
            f.txPackets++
        } else {
            f.dropped++
            // Important: Always release the descriptor regardless of success/failure in the hot path
            f.releaseToFill(desc) 
        }
    }
}

