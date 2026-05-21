// afxdp crate: XDP integration and mocks (scaffold)

pub fn available() -> bool {
    false
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn smoke() {
        assert!(!available());
    }
}
