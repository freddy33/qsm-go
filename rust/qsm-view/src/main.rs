//! `qsm-view <simulation name>` draws a saved simulation.
//!
//! Nodes share one sphere mesh and one material per event color, so Bevy
//! instances them. Connections, axes, and the lattice are gizmos. This file
//! is built only with `--features view`.

use std::num::NonZero;
use std::path::PathBuf;
use std::time::Instant;

use bevy::input::mouse::{MouseMotion, MouseScrollUnit, MouseWheel};
use bevy::prelude::*;
use bevy::render::view::screenshot::{save_to_disk, Capturing, Screenshot};
use bevy_egui::{egui, EguiContexts, EguiPlugin, EguiPrimaryContextPass};
use qsm_view::{
    build_frame, camera_fit, color_rgb, load_simulation, Frame, LoadedSim, ViewFilter, NODE_RADIUS,
    ORIGIN_RADIUS,
};

fn main() {
    let args = Args::parse();
    let data_root = m3data::default_data_root();
    let sim = match load_simulation(&data_root, &args.name) {
        Ok(sim) => sim,
        Err(err) => {
            eprintln!("qsm-view: {err}");
            std::process::exit(1);
        }
    };
    println!(
        "qsm-view: {}  nodes {}  time 0..={}  max_coord {}",
        sim.name,
        sim.nodes.len(),
        sim.max_time,
        sim.max_coord
    );
    if let Some(path) = &args.screenshot {
        if let Some(parent) = path.parent() {
            if !parent.as_os_str().is_empty() {
                if let Err(err) = std::fs::create_dir_all(parent) {
                    eprintln!("qsm-view: {err}");
                    std::process::exit(1);
                }
            }
        }
        println!("screenshot: {}", path.display());
    }

    let (focus, distance) = camera_fit(&sim.nodes);
    let filter = ViewFilter::at_time(sim.max_time);
    App::new()
        .add_plugins(
            DefaultPlugins
                .set(WindowPlugin {
                    primary_window: Some(Window {
                        title: format!("qsm-view {}", sim.name),
                        resolution: (1280, 720).into(),
                        ..default()
                    }),
                    ..default()
                })
                .set(bevy::log::LogPlugin {
                    level: bevy::log::Level::WARN,
                    filter: "wgpu=error,naga=warn".to_string(),
                    ..default()
                }),
        )
        .add_plugins(EguiPlugin::default())
        .insert_resource(ClearColor(Color::srgb(0.07, 0.08, 0.10)))
        .insert_resource(Session { sim, filter })
        .insert_resource(Drawn(Frame::default()))
        .insert_resource(Orbit {
            yaw: 0.85,
            pitch: 0.45,
            distance,
            focus: Vec3::new(focus[0], focus[1], focus[2]),
        })
        .insert_resource(UiBlocks::default())
        .insert_resource(Shot {
            path: args.screenshot,
            started: Instant::now(),
            frames: 0,
            fired: false,
        })
        .add_systems(Startup, setup)
        .add_systems(
            Update,
            (keys, sync_scene, draw_gizmos, apply_camera, grab_screenshot).chain(),
        )
        .add_systems(EguiPrimaryContextPass, panel)
        .run();
}

struct Args {
    name: String,
    screenshot: Option<PathBuf>,
}

impl Args {
    fn parse() -> Self {
        let mut args = std::env::args().skip(1);
        let Some(name) = args.next() else {
            usage();
        };
        let mut screenshot = None;
        while let Some(flag) = args.next() {
            if flag == "--screenshot" {
                let Some(path) = args.next() else {
                    usage();
                };
                screenshot = Some(PathBuf::from(path));
            } else {
                usage();
            }
        }
        Self { name, screenshot }
    }
}

fn usage() -> ! {
    eprintln!("usage: qsm-view <simulation name> [--screenshot <png>]");
    std::process::exit(2);
}

#[derive(Resource)]
struct Session {
    sim: LoadedSim,
    filter: ViewFilter,
}

#[derive(Resource)]
struct Drawn(Frame);

#[derive(Resource)]
struct Orbit {
    yaw: f32,
    pitch: f32,
    distance: f32,
    focus: Vec3,
}

impl Orbit {
    fn eye(&self) -> Transform {
        let dir = Vec3::new(
            self.yaw.cos() * self.pitch.cos(),
            self.pitch.sin(),
            self.yaw.sin() * self.pitch.cos(),
        );
        let eye = self.focus + dir * self.distance;
        Transform::from_translation(eye).looking_at(self.focus, Vec3::Y)
    }
}

#[derive(Resource, Default)]
struct UiBlocks {
    pointer: bool,
    keyboard: bool,
}

#[derive(Resource)]
struct Shot {
    path: Option<PathBuf>,
    started: Instant,
    frames: u32,
    fired: bool,
}

#[derive(Resource)]
struct Palette {
    node: Handle<Mesh>,
    origin: Handle<Mesh>,
    materials: [Handle<StandardMaterial>; 4],
}

