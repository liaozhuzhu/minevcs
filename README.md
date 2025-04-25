# Automatic Cloud Saving For Minecraft Java Edition

<p align="center">
  <img src="./build/appicon.png" width="200"/>
</p>

<p align="center">
  <img src="./assets/design.png" alt="design for minevcs"/>
</p>

## About

Enabling automatic cloud save for Minecraft Java Edition. **Currently only for MacOS**

## Future Goals

- Speed up upload
- Detect if change has actually occurred before pushing world
- Windows Support

Generating the `.dmg` (note to myself):
```bash
create-dmg --volname "MineVCS Installer" --window-pos 200 120 --window-size 500 300 --icon-size 100 --icon "minevcs.app" 125 150 --app-drop-link 375 150 MineVCS-Installer.dmg minevcs.app
