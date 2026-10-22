//! Path-node builders: root, intermediate, and last nodes inside one trio cube.
//!
//! A path walks three steps between main points. The builder records which
//! connection and which trio each of those steps uses.

use m3util::QsmError;

use crate::conn::{ConnectionDetails, ConnectionId, UnitDirection};
use crate::cube::CubeKey;
use crate::growth::GrowthContext;
use crate::pack::PointPackData;
use crate::point::{Point, ORIGIN};
use crate::trio::TrioIndex;

pub const NB_PATH_BUILDERS_PER_CONTEXT: usize = 10;

#[derive(Clone, Debug)]
pub struct LastBuilder {
    pub tr_idx: TrioIndex,
    pub next_main_conn_id: ConnectionId,
    pub next_inter_conn_id: ConnectionId,
}

#[derive(Clone, Debug)]
pub struct InterLink {
    pub conn_id: ConnectionId,
    pub node: LastBuilder,
}

#[derive(Clone, Debug)]
pub struct IntermediateBuilder {
    pub tr_idx: TrioIndex,
    pub links: [InterLink; 2],
}

#[derive(Clone, Debug)]
pub struct RootLink {
    pub conn_id: ConnectionId,
    pub node: IntermediateBuilder,
}

#[derive(Clone, Debug)]
pub struct RootBuilder {
    pub growth_ctx_id: i32,
    pub cube_id: i32,
    pub tr_idx: TrioIndex,
    pub links: [RootLink; 3],
}

/// Where a builder lives inside its cube. Equality replaces the Go pointer compare.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum NodeHandle {
    Root(i32),
    Intermediate { cube: i32, slot: u8 },
    Last { cube: i32, slot: u8, sub: u8 },
}

impl NodeHandle {
    /// Cube id stored as `path_builders_id`. Intermediate and last nodes keep the root cube.
    pub fn cube_id(self) -> i32 {
        match self {
            NodeHandle::Root(cube)
            | NodeHandle::Intermediate { cube, .. }
            | NodeHandle::Last { cube, .. } => cube,
        }
    }

    /// Index inside one cube, matching `pointdb` `RootPathBuilder` .. `LastPathBuilder32`.
    ///
    /// Root is 0, intermediate slot `s` is `1 + s`, and last `(slot, sub)` is `4 + slot*2 + sub`.
    pub fn path_builder_index(self) -> i32 {
        match self {
            NodeHandle::Root(_) => 0,
            NodeHandle::Intermediate { slot, .. } => 1 + i32::from(slot),
            NodeHandle::Last { slot, sub, .. } => 4 + i32::from(slot) * 2 + i32::from(sub),
        }
    }

    /// Inverse of [`path_builder_index`]. `cube` is the root cube id.
    pub fn from_path_builder_index(cube: i32, index: i32) -> Result<Self, QsmError> {
        match index {
            0 => Ok(NodeHandle::Root(cube)),
            1..=3 => Ok(NodeHandle::Intermediate {
                cube,
                slot: (index - 1) as u8,
            }),
            4..=9 => {
                let rel = index - 4;
                Ok(NodeHandle::Last {
                    cube,
                    slot: (rel / 2) as u8,
                    sub: (rel % 2) as u8,
                })
            }
            _ => Err(QsmError::new(format!(
                "path builder index {index} is outside 0..{NB_PATH_BUILDERS_PER_CONTEXT}"
            ))),
        }
    }

    pub fn trio_index(self, pack: &PointPackData) -> TrioIndex {
        match self {
            NodeHandle::Root(cube) => pack.root(cube).tr_idx,
            NodeHandle::Intermediate { cube, slot } => pack.intermediate(cube, slot).tr_idx,
            NodeHandle::Last { cube, slot, sub } => pack.last(cube, slot, sub).tr_idx,
        }
    }

    pub fn is_root(self) -> bool {
        matches!(self, NodeHandle::Root(_))
    }

