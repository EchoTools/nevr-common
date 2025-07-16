#!/bin/sh
for build_file in */build.go; do
    echo "Generating $build_file"
    go generate $build_file
done