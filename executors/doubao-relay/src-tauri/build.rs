fn main() {
    // `app.get_version` 的 rustcVersion:编译期执行 `rustc -V`("rustc 1.94.0 (...)" → "1.94.0")。
    let rustc = std::env::var("RUSTC").unwrap_or_else(|_| "rustc".into());
    let version = std::process::Command::new(rustc)
        .arg("-V")
        .output()
        .ok()
        .and_then(|o| String::from_utf8(o.stdout).ok())
        .and_then(|s| s.split_whitespace().nth(1).map(str::to_string))
        .unwrap_or_default();
    println!("cargo:rustc-env=DOUBAO_RELAY_RUSTC_VERSION={version}");
    println!("cargo:rerun-if-env-changed=RUSTC");
    if std::env::var("CARGO_CFG_TARGET_OS").as_deref() == Ok("windows") {
        // `tao` imports TaskDialogIndirect. Test harnesses do not receive the
        // desktop app manifest from Tauri, so explicitly select comctl32 v6.
        println!(
            "cargo:rustc-link-arg=/manifestdependency:type='win32' name='Microsoft.Windows.Common-Controls' version='6.0.0.0' processorArchitecture='*' publicKeyToken='6595b64144ccf1df' language='*'"
        );
    }
    tauri_build::build()
}
