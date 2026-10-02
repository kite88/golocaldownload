#!/usr/bin/env bash
#
# Launcher for the golocaldownload executable.
#
# Keep this file next to the binary (the release archives already ship it) and
# run it in the foreground; every argument is passed straight through to the
# program:
#
#     ./start.sh                      run with the embedded defaults
#     ./start.sh -config /etc/golocaldownload/env.ini
#
# Ctrl+C stops it - the program waits for in-flight downloads to finish first.
# To run it as a background service, start the binary from systemd / supervisor
# (or nohup) instead of using this script.

set -euo pipefail

cd "$(dirname "$0")"

bin=golocaldownload
if [ ! -f "$bin" ]; then
    echo "start.sh: $bin not found in $PWD" >&2
    echo "          keep this script in the same directory as the executable." >&2
    exit 1
fi

# Restore the executable bit: zip archives and some copy methods drop it.
if [ ! -x "$bin" ]; then
    chmod +x "$bin"
fi

exec "./$bin" "$@"
