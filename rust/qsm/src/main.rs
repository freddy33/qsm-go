fn main() {
    let mut args = std::env::args().skip(1);
    let Some(name) = args.next() else {
        usage();
    };
    if name.starts_with('-') {
        usage();
    }
    let mut fresh = false;
    for arg in args {
        if arg == "--new" && !fresh {
            fresh = true;
        } else {
            usage();
        }
    }
    let result = if fresh {
        m3data::run_new(m3data::default_data_root(), &name, m3data::config_root())
    } else {
        m3data::run_named(m3data::default_data_root(), &name)
    };
    match result {
        Ok(report) => println!("{report}"),
        Err(err) => {
            eprintln!("qsm: {err}");
            std::process::exit(1);
        }
    }
}

fn usage() -> ! {
    eprintln!("usage: qsm <simulation name> [--new]");
    std::process::exit(2);
}
