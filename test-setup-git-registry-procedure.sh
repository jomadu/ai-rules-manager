# install git repo test procedure

make build
rm -rf sandbox
mkdir sandbox
cd sandbox
../bin/arm version
../bin/arm install
../bin/arm clean all --force
../bin/arm config add registry default https://github.com/PatrickJS/awesome-cursorrules --type=git
../bin/arm config add channel q --directories ./amazonq/rules
../bin/arm install python --patterns "rules-new/python.mdc"
