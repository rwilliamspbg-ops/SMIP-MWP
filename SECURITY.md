# Security Policy

Responsible disclosure and high-level security considerations for SMIP-MWP.

Reporting
- If you discover a security issue, please open a private issue or contact the maintainers via email.

Key Management
- Keep long-term secrets off the repository and deploy them via a secret manager (HashiCorp Vault, AWS KMS, etc.).
- Rotate session keys frequently and enforce maximum session lifetime.

Session Rotation
- Rotate session keys on rekey events and after N seconds or M messages (whichever comes first).

Audit & Logging
- Record handshake events, session creation, and crypto errors to a secure audit log with access controls.

Threat Model Summary
- Protect against replay, crypto misuse, DoS by validating sizes and bounding caches.
- Avoid deterministic nonces and ensure unique nonces per AEAD operation.

Contact
- For responsible disclosure, contact the project maintainers.
