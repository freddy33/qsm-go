//! In-memory point pack: every connection, trio, growth context, cube, and path builder.
//!
//! The Go code stored this in PostgreSQL. The tables are derived, so the same
//! data is rebuilt here with the original ordering rules.

use std::cmp::Ordering;
use std::collections::HashMap;
use std::sync::OnceLock;

use crate::conn::{conn_vector_less, is_less_conn_id, ConnectionDetails, ConnectionId, EMPTY_CONN};
use crate::cube::{calculate_all_context_cubes, CubeKey};
use crate::growth::{all_growth_contexts, base_trio_index, GrowthContext, GrowthType};
use crate::pathbuilder::{calculate_all_path_builders, NodeHandle, RootBuilder};
use crate::perms::{
    all_base_trios, get_non_base_connections, init_mod4_permutations, init_mod8_permutations,
    init_valid_trios, BaseTrio,
};
use crate::point::{make_vector, Point, X_FIRST, Y_FIRST, Z_FIRST};
use crate::trio::{sort_three_conns, TrioDetails, TrioIndex};

pub const EXPECTED_CONNECTIONS: usize = 50;
pub const EXPECTED_TRIOS: usize = 200;
pub const EXPECTED_GROWTH_CONTEXTS: usize = 52;
pub const EXPECTED_CUBES: usize = 5192;

pub struct PointPackData {
    pub connections: Vec<ConnectionDetails>,
    pub connections_by_vector: HashMap<Point, ConnectionDetails>,
    pub trios: Vec<TrioDetails>,
    pub growth_contexts: Vec<GrowthContext>,
    pub base_trios: [BaseTrio; 8],
    pub valid_next_trio: [[TrioIndex; 2]; 12],
    pub mod4: [[TrioIndex; 4]; 12],
    pub mod8: [[TrioIndex; 8]; 12],
    pub cube_ids: HashMap<CubeKey, i32>,
    pub cubes_by_id: Vec<CubeKey>,
    pub path_builders: Vec<Option<RootBuilder>>,
}

impl PointPackData {
    /// Build the full derived data set. Safe to call more than once; [`shared`] caches it.
    pub fn calculate() -> Self {
        let base_trios = all_base_trios();
        let valid_next_trio = init_valid_trios();
        let mod4 = init_mod4_permutations();
        let mod8 = init_mod8_permutations();
        let (connections, connections_by_vector) = calculate_connections(&base_trios);
        let trios = calculate_trios(&base_trios, &connections_by_vector);
        let growth_contexts = all_growth_contexts();
        let mut pack = PointPackData {
            connections,
            connections_by_vector,
            trios,
            growth_contexts,
            base_trios,
            valid_next_trio,
            mod4,
            mod8,
            cube_ids: HashMap::new(),
            cubes_by_id: Vec::new(),
            path_builders: Vec::new(),
        };
        let (cube_ids, cubes_by_id) = calculate_all_context_cubes(&pack);
        pack.cube_ids = cube_ids;
        pack.cubes_by_id = cubes_by_id;
        pack.path_builders = calculate_all_path_builders(&pack);
        pack
    }

