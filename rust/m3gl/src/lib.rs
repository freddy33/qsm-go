//! Drawing filters, colors, and object types for the QSM viewer.
//!
//! The window is the Bevy app `qsm-view`. These types keep the color, dimmer,
//! and display-filter rules from `ui/m3gl/m3drawing.go`. The Go packages still
//! open a GLFW window; the Rust viewer does not.

use m3point::{ConnectionId, Point, ORIGIN};
use m3space::{DistAndTime, EventColor};
use m3util::count_the_ones;

pub const NO_DIMMER: f32 = 1.0;
pub const DEFAULT_GREY_DIMMER: f32 = 0.7;
pub const DEFAULT_OLD_EVENT_DIMMER: f32 = 0.3;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
#[repr(u8)]
pub enum ObjectType {
    AxeX = 0,
    AxeY = 1,
    AxeZ = 2,
    NodeEmpty = 3,
    NodeActive = 4,
    Connection00 = 5,
}

impl ObjectType {
    pub fn is_axe(self) -> bool {
        (self as u8) <= ObjectType::AxeZ as u8
    }
    pub fn is_node(self) -> bool {
        (ObjectType::NodeEmpty as u8..=ObjectType::NodeActive as u8).contains(&(self as u8))
    }
    pub fn is_connection(self) -> bool {
        (self as u8) >= ObjectType::Connection00 as u8
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SpaceDrawingFilter {
    pub display_empty_nodes: bool,
    pub display_empty_connections: bool,
    pub event_color_mask: u8,
    pub event_outgrowth_many_colors_threshold: u8,
    pub active_threshold: DistAndTime,
}

impl SpaceDrawingFilter {
    pub fn colors_increase(&mut self) {
        self.event_outgrowth_many_colors_threshold += 1;
        if self.event_outgrowth_many_colors_threshold > 4 {
            self.event_outgrowth_many_colors_threshold = 4;
        }
    }

    pub fn colors_decrease(&mut self) {
        self.event_outgrowth_many_colors_threshold =
            self.event_outgrowth_many_colors_threshold.saturating_sub(1);
    }

    pub fn color_mask_switch(&mut self, color: EventColor) {
        self.event_color_mask ^= color as u8;
    }
}

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct SpaceDrawingColor {
    pub obj_colors: u8,
    pub dim_colors: u8,
}

impl SpaceDrawingColor {
    pub fn has_color(self, c: EventColor) -> bool {
        self.obj_colors & (c as u8) != 0
    }

    pub fn has_dimming(self, c: EventColor) -> bool {
        self.dim_colors & (c as u8) != 0
    }

    pub fn how_many_colors(self) -> u8 {
        count_the_ones(self.obj_colors)
    }

    pub fn single_color(self) -> u8 {
        if self.obj_colors == 0 {
            return 0;
        }
        for c in m3space::ALL_COLORS {
            if self.has_color(c) {
                return c as u8;
            }
        }
        0
    }

    pub fn second_color(self) -> u8 {
        if self.obj_colors == 0 {
            return 0;
        }
        let mut found_one = false;
        for c in m3space::ALL_COLORS {
            if self.has_color(c) {
                if found_one {
                    return c as u8;
                }
                found_one = true;
            }
        }
        0
    }

    pub fn color(self, blink_value: f64) -> i32 {
        let color_switch = blink_value as u8;
        let color = 1u8 << color_switch;
        match self.how_many_colors() {
            0 => 0,
            1 => i32::from(self.single_color()),
            2 => {
                if i32::from(color_switch) % 2 == 0 {
                    i32::from(self.single_color())
                } else {
                    i32::from(self.second_color())
                }
            }
            3 | 4 => {
                if self.obj_colors & color != 0 {
                    i32::from(color)
                } else {
                    0
                }
            }
            _ => 0,
        }
    }

    pub fn dimmer(self, blink_value: f64) -> f32 {
        let color_switch = blink_value as u8;
        let color = 1u8 << color_switch;
        match self.how_many_colors() {
            0 => DEFAULT_GREY_DIMMER,
            1 => {
                if self.has_dimming(
                    EventColor::from_u8(self.single_color()).unwrap_or(EventColor::Red),
                ) && self.single_color() != 0
                {
                    DEFAULT_OLD_EVENT_DIMMER
                } else if self.single_color() != 0
                    && self.has_dimming(EventColor::from_u8(self.single_color()).unwrap())
                {
                    DEFAULT_OLD_EVENT_DIMMER
                } else {
                    NO_DIMMER
                }
            }
            2 => {
                let which = if i32::from(color_switch) % 2 == 0 {
                    self.single_color()
                } else {
                    self.second_color()
                };
                match EventColor::from_u8(which) {
                    Some(c) if self.has_dimming(c) => DEFAULT_OLD_EVENT_DIMMER,
                    _ => NO_DIMMER,
                }
            }
            3 | 4 => {
                if self.obj_colors & color != 0 {
                    match EventColor::from_u8(color) {
                        Some(c) if self.has_dimming(c) => DEFAULT_OLD_EVENT_DIMMER,
                        _ => NO_DIMMER,
                    }
                } else {
                    DEFAULT_GREY_DIMMER
                }
            }
            _ => DEFAULT_GREY_DIMMER,
        }
    }
}

pub fn connection_object_type(cd_id: ConnectionId) -> u8 {
    let base = ObjectType::Connection00 as i16;
    if cd_id.0 > 0 {
        (base + i16::from(cd_id.0) * 2) as u8
    } else {
        (base + 1 - i16::from(cd_id.0) * 2) as u8
    }
}

#[derive(Clone, Debug)]
pub struct NodeDrawingElement {
    pub object_type: ObjectType,
    pub color: SpaceDrawingColor,
    pub has_root: bool,
    pub point: Point,
}

impl NodeDrawingElement {
    pub fn from_mask(color_mask: u8, has_root: bool, point: Point) -> Self {
        let color = SpaceDrawingColor {
            obj_colors: color_mask,
            dim_colors: 0,
        };
        let object_type = if color_mask != 0 {
            ObjectType::NodeActive
        } else {
            ObjectType::NodeEmpty
        };
        Self {
            object_type,
            color,
            has_root,
            point,
        }
    }

    pub fn display(&self, filter: &SpaceDrawingFilter) -> bool {
        if self.object_type == ObjectType::NodeActive {
            if self.has_root {
                return true;
            }
            return self.color.obj_colors & filter.event_color_mask != 0
                && self.color.how_many_colors() >= filter.event_outgrowth_many_colors_threshold;
        }
        filter.display_empty_nodes
    }
}

#[derive(Clone, Debug)]
pub struct ConnectionDrawingElement {
    pub object_type: u8,
    pub color: SpaceDrawingColor,
    pub pos: Point,
}

impl ConnectionDrawingElement {
    pub fn new(color_mask: u8, point: Point, conn_id: ConnectionId) -> Self {
        Self {
            object_type: connection_object_type(conn_id),
            color: SpaceDrawingColor {
                obj_colors: color_mask,
                dim_colors: 0,
            },
            pos: point,
        }
    }

    pub fn display(&self, filter: &SpaceDrawingFilter) -> bool {
        if self.color.obj_colors & filter.event_color_mask != 0
            && self.color.how_many_colors() >= filter.event_outgrowth_many_colors_threshold
        {
            return true;
        }
        filter.display_empty_connections
    }
}

#[derive(Clone, Copy, Debug)]
pub struct AxeDrawingElement {
    pub object_type: ObjectType,
    pub max: m3point::CInt,
    pub neg: bool,
}

impl AxeDrawingElement {
    pub fn display(self) -> bool {
        let _ = self;
        true
    }

    pub fn color(self) -> i32 {
        match self.object_type {
            ObjectType::AxeX => EventColor::Red as i32,
            ObjectType::AxeY => EventColor::Green as i32,
            ObjectType::AxeZ => EventColor::Blue as i32,
            _ => 0,
        }
    }

    pub fn dimmer(self) -> f32 {
        let _ = self;
        1.0
    }

    pub fn pos(self) -> Point {
        if self.neg {
            match self.object_type {
                ObjectType::AxeX => return Point([-self.max, 0, 0]),
                ObjectType::AxeY => return Point([0, -self.max, 0]),
                ObjectType::AxeZ => return Point([0, 0, -self.max]),
                _ => {}
            }
        }
        ORIGIN
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn filter() -> SpaceDrawingFilter {
        SpaceDrawingFilter {
            display_empty_nodes: false,
            display_empty_connections: false,
            event_color_mask: EventColor::Red as u8 | EventColor::Green as u8,
            event_outgrowth_many_colors_threshold: 1,
            active_threshold: 0,
        }
    }

    #[test]
    fn filter_thresholds_and_color_mask() {
        let mut f = filter();
        f.colors_increase();
        assert_eq!(f.event_outgrowth_many_colors_threshold, 2);
        for _ in 0..10 {
            f.colors_increase();
        }
        assert_eq!(f.event_outgrowth_many_colors_threshold, 4);
        f.colors_decrease();
        assert_eq!(f.event_outgrowth_many_colors_threshold, 3);
        f.event_outgrowth_many_colors_threshold = 0;
        f.colors_decrease();
        assert_eq!(f.event_outgrowth_many_colors_threshold, 0);
        let before = f.event_color_mask;
        f.color_mask_switch(EventColor::Red);
        assert_ne!(f.event_color_mask, before);
        f.color_mask_switch(EventColor::Red);
        assert_eq!(f.event_color_mask, before);
    }

    #[test]
    fn node_and_connection_display_rules() {
        let f = filter();
        let empty = NodeDrawingElement::from_mask(0, false, ORIGIN);
        assert!(!empty.display(&f));
        let mut show_empty = f.clone();
        show_empty.display_empty_nodes = true;
        assert!(empty.display(&show_empty));

        let red = NodeDrawingElement::from_mask(EventColor::Red as u8, false, Point([1, 0, 0]));
        assert!(red.display(&f));
        assert_eq!(red.object_type, ObjectType::NodeActive);
        let blue = NodeDrawingElement::from_mask(EventColor::Blue as u8, false, ORIGIN);
        assert!(!blue.display(&f));
        let root = NodeDrawingElement::from_mask(EventColor::Blue as u8, true, ORIGIN);
        assert!(root.display(&f));

        let conn =
            ConnectionDrawingElement::new(EventColor::Red as u8, Point([1, 1, 0]), ConnectionId(4));
        assert!(conn.display(&f));
        assert!(conn.object_type >= ObjectType::Connection00 as u8);
        let hidden =
            ConnectionDrawingElement::new(EventColor::Yellow as u8, ORIGIN, ConnectionId(-4));
        assert!(!hidden.display(&f));
    }

    #[test]
    fn blink_color_and_axes() {
        let mut sdc = SpaceDrawingColor {
            obj_colors: EventColor::Red as u8 | EventColor::Green as u8,
            dim_colors: EventColor::Green as u8,
        };
        assert_eq!(sdc.how_many_colors(), 2);
        assert_eq!(sdc.color(0.0), EventColor::Red as i32);
        assert_eq!(sdc.color(1.0), EventColor::Green as i32);
        assert_eq!(sdc.dimmer(0.0), NO_DIMMER);
        assert_eq!(sdc.dimmer(1.0), DEFAULT_OLD_EVENT_DIMMER);
        sdc.obj_colors = 0;
        assert_eq!(sdc.color(0.0), 0);
        assert_eq!(sdc.dimmer(0.0), DEFAULT_GREY_DIMMER);

        let axe = AxeDrawingElement {
            object_type: ObjectType::AxeY,
            max: 9,
            neg: true,
        };
        assert!(axe.display());
        assert_eq!(axe.color(), EventColor::Green as i32);
        assert_eq!(axe.dimmer(), 1.0);
        assert_eq!(axe.pos(), Point([0, -9, 0]));
        assert!(ObjectType::AxeX.is_axe());
        assert!(ObjectType::NodeActive.is_node());
        assert!(ObjectType::Connection00.is_connection());
        assert_eq!(
            connection_object_type(ConnectionId(1)),
            ObjectType::Connection00 as u8 + 2
        );
        assert_eq!(
            connection_object_type(ConnectionId(-1)),
            ObjectType::Connection00 as u8 + 1 + 2
        );
    }
}
