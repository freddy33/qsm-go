//! `data/<simulation>/inputs.json`, the initial state of a run.
//!
//! Explicit events are created as written. Shortcut objects such as `"pyramid"`
//! expand to events. A new shortcut is another enum variant, serialized as its
//! own key, so existing files keep loading.

use std::path::Path;

use m3point::{GrowthType, Point, PointPackData, ALL_GROWTH_TYPES};
use m3space::{pyramid_centers, EventColor, ALL_COLORS, MIN_MAX_COORD};
use serde::{Deserialize, Serialize};

use crate::error::{Error, Result};

/// Top-level document read by `qsm <name>`.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SimulationInputs {
    /// Human-readable note. Ignored by the stepper.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub doc: String,
    /// Last creation time to compute. Roots are time 0 when events are created
    /// at 0, so `3` stores times 0, 1, 2 and 3.
    pub max_time_steps: i32,
    #[serde(default)]
    pub space: SpaceParams,
    /// Events created as themselves, before shortcuts.
    #[serde(default)]
    pub events: Vec<EventInput>,
    /// Named expansions. Each element is one object with a single key.
    #[serde(default)]
    pub shortcuts: Vec<Shortcut>,
}

/// Columns of a `spaces` row that are chosen up front, plus the coord limit.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SpaceParams {
    #[serde(default = "default_active_threshold")]
    pub active_path_node_threshold: i32,
    #[serde(default = "default_max_trios")]
    pub max_trios_per_point: i32,
    #[serde(default = "default_max_path_nodes")]
    pub max_path_nodes_per_point: i32,
    #[serde(default = "default_max_coord")]
    pub max_coord: i32,
}

impl Default for SpaceParams {
    fn default() -> Self {
        Self {
            active_path_node_threshold: default_active_threshold(),
            max_trios_per_point: default_max_trios(),
            max_path_nodes_per_point: default_max_path_nodes(),
            max_coord: default_max_coord(),
        }
    }
}

fn default_active_threshold() -> i32 {
    // `ui/m3gl.MakeWorld` is called from playgl with active threshold 3.
    3
}

fn default_max_trios() -> i32 {
    // `MakeWorld` passes this to `CreateSpace` as max trios per point.
    2
}

fn default_max_path_nodes() -> i32 {
    // `MakeWorld` passes this to `CreateSpace` as max path nodes per point.
    4
}

fn default_max_coord() -> i32 {
    MIN_MAX_COORD
}

/// One event, in the same fields as Go `CreateEvent`.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct EventInput {
    pub x: i32,
    pub y: i32,
    pub z: i32,
    pub creation_time: i32,
    /// `1` red, `2` green, `4` blue, `8` yellow.
    pub color: u8,
    pub growth_type: u8,
    pub growth_index: i32,
    pub growth_offset: i32,
}

impl EventInput {
    pub fn point(&self) -> Point {
        Point([self.x, self.y, self.z])
    }
}

/// A named expansion. New shortcuts are new variants.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Shortcut {
    Pyramid(PyramidShortcut),
}

/// `CreatePyramidWithParams` from `ui/m3gl/m3gl.go` and `backend/spacedb/pyramid.go`.
///
/// Four events, creation order red, green, blue, yellow, at
/// `{3,0,3}`, `{-3,3,3}`, `{-3,-3,3}`, `{0,0,-3}` multiplied by `size`.
/// playgl uses size 4, growth type 8, index 0, offsets 0, 2, 4, 6.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct PyramidShortcut {
    #[serde(default = "default_pyramid_size")]
    pub size: i32,
    #[serde(default = "default_pyramid_growth_type")]
    pub growth_type: u8,
    #[serde(default)]
    pub growth_index: i32,
    #[serde(default = "default_pyramid_offsets")]
    pub offsets: Vec<i32>,
    #[serde(default)]
    pub creation_time: i32,
}

fn default_pyramid_size() -> i32 {
    4
}

fn default_pyramid_growth_type() -> u8 {
    8
}

fn default_pyramid_offsets() -> Vec<i32> {
    vec![0, 2, 4, 6]
}

impl Default for PyramidShortcut {
    fn default() -> Self {
        Self {
            size: default_pyramid_size(),
            growth_type: default_pyramid_growth_type(),
            growth_index: 0,
            offsets: default_pyramid_offsets(),
            creation_time: 0,
        }
    }
}

/// Event after shortcut expansion and growth-context lookup.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ResolvedEvent {
    pub point: Point,
    pub creation_time: i32,
    pub color: EventColor,
    pub growth_ctx_id: i32,
    pub growth_offset: i32,
}

/// The file written when a new simulation folder has no `inputs.json`.
pub fn default_inputs() -> SimulationInputs {
    SimulationInputs {
        doc: "Default is the playgl pyramid: CreatePyramidWithParams(space, 4, growth types 8, indexes 0, offsets 0,2,4,6). max_time_steps is the last creation time stored. Time 0 is the four root nodes.".to_string(),
        max_time_steps: 3,
        space: SpaceParams::default(),
        events: Vec::new(),
        shortcuts: vec![Shortcut::Pyramid(PyramidShortcut::default())],
    }
}

