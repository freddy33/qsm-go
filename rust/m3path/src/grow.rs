//! In-memory path growth, one context at a time.
//!
//! This follows `PathContextDb.calculateNextMaxDist` and `makeNewNodes` with
//! `nbParallelProcesses = 1`. Open nodes at the current distance are walked in
//! point order so ids assigned by the caller stay stable across a resume.
//! Go map iteration is random; counts and geometry do not depend on that order,
//! but which parent wins a contested `setFrom` can. Point order makes the winner
//! deterministic.

use std::collections::BTreeMap;

use m3point::{ConnectionId, NodeHandle, Point, PointPackData, TrioIndex, ORIGIN};
use m3util::QsmError;

use crate::{
    ConnectionState, ConnectionsState, DEAD_END_ID, LINK_ID_NOT_SET, NB_CONNECTIONS,
    NEXT_LINK_ID_NOT_ASSIGNED,
};

/// One path node while it is growing. `id` is 0 until the simulation assigns it.
#[derive(Clone, Debug)]
pub struct LivePathNode {
    pub id: i64,
    pub point: Point,
    pub d: i32,
    pub builder: NodeHandle,
    pub connections: ConnectionsState,
    /// Index into [`PathGrowth::nodes`] for a resolved neighbour, when known.
    pub link_at: [Option<usize>; NB_CONNECTIONS],
    pub dirty: bool,
}

/// A path context and every node grown from its root.
#[derive(Clone, Debug)]
pub struct PathGrowth {
    pub id: i32,
    pub growth_ctx_id: i32,
    pub growth_offset: i32,
    pub nodes: Vec<LivePathNode>,
    by_point: BTreeMap<Point, usize>,
}

impl PathGrowth {
    /// Root node at the origin. The path builder is the cube for this growth context and offset.
    pub fn new_root(
        pack: &PointPackData,
        id: i32,
        growth_ctx_id: i32,
        growth_offset: i32,
    ) -> Result<Self, QsmError> {
        let ctx = *pack.growth_by_id(growth_ctx_id);
        let builder = pack.path_node_builder(&ctx, growth_offset, ORIGIN);
        let mut growth = Self {
            id,
            growth_ctx_id,
            growth_offset,
            nodes: Vec::new(),
            by_point: BTreeMap::new(),
        };
        growth.push_node(pack, ORIGIN, 0, builder);
        Ok(growth)
    }

    /// Rebuild a context from nodes that were already saved. `nodes` keep their ids.
    pub fn from_stored(
        id: i32,
        growth_ctx_id: i32,
        growth_offset: i32,
        nodes: Vec<LivePathNode>,
    ) -> Result<Self, QsmError> {
        if nodes.is_empty() {
            return Err(QsmError::new(format!(
                "path context {id} has no nodes to resume"
            )));
        }
        let mut by_point = BTreeMap::new();
        for (index, node) in nodes.iter().enumerate() {
            if by_point.insert(node.point, index).is_some() {
                return Err(QsmError::new(format!(
                    "path context {id} has two nodes at {}",
                    node.point
                )));
            }
        }
        Ok(Self {
            id,
            growth_ctx_id,
            growth_offset,
            nodes,
            by_point,
        })
    }

    pub fn max_dist(&self) -> i32 {
        self.nodes.iter().map(|node| node.d).max().unwrap_or(0)
    }

    pub fn count_at(&self, distance: i32) -> usize {
        self.nodes.iter().filter(|node| node.d == distance).count()
    }

    /// Grow until `max_dist` reaches `target`, one shell at a time.
    pub fn grow_to(&mut self, pack: &PointPackData, target: i32) -> Result<(), QsmError> {
        while self.max_dist() < target {
            self.step(pack)?;
        }
        Ok(())
    }