    pub fn is_intermediate(self) -> bool {
        matches!(self, NodeHandle::Intermediate { .. })
    }

    pub fn is_last(self) -> bool {
        matches!(self, NodeHandle::Last { .. })
    }
}

struct NextMain {
    ud: UnitDirection,
    lip: Point,
    back_conn: ConnectionDetails,
    next_main_conn: ConnectionId,
    next_inter_conn: ConnectionId,
    tr_idx: TrioIndex,
}

pub fn calculate_all_path_builders(pack: &PointPackData) -> Vec<Option<RootBuilder>> {
    let mut res = Vec::with_capacity(pack.cubes_by_id.len());
    res.push(None);
    for cube_id in 1..pack.cubes_by_id.len() {
        let key = pack.cubes_by_id[cube_id];
        let ctx = *pack.growth_by_id(key.growth_ctx_id);
        res.push(Some(populate(pack, cube_id as i32, &ctx, key)));
    }
    res
}

fn populate(pack: &PointPackData, cube_id: i32, ctx: &GrowthContext, key: CubeKey) -> RootBuilder {
    let cube = key.cube;
    let tr_idx = cube.center;
    let td = *pack.trio(tr_idx);
    let mut links = Vec::with_capacity(3);
    for cd in &td.conns {
        let ip = cd.vector;
        let dirs = cd.directions();
        let mut next_mains = [empty_next(dirs[0]), empty_next(dirs[1])];
        for (j, ud) in dirs.iter().enumerate() {
            let nmp = ud.first_point();
            let next_td = pack.trio(cube.center_face(*ud));
            let back_conn = ud.opposite().find_connection(next_td);
            next_mains[j] = NextMain {
                ud: *ud,
                lip: nmp.add(back_conn.vector),
                back_conn,
                next_main_conn: back_conn.neg_id(),
                next_inter_conn: ConnectionId::NIL,
                tr_idx: TrioIndex::NIL,
            };
        }
        let ip_conns = [
            pack.conn_by_points(ip, next_mains[0].lip),
            pack.conn_by_points(ip, next_mains[1].lip),
        ];
        let i_td = pack
            .trios
            .iter()
            .find(|poss| poss.has_connections(&[cd.neg_id(), ip_conns[0].id, ip_conns[1].id]))
            .unwrap_or_else(|| {
                panic!(
                    "did not find trio matching {} {} {} in {ctx} cube center {}",
                    cd.neg_id(),
                    ip_conns[0].id,
                    ip_conns[1].id,
                    cube.center
                )
            });
        let mut inter_links = Vec::with_capacity(2);
        for j in 0..2 {
            let back_uds = next_mains[j].back_conn.directions();
            let mut found_ud = false;
            for back_ud in back_uds {
                if back_ud.opposite() == next_mains[j].ud {
                    found_ud = true;
                } else {
                    let next_inter_td = pack.trio(cube.middle_edge(next_mains[j].ud, back_ud));
                    let next_inter_back = back_ud.opposite().find_connection(next_inter_td);
                    let next_inter_near = next_mains[j]
                        .ud
                        .first_point()
                        .add(back_ud.first_point())
                        .add(next_inter_back.vector);
                    let lip_to_other = pack.conn_by_points(next_mains[j].lip, next_inter_near);
                    next_mains[j].next_inter_conn = lip_to_other.id;
                    let wanted = [
                        ip_conns[j].neg_id(),
                        next_mains[j].next_inter_conn,
                        next_mains[j].next_main_conn,
                    ];
                    let li_td = pack
                        .trios
                        .iter()
                        .find(|poss| poss.has_connections(&wanted))
                        .unwrap_or_else(|| {
                            panic!(
                                "did not find last trio matching {} {} {} in {ctx}",
                                wanted[0], wanted[1], wanted[2]
                            )
                        });
                    next_mains[j].tr_idx = li_td.id;
                }
            }
            if !found_ud {
                panic!(
                    "direction mess between {td} and {i_td} for {:?}",
                    next_mains[j].ud
                );
            }
            inter_links.push(InterLink {
                conn_id: ip_conns[j].id,
                node: LastBuilder {
                    tr_idx: next_mains[j].tr_idx,
                    next_main_conn_id: next_mains[j].next_main_conn,
                    next_inter_conn_id: next_mains[j].next_inter_conn,
                },
            });
        }
        let inter_links: [InterLink; 2] = inter_links.try_into().expect("two intermediate links");
        links.push(RootLink {
            conn_id: cd.id,
            node: IntermediateBuilder {
                tr_idx: i_td.id,
                links: inter_links,
            },
        });
    }
    let links: [RootLink; 3] = links.try_into().expect("three root links");
    let root = RootBuilder {
        growth_ctx_id: ctx.id,
        cube_id,
        tr_idx,
        links,
    };
    verify_root(pack, &root);
    root
}

