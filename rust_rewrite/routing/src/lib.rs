// routing crate: route table and lookup (scaffold)

use std::collections::HashMap;

pub struct Table {
    routes: HashMap<[u8;32], [u8;32]>,
}

impl Table {
    pub fn new() -> Self {
        Self { routes: HashMap::new() }
    }
    pub fn update(&mut self, dest: [u8;32], next: [u8;32]) {
        self.routes.insert(dest, next);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn update_route() {
        let mut t = Table::new();
        t.update([1;32], [2;32]);
    }
}
