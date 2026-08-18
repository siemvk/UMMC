# UMMC

UMMC is a simple tool for modding Undertale & Deltarune on macOS. It allows you to easily patch the games, create backups, and add full mods for any chapter.

> [!NOTE]
> The deltarune support is not great yet, for example, it does not yet support modding the full game, only the chapters. If you want to mod the full game, you can use UMMC to mod each chapter separately for now.

## Features
- Making backups of the original game files (Undertale & Deltarune per chapter)
- Deltarune support with chapter selection (`--deltarune 1`, `-d 2`, etc.)
- (Planned) Manage save files
- Install mods to Undertale and Deltarune chapters
- Download Windows versions of Undertale or Deltarune via SteamCMD easily

## Installation

As of now, UMMC is tested on macOS with the steam version of the game, but it should work on Linux too.

### Quick Install

Run this command in your terminal:
```bash
curl -fsSL https://raw.githubusercontent.com/siemvk/UMMC/main/install.sh | bash
```

Or if you have already cloned the repository locally:
```bash
./install.sh
```

This script will:
- Check for required dependencies (`go`, `xdelta3`, `steamcmd`) and install any missing ones via Homebrew on macOS.
- Build the `UMMC` binary.
- Install `UMMC` to `~/.local/bin` (or a custom directory specified with `-b / --bin-dir`).

### Manual Build

If you prefer to build manually:
```bash
go mod download
go build -o UMMC .
```

The non-Go dependencies UMMC needs are:
- [go](https://go.dev/doc/install) (for building UMMC) [`brew install go`](https://formulae.brew.sh/formula/go)
- [steamcmd](https://developer.valvesoftware.com/wiki/SteamCMD) [`brew install --cask steamcmd`](https://formulae.brew.sh/cask/steamcmd) see [this](./notes/nosteamcli.md) if you don't trust steamcmd
- [xdelta3](https://github.com/jmacd/xdelta) [`brew install xdelta3`](https://formulae.brew.sh/formula/xdelta)

## Usage

### Undertale
After you have installed UMMC, make an initial backup of your game files:
```bash
UMMC backup create
```
Download Windows version (if modding with Windows `data.win`):
```bash
UMMC download-win -u <your steam username>
UMMC inject
```
see [this](./notes/nosteamcli.md) if you don't trust steamcmd or my program with your steam credentials.
> [!TIP]
> just use my thing, trust me its safe, I don't store your credentials anywhere.

Add and play Undertale mods:
```bash
UMMC mods create <path to mod>
UMMC mods play <mod name or id>
```

### Deltarune Support
UMMC fully supports Deltarune with chapter-specific modding. Use `--deltarune <chapter>` or `-d <chapter>`:

- **Create a Deltarune Backup**:
  ```bash
  UMMC backup create --deltarune 1
  UMMC backup create -d 2
  ```

- **Add a Deltarune Mod**:
  ```bash
  UMMC mods create <path to mod> --deltarune 1
  UMMC mods create <path to mod> -d 5
  ```

- **List Mods or Backups (Filtered by Game)**:
  ```bash
  UMMC mods list --undertale  # or -u
  UMMC mods list --deltarune  # or -d
  UMMC backup list -u
  UMMC backup list -d 2
  ```

- **Play a Deltarune Mod**:
  ```bash
  UMMC mods play <mod name or id>
  ```
  *(UMMC automatically detects the chapter from the database!)*

- **Download Deltarune Windows via SteamCMD & Inject**:
  ```bash
  UMMC download-win -U <username> --deltarune
  UMMC inject --deltarune 1
  ```

> [!TIP]
> You can add `--help` to any command to see all available options. For example, `UMMC mods create --help`.

## What makes a mod not work with UMMC on macOS
UMMC is a bit more than just a simple patcher as it can download windows versions of Undertale/Deltarune and patch them to work on macOS. However, some mods are not compatible with UMMC on macOS. This is usually because the mod uses a custom `.exe` file, which cannot be patched to work on macOS. If a mod uses a custom `.exe` file, it's unlikely that it will work with UMMC natively on macOS (at that point, consider Wine, Whisky, or a VM).

Maybe custom language files?????

## Mods confirmed to work with UMMC on macOS
- [C!UNDERTALE - REDUX UPDATE](https://gamebanana.com/mods/601488)
- [Undertale - Just have fun](https://gamebanana.com/mods/download/542409#FileInfo_1759398) (See [my notes](./notes/JHF.md) for instructions.)
- [Undertale Connect v1.3.4](https://landimizer.itch.io/ut-connect)
- [Undertale, but you play as Sans](https://gamebanana.com/mods/514736)
- [Undertale with a Gaster Blaster](https://gamebanana.com/mods/428457)
- [Undertale Random Souls](https://gamebanana.com/mods/514890)
- [Undertale Mouse Mod](https://gamebanana.com/mods/514892)
- [UNDERTALE: Wind Challenge](https://gamebanana.com/mods/565434)

## Mods that sadly do not work with UMMC on macOS
- [Undertale Together](https://www.moddb.com/mods/undertale-together) (the other versions of this mod also dont work.) (unknown why its broken)
- [Undertale Red & Yellow](https://gamejolt.com/games/undertale-red-yellow/877387) (.exe file is not compatible with macOS)
- [UNDERTALE Hard Mode: Director's Cut](https://gamejolt.com/games/uthardmodedc/973954) (.exe file is not compatible with macOS)
- [Undertale but a Gaster Blaster spawns every second](https://gamebanana.com/mods/538168) (load in idk why it crashes, but it does)
- [UNDERTALE: Monster Arena](https://gamebanana.com/mods/691720) (.exe file is not compatible with macOS)