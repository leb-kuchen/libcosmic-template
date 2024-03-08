
## About
Generates the boilerplate for [libcosmic](github.com/pop-os/libcosmic) applications.
Currently only applets are supported.
Includes the justfile for installing the applet and icons aswell as translation support.


## Install 
Go(>=1.22) is required.
```sh
go install github.com/leb-kuchen/libcosmic-template
```
## Usage
- icon string
    Icon name (default "display-symbolic")
- icon-files string
    path to icon files(Seperated by whitespace)
- id string
    App ID (default "com.system76.CosmicAppletExample")
- name string
    App name (default "cosmic-applet-example")

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
libcosmic-template -id org.example.com -name example-example-applet -icon "some-icon"
```