    /// One `calculateNextMaxDist`: open the current shell and create the next one.
    pub fn step(&mut self, pack: &PointPackData) -> Result<(), QsmError> {
        let current_d = self.max_dist();
        let next_d = current_d + 1;
        let offset = self.growth_offset;
        let mut open: Vec<usize> = self
            .nodes
            .iter()
            .enumerate()
            .filter(|(_, node)| node.d == current_d && node.connections.has_open_connections())
            .map(|(index, _)| index)
            .collect();
        open.sort_by_key(|&index| self.nodes[index].point);

        for from_idx in open {
            let trio = TrioIndex(self.nodes[from_idx].connections.trio_id);
            if trio == TrioIndex::NIL {
                return Err(QsmError::new(format!(
                    "path node {} has no trio",
                    self.nodes[from_idx].id
                )));
            }
            let details = *pack.trio(trio);
            let builder = self.nodes[from_idx].builder;
            let from_point = self.nodes[from_idx].point;
            for conn in 0..NB_CONNECTIONS {
                if self.nodes[from_idx]
                    .connections
                    .connection_state(conn as i32)
                    != ConnectionState::NotSet
                {
                    continue;
                }
                let connection = details.conns[conn];
                let (next_builder, next_point) =
                    pack.next_builder(builder, from_point, connection.id, offset)?;
                if let Some(existing) = self.by_point.get(&next_point).copied() {
                    if self.nodes[existing].d == current_d {
                        // Same shell: d + 1 != d, so the link is a dead end.
                        self.set_dead_end(from_idx, conn);
                    } else {
                        self.create_connection(pack, from_idx, conn, connection.id, existing);
                    }
                } else {
                    let next_idx = self.push_node(pack, next_point, next_d, next_builder);
                    self.create_connection(pack, from_idx, conn, connection.id, next_idx);
                }
            }
        }
        Ok(())
    }

    /// Copy a newly assigned next-node id into a `Next` link that was still -5.
    pub fn backfill_next_links(&mut self) {
        let ids: Vec<i64> = self.nodes.iter().map(|node| node.id).collect();
        for node in &mut self.nodes {
            for conn in 0..NB_CONNECTIONS {
                if node.connections.connection_state(conn as i32) != ConnectionState::Next {
                    continue;
                }
                if node.connections.link_ids[conn] != NEXT_LINK_ID_NOT_ASSIGNED {
                    continue;
                }
                let Some(next_idx) = node.link_at[conn] else {
                    continue;
                };
                let next_id = ids[next_idx];
                if next_id > 0 {
                    node.connections.link_ids[conn] = next_id;
                    node.dirty = true;
                }
            }
        }
    }

    fn push_node(
        &mut self,
        pack: &PointPackData,
        point: Point,
        distance: i32,
        builder: NodeHandle,
    ) -> usize {
        let index = self.nodes.len();
        self.nodes.push(LivePathNode {
            id: 0,
            point,
            d: distance,
            builder,
            connections: ConnectionsState {
                trio_id: builder.trio_index(pack).0,
                connection_mask: ConnectionState::NotSet as u16,
                link_ids: [LINK_ID_NOT_SET; NB_CONNECTIONS],
            },
            link_at: [None; NB_CONNECTIONS],
            dirty: true,
        });
        self.by_point.insert(point, index);
        index
    }

    fn set_dead_end(&mut self, index: usize, conn: usize) {
        let node = &mut self.nodes[index];
        node.connections
            .set_connection_state(conn as i32, ConnectionState::Blocked);
        node.connections.link_ids[conn] = DEAD_END_ID;
        node.link_at[conn] = None;
        node.dirty = true;
    }

    fn create_connection(
        &mut self,
        pack: &PointPackData,
        from_idx: usize,
        conn: usize,
        conn_id: ConnectionId,
        next_idx: usize,
    ) {
        if self.nodes[next_idx].d != self.nodes[from_idx].d + 1 {
            self.set_dead_end(from_idx, conn);
            return;
        }
        if self.set_from(pack, next_idx, conn_id.neg(), from_idx) {
            let next_id = self.nodes[next_idx].id;
            let link = if next_id <= 0 {
                NEXT_LINK_ID_NOT_ASSIGNED
            } else {
                next_id
            };
            let node = &mut self.nodes[from_idx];
            node.connections
                .set_connection_state(conn as i32, ConnectionState::Next);
            node.connections.link_ids[conn] = link;
            node.link_at[conn] = Some(next_idx);
            node.dirty = true;
        } else {
            self.set_dead_end(from_idx, conn);
        }
    }

