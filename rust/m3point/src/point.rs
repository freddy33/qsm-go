//! Integer 3D points. Every operation is integer arithmetic, matching `m3point.Point`.

use std::fmt;

use rand::Rng;

pub type CInt = i32;
pub type DInt = i64;

pub const THREE: CInt = 3;

/// A location in the integer lattice. The Go type is `[3]CInt`.
#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash, Debug, Default)]
pub struct Point(pub [CInt; 3]);

pub const ORIGIN: Point = Point([0, 0, 0]);
pub const X_FIRST: Point = Point([THREE, 0, 0]);
pub const Y_FIRST: Point = Point([0, THREE, 0]);
pub const Z_FIRST: Point = Point([0, 0, THREE]);

#[inline]
pub fn abs_cint(i: CInt) -> CInt {
    if i < 0 {
        -i
    } else {
        i
    }
}

#[inline]
pub fn abs_dint(i: DInt) -> DInt {
    if i < 0 {
        -i
    } else {
        i
    }
}

#[inline]
pub fn abs_dint_from_c(i: CInt) -> DInt {
    if i < 0 {
        DInt::from(-i)
    } else {
        DInt::from(i)
    }
}

/// Squared Euclidean distance between two points.
pub fn ds(p1: Point, p2: Point) -> DInt {
    let x = DInt::from(p2.x() - p1.x());
    let y = DInt::from(p2.y() - p1.y());
    let z = DInt::from(p2.z() - p1.z());
    x * x + y * y + z * z
}

impl Point {
    #[inline]
    pub fn x(self) -> CInt {
        self.0[0]
    }
    #[inline]
    pub fn y(self) -> CInt {
        self.0[1]
    }
    #[inline]
    pub fn z(self) -> CInt {
        self.0[2]
    }

    pub fn mul(self, m: CInt) -> Self {
        Point([self.0[0] * m, self.0[1] * m, self.0[2] * m])
    }

    pub fn add(self, other: Point) -> Self {
        Point([
            self.0[0] + other.0[0],
            self.0[1] + other.0[1],
            self.0[2] + other.0[2],
        ])
    }

    pub fn sub(self, other: Point) -> Self {
        Point([
            self.0[0] - other.0[0],
            self.0[1] - other.0[1],
            self.0[2] - other.0[2],
        ])
    }

    pub fn neg(self) -> Self {
        Point([-self.0[0], -self.0[1], -self.0[2]])
    }

    /// Positive quarter turn around X: `(x, y, z) -> (x, -z, y)`.
    pub fn rot_plus_x(self) -> Self {
        Point([self.0[0], -self.0[2], self.0[1]])
    }
    pub fn rot_neg_x(self) -> Self {
        Point([self.0[0], self.0[2], -self.0[1]])
    }
    pub fn rot_plus_y(self) -> Self {
        Point([self.0[2], self.0[1], -self.0[0]])
    }
    pub fn rot_neg_y(self) -> Self {
        Point([-self.0[2], self.0[1], self.0[0]])
    }
    pub fn rot_plus_z(self) -> Self {
        Point([-self.0[1], self.0[0], self.0[2]])
    }
    pub fn rot_neg_z(self) -> Self {
        Point([self.0[1], -self.0[0], self.0[2]])
    }

    pub fn distance_squared(self) -> DInt {
        let x = DInt::from(self.0[0]);
        let y = DInt::from(self.0[1]);
        let z = DInt::from(self.0[2]);
        x * x + y * y + z * z
    }

    pub fn is_main_point(self) -> bool {
        self.0.iter().all(|c| c % THREE == 0)
    }

    pub fn is_only_one_and_zero(self) -> bool {
        self.0.iter().all(|c| *c == 0 || *c == 1 || *c == -1)
    }

    pub fn is_only_two_one_and_zero(self) -> bool {
        self.0
            .iter()
            .all(|c| *c == 0 || *c == 1 || *c == -1 || *c == 2 || *c == -2)
    }