    pub fn shared() -> &'static PointPackData {
        static PACK: OnceLock<PointPackData> = OnceLock::new();
        PACK.get_or_init(PointPackData::calculate)
    }

    pub fn max_conn_id(&self) -> ConnectionId {
        self.connections
            .last()
            .map(|c| c.pos_id())
            .unwrap_or(ConnectionId::NIL)
    }

    pub fn conn_by_id(&self, id: ConnectionId) -> ConnectionDetails {
        if id.0 > 0 {
            self.connections[2 * id.0 as usize - 2]
        } else if id.0 < 0 {
            let idx = (-2 * i32::from(id.0) - 1) as usize;
            self.connections[idx]
        } else {
            EMPTY_CONN
        }
    }

    pub fn conn_by_vector(&self, vector: Point) -> ConnectionDetails {
        self.connections_by_vector
            .get(&vector)
            .copied()
            .unwrap_or(EMPTY_CONN)
    }

    pub fn conn_by_points(&self, p1: Point, p2: Point) -> ConnectionDetails {
        self.conn_by_vector(make_vector(p1, p2))
    }

    pub fn trio(&self, idx: TrioIndex) -> &TrioDetails {
        &self.trios[idx.0 as usize]
    }

    pub fn growth_by_id(&self, id: i32) -> &GrowthContext {
        &self.growth_contexts[id as usize]
    }

    pub fn growth_by_type_and_index(&self, growth_type: GrowthType, index: i32) -> &GrowthContext {
        self.growth_contexts
            .iter()
            .find(|g| g.growth_type == growth_type && g.growth_index == index)
            .unwrap_or_else(|| panic!("could not find trio Context for {} {index}", growth_type.0))
    }

    pub fn base_trio_index(
        &self,
        ctx: &GrowthContext,
        div_by_three: u64,
        offset: i32,
    ) -> TrioIndex {
        base_trio_index(
            ctx,
            div_by_three,
            offset,
            &self.valid_next_trio,
            &self.mod4,
            &self.mod8,
        )
    }

    pub fn trio_for_main(&self, ctx: &GrowthContext, main: Point, offset: i32) -> &TrioDetails {
        let idx = self.base_trio_index(ctx, ctx.base_div_by_three(main), offset);
        self.trio(idx)
    }

    pub fn cube_id_by_key(&self, key: CubeKey) -> i32 {
        *self.cube_ids.get(&key).unwrap_or_else(|| {
            panic!(
                "trying to find cube which does not exist for ctx {}",
                key.growth_ctx_id
            )
        })
    }

    pub fn cube_by_id(&self, cube_id: i32) -> CubeKey {
        self.cubes_by_id[cube_id as usize]
    }

    pub fn path_node_builder(&self, ctx: &GrowthContext, offset: i32, c: Point) -> NodeHandle {
        let cube = self.create_trio_cube(ctx, offset, c);
        let key = CubeKey {
            growth_ctx_id: ctx.id,
            cube,
        };
        let cube_id = self.cube_id_by_key(key);
        NodeHandle::Root(cube_id)
    }
}

fn calculate_connections(
    base: &[BaseTrio; 8],
) -> (Vec<ConnectionDetails>, HashMap<Point, ConnectionDetails>) {
    let mut conn_map: HashMap<Point, ConnectionDetails> = HashMap::new();
    for tr in base {
        for vec in tr.0 {
            add_conn_detail(&mut conn_map, vec);
        }
        for t_b in base {
            for conn in get_non_base_connections(*tr, *t_b) {
                add_conn_detail(&mut conn_map, conn);
            }
        }
    }
    let nb = ConnectionId((conn_map.len() / 2) as i8);

    let mut ordered: Vec<Point> = conn_map.keys().copied().collect();
    ordered.sort_by(|a, b| {
        let ca = &conn_map[a];
        let cb = &conn_map[b];
        if conn_vector_less(ca, cb) {
            Ordering::Less
        } else if conn_vector_less(cb, ca) {
            Ordering::Greater
        } else {
            Ordering::Equal
        }
    });

    let mut current = 1i8;
    for vec1 in ordered {
        if conn_map[&vec1].id.0 != 0 {
            continue;
        }
        let vec2 = vec1.neg();
        let mut pos_vec = vec1;
        let mut neg_vec = vec2;
        for c in vec1.0 {
            if c > 0 {
                pos_vec = vec1;
                neg_vec = vec2;
                break;
            } else if c < 0 {
                pos_vec = vec2;
                neg_vec = vec1;
                break;
            }
        }
        conn_map.get_mut(&pos_vec).unwrap().id = ConnectionId(current);
        conn_map.get_mut(&neg_vec).unwrap().id = ConnectionId(-current);
        current += 1;
    }

    let mut res: Vec<ConnectionDetails> = conn_map.values().copied().collect();
    res.sort_by(|a, b| {
        if is_less_conn_id(a, b) {
            Ordering::Less
        } else if is_less_conn_id(b, a) {
            Ordering::Greater
        } else {
            Ordering::Equal
        }
    });
    // Refresh the map so stored copies carry the assigned ids.
    for cd in &res {
        conn_map.insert(cd.vector, *cd);
    }
    let last_pos = res[res.len() - 2].id;
    if last_pos != nb {
        panic!("Calculating connection details failed since {last_pos} != {nb}");
    }
    (res, conn_map)
}

