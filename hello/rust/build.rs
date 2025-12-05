use std::fs;
use std::path::Path;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source_dir = Path::new("../src/main/protobuf");
    let dest_dir = Path::new("proto");

    fs::create_dir_all(dest_dir)?;

    if source_dir.exists() {
        for entry in fs::read_dir(source_dir)? {
            let entry = entry?;
            let path = entry.path();
            if path.extension().map_or(false, |ext| ext == "proto") {
                let file_name = path.file_name().unwrap();
                let dest_path = dest_dir.join(file_name);
                fs::copy(&path, &dest_path)?;
                println!("cargo:rerun-if-changed={}", path.display());
            }
        }
    }

    // Step 2: Generate Rust code from proto files
    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(
            &["proto/hello.proto"],
            &["proto"],
        )?;
    Ok(())
}
