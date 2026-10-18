//! Base trios, valid next-trio pairs, and the modulo-4 / modulo-8 permutations.

use crate::point::{Point, X_FIRST, Y_FIRST, Z_FIRST};
use crate::trio::TrioIndex;

/// Three base connecting vectors, slotted so index 0 has `z = 0`, 1 has `y = 0`, 2 has `x = 0`.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct BaseTrio(pub [Point; 3]);

const REVERSE3: [usize; 3] = [2, 1, 0];

pub fn make_base_connecting_vectors_trio(points: [Point; 3]) -> BaseTrio {
    for p in points {
        if !p.is_base_connecting_vector() {
            panic!("Trying to create a base trio out of non base vector {p}");
        }
    }
    let mut res = BaseTrio([Point([0, 0, 0]); 3]);
    for p in points {
        for i in 0..3 {
            if p.0[i] == 0 {
                res.0[REVERSE3[i]] = p;
            }
        }
    }
    res
}

impl BaseTrio {
    pub fn rot_plus_x(self) -> Self {
        make_base_connecting_vectors_trio([
            self.0[0].rot_plus_x(),
            self.0[1].rot_plus_x(),
            self.0[2].rot_plus_x(),
        ])
    }

    pub fn neg(self) -> Self {
        make_base_connecting_vectors_trio([self.0[0].neg(), self.0[1].neg(), self.0[2].neg()])
    }

    pub fn axis_vector(self, axis: usize, sign: i32) -> Point {
        for vec in self.0 {
            if vec.0[axis] == sign {
                return vec;
            }
        }
        panic!("Impossible! For all base trio there should be a signed vector on axis {axis}");
    }

    fn ds_index(self) -> i32 {
        if self.0[0].distance_squared() == 1 {
            1
        } else {
            match self.0[1].distance_squared() {
                2 => 0,
                3 => 2,
                5 => 3,
                _ => -1,
            }
        }
    }
}

/// `T[i1]` is the negation of `T[i2]` for base indexes.
pub fn is_prime(i1: TrioIndex, i2: TrioIndex) -> bool {
    if !i1.is_base() || !i2.is_base() {
        return false;
    }
    if i1 > i2 {
        return i1.0 == i2.0 + 4;
    }
    if i2 > i1 {
        return i2.0 == i1.0 + 4;
    }
    false
}

pub fn all_base_trios() -> [BaseTrio; 8] {
    let mut all = [BaseTrio([Point([0, 0, 0]); 3]); 8];
    all[0] = make_base_connecting_vectors_trio([
        Point([1, 1, 0]),
        Point([-1, 0, -1]),
        Point([0, -1, 1]),
    ]);
    for i in 1..4 {
        all[i] = all[i - 1].rot_plus_x();
    }
    for i in 0..4 {
        all[i + 4] = all[i].neg();
    }
    all
}

pub fn init_valid_trios() -> [[TrioIndex; 2]; 12] {
    let mut valid = [[TrioIndex::NIL; 2]; 12];
    let mut idx = 0;
    for i in 0..4u8 {
        for j in 4..8u8 {
            if !is_prime(TrioIndex(i), TrioIndex(j)) {
                valid[idx] = [TrioIndex(i), TrioIndex(j)];
                idx += 1;
            }
        }
    }
    assert_eq!(idx, 12, "valid next trio table must have 12 rows");
    valid
}

struct PermBuilder {
    size: usize,
    col_idx: usize,
    collector: Vec<Vec<TrioIndex>>,
}

pub fn same_permutation(p1: &[TrioIndex], p2: &[TrioIndex]) -> bool {
    if p1.len() != p2.len() {
        panic!("cannot test 2 permutation of different sizes {p1:?} {p2:?}");
    }
    let perm_size = p1.len();
    let mut idx0 = None;
    for (idx, value) in p2.iter().enumerate() {
        if *value == p1[0] {
            idx0 = Some(idx);
            break;
        }
    }
    let Some(idx0) = idx0 else {
        return false;
    };
    for idx in 0..perm_size {
        if p2[(idx0 + idx) % perm_size] != p1[idx] {
            return false;
        }
    }
    true
}

impl PermBuilder {
    fn fill(&mut self, pos: usize, current: Vec<TrioIndex>) {
        if pos == self.size {
            let exists = self.collector[..self.col_idx]
                .iter()
                .any(|c| same_permutation(c, &current));
            if !exists {
                if self.col_idx >= self.collector.len() {
                    self.collector.push(current);
                } else {
                    self.collector[self.col_idx] = current;
                }
                self.col_idx += 1;
            }
            return;
        }
        for i in 0..4 {
            let new_index = if pos % 2 == 1 {
                TrioIndex((i + 4) as u8)
            } else {
                TrioIndex(i as u8)
            };
            let mut usable = true;
            if pos >= 1 {
                for j in 0..pos - 1 {
                    if current[j] == new_index {
                        usable = false;
                    }
                }
                if is_prime(new_index, current[pos - 1]) {
                    usable = false;
                }
            }
            if pos + 1 == self.size && is_prime(new_index, current[0]) {
                usable = false;
            }
            if usable {
                let mut perm = current.clone();
                perm[pos] = new_index;
                self.fill(pos + 1, perm);
            }
        }
    }
}

