//! A cube of trio indexes around one main point: center, six faces, twelve edges.

use std::cmp::Ordering;
use std::collections::HashMap;

use crate::conn::UnitDirection;
use crate::growth::GrowthContext;
use crate::pack::PointPackData;
use crate::point::{CInt, Point, ORIGIN, THREE};
use crate::trio::TrioIndex;

pub const TOTAL_NUMBER_OF_CUBES: usize = 5192;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, Default)]
pub struct CubeOfTrioIndex {
    pub center: TrioIndex,
    pub center_faces: [TrioIndex; 6],
    pub middle_edges: [TrioIndex; 12],
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct CubeKey {
    pub growth_ctx_id: i32,
    pub cube: CubeOfTrioIndex,
}

impl Default for CubeKey {
    fn default() -> Self {
        Self {
            growth_ctx_id: -1,
            cube: CubeOfTrioIndex {
                center: TrioIndex::NIL,
                center_faces: [TrioIndex::NIL; 6],
                middle_edges: [TrioIndex::NIL; 12],
            },
        }
    }
}

/// Index of the edge between two unit directions, in the Go order.
pub fn middle_edge_index(mut ud1: UnitDirection, mut ud2: UnitDirection) -> usize {
    if ud1 == ud2 {
        panic!("Cannot find middle edge for 2 identical unit direction {ud1}");
    }
    if ud1 > ud2 {
        std::mem::swap(&mut ud1, &mut ud2);
    }
    if (ud1 as u8) % 2 == 0 && ud1 as u8 == ud2 as u8 - 1 {
        panic!("Cannot find middle edge for {ud1} and {ud2} since they are on same axis");
    }
    match ud1 {
        UnitDirection::PlusX => (ud2 as u8 - 2) as usize,
        UnitDirection::MinusX => (4 + ud2 as u8 - 2) as usize,
        UnitDirection::PlusY => (8 + ud2 as u8 - 4) as usize,
        UnitDirection::MinusY => (8 + 2 + ud2 as u8 - 4) as usize,
        _ => panic!("Cannot find middle edge for {ud1} and {ud2}"),
    }
}

impl CubeOfTrioIndex {
    pub fn center_face(self, ud: UnitDirection) -> TrioIndex {
        self.center_faces[ud as usize]
    }

