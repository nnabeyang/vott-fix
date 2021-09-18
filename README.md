# VoTT-Fix

VoTT-Fix is a tool to fix [VoTT](https://github.com/microsoft/VoTT) files (especially for local storage project).

# Usage
```sh
vott-fix -target /path/to/target # target directory contains *.vott and *-asset.json
```

# Example

- Download the sample VoTT project at: https://drive.google.com/file/d/1djlcSk6B8QMJRp7mOfCBXrrkqiVqLY3N/view?usp=sharing
- Add Security Token to VoTT
  - Copy token in `sample_project/token.txt`.
  - Push Button to open Application Settings.
  - Add the token whose name is "Sample Token".
<img class="screen-shot" src="https://raw.githubusercontent.com/nnabeyang/vott-fix/master/asset/image/01.png" width="500px" style="max-width: 500px; border: 1px solid rgba(0,0,0,0.1); box-shadow: 1px 1px 1px rgba(0,0,0,0.5);">
<img class="screen-shot" src="https://raw.githubusercontent.com/nnabeyang/vott-fix/master/asset/image/02.png" width="500px" style="max-width: 500px; border: 1px solid rgba(0,0,0,0.1); box-shadow: 1px 1px 1px rgba(0,0,0,0.5);">

- fix VoTT files with vott-fix
```sh
vott-fix -target /path/to/sample_project/project/dist
```
- Open `Sample.vott` by clicking "Open Local Project".
<img class="screen-shot" src="https://raw.githubusercontent.com/nnabeyang/vott-fix/master/asset/image/03.png" width="500px" style="max-width: 500px; border: 1px solid rgba(0,0,0,0.1); box-shadow: 1px 1px 1px rgba(0,0,0,0.5);">

- When the app transitions to the screen as shown below, the project has been loaded successfully.
<img class="screen-shot" src="https://raw.githubusercontent.com/nnabeyang/vott-fix/master/asset/image/04.png" width="500px" style="max-width: 500px; border: 1px solid rgba(0,0,0,0.1); box-shadow: 1px 1px 1px rgba(0,0,0,0.5);">

# License
MIT

# Author
[Noriaki Watanabe@nnabeyang](https://twitter.com/nnabeyang)
