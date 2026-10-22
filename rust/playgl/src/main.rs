//! Pack summary for the static reference tables.
//!
//! The interactive window is `qsm-view <simulation name>`.

fn main() {
    let pack = m3point::PointPackData::shared();
    println!("QSM pack");
    println!("connections: {}", pack.connections.len());
    println!("trios: {}", pack.trios.len());
    println!("growth contexts: {}", pack.growth_contexts.len());
    println!("cubes: {}", pack.cube_ids.len());
    println!(
        "path builders: {}",
        pack.path_builders.iter().filter(|b| b.is_some()).count()
    );
    println!();
    println!("viewer: qsm-view <simulation name>");
    let _ = m3gl::ObjectType::NodeActive;
}
