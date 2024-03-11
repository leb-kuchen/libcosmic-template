
## About
Generates the boilerplate for [libcosmic](https://github.com/pop-os/libcosmic) applications on the [COSMIC DE](https://github.com/pop-os/cosmic-epoch).
Currently only applets are supported.
Includes the justfile for installing the applet and icons aswell as translation support.


## Install 
Go(>=1.22) is required to build. Prebuilt binaries are available in the releases section.
```sh
go install github.com/leb-kuchen/libcosmic-template@latest
```
## Usage
  --icon string          Icon name (default "display-symbolic")
  --icon-files strings   path to icon files
  --id string            App ID (default "com.system76.CosmicAppletExample")
  -i, --interactive      Activate interactive mode
  -n, --name string      App name (default "cosmic-applet-example")

## Getting started
```sh
libcosmic-template
cd cosmic-applet-example
cargo b -r
sudo just install
```
The example applet should now appear in the panel settings.
## Example
```sh
libcosmic-template -id org.example1.com -name example-example-applet1 -icon "zoom-original-symbolic.svg" -icon-files "zoom-original-symbolic.svg"
```


