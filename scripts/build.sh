#!/bin/sh

set -e

build_ui() (
    cd internal/server/src && \
    rm ../public/dist/* && \
    npx esbuild "*.js" "*.css" \
        --bundle --minify \
        --outdir=../public/dist \
        --entry-names=[dir]/[name]
)

build_dashi() {
    go build ./cmd/dashi
}

build_ui
build_dashi
