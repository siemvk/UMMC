#!/bin/bash

echo "Did handoff to bash script"

if ! command -v steamcmd &> /dev/null; then
    echo "Error: steamcmd is not installed or not found in PATH."
    echo "Please install steamcmd (e.g. 'brew install steamcmd') and try again."
    exit 1
fi

APP_ID="${APP_ID:-391540}"
TARGET_DIR="${TARGET_DIR:-$HOME/UMMC/windows/}"
GAME_NAME="${GAME_NAME:-Undertale}"

if [ -z "$USERNAME" ]; then
    read -p "Enter Steam Username: " USERNAME
fi

echo "Starting $GAME_NAME Windows download via SteamCMD (App ID $APP_ID)..."

steamcmd \
  +@sSteamCmdForcePlatformType windows \
  +force_install_dir "$TARGET_DIR" \
  +login "$USERNAME" \
  +app_update "$APP_ID" validate \
  +quit

echo "Download process complete! You can now use `--win` when defining mods"
exit 0