fn add_conn_detail(conn_map: &mut HashMap<Point, ConnectionDetails>, conn_vector: Point) {
    let ds = conn_vector.distance_squared();
    if ds == 0 {
        panic!("zero vector cannot be a connection");
    }
    if !(ds == 1 || ds == 2 || ds == 3 || ds == 5) {
        panic!("vector {conn_vector} of ds={ds} cannot be a connection");
    }
    if conn_map.contains_key(&conn_vector) {
        return;
    }
    let neg = conn_vector.neg();
    conn_map.insert(
        conn_vector,
        ConnectionDetails {
            id: ConnectionId::NIL,
            vector: conn_vector,
            conn_ds: ds,
        },
    );
    conn_map.insert(
        neg,
        ConnectionDetails {
            id: ConnectionId::NIL,
            vector: neg,
            conn_ds: ds,
        },
    );
}

fn calculate_trios(
    base: &[BaseTrio; 8],
    conn_map: &HashMap<Point, ConnectionDetails>,
) -> Vec<TrioDetails> {
    let mut res: Vec<TrioDetails> = Vec::with_capacity(200);
    for (i, tr) in base.iter().enumerate() {
        let mut td = make_trio_details(conn_map, [tr.0[0], tr.0[1], tr.0[2]]);
        td.id = TrioIndex(i as u8);
        add_unique_trio(&mut res, td);
    }
    for t_a in base {
        for t_b in base {
            for t_c in base {
                for next in get_next_trios(conn_map, *t_a, *t_b, *t_c) {
                    add_unique_trio(&mut res, next);
                }
            }
        }
    }
    res.sort_by(|a, b| cmp_trio(a, b, base));
    for (i, td) in res.iter_mut().enumerate() {
        let tr_idx = TrioIndex(i as u8);
        if i < base.len() {
            if td.id != tr_idx {
                panic!("incorrect Id for base trio details {td} at {i}");
            }
        } else if td.id != TrioIndex::NIL {
            panic!("incorrect Id for non base trio details {td} at {i}");
        } else {
            td.id = tr_idx;
        }
    }
    res
}

fn make_trio_details(
    conn_map: &HashMap<Point, ConnectionDetails>,
    points: [Point; 3],
) -> TrioDetails {
    let mut cds = [EMPTY_CONN; 3];
    for (i, p) in points.iter().enumerate() {
        cds[i] = *conn_map
            .get(p)
            .unwrap_or_else(|| panic!("trying to create trio with vector not a connection {p}"));
    }
    sort_three_conns(&mut cds);
    TrioDetails {
        id: TrioIndex::NIL,
        conns: cds,
    }
}

fn vectors_of(td: &TrioDetails) -> [Point; 3] {
    [td.conns[0].vector, td.conns[1].vector, td.conns[2].vector]
}

fn add_unique_trio(list: &mut Vec<TrioDetails>, td: TrioDetails) {
    let key = vectors_of(&td);
    if list.iter().any(|existing| vectors_of(existing) == key) {
        return;
    }
    list.push(td);
}

fn cmp_trio(tr1: &TrioDetails, tr2: &TrioDetails, base: &[BaseTrio; 8]) -> Ordering {
    let ds1 = tr1.ds_index();
    let ds2 = tr2.ds_index();
    if ds1 != ds2 {
        return ds1.cmp(&ds2);
    }
    if ds1 == 0 {
        let k = base_slot(tr1, base);
        let l = base_slot(tr2, base);
        return k.cmp(&l);
    }
    for k in 0..3 {
        let cd1 = &tr1.conns[k];
        let cd2 = &tr2.conns[k];
        if cd1.id != cd2.id {
            return if is_less_conn_id(cd1, cd2) {
                Ordering::Less
            } else {
                Ordering::Greater
            };
        }
    }
    Ordering::Equal
}

