# UMMC

UMMC is a simple tool for moding undertale. It allows you to easily patch the game and add full mods to it. 

## Features
- Making backups of the original game files
- (Planned) Manage save files
- Install mods to the game
- download the windows version of undertale from steam easily

## Installation

As of now, UMMC is only tested on macOS, but it should work on linux too. 

You have to build UMMC from source, as I am not emotionally ready to try and get it on homebrew.
```bash
go mod download
go build
```
The non Go dependencies UMMC needs are:
- [go](https://go.dev/doc/install) (for building UMMC) [brew install go](https://formulae.brew.sh/formula/go)
- [steamcmd](https://developer.valvesoftware.com/wiki/SteamCMD) (only if you want to easily download the windows version of undertale) [brew install steamcmd](https://formulae.brew.sh/cask/steamcmd)
- [xdelta3](https://github.com/jmacd/xdelta) (for patching the game files, needed for modding) [brew install xdelta3](https://formulae.brew.sh/formula/xdelta)



## Usage

After you have installed UMMC, install undertale via steam. Next make a initial backup of your game files to prevent unnecessary redownloads later. 
you can do this by running:
```bash
UMMC backup create
```
If you dont have the latest version (1.0.8) of undertale, you can specify the version you are backing up by running:
```bash
UMMC backup create --version <version>
```
If you are getting an error about files already existing, check that you are not using ~/UMMC.
> [!NOTE]
> this uses valve's offical steamcmd tool to download the windows version of undertale. I can assure you that you can trust this tool, but if you are still concerned, you can follow the guide [here]() to do it without steamcmd.
Next we will download the windows version of undertale as most mods are made for the windows `data.win` file instead of the macos `game.ios` file. You can do this by running:
```bash
UMMC download-win -u <your steam username>
```
this wil prompt you for your steam login password and steam guard code if you have it enabled.
If you are getting an error about steamcmd not being found, install steamcmd and add it to your path. You can find instructions for this [here](https://developer.valvesoftware.com/wiki/SteamCMD). Alternatively, you can folow the guide [here]() to do it without steamcmd.
Now you can make a patched version of the game and make a backup of the patched version. You can do this by running:
```bash
UMMC inject
UMMC backup create 
```
Now you can install mods to the game. You can do this by running:
```bash
UMMC mods add <path to mod>
```
you can do this to see all the options for adding mods:
```bash
UMMC mods add --help
```
> [!TIP]
> You can add `--help` to any command to see all the options for that command. For example, you can run `UMMC mods load --help` to see all the options for loading mods.
after you have added a mod, you can install it to the game by running:
```bash
UMMC mods load <mod name or unique id>
```
You can see all the available mods and their id's by running:
```bash
UMMC mods list
```

## What makes a mod not work with UMMC on macOS
UMMC is a bit more then just a simple patcher as it can download windows versions of undertale and patch them to work on macOS. However, some mods are not compatible with UMMC on macOS. This is usually because the mod uses a custom undertale.exe file, which i cannot patch to work on macOS. If a mod uses a custom undertale.exe file, its unlikely that it will work with UMMC on macOS. You can check if the mod has a macos version, but if it does not, it wont work ):. at that point, you can 1. ask the mod author polietly if they can make a macos version, or 2. run it via wine, whisky or a virtual machine.

## Mods confirmed to work with UMMC on macOS
- [C!UNDERTALE - REDUX UPDATE](https://gamebanana.com/mods/601488)
- [Undertale - Just have fun](https://gamebanana.com/mods/download/542409#FileInfo_1759398) (See [my notes](./notes/JHF.md) for instructions.)
- [Undertale Connect v1.3.4](https://landimizer.itch.io/ut-connect)

## Mods that sadly do not work with UMMC on macOS
- [Undertale Together](https://www.moddb.com/mods/undertale-together) (the other versions of this mod also dont work.)
- [Undertale Red & Yellow](https://gamejolt.com/games/undertale-red-yellow/877387)