impl Palette {
    fn material(&self, color: u8) -> Handle<StandardMaterial> {
        let index = match color {
            1 => 0,
            2 => 1,
            4 => 2,
            _ => 3,
        };
        self.materials[index].clone()
    }
}

#[derive(Component)]
struct SceneMarker;

#[derive(Component)]
struct OrbitCam;

fn setup(
    mut commands: Commands,
    mut meshes: ResMut<Assets<Mesh>>,
    mut materials: ResMut<Assets<StandardMaterial>>,
) {
    let node = meshes.add(Sphere::new(NODE_RADIUS).mesh().uv(16, 12));
    let origin = meshes.add(Sphere::new(ORIGIN_RADIUS).mesh().uv(20, 16));
    let materials = [1u8, 2, 4, 8].map(|color| {
        let [r, g, b] = color_rgb(color);
        materials.add(StandardMaterial {
            base_color: Color::srgb(r, g, b),
            unlit: true,
            ..default()
        })
    });
    commands.insert_resource(Palette {
        node,
        origin,
        materials,
    });
    commands.spawn((
        Camera3d::default(),
        Projection::Perspective(PerspectiveProjection {
            far: 8000.0,
            ..default()
        }),
        Orbit::default_transform(),
        OrbitCam,
    ));
}

impl Orbit {
    fn default_transform() -> Transform {
        Transform::from_xyz(40.0, 30.0, 40.0).looking_at(Vec3::ZERO, Vec3::Y)
    }
}

fn apply_camera(orbit: Res<Orbit>, mut cameras: Query<&mut Transform, With<OrbitCam>>) {
    if let Ok(mut transform) = cameras.single_mut() {
        *transform = orbit.eye();
    }
}

fn keys(
    mut session: ResMut<Session>,
    keys: Res<ButtonInput<KeyCode>>,
    mut orbit: ResMut<Orbit>,
    mut motion: MessageReader<MouseMotion>,
    mut wheel: MessageReader<MouseWheel>,
    mouse: Res<ButtonInput<MouseButton>>,
    blocks: Res<UiBlocks>,
) {
    let motions: Vec<Vec2> = motion.read().map(|ev| ev.delta).collect();
    let wheels: Vec<(f32, bool)> = wheel
        .read()
        .map(|ev| (ev.y, matches!(ev.unit, MouseScrollUnit::Pixel)))
        .collect();
    if !blocks.keyboard {
        let max_time = session.sim.max_time;
        let mut time = session.filter.time;
        let mut mask = session.filter.color_mask;
        let mut lines = session.filter.show_lines;
        let mut axes = session.filter.show_axes;
        if keys.just_pressed(KeyCode::ArrowLeft) {
            time = (time - 1).max(0);
        }
        if keys.just_pressed(KeyCode::ArrowRight) {
            time = (time + 1).min(max_time);
        }
        if keys.just_pressed(KeyCode::Home) {
            time = 0;
        }
        if keys.just_pressed(KeyCode::End) {
            time = max_time;
        }
        for (key, bit) in [
            (KeyCode::Digit1, 1u8),
            (KeyCode::Digit2, 2),
            (KeyCode::Digit3, 4),
            (KeyCode::Digit4, 8),
        ] {
            if keys.just_pressed(key) {
                mask ^= bit;
            }
        }
        if keys.just_pressed(KeyCode::KeyL) {
            lines = !lines;
        }
        if keys.just_pressed(KeyCode::KeyX) {
            axes = !axes;
        }
        if time != session.filter.time
            || mask != session.filter.color_mask
            || lines != session.filter.show_lines
            || axes != session.filter.show_axes
        {
            session.filter.time = time;
            session.filter.color_mask = mask;
            session.filter.show_lines = lines;
            session.filter.show_axes = axes;
            session.filter.revision = session.filter.revision.wrapping_add(1);
        }
    }
    if blocks.pointer {
        return;
    }
    let orbit = orbit.as_mut();
    if mouse.pressed(MouseButton::Left) {
        for delta in &motions {
            orbit.yaw -= delta.x * 0.005;
            orbit.pitch = (orbit.pitch - delta.y * 0.005).clamp(-1.45, 1.45);
        }
    }
    if mouse.pressed(MouseButton::Right) || mouse.pressed(MouseButton::Middle) {
        let transform = orbit.eye();
        let right = *transform.right();
        let up = *transform.up();
        for delta in &motions {
            let scale = orbit.distance * 0.0015;
            orbit.focus += (-right * delta.x + up * delta.y) * scale;
        }
    }
    for (y, pixel) in &wheels {
        let dy = if *pixel { *y / 16.0 } else { *y };
        orbit.distance = (orbit.distance * (1.0 - dy * 0.08)).clamp(2.0, 4000.0);
    }
}