    pub fn is_base_connecting_vector(self) -> bool {
        self.is_only_one_and_zero() && self.distance_squared() == 2
    }

    pub fn is_connection_vector(self) -> bool {
        if !self.is_only_two_one_and_zero() {
            return false;
        }
        let size = self.distance_squared();
        size == 1 || size == 2 || size == 3 || size == 5
    }

    pub fn sum_of_positive_coord(self) -> DInt {
        let mut res = 0;
        for c in self.0 {
            if c > 0 {
                res += DInt::from(c);
            }
        }
        res
    }

    /// Nearest main point (coordinates divisible by 3), using the Go switch on `c % 3`.
    pub fn near_main_point(self) -> Self {
        let mut res = Point([0, 0, 0]);
        for i in 0..3 {
            let c = self.0[i];
            res.0[i] = match c % THREE {
                0 => c,
                1 => c - 1,
                2 => c + 1,
                -1 => c + 1,
                -2 => c - 1,
                _ => 0,
            };
        }
        res
    }

    pub fn murmur_hash(self) -> u32 {
        crate::murmur::murmur_hash_u32(&[self.0[0] as u32, self.0[1] as u32, self.0[2] as u32])
    }
}

impl fmt::Display for Point {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "[ {} , {} , {} ]",
            fmt_signed(self.0[0]),
            fmt_signed(self.0[1]),
            fmt_signed(self.0[2])
        )
    }
}

/// Go `% d`: a leading space for non-negative numbers.
fn fmt_signed(v: CInt) -> String {
    if v < 0 {
        format!("{v}")
    } else {
        format!(" {v}")
    }
}

pub fn make_vector(p1: Point, p2: Point) -> Point {
    p2.sub(p1)
}

pub fn create_random_point(max: CInt) -> Point {
    Point([random_cint(max), random_cint(max), random_cint(max)])
}