fn base_slot(tr: &TrioDetails, base: &[BaseTrio; 8]) -> usize {
    let key = vectors_of(tr);
    for (bi, bt) in base.iter().enumerate() {
        if bt.0 == key {
            return bi;
        }
    }
    0
}

fn connector(from: Point, axis: usize, trio: BaseTrio) -> Point {
    let first = match axis {
        0 => X_FIRST,
        1 => Y_FIRST,
        2 => Z_FIRST,
        _ => unreachable!(),
    };
    if from.0[axis] > 0 {
        first.add(trio.axis_vector(axis, -1)).sub(from)
    } else {
        first.neg().add(trio.axis_vector(axis, 1)).sub(from)
    }
}

fn get_next_trios(
    conn_map: &HashMap<Point, ConnectionDetails>,
    t_a: BaseTrio,
    t_b: BaseTrio,
    t_c: BaseTrio,
) -> Vec<TrioDetails> {
    let mut res = Vec::new();
    let same = t_b == t_c;

    let no_z = t_a.0[0];
    let x_b = connector(no_z, 0, t_b);
    let y_b = connector(no_z, 1, t_b);
    if same {
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_z.neg(), x_b, y_b]),
        );
    } else {
        let x_c = connector(no_z, 0, t_c);
        let y_c = connector(no_z, 1, t_c);
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_z.neg(), x_b, y_c]),
        );
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_z.neg(), x_c, y_b]),
        );
    }

    let no_y = t_a.0[1];
    let x_b = connector(no_y, 0, t_b);
    let z_b = connector(no_y, 2, t_b);
    if same {
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_y.neg(), x_b, z_b]),
        );
    } else {
        let x_c = connector(no_y, 0, t_c);
        let z_c = connector(no_y, 2, t_c);
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_y.neg(), x_b, z_c]),
        );
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_y.neg(), x_c, z_b]),
        );
    }

    let no_x = t_a.0[2];
    let y_b = connector(no_x, 1, t_b);
    let z_b = connector(no_x, 2, t_b);
    if same {
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_x.neg(), y_b, z_b]),
        );
    } else {
        let y_c = connector(no_x, 1, t_c);
        let z_c = connector(no_x, 2, t_c);
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_x.neg(), y_b, z_c]),
        );
        add_unique_trio(
            &mut res,
            make_trio_details(conn_map, [no_x.neg(), y_c, z_b]),
        );
    }
    res
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::conn::UnitDirection;
    use crate::growth::ALL_GROWTH_TYPES;
    use crate::point::{CInt, ORIGIN};
    use crate::trio::possible_ds_arrays;

    fn pack() -> &'static PointPackData {
        PointPackData::shared()
    }

    #[test]
    fn connection_details_count_and_opposites() {
        let ppd = pack();
        assert_eq!(ppd.connections.len(), EXPECTED_CONNECTIONS);
        assert_eq!(ppd.connections_by_vector.len(), EXPECTED_CONNECTIONS);
        for (k, v) in &ppd.connections_by_vector {
            assert_eq!(*k, v.vector);
            assert_eq!(k.distance_squared(), v.distance_squared());
            let current = v.pos_id();
            let mut same = 0;
            for nv in ppd.connections_by_vector.values() {
                if nv.pos_id() == current {
                    same += 1;
                    if nv.vector != v.vector {
                        assert_eq!(nv.id, v.id.neg());
                        assert_eq!(nv.vector, v.vector.neg());
                    }
                }
            }
            assert_eq!(same, 2, "conn {}", current.0);
        }
        for (i, t_a) in ppd.base_trios.iter().enumerate() {
            for (j, t_b) in ppd.base_trios.iter().enumerate() {
                for (k, conn_vector) in get_non_base_connections(*t_a, *t_b).iter().enumerate() {
                    let details = ppd.conn_by_vector(*conn_vector);
                    assert!(
                        details.is_valid(),
                        "missing connection between trios ({i},{j}) slot {k} vector {conn_vector}"
                    );
                    assert_eq!(details.vector, *conn_vector);
                }
            }
        }
    }

    #[test]
    fn all_trio_details_shape() {
        let ppd = pack();
        let all = &ppd.trios;
        assert_eq!(all.len(), EXPECTED_TRIOS);
        for (i, td) in all.iter().enumerate() {
            let cds = td.conns;
            assert!(cds[0].pos_id().0 <= cds[1].pos_id().0, "{td}");
            assert!(cds[1].pos_id().0 <= cds[2].pos_id().0, "{td}");
            assert_eq!(td.is_base_trio(), i < 8, "{td}");
            if i < 8 {
                for c in &cds {
                    assert!(c.is_base(), "{c} in {td}");
                }
                for ud in UnitDirection::all() {
                    let _ = ud.find_connection(td);
                    let _ = ud.opposite().find_connection(td);
                }
            }
            if i > 0 {
                assert!(all[i - 1].ds_index() <= td.ds_index());
            }
        }
    }

    #[test]
    fn trio_details_per_ds_index() {
        let ppd = pack();
        let possible = possible_ds_arrays();
        let mut index_in_poss = vec![0i32; ppd.trios.len()];
        for (i, td) in ppd.trios.iter().enumerate() {
            let ds_array = [
                td.conns[0].conn_ds,
                td.conns[1].conn_ds,
                td.conns[2].conn_ds,
            ];
            let mut found = None;
            for (k, pos) in possible.iter().enumerate() {
                if *pos == ds_array {
                    found = Some(k);
                    index_in_poss[i] = k as i32;
                }
            }
            let found = found.expect("ds array not in the legal list");
            assert_eq!(found as i32, td.ds_index(), "{td}");
        }
        let mut count = [0i32; 7];
        let mut by_first = [[0i32; 10]; 7];
        for (i, td) in ppd.trios.iter().enumerate() {
            if i > 0 {
                assert!(index_in_poss[i - 1] <= index_in_poss[i]);
            }
            let ds_index = td.ds_index() as usize;
            count[ds_index] += 1;
            by_first[ds_index][td.conns[0].pos_id().0 as usize] += 1;
        }
        assert_eq!(count[0], 8);
        assert_eq!(count[1], 3 * 2 * 2);
        assert_eq!(count[2], 3 * 8 * 2);
        assert_eq!(count[3], 3 * 4 * 2);
        assert_eq!(count[4], 3 * 8 * 2);
        assert_eq!(count[5], 3 * 8 * 2);
        assert_eq!(count[6], 3 * 2 * 2);
        for (i, v) in by_first[0].iter().enumerate() {
            if i == 4 || i == 5 {
                assert_eq!(*v, 4, "index 0 conn {i}");
            } else {
                assert_eq!(*v, 0, "index 0 conn {i}");
            }
        }
        for (i, v) in by_first[1].iter().enumerate() {
            if i == 1 {
                assert_eq!(*v, 8);
            } else if i == 2 {
                assert_eq!(*v, 4);
            } else {
                assert_eq!(*v, 0);
            }
        }
        for (i, v) in by_first[2].iter().enumerate() {
            if i == 1 || i == 2 || i == 3 {
                assert_eq!(*v, 16, "index 2 conn {i}");
            } else {
                assert_eq!(*v, 0);
            }
        }
        for (i, v) in by_first[3].iter().enumerate() {
            if i == 1 || i == 2 || i == 3 {
                assert_eq!(*v, 8, "index 3 conn {i}");
            } else {
                assert_eq!(*v, 0);
            }
        }
        for (i, v) in by_first[4].iter().enumerate() {
            if (4..=9).contains(&i) {
                assert_eq!(*v, 8, "index 4 conn {i}");
            } else {
                assert_eq!(*v, 0);
            }
        }
        for (i, v) in by_first[5].iter().enumerate() {
            if (4..=9).contains(&i) {
                assert_eq!(*v, 8, "index 5 conn {i}");
            } else {
                assert_eq!(*v, 0);
            }
        }
        for (i, v) in by_first[6].iter().enumerate() {
            if (4..=9).contains(&i) {
                assert_eq!(*v, 2, "index 6 conn {i}");
            } else {
                assert_eq!(*v, 0);
            }
        }
    }

    #[test]
    fn trio_connection_methods() {
        let ppd = pack();
        let td0 = ppd.trio(TrioIndex(0));
        assert!(td0.has_connection(ConnectionId(4)));
        assert!(!td0.has_connection(ConnectionId(-4)));
        assert!(td0.has_connection(ConnectionId(-6)));
        assert!(!td0.has_connection(ConnectionId(6)));
        assert!(td0.has_connection(ConnectionId(-9)));
        assert!(!td0.has_connection(ConnectionId(9)));
        let failed = td0.other_connections_from(ConnectionId(-4));
        assert!(failed[0].is_none() && failed[1].is_none());
        let oc = td0.other_connections_from(ConnectionId(4));
        assert_eq!(oc[0].unwrap(), ppd.conn_by_id(ConnectionId(-6)));
        assert_eq!(oc[1].unwrap(), ppd.conn_by_id(ConnectionId(-9)));

        let td92 = ppd.trio(TrioIndex(92));
        assert!(td92.has_connection(ConnectionId(4)));
        assert!(!td92.has_connection(ConnectionId(-4)));
        assert!(td92.has_connection(ConnectionId(12)));
        assert!(td92.has_connection(ConnectionId(-12)));
        let oc = td92.other_connections_from(ConnectionId(4));
        assert_eq!(oc[0].unwrap(), ppd.conn_by_id(ConnectionId(12)));
        assert_eq!(oc[1].unwrap(), ppd.conn_by_id(ConnectionId(-12)));
    }

    #[test]
    fn growth_context_neighbors_stay_connected() {
        let ppd = pack();
        assert_eq!(ppd.growth_contexts.len(), EXPECTED_GROWTH_CONTEXTS);
        let min = CInt::from(-5);
        let max = CInt::from(5);
        for ctx_type in ALL_GROWTH_TYPES {
            for p_idx in 0..ctx_type.nb_indexes() {
                let ctx = ppd.growth_by_type_and_index(ctx_type, p_idx);
                for x in min..max {
                    for y in min..max {
                        for z in min..max {
                            let main = Point([x, y, z]).mul(3);
                            let conns = ppd.trio_for_main(ctx, main, 0).conns;
                            for conn in conns {
                                let c_vec = conn.vector;
                                assert_valid(ppd, main, main.add(c_vec));
                                check_axis(ppd, ctx, main, c_vec, 0);
                                check_axis(ppd, ctx, main, c_vec, 1);
                                check_axis(ppd, ctx, main, c_vec, 2);
                            }
                        }
                    }
                }
            }
        }
    }

    fn check_axis(
        ppd: &PointPackData,
        ctx: &GrowthContext,
        main: Point,
        c_vec: Point,
        axis: usize,
    ) {
        let step = match axis {
            0 => X_FIRST,
            1 => Y_FIRST,
            _ => Z_FIRST,
        };
        let next_main = match c_vec.0[axis] {
            0 => return,
            1 => main.add(step),
            -1 => main.sub(step),
            other => panic!("connecting vector axis {axis} is {other}"),
        };
        if next_main == ORIGIN && main != ORIGIN.sub(step) && c_vec.0[axis] != 0 {
            // Origin is a real main point. Only skip when the switch left it unset,
            // which this match does not do.
        }
        let next_conns = ppd.trio_for_main(ctx, next_main, 0).conns;
        for conn in next_conns {
            if conn.vector.0[axis] == -c_vec.0[axis] {
                assert_valid(ppd, main.add(c_vec), next_main.add(conn.vector));
            }
        }
    }

    fn assert_valid(ppd: &PointPackData, p1: Point, p2: Point) {
        let c1 = ppd.conn_by_points(p1, p2);
        assert_ne!(c1, EMPTY_CONN, "{p1} -> {p2}");
        assert_eq!(c1.vector, make_vector(p1, p2));
        let c2 = ppd.conn_by_points(p2, p1);
        assert_ne!(c2, EMPTY_CONN);
        assert_eq!(c2.vector, make_vector(p2, p1));
    }
}
