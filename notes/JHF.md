# Undertale just have fun

> [!IMPORTANT]
> Make sure you get the offical macos version of this mod!

For this mod to work, you need to have these args when you add the mod to UMMC:
```
--install-to-app-root --macos-mod
```
You will also need to have the following file changes:

1. Delete all files not in either the root of the mod or the `Runner and Assets` folder
2. rename `Runner and Assets` to `Contents`
3. In your new `Contents` folder, move `Mac_Runner` into a new folder called `MacOS`

After you have done this, you can add the mod to UMMC (with the args I mentioned) and it should work.