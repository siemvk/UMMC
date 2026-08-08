#!/bin/bash

echo "Did handoff to bash script"

if ! command -v steamcmd &> /dev/null; then
    echo "Error: steamcmd is not installed or not found in PATH."
    echo "Please install steamcmd (e.g. 'brew install steamcmd') and try again."
    exit 1
fi

TARGET_DIR="${TARGET_DIR:-$HOME/UMMC/windows/}"

if [ -z "$USERNAME" ]; then
    read -p "Enter Steam Username: " USERNAME
fi

echo "Starting Undertale Windows download via SteamCMD..."

steamcmd \
  +@sSteamCmdForcePlatformType windows \
  +force_install_dir "$TARGET_DIR" \
  +login "$USERNAME" \
  +app_update 391540 validate \
  +quit

echo "Download process complete! you can now use `--win` when defining mods"
exit 0