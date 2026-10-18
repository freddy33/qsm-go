//! Growth contexts select which base trio sits on each main point.

use std::fmt;

use crate::point::{abs_dint_from_c, Point, THREE};
use crate::trio::TrioIndex;
use m3util::{pos_mod2, pos_mod4, pos_mod8};

/// How outgrowth walks from one main point to the next.
///
/// The five live types are 1, 2, 3, 4 and 8. Type 0 is reserved in the Go comments.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct GrowthType(pub u8);

pub const ALL_GROWTH_TYPES: [GrowthType; 5] = [
    GrowthType(1),
    GrowthType(2),
    GrowthType(3),
    GrowthType(4),
    GrowthType(8),
];

pub const TOTAL_NB_CONTEXTS: usize = 8 + 12 + 8 + 12 + 12;

impl GrowthType {
    pub fn is_permutation(self) -> bool {
        self.0 == 2 || self.0 == 4 || self.0 == 8
    }

    pub fn modulo(self) -> i32 {
        i32::from(self.0)
    }

    pub fn nb_indexes(self) -> i32 {
        if self.is_permutation() {
            12
        } else {
            8
        }
    }

    pub fn max_offset(self) -> i32 {
        match self.0 {
            1 => 1,
            3 => 3,
            2 => 2,
            4 => 4,
            8 => 8,
            _ => 0,
        }
    }
}

impl fmt::Display for GrowthType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "CtxType{}", self.0)
    }
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct GrowthContext {
    pub id: i32,
    pub growth_type: GrowthType,
    pub growth_index: i32,
}

impl GrowthContext {
    pub fn base_div_by_three(self, main_point: Point) -> u64 {
        if !main_point.is_main_point() {
            panic!(
                "cannot ask for trio index on non nextMainPoint Pos {main_point} in context {self}!"
            );
        }
        (abs_dint_from_c(main_point.0[0]) / D3
            + abs_dint_from_c(main_point.0[1]) / D3
            + abs_dint_from_c(main_point.0[2]) / D3) as u64
    }
}

const D3: i64 = THREE as i64;

impl fmt::Display for GrowthContext {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "GrowthCtx{}-T{}-Idx{:02}",
            self.id, self.growth_type.0, self.growth_index
        )
    }
}

/// Trio index on a main point for this context, division, and offset.
///
/// `valid_next`, `mod4` and `mod8` are the tables built by [`crate::perms`].
pub fn base_trio_index(
    ctx: &GrowthContext,
    div_by_three: u64,
    offset: i32,
    valid_next: &[[TrioIndex; 2]; 12],
    mod4: &[[TrioIndex; 4]; 12],
    mod8: &[[TrioIndex; 8]; 12],
) -> TrioIndex {
    let ctx_tr = TrioIndex(ctx.growth_index as u8);
    if ctx.growth_type.0 == 1 {
        return ctx_tr;
    }
    if ctx.growth_type.0 == 3 {
        if pos_mod2(div_by_three) == 0 {
            return ctx_tr;
        }
        let mod3 = (((div_by_three - 1) / 2 + offset as u64) % 3) as i32;
        if ctx.growth_index < 4 {
            return valid_next[(3 * ctx.growth_index + mod3) as usize][1];
        }
        let mut count = 0;
        for valid in valid_next {
            if valid[1] == ctx_tr {
                if count == mod3 {
                    return valid[0];
                }
                count += 1;
            }
        }
        panic!(
            "did not find valid trio for div by three value {div_by_three} in context {ctx}-{offset}!"
        );
    }

    let div_off = offset as u64 + div_by_three;
    match ctx.growth_type.0 {
        2 => {
            let idx = pos_mod2(div_off) as usize;
            valid_next[ctx.growth_index as usize][idx]
        }
        4 => {
            let idx = pos_mod4(div_off) as usize;
            mod4[ctx.growth_index as usize][idx]
        }
        8 => {
            let idx = pos_mod8(div_off) as usize;
            mod8[ctx.growth_index as usize][idx]
        }
        _ => panic!(
            "event permutation type {} in context {ctx}-{offset} is invalid!",
            ctx.growth_index
        ),
    }
}

pub fn all_growth_contexts() -> Vec<GrowthContext> {
    let mut res = Vec::with_capacity(TOTAL_NB_CONTEXTS);
    let mut idx = 0;
    for ctx_type in ALL_GROWTH_TYPES {
        let nb = ctx_type.nb_indexes();
        for p_idx in 0..nb {
            res.push(GrowthContext {
                id: idx,
                growth_type: ctx_type,
                growth_index: p_idx,
            });
            idx += 1;
        }
    }
    res
}
