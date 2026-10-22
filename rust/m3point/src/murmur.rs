//! MurmurHash3 finalizer used by points and ids. The mixing constants are the Go ones.

pub const LOW: u64 = 0x0000_0000_ffff_ffff;
pub const HIGH: u64 = 0xffff_ffff_0000_0000;
const C1: u32 = 0xcc9e_2d51;
const C2: u32 = 0x1b87_3593;
const R1A: u32 = 15;
const R1B: u32 = 17;
const R2A: u32 = 13;
const R2B: u32 = 19;
const M: u32 = 4;
const N: u32 = 0xe654_6b64;

pub fn murmur_hash_u32(data: &[u32]) -> u32 {
    let mut h1: u32 = 0;
    for k in data {
        let mut k1 = k.wrapping_mul(C1);
        k1 = (k1 << R1A) | (k1 >> R1B);
        k1 = k1.wrapping_mul(C2);
        h1 ^= k1;
        h1 = (h1 << R2A) | (h1 >> R2B);
        h1 = h1.wrapping_mul(M).wrapping_add(h1).wrapping_add(N);
    }
    h1 ^= h1 >> 16;
    h1 = h1.wrapping_mul(0x85eb_ca6b);
    h1 ^= h1 >> 13;
    h1 = h1.wrapping_mul(0xc2b2_ae35);
    h1 ^= h1 >> 16;
    h1
}

pub fn murmur_hash_to_int(m: u32, size: i32) -> i32 {
    let res = (m as i32).wrapping_rem(size);
    if res < 0 {
        -res
    } else {
        res
    }
}

pub fn murmur_i32(id: i32) -> u32 {
    murmur_hash_u32(&[id as u32])
}

pub fn murmur_i64(id: i64) -> u32 {
    let c64 = id as u64;
    murmur_hash_u32(&[(c64 & LOW) as u32, ((c64 & HIGH) >> 32) as u32])
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::point::{create_random_point, CInt};

    #[test]
    fn hash_index_stays_inside_table() {
        let h = murmur_hash_u32(&[1, 2, 3]);
        let idx = murmur_hash_to_int(h, 64);
        assert!((0..64).contains(&idx));
    }

    #[test]
    fn hash_code_conflicts_stay_bounded() {
        run_hash(100, 2.0, 2.0, 10);
        run_hash(200, 7.0, 1.7, 10);
        run_hash(500, 7000.0, 2.5, 10);
    }

    fn run_hash(rd_max: CInt, run_ratio: f64, hash_size_ratio: f64, max_fails: usize) {
        let max_elements = (rd_max as f64).powi(3);
        let nb_run = (max_elements / run_ratio) as usize;
        let hash_size = (nb_run as f64 * hash_size_ratio) as i32;
        let mut data = Vec::with_capacity(nb_run);
        for _ in 0..nb_run {
            data.push(create_random_point(rd_max));
        }
        let mut buckets: std::collections::HashMap<i32, Vec<crate::point::Point>> =
            std::collections::HashMap::new();
        let mut no_more_space = 0;
        for p in &data {
            let m_hash = p.murmur_hash();
            let hash = murmur_hash_to_int(m_hash, hash_size);
            assert!(
                hash >= 0 && hash < hash_size,
                "hash {hash} not in 0..{hash_size}"
            );
            let points = buckets.entry(hash).or_default();
            if points.iter().any(|op| op == p) {
                continue;
            }
            if points.len() + 1 > max_fails {
                no_more_space += 1;
                break;
            }
            points.push(*p);
        }
        assert_eq!(no_more_space, 0);
    }
}