fn empty_next(ud: UnitDirection) -> NextMain {
    NextMain {
        ud,
        lip: ORIGIN,
        back_conn: ConnectionDetails::default(),
        next_main_conn: ConnectionId::NIL,
        next_inter_conn: ConnectionId::NIL,
        tr_idx: TrioIndex::NIL,
    }
}

fn verify_root(pack: &PointPackData, root: &RootBuilder) {
    let td = pack.trio(root.tr_idx);
    for (link_idx, pl) in root.links.iter().enumerate() {
        if !td.has_connection(pl.conn_id) {
            panic!("{} missing link {link_idx} {}", root.tr_idx, pl.conn_id);
        }
        for (i, other) in root.links.iter().enumerate() {
            if link_idx != i && pl.conn_id == other.conn_id {
                panic!("duplicate root link {}", pl.conn_id);
            }
        }
        let inter = &pl.node;
        let itd = pack.trio(inter.tr_idx);
        for (j, ipl) in inter.links.iter().enumerate() {
            if !itd.has_connection(ipl.conn_id) {
                panic!("intermediate missing {j} {}", ipl.conn_id);
            }
            let last = &ipl.node;
            let ltd = pack.trio(last.tr_idx);
            if !ltd.has_connection(last.next_main_conn_id)
                || !ltd.has_connection(last.next_inter_conn_id)
                || last.next_main_conn_id == last.next_inter_conn_id
            {
                panic!(
                    "last trio {} connections {} {}",
                    last.tr_idx, last.next_main_conn_id, last.next_inter_conn_id
                );
            }
        }
    }
}

impl PointPackData {
    pub fn root(&self, cube: i32) -> &RootBuilder {
        self.path_builders[cube as usize]
            .as_ref()
            .unwrap_or_else(|| panic!("no root builder for cube {cube}"))
    }

    pub fn intermediate(&self, cube: i32, slot: u8) -> &IntermediateBuilder {
        &self.root(cube).links[slot as usize].node
    }

    pub fn last(&self, cube: i32, slot: u8, sub: u8) -> &LastBuilder {
        &self.intermediate(cube, slot).links[sub as usize].node
    }

    pub fn nb_path_builders(&self) -> usize {
        self.path_builders.len()
    }

