//! Integer lattice for the Quantum State Machine.
//!
//! Points, connection vectors, trios, growth contexts, trio cubes, and path
//! builders are rebuilt in process. The numbers match the Go `m3point` and
//! `pointdb` packages: 50 connections, 200 trios, 52 growth contexts, 5192 cubes.

pub mod conn;
pub mod cube;
pub mod growth;
pub mod murmur;
pub mod pack;
pub mod pathbuilder;
pub mod perms;
pub mod point;
pub mod pointmap;
pub mod trio;

pub use conn::{ConnectionDetails, ConnectionId, UnitDirection, EMPTY_CONN};
pub use cube::{middle_edge_index, CubeKey, CubeOfTrioIndex, TOTAL_NUMBER_OF_CUBES};
pub use growth::{
    all_growth_contexts, base_trio_index, GrowthContext, GrowthType, ALL_GROWTH_TYPES,
    TOTAL_NB_CONTEXTS,
};
pub use murmur::{murmur_hash_to_int, murmur_hash_u32, murmur_i32, murmur_i64};
pub use pack::{
    PointPackData, EXPECTED_CONNECTIONS, EXPECTED_CUBES, EXPECTED_GROWTH_CONTEXTS, EXPECTED_TRIOS,
};
pub use pathbuilder::{
    InterLink, IntermediateBuilder, LastBuilder, NodeHandle, RootBuilder, RootLink,
    NB_PATH_BUILDERS_PER_CONTEXT,
};
pub use perms::{
    all_base_trios, get_non_base_connections, init_mod4_permutations, init_mod8_permutations,
    init_valid_trios, is_prime, same_permutation, BaseTrio,
};
pub use point::{
    abs_cint, abs_dint, abs_dint_from_c, create_random_point, ds, make_vector, random_cint, CInt,
    DInt, Point, ORIGIN, THREE, X_FIRST, Y_FIRST, Z_FIRST,
};
pub use pointmap::PointHashMap;
pub use trio::{TrioDetails, TrioIndex, NB_TRIO_DS_INDEX};
