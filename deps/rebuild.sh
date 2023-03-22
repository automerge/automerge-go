#!/usr/bin/env bash

set -e
set -x

if [[ "$(uname)" != "Darwin" || "$(uname -p)" != "arm" ]];
then
    echo "Rebuilding must happen on a Mac with M1"
    exit 1
fi

which cargo || which cmake || (
    echo "Missing dependencies!"
    echo "See https://github.com/automerge/automerge-rs for instructions on setting up the build environment."
    exit 1
)

deps="$(realpath "$(dirname "$0")")"
automerge_c="$deps/automerge-rs/rust/automerge-c"

if [[ "$1" == "clean" ]]; then
    rm -rf "$deps/build"

elif [[ "$1" == "local" ]]; then
    true

elif [[ "$1" != "" ]]; then
    echo "Unknown argument: $1"
    exit 1
fi

mkdir -p "$deps/build"

cmake -B "$deps/build" -S "$automerge_c"
cmake --build "$deps/build"
cp "$deps/build/include/automerge-c/automerge.h" "$deps/.."
cp "$deps/build/Cargo/target/aarch64-apple-darwin/release/libautomerge_core.a" "$deps/libautomerge_core_darwin_arm64.a"

if [[ "$1" == "local" ]]; then
    exit
fi

# cross@0.2.5 is broken on my machine: https://github.com/cross-rs/cross/issues/1214
which cross || cargo install cross@0.2.4

function build() {
    target="$1"
    output="$2"

    RUSTFLAGS="-C panic=abort" cross +nightly build --release --manifest-path="$automerge_c/Cargo.toml" -Z build-std=std,panic_abort --target "$target" --target-dir "$deps/build"

    cp "$deps/build/$target/release/libautomerge_core.a" "$deps/$output"
}

build x86_64-apple-darwin libautomerge_core_darwin_amd64.a
build aarch64-unknown-linux-gnu libautomerge_core_linux_arm64.a
build x86_64-unknown-linux-gnu libautomerge_core_linux_amd64.a