fn collect_perms(size: usize, seed: Option<TrioIndex>) -> Vec<Vec<TrioIndex>> {
    let mut builder = PermBuilder {
        size,
        col_idx: 0,
        collector: vec![Vec::new(); 12],
    };
    let mut first = vec![TrioIndex(0); size];
    let start = if let Some(fixed) = seed {
        first[0] = fixed;
        1
    } else {
        0
    };
    builder.fill(start, first);
    builder.collector.truncate(builder.col_idx);
    builder.collector
}

pub fn init_mod4_permutations() -> [[TrioIndex; 4]; 12] {
    let perms = collect_perms(4, None);
    if perms.len() != 12 {
        panic!("mod4 produced {} permutations: {:?}", perms.len(), perms);
    }
    let mut out = [[TrioIndex::NIL; 4]; 12];
    for (p_idx, perm) in perms.iter().enumerate() {
        for i in 0..4 {
            out[p_idx][i] = perm[i];
        }
    }
    out
}

pub fn init_mod8_permutations() -> [[TrioIndex; 8]; 12] {
    let perms = collect_perms(8, Some(TrioIndex(0)));
    if perms.len() != 12 {
        panic!("mod8 produced {} permutations: {:?}", perms.len(), perms);
    }
    let mut out = [[TrioIndex::NIL; 8]; 12];
    for (p_idx, perm) in perms.iter().enumerate() {
        for i in 0..8 {
            out[p_idx][i] = perm[i];
        }
    }
    out
}

/// Six non-base connections between two base trios, ordered `+X,-X,+Y,-Y,+Z,-Z`.
pub fn get_non_base_connections(t_a: BaseTrio, t_b: BaseTrio) -> [Point; 6] {
    let mut res = [Point([0, 0, 0]); 6];
    for a_vec in t_a.0 {
        match a_vec.x() {
            0 => {}
            1 => {
                let b_vec = t_b.axis_vector(0, -1);
                res[0] = X_FIRST.add(b_vec).sub(a_vec);
            }
            -1 => {
                let b_vec = t_b.axis_vector(0, 1);
                res[1] = X_FIRST.neg().add(b_vec).sub(a_vec);
            }
            _ => {}
        }
        match a_vec.y() {
            0 => {}
            1 => {
                let b_vec = t_b.axis_vector(1, -1);
                res[2] = Y_FIRST.add(b_vec).sub(a_vec);
            }
            -1 => {
                let b_vec = t_b.axis_vector(1, 1);
                res[3] = Y_FIRST.neg().add(b_vec).sub(a_vec);
            }
            _ => {}
        }
        match a_vec.z() {
            0 => {}
            1 => {
                let b_vec = t_b.axis_vector(2, -1);
                res[4] = Z_FIRST.add(b_vec).sub(a_vec);
            }
            -1 => {
                let b_vec = t_b.axis_vector(2, 1);
                res[5] = Z_FIRST.neg().add(b_vec).sub(a_vec);
            }
            _ => {}
        }
    }
    res
}

