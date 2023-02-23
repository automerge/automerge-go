#!/usr/bin/env bash

set -e
set -x

if [[ "$(uname)" != "Darwin" || "$(uname -p)" != "arm" ]];
then
    echo "Rebuilding must happen on a Mac with M1"
    exit 1
fi

which cargo || (
    echo "Missing dependencies!"
    echo "See https://github.com/automerge/automerge-rs for instructions on setting up the build environment."
    exit 1
)

# cross@0.2.5 is broken on my machine: https://github.com/cross-rs/cross/issues/1214
which cross || cargo install cross@0.2.4

deps="$(realpath "$(dirname "$0")")"
automerge_c="$deps/automerge-rs/rust/automerge-c"

if [[ "$1" == "clean" ]]; then
    rm -rf "$deps/build"

elif [[ "$1" != "" ]]; then
    echo "Unknown argument: $1"
    exit 1
fi

mkdir -p "$deps/build"

CBINDGEN_TARGET_DIR="$deps/.." cargo build -r --manifest-path="$automerge_c/Cargo.toml" --target aarch64-apple-darwin --target-dir "$deps/build"
cargo build -r --manifest-path="$automerge_c/Cargo.toml" --target x86_64-apple-darwin --target-dir "$deps/build"
cross build -r --manifest-path="$automerge_c/Cargo.toml" --target aarch64-unknown-linux-gnu --target-dir "$deps/build"
cross build -r --manifest-path="$automerge_c/Cargo.toml" --target x86_64-unknown-linux-gnu --target-dir "$deps/build"

mkdir -p "$deps/include"
mkdir -p "$deps/darwin_arm64"
mkdir -p "$deps/darwin_amd64"
mkdir -p "$deps/linux_arm64"
mkdir -p "$deps/linux_amd64"

cp "$deps/build/include/automerge-c/automerge.h" "$deps/../"
cp "$deps/build/aarch64-apple-darwin/release/libautomerge_core.a" "$deps/darwin_arm64/"
cp "$deps/build/x86_64-apple-darwin/release/libautomerge_core.a" "$deps/darwin_amd64/"
cp "$deps/build/aarch64-unknown-linux-gnu/release/libautomerge_core.a" "$deps/linux_arm64/"
cp "$deps/build/x86_64-unknown-linux-gnu/release/libautomerge_core.a" "$deps/linux_amd64/"