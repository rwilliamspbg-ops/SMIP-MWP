// datapath crate: forwarder core (scaffold)

use wire::Header;
use routing::Table;

pub struct Forwarder {
    pub cfg: (),
    pub routes: Table,
}

impl Forwarder {
    pub fn new(routes: Table) -> Self {
        Self { cfg: (), routes }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use routing::Table;

    #[test]
    fn make_forwarder() {
        let rt = Table::new();
        let _f = Forwarder::new(rt);
    }
}