fn sync_scene(
    mut commands: Commands,
    session: Res<Session>,
    palette: Option<Res<Palette>>,
    mut drawn: ResMut<Drawn>,
    existing: Query<Entity, With<SceneMarker>>,
    mut applied: Local<u64>,
) {
    let Some(palette) = palette else {
        return;
    };
    if *applied == session.filter.revision {
        return;
    }
    *applied = session.filter.revision;
    for entity in &existing {
        commands.entity(entity).despawn();
    }
    let frame = build_frame(&session.sim.nodes, &session.filter, session.sim.max_coord);
    for marker in &frame.markers {
        let mesh = if marker.origin {
            palette.origin.clone()
        } else {
            palette.node.clone()
        };
        commands.spawn((
            Mesh3d(mesh),
            MeshMaterial3d(palette.material(marker.color)),
            Transform::from_xyz(marker.position[0], marker.position[1], marker.position[2]),
            SceneMarker,
        ));
    }
    drawn.0 = frame;
}

fn draw_gizmos(mut gizmos: Gizmos, drawn: Res<Drawn>, session: Res<Session>) {
    if session.filter.show_axes {
        for seg in &drawn.0.axes {
            gizmos.line(vec3(seg.a), vec3(seg.b), line_color(seg.color, true));
        }
    }
    if session.filter.show_lines {
        for seg in &drawn.0.segments {
            gizmos.line(vec3(seg.a), vec3(seg.b), line_color(seg.color, false));
        }
    }
}

fn vec3(p: [f32; 3]) -> Vec3 {
    Vec3::new(p[0], p[1], p[2])
}

fn line_color(color: u8, axis: bool) -> Color {
    if color == 0 {
        return Color::srgb(0.28, 0.30, 0.34);
    }
    let [r, g, b] = color_rgb(color);
    if axis {
        Color::srgb(r, g, b)
    } else {
        Color::srgb(r * 0.75, g * 0.75, b * 0.75)
    }
}

fn panel(
    mut contexts: EguiContexts,
    mut session: ResMut<Session>,
    drawn: Res<Drawn>,
    mut blocks: ResMut<UiBlocks>,
) -> Result {
    let ctx = contexts.ctx_mut()?;
    let max_time = session.sim.max_time;
    let name = session.sim.name.clone();
    let run_name = session.sim.run_name.clone();
    let mut changed = false;
    egui::Window::new("Simulation")
        .default_pos([12.0, 12.0])
        .resizable(false)
        .show(ctx, |ui| {
            ui.label(name);
            if let Some(run) = run_name {
                ui.label(run);
            }
            ui.label(format!("time {} / {max_time}", session.filter.time));
            changed |= ui
                .add(
                    egui::Slider::new(&mut session.filter.time, 0..=max_time).text("creation time"),
                )
                .changed();
            ui.label(format!(
                "nodes {}   origins {}   lines {}",
                drawn.0.node_count(),
                drawn.0.origin_count(),
                drawn.0.line_count()
            ));
            ui.horizontal(|ui| {
                changed |= color_box(ui, &mut session.filter.color_mask, 1, "red");
                changed |= color_box(ui, &mut session.filter.color_mask, 2, "green");
                changed |= color_box(ui, &mut session.filter.color_mask, 4, "blue");
                changed |= color_box(ui, &mut session.filter.color_mask, 8, "yellow");
            });
            changed |= ui
                .checkbox(&mut session.filter.show_lines, "lines (L)")
                .changed();
            changed |= ui
                .checkbox(&mut session.filter.show_axes, "axes (X)")
                .changed();
            ui.label("orbit: left drag    pan: right drag    zoom: scroll");
            ui.label("time: left / right, Home / End    colors: 1 2 3 4");
        });
    if changed {
        session.filter.revision = session.filter.revision.wrapping_add(1);
    }
    blocks.pointer = ctx.egui_wants_pointer_input();
    blocks.keyboard = ctx.egui_wants_keyboard_input();
    Ok(())
}

fn color_box(ui: &mut egui::Ui, mask: &mut u8, bit: u8, label: &str) -> bool {
    let mut on = *mask & bit != 0;
    let changed = ui.checkbox(&mut on, label).changed();
    if changed {
        if on {
            *mask |= bit;
        } else {
            *mask &= !bit;
        }
    }
    changed
}

fn grab_screenshot(
    mut commands: Commands,
    mut shot: ResMut<Shot>,
    capturing: Query<(), With<Capturing>>,
    mut exit: MessageWriter<AppExit>,
) {
    let Some(path) = shot.path.clone() else {
        return;
    };
    shot.frames += 1;
    if !shot.fired && shot.frames >= 12 {
        commands
            .spawn(Screenshot::primary_window())
            .observe(save_to_disk(path.clone()));
        shot.fired = true;
        return;
    }
    if shot.fired && capturing.is_empty() && path.is_file() {
        exit.write(AppExit::Success);
        return;
    }
    if shot.started.elapsed().as_secs() > 180 {
        eprintln!("qsm-view: screenshot timed out ({})", path.display());
        exit.write(AppExit::Error(NonZero::new(1).unwrap()));
    }
}
