#!/bin/bash
cd ..
make build
rm -rf sandbox
rm -rf ~/.arm/cache
mkdir sandbox
cp bin/arm sandbox/
cd sandbox
./arm config add registry default https://github.com/PatrickJS/awesome-cursorrules --type git
./arm config add channel q --directories .amazonq/rules
./arm install python-rules@main --patterns "rules-new/python.mdc"
cat arm.lock
cat arm.json
cat .armrc.json
tree -a
./arm outdated
tree -a -L 4 ~/.arm/cache