    pub fn next_builder(
        &self,
        handle: NodeHandle,
        from: Point,
        conn_id: ConnectionId,
        offset: i32,
    ) -> Result<(NodeHandle, Point), QsmError> {
        match handle {
            NodeHandle::Root(cube) => {
                let root = self.root(cube);
                for (slot, pl) in root.links.iter().enumerate() {
                    if pl.conn_id == conn_id {
                        let point = from.add(self.conn_by_id(conn_id).vector);
                        return Ok((
                            NodeHandle::Intermediate {
                                cube,
                                slot: slot as u8,
                            },
                            point,
                        ));
                    }
                }
                Err(QsmError::new(format!(
                    "trying to get next path node builder on connection {conn_id} which does not exist on root cube {cube}"
                )))
            }
            NodeHandle::Intermediate { cube, slot } => {
                let inter = self.intermediate(cube, slot);
                for (sub, pl) in inter.links.iter().enumerate() {
                    if pl.conn_id == conn_id {
                        let point = from.add(self.conn_by_id(conn_id).vector);
                        return Ok((
                            NodeHandle::Last {
                                cube,
                                slot,
                                sub: sub as u8,
                            },
                            point,
                        ));
                    }
                }
                Err(QsmError::new(format!(
                    "trying to get next path node builder on connection {conn_id} which does not exist on intermediate"
                )))
            }
            NodeHandle::Last { cube, slot, sub } => {
                let last = self.last(cube, slot, sub).clone();
                let ctx = *self.growth_by_id(self.root(cube).growth_ctx_id);
                let next_main_point = from.near_main_point();
                let stepped = from.add(self.conn_by_id(last.next_main_conn_id).vector);
                if next_main_point != stepped {
                    return Err(QsmError::new(format!(
                        "last inter main path node does not give a main point using {from} and {}",
                        last.next_main_conn_id
                    )));
                }
                let next_main = self.path_node_builder(&ctx, offset, next_main_point);
                if last.next_main_conn_id == conn_id {
                    return Ok((next_main, next_main_point));
                }
                if last.next_inter_conn_id == conn_id {
                    let (back, o_inter) = self.next_builder(
                        next_main,
                        next_main_point,
                        last.next_main_conn_id.neg(),
                        offset,
                    )?;
                    if from != o_inter {
                        return Err(QsmError::new(format!(
                            "back calculation on last inter path node failed {from} != {o_inter}"
                        )));
                    }
                    return self.next_builder(back, from, conn_id, offset);
                }
                Err(QsmError::new(format!(
                    "trying to get next path node builder on connection {conn_id} which does not exist on last"
                )))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::growth::ALL_GROWTH_TYPES;
    use crate::pack::PointPackData;
    use crate::point::{X_FIRST, Y_FIRST, Z_FIRST};
    use m3util::{pos_mod2, pos_mod4, pos_mod8};

    #[test]
    fn all_path_builders_walk_forward_and_back() {
        let ppd = PointPackData::shared();
        assert_eq!(
            ppd.nb_path_builders(),
            crate::cube::TOTAL_NUMBER_OF_CUBES + 1
        );
        assert_eq!(
            ppd.path_builders.iter().filter(|b| b.is_some()).count(),
            crate::cube::TOTAL_NUMBER_OF_CUBES
        );
        let mut nb_test = 0;
        for ctx_type in ALL_GROWTH_TYPES {
            for p_idx in 0..ctx_type.nb_indexes() {
                let ctx = ppd.growth_by_type_and_index(ctx_type, p_idx);
                for offset in 0..ctx_type.max_offset() {
                    let mut center = ORIGIN;
                    for div in 0..(8 * 3) {
                        nb_test += 1;
                        if div != 0 {
                            center = match div % 3 {
                                0 => center.add(X_FIRST),
                                1 => center.add(Y_FIRST),
                                _ => center.add(Z_FIRST),
                            };
                        }
                        assert_eq!(div, ctx.base_div_by_three(center), "{ctx}");
                        let pnb = ppd.path_node_builder(ctx, offset, center);
                        assert!(pnb.is_root());
                        let trio_idx = pnb.trio_index(ppd);
                        assert_ne!(trio_idx, TrioIndex::NIL);
                        assert!(trio_idx.is_base(), "{trio_idx}");
                        match ctx_type.0 {
                            1 => assert_eq!(trio_idx, TrioIndex(p_idx as u8)),
                            3 => {
                                if pos_mod2(div) == 0 {
                                    assert_eq!(trio_idx, TrioIndex(p_idx as u8));
                                } else {
                                    assert_ne!(trio_idx, TrioIndex(p_idx as u8));
                                }
                            }
                            2 => {
                                let idx = pos_mod2(div + offset as u64) as usize;
                                assert_eq!(trio_idx, ppd.valid_next_trio[p_idx as usize][idx]);
                            }
                            4 => {
                                let idx = pos_mod4(div + offset as u64) as usize;
                                assert_eq!(trio_idx, ppd.mod4[p_idx as usize][idx]);
                            }
                            8 => {
                                let idx = pos_mod8(div + offset as u64) as usize;
                                assert_eq!(trio_idx, ppd.mod8[p_idx as usize][idx]);
                            }
                            _ => panic!("unexpected type"),
                        }
                        let td = *ppd.trio(trio_idx);
                        let NodeHandle::Root(cube) = pnb else {
                            unreachable!()
                        };
                        let root = ppd.root(cube);
                        let mut nb_intermediate = 0;
                        let mut nb_last = 0;
                        for (i, cd) in td.conns.iter().enumerate() {
                            assert_eq!(cd.id, root.links[i].conn_id);
                            let (npnb, np) = ppd
                                .next_builder(pnb, center, cd.id, offset)
                                .expect("next from root");
                            assert!(npnb.is_intermediate());
                            assert!(!npnb.trio_index(ppd).is_base());
                            assert_eq!(np, center.add(cd.vector));
                            nb_intermediate += 1;
                            let ntd = *ppd.trio(npnb.trio_index(ppd));
                            let NodeHandle::Intermediate { slot, .. } = npnb else {
                                unreachable!()
                            };
                            let inter = ppd.intermediate(cube, slot);
                            for ncd in ntd.conns {
                                if ncd.neg_id() == cd.id {
                                    continue;
                                }
                                let mut found_last = None;
                                for (sub, ipl) in inter.links.iter().enumerate() {
                                    if ipl.conn_id == ncd.id {
                                        let handle = NodeHandle::Last {
                                            cube,
                                            slot,
                                            sub: sub as u8,
                                        };
                                        assert!(!handle.trio_index(ppd).is_base());
                                        found_last = Some(handle);
                                    }
                                }
                                let lipnb = found_last.expect("inter connection present");
                                let (last_handle, lip) = ppd
                                    .next_builder(npnb, np, ncd.id, offset)
                                    .expect("next from intermediate");
                                assert_eq!(last_handle, lipnb);
                                assert_eq!(lip, np.add(ncd.vector));
                                nb_last += 1;
                                let NodeHandle::Last { sub, .. } = lipnb else {
                                    unreachable!()
                                };
                                let last = ppd.last(cube, slot, sub);
                                let (next_main, nmp) = ppd
                                    .next_builder(last_handle, lip, last.next_main_conn_id, offset)
                                    .expect("next main");
                                assert!(next_main.is_root());
                                assert!(nmp.is_main_point());
                                assert!(next_main.trio_index(ppd).is_base());
                                assert_eq!(lip.near_main_point(), nmp);
                                let (back, o_lip) = ppd
                                    .next_builder(
                                        next_main,
                                        nmp,
                                        last.next_main_conn_id.neg(),
                                        offset,
                                    )
                                    .expect("back to intermediate");
                                assert_eq!(o_lip, lip);
                                assert_eq!(back.trio_index(ppd), last_handle.trio_index(ppd));
                                let (next_inter, nlip) = ppd
                                    .next_builder(last_handle, lip, last.next_inter_conn_id, offset)
                                    .expect("next inter");
                                assert!(next_inter.is_last());
                                assert_ne!(nlip.near_main_point(), nmp);
                            }
                        }
                        assert_eq!(nb_intermediate, 3);
                        assert_eq!(nb_last, 6);
                    }
                }
            }
        }
        assert_eq!(nb_test, 4800);
    }
}
