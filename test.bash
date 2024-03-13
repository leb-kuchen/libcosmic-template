#!/usr/bin/env bash 
# todo run simultaneously
./libcosmic-template --no-confirm
cd cosmic-applet-example
cargo c
cd ..
rm -rf  cosmic-applet-example
./libcosmic-template --no-confirm --config=false
cd cosmic-applet-example
cargo c
cd ..
rm -rf cosmic-applet-example
./libcosmic-template --no-confirm --no-example
cd cosmic-applet-example
cargo c
cd ..
rm -rf cosmic-applet-example

