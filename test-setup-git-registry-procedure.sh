# install git repo test procedure

make build
rm -rf sandbox
mkdir sandbox
cd sandbox
../bin/arm version
../bin/arm install
../bin/arm clean all --force
../bin/arm config add registry sdlc-rules https://github.com/jomadu/sdlc-rules --type=git
../bin/arm config add channel q --directories ./amazonq/rules
../bin/arm install sdlc-rules/its-a-me@main --patterns "rules/its-a-me.md" --channels q
