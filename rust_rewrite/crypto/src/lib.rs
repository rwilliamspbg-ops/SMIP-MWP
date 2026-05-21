// crypto crate: key exchange, session, AEAD wrappers (scaffold)

pub fn supported() -> bool {
    true
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn smoke() {
        assert!(supported());
    }
}
