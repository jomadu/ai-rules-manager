#!/bin/bash
cd ..
make build
rm -rf sandbox
mkdir sandbox
cp bin/arm sandbox/
cd sandbox
./arm config add registry default https://github.com/PatrickJS/awesome-cursorrules --type git
./arm config add channel q --directories .amazonq/rules
./arm install python-rules@main --patterns "rules-new/python.mdc"
cat arm.lock
cat arm.json
cat .armrc
tree -a
./arm outdated
tree -a -L 4 ~/.arm/cache
cat ~/.arm/cache/ruleset-map.json
cat ~/.arm/cache/registry-map.json
cat ~/.arm/cache/registries/4d3bb29db1e2e8831cd5a243ad652a476751fa22e7003ab6b80cf62c35053aec/cache-info.json
cat ~/.arm/cache/registries/4d3bb29db1e2e8831cd5a243ad652a476751fa22e7003ab6b80cf62c35053aec/metadata.json
cat ~/.arm/cache/registries/4d3bb29db1e2e8831cd5a243ad652a476751fa22e7003ab6b80cf62c35053aec/versions.json
