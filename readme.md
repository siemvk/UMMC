# UMMC

UMMC is a simple tool for moding undertale. It allows you to easily patch the game and add full mods to it. 

## Features
- Making backups of the original game files
- (Planned) Manage save files
- (Planned) Install mods to the game
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
> ![NOTE]
> this uses valve's offical steamcmd tool to download the windows version of undertale. I can assure you that you can trust this tool, but if you are still concerned, you can follow the guide [here]() to do it without steamcmd.
Next we will download the windows version of undertale as most mods are made for the windows `data.win` file instead of the macos `game.ios` file. You can do this by running:
```bash
UMMC download-win -u <your steam username>
```
this wil prompt you for your steam login password and steam guard code if you have it enabled.
If you are getting an error about steamcmd not being found, install steamcmd and add it to your path. You can find instructions for this [here](https://developer.valvesoftware.com/wiki/SteamCMD). Alternatively, you can folow the guide [here]() to do it without steamcmd.

## Mods confirmed to work with UMMC on macOS
- [C!UNDERTALE - REDUX UPDATE](https://gamebanana.com/mods/601488)
- [Undertale - Just have fun](https://gamebanana.com/mods/download/542409#FileInfo_1759398) (See [my notes](./notes/JHF.md) for instructions.)

## Mods that sadly do not work with UMMC on macOS