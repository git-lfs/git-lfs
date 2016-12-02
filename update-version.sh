#!/bin/bash

VERSION_STRING=$1
VERSION_ARRAY=( ${VERSION_STRING//./ } )
VERSION_MAJOR=${VERSION_ARRAY[0]}
VERSION_MINOR=${VERSION_ARRAY[1]}
VERSION_PATCH=${VERSION_ARRAY[2]}

sed -i "s,\(Version = \"\).*\(\"\),\1$VERSION_STRING\2," config/version.go
sed -i "s,\(Version:[[:space:]]*\).*,\1$VERSION_STRING," rpm/SPECS/git-lfs.spec

sed -i "s,\(FILEVERSION \).*,\1$VERSION_MAJOR\,$VERSION_MINOR\,$VERSION_PATCH\,0," script/windows-installer/resources.rc
sed -i "s,\([[:space:]]*VALUE \"ProductVersion\"\, \"\).*\(\\\\0\"\),\1$VERSION_STRING\2," script/windows-installer/resources.rc