    pub fn middle_edge(self, ud1: UnitDirection, ud2: UnitDirection) -> TrioIndex {
        self.middle_edges[middle_edge_index(ud1, ud2)]
    }
}

impl PointPackData {
    pub fn create_trio_cube(&self, ctx: &GrowthContext, offset: i32, c: Point) -> CubeOfTrioIndex {
        let idx = |p: Point| self.base_trio_index(ctx, ctx.base_div_by_three(p), offset);
        let mut res = CubeOfTrioIndex {
            center: idx(c),
            center_faces: [TrioIndex::NIL; 6],
            middle_edges: [TrioIndex::NIL; 12],
        };
        use crate::point::{X_FIRST, Y_FIRST, Z_FIRST};
        use UnitDirection::*;
        res.center_faces[PlusX as usize] = idx(c.add(X_FIRST));
        res.center_faces[MinusX as usize] = idx(c.sub(X_FIRST));
        res.center_faces[PlusY as usize] = idx(c.add(Y_FIRST));
        res.center_faces[MinusY as usize] = idx(c.sub(Y_FIRST));
        res.center_faces[PlusZ as usize] = idx(c.add(Z_FIRST));
        res.center_faces[MinusZ as usize] = idx(c.sub(Z_FIRST));

        let mut edge = |ud1, ud2, p: Point| {
            res.middle_edges[middle_edge_index(ud1, ud2)] = idx(p);
        };
        edge(PlusX, PlusY, c.add(X_FIRST).add(Y_FIRST));
        edge(PlusX, MinusY, c.add(X_FIRST).sub(Y_FIRST));
        edge(PlusX, PlusZ, c.add(X_FIRST).add(Z_FIRST));
        edge(PlusX, MinusZ, c.add(X_FIRST).sub(Z_FIRST));
        edge(MinusX, PlusY, c.sub(X_FIRST).add(Y_FIRST));
        edge(MinusX, MinusY, c.sub(X_FIRST).sub(Y_FIRST));
        edge(MinusX, PlusZ, c.sub(X_FIRST).add(Z_FIRST));
        edge(MinusX, MinusZ, c.sub(X_FIRST).sub(Z_FIRST));
        edge(PlusY, PlusZ, c.add(Y_FIRST).add(Z_FIRST));
        edge(PlusY, MinusZ, c.add(Y_FIRST).sub(Z_FIRST));
        edge(MinusY, PlusZ, c.sub(Y_FIRST).add(Z_FIRST));
        edge(MinusY, MinusZ, c.sub(Y_FIRST).sub(Z_FIRST));
        res
    }
}

pub fn calculate_all_context_cubes(pack: &PointPackData) -> (HashMap<CubeKey, i32>, Vec<CubeKey>) {
    let mut map = HashMap::with_capacity(TOTAL_NUMBER_OF_CUBES);
    let mut by_id = Vec::with_capacity(TOTAL_NUMBER_OF_CUBES + 1);
    by_id.push(CubeKey::default());
    let mut cube_idx = 1i32;
    for growth_ctx in &pack.growth_contexts {
        let mut cubes = populate_cubes(pack, growth_ctx);
        cubes.sort_by(cmp_cube);
        for cube in cubes {
            let key = CubeKey {
                growth_ctx_id: growth_ctx.id,
                cube,
            };
            if !map.contains_key(&key) {
                map.insert(key, cube_idx);
                by_id.push(key);
                cube_idx += 1;
            }
        }
    }
    (map, by_id)
}

fn populate_limit(ctx: &GrowthContext) -> CInt {
    match ctx.growth_type.0 {
        1 => 1,
        3 => 6,
        2 => 1,
        4 => 4,
        8 => 8,
        other => panic!("unknown growth type {other}"),
    }
}

fn populate_cubes(pack: &PointPackData, ctx: &GrowthContext) -> Vec<CubeOfTrioIndex> {
    let mut all: HashMap<CubeOfTrioIndex, i32> = HashMap::new();
    let max_offset = ctx.growth_type.max_offset();
    for offset in 0..max_offset {
        let cube = pack.create_trio_cube(ctx, offset, ORIGIN);
        *all.entry(cube).or_insert(0) += 1;
    }
    let max = populate_limit(ctx);
    for x in -max..=max {
        for y in -max..=max {
            for z in -max..=max {
                let cube = pack.create_trio_cube(ctx, 0, Point([x, y, z]).mul(THREE));
                *all.entry(cube).or_insert(0) += 1;
            }
        }
    }
    all.into_keys().collect()
}

fn cmp_cube(c1: &CubeOfTrioIndex, c2: &CubeOfTrioIndex) -> Ordering {
    let center = (c1.center.0 as i32).cmp(&(c2.center.0 as i32));
    if center != Ordering::Equal {
        return center;
    }
    for i in 0..c1.center_faces.len() {
        let d = (c1.center_faces[i].0 as i32).cmp(&(c2.center_faces[i].0 as i32));
        if d != Ordering::Equal {
            return d;
        }
    }
    for i in 0..c1.middle_edges.len() {
        let d = (c1.middle_edges[i].0 as i32).cmp(&(c2.middle_edges[i].0 as i32));
        if d != Ordering::Equal {
            return d;
        }
    }
    Ordering::Equal
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::growth::ALL_GROWTH_TYPES;
    use crate::pack::PointPackData;
    use std::collections::HashSet;

    #[test]
    fn trio_cube_maps_cover_the_lattice() {
        let ppd = PointPackData::shared();
        assert_eq!(ppd.cube_ids.len(), TOTAL_NUMBER_OF_CUBES);
        for ctx_type in ALL_GROWTH_TYPES {
            for p_idx in 0..ctx_type.nb_indexes() {
                let ctx = ppd.growth_by_type_and_index(ctx_type, p_idx);
                let (max, cubes) = find_nb_cubes(ppd, ctx);
                let big = distinct_cubes(ppd, ctx, max * 3);
                assert_eq!(cubes.len(), big.len(), "{ctx} max={max}");
                let saved = cubes_for_ctx(ppd, ctx.id);
                assert_eq!(cubes.len(), saved.len(), "{ctx}");
                for offset in 0..ctx_type.max_offset() {
                    assert_with_offset(ppd, ctx, &saved, max + 1, offset);
                }
            }
        }
    }

    fn cubes_for_ctx(ppd: &PointPackData, ctx_id: i32) -> HashSet<CubeOfTrioIndex> {
        ppd.cube_ids
            .keys()
            .filter(|k| k.growth_ctx_id == ctx_id)
            .map(|k| k.cube)
            .collect()
    }

    fn find_nb_cubes(
        ppd: &PointPackData,
        ctx: &GrowthContext,
    ) -> (CInt, HashMap<CubeOfTrioIndex, i32>) {
        let mut nb = 0;
        let mut max = 1;
        let mut latest = HashMap::new();
        while max < 30 {
            latest = distinct_cubes(ppd, ctx, max);
            if nb == latest.len() {
                break;
            }
            nb = latest.len();
            max += 1;
        }
        (max - 1, latest)
    }

    fn distinct_cubes(
        ppd: &PointPackData,
        ctx: &GrowthContext,
        max: CInt,
    ) -> HashMap<CubeOfTrioIndex, i32> {
        let mut all = HashMap::new();
        for offset in 0..ctx.growth_type.max_offset() {
            let cube = ppd.create_trio_cube(ctx, offset, ORIGIN);
            *all.entry(cube).or_insert(0) += 1;
        }
        for x in -max..=max {
            for y in -max..=max {
                for z in -max..=max {
                    let cube = ppd.create_trio_cube(ctx, 0, Point([x, y, z]).mul(THREE));
                    *all.entry(cube).or_insert(0) += 1;
                }
            }
        }
        all
    }

    fn assert_with_offset(
        ppd: &PointPackData,
        ctx: &GrowthContext,
        saved: &HashSet<CubeOfTrioIndex>,
        max: CInt,
        offset: i32,
    ) {
        for x in -max..=max {
            for y in -max..=max {
                for z in -max..=max {
                    let mp = Point([x, y, z]).mul(THREE);
                    let cube = ppd.create_trio_cube(ctx, offset, mp);
                    assert!(saved.contains(&cube), "{ctx} offset {offset} at {mp}");
                }
            }
        }
    }
}
