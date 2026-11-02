//! 应用数据目录布局(07-data-model "文件布局")与文件删除辅助。
//!
//! ```text
//! <root>/relay.db            SQLite 主库(+ -wal / -shm)
//! <root>/uploads/{id}.{ext}  预上传的参考图/视频
//! <root>/results/{task_id}/  结果文件(image_1.png / video_1.mp4 …)
//! <root>/logs/               日志(macOS 实际按 spec 10 落在 ~/Library/Logs,此处仅保证目录存在)
//! ```

use std::io;
use std::path::{Component, Path, PathBuf};

pub const DB_FILE: &str = "relay.db";
pub const UPLOADS_DIR: &str = "uploads";
pub const RESULTS_DIR: &str = "results";
pub const LOGS_DIR: &str = "logs";

/// 创建 root 及 uploads/ results/ logs/ 三个子目录(幂等)。
pub fn ensure_layout(root: &Path) -> io::Result<()> {
    for dir in [UPLOADS_DIR, RESULTS_DIR, LOGS_DIR] {
        std::fs::create_dir_all(root.join(dir))?;
    }
    Ok(())
}

/// `<root>/relay.db`。
pub fn db_path(root: &Path) -> PathBuf {
    root.join(DB_FILE)
}

/// `<root>/results/{task_id}`。
pub fn result_dir(root: &Path, task_id: &str) -> PathBuf {
    root.join(RESULTS_DIR).join(task_id)
}

/// `<root>/uploads/{id}.{ext}`;`ext` 可带或不带前导点,空则不加扩展名。
pub fn upload_path(root: &Path, id: &str, ext: &str) -> PathBuf {
    let ext = ext.trim_start_matches('.');
    let name = if ext.is_empty() {
        id.to_string()
    } else {
        format!("{id}.{ext}")
    };
    root.join(UPLOADS_DIR).join(name)
}

/// 相对路径是否可以安全地拼到 root 下(纯相对、不含 `..`、非绝对)。
pub fn is_safe_relative(rel: &Path) -> bool {
    !rel.as_os_str().is_empty()
        && rel
            .components()
            .all(|c| matches!(c, Component::Normal(_) | Component::CurDir))
}

/// 删除文件;不存在视为成功。返回是否真的删掉了一个文件。
pub fn remove_file_if_exists(path: &Path) -> io::Result<bool> {
    match std::fs::remove_file(path) {
        Ok(()) => Ok(true),
        Err(e) if e.kind() == io::ErrorKind::NotFound => Ok(false),
        Err(e) => Err(e),
    }
}

/// 递归删除目录;不存在视为成功。返回删掉的普通文件数。
pub fn remove_dir_if_exists(path: &Path) -> io::Result<u64> {
    if !path.exists() {
        return Ok(0);
    }
    let files = count_files(path)?;
    std::fs::remove_dir_all(path)?;
    Ok(files)
}

fn count_files(dir: &Path) -> io::Result<u64> {
    let mut n = 0;
    for entry in std::fs::read_dir(dir)? {
        let entry = entry?;
        let ty = entry.file_type()?;
        if ty.is_dir() {
            n += count_files(&entry.path())?;
        } else {
            n += 1;
        }
    }
    Ok(n)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn layout_paths_follow_spec() {
        let root = Path::new("/tmp/app");
        assert_eq!(db_path(root), Path::new("/tmp/app/relay.db"));
        assert_eq!(
            result_dir(root, "01HQZ9"),
            Path::new("/tmp/app/results/01HQZ9")
        );
        assert_eq!(
            upload_path(root, "01HQZ8", "png"),
            Path::new("/tmp/app/uploads/01HQZ8.png")
        );
        assert_eq!(
            upload_path(root, "01HQZ8", ".jpg"),
            Path::new("/tmp/app/uploads/01HQZ8.jpg")
        );
        assert_eq!(
            upload_path(root, "01HQZ8", ""),
            Path::new("/tmp/app/uploads/01HQZ8")
        );
    }

    #[test]
    fn ensure_layout_is_idempotent() {
        let dir = tempfile::tempdir().unwrap();
        ensure_layout(dir.path()).unwrap();
        ensure_layout(dir.path()).unwrap();
        for d in ["uploads", "results", "logs"] {
            assert!(dir.path().join(d).is_dir(), "{d}");
        }
    }

    #[test]
    fn safe_relative_rejects_traversal_and_absolute() {
        assert!(is_safe_relative(Path::new("results/01/image_1.png")));
        assert!(!is_safe_relative(Path::new("../x")));
        assert!(!is_safe_relative(Path::new("results/../../x")));
        assert!(!is_safe_relative(Path::new("/etc/passwd")));
        assert!(!is_safe_relative(Path::new("")));
    }

    #[test]
    fn remove_helpers_tolerate_missing_and_count_files() {
        let dir = tempfile::tempdir().unwrap();
        let missing = dir.path().join("nope");
        assert!(!remove_file_if_exists(&missing).unwrap());
        assert_eq!(remove_dir_if_exists(&missing).unwrap(), 0);

        let sub = dir.path().join("r").join("t1");
        std::fs::create_dir_all(sub.join("nested")).unwrap();
        std::fs::write(sub.join("a.png"), b"x").unwrap();
        std::fs::write(sub.join("nested").join("b.png"), b"y").unwrap();
        assert_eq!(remove_dir_if_exists(&sub).unwrap(), 2);
        assert!(!sub.exists());

        let f = dir.path().join("f.txt");
        std::fs::write(&f, b"z").unwrap();
        assert!(remove_file_if_exists(&f).unwrap());
        assert!(!f.exists());
    }
}
