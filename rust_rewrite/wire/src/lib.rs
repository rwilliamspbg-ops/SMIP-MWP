// wire crate: protocol types and parsing for SMIP

pub const HEADER_SIZE: usize = 32;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Header {
    pub src: [u8; 32],
    pub dst: [u8; 32],
    pub flow_label: u32,
    pub seq: u64,
    pub flags: u16,
    pub length: u16,
}

impl Header {
    pub fn new() -> Self {
        Self {
            src: [0; 32],
            dst: [0; 32],
            flow_label: 0,
            seq: 0,
            flags: 0,
            length: 0,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn header_basic() {
        let h = Header::new();
        assert_eq!(h.length, 0);
    }
}
