//! Text stand-in for the Go `playgl` OpenGL viewer.
//!
//! TODO: replace this process with a `glow`+`glutin` or `wgpu` window that
//! draws [`m3gl`] elements. The Go viewer lives in `ui/playgl` and `ui/m3gl`
//! and is left unchanged.

fn main() {
    let pack = m3point::PointPackData::shared();
    println!("QSM playgl (Rust stub)");
    println!("connections: {}", pack.connections.len());
    println!("trios: {}", pack.trios.len());
    println!("growth contexts: {}", pack.growth_contexts.len());
    println!("cubes: {}", pack.cube_ids.len());
    println!(
        "path builders: {}",
        pack.path_builders.iter().filter(|b| b.is_some()).count()
    );
    println!();
    println!("TODO: the OpenGL viewer is not ported.");
    println!("The Go UI uses GLFW and OpenGL (ui/playgl, ui/m3gl).");
    println!("A Rust window should use glow+glutin or wgpu and the drawing");
    println!("types in the m3gl crate (filters, colors, axes, nodes).");
    let _ = m3gl::ObjectType::NodeActive;
}
