fn main() {
    let mut args = std::env::args().skip(1);
    let Some(name) = args.next() else {
        eprintln!("usage: qsm <simulation name>");
        std::process::exit(2);
    };
    if args.next().is_some() {
        eprintln!("usage: qsm <simulation name>");
        std::process::exit(2);
    }
    match m3data::run_named(m3data::default_data_root(), &name) {
        Ok(report) => println!("{report}"),
        Err(err) => {
            eprintln!("qsm: {err}");
            std::process::exit(1);
        }
    }
}
