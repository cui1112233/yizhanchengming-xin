"""Start the supplied pywebview application with a loopback CDP port.

Uses pywebview's existing REMOTE_DEBUGGING_PORT setting. Does not modify the
original application, inject page JavaScript, or change authentication/storage.
Run with the target application's own Python interpreter.
"""
import argparse
import os
from pathlib import Path
import runpy
import socket
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--app-dir", required=True)
    parser.add_argument("--port", type=int, default=9226)
    args = parser.parse_args()
    base = Path(args.app_dir).resolve(strict=True)
    entry = base / "run.py"
    if not entry.is_file() or not (base / "modules" / "main.pyd").is_file():
        raise SystemExit("The selected folder is not the expected application.")
    if not 1024 <= args.port <= 65535:
        raise SystemExit("Choose a non-privileged local port.")
    # Do not attach a new launch to someone else's existing debug listener.
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", args.port))
    for folder in (base, base / "modules", base / "packages"):
        sys.path.insert(0, str(folder))
    os.chdir(base)
    # The target's run.py sets GPU flags itself. Its installed renderer adds
    # REMOTE_DEBUGGING_PORT separately to WebView2 CreationProperties.
    import webview
    webview.settings["REMOTE_DEBUGGING_PORT"] = args.port
    sys.argv = [str(entry)]
    print(f"Starting original run.py with local CDP port {args.port}", flush=True)
    runpy.run_path(str(entry), run_name="__main__")


if __name__ == "__main__":
    main()
