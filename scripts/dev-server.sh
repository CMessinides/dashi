#!/bin/sh

set -e

watch_ui() (
    cd internal/server/src && \
    npx esbuild "*.js" "*.css" \
        --bundle --watch=forever \
        --outdir=../public/dist \
        --entry-names=[dir]/[name] \
        --sourcemap
)

watch_server() {
    air
}

(trap 'kill 0' INT; \
    watch_ui & \
    watch_server & \
    wait)