    /// `setFrom`: the next node's connection whose id is the negative of the outgoing one.
    fn set_from(
        &mut self,
        pack: &PointPackData,
        node_idx: usize,
        conn_id: ConnectionId,
        from_idx: usize,
    ) -> bool {
        let trio = TrioIndex(self.nodes[node_idx].connections.trio_id);
        let details = *pack.trio(trio);
        let from_id = self.nodes[from_idx].id;
        for (slot, connection) in details.conns.iter().enumerate() {
            if connection.id != conn_id {
                continue;
            }
            if self.nodes[node_idx]
                .connections
                .connection_state(slot as i32)
                != ConnectionState::NotSet
            {
                return false;
            }
            let node = &mut self.nodes[node_idx];
            node.connections
                .set_connection_state(slot as i32, ConnectionState::From);
            node.connections.link_ids[slot] = from_id;
            node.link_at[slot] = Some(from_idx);
            node.dirty = true;
            return true;
        }
        false
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use m3point::GrowthType;

    #[test]
    fn type8_index0_offset0_matches_go_make_new_path_ctx() {
        let pack = PointPackData::shared();
        let ctx = pack.growth_by_type_and_index(GrowthType(8), 0);
        assert_eq!(ctx.id, 40);
        let mut growth = PathGrowth::new_root(pack, 1, ctx.id, 0).unwrap();
        assert_eq!(growth.count_at(0), 1);
        assert!(growth.nodes[0].builder.is_root());
        assert_eq!(growth.nodes[0].builder.cube_id(), 2601);
        assert_eq!(growth.nodes[0].builder.path_builder_index(), 0);
        growth.grow_to(pack, 1).unwrap();
        assert_eq!(growth.count_at(1), 3);
        assert_eq!(growth.max_dist(), 1);
    }

    #[test]
    fn type8_index2_offset1_matches_server_path_counts() {
        let pack = PointPackData::shared();
        let ctx = pack.growth_by_type_and_index(GrowthType(8), 2);
        let mut growth = PathGrowth::new_root(pack, 1, ctx.id, 1).unwrap();
        growth.grow_to(pack, 4).unwrap();
        // backend/m3server/server_path_test.go TestPathContextCreateAndIncrease
        // creates growth type 8, index 2, offset 1 and expects these active counts.
        assert_eq!(growth.count_at(2), 6);
        assert_eq!(growth.count_at(3), 12);
        assert_eq!(growth.count_at(4), 22);
        for node in &growth.nodes {
            let restored = NodeHandle::from_path_builder_index(
                node.builder.cube_id(),
                node.builder.path_builder_index(),
            )
            .unwrap();
            assert_eq!(restored, node.builder);
        }
    }

    #[test]
    fn pyramid_offsets_through_distance_3() {
        let pack = PointPackData::shared();
        let ctx = pack.growth_by_type_and_index(GrowthType(8), 0);
        let mut counts = Vec::new();
        for offset in [0, 2, 4, 6] {
            let mut growth = PathGrowth::new_root(pack, 1, ctx.id, offset).unwrap();
            growth.grow_to(pack, 3).unwrap();
            counts.push((0..=3).map(|d| growth.count_at(d)).collect::<Vec<_>>());
        }
        // Same shell sizes for every playgl pyramid offset. d = 0 and d = 1 match
        // TestMakeNewPathCtx (type 8, index 0, offset 0). d = 2 and d = 3 match the
        // type-8 counts in server_path_test.go (6 and 12), which uses another index.
        assert_eq!(
            counts,
            vec![
                vec![1, 3, 6, 12],
                vec![1, 3, 6, 12],
                vec![1, 3, 6, 12],
                vec![1, 3, 6, 12]
            ]
        );
    }
}