pub fn load_inputs(path: &Path) -> Result<SimulationInputs> {
    let text = std::fs::read_to_string(path)?;
    Ok(serde_json::from_str(&text)?)
}

pub fn save_inputs(path: &Path, inputs: &SimulationInputs) -> Result<()> {
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    let text = serde_json::to_string_pretty(inputs)?;
    std::fs::write(path, text + "\n")?;
    Ok(())
}

pub fn validate_inputs(inputs: &SimulationInputs) -> Result<()> {
    if inputs.max_time_steps < 0 {
        return Err(Error::new("max_time_steps must be >= 0"));
    }
    let space = &inputs.space;
    if space.max_coord <= 0 || space.max_coord % 3 != 0 {
        return Err(Error::new(
            "space.max_coord must be a positive multiple of 3",
        ));
    }
    if space.active_path_node_threshold < 0 {
        return Err(Error::new("space.active_path_node_threshold must be >= 0"));
    }
    if space.max_trios_per_point <= 0 || space.max_path_nodes_per_point <= 0 {
        return Err(Error::new(
            "space max_trios_per_point and max_path_nodes_per_point must be positive",
        ));
    }
    Ok(())
}

pub fn resolve_events(inputs: &SimulationInputs) -> Result<Vec<ResolvedEvent>> {
    validate_inputs(inputs)?;
    let pack = PointPackData::shared();
    let mut resolved = Vec::new();
    for event in &inputs.events {
        resolved.push(resolve_one(pack, event)?);
    }
    for shortcut in &inputs.shortcuts {
        match shortcut {
            Shortcut::Pyramid(spec) => {
                for event in expand_pyramid(spec)? {
                    resolved.push(resolve_one(pack, &event)?);
                }
            }
        }
    }
    Ok(resolved)
}

pub fn expand_pyramid(spec: &PyramidShortcut) -> Result<Vec<EventInput>> {
    if spec.size == 0 {
        return Err(Error::new("pyramid size must be non-zero"));
    }
    if spec.offsets.len() != 4 {
        return Err(Error::new(
            "pyramid offsets must have 4 values, one per center",
        ));
    }
    let centers = pyramid_centers(spec.size);
    let mut events = Vec::with_capacity(4);
    for ((center, offset), color) in centers
        .into_iter()
        .zip(spec.offsets.iter().copied())
        .zip(ALL_COLORS)
    {
        events.push(EventInput {
            x: center.0[0],
            y: center.0[1],
            z: center.0[2],
            creation_time: spec.creation_time,
            color: color as u8,
            growth_type: spec.growth_type,
            growth_index: spec.growth_index,
            growth_offset: offset,
        });
    }
    Ok(events)
}

fn resolve_one(pack: &PointPackData, event: &EventInput) -> Result<ResolvedEvent> {
    let growth_type = GrowthType(event.growth_type);
    if !ALL_GROWTH_TYPES.contains(&growth_type) {
        return Err(Error::new(format!(
            "growth_type {} is not one of 1, 2, 3, 4, 8",
            event.growth_type
        )));
    }
    if event.growth_index < 0 || event.growth_index >= growth_type.nb_indexes() {
        return Err(Error::new(format!(
            "growth_index {} is outside 0..{} for type {}",
            event.growth_index,
            growth_type.nb_indexes(),
            event.growth_type
        )));
    }
    if event.growth_offset < 0 || event.growth_offset >= growth_type.max_offset() {
        return Err(Error::new(format!(
            "growth_offset {} is outside 0..{} for type {}",
            event.growth_offset,
            growth_type.max_offset(),
            event.growth_type
        )));
    }
    let Some(color) = EventColor::from_u8(event.color) else {
        return Err(Error::new(format!(
            "color {} is not 1 (red), 2 (green), 4 (blue), or 8 (yellow)",
            event.color
        )));
    };
    let ctx = pack.growth_by_type_and_index(growth_type, event.growth_index);
    Ok(ResolvedEvent {
        point: event.point(),
        creation_time: event.creation_time,
        color,
        growth_ctx_id: ctx.id,
        growth_offset: event.growth_offset,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn playgl_pyramid_expands_to_the_four_go_events() {
        let events = expand_pyramid(&PyramidShortcut::default()).unwrap();
        let expected = [
            (12, 0, 12, EventColor::Red as u8, 0),
            (-12, 12, 12, EventColor::Green as u8, 2),
            (-12, -12, 12, EventColor::Blue as u8, 4),
            (0, 0, -12, EventColor::Yellow as u8, 6),
        ];
        assert_eq!(events.len(), 4);
        for (event, (x, y, z, color, offset)) in events.iter().zip(expected) {
            assert_eq!((event.x, event.y, event.z), (x, y, z));
            assert_eq!(event.color, color);
            assert_eq!(event.growth_offset, offset);
            assert_eq!(event.growth_type, 8);
            assert_eq!(event.growth_index, 0);
            assert_eq!(event.creation_time, 0);
        }
        let resolved = resolve_events(&default_inputs()).unwrap();
        assert!(resolved.iter().all(|event| event.growth_ctx_id == 40));
    }
}