#[allow(dead_code)]
fn _base_ds(t: BaseTrio) -> i32 {
    t.ds_index()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn same_permutation_rotations() {
        let a = [TrioIndex(1), TrioIndex(2), TrioIndex(3), TrioIndex(4)];
        assert!(same_permutation(
            &a,
            &[TrioIndex(1), TrioIndex(2), TrioIndex(3), TrioIndex(4)]
        ));
        assert!(same_permutation(
            &a,
            &[TrioIndex(4), TrioIndex(1), TrioIndex(2), TrioIndex(3)]
        ));
        assert!(same_permutation(
            &a,
            &[TrioIndex(3), TrioIndex(4), TrioIndex(1), TrioIndex(2)]
        ));
        assert!(same_permutation(
            &a,
            &[TrioIndex(2), TrioIndex(3), TrioIndex(4), TrioIndex(1)]
        ));
        assert!(!same_permutation(
            &a,
            &[TrioIndex(1), TrioIndex(2), TrioIndex(4), TrioIndex(3)]
        ));
        assert!(!same_permutation(
            &a,
            &[TrioIndex(3), TrioIndex(1), TrioIndex(2), TrioIndex(4)]
        ));
    }

    #[test]
    fn perm_builder_fills_twelve_mod4_rows() {
        let p = init_mod4_permutations();
        assert_eq!(p.len(), 12);
        for row in &p {
            assert_eq!(row.len(), 4);
        }
    }

    #[test]
    fn initial_trio_connecting_vectors() {
        let all = all_base_trios();
        assert_eq!(all[0].0[0], Point([1, 1, 0]));
        assert_eq!(all[0].0[1], Point([-1, 0, -1]));
        assert_eq!(
            all[0].0[1],
            all[0].0[0].rot_plus_x().rot_plus_y().rot_plus_y()
        );
        assert_eq!(all[0].0[2], Point([0, -1, 1]));
        assert_eq!(
            all[0].0[2],
            all[0].0[0].rot_plus_y().rot_plus_x().rot_plus_x()
        );
        assert_eq!(all[4].0[0], Point([-1, -1, 0]));
        assert_eq!(all[4].0[1], Point([1, 0, 1]));
        assert_eq!(all[4].0[2], Point([0, 1, -1]));
    }

    #[test]
    fn all_base_trio_sum_to_origin() {
        use crate::point::{abs_cint, ORIGIN};
        let all = all_base_trios();
        for (i, tr) in all.iter().enumerate() {
            assert_eq!(tr.0[0].z(), 0, "Failed on trio {i}");
            assert_eq!(tr.0[1].y(), 0, "Failed on trio {i}");
            assert_eq!(tr.0[2].x(), 0, "Failed on trio {i}");
            let mut back = ORIGIN;
            for (j, vec) in tr.0.iter().enumerate() {
                for c in vec.0 {
                    let abs = abs_cint(c);
                    assert!(abs == 1 || abs == 0, "coord {i} {j} {vec}");
                }
                assert_eq!(vec.distance_squared(), 2, "Failed vec at {i} {j}");
                assert!(vec.is_base_connecting_vector());
                back = back.add(*vec);
            }
            assert_eq!(back, ORIGIN, "sum of trio {i}");
            for (j, t_b) in all.iter().enumerate() {
                for conn in get_non_base_connections(*tr, *t_b) {
                    assert!(conn.is_connection_vector(), "wrong {conn} at {i} {j}");
                    assert!(!conn.is_base_connecting_vector(), "wrong {conn} at {i} {j}");
                    let d = conn.distance_squared();
                    assert!(d == 1 || d == 3 || d == 5, "wrong {conn} at {i} {j}");
                }
            }
        }
    }

    #[test]
    fn full5_and_valid_and_mod_permutations() {
        let all = all_base_trios();
        let valid = init_valid_trios();
        let mut idx_map = [0i32; 8];
        for i in 0..4 {
            let next = [TrioIndex(i), TrioIndex(i + 4)];
            assert_valid_next(next, i as usize);
            for conn in get_non_base_connections(all[next[0].0 as usize], all[next[1].0 as usize]) {
                assert!(conn.is_only_two_one_and_zero());
                assert!(conn.is_connection_vector());
                assert!(!conn.is_base_connecting_vector());
                assert_eq!(conn.distance_squared(), 5);
            }
            idx_map[next[0].0 as usize] += 1;
            idx_map[next[1].0 as usize] += 1;
        }
        assert_eq!(idx_map, [1; 8]);

        let mut idx_map = [0i32; 8];
        for (i, next) in valid.iter().enumerate() {
            assert_valid_next(*next, i);
            for conn in get_non_base_connections(all[next[0].0 as usize], all[next[1].0 as usize]) {
                assert!(conn.is_only_one_and_zero(), "wrong {conn}");
                assert!(conn.is_connection_vector());
                assert!(!conn.is_base_connecting_vector());
                let d = conn.distance_squared();
                assert!(d == 1 || d == 3, "wrong {conn}");
            }
            idx_map[next[0].0 as usize] += 1;
            idx_map[next[1].0 as usize] += 1;
        }
        assert_eq!(idx_map, [3; 8]);

        let mod4 = init_mod4_permutations();
        let mut idx_map = [0i32; 8];
        for (i, permut) in mod4.iter().enumerate() {
            for j in 0..4 {
                let start = permut[j];
                let end = permut[(j + 1) % 4];
                assert_in_valid(start, end, &valid, &format!("mod4[{i}] idx {j}"));
                idx_map[permut[j].0 as usize] += 1;
            }
        }
        assert_eq!(idx_map, [6; 8]);

        let mod8 = init_mod8_permutations();
        for (i, permut) in mod8.iter().enumerate() {
            let mut idx_map = [0i32; 8];
            for j in 0..8 {
                let start = permut[j];
                let end = permut[(j + 1) % 8];
                assert_in_valid(start, end, &valid, &format!("mod8[{i}] idx {j}"));
                idx_map[permut[j].0 as usize] += 1;
            }
            assert_eq!(idx_map, [1; 8], "mod8 row {i}");
        }
    }

    fn assert_valid_next(next: [TrioIndex; 2], i: usize) {
        assert_ne!(next[0], next[1], "index {i}");
        assert!(next[0].0 <= 3, "index {i}");
        assert!((4..=7).contains(&next[1].0), "index {i}");
    }

    fn assert_in_valid(start: TrioIndex, end: TrioIndex, valid: &[[TrioIndex; 2]; 12], msg: &str) {
        assert_ne!(start, end, "{msg}");
        let mut trio = [TrioIndex::NIL, TrioIndex::NIL];
        if start.0 >= 4 {
            trio[1] = start;
        } else {
            trio[0] = start;
        }
        if end.0 >= 4 {
            trio[1] = end;
        } else {
            trio[0] = end;
        }
        assert!(trio[0].0 <= 3, "{msg}");
        assert!((4..=7).contains(&trio[1].0), "{msg}");
        assert!(
            valid.iter().any(|n| *n == trio),
            "missing {trio:?} for {msg}"
        );
    }
}