pub fn random_cint(max: CInt) -> CInt {
    let mut rng = rand::thread_rng();
    let r = rng.gen_range(0..max);
    if rng.gen::<f32>() < 0.5 {
        -r
    } else {
        r
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ds() {
        assert_eq!(ds(Point([1, 2, 3]), Point([1, 2, 3])), 0);
        assert_eq!(ds(Point([1, 2, 3]), Point([0, 2, 3])), 1);
        assert_eq!(ds(Point([1, 2, 3]), Point([2, 2, 3])), 1);
        assert_eq!(ds(Point([1, 2, 3]), Point([1, 3, 3])), 1);
        assert_eq!(ds(Point([1, 2, 3]), Point([1, 1, 3])), 1);
        assert_eq!(ds(Point([1, 2, 3]), Point([1, 2, 4])), 1);
        assert_eq!(ds(Point([1, 2, 3]), Point([1, 2, 2])), 1);
        assert_eq!(Point([1, 1, 1]).distance_squared(), 3);
        assert_eq!(Point([-1, 1, 1]).distance_squared(), 3);
        assert_eq!(Point([1, -1, 1]).distance_squared(), 3);
        assert_eq!(Point([1, 1, -1]).distance_squared(), 3);
        assert_eq!(Point([-1, -1, -1]).distance_squared(), 3);
        assert_eq!(Point([1, 2, 3]).distance_squared(), 14);
        assert_eq!(ds(Point([-3, -2, -1]), Point([-3, -2, -1])), 0);
        assert_eq!(ds(Point([-3, -2, -1]), Point([-2, -1, 0])), 3);
    }

    #[test]
    fn test_nb_pos_coord() {
        assert_eq!(ORIGIN.sum_of_positive_coord(), 0);
        assert_eq!(Point([-1, 0, 0]).sum_of_positive_coord(), 0);
        assert_eq!(Point([0, -1, 0]).sum_of_positive_coord(), 0);
        assert_eq!(Point([0, 0, -1]).sum_of_positive_coord(), 0);
        assert_eq!(Point([-34, -45, -14]).sum_of_positive_coord(), 0);
        assert_eq!(Point([34, -45, -14]).sum_of_positive_coord(), 34);
        assert_eq!(Point([-34, 45, -14]).sum_of_positive_coord(), 45);
        assert_eq!(Point([-34, -45, 14]).sum_of_positive_coord(), 14);
        assert_eq!(Point([1, 2, 3]).sum_of_positive_coord(), 6);
    }

    #[test]
    fn test_point_equal_and_ds() {
        let orig = Point([0, 0, 0]);
        let one_two_three = Point([1, 2, 3]);
        let p = Point([17, 11, 13]);
        assert_eq!(orig, ORIGIN);
        assert_eq!(Point([3, 0, 0]), X_FIRST);
        assert_eq!(Point([0, 3, 0]), Y_FIRST);
        assert_eq!(Point([0, 0, 3]), Z_FIRST);
        assert_eq!(ds(one_two_three, Point([0, 1, 2])), 3);
        assert_eq!(one_two_three, Point([1, 2, 3]));
        assert_eq!(ds(one_two_three, Point([-1, 2, 3])), 4);
        assert_eq!(ds(one_two_three, Point([1, -2, 3])), 16);
        assert_eq!(ds(one_two_three, Point([1, 2, -3])), 36);
        assert_eq!(p, Point([17, 11, 13]));
    }

    #[test]
    fn test_is_main_point() {
        assert!(ORIGIN.is_main_point());
        assert!(X_FIRST.is_main_point());
        assert!(X_FIRST.neg().is_main_point());
        assert!(Y_FIRST.is_main_point());
        assert!(Y_FIRST.neg().is_main_point());
        assert!(Z_FIRST.is_main_point());
        assert!(Z_FIRST.neg().is_main_point());
        assert!(Point([6, 3, 0]).is_main_point());
        assert!(Point([-6, -3, 0]).is_main_point());
        assert!(Point([-9, 12, 3]).is_main_point());
        assert!(Point([21, -12, -3]).is_main_point());
        assert!(!Point([1, 0, 0]).is_main_point());
        assert!(!Point([0, 1, 0]).is_main_point());
        assert!(!Point([0, 0, 1]).is_main_point());
        assert!(!Point([1, 0, 0]).neg().is_main_point());
        assert!(!Point([0, 1, 0]).neg().is_main_point());
        assert!(!Point([0, 0, 1]).neg().is_main_point());
        assert!(!Point([1, 0, 0]).mul(2).is_main_point());
        assert!(!Point([0, 1, 0]).mul(2).is_main_point());
        assert!(!Point([0, 0, 1]).mul(2).is_main_point());
        assert!(!Point([1, 0, 0]).mul(-2).is_main_point());
        assert!(!Point([0, 1, 0]).mul(-2).is_main_point());
        assert!(!Point([0, 0, 1]).mul(-2).is_main_point());
    }

    #[test]
    fn test_is_connection_vector() {
        assert!(Point([1, 1, 0]).is_base_connecting_vector());
        assert!(Point([0, 1, 1]).is_base_connecting_vector());
        assert!(Point([1, 0, 1]).is_base_connecting_vector());
        assert!(Point([1, -1, 0]).is_base_connecting_vector());
        assert!(Point([0, 1, -1]).is_base_connecting_vector());
        assert!(Point([-1, 0, 1]).is_base_connecting_vector());
        assert!(!ORIGIN.is_base_connecting_vector());
        assert!(!Point([1, 0, 0]).is_base_connecting_vector());
        assert!(!Point([-1, 0, 0]).is_base_connecting_vector());
        assert!(!Point([1, 1, 1]).is_base_connecting_vector());
        assert!(!Point([1, -1, 1]).is_base_connecting_vector());
        assert!(!Point([0, 0, 2]).is_base_connecting_vector());
        assert!(!Point([0, -2, 0]).is_base_connecting_vector());
        assert!(!Point([2, 0, 0]).is_base_connecting_vector());

        assert!(Point([1, 1, 0]).is_connection_vector());
        assert!(Point([1, 0, 0]).is_connection_vector());
        assert!(Point([-1, 0, 0]).is_connection_vector());
        assert!(Point([1, 1, 1]).is_connection_vector());
        assert!(Point([1, -1, 1]).is_connection_vector());
        assert!(Point([1, -2, 0]).is_connection_vector());
        assert!(Point([0, -2, 1]).is_connection_vector());
        assert!(!ORIGIN.is_connection_vector());
        assert!(!Point([2, 0, 0]).is_connection_vector());
        assert!(!Point([0, 2, 0]).is_connection_vector());
        assert!(!Point([0, 0, 2]).is_connection_vector());
        assert!(!Point([-2, 0, 0]).is_connection_vector());
        assert!(!Point([0, -2, 0]).is_connection_vector());
        assert!(!Point([0, 0, -2]).is_connection_vector());
        assert!(!Point([2, 1, 1]).is_connection_vector());
        assert!(!Point([2, 2, 1]).is_connection_vector());
        assert!(!Point([3, 0, 1]).is_connection_vector());
        assert!(!Point([2, 3, 0]).is_connection_vector());
        assert!(!Point([1, 0, 3]).is_connection_vector());
    }

    #[test]
    fn test_near_main_point() {
        assert_eq!(ORIGIN.near_main_point(), ORIGIN);
        assert_eq!(X_FIRST.near_main_point(), X_FIRST);
        assert_eq!(X_FIRST.neg().near_main_point(), X_FIRST.neg());
        assert_eq!(Y_FIRST.near_main_point(), Y_FIRST);
        assert_eq!(Y_FIRST.neg().near_main_point(), Y_FIRST.neg());
        assert_eq!(Z_FIRST.near_main_point(), Z_FIRST);
        assert_eq!(Z_FIRST.neg().near_main_point(), Z_FIRST.neg());
        assert_eq!(Point([1, 0, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([-1, 0, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([0, 1, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([0, -1, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([0, 0, 1]).near_main_point(), ORIGIN);
        assert_eq!(Point([0, 0, -1]).near_main_point(), ORIGIN);
        assert_eq!(Point([1, 1, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([-1, 1, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([-1, -1, 0]).near_main_point(), ORIGIN);
        assert_eq!(Point([1, 1, 1]).near_main_point(), ORIGIN);
        assert_eq!(Point([-1, -1, -1]).near_main_point(), ORIGIN);
        assert_eq!(Point([2, 0, 0]).near_main_point(), X_FIRST);
        assert_eq!(Point([2, 1, 0]).near_main_point(), X_FIRST);
        assert_eq!(Point([2, 1, 1]).near_main_point(), X_FIRST);
        assert_eq!(Point([2, 1, -1]).near_main_point(), X_FIRST);
        assert_eq!(Point([2, -1, -1]).near_main_point(), X_FIRST);
    }

    #[test]
    fn test_add_sub_mul_neg() {
        let orig = Point([0, 0, 0]);
        let one_two_three = Point([1, 2, 3]);
        let p = Point([17, 11, 13]);
        assert_eq!(orig.add(X_FIRST), Point([3, 0, 0]));
        assert_eq!(orig, ORIGIN);
        assert_eq!(orig.add(Y_FIRST), Point([0, 3, 0]));
        assert_eq!(orig.add(Z_FIRST), Point([0, 0, 3]));
        assert_eq!(X_FIRST.add(Y_FIRST).add(Z_FIRST), Point([3, 3, 3]));
        assert_eq!(X_FIRST.add(Y_FIRST).add(Z_FIRST).neg(), Point([-3, -3, -3]));
        assert_eq!(p.add(one_two_three), Point([18, 13, 16]));
        assert_eq!(p, Point([17, 11, 13]));
        assert_eq!(one_two_three, Point([1, 2, 3]));
        assert_eq!(orig.sub(X_FIRST), Point([-3, 0, 0]));
        assert_eq!(orig.sub(Y_FIRST), Point([0, -3, 0]));
        assert_eq!(orig.sub(Z_FIRST), Point([0, 0, -3]));
        assert_eq!(p.sub(one_two_three), Point([16, 9, 10]));
        assert_eq!(one_two_three.neg(), Point([-1, -2, -3]));
        assert_eq!(one_two_three.mul(2), Point([2, 4, 6]));
        assert_eq!(one_two_three.mul(-3), Point([-3, -6, -9]));
    }

    #[test]
    fn test_point_rotations() {
        let orig = Point([0, 0, 0]);
        let one_two_three = Point([1, 2, 3]);
        assert_eq!(one_two_three.rot_plus_x(), Point([1, -3, 2]));
        assert_eq!(one_two_three.rot_neg_x(), Point([1, 3, -2]));
        assert_eq!(one_two_three.rot_plus_y(), Point([3, 2, -1]));
        assert_eq!(one_two_three.rot_neg_y(), Point([-3, 2, 1]));
        assert_eq!(one_two_three.rot_plus_z(), Point([-2, 1, 3]));
        assert_eq!(one_two_three.rot_neg_z(), Point([2, -1, 3]));

        for _ in 0..100 {
            let random_point = create_random_point(100_000_000);
            assert_eq!(orig.sub(random_point), random_point.neg());
            assert_eq!(
                random_point.sub(random_point.add(one_two_three)),
                one_two_three.neg()
            );
            assert_eq!(
                random_point.sub(random_point.add(one_two_three)),
                one_two_three.mul(-1)
            );
            assert_eq!(random_point.add(random_point.neg()), orig);
            assert_eq!(random_point.add(random_point.mul(-1)), orig);
            assert_eq!(random_point.rot_plus_x().rot_neg_x(), random_point);
            assert_eq!(random_point.rot_neg_x().rot_plus_x(), random_point);
            assert_eq!(random_point.rot_plus_y().rot_neg_y(), random_point);
            assert_eq!(random_point.rot_neg_y().rot_plus_y(), random_point);
            assert_eq!(random_point.rot_plus_z().rot_neg_z(), random_point);
            assert_eq!(random_point.rot_neg_z().rot_plus_z(), random_point);
            assert_eq!(
                random_point
                    .rot_plus_x()
                    .rot_plus_x()
                    .rot_plus_x()
                    .rot_plus_x(),
                random_point
            );
            assert_eq!(
                random_point
                    .rot_plus_y()
                    .rot_plus_y()
                    .rot_plus_y()
                    .rot_plus_y(),
                random_point
            );
            assert_eq!(
                random_point
                    .rot_plus_z()
                    .rot_plus_z()
                    .rot_plus_z()
                    .rot_plus_z(),
                random_point
            );
            assert_eq!(
                random_point.rot_neg_x().rot_neg_x().rot_neg_x().rot_neg_x(),
                random_point
            );
            assert_eq!(
                random_point.rot_neg_y().rot_neg_y().rot_neg_y().rot_neg_y(),
                random_point
            );
            assert_eq!(
                random_point.rot_neg_z().rot_neg_z().rot_neg_z().rot_neg_z(),
                random_point
            );
            assert_eq!(
                random_point.rot_neg_x().rot_neg_x(),
                random_point.rot_plus_x().rot_plus_x()
            );
            assert_eq!(
                random_point.rot_neg_y().rot_neg_y(),
                random_point.rot_plus_y().rot_plus_y()
            );
            assert_eq!(
                random_point.rot_neg_z().rot_neg_z(),
                random_point.rot_plus_z().rot_plus_z()
            );
        }
    }
}
