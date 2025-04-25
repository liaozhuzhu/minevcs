# Automatic Cloud Saving For Minecraft Java Edition

<img src="./build/appicon.png" width="100">

## About

Enabling automatic cloud save for Minecraft Java Edition. 

## Future Goals

- Speed up upload

- Detect if change has actually occurred before pushing world

Generating the `.dmg` (note to myself):
```bash
create-dmg --volname "MineVCS Installer" --window-pos 200 120 --window-size 500 300 --icon-size 100 --icon "minevcs.app" 125 150 --app-drop-link 375 150 MineVCS-Installer.dmg minevcs.app
```