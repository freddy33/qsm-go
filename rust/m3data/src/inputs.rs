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

/// Space parameters for the one space this simulation folder is.
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
    let raw = std::fs::read(path)?;
    parse_inputs(&raw)
}

pub fn parse_inputs(raw: &[u8]) -> Result<SimulationInputs> {
    Ok(serde_json::from_slice(raw)?)
}

/// `simulation-configs/<name>-config.json`.
///
/// Every field may be omitted. Absent `shortcuts` means the default pyramid.
/// `"shortcuts": []` means no shortcuts. Event fields that are omitted use the
/// origin, creation time 0, red, and growth type 8. The result is a complete
/// [`SimulationInputs`]: the file written to `data/<name>/inputs.json` still
/// keeps shortcuts, with every default filled in.
#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
struct ConfigFile {
    #[serde(default)]
    doc: Option<String>,
    #[serde(default)]
    max_time_steps: Option<i32>,
    #[serde(default)]
    space: Option<ConfigSpace>,
    #[serde(default)]
    events: Vec<ConfigEvent>,
    /// `None` is a missing key (default pyramid). `Some` is the key as written,
    /// including an empty list.
    #[serde(default)]
    shortcuts: Option<Vec<Shortcut>>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
struct ConfigSpace {
    #[serde(default)]
    active_path_node_threshold: Option<i32>,
    #[serde(default)]
    max_trios_per_point: Option<i32>,
    #[serde(default)]
    max_path_nodes_per_point: Option<i32>,
    #[serde(default)]
    max_coord: Option<i32>,
}

#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
struct ConfigEvent {
    #[serde(default)]
    x: i32,
    #[serde(default)]
    y: i32,
    #[serde(default)]
    z: i32,
    #[serde(default)]
    creation_time: i32,
    #[serde(default = "default_event_color")]
    color: u8,
    #[serde(default = "default_pyramid_growth_type")]
    growth_type: u8,
    #[serde(default)]
    growth_index: i32,
    #[serde(default)]
    growth_offset: i32,
}

fn default_event_color() -> u8 {
    EventColor::Red as u8
}

/// Parse a minimal config and fill the same defaults as [`default_inputs`].
pub fn expand_config(raw: &[u8]) -> Result<SimulationInputs> {
    let config: ConfigFile = serde_json::from_slice(raw)?;
    let defaults = default_inputs();
    let base = SpaceParams::default();
    let space = match config.space {
        None => base.clone(),
        Some(partial) => SpaceParams {
            active_path_node_threshold: partial
                .active_path_node_threshold
                .unwrap_or(base.active_path_node_threshold),
            max_trios_per_point: partial
                .max_trios_per_point
                .unwrap_or(base.max_trios_per_point),
            max_path_nodes_per_point: partial
                .max_path_nodes_per_point
                .unwrap_or(base.max_path_nodes_per_point),
            max_coord: partial.max_coord.unwrap_or(base.max_coord),
        },
    };
    let events = config
        .events
        .into_iter()
        .map(|event| EventInput {
            x: event.x,
            y: event.y,
            z: event.z,
            creation_time: event.creation_time,
            color: event.color,
            growth_type: event.growth_type,
            growth_index: event.growth_index,
            growth_offset: event.growth_offset,
        })
        .collect();
    let shortcuts = config
        .shortcuts
        .unwrap_or_else(|| defaults.shortcuts.clone());
    let mut inputs = SimulationInputs {
        doc: config.doc.unwrap_or_default(),
        max_time_steps: config.max_time_steps.unwrap_or(defaults.max_time_steps),
        space,
        events,
        shortcuts,
    };
    if inputs.doc.is_empty() {
        let mut filled = inputs.clone();
        filled.doc = defaults.doc.clone();
        if filled == defaults {
            inputs.doc = defaults.doc;
        }
    }
    validate_inputs(&inputs)?;
    Ok(inputs)
}

pub fn load_config(path: &Path) -> Result<SimulationInputs> {
    let raw = std::fs::read(path)?;
    expand_config(&raw)
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

/// Explicit events, then each shortcut, in the order they are created.
pub fn expand_events(inputs: &SimulationInputs) -> Result<Vec<EventInput>> {
    let mut events = inputs.events.clone();
    for shortcut in &inputs.shortcuts {
        match shortcut {
            Shortcut::Pyramid(spec) => events.extend(expand_pyramid(spec)?),
        }
    }
    Ok(events)
}

pub fn resolve_events(inputs: &SimulationInputs) -> Result<Vec<ResolvedEvent>> {
    validate_inputs(inputs)?;
    let pack = PointPackData::shared();
    let mut resolved = Vec::new();
    for event in expand_events(inputs)? {
        resolved.push(resolve_one(pack, &event)?);
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
    use std::path::Path;

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

    #[test]
    fn an_empty_config_is_the_default_pyramid() {
        let inputs = expand_config(b"{}").unwrap();
        assert_eq!(inputs, default_inputs());
        let explicit =
            expand_config(br#"{ "max_time_steps": 3, "shortcuts": [ { "pyramid": {} } ] }"#)
                .unwrap();
        assert_eq!(explicit, default_inputs());
    }

    #[test]
    fn a_partial_pyramid_and_a_single_event_fill_defaults() {
        let pyramid = expand_config(
            br#"{ "max_time_steps": 4, "shortcuts": [ { "pyramid": { "size": 2 } } ] }"#,
        )
        .unwrap();
        assert_eq!(pyramid.max_time_steps, 4);
        assert_eq!(pyramid.space, SpaceParams::default());
        match &pyramid.shortcuts[0] {
            Shortcut::Pyramid(spec) => {
                assert_eq!(spec.size, 2);
                assert_eq!(spec.growth_type, 8);
                assert_eq!(spec.growth_index, 0);
                assert_eq!(spec.offsets, vec![0, 2, 4, 6]);
                assert_eq!(spec.creation_time, 0);
            }
        }
        let single = expand_config(
            br#"{ "max_time_steps": 3, "events": [ { "color": 1 } ], "shortcuts": [] }"#,
        )
        .unwrap();
        assert!(single.shortcuts.is_empty());
        assert_eq!(single.events.len(), 1);
        assert_eq!(single.events[0].x, 0);
        assert_eq!(single.events[0].y, 0);
        assert_eq!(single.events[0].z, 0);
        assert_eq!(single.events[0].creation_time, 0);
        assert_eq!(single.events[0].color, 1);
        assert_eq!(single.events[0].growth_type, 8);
        assert_eq!(single.events[0].growth_index, 0);
        assert_eq!(single.events[0].growth_offset, 0);
        assert_eq!(single.space.max_coord, SpaceParams::default().max_coord);
        assert!(resolve_events(&single).unwrap().len() == 1);
    }

    #[test]
    fn an_unknown_config_field_is_rejected() {
        let err = expand_config(br#"{ "nope": 1 }"#).unwrap_err();
        assert!(err.to_string().contains("nope"), "{err}");
    }

    #[test]
    fn the_checked_in_pyramid_config_matches_the_test_default() {
        let path = Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("../../simulation-configs/pyramid-config.json");
        let inputs = load_config(&path).unwrap();
        assert_eq!(inputs, default_inputs());
    }

    #[test]
    fn checked_in_variant_configs_fill_the_documented_defaults() {
        let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("../../simulation-configs");
        let small = load_config(&root.join("pyramid-size2-config.json")).unwrap();
        assert_eq!(small.max_time_steps, 4);
        assert_eq!(small.space, SpaceParams::default());
        match &small.shortcuts[0] {
            Shortcut::Pyramid(spec) => {
                assert_eq!(spec.size, 2);
                assert_eq!(spec.growth_type, 8);
                assert_eq!(spec.offsets, vec![0, 2, 4, 6]);
            }
        }
        let single = load_config(&root.join("single-event-config.json")).unwrap();
        assert!(single.shortcuts.is_empty());
        assert_eq!(single.events[0].growth_type, 8);
        assert_eq!(single.events[0].color, 1);
        assert_eq!(resolve_events(&single).unwrap().len(), 1);
    }
}
